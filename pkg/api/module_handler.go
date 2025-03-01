package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ModuleHandler handles module API endpoints
type ModuleHandler struct {
	*Handler
	moduleManager interface{} // Will be replaced with actual module manager type
	modulePath    string
}

// NewModuleHandler creates a new module handler
func NewModuleHandler(base *Handler, moduleManager interface{}, modulePath string) *ModuleHandler {
	return &ModuleHandler{
		Handler:       base,
		moduleManager: moduleManager,
		modulePath:    modulePath,
	}
}

// RegisterRoutes registers the module API routes
func (h *ModuleHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/modules", h.AuthMiddleware(h.handleModules))
	mux.HandleFunc("/api/v1/modules/", h.AuthMiddleware(h.handleModule))
}

// handleModules handles GET and POST requests to /api/v1/modules
func (h *ModuleHandler) handleModules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getModules(w, r)
	case http.MethodPost:
		h.uploadModule(w, r)
	default:
		h.RespondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
	}
}

// handleModule handles GET and DELETE requests to /api/v1/modules/{name}
func (h *ModuleHandler) handleModule(w http.ResponseWriter, r *http.Request) {
	// Extract the module name from the URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		h.RespondError(w, http.StatusBadRequest, fmt.Errorf("invalid URL"))
		return
	}
	name := parts[4]

	switch r.Method {
	case http.MethodGet:
		h.getModule(w, r, name)
	case http.MethodDelete:
		h.deleteModule(w, r, name)
	default:
		h.RespondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
	}
}

// getModules returns a list of all available modules
func (h *ModuleHandler) getModules(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would get the modules from the module manager
	modules := []map[string]interface{}{
		{
			"name":        "shell",
			"description": "Provides shell access to the client",
			"version":     "1.0.0",
		},
		{
			"name":        "file",
			"description": "Provides file system access",
			"version":     "1.0.0",
		},
	}

	h.RespondJSON(w, http.StatusOK, modules, "Modules retrieved successfully")
}

// uploadModule uploads a new module
func (h *ModuleHandler) uploadModule(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	// Get the file from the form
	file, handler, err := r.FormFile("module")
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Errorf("failed to get file: %w", err))
		return
	}
	defer file.Close()

	// Create the module directory if it doesn't exist
	if err := os.MkdirAll(h.modulePath, 0755); err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Errorf("failed to create module directory: %w", err))
		return
	}

	// Create the module file
	modulePath := filepath.Join(h.modulePath, handler.Filename)
	moduleFile, err := os.Create(modulePath)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Errorf("failed to create module file: %w", err))
		return
	}
	defer moduleFile.Close()

	// Copy the file
	if _, err := io.Copy(moduleFile, file); err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save module file: %w", err))
		return
	}

	// In a real implementation, this would register the module with the module manager
	h.RespondJSON(w, http.StatusCreated, nil, "Module uploaded successfully")
}

// getModule returns a specific module
func (h *ModuleHandler) getModule(w http.ResponseWriter, r *http.Request, name string) {
	// In a real implementation, this would get the module from the module manager
	module := map[string]interface{}{
		"name":        name,
		"description": "Module description",
		"version":     "1.0.0",
	}

	h.RespondJSON(w, http.StatusOK, module, "Module retrieved successfully")
}

// deleteModule deletes a specific module
func (h *ModuleHandler) deleteModule(w http.ResponseWriter, r *http.Request, name string) {
	// In a real implementation, this would delete the module using the module manager
	h.RespondJSON(w, http.StatusOK, nil, "Module deleted successfully")
}
