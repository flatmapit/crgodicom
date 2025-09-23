package pacs

import (
	"context"
	"fmt"
	"os"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/dcmtk"
	"github.com/sirupsen/logrus"
)

// DCMTK handles all DICOM protocol details, so we only need the Client struct

// Client represents a DICOM PACS client using DCMTK
type Client struct {
	config *config.PACSConfig
}

// NewClient creates a new PACS client
func NewClient(cfg *config.PACSConfig) *Client {
	return &Client{
		config: cfg,
	}
}

// Connect establishes a DICOM association with the PACS server using DCMTK
func (c *Client) Connect(ctx context.Context) error {
	logrus.Infof("DCMTK will handle connection to PACS at %s:%d", c.config.Host, c.config.Port)
	logrus.Infof("AEC: %s, AET: %s", c.config.AEC, c.config.AET)
	
	// DCMTK handles connection internally, no need to maintain connection state
	return nil
}

// Disconnect closes the DICOM association and connection to the PACS server
func (c *Client) Disconnect() error {
	logrus.Info("DCMTK handles disconnection automatically")
	return nil
}

// CEcho performs a C-ECHO request to verify connectivity using DCMTK
func (c *Client) CEcho(ctx context.Context) error {
	logrus.Info("Performing C-ECHO request using DCMTK")

	err := dcmtk.Echo(c.config.Host, c.config.Port, c.config.AET, c.config.AEC)
	if err != nil {
		return fmt.Errorf("C-ECHO failed: %w", err)
	}

	logrus.Info("C-ECHO completed successfully")
	return nil
}

// CStore performs a C-STORE request to send a DICOM file using DCMTK
func (c *Client) CStore(ctx context.Context, dicomData []byte, sopInstanceUID string) error {
	logrus.Infof("Performing C-STORE request for SOP Instance UID: %s", sopInstanceUID)

	// Write DICOM data to temporary file
	tempFile := fmt.Sprintf("/tmp/crgodicom_store_%s.dcm", sopInstanceUID)
	if err := c.writeTempDicomFile(tempFile, dicomData); err != nil {
		return fmt.Errorf("failed to create temporary DICOM file: %w", err)
	}

	// Use DCMTK to send the file
	err := dcmtk.Store(c.config.Host, c.config.Port, c.config.AET, c.config.AEC, tempFile)
	if err != nil {
		return fmt.Errorf("C-STORE failed: %w", err)
	}

	logrus.Info("C-STORE completed successfully")
	return nil
}

// writeTempDicomFile writes DICOM data to a temporary file
func (c *Client) writeTempDicomFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

