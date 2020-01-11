package main

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/filter"
	"github.com/arnisoph/postisto/pkg/log"
	"github.com/arnisoph/postisto/pkg/server"
	"github.com/urfave/cli/v2"
	goLog "log"
	"os"
	"time"
)

func main() {
	var configPath string
	var logLevel string
	var logJSON bool
	var pollInterval time.Duration

	app := &cli.App{
		Name:  "poŝtisto",
		Usage: "quite okay mail-sorting",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "config file or directory path",
				Value:       "config/",
				EnvVars:     []string{"CONFIG_PATH"},
				Destination: &configPath,
			},
			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				Usage:       "log level e.g. trace, debug, info or error (WARNING: trace exposes account credentials and more sensitive data)",
				Value:       "info",
				EnvVars:     []string{"LOG_LEVEL"},
				Destination: &logLevel,
			},
			&cli.BoolFlag{
				Name:        "log-json",
				Aliases:     []string{"j"},
				Usage:       "format log output as JSON",
				Value:       false,
				EnvVars:     []string{"LOG_JSON"},
				Destination: &logJSON,
			},
			&cli.DurationFlag{
				Name:        "poll-interval",
				Aliases:     []string{"i"},
				Usage:       "duration to wait between checking for new messages in input mailbox",
				Value:       5,
				EnvVars:     []string{"POLL_INTERVAL"},
				Destination: &pollInterval,
			},
		},
		Action: func(c *cli.Context) error {
			return startApp(c, configPath, logLevel, logJSON, pollInterval)
		},
	}

	if err := app.Run(os.Args); err != nil {
		goLog.Fatalln("Failed to start app:", err)
	}
}

func startApp(c *cli.Context, configPath string, logLevel string, logJSON bool, pollInterval time.Duration) error {

	if err := log.InitWithConfig(logLevel, logJSON); err != nil {
		return err
	}

	var cfg *config.Config
	var err error

	if cfg, err = config.NewConfigFromFile(configPath); err != nil {
		return err
	}

	if len(cfg.Accounts) == 0 {
		return fmt.Errorf("no (enabled) account configuration found. nothing to do")
	}

	if len(cfg.Filters) == 0 {
		return fmt.Errorf("no filter configuration found. nothing to do")
	}

	type accTuple struct {
		acc     *config.Account
		filters *map[string]filter.Filter
	}
	var accs []accTuple
	for name, acc := range cfg.Accounts {
		filters, ok := cfg.Filters[name]
		if !ok {
			return fmt.Errorf("no filter configuration found for account %v. nothing to do", name)
		}

		accs = append(accs, accTuple{acc: &acc, filters: &filters})
		if err := acc.Connection.Connect(); err != nil {
			return err
		}
	}

	for {
		for _, accTuple := range accs {
			if err := filter.EvaluateFilterSetsOnMsgs(&accTuple.acc.Connection, *accTuple.acc.InputMailbox, []string{server.SeenFlag, server.FlaggedFlag}, *accTuple.acc.FallbackMailbox, *accTuple.filters); err != nil {
				return fmt.Errorf("failed to run filter engine: %v", err)
			}
		}

		time.Sleep(pollInterval)
	}
}
