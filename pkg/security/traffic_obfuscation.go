package security

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"sync"
	"time"
)

// ObfuscationProfile defines a traffic obfuscation profile
type ObfuscationProfile struct {
	Name                string
	Protocol            string
	HeaderTemplate      []byte
	FooterTemplate      []byte
	ChunkSize           int
	JitterEnabled       bool
	JitterMin           time.Duration
	JitterMax           time.Duration
	PaddingEnabled      bool
	PaddingMin          int
	PaddingMax          int
	ProtocolFingerprints map[string][]byte
	TransformationRules  []TransformationRule
}

// TransformationRule defines a rule for transforming traffic
type TransformationRule struct {
	Pattern     []byte
	Replacement []byte
	Position    string // "header", "body", "footer", "all"
}

// TrafficObfuscator implements traffic obfuscation techniques
type TrafficObfuscator struct {
	profiles    map[string]*ObfuscationProfile
	activeProfile string
	mutex       sync.RWMutex
	jitterRand  *rand.Rand
	paddingRand *rand.Rand
}

// NewTrafficObfuscator creates a new traffic obfuscator
func NewTrafficObfuscator() *TrafficObfuscator {
	return &TrafficObfuscator{
		profiles:    make(map[string]*ObfuscationProfile),
		activeProfile: "",
		jitterRand:  rand.New(rand.NewSource(time.Now().UnixNano())),
		paddingRand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RegisterProfile registers an obfuscation profile
func (t *TrafficObfuscator) RegisterProfile(profile *ObfuscationProfile) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.profiles[profile.Name] = profile

	// Set as active profile if none is set
	if t.activeProfile == "" {
		t.activeProfile = profile.Name
	}
}

// SetActiveProfile sets the active obfuscation profile
func (t *TrafficObfuscator) SetActiveProfile(name string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, exists := t.profiles[name]; !exists {
		return fmt.Errorf("profile not found: %s", name)
	}

	t.activeProfile = name
	return nil
}

// GetActiveProfile returns the active obfuscation profile
func (t *TrafficObfuscator) GetActiveProfile() *ObfuscationProfile {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.profiles[t.activeProfile]
}

// ObfuscateOutgoing obfuscates outgoing traffic
func (t *TrafficObfuscator) ObfuscateOutgoing(data []byte) ([]byte, error) {
	t.mutex.RLock()
	profile, exists := t.profiles[t.activeProfile]
	t.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no active profile")
	}

	// Apply transformations
	transformed := t.applyTransformations(data, profile)

	// Add protocol fingerprints
	fingerprinted := t.addProtocolFingerprints(transformed, profile)

	// Add header and footer
	var result bytes.Buffer
	result.Write(profile.HeaderTemplate)
	result.Write(fingerprinted)
	result.Write(profile.FooterTemplate)

	// Add padding if enabled
	if profile.PaddingEnabled {
		padded := t.addPadding(result.Bytes(), profile)
		return padded, nil
	}

	return result.Bytes(), nil
}

// DeobfuscateIncoming deobfuscates incoming traffic
func (t *TrafficObfuscator) DeobfuscateIncoming(data []byte) ([]byte, error) {
	t.mutex.RLock()
	profile, exists := t.profiles[t.activeProfile]
	t.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no active profile")
	}

	// Remove header and footer
	headerLen := len(profile.HeaderTemplate)
	footerLen := len(profile.FooterTemplate)

	if len(data) < headerLen+footerLen {
		return nil, fmt.Errorf("data too short")
	}

	// Extract the body
	body := data[headerLen : len(data)-footerLen]

	// Remove padding if enabled
	if profile.PaddingEnabled {
		unpadded, err := t.removePadding(body)
		if err != nil {
			return nil, err
		}
		body = unpadded
	}

	// Remove protocol fingerprints
	unfingerprinted := t.removeProtocolFingerprints(body, profile)

	// Reverse transformations
	untransformed := t.reverseTransformations(unfingerprinted, profile)

	return untransformed, nil
}

// ApplyJitter applies jitter to communication timing
func (t *TrafficObfuscator) ApplyJitter() time.Duration {
	t.mutex.RLock()
	profile, exists := t.profiles[t.activeProfile]
	t.mutex.RUnlock()

	if !exists || !profile.JitterEnabled {
		return 0
	}

	// Calculate random jitter within range
	jitterRange := profile.JitterMax - profile.JitterMin
	if jitterRange <= 0 {
		return profile.JitterMin
	}

	jitterNs := t.jitterRand.Int63n(int64(jitterRange))
	return profile.JitterMin + time.Duration(jitterNs)
}

// applyTransformations applies transformation rules to data
func (t *TrafficObfuscator) applyTransformations(data []byte, profile *ObfuscationProfile) []byte {
	result := data

	for _, rule := range profile.TransformationRules {
		if rule.Position == "body" || rule.Position == "all" {
			result = bytes.ReplaceAll(result, rule.Pattern, rule.Replacement)
		}
	}

	return result
}

// reverseTransformations reverses transformation rules
func (t *TrafficObfuscator) reverseTransformations(data []byte, profile *ObfuscationProfile) []byte {
	result := data

	// Apply transformations in reverse order
	for i := len(profile.TransformationRules) - 1; i >= 0; i-- {
		rule := profile.TransformationRules[i]
		if rule.Position == "body" || rule.Position == "all" {
			result = bytes.ReplaceAll(result, rule.Replacement, rule.Pattern)
		}
	}

	return result
}

// addProtocolFingerprints adds protocol fingerprints to data
func (t *TrafficObfuscator) addProtocolFingerprints(data []byte, profile *ObfuscationProfile) []byte {
	if len(profile.ProtocolFingerprints) == 0 {
		return data
	}

	// In a real implementation, this would add protocol-specific fingerprints
	// to make the traffic look like legitimate protocol traffic
	return data
}

// removeProtocolFingerprints removes protocol fingerprints from data
func (t *TrafficObfuscator) removeProtocolFingerprints(data []byte, profile *ObfuscationProfile) []byte {
	if len(profile.ProtocolFingerprints) == 0 {
		return data
	}

	// In a real implementation, this would remove protocol-specific fingerprints
	return data
}

// addPadding adds random padding to data
func (t *TrafficObfuscator) addPadding(data []byte, profile *ObfuscationProfile) []byte {
	if !profile.PaddingEnabled {
		return data
	}

	// Calculate random padding length
	paddingRange := profile.PaddingMax - profile.PaddingMin
	var paddingLen int
	if paddingRange <= 0 {
		paddingLen = profile.PaddingMin
	} else {
		paddingLen = profile.PaddingMin + t.paddingRand.Intn(paddingRange+1)
	}

	// Create padding
	padding := make([]byte, paddingLen+4) // +4 for length field
	if _, err := io.ReadFull(rand.Reader, padding[4:]); err != nil {
		// If random generation fails, use deterministic padding
		for i := 4; i < len(padding); i++ {
			padding[i] = byte(i % 256)
		}
	}

	// Store padding length at the beginning
	binary.LittleEndian.PutUint32(padding, uint32(paddingLen))

	// Append padding to data
	result := make([]byte, len(data)+len(padding))
	copy(result, data)
	copy(result[len(data):], padding)

	return result
}

// removePadding removes padding from data
func (t *TrafficObfuscator) removePadding(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short for padding")
	}

	// Extract padding length from the last 4 bytes
	paddingLen := int(binary.LittleEndian.Uint32(data[len(data)-4:]))
	
	// Validate padding length
	if paddingLen < 0 || paddingLen+4 > len(data) {
		return nil, fmt.Errorf("invalid padding length")
	}

	// Remove padding
	return data[:len(data)-(paddingLen+4)], nil
}

// CreateHTTPProfile creates an HTTP-like obfuscation profile
func CreateHTTPProfile() *ObfuscationProfile {
	return &ObfuscationProfile{
		Name:     "http",
		Protocol: "http",
		HeaderTemplate: []byte("GET /index.html HTTP/1.1\r\n" +
			"Host: example.com\r\n" +
			"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36\r\n" +
			"Accept: */*\r\n" +
			"Connection: keep-alive\r\n\r\n"),
		FooterTemplate: []byte("\r\n\r\n"),
		ChunkSize:      1024,
		JitterEnabled:  true,
		JitterMin:      100 * time.Millisecond,
		JitterMax:      500 * time.Millisecond,
		PaddingEnabled: true,
		PaddingMin:     16,
		PaddingMax:     64,
		ProtocolFingerprints: map[string][]byte{
			"http_get":    []byte("GET"),
			"http_post":   []byte("POST"),
			"http_cookie": []byte("Cookie:"),
		},
		TransformationRules: []TransformationRule{
			{
				Pattern:     []byte{0x00, 0x01, 0x02, 0x03},
				Replacement: []byte("Content-Length: 1234\r\n"),
				Position:    "header",
			},
		},
	}
}

// CreateDNSProfile creates a DNS-like obfuscation profile
func CreateDNSProfile() *ObfuscationProfile {
	return &ObfuscationProfile{
		Name:     "dns",
		Protocol: "dns",
		HeaderTemplate: []byte{
			0x00, 0x01, // Transaction ID
			0x01, 0x00, // Flags
			0x00, 0x01, // Questions
			0x00, 0x00, // Answer RRs
			0x00, 0x00, // Authority RRs
			0x00, 0x00, // Additional RRs
		},
		FooterTemplate: []byte{
			0x00, 0x00, // Type
			0x00, 0x01, // Class
		},
		ChunkSize:      63, // DNS label max length
		JitterEnabled:  true,
		JitterMin:      500 * time.Millisecond,
		JitterMax:      2 * time.Second,
		PaddingEnabled: true,
		PaddingMin:     4,
		PaddingMax:     16,
		ProtocolFingerprints: map[string][]byte{
			"dns_query": []byte{0x01, 0x00}, // Standard query
		},
		TransformationRules: []TransformationRule{
			{
				Pattern:     []byte{0xFF, 0xFE, 0xFD, 0xFC},
				Replacement: []byte{0x03, 'w', 'w', 'w'},
				Position:    "body",
			},
		},
	}
}

// CreateTLSProfile creates a TLS-like obfuscation profile
func CreateTLSProfile() *ObfuscationProfile {
	return &ObfuscationProfile{
		Name:     "tls",
		Protocol: "tls",
		HeaderTemplate: []byte{
			0x16,       // Content type: Handshake
			0x03, 0x01, // TLS version: TLS 1.0
			0x00, 0x00, // Length (placeholder)
		},
		FooterTemplate: []byte{
			0x14,       // Content type: Change Cipher Spec
			0x03, 0x01, // TLS version: TLS 1.0
			0x00, 0x01, // Length
			0x01,       // Change cipher spec message
		},
		ChunkSize:      1024,
		JitterEnabled:  true,
		JitterMin:      50 * time.Millisecond,
		JitterMax:      200 * time.Millisecond,
		PaddingEnabled: true,
		PaddingMin:     32,
		PaddingMax:     128,
		ProtocolFingerprints: map[string][]byte{
			"tls_client_hello": []byte{0x16, 0x03, 0x01},
		},
		TransformationRules: []TransformationRule{
			{
				Pattern:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
				Replacement: []byte{0x01, 0x00, 0x00, 0x00}, // Client Hello
				Position:    "header",
			},
		},
	}
}
