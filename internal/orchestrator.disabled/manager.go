package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/response"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

// Manager is the main orchestrator service
type Manager struct {
	*BaseOrchestrator
	workflowEngine WorkflowEngine
	parser         TodoParser
	router         TaskRouter
	pluginMgr      interfaces.PluginManager
	formatter      response.Formatter
	mu             sync.RWMutex
}

// NewManager creates a new orchestrator manager
func NewManager(pluginMgr interfaces.PluginManager, config map[string]interface{}) *Manager {
	base := NewBaseOrchestrator("orchestrator", config)

	parser := NewTodoParser()
	router := NewTaskRouter(pluginMgr)
	workflowEngine := NewWorkflowEngine(pluginMgr, parser, router)
	formatter := response.NewAutoFormatter()

	return &Manager{
		BaseOrchestrator: base,
		workflowEngine:   workflowEngine,
		parser:           parser,
		router:           router,
		pluginMgr:        pluginMgr,
		formatter:        formatter,
	}
}

// Process handles orchestrator requests
func (m *Manager) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	switch input.Type {
	case "manager":
		return m.processManagerRequest(ctx, input)
	case "workflow-progress":
		return m.processWorkflowProgress(ctx, input)
	case "error-recovery":
		return m.processErrorRecovery(ctx, input)
	default:
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("unknown orchestrator function: %s", input.Type),
		}, nil
	}
}

// processManagerRequest handles the main orchestrator manager function
func (m *Manager) processManagerRequest(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract todos from payload
	todosInterface, exists := input.Payload["todos"]
	if !exists {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "todos field not found in input payload",
		}, nil
	}

	// Convert to string slice
	todosRaw := fmt.Sprintf("%v", todosInterface)
	todos := strings.Split(todosRaw, "\n")

	// Clean up todos
	var cleanTodos []string
	for _, todo := range todos {
		todo = strings.TrimSpace(todo)
		if todo != "" {
			cleanTodos = append(cleanTodos, todo)
		}
	}

	if len(cleanTodos) == 0 {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "no valid todos found",
		}, nil
	}

	// Extract workflow name
	nameInterface, _ := input.Payload["name"]
	name := fmt.Sprintf("workflow-%d", time.Now().Unix())
	if nameInterface != nil {
		name = fmt.Sprintf("%v", nameInterface)
	}

	// Extract context
	workflowContext, _ := input.Payload["context"]
	if workflowContext == nil {
		workflowContext = make(map[string]interface{})
	}

	// Create and execute workflow
	workflow, err := m.workflowEngine.CreateWorkflow(name, cleanTodos, workflowContext.(map[string]interface{}))
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to create workflow: %v", err),
		}, nil
	}

	// Execute workflow
	result, err := m.workflowEngine.ExecuteWorkflow(ctx, workflow)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to execute workflow: %v", err),
		}, nil
	}

	// Format the response for model
	formattedResponse, err := m.formatter.FormatAgentOutput("orchestrator", interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"workflow_id":     result.WorkflowID,
			"status":          result.Status,
			"total_tasks":     result.TotalTasks,
			"completed_tasks": result.CompletedTasks,
			"failed_tasks":    result.FailedTasks,
			"duration":        result.Duration.String(),
			"summary":         result.Summary,
			"tasks":           result.Tasks,
			"context":         result.Context,
		},
	})

	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to format response: %v", err),
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"function_response": formattedResponse,
		},
	}, nil
}

// processWorkflowProgress handles workflow progress requests
func (m *Manager) processWorkflowProgress(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract workflow ID
	workflowIDInterface, exists := input.Payload["workflow_id"]
	if !exists {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "workflow_id not specified in payload",
		}, nil
	}

	workflowID := fmt.Sprintf("%v", workflowIDInterface)

	// Get workflow
	workflow, exists := m.workflowEngine.GetWorkflow(workflowID)
	if !exists {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("workflow %s not found", workflowID),
		}, nil
	}

	// Calculate progress
	totalTasks := len(workflow.Tasks)
	completedTasks := 0
	failedTasks := 0
	var totalDuration time.Duration

	for _, task := range workflow.Tasks {
		switch task.Status {
		case TaskStatusCompleted:
			completedTasks++
		case TaskStatusFailed:
			failedTasks++
		}
		totalDuration += task.Duration
	}

	// Generate progress summary
	status := string(workflow.Status)
	progress := float64(completedTasks) / float64(totalTasks) * 100

	// Format the response for model
	formattedResponse, err := m.formatter.FormatAgentOutput("orchestrator", interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"workflow_id":     workflow.ID,
			"name":            workflow.Name,
			"status":          status,
			"progress":        progress,
			"total_tasks":     totalTasks,
			"completed_tasks": completedTasks,
			"failed_tasks":    failedTasks,
			"duration":        totalDuration.String(),
			"created_at":      workflow.CreatedAt.Format(time.RFC3339),
			"started_at":      workflow.StartedAt.Format(time.RFC3339),
			"completed_at":    workflow.CompletedAt.Format(time.RFC3339),
			"tasks":           workflow.Tasks,
		},
	})

	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to format response: %v", err),
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"function_response": formattedResponse,
		},
	}, nil
}

// processErrorRecovery handles error recovery requests
func (m *Manager) processErrorRecovery(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract workflow ID
	workflowIDInterface, exists := input.Payload["workflow_id"]
	if !exists {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "workflow_id not specified in payload",
		}, nil
	}

	workflowID := fmt.Sprintf("%v", workflowIDInterface)

	// Get workflow
	workflow, exists := m.workflowEngine.GetWorkflow(workflowID)
	if !exists {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("workflow %s not found", workflowID),
		}, nil
	}

	// Check if workflow is in failed state
	if workflow.Status != WorkflowStatusFailed && workflow.Status != WorkflowStatusCancelled {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("workflow %s is not in a failed state (current: %s)", workflowID, workflow.Status),
		}, nil
	}

	// Find failed tasks
	var failedTasks []Task
	for _, task := range workflow.Tasks {
		if task.Status == TaskStatusFailed {
			// Reset task status for retry
			task.Status = TaskStatusPending
			task.Error = ""
			failedTasks = append(failedTasks, task)
		}
	}

	// Retry workflow with only failed tasks
	if len(failedTasks) > 0 {
		retryWorkflow := &Workflow{
			ID:        workflow.ID + "-retry",
			Name:      workflow.Name + " (retry)",
			Status:    WorkflowStatusPending,
			CreatedAt: time.Now(),
			Context:   workflow.Context,
		}

		// Copy only failed tasks
		for _, task := range failedTasks {
			retryWorkflow.Tasks = append(retryWorkflow.Tasks, task)
		}

		// Execute retry workflow
		result, err := m.workflowEngine.ExecuteWorkflow(ctx, retryWorkflow)
		if err != nil {
			return interfaces.AgentOutput{
				Success: false,
				Error:   fmt.Sprintf("failed to execute retry workflow: %v", err),
			}, nil
		}

		// Format the response for model
		formattedResponse, err := m.formatter.FormatAgentOutput("orchestrator", interfaces.AgentOutput{
			Success: true,
			Data: map[string]interface{}{
				"recovery_successful": true,
				"failed_tasks_count":  len(failedTasks),
				"retry_workflow_id":   result.WorkflowID,
				"retry_result":        result,
				"message":             fmt.Sprintf("Attempted to recover %d failed tasks", len(failedTasks)),
			},
		})

		if err != nil {
			return interfaces.AgentOutput{
				Success: false,
				Error:   fmt.Sprintf("failed to format response: %v", err),
			}, nil
		}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"function_response": formattedResponse,
		},
	}, nil
}

// GetAvailableOrchestrators returns list of available orchestrator functions

// GetAvailableOrchestrators returns list of available orchestrator functions
func (m *Manager) GetAvailableOrchestrators() map[string]string {
	return map[string]string{
		"orch.manager":           "Main orchestrator - processes todo lists and coordinates multiple agents",
		"orch.workflow-progress": "Check workflow status and progress",
		"orch.error-recovery":    "Handle failed workflows and retry tasks",
	}
}

// Name returns orchestrator name
func (m *Manager) Name() string {
	return "orchestrator"
}
