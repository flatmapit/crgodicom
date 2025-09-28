package pacs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/sirupsen/logrus"
)

// DCMTK handles all DICOM protocol details, so we only need the Client struct

// Client represents a DICOM PACS client using DCMTK
type Client struct {
	config  *config.PACSConfig
	verbose bool
	debug   bool
}

// ClientOptions contains options for creating a PACS client
type ClientOptions struct {
	Verbose bool
	Debug   bool
}

// NewClient creates a new PACS client
func NewClient(cfg *config.PACSConfig) *Client {
	return &Client{
		config:  cfg,
		verbose: false,
		debug:   false,
	}
}

// NewClientWithOptions creates a new PACS client with options
func NewClientWithOptions(cfg *config.PACSConfig, opts *ClientOptions) *Client {
	verbose := false
	debug := false
	if opts != nil {
		verbose = opts.Verbose
		debug = opts.Debug
	}
	return &Client{
		config:  cfg,
		verbose: verbose,
		debug:   debug,
	}
}

// Connect establishes a DICOM association with the PACS server using DCMTK
func (c *Client) Connect(ctx context.Context) error {
	if c.verbose {
		logrus.Infof("🔗 Connecting to PACS server at %s:%d", c.config.Host, c.config.Port)
		logrus.Infof("📡 AEC: %s, AET: %s", c.config.AEC, c.config.AET)
		logrus.Infof("⏱️  Timeout: %ds", c.config.Timeout)
	}

	// Validate PACS configuration
	if err := c.validateConfig(); err != nil {
		return fmt.Errorf("invalid PACS configuration: %w", err)
	}

	// DCMTK handles connection internally, no need to maintain connection state
	if c.verbose {
		logrus.Info("✅ PACS connection configuration validated")
	}
	return nil
}

// validateConfig validates the PACS configuration
func (c *Client) validateConfig() error {
	if c.config.Host == "" {
		return fmt.Errorf("PACS host is required")
	}
	if c.config.Port <= 0 || c.config.Port > 65535 {
		return fmt.Errorf("invalid PACS port: %d (must be 1-65535)", c.config.Port)
	}
	if c.config.AEC == "" {
		return fmt.Errorf("PACS AEC (Application Entity Caller) is required")
	}
	if c.config.AET == "" {
		return fmt.Errorf("PACS AET (Application Entity Title) is required")
	}
	if c.config.Timeout <= 0 {
		return fmt.Errorf("invalid timeout: %ds (must be > 0)", c.config.Timeout)
	}
	return nil
}

// Disconnect closes the DICOM association and connection to the PACS server
func (c *Client) Disconnect() error {
	if c.verbose {
		logrus.Info("🔌 Disconnecting from PACS server")
		logrus.Info("✅ DCMTK handles disconnection automatically")
	}
	return nil
}

// CEcho performs a C-ECHO request to verify connectivity using DCMTK
func (c *Client) CEcho(ctx context.Context) error {
	if c.verbose {
		logrus.Info("📡 Performing C-ECHO request to verify PACS connectivity")
		logrus.Infof("🎯 Target: %s:%d (AEC: %s, AET: %s)", c.config.Host, c.config.Port, c.config.AEC, c.config.AET)
	}

	err := c.performEchoWithDCMTK()
	if err != nil {
		if c.debug {
			logrus.Errorf("❌ C-ECHO failed with detailed error: %v", err)
		}
		return fmt.Errorf("C-ECHO failed: %w", err)
	}

	if c.verbose {
		logrus.Info("✅ C-ECHO completed successfully - PACS is reachable")
	}
	return nil
}

// performEchoWithDCMTK performs the actual C-ECHO using DCMTK tools
func (c *Client) performEchoWithDCMTK() error {
	// Use echoscu command from DCMTK
	cmd := exec.Command("echoscu",
		"-aec", c.config.AEC,
		"-aet", c.config.AET,
		c.config.Host,
		fmt.Sprintf("%d", c.config.Port))

	if c.debug {
		logrus.Debugf("🔧 Running command: %s", strings.Join(cmd.Args, " "))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if c.debug {
			logrus.Errorf("📋 echoscu output: %s", string(output))
		}
		return fmt.Errorf("echoscu command failed: %w", err)
	}

	if c.debug {
		logrus.Debugf("📋 echoscu output: %s", string(output))
	}

	return nil
}

// CStore performs a C-STORE request to send a DICOM file using DCMTK
func (c *Client) CStore(ctx context.Context, dicomData []byte, sopInstanceUID string) error {
	if c.verbose {
		logrus.Infof("📤 Performing C-STORE request for SOP Instance UID: %s", sopInstanceUID)
		logrus.Infof("📊 Data size: %d bytes", len(dicomData))
	}

	// Write DICOM data to temporary file
	tempFile := fmt.Sprintf("/tmp/crgodicom_store_%s.dcm", sopInstanceUID)
	if err := c.writeTempDicomFile(tempFile, dicomData); err != nil {
		return fmt.Errorf("failed to create temporary DICOM file: %w", err)
	}
	defer os.Remove(tempFile) // Clean up temp file

	if c.debug {
		logrus.Debugf("📁 Created temporary file: %s", tempFile)
	}

	// Use DCMTK to send the file
	err := c.performStoreWithDCMTK(tempFile, sopInstanceUID)
	if err != nil {
		if c.debug {
			logrus.Errorf("❌ C-STORE failed with detailed error: %v", err)
		}
		return fmt.Errorf("C-STORE failed: %w", err)
	}

	if c.verbose {
		logrus.Info("✅ C-STORE completed successfully")
	}
	return nil
}

// performStoreWithDCMTK performs the actual C-STORE using DCMTK tools
func (c *Client) performStoreWithDCMTK(tempFile, sopInstanceUID string) error {
	// Use storescu command from DCMTK with JPEG support
	cmd := exec.Command("storescu",
		"-aec", c.config.AEC,
		"-aet", c.config.AET,
		"-xy", // Propose JPEG lossy transfer syntax
		c.config.Host,
		fmt.Sprintf("%d", c.config.Port),
		tempFile)

	if c.debug {
		logrus.Debugf("🔧 Running command: %s", strings.Join(cmd.Args, " "))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if c.debug {
			logrus.Errorf("📋 storescu output: %s", string(output))
		}
		return fmt.Errorf("storescu command failed: %w", err)
	}

	if c.debug {
		logrus.Debugf("📋 storescu output: %s", string(output))
	}

	return nil
}

// writeTempDicomFile writes DICOM data to a temporary file
func (c *Client) writeTempDicomFile(filename string, data []byte) error {
	if c.debug {
		logrus.Debugf("📝 Writing %d bytes to temporary file: %s", len(data), filename)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create temporary file %s: %w", filename, err)
	}
	defer file.Close()

	bytesWritten, err := file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to temporary file %s: %w", filename, err)
	}

	if bytesWritten != len(data) {
		return fmt.Errorf("incomplete write to temporary file %s: wrote %d/%d bytes", filename, bytesWritten, len(data))
	}

	if c.debug {
		logrus.Debugf("✅ Successfully wrote %d bytes to %s", bytesWritten, filename)
	}

	return nil
}

// findPresentationContext finds the appropriate presentation context for a SOP class
func (c *Client) findPresentationContext(sopClassUID string) uint8 {
	// Stub implementation - return a default context ID
	switch sopClassUID {
	case "1.2.840.10008.5.1.4.1.1.2": // CT Image Storage
		return 1
	case "1.2.840.10008.5.1.4.1.1.1": // Computed Radiography Image Storage
		return 3
	case "1.2.840.10008.5.1.4.1.1.4": // MR Image Storage
		return 5
	default:
		return 15 // Secondary Capture Image Storage
	}
}
