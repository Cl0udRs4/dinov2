package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// HTTPListener implements the Listener interface for HTTP/HTTP2 protocol
type HTTPListener struct {
	config     HTTPConfig
	server     *http.Server
	status     string
	statusLock sync.RWMutex
	handlers   map[string]http.HandlerFunc
}

// HTTPConfig holds configuration for the HTTP listener
type HTTPConfig struct {
	Address      string
	Port         int
	TLSCertFile  string
	TLSKeyFile   string
	UseHTTP2     bool
	AllowHTTP2H2C bool // Allow HTTP/2 cleartext (h2c)
	Options      map[string]interface{}
}

// NewHTTPListener creates a new HTTP listener
func NewHTTPListener(config HTTPConfig) *HTTPListener {
	return &HTTPListener{
		config:   config,
		status:   "stopped",
		handlers: make(map[string]http.HandlerFunc),
	}
}

// Start implements the Listener interface
func (l *HTTPListener) Start() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == "running" {
		return fmt.Errorf("HTTP listener is already running")
	}

	// Create a new HTTP server
	addr := fmt.Sprintf("%s:%d", l.config.Address, l.config.Port)
	
	// Create a router/mux for handling requests
	mux := http.NewServeMux()
	
	// Register handlers
	for path, handler := range l.handlers {
		mux.HandleFunc(path, handler)
	}
	
	// If no handlers are registered, add a default one
	if len(l.handlers) == 0 {
		mux.HandleFunc("/", l.defaultHandler)
	}
	
	var handler http.Handler = mux
	
	// If HTTP/2 cleartext (h2c) is enabled, wrap the handler
	if l.config.AllowHTTP2H2C {
		h2s := &http2.Server{}
		handler = h2c.NewHandler(mux, h2s)
	}
	
	l.server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	
	// Configure HTTP/2 if requested
	if l.config.UseHTTP2 {
		if err := http2.ConfigureServer(l.server, &http2.Server{}); err != nil {
			l.status = "error"
			return fmt.Errorf("failed to configure HTTP/2: %w", err)
		}
	}
	
	// Start the server in a goroutine
	go func() {
		var err error
		
		// Check if TLS is configured
		if l.config.TLSCertFile != "" && l.config.TLSKeyFile != "" {
			err = l.server.ListenAndServeTLS(l.config.TLSCertFile, l.config.TLSKeyFile)
		} else {
			err = l.server.ListenAndServe()
		}
		
		if err != nil && err != http.ErrServerClosed {
			l.statusLock.Lock()
			l.status = "error"
			l.statusLock.Unlock()
			fmt.Printf("Error starting HTTP listener: %v\n", err)
		}
	}()
	
	l.status = "running"
	return nil
}

// Stop implements the Listener interface
func (l *HTTPListener) Stop() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status != "running" {
		return nil // Already stopped
	}

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if l.server != nil {
		err := l.server.Shutdown(ctx)
		if err != nil {
			l.status = "error"
			return fmt.Errorf("error stopping HTTP listener: %w", err)
		}
	}

	l.status = "stopped"
	return nil
}

// Status implements the Listener interface
func (l *HTTPListener) Status() string {
	l.statusLock.RLock()
	defer l.statusLock.RUnlock()
	return l.status
}

// Configure implements the Listener interface
func (l *HTTPListener) Configure(config interface{}) error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == "running" {
		return fmt.Errorf("cannot configure a running HTTP listener")
	}

	httpConfig, ok := config.(HTTPConfig)
	if !ok {
		return fmt.Errorf("invalid configuration type for HTTP listener")
	}

	l.config = httpConfig
	return nil
}

// RegisterHandler registers a handler for a specific path
func (l *HTTPListener) RegisterHandler(path string, handler http.HandlerFunc) {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()
	
	l.handlers[path] = handler
}

// defaultHandler is the default HTTP handler
func (l *HTTPListener) defaultHandler(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would pass the request to the protocol layer
	fmt.Printf("Received HTTP request from %s: %s %s\n", r.RemoteAddr, r.Method, r.URL.Path)
	
	// Send a generic response
	w.Header().Set("Server", "Microsoft-IIS/10.0")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<html><body><h1>It works!</h1></body></html>"))
}

// CreateTLSConfig creates a TLS configuration for the HTTP server
func CreateTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}
