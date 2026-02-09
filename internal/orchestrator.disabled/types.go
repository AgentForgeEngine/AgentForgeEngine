package orchestrator

import (
	"context"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

// Orchestrator interface defines the contract for all orchestrator functions
type Orchestrator interface {
	Name() string
	Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error)
	HealthCheck() error
	Shutdown() error
}

// BaseOrchestrator provides common functionality for all orchestrators
type BaseOrchestrator struct {
	name       string
	ctx        context.Context
	cancelFunc context.CancelFunc
	config     map[string]interface{}
}

// NewBaseOrchestrator creates a new base orchestrator
func NewBaseOrchestrator(name string, config map[string]interface{}) *BaseOrchestrator {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseOrchestrator{
		name:       name,
		ctx:        ctx,
		cancelFunc: cancel,
		config:     config,
	}
}

// Name returns the orchestrator name
func (bo *BaseOrchestrator) Name() string {
	return bo.name
}

// Context returns the orchestrator's context
func (bo *BaseOrchestrator) Context() context.Context {
	return bo.ctx
}

// Config returns the orchestrator's configuration
func (bo *BaseOrchestrator) Config() map[string]interface{} {
	return bo.config
}

// Shutdown gracefully shuts down the orchestrator
func (bo *BaseOrchestrator) Shutdown() error {
	bo.cancelFunc()
	return nil
}

// HealthCheck performs basic health check
func (bo *BaseOrchestrator) HealthCheck() error {
	select {
	case <-bo.ctx.Done():
		return bo.ctx.Err()
	default:
		return nil
	}
}

// Workflow represents a collection of tasks to be executed
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Todos       []string               `json:"todos"`
	Tasks       []Task                 `json:"tasks"`
	Status      WorkflowStatus         `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Context     map[string]interface{} `json:"context"`
	Results     []TaskResult           `json:"results"`
}

// Task represents a single task to be executed by an agent
type Task struct {
	ID          string                 `json:"id"`
	AgentName   string                 `json:"agent_name"`
	Description string                 `json:"description"`
	Arguments   map[string]interface{} `json:"arguments"`
	Status      TaskStatus             `json:"status"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Context     map[string]interface{} `json:"context"`
	DependsOn   []string               `json:"depends_on,omitempty"`
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID    string                 `json:"task_id"`
	AgentName string                 `json:"agent_name"`
	Success   bool                   `json:"success"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

// WorkflowStatus represents the status of a workflow
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusPaused    WorkflowStatus = "paused"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// WorkflowResult represents the final result of a workflow execution
type WorkflowResult struct {
	WorkflowID     string                 `json:"workflow_id"`
	Status         WorkflowStatus         `json:"status"`
	TotalTasks     int                    `json:"total_tasks"`
	CompletedTasks int                    `json:"completed_tasks"`
	FailedTasks    int                    `json:"failed_tasks"`
	Duration       time.Duration          `json:"duration"`
	Tasks          []TaskResult           `json:"tasks"`
	Context        map[string]interface{} `json:"context"`
	Error          string                 `json:"error,omitempty"`
	Summary        string                 `json:"summary"`
}

// ParsedTodo represents a parsed todo item
type ParsedTodo struct {
	Original   string                 `json:"original"`
	Cleaned    string                 `json:"cleaned"`
	AgentName  string                 `json:"agent_name"`
	Arguments  map[string]interface{} `json:"arguments"`
	Confidence float64                `json:"confidence"`
}

// WorkflowEngine handles workflow execution and task management
type WorkflowEngine interface {
	CreateWorkflow(name string, todos []string, context map[string]interface{}) (*Workflow, error)
	ExecuteWorkflow(ctx context.Context, workflow *Workflow) (*WorkflowResult, error)
	GetWorkflow(workflowID string) (*Workflow, bool)
	ListWorkflows() []Workflow
	CancelWorkflow(workflowID string) error
}

// TodoParser handles parsing of todo lists into tasks
type TodoParser interface {
	ParseTodo(todo string) (*ParsedTodo, error)
	ParseMultiple(todos []string) ([]*ParsedTodo, error)
}

// TaskRouter handles routing of tasks to appropriate agents
type TaskRouter interface {
	RouteTask(todo *ParsedTodo) (string, map[string]interface{}, error)
	ListAgents() []string
	GetAgentCapabilities(agentName string) ([]string, error)
}
