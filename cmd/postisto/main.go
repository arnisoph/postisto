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

var build string

func main() {
	app := newApp()

	if err := app.Run(os.Args); err != nil {
		goLog.Fatalln("Failed to start app:", err)
	}
}

func newApp() *cli.App {
	var configPath string
	var logLevel string
	var logJSON bool
	var pollInterval time.Duration
	var onetime bool

	app := cli.App{
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
				Value:       time.Second * 5,
				EnvVars:     []string{"POLL_INTERVAL"},
				Destination: &pollInterval,
			},
			&cli.BoolFlag{
				Name:        "onetime",
				Usage:       "run filter only once and exit the program afterwards",
				Value:       false,
				EnvVars:     []string{"ONETIME"},
				Destination: &onetime,
			},
		},
		Action: func(c *cli.Context) error {
			return runApp(configPath, logLevel, logJSON, pollInterval, onetime)
		},
		Version: build,
	}

	return &app
}

func runApp(configPath string, logLevel string, logJSON bool, pollInterval time.Duration, onetime bool) error {

	if err := log.InitWithConfig(logLevel, logJSON); err != nil {
		return err
	}

	log.Info("Welcome, thanks for using poŝtisto! If you experience any problems or questions please raise an issue on Github (https://github.com/arnisoph/postisto).")

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

	type accInfo struct {
		name    string
		acc     *config.Account
		filters map[string]filter.Filter
	}

	var accs []*accInfo
	for name, _ := range cfg.Accounts {
		filters, ok := cfg.Filters[name]
		if !ok {
			return fmt.Errorf("no filter configuration found for account %v. nothing to do", name)
		}

		acc := cfg.Accounts[name]

		if err := acc.Connection.Connect(); err != nil {
			return fmt.Errorf("failed to initially connect to server %q with username %q", acc.Connection.Server, acc.Connection.Username)
		}

		accs = append(accs, &accInfo{name: name, acc: &acc, filters: filters})
	}

	if onetime {
		log.Info("Entering mail search & filter loop once and exit then immediately")
	} else {
		log.Info("Entering continuously running mail search & filter loop. Waiting for mails...")
	}

	for {
		for _, accInfo := range accs {
			if err := filter.EvaluateFilterSetsOnMsgs(&accInfo.acc.Connection, *accInfo.acc.InputMailbox, []string{server.SeenFlag, server.FlaggedFlag}, *accInfo.acc.FallbackMailbox, accInfo.filters); err != nil {
				if server.IsDisconnected(err) {
					// this can happen, so let's just reconnect
					//TODO implement exponential backoff to avoid too much noise?

					time.Sleep(time.Second * 3)
					if err := accInfo.acc.Connection.Connect(); err != nil {
						return fmt.Errorf("failed to reconnect to server %q with username %q", accInfo.acc.Connection.Server, accInfo.acc.Connection.Username)
					}

					continue
				}

				return fmt.Errorf("failed to run filter engine: %v", err)
			}
		}

		if onetime {
			return nil
		}

		time.Sleep(pollInterval)
	}
}
