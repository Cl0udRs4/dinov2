package task

import (
	"errors"
	"sync"
	"time"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// TaskPriority represents the priority level of a task
type TaskPriority int

const (
	TaskPriorityLow    TaskPriority = 0
	TaskPriorityNormal TaskPriority = 50
	TaskPriorityHigh   TaskPriority = 100
)

// TaskType represents the type of task
type TaskType string

const (
	TaskTypeCommand        TaskType = "command"
	TaskTypeModuleLoad     TaskType = "module_load"
	TaskTypeModuleExec     TaskType = "module_exec"
	TaskTypeProtocolSwitch TaskType = "protocol_switch"
	TaskTypeKeyExchange    TaskType = "key_exchange"
)

// Task represents a task to be executed by a client
type Task struct {
	ID          uint32
	Type        TaskType
	ClientID    string
	Data        []byte
	Priority    TaskPriority
	Status      TaskStatus
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	Result      []byte
	Error       string
	DependsOn   []uint32 // IDs of tasks that must complete before this one
}

// Manager handles task creation, scheduling, and tracking
type Manager struct {
	tasks       map[uint32]*Task
	nextID      uint32
	mutex       sync.RWMutex
	pendingChan chan *Task
}

// NewManager creates a new task manager
func NewManager() *Manager {
	return &Manager{
		tasks:       make(map[uint32]*Task),
		nextID:      1,
		pendingChan: make(chan *Task, 100),
	}
}

// CreateTask creates a new task and adds it to the manager
func (m *Manager) CreateTask(taskType TaskType, clientID string, data []byte, priority TaskPriority, dependsOn []uint32) (*Task, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Create the task
	task := &Task{
		ID:        m.nextID,
		Type:      taskType,
		ClientID:  clientID,
		Data:      data,
		Priority:  priority,
		Status:    TaskStatusPending,
		CreatedAt: time.Now(),
		DependsOn: dependsOn,
	}

	// Increment the next ID
	m.nextID++

	// Add the task to the manager
	m.tasks[task.ID] = task

	// Check if the task can be scheduled immediately
	if len(dependsOn) == 0 {
		// No dependencies, can be scheduled immediately
		go func() {
			m.pendingChan <- task
		}()
	} else {
		// Check if all dependencies are completed
		canSchedule := true
		for _, depID := range dependsOn {
			depTask, exists := m.tasks[depID]
			if !exists || depTask.Status != TaskStatusCompleted {
				canSchedule = false
				break
			}
		}

		if canSchedule {
			go func() {
				m.pendingChan <- task
			}()
		}
	}

	return task, nil
}

// GetTask retrieves a task by ID
func (m *Manager) GetTask(id uint32) (*Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	task, exists := m.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}

	return task, nil
}

// UpdateTaskStatus updates the status of a task
func (m *Manager) UpdateTaskStatus(id uint32, status TaskStatus, result []byte, errorMsg string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task, exists := m.tasks[id]
	if !exists {
		return errors.New("task not found")
	}

	// Update the task status
	task.Status = status
	
	// Update additional fields based on status
	switch status {
	case TaskStatusRunning:
		task.StartedAt = time.Now()
	case TaskStatusCompleted, TaskStatusFailed, TaskStatusCancelled:
		task.CompletedAt = time.Now()
		task.Result = result
		task.Error = errorMsg

		// If completed, check if any dependent tasks can now be scheduled
		if status == TaskStatusCompleted {
			m.checkDependentTasks(id)
		}
	}

	return nil
}

// checkDependentTasks checks if any tasks that depend on the given task ID can now be scheduled
func (m *Manager) checkDependentTasks(completedTaskID uint32) {
	// Find all tasks that depend on the completed task
	for _, task := range m.tasks {
		if task.Status == TaskStatusPending {
			// Check if this task depends on the completed task
			isDependentTask := false
			for _, depID := range task.DependsOn {
				if depID == completedTaskID {
					isDependentTask = true
					break
				}
			}

			if isDependentTask {
				// Check if all dependencies are now completed
				allDepsCompleted := true
				for _, depID := range task.DependsOn {
					depTask, exists := m.tasks[depID]
					if !exists || depTask.Status != TaskStatusCompleted {
						allDepsCompleted = false
						break
					}
				}

				if allDepsCompleted {
					// All dependencies are completed, schedule the task
					go func(t *Task) {
						m.pendingChan <- t
					}(task)
				}
			}
		}
	}
}

// GetPendingTask returns the next pending task from the queue
func (m *Manager) GetPendingTask() *Task {
	return <-m.pendingChan
}

// ListTasks returns a list of all tasks
func (m *Manager) ListTasks() []*Task {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	tasks := make([]*Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// ListClientTasks returns a list of tasks for a specific client
func (m *Manager) ListClientTasks(clientID string) []*Task {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	tasks := make([]*Task, 0)
	for _, task := range m.tasks {
		if task.ClientID == clientID {
			tasks = append(tasks, task)
		}
	}

	return tasks
}
