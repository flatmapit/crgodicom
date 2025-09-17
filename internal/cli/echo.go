package cli

import (
	"context"
	"time"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/pacs"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// EchoCommand returns the C-ECHO command
func EchoCommand() *cli.Command {
	return &cli.Command{
		Name:  "echo",
		Usage: "Send C-ECHO request to PACS server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Usage: "PACS host address",
				Value: "localhost",
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "PACS port",
				Value: 4242,
			},
			&cli.StringFlag{
				Name:  "aec",
				Usage: "Application Entity Caller",
				Value: "DICOM_CLIENT",
			},
			&cli.StringFlag{
				Name:  "aet",
				Usage: "Application Entity Title",
				Value: "PACS1",
			},
			&cli.IntFlag{
				Name:  "timeout",
				Usage: "Connection timeout in seconds",
				Value: 30,
			},
		},
		Action: echoAction,
	}
}

func echoAction(c *cli.Context) error {
	logrus.Infof("Sending C-ECHO to PACS %s:%d (AEC: %s, AET: %s)",
		c.String("host"), c.Int("port"), c.String("aec"), c.String("aet"))

	// Create PACS configuration
	pacsConfig := &config.PACSConfig{
		Host:    c.String("host"),
		Port:    c.Int("port"),
		AEC:     c.String("aec"),
		AET:     c.String("aet"),
		Timeout: c.Int("timeout"),
	}

	// Create PACS client
	client := pacs.NewClient(pacsConfig)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pacsConfig.Timeout)*time.Second)
	defer cancel()

	// Connect to PACS
	logrus.Info("Establishing DICOM association...")
	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Disconnect()

	logrus.Info("DICOM association established successfully")

	// Send C-ECHO
	logrus.Info("Sending C-ECHO request...")
	if err := client.CEcho(ctx); err != nil {
		return err
	}

	logrus.Info("C-ECHO successful - PACS is responding")
	return nil
}
