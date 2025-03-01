package api

import (
	"encoding/json"
	"net/http"
	
	"dinoc2/pkg/module/loader"
)

// ModuleLoadRequest represents a request to load a module
type ModuleLoadRequest struct {
	Name       string           `json:"name"`
	Path       string           `json:"path"`
	LoaderType loader.LoaderType `json:"loader_type"`
}

// ModuleExecRequest represents a request to execute a module
type ModuleExecRequest struct {
	Name    string        `json:"name"`
	Command string        `json:"command"`
	Args    []interface{} `json:"args"`
}

// handleListModules handles GET /api/modules
func (r *Router) handleListModules(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	modules := r.moduleManager.ListModules()
	writeJSON(w, modules, http.StatusOK)
}

// handleLoadModule handles POST /api/modules/load
func (r *Router) handleLoadModule(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var moduleReq ModuleLoadRequest
	if err := json.NewDecoder(req.Body).Decode(&moduleReq); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Load module
	module, err := r.moduleManager.LoadModule(
		moduleReq.Name,
		moduleReq.Path,
		moduleReq.LoaderType,
	)
	
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	writeJSON(w, module, http.StatusOK)
}

// handleExecModule handles POST /api/modules/exec
func (r *Router) handleExecModule(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var execReq ModuleExecRequest
	if err := json.NewDecoder(req.Body).Decode(&execReq); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Execute module
	result, err := r.moduleManager.ExecModule(
		execReq.Name,
		execReq.Command,
		execReq.Args...,
	)
	
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	writeJSON(w, result, http.StatusOK)
}
