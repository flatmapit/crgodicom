package cli

import (
	"fmt"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// VerifyCommand returns the verify command
func VerifyCommand() *cli.Command {
	return &cli.Command{
		Name:  "verify",
		Usage: "Verify PACS connection using C-ECHO",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Usage: "PACS host address",
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "PACS port",
				Value: 11112,
			},
			&cli.StringFlag{
				Name:  "aec",
				Usage: "Application Entity Caller",
			},
			&cli.StringFlag{
				Name:  "aet",
				Usage: "Application Entity Title",
			},
			&cli.IntFlag{
				Name:  "timeout",
				Usage: "Connection timeout in seconds",
				Value: 10,
			},
		},
		Action: verifyAction,
	}
}

func verifyAction(c *cli.Context) error {
	// Get configuration from context
	cfg, ok := c.Context.Value("config").(*config.Config)
	if !ok {
		return fmt.Errorf("configuration not found in context")
	}

	// Build PACS connection parameters
	pacsConfig := config.PACSConfig{
		Host:    c.String("host"),
		Port:    c.Int("port"),
		AEC:     c.String("aec"),
		AET:     c.String("aet"),
		Timeout: c.Int("timeout"),
	}

	// Use default PACS config if not specified via CLI
	if pacsConfig.Host == "" {
		pacsConfig = cfg.DefaultPACS
		logrus.Info("Using default PACS configuration")
	}

	// Validate required PACS parameters
	if pacsConfig.Host == "" || pacsConfig.AEC == "" || pacsConfig.AET == "" {
		return fmt.Errorf("PACS connection requires host, aec, and aet parameters")
	}

	logrus.Infof("Verifying PACS connection to %s:%d (AEC: %s, AET: %s)",
		pacsConfig.Host, pacsConfig.Port, pacsConfig.AEC, pacsConfig.AET)
	logrus.Infof("Timeout: %ds", pacsConfig.Timeout)

	// TODO: Implement actual PACS verification (C-ECHO)
	// For now, just log the parameters
	logrus.Info("PACS verification not yet implemented")

	return nil
}
