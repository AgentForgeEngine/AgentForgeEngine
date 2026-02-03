package main

import (
	"context"
	"testing"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

func TestTaskAgent_ExecuteTask(t *testing.T) {
	agent := NewTaskAgent()

	ctx := context.Background()

	// Test successful command execution
	input := interfaces.AgentInput{
		Type: "execute",
		Payload: map[string]interface{}{
			"command": "echo",
			"args":    []string{"hello", "world"},
		},
	}

	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !output.Success {
		t.Error("Expected successful execution")
	}

	taskID, ok := output.Data["task_id"].(string)
	if !ok {
		t.Fatal("Expected task_id in output")
	}

	// Wait for task to complete
	time.Sleep(100 * time.Millisecond)

	// Check task status
	statusInput := interfaces.AgentInput{
		Type: "status",
		Payload: map[string]interface{}{
			"task_id": taskID,
		},
	}

	statusOutput, err := agent.Process(ctx, statusInput)
	if err != nil {
		t.Fatalf("Status check failed: %v", err)
	}

	if !statusOutput.Success {
		t.Error("Expected successful status check")
	}

	taskData := statusOutput.Data
	status, ok := taskData["status"].(string)
	if !ok {
		t.Fatal("Expected status string")
	}

	if status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", status)
	}
}

func TestTaskAgent_ListTasks(t *testing.T) {
	agent := NewTaskAgent()

	ctx := context.Background()

	// Start a task
	input := interfaces.AgentInput{
		Type: "execute",
		Payload: map[string]interface{}{
			"command": "echo",
			"args":    []string{"test"},
		},
	}

	_, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// List tasks
	listInput := interfaces.AgentInput{
		Type: "list",
	}

	output, err := agent.Process(ctx, listInput)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if !output.Success {
		t.Error("Expected successful list")
	}

	taskData := output.Data
	tasks, ok := taskData["tasks"].([]interface{})
	if !ok {
		t.Fatal("Expected tasks array")
	}

	if len(tasks) == 0 {
		t.Error("Expected at least one task")
	}
}

func TestTaskAgent_CancelTask(t *testing.T) {
	agent := NewTaskAgent()

	ctx := context.Background()

	// Start a long-running task
	input := interfaces.AgentInput{
		Type: "execute",
		Payload: map[string]interface{}{
			"command": "sleep",
			"args":    []string{"10"},
		},
	}

	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	taskID, ok := output.Data["task_id"].(string)
	if !ok {
		t.Fatal("Expected task_id in output")
	}

	// Cancel task
	cancelInput := interfaces.AgentInput{
		Type: "cancel",
		Payload: map[string]interface{}{
			"task_id": taskID,
		},
	}

	cancelOutput, err := agent.Process(ctx, cancelInput)
	if err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	if !cancelOutput.Success {
		t.Error("Expected successful cancellation")
	}

	// Check status
	time.Sleep(50 * time.Millisecond)

	statusInput := interfaces.AgentInput{
		Type: "status",
		Payload: map[string]interface{}{
			"task_id": taskID,
		},
	}

	statusOutput, err := agent.Process(ctx, statusInput)
	if err != nil {
		t.Fatalf("Status check failed: %v", err)
	}

	if !statusOutput.Success {
		t.Error("Expected successful status check")
	}

	taskData := statusOutput.Data
	status, ok := taskData["status"].(string)
	if !ok {
		t.Fatal("Expected status string")
	}

	if status != "cancelled" {
		t.Errorf("Expected status 'cancelled', got '%s'", status)
	}
}

func TestTaskAgent_Initialize(t *testing.T) {
	agent := NewTaskAgent()

	// Test with config
	config := map[string]interface{}{
		"max_concurrent_tasks": 3,
	}

	err := agent.Initialize(config)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	// Test without config
	agent2 := NewTaskAgent()
	err = agent2.Initialize(nil)
	if err != nil {
		t.Errorf("Initialize failed with nil config: %v", err)
	}
}

func TestTaskAgent_HealthCheck(t *testing.T) {
	agent := NewTaskAgent()

	err := agent.HealthCheck()
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestTaskAgent_Shutdown(t *testing.T) {
	agent := NewTaskAgent()

	// Start a task
	ctx := context.Background()
	input := interfaces.AgentInput{
		Type: "execute",
		Payload: map[string]interface{}{
			"command": "echo",
			"args":    []string{"test"},
		},
	}

	_, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Shutdown
	err = agent.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Check that tasks are marked as cancelled
	listInput := interfaces.AgentInput{
		Type: "list",
	}

	output, err := agent.Process(ctx, listInput)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	taskData := output.Data
	tasks, ok := taskData["tasks"].([]interface{})
	if !ok {
		t.Fatal("Expected tasks array")
	}

	for _, taskInterface := range tasks {
		task, ok := taskInterface.(map[string]interface{})
		if !ok {
			continue
		}

		status, ok := task["status"].(string)
		if ok && status == "running" {
			t.Error("Expected no running tasks after shutdown")
			break
		}
	}
}
