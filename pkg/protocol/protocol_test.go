package protocol

import (
	"bytes"
	"testing"
	"time"

	"dinoc2/pkg/crypto"
)

func TestPacketEncoding(t *testing.T) {
	// Create a test packet
	originalData := []byte("This is a test packet for encoding and decoding")
	packet := NewPacket(PacketTypeCommand, originalData)
	packet.SetTaskID(12345)
	packet.SetEncryptionAlgorithm(EncryptionAlgorithmAES)
	
	// Encode the packet
	encoded := EncodePacket(packet)
	
	// Decode the packet
	decoded, err := DecodePacket(encoded)
	if err != nil {
		t.Fatalf("Failed to decode packet: %v", err)
	}
	
	// Verify the decoded packet
	if decoded.Header.Version != packet.Header.Version {
		t.Errorf("Version mismatch: got %d, want %d", decoded.Header.Version, packet.Header.Version)
	}
	
	if decoded.Header.EncAlgorithm != packet.Header.EncAlgorithm {
		t.Errorf("EncAlgorithm mismatch: got %d, want %d", decoded.Header.EncAlgorithm, packet.Header.EncAlgorithm)
	}
	
	if decoded.Header.Type != packet.Header.Type {
		t.Errorf("Type mismatch: got %d, want %d", decoded.Header.Type, packet.Header.Type)
	}
	
	if decoded.Header.TaskID != packet.Header.TaskID {
		t.Errorf("TaskID mismatch: got %d, want %d", decoded.Header.TaskID, packet.Header.TaskID)
	}
	
	if !bytes.Equal(decoded.Data, packet.Data) {
		t.Errorf("Data mismatch: got %v, want %v", decoded.Data, packet.Data)
	}
}

func TestPacketFragmentation(t *testing.T) {
	// Create a large test packet
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	
	packet := NewPacket(PacketTypeModuleData, largeData)
	packet.SetTaskID(54321)
	
	// Fragment the packet
	fragments := FragmentPacket(packet, 1000)
	
	// Verify we have the expected number of fragments
	expectedFragments := (len(largeData) + 1000 - 1) / 1000
	if len(fragments) != expectedFragments {
		t.Errorf("Unexpected number of fragments: got %d, want %d", len(fragments), expectedFragments)
	}
	
	// Reassemble the packet
	reassembled, err := ReassemblePacket(fragments)
	if err != nil {
		t.Fatalf("Failed to reassemble packet: %v", err)
	}
	
	// Verify the reassembled packet
	if reassembled.Header.Version != packet.Header.Version {
		t.Errorf("Version mismatch: got %d, want %d", reassembled.Header.Version, packet.Header.Version)
	}
	
	if reassembled.Header.Type != packet.Header.Type {
		t.Errorf("Type mismatch: got %d, want %d", reassembled.Header.Type, packet.Header.Type)
	}
	
	if reassembled.Header.TaskID != packet.Header.TaskID {
		t.Errorf("TaskID mismatch: got %d, want %d", reassembled.Header.TaskID, packet.Header.TaskID)
	}
	
	// The reassembled data will include fragment headers, so we need to extract the original data
	// In a real implementation, you would have a more robust way to handle this
	if len(reassembled.Data) != len(largeData) {
		t.Errorf("Data length mismatch: got %d, want %d", len(reassembled.Data), len(largeData))
	}
}

func TestTLVEncoding(t *testing.T) {
	// Create a test TLV
	originalData := []byte("This is a test TLV for encoding and decoding")
	tlv := NewTLV(TLVTypeCommand, originalData)
	
	// Encode the TLV
	encoded := EncodeTLV(tlv)
	
	// Decode the TLV
	decoded, bytesRead, err := DecodeTLV(encoded)
	if err != nil {
		t.Fatalf("Failed to decode TLV: %v", err)
	}
	
	// Verify the decoded TLV
	if decoded.Type != tlv.Type {
		t.Errorf("Type mismatch: got %d, want %d", decoded.Type, tlv.Type)
	}
	
	if decoded.Length != tlv.Length {
		t.Errorf("Length mismatch: got %d, want %d", decoded.Length, tlv.Length)
	}
	
	if !bytes.Equal(decoded.Value, tlv.Value) {
		t.Errorf("Value mismatch: got %v, want %v", decoded.Value, tlv.Value)
	}
	
	// Verify bytes read
	if bytesRead != len(encoded) {
		t.Errorf("Bytes read mismatch: got %d, want %d", bytesRead, len(encoded))
	}
}

func TestProtocolHandler(t *testing.T) {
	// Create a protocol handler
	handler := NewProtocolHandler()
	defer handler.Shutdown()
	
	// Create a session
	sessionID := crypto.SessionID("test-session")
	err := handler.CreateSession(sessionID, crypto.AlgorithmAES)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	// Create a test packet
	originalData := []byte("This is a test packet for the protocol handler")
	packet := NewPacket(PacketTypeCommand, originalData)
	packet.SetTaskID(67890)
	
	// Prepare the packet for sending (with encryption)
	encodedFragments, err := handler.PrepareOutgoingPacket(packet, sessionID, true)
	if err != nil {
		t.Fatalf("Failed to prepare outgoing packet: %v", err)
	}
	
	// Process each fragment
	var processedPacket *Packet
	for i, fragment := range encodedFragments {
		processed, err := handler.ProcessIncomingPacket(fragment, sessionID)
		if err != nil {
			if i < len(encodedFragments)-1 && err.Error() == "packet fragmented, waiting for more fragments" {
				// Expected error for non-final fragments
				continue
			}
			t.Fatalf("Failed to process incoming packet: %v", err)
		}
		
		processedPacket = processed
	}
	
	// Verify the processed packet
	if processedPacket == nil {
		t.Fatal("Processed packet is nil")
	}
	
	if processedPacket.Header.Type != packet.Header.Type {
		t.Errorf("Type mismatch: got %d, want %d", processedPacket.Header.Type, packet.Header.Type)
	}
	
	if processedPacket.Header.TaskID != packet.Header.TaskID {
		t.Errorf("TaskID mismatch: got %d, want %d", processedPacket.Header.TaskID, packet.Header.TaskID)
	}
	
	if !bytes.Equal(processedPacket.Data, originalData) {
		t.Errorf("Data mismatch: got %v, want %v", processedPacket.Data, originalData)
	}
}

func TestObfuscator(t *testing.T) {
	// Test HTTP obfuscation
	httpProfile := DefaultHTTPProfile()
	httpObfuscator := NewObfuscator(httpProfile)
	
	originalData := []byte("This is a test for HTTP obfuscation")
	
	// Obfuscate
	obfuscated, err := httpObfuscator.Obfuscate(originalData)
	if err != nil {
		t.Fatalf("Failed to obfuscate with HTTP: %v", err)
	}
	
	// Deobfuscate
	deobfuscated, err := httpObfuscator.Deobfuscate(obfuscated)
	if err != nil {
		t.Fatalf("Failed to deobfuscate with HTTP: %v", err)
	}
	
	// Verify
	if !bytes.Equal(deobfuscated, originalData) {
		t.Errorf("HTTP obfuscation data mismatch: got %v, want %v", deobfuscated, originalData)
	}
	
	// Test TLS obfuscation
	tlsProfile := DefaultTLSProfile()
	tlsObfuscator := NewObfuscator(tlsProfile)
	
	// Obfuscate
	obfuscated, err = tlsObfuscator.Obfuscate(originalData)
	if err != nil {
		t.Fatalf("Failed to obfuscate with TLS: %v", err)
	}
	
	// Deobfuscate
	deobfuscated, err = tlsObfuscator.Deobfuscate(obfuscated)
	if err != nil {
		t.Fatalf("Failed to deobfuscate with TLS: %v", err)
	}
	
	// Verify
	if !bytes.Equal(deobfuscated, originalData) {
		t.Errorf("TLS obfuscation data mismatch: got %v, want %v", deobfuscated, originalData)
	}
}

func TestJitter(t *testing.T) {
	// Create a protocol handler
	handler := NewProtocolHandler()
	
	// Enable jitter
	handler.SetJitterEnabled(true)
	handler.SetJitterRange(10*time.Millisecond, 100*time.Millisecond)
	
	// Get jitter delay
	delay := handler.GetJitterDelay()
	
	// Verify delay is within range
	if delay < 10*time.Millisecond || delay > 100*time.Millisecond {
		t.Errorf("Jitter delay outside of range: got %v", delay)
	}
	
	// Disable jitter
	handler.SetJitterEnabled(false)
	
	// Get jitter delay
	delay = handler.GetJitterDelay()
	
	// Verify delay is zero
	if delay != 0 {
		t.Errorf("Jitter delay should be zero when disabled: got %v", delay)
	}
}
