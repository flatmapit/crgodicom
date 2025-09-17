package pacs

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/sirupsen/logrus"
)

// Client represents a DICOM PACS client
type Client struct {
	config *config.PACSConfig
	conn   net.Conn
}

// NewClient creates a new PACS client
func NewClient(cfg *config.PACSConfig) *Client {
	return &Client{
		config: cfg,
	}
}

// Connect establishes a connection to the PACS server
func (c *Client) Connect(ctx context.Context) error {
	address := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	
	logrus.Infof("Connecting to PACS at %s", address)
	
	// Create connection with timeout
	dialer := &net.Dialer{
		Timeout: time.Duration(c.config.Timeout) * time.Second,
	}
	
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to PACS: %w", err)
	}
	
	c.conn = conn
	logrus.Infof("Connected to PACS server %s", address)
	
	return nil
}

// Disconnect closes the connection to the PACS server
func (c *Client) Disconnect() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		if err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
		logrus.Info("Disconnected from PACS server")
	}
	return nil
}

// CEcho performs a C-ECHO request to verify connectivity
func (c *Client) CEcho(ctx context.Context) error {
	logrus.Info("Performing C-ECHO request")
	
	// TODO: Implement actual DICOM C-ECHO protocol
	// For now, just verify the connection is alive
	if c.conn == nil {
		return fmt.Errorf("not connected to PACS server")
	}
	
	// Simple connectivity test
	deadline := time.Now().Add(time.Duration(c.config.Timeout) * time.Second)
	if err := c.conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}
	
	// Send a simple ping-like message
	testMsg := []byte("DICOM_C_ECHO_REQUEST")
	if _, err := c.conn.Write(testMsg); err != nil {
		return fmt.Errorf("failed to send C-ECHO request: %w", err)
	}
	
	// Read response
	buffer := make([]byte, 1024)
	if _, err := c.conn.Read(buffer); err != nil {
		return fmt.Errorf("failed to receive C-ECHO response: %w", err)
	}
	
	logrus.Info("C-ECHO successful - PACS server is responding")
	return nil
}

// CStore performs a C-STORE request to send a DICOM file
func (c *Client) CStore(ctx context.Context, dicomData []byte, sopInstanceUID string) error {
	logrus.Infof("Performing C-STORE request for SOP Instance UID: %s", sopInstanceUID)
	
	if c.conn == nil {
		return fmt.Errorf("not connected to PACS server")
	}
	
	// TODO: Implement actual DICOM C-STORE protocol
	// For now, just send the DICOM data
	deadline := time.Now().Add(time.Duration(c.config.Timeout) * time.Second)
	if err := c.conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}
	
	// Send DICOM data length first
	length := uint32(len(dicomData))
	lengthBytes := []byte{
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}
	
	if _, err := c.conn.Write(lengthBytes); err != nil {
		return fmt.Errorf("failed to send data length: %w", err)
	}
	
	// Send DICOM data
	if _, err := c.conn.Write(dicomData); err != nil {
		return fmt.Errorf("failed to send DICOM data: %w", err)
	}
	
	// Read response
	buffer := make([]byte, 1024)
	if _, err := c.conn.Read(buffer); err != nil {
		return fmt.Errorf("failed to receive C-STORE response: %w", err)
	}
	
	logrus.Infof("C-STORE successful for SOP Instance UID: %s", sopInstanceUID)
	return nil
}

// IsConnected returns true if the client is connected to a PACS server
func (c *Client) IsConnected() bool {
	return c.conn != nil
}

// GetConfig returns the PACS configuration
func (c *Client) GetConfig() *config.PACSConfig {
	return c.config
}
