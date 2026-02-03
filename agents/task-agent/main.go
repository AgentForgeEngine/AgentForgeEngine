package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type TaskAgent struct {
	name        string
	maxTasks    int
	activeTasks map[string]*task
}

type task struct {
	id       string
	command  string
	args     []string
	status   string
	started  time.Time
	finished *time.Time
	output   string
	error    string
}

func NewTaskAgent() *TaskAgent {
	return &TaskAgent{
		name:        "task-agent",
		maxTasks:    5,
		activeTasks: make(map[string]*task),
	}
}

func (ta *TaskAgent) Name() string {
	return ta.name
}

func (ta *TaskAgent) Initialize(config map[string]interface{}) error {
	// Set max concurrent tasks from config
	if maxTasks, ok := config["max_concurrent_tasks"].(int); ok {
		ta.maxTasks = maxTasks
	}

	return nil
}

func (ta *TaskAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	switch input.Type {
	case "execute":
		return ta.executeTask(ctx, input)
	case "status":
		return ta.getTaskStatus(input)
	case "list":
		return ta.listTasks()
	case "cancel":
		return ta.cancelTask(input)
	default:
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("unknown task type: %s", input.Type),
		}, nil
	}
}

func (ta *TaskAgent) HealthCheck() error {
	// Check if we can execute basic commands
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "echo", "health-check")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

func (ta *TaskAgent) Shutdown() error {
	// Cancel all active tasks
	for id, task := range ta.activeTasks {
		task.status = "cancelled"
		now := time.Now()
		task.finished = &now
		ta.activeTasks[id] = task
	}

	return nil
}

func (ta *TaskAgent) executeTask(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Check task limit
	if len(ta.activeTasks) >= ta.maxTasks {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("maximum concurrent tasks (%d) reached", ta.maxTasks),
		}, nil
	}

	// Extract command from payload
	command, ok := input.Payload["command"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "command not specified in payload",
		}, nil
	}

	// Extract args if present
	var args []string
	if argsInterface, ok := input.Payload["args"]; ok {
		if argsSlice, ok := argsInterface.([]interface{}); ok {
			for _, arg := range argsSlice {
				if argStr, ok := arg.(string); ok {
					args = append(args, argStr)
				}
			}
		}
	}

	// Generate task ID
	taskID := generateTaskID(command)

	// Create and execute task
	task := &task{
		id:      taskID,
		command: command,
		args:    args,
		status:  "running",
		started: time.Now(),
	}

	ta.activeTasks[taskID] = task

	// Execute command in goroutine
	go ta.runCommand(ctx, task)

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"task_id": taskID,
			"status":  "started",
		},
	}, nil
}

func (ta *TaskAgent) runCommand(ctx context.Context, task *task) {
	cmd := exec.CommandContext(ctx, task.command, task.args...)
	output, err := cmd.CombinedOutput()

	task.status = "completed"
	now := time.Now()
	task.finished = &now

	if err != nil {
		task.status = "failed"
		task.error = err.Error()
	} else {
		task.output = string(output)
	}

	ta.activeTasks[task.id] = task
}

func (ta *TaskAgent) getTaskStatus(input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	taskID, ok := input.Payload["task_id"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "task_id not specified in payload",
		}, nil
	}

	task, exists := ta.activeTasks[taskID]
	if !exists {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("task %s not found", taskID),
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"task_id":  task.id,
			"command":  task.command,
			"args":     task.args,
			"status":   task.status,
			"started":  task.started,
			"finished": task.finished,
			"output":   task.output,
			"error":    task.error,
		},
	}, nil
}

func (ta *TaskAgent) listTasks() (interfaces.AgentOutput, error) {
	var tasks []map[string]interface{}

	for _, task := range ta.activeTasks {
		taskData := map[string]interface{}{
			"task_id":  task.id,
			"command":  task.command,
			"args":     task.args,
			"status":   task.status,
			"started":  task.started,
			"finished": task.finished,
		}
		if task.output != "" {
			taskData["output"] = task.output
		}
		if task.error != "" {
			taskData["error"] = task.error
		}
		tasks = append(tasks, taskData)
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"tasks":     tasks,
			"total":     len(tasks),
			"max_tasks": ta.maxTasks,
			"available": ta.maxTasks - len(tasks),
		},
	}, nil
}

func (ta *TaskAgent) cancelTask(input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	taskID, ok := input.Payload["task_id"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "task_id not specified in payload",
		}, nil
	}

	task, exists := ta.activeTasks[taskID]
	if !exists {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("task %s not found", taskID),
		}, nil
	}

	if task.status != "running" {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("task %s is not running (status: %s)", taskID, task.status),
		}, nil
	}

	// Mark as cancelled
	task.status = "cancelled"
	now := time.Now()
	task.finished = &now
	ta.activeTasks[taskID] = task

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"task_id": taskID,
			"status":  "cancelled",
		},
	}, nil
}

func generateTaskID(command string) string {
	timestamp := time.Now().UnixNano()
	hash := strings.ToLower(command)
	hash = strings.ReplaceAll(hash, " ", "_")
	hash = strings.ReplaceAll(hash, "/", "_")
	return fmt.Sprintf("%s_%d", hash, timestamp)
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewTaskAgent()
