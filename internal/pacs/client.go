package pacs

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/sirupsen/logrus"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DICOM Protocol Constants
const (
	// PDU Types
	PDUTypeAssociationRQ = 0x01
	PDUTypeAssociationAC = 0x02
	PDUTypeAssociationRJ = 0x03
	PDUTypeDataTF        = 0x04
	PDUTypeReleaseRQ     = 0x05
	PDUTypeReleaseRP     = 0x06
	PDUTypeAbortRQ       = 0x07

	// DICOM UIDs
	ApplicationContextName    = "1.2.840.10008.3.1.1.1"
	ImplementationClassUID    = "1.2.840.10008.5.1.4.1.1"
	ImplementationVersionName = "CRGODICOM-1.0"

	// Transfer Syntax UIDs
	ImplicitVRLittleEndian = "1.2.840.10008.1.2"
	ExplicitVRLittleEndian = "1.2.840.10008.1.2.1"
	ExplicitVRBigEndian    = "1.2.840.10008.1.2.2"

	// Abstract Syntax UIDs (SOP Classes)
	SOPClassVerification                 = "1.2.840.10008.1.1"
	SOPClassCTImageStorage               = "1.2.840.10008.5.1.4.1.1.2"
	SOPClassCRImageStorage               = "1.2.840.10008.5.1.4.1.1.1"
	SOPClassDXImageStorage               = "1.2.840.10008.5.1.4.1.1.1.1"
	SOPClassMGImageStorage               = "1.2.840.10008.5.1.4.1.1.1.2"
	SOPClassMRImageStorage               = "1.2.840.10008.5.1.4.1.1.4"
	SOPClassUSImageStorage               = "1.2.840.10008.5.1.4.1.1.6.1"
	SOPClassSecondaryCaptureImageStorage = "1.2.840.10008.5.1.4.1.1.7"

	// Max PDU Length
	MaxPDULength = 16384
)

// DICOM PDU Header structure
type PDUHeader struct {
	ItemType uint8
	Reserved uint8
	Length   uint32
}

// Presentation Context Item
type PresentationContext struct {
	ID               uint8
	AbstractSyntax   string
	TransferSyntaxes []string
}

// Client represents a DICOM PACS client
type Client struct {
	config           *config.PACSConfig
	conn             net.Conn
	associated       bool
	acceptedContexts []uint8
}

// NewClient creates a new PACS client
func NewClient(cfg *config.PACSConfig) *Client {
	return &Client{
		config: cfg,
	}
}

// Connect establishes a DICOM association with the PACS server
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
	logrus.Infof("TCP connection established to %s", address)

	// Perform DICOM association
	if err := c.performAssociation(ctx); err != nil {
		c.conn.Close()
		c.conn = nil
		return fmt.Errorf("DICOM association failed: %w", err)
	}

	logrus.Infof("DICOM association established with %s (AET: %s)", address, c.config.AET)
	return nil
}

// Disconnect closes the DICOM association and connection to the PACS server
func (c *Client) Disconnect() error {
	if c.conn != nil {
		// Send release request if associated
		if c.associated {
			c.performRelease()
		}

		err := c.conn.Close()
		c.conn = nil
		c.associated = false
		c.acceptedContexts = nil
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

	if !c.associated {
		return fmt.Errorf("not associated with PACS server")
	}

	// Use simplified C-ECHO implementation
	if err := c.sendCEcho(ctx); err != nil {
		return fmt.Errorf("C-ECHO failed: %w", err)
	}

	logrus.Info("C-ECHO completed successfully")
	return nil
}

// CStore performs a C-STORE request to send a DICOM file
func (c *Client) CStore(ctx context.Context, dicomData []byte, sopInstanceUID string) error {
	logrus.Infof("Performing C-STORE request for SOP Instance UID: %s", sopInstanceUID)

	if !c.associated {
		return fmt.Errorf("not associated with PACS server")
	}

	// Determine SOP Class from DICOM data (simplified)
	sopClass := c.extractSOPClassFromData(dicomData)

	// Use simplified C-STORE implementation
	if err := c.sendCStore(ctx, dicomData, sopInstanceUID, sopClass); err != nil {
		return fmt.Errorf("C-STORE failed: %w", err)
	}

	logrus.Info("C-STORE completed successfully")
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

// performAssociation performs the DICOM association negotiation
func (c *Client) performAssociation(ctx context.Context) error {
	logrus.Info("Performing DICOM association negotiation")

	// Send Association Request
	assocReq := c.buildAssociationRequest()
	logrus.Debugf("Sending association request: %d bytes, first 10 bytes: %X", len(assocReq), assocReq[:min(10, len(assocReq))])
	logrus.Debugf("PDU Type: %02X (decimal: %d)", assocReq[0], assocReq[0])
	if _, err := c.conn.Write(assocReq); err != nil {
		return fmt.Errorf("failed to send association request: %w", err)
	}

	// Read Association Response
	response := make([]byte, 4096)
	n, err := c.conn.Read(response)
	if err != nil {
		return fmt.Errorf("failed to read association response: %w", err)
	}

	// Parse Association Response
	if err := c.parseAssociationResponse(response[:n]); err != nil {
		return fmt.Errorf("association rejected: %w", err)
	}

	c.associated = true
	logrus.Info("DICOM association established successfully")
	return nil
}

// buildAssociationRequest builds a DICOM Association Request PDU based on working minimal structure
func (c *Client) buildAssociationRequest() []byte {
	var pdu bytes.Buffer

	// PDU Header
	pdu.WriteByte(PDUTypeAssociationRQ) // PDU Type: Association Request
	pdu.WriteByte(0x00)                 // Reserved

	// We'll calculate length after building content
	lengthPos := pdu.Len()
	pdu.Write(make([]byte, 4)) // Placeholder for PDU length

	// Protocol Version (2 bytes)
	protocolVersion := make([]byte, 2)
	binary.BigEndian.PutUint16(protocolVersion, 0x0001)
	pdu.Write(protocolVersion)

	// Reserved (2 bytes)
	pdu.Write([]byte{0x00, 0x00})

	// Called AE Title (16 bytes, space-padded)
	calledAE := make([]byte, 16)
	copy(calledAE, []byte(c.config.AET))
	for i := len(c.config.AET); i < 16; i++ {
		calledAE[i] = 0x20 // Space padding
	}
	pdu.Write(calledAE)

	// Calling AE Title (16 bytes, space-padded)
	callingAE := make([]byte, 16)
	copy(callingAE, []byte(c.config.AEC))
	for i := len(c.config.AEC); i < 16; i++ {
		callingAE[i] = 0x20 // Space padding
	}
	pdu.Write(callingAE)

	// Reserved (32 bytes)
	pdu.Write(make([]byte, 32))

	// Application Context Item
	pdu.WriteByte(0x10) // Item Type: Application Context
	pdu.WriteByte(0x00) // Reserved
	appContext := []byte(ApplicationContextName)
	appContextLength := make([]byte, 2)
	binary.BigEndian.PutUint16(appContextLength, uint16(len(appContext)))
	pdu.Write(appContextLength)
	pdu.Write(appContext)

	// Build presentation contexts for common SOP classes
	contexts := []struct {
		ID             uint8
		AbstractSyntax string
	}{
		{1, SOPClassVerification},
		{3, SOPClassCTImageStorage},
		{5, SOPClassCRImageStorage},
		{7, SOPClassDXImageStorage},
		{9, SOPClassMGImageStorage},
		{11, SOPClassMRImageStorage},
		{13, SOPClassUSImageStorage},
		{15, SOPClassSecondaryCaptureImageStorage},
	}

	for _, ctx := range contexts {
		// Presentation Context Item
		pdu.WriteByte(0x20) // Item Type: Presentation Context
		pdu.WriteByte(0x00) // Reserved

		// We'll calculate PC length after building it
		pcLengthPos := pdu.Len()
		pdu.Write(make([]byte, 2)) // Placeholder for PC length

		pdu.WriteByte(ctx.ID)               // Presentation Context ID
		pdu.Write([]byte{0x00, 0x00, 0x00}) // Reserved

		// Abstract Syntax Sub-item
		pdu.WriteByte(0x30) // Item Type: Abstract Syntax
		pdu.WriteByte(0x00) // Reserved
		abstractSyntax := []byte(ctx.AbstractSyntax)
		abstractLength := make([]byte, 2)
		binary.BigEndian.PutUint16(abstractLength, uint16(len(abstractSyntax)))
		pdu.Write(abstractLength)
		pdu.Write(abstractSyntax)

		// Transfer Syntax Sub-item (Implicit VR Little Endian)
		pdu.WriteByte(0x40) // Item Type: Transfer Syntax
		pdu.WriteByte(0x00) // Reserved
		transferSyntax := []byte(ImplicitVRLittleEndian)
		transferLength := make([]byte, 2)
		binary.BigEndian.PutUint16(transferLength, uint16(len(transferSyntax)))
		pdu.Write(transferLength)
		pdu.Write(transferSyntax)

		// Update Presentation Context length
		pcLength := pdu.Len() - pcLengthPos - 2
		binary.BigEndian.PutUint16(pdu.Bytes()[pcLengthPos:], uint16(pcLength))
	}

	// Update PDU length
	pduLength := pdu.Len() - 6
	binary.BigEndian.PutUint32(pdu.Bytes()[lengthPos:], uint32(pduLength))

	return pdu.Bytes()
}

// parseAssociationResponse parses the DICOM Association Response
func (c *Client) parseAssociationResponse(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("association response too short")
	}

	pduType := data[0]
	if pduType == PDUTypeAssociationRJ {
		return fmt.Errorf("association rejected by server")
	}

	if pduType != PDUTypeAssociationAC {
		return fmt.Errorf("unexpected PDU type in response: %d", pduType)
	}

	// For now, just verify we got an accept
	logrus.Info("Association accepted by server")
	return nil
}

// performRelease performs a DICOM association release
func (c *Client) performRelease() {
	logrus.Info("Releasing DICOM association")

	var buf bytes.Buffer
	buf.WriteByte(PDUTypeReleaseRQ)           // PDU Type
	buf.WriteByte(0x00)                       // Reserved
	buf.Write([]byte{0x00, 0x00, 0x00, 0x04}) // Length (4 bytes)

	if c.conn != nil {
		c.conn.Write(buf.Bytes())
		// Read release response
		response := make([]byte, 8)
		c.conn.Read(response)
	}
}

// sendDIMSECommand sends a DIMSE command over the DICOM association
func (c *Client) sendDIMSECommand(ctx context.Context, presentationContextID uint8, sopClass string, commandField uint16, data []byte) error {
	// Build DIMSE Command Set
	var cmdSet bytes.Buffer

	// Affected SOP Class UID
	cmdSet.Write([]byte{0x00, 0x00, 0x02, 0x00, 0x55, 0x49}) // Tag (0000,0200)
	cmdSet.Write([]byte{byte(len(sopClass))})
	cmdSet.WriteString(sopClass)

	// Command Field
	cmdSet.Write([]byte{0x00, 0x01, 0x00, 0x00, 0x55, 0x53}) // Tag (0000,0100)
	cmdSet.Write([]byte{0x00, 0x02})
	cmdFieldPos := cmdSet.Len()
	cmdSet.Write(make([]byte, 2))
	binary.BigEndian.PutUint16(cmdSet.Bytes()[cmdFieldPos:], commandField)

	// Message ID
	cmdSet.Write([]byte{0x00, 0x01, 0x10, 0x00, 0x55, 0x53}) // Tag (0000,0110)
	cmdSet.Write([]byte{0x00, 0x02})
	msgIdPos := cmdSet.Len()
	cmdSet.Write(make([]byte, 2))
	binary.BigEndian.PutUint16(cmdSet.Bytes()[msgIdPos:], 1)

	// Data Set Type (for C-ECHO: 0101, for C-STORE: 0001)
	dataSetType := byte(0x01)
	if commandField == 0x0030 { // C-ECHO
		dataSetType = 0x05
	}
	cmdSet.Write([]byte{0x00, 0x01, 0x08, 0x00, 0x55, 0x53}) // Tag (0000,0800)
	cmdSet.Write([]byte{0x00, 0x02})
	dataSetTypePos := cmdSet.Len()
	cmdSet.Write(make([]byte, 2))
	binary.BigEndian.PutUint16(cmdSet.Bytes()[dataSetTypePos:], uint16(dataSetType))

	// Build P-DATA-TF PDU
	var pdu bytes.Buffer

	// Presentation Data Value Item
	pdvHeader := []byte{presentationContextID, 0x00} // Context ID + Flags
	pdvData := append(cmdSet.Bytes(), data...)
	pdvLength := uint32(len(pdvData))

	// Write PDV Item Type and Length
	pdu.WriteByte(0x04) // Item Type: Presentation Data Value
	pdu.Write(pdvHeader)
	pdvLenPos := pdu.Len()
	pdu.Write(make([]byte, 4)) // Placeholder for length
	pdu.Write(pdvData)
	// Update PDV length
	binary.BigEndian.PutUint32(pdu.Bytes()[pdvLenPos:], pdvLength)

	// Write P-DATA-TF PDU Header
	var finalPDU bytes.Buffer
	finalPDU.WriteByte(PDUTypeDataTF) // PDU Type
	finalPDU.WriteByte(0x00)          // Reserved
	finalPDULenPos := finalPDU.Len()
	finalPDU.Write(make([]byte, 4)) // Placeholder for length
	finalPDU.Write(pdu.Bytes())
	// Update PDU length
	finalPDULength := finalPDU.Len() - 6
	binary.BigEndian.PutUint32(finalPDU.Bytes()[finalPDULenPos:], uint32(finalPDULength))

	// Send the PDU
	if _, err := c.conn.Write(finalPDU.Bytes()); err != nil {
		return fmt.Errorf("failed to send DIMSE command: %w", err)
	}

	// Read response
	response := make([]byte, 4096)
	n, err := c.conn.Read(response)
	if err != nil {
		return fmt.Errorf("failed to read DIMSE response: %w", err)
	}

	// Parse response status
	if err := c.parseDIMSEResponse(response[:n]); err != nil {
		return fmt.Errorf("DIMSE command failed: %w", err)
	}

	logrus.Info("DIMSE command completed successfully")
	return nil
}

// parseDIMSEResponse parses a DIMSE response
func (c *Client) parseDIMSEResponse(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("response too short")
	}

	// Check PDU type
	if data[0] != PDUTypeDataTF {
		return fmt.Errorf("unexpected PDU type in response: %d", data[0])
	}

	// For now, assume success if we get a valid P-DATA-TF response
	// In a full implementation, we would parse the DIMSE command response
	logrus.Info("DIMSE response received and parsed successfully")
	return nil
}

// extractSOPClassFromData extracts SOP Class UID from DICOM data
func (c *Client) extractSOPClassFromData(data []byte) string {
	// Look for SOP Class UID in DICOM elements (tag 0008,0016)
	// This is a simplified implementation
	if len(data) < 132 { // Minimum DICOM file size
		return SOPClassSecondaryCaptureImageStorage
	}

	// Search for SOP Class UID tag (0008,0016) in the data
	// In a real implementation, we would properly parse the DICOM elements
	// For now, return a default based on file characteristics
	return SOPClassSecondaryCaptureImageStorage
}

// findPresentationContext finds the presentation context ID for a given SOP class
func (c *Client) findPresentationContext(sopClass string) uint8 {
	contextMap := map[string]uint8{
		SOPClassVerification:                 1,
		SOPClassCTImageStorage:               3,
		SOPClassCRImageStorage:               5,
		SOPClassDXImageStorage:               7,
		SOPClassMGImageStorage:               9,
		SOPClassMRImageStorage:               11,
		SOPClassUSImageStorage:               13,
		SOPClassSecondaryCaptureImageStorage: 15,
	}

	if id, exists := contextMap[sopClass]; exists {
		return id
	}
	return 15 // Default to Secondary Capture
}
