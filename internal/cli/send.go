package cli

import (
	"fmt"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// SendCommand returns the send command
func SendCommand() *cli.Command {
	return &cli.Command{
		Name:  "send",
		Usage: "Send DICOM study to PACS",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "study-id",
				Usage:    "Study Instance UID (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "host",
				Usage:   "PACS host address",
			},
			&cli.IntFlag{
				Name:    "port",
				Usage:   "PACS port",
				Value:   11112,
			},
			&cli.StringFlag{
				Name:    "aec",
				Usage:   "Application Entity Caller",
			},
			&cli.StringFlag{
				Name:    "aet",
				Usage:   "Application Entity Title",
			},
			&cli.StringFlag{
				Name:    "output-dir",
				Usage:   "Studies directory",
				Value:   "studies",
			},
			&cli.IntFlag{
				Name:    "timeout",
				Usage:   "Connection timeout in seconds",
				Value:   30,
			},
			&cli.IntFlag{
				Name:    "retries",
				Usage:   "Retry attempts",
				Value:   3,
			},
		},
		Action: sendAction,
	}
}

func sendAction(c *cli.Context) error {
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

	studyID := c.String("study-id")
	outputDir := c.String("output-dir")
	retries := c.Int("retries")

	logrus.Infof("Sending study %s to PACS %s:%d (AEC: %s, AET: %s)", 
		studyID, pacsConfig.Host, pacsConfig.Port, pacsConfig.AEC, pacsConfig.AET)
	logrus.Infof("Studies directory: %s, Retries: %d, Timeout: %ds", 
		outputDir, retries, pacsConfig.Timeout)

	// TODO: Implement actual PACS sending
	// For now, just log the parameters
	logrus.Info("PACS sending not yet implemented")

	return nil
}
