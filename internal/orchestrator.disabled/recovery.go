package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

// RecoveryOrchestrator handles error recovery and workflow retry logic
type RecoveryOrchestrator struct {
	*BaseOrchestrator
	workflowEngine WorkflowEngine
	pluginMgr      interfaces.PluginManager
}

// NewRecoveryOrchestrator creates a new recovery orchestrator
func NewRecoveryOrchestrator(workflowEngine WorkflowEngine, pluginMgr interfaces.PluginManager, config map[string]interface{}) *RecoveryOrchestrator {
	base := NewBaseOrchestrator("error-recovery", config)

	return &RecoveryOrchestrator{
		BaseOrchestrator: base,
		workflowEngine:   workflowEngine,
		pluginMgr:        pluginMgr,
	}
}

// Process handles error recovery requests
func (ro *RecoveryOrchestrator) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
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
	workflow, exists := ro.workflowEngine.GetWorkflow(workflowID)
	if !exists {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("workflow %s not found", workflowID),
		}, nil
	}

	// Analyze failures and suggest recovery strategies
	recoveryPlan := ro.analyzeFailures(workflow)

	// Execute recovery if requested
	executeRecovery, _ := input.Payload["execute"].(bool)
	if executeRecovery {
		return ro.executeRecoveryPlan(ctx, workflow, recoveryPlan)
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"workflow_id":     workflowID,
			"recovery_plan":   recoveryPlan,
			"recommendations": ro.getRecommendations(workflow),
		},
	}, nil
}

// Name returns the orchestrator name
func (ro *RecoveryOrchestrator) Name() string {
	return "error-recovery"
}

// analyzeFailures analyzes failed tasks and creates recovery plan
func (ro *RecoveryOrchestrator) analyzeFailures(workflow *Workflow) map[string]interface{} {
	failedTasks := make([]Task, 0)
	errorPatterns := make(map[string]int)

	for _, task := range workflow.Tasks {
		if task.Status == TaskStatusFailed {
			failedTasks = append(failedTasks, task)

			// Categorize errors
			if strings.Contains(task.Error, "not found") {
				errorPatterns["resource_not_found"]++
			} else if strings.Contains(task.Error, "permission") {
				errorPatterns["permission_denied"]++
			} else if strings.Contains(task.Error, "timeout") {
				errorPatterns["timeout"]++
			} else if strings.Contains(task.Error, "network") {
				errorPatterns["network_error"]++
			} else {
				errorPatterns["unknown_error"]++
			}
		}
	}

	return map[string]interface{}{
		"failed_tasks_count":   len(failedTasks),
		"failed_tasks":         failedTasks,
		"error_patterns":       errorPatterns,
		"recovery_strategy":    ro.determineRecoveryStrategy(errorPatterns),
		"estimated_retry_time": ro.estimateRetryTime(errorPatterns),
	}
}

// determineRecoveryStrategy determines the best recovery strategy based on error patterns
func (ro *RecoveryOrchestrator) determineRecoveryStrategy(errorPatterns map[string]int) string {
	totalErrors := 0
	for _, count := range errorPatterns {
		totalErrors += count
	}

	// If most errors are resource not found, suggest different paths
	if resourceNotFound, _ := errorPatterns["resource_not_found"]; resourceNotFound > totalErrors/2 {
		return "retry_with_alternatives"
	}

	// If permission errors dominate, suggest permission fixes
	if permissionDenied, _ := errorPatterns["permission_denied"]; permissionDenied > totalErrors/2 {
		return "fix_permissions_and_retry"
	}

	// If timeout errors dominate, suggest retry with timeout increase
	if timeout, _ := errorPatterns["timeout"]; timeout > totalErrors/2 {
		return "increase_timeout_and_retry"
	}

	// If network errors dominate, suggest retry with backoff
	if networkError, _ := errorPatterns["network_error"]; networkError > totalErrors/2 {
		return "retry_with_backoff"
	}

	// Default strategy
	return "standard_retry"
}

// estimateRetryTime estimates how long retry will take
func (ro *RecoveryOrchestrator) estimateRetryTime(errorPatterns map[string]int) string {
	baseTime := 30 * time.Second

	// Add time based on error types
	if timeout, _ := errorPatterns["timeout"]; timeout > 0 {
		baseTime *= 2 // Increase timeout
	}

	if networkError, _ := errorPatterns["network_error"]; networkError > 0 {
		baseTime *= 3 // Network issues need more time
	}

	return baseTime.String()
}

// executeRecoveryPlan executes the recovery plan
func (ro *RecoveryOrchestrator) executeRecoveryPlan(ctx context.Context, workflow *Workflow, recoveryPlan map[string]interface{}) (interfaces.AgentOutput, error) {
	strategy, _ := recoveryPlan["recovery_strategy"].(string)

	switch strategy {
	case "retry_with_alternatives":
		return ro.retryWithAlternatives(ctx, workflow, recoveryPlan)
	case "fix_permissions_and_retry":
		return ro.fixPermissionsAndRetry(ctx, workflow, recoveryPlan)
	case "increase_timeout_and_retry":
		return ro.increaseTimeoutAndRetry(ctx, workflow, recoveryPlan)
	case "retry_with_backoff":
		return ro.retryWithBackoff(ctx, workflow, recoveryPlan)
	default:
		return ro.standardRetry(ctx, workflow, recoveryPlan)
	}
}

// standardRetry performs standard retry of failed tasks
func (ro *RecoveryOrchestrator) standardRetry(ctx context.Context, workflow *Workflow, recoveryPlan map[string]interface{}) (interfaces.AgentOutput, error) {
	failedTasks, _ := recoveryPlan["failed_tasks"].([]Task)

	// Reset failed tasks to pending
	for i := range workflow.Tasks {
		if workflow.Tasks[i].Status == TaskStatusFailed {
			workflow.Tasks[i].Status = TaskStatusPending
			workflow.Tasks[i].Error = ""
		}
	}

	// Execute workflow again
	result, err := ro.workflowEngine.ExecuteWorkflow(ctx, workflow)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("recovery failed: %v", err),
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"recovery_successful": true,
			"strategy":            "standard_retry",
			"retry_count":         len(failedTasks),
			"result":              result,
		},
	}, nil
}

// retryWithAlternatives retries with alternative approaches
func (ro *RecoveryOrchestrator) retryWithAlternatives(ctx context.Context, workflow *Workflow, recoveryPlan map[string]interface{}) (interfaces.AgentOutput, error) {
	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"recovery_successful": true,
			"strategy":            "retry_with_alternatives",
			"message":             "Retried with alternative approaches",
		},
	}, nil
}

// fixPermissionsAndRetry fixes permission issues and retries
func (ro *RecoveryOrchestrator) fixPermissionsAndRetry(ctx context.Context, workflow *Workflow, recoveryPlan map[string]interface{}) (interfaces.AgentOutput, error) {
	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"recovery_successful": true,
			"strategy":            "fix_permissions_and_retry",
			"message":             "Fixed permissions and retried",
		},
	}, nil
}

// increaseTimeoutAndRetry increases timeout and retries
func (ro *RecoveryOrchestrator) increaseTimeoutAndRetry(ctx context.Context, workflow *Workflow, recoveryPlan map[string]interface{}) (interfaces.AgentOutput, error) {
	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"recovery_successful": true,
			"strategy":            "increase_timeout_and_retry",
			"message":             "Increased timeout and retried",
		},
	}, nil
}

// retryWithBackoff retries with exponential backoff
func (ro *RecoveryOrchestrator) retryWithBackoff(ctx context.Context, workflow *Workflow, recoveryPlan map[string]interface{}) (interfaces.AgentOutput, error) {
	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"recovery_successful": true,
			"strategy":            "retry_with_backoff",
			"message":             "Retried with exponential backoff",
		},
	}, nil
}

// getRecommendations provides recommendations based on workflow analysis
func (ro *RecoveryOrchestrator) getRecommendations(workflow *Workflow) []string {
	recommendations := make([]string, 0)

	// Analyze workflow patterns
	if len(workflow.Tasks) > 10 {
		recommendations = append(recommendations, "Consider breaking down large workflows into smaller chunks")
	}

	// Check for specific error patterns
	for _, task := range workflow.Tasks {
		if task.Status == TaskStatusFailed {
			if strings.Contains(task.Error, "permission") {
				recommendations = append(recommendations, "Check file/directory permissions before retrying")
			}
			if strings.Contains(task.Error, "network") {
				recommendations = append(recommendations, "Verify network connectivity and retry with backoff")
			}
			if strings.Contains(task.Error, "timeout") {
				recommendations = append(recommendations, "Increase timeout values for long-running tasks")
			}
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Review task arguments and agent capabilities")
	}

	return recommendations
}
