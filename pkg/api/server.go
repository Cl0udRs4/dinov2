package api

import (
	"fmt"
	"net/http"
)

// Server represents the API server
type Server struct {
	mux            *http.ServeMux
	authenticator  interface{} // Will be replaced with actual authenticator type
	listenerManager interface{} // Will be replaced with actual listener manager type
	moduleManager  interface{} // Will be replaced with actual module manager type
	clientManager  interface{} // Will be replaced with actual client manager type
	modulePath     string
}

// NewServer creates a new API server
func NewServer(authenticator interface{}, listenerManager interface{}, moduleManager interface{}, clientManager interface{}, modulePath string) *Server {
	return &Server{
		mux:            http.NewServeMux(),
		authenticator:  authenticator,
		listenerManager: listenerManager,
		moduleManager:  moduleManager,
		clientManager:  clientManager,
		modulePath:     modulePath,
	}
}

// Start starts the API server
func (s *Server) Start(address string, port int) error {
	// Create the base handler
	baseHandler := NewHandler(s.authenticator)

	// Register the listener handler
	listenerHandler := NewListenerHandler(baseHandler, s.listenerManager)
	listenerHandler.RegisterRoutes(s.mux)

	// Register the client handler
	clientHandler := NewClientHandler(baseHandler, s.clientManager)
	clientHandler.RegisterRoutes(s.mux)

	// Register the module handler
	moduleHandler := NewModuleHandler(baseHandler, s.moduleManager, s.modulePath)
	moduleHandler.RegisterRoutes(s.mux)

	// Start the server
	addr := fmt.Sprintf("%s:%d", address, port)
	fmt.Printf("Starting API server on %s\n", addr)
	return http.ListenAndServe(addr, s.mux)
}
