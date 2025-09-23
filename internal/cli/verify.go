package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/pacs"
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

	// Perform actual PACS verification using C-ECHO
	return performPACSVerification(&pacsConfig)
}

// performPACSVerification performs actual C-ECHO test with the PACS server
func performPACSVerification(pacsConfig *config.PACSConfig) error {
	logrus.Info("🔍 Starting PACS verification...")

	// Create PACS client
	client := pacs.NewClient(pacsConfig)
	if client == nil {
		return fmt.Errorf("failed to create PACS client")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pacsConfig.Timeout)*time.Second)
	defer cancel()

	// Step 1: Connect to PACS (includes association negotiation)
	logrus.Info("📡 Connecting to PACS server and establishing association...")
	if err := client.Connect(ctx); err != nil {
		logrus.Errorf("❌ Connection/association failed: %v", err)
		return fmt.Errorf("failed to connect to PACS: %w", err)
	}
	logrus.Info("✅ Connected to PACS server and association established")

	// Step 2: Perform C-ECHO
	logrus.Info("🏓 Performing C-ECHO test...")
	if err := client.CEcho(ctx); err != nil {
		logrus.Errorf("❌ C-ECHO failed: %v", err)
		client.Disconnect()
		return fmt.Errorf("C-ECHO test failed: %w", err)
	}
	logrus.Info("✅ C-ECHO test successful")

	// Step 3: Clean up connection
	logrus.Info("🔚 Disconnecting from PACS...")
	client.Disconnect()
	logrus.Info("✅ PACS verification completed successfully!")

	// Print success summary
	fmt.Printf("✅ PACS Verification SUCCESSFUL\n")
	fmt.Printf("   Host: %s:%d\n", pacsConfig.Host, pacsConfig.Port)
	fmt.Printf("   AEC: %s, AET: %s\n", pacsConfig.AEC, pacsConfig.AET)
	fmt.Printf("   C-ECHO: ✅ PASSED\n")

	return nil
}
