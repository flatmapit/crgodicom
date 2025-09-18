package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	internalcli "github.com/flatmapit/crgodicom/internal/cli"
	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	Version   = "0.0.1-beta"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	// Create CLI application
	app := &cli.App{
		Name:    "crgodicom",
		Usage:   "A cross-platform CLI utility for creating synthetic DICOM data and sending it to PACS systems",
		Version: fmt.Sprintf("%s (built: %s, commit: %s)", Version, BuildDate, GitCommit),
		Authors: []*cli.Author{
			{
				Name:  "flatmapit.com",
				Email: "contact@flatmapit.com",
			},
		},
		Copyright: "© 2025 flatmapit.com - Licensed under the MIT License",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Configuration file path",
				Value:   "crgodicom.yaml",
			},
			&cli.StringFlag{
				Name:  "log-file",
				Usage: "Log file path",
			},
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "Log level (DEBUG, INFO, WARNING, ERROR, CRITICAL)",
				Value: "INFO",
			},
		},
		Before: func(c *cli.Context) error {
			// Initialize configuration
			cfg, err := config.LoadConfig(c.String("config"))
			if err != nil {
				logrus.Warnf("Failed to load config file %s: %v", c.String("config"), err)
				cfg = config.DefaultConfig()
			}

			// Override config with CLI flags
			if c.String("log-file") != "" {
				cfg.Logging.File = c.String("log-file")
			}
			if c.String("log-level") != "" {
				cfg.Logging.Level = c.String("log-level")
			}

			// Initialize logging
			if err := initLogging(cfg.Logging); err != nil {
				return fmt.Errorf("failed to initialize logging: %w", err)
			}

			// Store config in context
			c.Context = context.WithValue(c.Context, "config", cfg)
			return nil
		},
		Commands: []*cli.Command{
			internalcli.CreateCommand(),
			internalcli.ListCommand(),
			internalcli.SendCommand(),
			internalcli.StoreCommand(),
			internalcli.DCMTKCommand(),
			internalcli.AssociateCommand(),
			internalcli.EchoCommand(),
			internalcli.VerifyCommand(),
			internalcli.ExportCommand(),
			internalcli.CreateTemplateCommand(),
			internalcli.CreateCheckDCMTKCommand(),
			internalcli.CreateORMCommand(),
			// Future: internalcli.QueryCommand(),
		},
	}

	// Run the application
	if err := app.RunContext(ctx, os.Args); err != nil {
		logrus.Fatalf("Application error: %v", err)
	}
}

// initLogging initializes the logging system
func initLogging(cfg config.LoggingConfig) error {
	// Parse log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Set formatter
	if cfg.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set log file if specified
	if cfg.File != "" {
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", cfg.File, err)
		}
		logrus.SetOutput(file)
	}

	return nil
}
