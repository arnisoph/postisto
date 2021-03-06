package config

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/filter"
	"github.com/arnisoph/postisto/pkg/log"
	"github.com/arnisoph/postisto/pkg/server"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Accounts map[string]Account                  `yaml:"accounts"`
	Filters  map[string]map[string]filter.Filter `yaml:"filters"`
}

type Account struct {
	Enable          bool              `yaml:"enable"`
	Connection      server.Connection `yaml:"connection"`
	InputMailbox    *string           `yaml:"input"`
	FallbackMailbox *string           `yaml:"fallback"`
}

func NewConfig() *Config {
	return new(Config)
}

func NewConfigFromFile(configPath string) (*Config, error) {
	cfg := NewConfig()
	var configFiles []string
	passwords := map[string]string{}
	var err error

	log.Debugw("Starting to parse config", "configPath", configPath)

	log.Debugw("configPath is a directory. Starting to recursively walk through the directory tree.", "configPath", configPath)
	configFiles, passwords, err = walkConfigPath(configPath)
	if err != nil {
		log.Errorw("Failed to parse dir", err, "configPath", configPath)
		return nil, err
	}

	for _, file := range configFiles {
		log.Debugw("Parsing config YAML file", "file", file)

		fileCfg := new(Config)
		yamlFile, err := ioutil.ReadFile(file)

		if err != nil {
			log.Errorw("Failed to read file", err, "file", file)
			return nil, err
		}

		// YAML to Config struct
		err = yaml.Unmarshal(yamlFile, &fileCfg)

		if err != nil {
			log.Errorw("Failed to parse YAML file", err, "file", file)
			return nil, err
		}
		log.Debugw("Successfully parsed YAML file", "file", file, "parsedFile", string(yamlFile))

		// Merge configs from files
		if err := mergo.Merge(cfg, fileCfg, mergo.WithOverride, mergo.WithTypeCheck); err != nil {
			log.Errorw("Failed to merge YAML file", err, "file", file)
			return nil, err
		}
	}

	log.Debugw("Successfully parsed all YAML files, checking for validity now", "cfg", cfg)
	newCfg, err := cfg.validate(passwords)
	if err != nil {
		return nil, err
	}

	log.Debugw("Configuration successfully loaded & validated", "configPath", configPath)
	return newCfg, nil
}

func (cfg Config) validate(passwords map[string]string) (*Config, error) {
	valCfg := Config{
		Accounts: map[string]Account{},
		Filters:  map[string]map[string]filter.Filter{},
	}

	// Accounts
	if len(cfg.Accounts) == 0 {
		log.Info("Warning: no accounts configured")
	}

	for accName, acc := range cfg.Accounts {
		if !acc.Enable {
			continue
		}

		newAcc := Account{
			Connection:      acc.Connection,
			InputMailbox:    acc.InputMailbox,
			FallbackMailbox: acc.FallbackMailbox,
		}
		// Connection
		if strings.TrimSpace(acc.Connection.Server) == "" {
			return nil, fmt.Errorf("server not configured")
		}

		if filePwd, ok := passwords[accName]; ok {
			log.Debugw("Setting pwd for an account from previously loaded pwd file", "account", accName)
			newAcc.Connection.Password = strings.TrimSpace(filePwd)
		}

		// Input
		if newAcc.InputMailbox == nil || *newAcc.InputMailbox == "" {
			newAcc.InputMailbox = new(string)
			*newAcc.InputMailbox = "INBOX"
		}

		if newAcc.FallbackMailbox == nil {
			newAcc.FallbackMailbox = new(string)
			*newAcc.FallbackMailbox = "INBOX"
		}

		valCfg.Accounts[accName] = newAcc
	}

	// Filters
	valCfg.Filters = cfg.Filters

	if len(valCfg.Filters) == 0 {
		log.Info("Warning: no filters configured")
	}

	return &valCfg, nil
}

func walkConfigPath(configPath string) ([]string, map[string]string, error) {

	var configFiles []string
	passwords := map[string]string{}

	err := filepath.Walk(configPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorw("Failed to load path", err, "path", path)
			return err
		}

		log.Debugw("Checking a file", "path", path)

		if stat, err := os.Stat(path); err != nil {
			log.Errorw("Failed to load path", err, "path", path)
			return err
		} else if !stat.IsDir() && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			configFiles = append(configFiles, path)
		} else if !stat.IsDir() && strings.HasPrefix(filepath.Base(path), ".postisto") && strings.HasSuffix(path, ".pwd") {
			pathFields := strings.Split(path, ".")

			log.Debugw("Starting to read postisto pwd file", "path", path)

			password, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			if string(password) == "" {
				return fmt.Errorf("postisto pwd file is empty")
			}

			passwords[pathFields[len(pathFields)-2]] = string(password)

			log.Infow("Successfully read postisto pwd file. Deleting it now to prevent others to obtain the plaintext password!", "path", path)
			return os.Remove(path)
		}

		return nil
	})

	return configFiles, passwords, err
}
