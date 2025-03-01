package api

import (
	"net/http"
)

// HTTPHandlerProvider defines an interface for providing HTTP handlers
type HTTPHandlerProvider interface {
	// ServeHTTP handles HTTP requests
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}
