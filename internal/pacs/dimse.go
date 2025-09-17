package pacs

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// DIMSECommand represents a DIMSE command
type DIMSECommand struct {
	CommandField     uint16
	MessageID        uint16
	AffectedSOPClass string
	DataSetType      uint16
	Data             []byte
}

// sendCEcho sends a C-ECHO request using minimal DIMSE format
func (c *Client) sendCEcho(ctx context.Context) error {
	logrus.Info("Sending simplified C-ECHO request")

	// Build minimal C-ECHO command (Implicit VR Little Endian)
	var cmd bytes.Buffer

	// Command Group Length (0000,0000) - will calculate at end
	cmd.Write([]byte{0x00, 0x00, 0x00, 0x00}) // Tag
	groupLengthPos := cmd.Len()
	cmd.Write([]byte{0x00, 0x00, 0x00, 0x00}) // Length placeholder

	// Affected SOP Class UID (0000,0002)
	cmd.Write([]byte{0x00, 0x00, 0x00, 0x02}) // Tag
	sopClass := SOPClassVerification
	sopLen := len(sopClass)
	if sopLen%2 == 1 {
		sopClass += "\x00" // Pad to even length
		sopLen++
	}
	cmd.Write([]byte{byte(sopLen >> 8), byte(sopLen)}) // Length
	cmd.WriteString(sopClass)

	// Command Field (0000,0100)
	cmd.Write([]byte{0x00, 0x00, 0x01, 0x00}) // Tag
	cmd.Write([]byte{0x00, 0x02})             // Length
	cmd.Write([]byte{0x00, 0x30})             // C-ECHO-RQ

	// Message ID (0000,0110)
	cmd.Write([]byte{0x00, 0x00, 0x01, 0x10}) // Tag
	cmd.Write([]byte{0x00, 0x02})             // Length
	cmd.Write([]byte{0x00, 0x01})             // Message ID = 1

	// Data Set Type (0000,0800)
	cmd.Write([]byte{0x00, 0x00, 0x08, 0x00}) // Tag
	cmd.Write([]byte{0x00, 0x02})             // Length
	cmd.Write([]byte{0x01, 0x01})             // No Data Set Present

	// Update Command Group Length
	groupLength := uint32(cmd.Len() - 8) // Exclude group length element
	binary.BigEndian.PutUint32(cmd.Bytes()[groupLengthPos:], groupLength)

	// Send as P-DATA-TF PDU
	return c.sendPDataTF(1, 0x03, cmd.Bytes()) // PC ID=1, Command+Last Fragment
}

// sendPDataTF sends a P-DATA-TF PDU with the given data
func (c *Client) sendPDataTF(presentationContextID uint8, messageControlHeader uint8, data []byte) error {
	var pdu bytes.Buffer

	// P-DATA-TF PDU Header
	pdu.WriteByte(PDUTypeDataTF) // PDU Type
	pdu.WriteByte(0x00)          // Reserved

	// Calculate lengths
	pdvLength := uint32(2 + len(data)) // PDV header + data
	pduLength := 4 + pdvLength         // PDV length field + PDV

	// PDU Length
	pduLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(pduLengthBytes, pduLength)
	pdu.Write(pduLengthBytes)

	// PDV Item Length
	pdvLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(pdvLengthBytes, pdvLength)
	pdu.Write(pdvLengthBytes)

	// PDV Header
	pdu.WriteByte(presentationContextID) // Presentation Context ID
	pdu.WriteByte(messageControlHeader)  // Message Control Header

	// PDV Data
	pdu.Write(data)

	logrus.Debugf("Sending P-DATA-TF: %d bytes, PC=%d, MCH=0x%02X", pdu.Len(), presentationContextID, messageControlHeader)

	// Send PDU
	if _, err := c.conn.Write(pdu.Bytes()); err != nil {
		return fmt.Errorf("failed to send P-DATA-TF: %w", err)
	}

	// Read response with timeout
	c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	response := make([]byte, 4096)
	n, err := c.conn.Read(response)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	logrus.Debugf("Received response: %d bytes, PDU type: 0x%02X", n, response[0])

	// Check response type
	if response[0] == PDUTypeDataTF {
		logrus.Info("Received P-DATA-TF response")
		return nil
	} else if response[0] == PDUTypeAbortRQ {
		return fmt.Errorf("received A-ABORT from PACS")
	} else {
		return fmt.Errorf("unexpected response PDU type: 0x%02X", response[0])
	}
}

// sendCStore sends a C-STORE request using minimal DIMSE format
func (c *Client) sendCStore(ctx context.Context, dicomData []byte, sopInstanceUID string, sopClassUID string) error {
	logrus.Infof("Sending C-STORE for SOP Instance: %s", sopInstanceUID)

	// Find appropriate presentation context
	contextID := c.findPresentationContext(sopClassUID)
	if contextID == 0 {
		contextID = 15 // Fallback to Secondary Capture
		logrus.Warnf("Using fallback presentation context %d for SOP class %s", contextID, sopClassUID)
	}

	// Build C-STORE command
	var cmd bytes.Buffer

	// Command Group Length (0000,0000)
	cmd.Write([]byte{0x00, 0x00, 0x00, 0x00}) // Tag
	groupLengthPos := cmd.Len()
	cmd.Write([]byte{0x00, 0x00, 0x00, 0x00}) // Length placeholder

	// Affected SOP Class UID (0000,0002)
	cmd.Write([]byte{0x00, 0x00, 0x00, 0x02}) // Tag
	sopLen := len(sopClassUID)
	if sopLen%2 == 1 {
		sopClassUID += "\x00" // Pad to even length
		sopLen++
	}
	cmd.Write([]byte{byte(sopLen >> 8), byte(sopLen)}) // Length
	cmd.WriteString(sopClassUID)

	// Command Field (0000,0100)
	cmd.Write([]byte{0x00, 0x00, 0x01, 0x00}) // Tag
	cmd.Write([]byte{0x00, 0x02})             // Length
	cmd.Write([]byte{0x00, 0x01})             // C-STORE-RQ

	// Message ID (0000,0110)
	cmd.Write([]byte{0x00, 0x00, 0x01, 0x10}) // Tag
	cmd.Write([]byte{0x00, 0x02})             // Length
	cmd.Write([]byte{0x00, 0x01})             // Message ID = 1

	// Affected SOP Instance UID (0000,1000)
	cmd.Write([]byte{0x00, 0x00, 0x10, 0x00}) // Tag
	sopInstLen := len(sopInstanceUID)
	if sopInstLen%2 == 1 {
		sopInstanceUID += "\x00" // Pad to even length
		sopInstLen++
	}
	cmd.Write([]byte{byte(sopInstLen >> 8), byte(sopInstLen)}) // Length
	cmd.WriteString(sopInstanceUID)

	// Data Set Type (0000,0800)
	cmd.Write([]byte{0x00, 0x00, 0x08, 0x00}) // Tag
	cmd.Write([]byte{0x00, 0x02})             // Length
	cmd.Write([]byte{0x00, 0x01})             // Data Set Present

	// Update Command Group Length
	groupLength := uint32(cmd.Len() - 8)
	binary.BigEndian.PutUint32(cmd.Bytes()[groupLengthPos:], groupLength)

	// Send command first
	if err := c.sendPDataTF(contextID, 0x01, cmd.Bytes()); err != nil { // Command, not last
		return fmt.Errorf("failed to send C-STORE command: %w", err)
	}

	// Send dataset
	if err := c.sendPDataTF(contextID, 0x02, dicomData); err != nil { // Dataset, last fragment
		return fmt.Errorf("failed to send C-STORE dataset: %w", err)
	}

	logrus.Info("C-STORE completed successfully")
	return nil
}
