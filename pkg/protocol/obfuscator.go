package protocol

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"strings"
)

// ObfuscationType represents the type of traffic obfuscation
type ObfuscationType int

const (
	ObfuscationNone ObfuscationType = iota
	ObfuscationHTTP
	ObfuscationDNS
	ObfuscationTLS
	ObfuscationCustom
)

// ObfuscationProfile contains settings for traffic obfuscation
type ObfuscationProfile struct {
	Type            ObfuscationType
	PaddingEnabled  bool
	PaddingMinBytes int
	PaddingMaxBytes int
	JitterEnabled   bool
	JitterMinMS     int
	JitterMaxMS     int
	CustomHeaders   map[string]string
	MimicryTemplate []byte
}

// DefaultHTTPProfile returns a default HTTP obfuscation profile
func DefaultHTTPProfile() *ObfuscationProfile {
	return &ObfuscationProfile{
		Type:            ObfuscationHTTP,
		PaddingEnabled:  true,
		PaddingMinBytes: 10,
		PaddingMaxBytes: 100,
		JitterEnabled:   true,
		JitterMinMS:     10,
		JitterMaxMS:     100,
		CustomHeaders: map[string]string{
			"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			"Accept-Language": "en-US,en;q=0.5",
			"Accept-Encoding": "gzip, deflate, br",
			"Connection":      "keep-alive",
			"Cache-Control":   "max-age=0",
		},
	}
}

// DefaultTLSProfile returns a default TLS obfuscation profile
func DefaultTLSProfile() *ObfuscationProfile {
	return &ObfuscationProfile{
		Type:            ObfuscationTLS,
		PaddingEnabled:  true,
		PaddingMinBytes: 20,
		PaddingMaxBytes: 200,
		JitterEnabled:   true,
		JitterMinMS:     15,
		JitterMaxMS:     150,
	}
}

// Obfuscator handles traffic obfuscation
type Obfuscator struct {
	profile *ObfuscationProfile
}

// NewObfuscator creates a new traffic obfuscator with the specified profile
func NewObfuscator(profile *ObfuscationProfile) *Obfuscator {
	if profile == nil {
		profile = &ObfuscationProfile{
			Type: ObfuscationNone,
		}
	}
	
	return &Obfuscator{
		profile: profile,
	}
}

// Obfuscate applies obfuscation to the data
func (o *Obfuscator) Obfuscate(data []byte) ([]byte, error) {
	switch o.profile.Type {
	case ObfuscationNone:
		return data, nil
	case ObfuscationHTTP:
		return o.obfuscateHTTP(data)
	case ObfuscationDNS:
		return o.obfuscateDNS(data)
	case ObfuscationTLS:
		return o.obfuscateTLS(data)
	case ObfuscationCustom:
		return o.obfuscateCustom(data)
	default:
		return data, nil
	}
}

// Deobfuscate removes obfuscation from the data
func (o *Obfuscator) Deobfuscate(data []byte) ([]byte, error) {
	switch o.profile.Type {
	case ObfuscationNone:
		return data, nil
	case ObfuscationHTTP:
		return o.deobfuscateHTTP(data)
	case ObfuscationDNS:
		return o.deobfuscateDNS(data)
	case ObfuscationTLS:
		return o.deobfuscateTLS(data)
	case ObfuscationCustom:
		return o.deobfuscateCustom(data)
	default:
		return data, nil
	}
}

// obfuscateHTTP applies HTTP obfuscation
func (o *Obfuscator) obfuscateHTTP(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	
	// Encode data as base64
	encodedData := base64.StdEncoding.EncodeToString(data)
	
	// Create HTTP request
	buf.WriteString("POST /api/data HTTP/1.1\r\n")
	buf.WriteString("Host: example.com\r\n")
	
	// Add custom headers
	for key, value := range o.profile.CustomHeaders {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	
	// Add content headers
	buf.WriteString("Content-Type: application/json\r\n")
	buf.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(encodedData)+4)) // +4 for the JSON quotes and braces
	buf.WriteString("\r\n")
	
	// Add data as JSON
	buf.WriteString(fmt.Sprintf(`{"d":"%s"}`, encodedData))
	
	// Add padding if enabled
	if o.profile.PaddingEnabled {
		paddingSize := o.profile.PaddingMinBytes
		if o.profile.PaddingMaxBytes > o.profile.PaddingMinBytes {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(o.profile.PaddingMaxBytes-o.profile.PaddingMinBytes+1)))
			if err != nil {
				return nil, err
			}
			paddingSize += int(n.Int64())
		}
		
		padding := make([]byte, paddingSize)
		if _, err := io.ReadFull(rand.Reader, padding); err != nil {
			return nil, err
		}
		
		// Add padding as a comment
		buf.WriteString(fmt.Sprintf("\r\n<!-- %s -->", base64.StdEncoding.EncodeToString(padding)))
	}
	
	return buf.Bytes(), nil
}

// deobfuscateHTTP removes HTTP obfuscation
func (o *Obfuscator) deobfuscateHTTP(data []byte) ([]byte, error) {
	// Find the JSON data
	start := bytes.Index(data, []byte(`{"d":"`))
	if start == -1 {
		return nil, fmt.Errorf("invalid HTTP obfuscated data: missing JSON")
	}
	
	start += 6 // Skip {"d":"
	
	// Find the end of the base64 data
	end := bytes.Index(data[start:], []byte(`"`))
	if end == -1 {
		return nil, fmt.Errorf("invalid HTTP obfuscated data: missing end quote")
	}
	
	// Extract the base64 data
	encodedData := data[start : start+end]
	
	// Decode base64
	decodedData, err := base64.StdEncoding.DecodeString(string(encodedData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}
	
	return decodedData, nil
}

// obfuscateDNS applies DNS obfuscation
func (o *Obfuscator) obfuscateDNS(data []byte) ([]byte, error) {
	// Encode data as base32 (DNS-friendly encoding)
	encodedData := base32.StdEncoding.EncodeToString(data)
	
	// Split into DNS-like segments (max 63 chars per label)
	var segments []string
	for i := 0; i < len(encodedData); i += 63 {
		end := i + 63
		if end > len(encodedData) {
			end = len(encodedData)
		}
		segments = append(segments, encodedData[i:end])
	}
	
	// Create DNS query format
	domain := strings.Join(segments, ".")
	domain += ".example.com"
	
	// Create DNS query packet
	var buf bytes.Buffer
	
	// Transaction ID (2 bytes)
	transactionID := make([]byte, 2)
	if _, err := io.ReadFull(rand.Reader, transactionID); err != nil {
		return nil, err
	}
	buf.Write(transactionID)
	
	// Flags (2 bytes) - Standard query
	buf.Write([]byte{0x01, 0x00})
	
	// Questions (2 bytes) - 1 question
	buf.Write([]byte{0x00, 0x01})
	
	// Answer RRs, Authority RRs, Additional RRs (6 bytes) - All 0
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	
	// Write domain name
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		buf.WriteByte(byte(len(label)))
		buf.WriteString(label)
	}
	
	// Terminating byte for domain name
	buf.WriteByte(0x00)
	
	// Type (2 bytes) - A record
	buf.Write([]byte{0x00, 0x01})
	
	// Class (2 bytes) - IN
	buf.Write([]byte{0x00, 0x01})
	
	return buf.Bytes(), nil
}

// deobfuscateDNS removes DNS obfuscation
func (o *Obfuscator) deobfuscateDNS(data []byte) ([]byte, error) {
	// Skip DNS header (12 bytes)
	if len(data) < 12 {
		return nil, fmt.Errorf("invalid DNS obfuscated data: too short")
	}
	
	// Extract domain name
	var domain strings.Builder
	pos := 12 // Start after header
	
	for {
		if pos >= len(data) {
			return nil, fmt.Errorf("invalid DNS obfuscated data: unexpected end")
		}
		
		labelLen := int(data[pos])
		pos++
		
		if labelLen == 0 {
			break // End of domain name
		}
		
		if pos+labelLen > len(data) {
			return nil, fmt.Errorf("invalid DNS obfuscated data: label too long")
		}
		
		if domain.Len() > 0 {
			domain.WriteByte('.')
		}
		
		domain.Write(data[pos : pos+labelLen])
		pos += labelLen
	}
	
	// Remove the example.com suffix
	domainStr := domain.String()
	if !strings.HasSuffix(domainStr, ".example.com") {
		return nil, fmt.Errorf("invalid DNS obfuscated data: missing domain suffix")
	}
	
	domainStr = strings.TrimSuffix(domainStr, ".example.com")
	
	// Join segments and decode base32
	encodedData := strings.ReplaceAll(domainStr, ".", "")
	
	// Decode base32
	decodedData, err := base32.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base32 data: %w", err)
	}
	
	return decodedData, nil
}

// generateRandomInt generates a random integer between min and max (inclusive)
func generateRandomInt(min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}
	
	if min == max {
		return min, nil
	}
	
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	
	return min + int(n.Int64()), nil
}

// generateRandomInt64 is a local helper function
func generateRandomInt64(max int64) int64 {
	if max <= 0 {
		return 0
	}
	
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0
	}
	
	return n.Int64()
}

// obfuscateTLS applies TLS obfuscation
func (o *Obfuscator) obfuscateTLS(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	
	// Encode data as base64
	encodedData := base64.StdEncoding.EncodeToString(data)
	
	// Create TLS-like header
	// Record type: Application Data (23)
	buf.WriteByte(23)
	
	// TLS version: 1.2 (0x0303)
	buf.Write([]byte{0x03, 0x03})
	
	// Length (2 bytes) - will be filled in later
	lengthPos := buf.Len()
	buf.Write([]byte{0x00, 0x00})
	
	// Add encoded data
	dataStart := buf.Len()
	buf.WriteString(encodedData)
	dataLength := buf.Len() - dataStart
	
	// Add padding if enabled
	if o.profile.PaddingEnabled {
		paddingSize, err := generateRandomInt(o.profile.PaddingMinBytes, o.profile.PaddingMaxBytes)
		if err != nil {
			return nil, err
		}
		
		padding := make([]byte, paddingSize)
		if _, err := io.ReadFull(rand.Reader, padding); err != nil {
			return nil, err
		}
		
		buf.Write(padding)
		dataLength += paddingSize
	}
	
	// Update length field
	result := buf.Bytes()
	binary.BigEndian.PutUint16(result[lengthPos:lengthPos+2], uint16(dataLength))
	
	return result, nil
}

// deobfuscateTLS removes TLS obfuscation
func (o *Obfuscator) deobfuscateTLS(data []byte) ([]byte, error) {
	// Verify TLS header
	if len(data) < 5 {
		return nil, fmt.Errorf("invalid TLS obfuscated data: too short")
	}
	
	// Check record type (23 = Application Data)
	if data[0] != 23 {
		return nil, fmt.Errorf("invalid TLS obfuscated data: wrong record type")
	}
	
	// Check TLS version (0x0303 = TLS 1.2)
	if data[1] != 0x03 || data[2] != 0x03 {
		return nil, fmt.Errorf("invalid TLS obfuscated data: wrong TLS version")
	}
	
	// Get data length
	dataLength := int(binary.BigEndian.Uint16(data[3:5]))
	
	// Verify data length
	if len(data) < 5+dataLength {
		return nil, fmt.Errorf("invalid TLS obfuscated data: insufficient data")
	}
	
	// Extract encoded data (base64)
	encodedData := string(data[5:])
	
	// Find the end of the base64 data (it's followed by random padding)
	// Base64 is always a multiple of 4 characters and ends with 0-2 '=' characters
	validLen := 0
	for i := 0; i <= len(encodedData); i += 4 {
		if i > len(encodedData) {
			break
		}
		
		// Try to decode this segment
		segment := encodedData[:i]
		if _, err := base64.StdEncoding.DecodeString(segment); err == nil {
			validLen = i
		}
	}
	
	if validLen == 0 {
		return nil, fmt.Errorf("invalid TLS obfuscated data: no valid base64 data found")
	}
	
	// Decode base64
	decodedData, err := base64.StdEncoding.DecodeString(encodedData[:validLen])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}
	
	return decodedData, nil
}

// obfuscateCustom applies custom obfuscation
func (o *Obfuscator) obfuscateCustom(data []byte) ([]byte, error) {
	if o.profile.MimicryTemplate == nil {
		return nil, fmt.Errorf("custom obfuscation requires a mimicry template")
	}
	
	// Create a copy of the template
	result := make([]byte, len(o.profile.MimicryTemplate))
	copy(result, o.profile.MimicryTemplate)
	
	// Encode data as base64
	encodedData := base64.StdEncoding.EncodeToString(data)
	
	// Find placeholder in template (if any)
	placeholder := []byte("{{DATA}}")
	placeholderPos := bytes.Index(result, placeholder)
	
	if placeholderPos >= 0 {
		// Replace placeholder with encoded data
		before := result[:placeholderPos]
		after := result[placeholderPos+len(placeholder):]
		
		var buf bytes.Buffer
		buf.Write(before)
		buf.WriteString(encodedData)
		buf.Write(after)
		
		result = buf.Bytes()
	} else {
		// No placeholder, append data to the end
		result = append(result, []byte(encodedData)...)
	}
	
	return result, nil
}

// deobfuscateCustom removes custom obfuscation
func (o *Obfuscator) deobfuscateCustom(data []byte) ([]byte, error) {
	// Find base64 encoded data
	// This is a simplified approach - in a real implementation, you would need
	// a more robust way to identify the encoded data within the custom template
	
	// Look for base64 pattern (alphanumeric + /+ and possibly ending with =)
	re := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=")
	
	// Find the longest continuous sequence of base64 characters
	start := -1
	end := -1
	currentStart := -1
	currentLength := 0
	maxLength := 0
	
	for i, b := range data {
		isBase64 := false
		for _, c := range re {
			if b == c {
				isBase64 = true
				break
			}
		}
		
		if isBase64 {
			if currentStart == -1 {
				currentStart = i
			}
			currentLength++
		} else {
			if currentStart != -1 {
				if currentLength > maxLength {
					maxLength = currentLength
					start = currentStart
					end = i
				}
				currentStart = -1
				currentLength = 0
			}
		}
	}
	
	// Check if we found a base64 sequence
	if start == -1 || maxLength < 4 { // Base64 is always at least 4 characters
		return nil, fmt.Errorf("no base64 data found in custom obfuscated data")
	}
	
	// Extract and decode the base64 data
	encodedData := string(data[start:end])
	
	// Ensure the length is a multiple of 4
	padding := 4 - (len(encodedData) % 4)
	if padding < 4 {
		encodedData += strings.Repeat("=", padding)
	}
	
	// Decode base64
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}
	
	return decodedData, nil
}
