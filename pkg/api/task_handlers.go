package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"dinoc2/pkg/task"
)

// TaskRequest represents a request to create a task
type TaskRequest struct {
	Type      string           `json:"type"`
	ClientID  string           `json:"client_id"`
	Data      []byte           `json:"data"`
	Priority  task.TaskPriority `json:"priority"`
	DependsOn []uint32         `json:"depends_on"`
}

// handleListTasks handles GET /api/tasks
func (r *Router) handleListTasks(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	clientID := req.URL.Query().Get("client_id")
	var tasks []*task.Task
	
	if clientID != "" {
		// List tasks for a specific client
		tasks = r.taskManager.ListClientTasks(clientID)
	} else {
		// List all tasks
		tasks = r.taskManager.ListTasks()
	}
	
	writeJSON(w, tasks, http.StatusOK)
}

// handleCreateTask handles POST /api/tasks/create
func (r *Router) handleCreateTask(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var taskReq TaskRequest
	if err := json.NewDecoder(req.Body).Decode(&taskReq); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Create task
	task, err := r.taskManager.CreateTask(
		task.TaskType(taskReq.Type),
		taskReq.ClientID,
		taskReq.Data,
		taskReq.Priority,
		taskReq.DependsOn,
	)
	
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	writeJSON(w, task, http.StatusOK)
}

// handleTaskStatus handles GET /api/tasks/status
func (r *Router) handleTaskStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	idStr := req.URL.Query().Get("id")
	if idStr == "" {
		writeError(w, "Task ID is required", http.StatusBadRequest)
		return
	}
	
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		writeError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	
	task, err := r.taskManager.GetTask(uint32(id))
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	writeJSON(w, task, http.StatusOK)
}
