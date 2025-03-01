package listener

import (
	"dinoc2/pkg/listener/dns"
	"dinoc2/pkg/listener/http"
	"dinoc2/pkg/listener/icmp"
	"dinoc2/pkg/listener/websocket"
)

// DNSListenerAdapter adapts the DNS listener to the Listener interface
type DNSListenerAdapter struct {
	listener *dns.DNSListener
}

// NewDNSListenerAdapter creates a new DNS listener adapter
func NewDNSListenerAdapter(listener *dns.DNSListener) *DNSListenerAdapter {
	return &DNSListenerAdapter{listener: listener}
}

// Start implements the Listener interface
func (a *DNSListenerAdapter) Start() error {
	return a.listener.Start()
}

// Stop implements the Listener interface
func (a *DNSListenerAdapter) Stop() error {
	return a.listener.Stop()
}

// Status implements the Listener interface
func (a *DNSListenerAdapter) Status() ListenerStatus {
	status := a.listener.Status()
	switch status {
	case "running":
		return StatusRunning
	case "stopped":
		return StatusStopped
	case "error":
		return StatusError
	default:
		return StatusUnknown
	}
}

// Configure implements the Listener interface
func (a *DNSListenerAdapter) Configure(config ListenerConfig) error {
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
	
	return a.listener.Configure(dnsConfig)
}

// ICMPListenerAdapter adapts the ICMP listener to the Listener interface
type ICMPListenerAdapter struct {
	listener *icmp.ICMPListener
}

// NewICMPListenerAdapter creates a new ICMP listener adapter
func NewICMPListenerAdapter(listener *icmp.ICMPListener) *ICMPListenerAdapter {
	return &ICMPListenerAdapter{listener: listener}
}

// Start implements the Listener interface
func (a *ICMPListenerAdapter) Start() error {
	return a.listener.Start()
}

// Stop implements the Listener interface
func (a *ICMPListenerAdapter) Stop() error {
	return a.listener.Stop()
}

// Status implements the Listener interface
func (a *ICMPListenerAdapter) Status() ListenerStatus {
	status := a.listener.Status()
	switch status {
	case "running":
		return StatusRunning
	case "stopped":
		return StatusStopped
	case "error":
		return StatusError
	default:
		return StatusUnknown
	}
}

// Configure implements the Listener interface
func (a *ICMPListenerAdapter) Configure(config ListenerConfig) error {
	icmpConfig := icmp.ICMPConfig{
		ListenAddress: config.Address,
	}
	
	// Extract ICMP-specific options
	if config.Options != nil {
		if protocol, ok := config.Options["protocol"].(string); ok {
			icmpConfig.Protocol = protocol
		}
	}
	
	return a.listener.Configure(icmpConfig)
}

// HTTPListenerAdapter adapts the HTTP listener to the Listener interface
type HTTPListenerAdapter struct {
	listener *http.HTTPListener
}

// NewHTTPListenerAdapter creates a new HTTP listener adapter
func NewHTTPListenerAdapter(listener *http.HTTPListener) *HTTPListenerAdapter {
	return &HTTPListenerAdapter{listener: listener}
}

// Start implements the Listener interface
func (a *HTTPListenerAdapter) Start() error {
	return a.listener.Start()
}

// Stop implements the Listener interface
func (a *HTTPListenerAdapter) Stop() error {
	return a.listener.Stop()
}

// Status implements the Listener interface
func (a *HTTPListenerAdapter) Status() ListenerStatus {
	status := a.listener.Status()
	switch status {
	case "running":
		return StatusRunning
	case "stopped":
		return StatusStopped
	case "error":
		return StatusError
	default:
		return StatusUnknown
	}
}

// Configure implements the Listener interface
func (a *HTTPListenerAdapter) Configure(config ListenerConfig) error {
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
	
	return a.listener.Configure(httpConfig)
}

// WebSocketListenerAdapter adapts the WebSocket listener to the Listener interface
type WebSocketListenerAdapter struct {
	listener *websocket.WebSocketListener
}

// NewWebSocketListenerAdapter creates a new WebSocket listener adapter
func NewWebSocketListenerAdapter(listener *websocket.WebSocketListener) *WebSocketListenerAdapter {
	return &WebSocketListenerAdapter{listener: listener}
}

// Start implements the Listener interface
func (a *WebSocketListenerAdapter) Start() error {
	return a.listener.Start()
}

// Stop implements the Listener interface
func (a *WebSocketListenerAdapter) Stop() error {
	return a.listener.Stop()
}

// Status implements the Listener interface
func (a *WebSocketListenerAdapter) Status() ListenerStatus {
	status := a.listener.Status()
	switch status {
	case "running":
		return StatusRunning
	case "stopped":
		return StatusStopped
	case "error":
		return StatusError
	default:
		return StatusUnknown
	}
}

// Configure implements the Listener interface
func (a *WebSocketListenerAdapter) Configure(config ListenerConfig) error {
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
	
	return a.listener.Configure(wsConfig)
}
