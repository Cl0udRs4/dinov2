package listener

import (
	"errors"
	"fmt"

	"dinoc2/pkg/listener/dns"
	"dinoc2/pkg/listener/http"
	"dinoc2/pkg/listener/icmp"
	"dinoc2/pkg/listener/websocket"
)

// ListenerType represents the type of listener
type ListenerType string

const (
	ListenerTypeTCP       ListenerType = "tcp"
	ListenerTypeDNS       ListenerType = "dns"
	ListenerTypeICMP      ListenerType = "icmp"
	ListenerTypeHTTP      ListenerType = "http"
	ListenerTypeWebSocket ListenerType = "websocket"
)

// CreateListener creates a new listener of the specified type
func CreateListener(listenerType ListenerType, config ListenerConfig) (Listener, error) {
	switch listenerType {
	case ListenerTypeTCP:
		return NewTCPListener(config), nil
	case ListenerTypeDNS:
		// Convert generic config to DNS-specific config
		dnsConfig := dns.DNSConfig{
			Address: config.Address,
			Port:    config.Port,
		}
		
		// Extract DNS-specific options
		if config.Options != nil {
			if domain, ok := config.Options["domain"].(string); ok {
				dnsConfig.Domain = domain
			}
			if ttl, ok := config.Options["ttl"].(uint32); ok {
				dnsConfig.TTL = ttl
			}
		}
		
		dnsListener := dns.NewDNSListener(dnsConfig)
		return NewDNSListenerAdapter(dnsListener), nil
	case ListenerTypeICMP:
		// Convert generic config to ICMP-specific config
		icmpConfig := icmp.ICMPConfig{
			ListenAddress: config.Address,
		}
		
		// Extract ICMP-specific options
		if config.Options != nil {
			if protocol, ok := config.Options["protocol"].(string); ok {
				icmpConfig.Protocol = protocol
			}
		}
		
		icmpListener := icmp.NewICMPListener(icmpConfig)
		return NewICMPListenerAdapter(icmpListener), nil
	case ListenerTypeHTTP:
		// Convert generic config to HTTP-specific config
		httpConfig := http.HTTPConfig{
			Address: config.Address,
			Port:    config.Port,
		}
		
		// Extract HTTP-specific options
		if config.Options != nil {
			if certFile, ok := config.Options["tls_cert_file"].(string); ok {
				httpConfig.TLSCertFile = certFile
			}
			if keyFile, ok := config.Options["tls_key_file"].(string); ok {
				httpConfig.TLSKeyFile = keyFile
			}
			if useHTTP2, ok := config.Options["use_http2"].(bool); ok {
				httpConfig.UseHTTP2 = useHTTP2
			}
			if allowH2C, ok := config.Options["allow_h2c"].(bool); ok {
				httpConfig.AllowHTTP2H2C = allowH2C
			}
		}
		
		httpListener := http.NewHTTPListener(httpConfig)
		return NewHTTPListenerAdapter(httpListener), nil
	case ListenerTypeWebSocket:
		// Convert generic config to WebSocket-specific config
		wsConfig := websocket.WebSocketConfig{
			Address: config.Address,
			Port:    config.Port,
		}
		
		// Extract WebSocket-specific options
		if config.Options != nil {
			if path, ok := config.Options["path"].(string); ok {
				wsConfig.Path = path
			}
			if certFile, ok := config.Options["tls_cert_file"].(string); ok {
				wsConfig.TLSCertFile = certFile
			}
			if keyFile, ok := config.Options["tls_key_file"].(string); ok {
				wsConfig.TLSKeyFile = keyFile
			}
		}
		
		wsListener := websocket.NewWebSocketListener(wsConfig)
		return NewWebSocketListenerAdapter(wsListener), nil
	default:
		return nil, fmt.Errorf("unsupported listener type: %s", listenerType)
	}
}

// ValidateListenerConfig validates a listener configuration
func ValidateListenerConfig(listenerType ListenerType, config ListenerConfig) error {
	// Common validation
	if config.Address == "" {
		return errors.New("listener address is required")
	}
	
	// Protocol-specific validation
	switch listenerType {
	case ListenerTypeTCP, ListenerTypeDNS, ListenerTypeHTTP, ListenerTypeWebSocket:
		if config.Port <= 0 || config.Port > 65535 {
			return errors.New("invalid port number")
		}
	case ListenerTypeDNS:
		if config.Options == nil || config.Options["domain"] == nil {
			return errors.New("DNS listener requires a domain")
		}
	}
	
	return nil
}
