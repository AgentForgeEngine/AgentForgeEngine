package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

// WorkflowEngineImpl implements WorkflowEngine interface
type WorkflowEngineImpl struct {
	workflows map[string]*Workflow
	mu        sync.RWMutex
	pluginMgr interfaces.PluginManager
	parser    TodoParser
	router    TaskRouter
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(pluginMgr interfaces.PluginManager, parser TodoParser, router TaskRouter) *WorkflowEngineImpl {
	return &WorkflowEngineImpl{
		workflows: make(map[string]*Workflow),
		pluginMgr: pluginMgr,
		parser:    parser,
		router:    router,
	}
}

// CreateWorkflow creates a new workflow from todos
func (we *WorkflowEngineImpl) CreateWorkflow(name string, todos []string, context map[string]interface{}) (*Workflow, error) {
	we.mu.Lock()
	defer we.mu.Unlock()

	// Generate workflow ID
	workflowID := fmt.Sprintf("workflow-%d-%d", time.Now().Unix(), len(we.workflows))

	// Parse todos into tasks
	parsedTodos, err := we.parser.ParseMultiple(todos)
	if err != nil {
		return nil, fmt.Errorf("failed to parse todos: %w", err)
	}

	// Create tasks from parsed todos
	tasks := make([]Task, 0, len(parsedTodos))
	for i, parsed := range parsedTodos {
		task := Task{
			ID:          fmt.Sprintf("%s-task-%d", workflowID, i),
			AgentName:   parsed.AgentName,
			Description: parsed.Cleaned,
			Arguments:   parsed.Arguments,
			Status:      TaskStatusPending,
			Context:     make(map[string]interface{}),
		}
		tasks = append(tasks, task)
	}

	// Create workflow
	workflow := &Workflow{
		ID:        workflowID,
		Name:      name,
		Todos:     todos,
		Tasks:     tasks,
		Status:    WorkflowStatusPending,
		CreatedAt: time.Now(),
		Context:   context,
		Results:   make([]TaskResult, 0),
	}

	we.workflows[workflowID] = workflow
	return workflow, nil
}

// ExecuteWorkflow executes a workflow
func (we *WorkflowEngineImpl) ExecuteWorkflow(ctx context.Context, workflow *Workflow) (*WorkflowResult, error) {
	we.mu.Lock()
	if workflow.Status != WorkflowStatusPending {
		we.mu.Unlock()
		return nil, fmt.Errorf("workflow %s is not pending", workflow.ID)
	}

	workflow.Status = WorkflowStatusRunning
	now := time.Now()
	workflow.StartedAt = &now
	we.mu.Unlock()

	// Execute tasks sequentially (simple implementation)
	var results []TaskResult
	completedTasks := 0
	failedTasks := 0
	var totalDuration time.Duration

	for i := range workflow.Tasks {
		task := &workflow.Tasks[i]
		task.Status = TaskStatusRunning
		taskStart := time.Now()

		// Get the agent for this task
		agent, exists := we.pluginMgr.GetAgent(task.AgentName)
		if !exists {
			// Handle unknown agents by routing
			agentName, args, err := we.router.RouteTask(nil) // Use parsed todo info
			if err != nil {
				task.Status = TaskStatusFailed
				task.Error = fmt.Sprintf("failed to route task: %v", err)
				failedTasks++
				results = append(results, TaskResult{
					TaskID:    task.ID,
					AgentName: task.AgentName,
					Success:   false,
					Error:     task.Error,
					Duration:  time.Since(taskStart),
					Timestamp: time.Now(),
				})
				continue
			}
			agent, exists = we.pluginMgr.GetAgent(agentName)
			if !exists {
				task.Status = TaskStatusFailed
				task.Error = fmt.Sprintf("agent %s not found", agentName)
				failedTasks++
				results = append(results, TaskResult{
					TaskID:    task.ID,
					AgentName: task.AgentName,
					Success:   false,
					Error:     task.Error,
					Duration:  time.Since(taskStart),
					Timestamp: time.Now(),
				})
				continue
			}
			task.Arguments = args
		}

		// Execute the task
		input := interfaces.AgentInput{
			Type:    "execute",
			Payload: task.Arguments,
			Metadata: map[string]interface{}{
				"workflow_id": workflow.ID,
				"task_id":     task.ID,
			},
		}

		output, err := agent.Process(ctx, input)
		taskEnd := time.Now()
		taskDuration := taskEnd.Sub(taskStart)
		totalDuration += taskDuration

		if err != nil || !output.Success {
			task.Status = TaskStatusFailed
			task.Error = err.Error()
			if output.Error != "" {
				task.Error = output.Error
			}
			failedTasks++
			results = append(results, TaskResult{
				TaskID:    task.ID,
				AgentName: task.AgentName,
				Success:   false,
				Error:     task.Error,
				Data:      output.Data,
				Duration:  taskDuration,
				Timestamp: time.Now(),
			})
		} else {
			task.Status = TaskStatusCompleted
			completedTasks++
			results = append(results, TaskResult{
				TaskID:    task.ID,
				AgentName: task.AgentName,
				Success:   true,
				Data:      output.Data,
				Duration:  taskDuration,
				Timestamp: time.Now(),
			})
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			workflow.Status = WorkflowStatusCancelled
			return &WorkflowResult{
				WorkflowID:     workflow.ID,
				Status:         WorkflowStatusCancelled,
				TotalTasks:     len(workflow.Tasks),
				CompletedTasks: completedTasks,
				FailedTasks:    failedTasks,
				Duration:       totalDuration,
				Tasks:          results,
				Context:        workflow.Context,
				Error:          ctx.Err().Error(),
			}, ctx.Err()
		default:
		}
	}

	// Update workflow status
	we.mu.Lock()
	defer we.mu.Unlock()

	workflow.Status = WorkflowStatusCompleted
	completedAt := time.Now()
	workflow.CompletedAt = &completedAt
	workflow.Results = results

	// Generate summary
	summary := fmt.Sprintf("Workflow '%s' completed: %d/%d tasks successful (%d failed)",
		workflow.Name, completedTasks, len(workflow.Tasks), failedTasks)

	return &WorkflowResult{
		WorkflowID:     workflow.ID,
		Status:         WorkflowStatusCompleted,
		TotalTasks:     len(workflow.Tasks),
		CompletedTasks: completedTasks,
		FailedTasks:    failedTasks,
		Duration:       totalDuration,
		Tasks:          results,
		Context:        workflow.Context,
		Summary:        summary,
	}, nil
}

// GetWorkflow retrieves a workflow by ID
func (we *WorkflowEngineImpl) GetWorkflow(workflowID string) (*Workflow, bool) {
	we.mu.RLock()
	defer we.mu.RUnlock()

	workflow, exists := we.workflows[workflowID]
	return workflow, exists
}

// ListWorkflows returns all workflows
func (we *WorkflowEngineImpl) ListWorkflows() []Workflow {
	we.mu.RLock()
	defer we.mu.RUnlock()

	workflows := make([]Workflow, 0, len(we.workflows))
	for _, workflow := range we.workflows {
		workflows = append(workflows, *workflow)
	}
	return workflows
}

// CancelWorkflow cancels a workflow
func (we *WorkflowEngineImpl) CancelWorkflow(workflowID string) error {
	we.mu.Lock()
	defer we.mu.Unlock()

	workflow, exists := we.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow %s not found", workflowID)
	}

	if workflow.Status == WorkflowStatusCompleted || workflow.Status == WorkflowStatusCancelled {
		return fmt.Errorf("workflow %s is already completed", workflowID)
	}

	workflow.Status = WorkflowStatusCancelled
	return nil
}
