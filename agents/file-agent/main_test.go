package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

func TestFileAgent_ReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	agent := NewFileAgent()

	config := map[string]interface{}{
		"working_directory": tmpDir,
	}
	err := agent.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!"
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	input := interfaces.AgentInput{
		Type: "read",
		Payload: map[string]interface{}{
			"filename": "test.txt",
		},
	}

	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !output.Success {
		t.Errorf("Expected successful read, got error: %s", output.Error)
	}

	fileContent, ok := output.Data["content"].(string)
	if !ok {
		t.Fatal("Expected content in output data")
	}

	if fileContent != content {
		t.Errorf("Expected content '%s', got '%s'", content, fileContent)
	}
}

func TestFileAgent_WriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	agent := NewFileAgent()

	config := map[string]interface{}{
		"working_directory": tmpDir,
	}
	err := agent.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	ctx := context.Background()
	content := "Test content for writing"
	input := interfaces.AgentInput{
		Type: "write",
		Payload: map[string]interface{}{
			"filename": "write-test.txt",
			"content":  content,
		},
	}

	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !output.Success {
		t.Errorf("Expected successful write, got error: %s", output.Error)
	}

	// Verify file was written
	testFile := filepath.Join(tmpDir, "write-test.txt")
	writtenContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(writtenContent) != content {
		t.Errorf("Expected written content '%s', got '%s'", content, string(writtenContent))
	}
}

func TestFileAgent_ListFiles(t *testing.T) {
	tmpDir := t.TempDir()
	agent := NewFileAgent()

	config := map[string]interface{}{
		"working_directory": tmpDir,
	}
	err := agent.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test files
	err = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "file2.md"), []byte("content2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create test directory
	err = os.Mkdir(filepath.Join(tmpDir, "testdir"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	ctx := context.Background()
	input := interfaces.AgentInput{
		Type: "list",
	}

	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !output.Success {
		t.Errorf("Expected successful list, got error: %s", output.Error)
	}

	data := output.Data
	files, ok := data["files"].([]interface{})
	if !ok {
		t.Fatal("Expected files array in output data")
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	dirs, ok := data["dirs"].([]interface{})
	if !ok {
		t.Fatal("Expected dirs array in output data")
	}

	if len(dirs) != 1 {
		t.Errorf("Expected 1 directory, got %d", len(dirs))
	}
}

func TestFileAgent_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	agent := NewFileAgent()

	config := map[string]interface{}{
		"working_directory": tmpDir,
	}
	err := agent.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test file
	testFile := filepath.Join(tmpDir, "exists.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()

	// Test existing file
	input := interfaces.AgentInput{
		Type: "exists",
		Payload: map[string]interface{}{
			"filename": "exists.txt",
		},
	}

	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !output.Success {
		t.Error("Expected successful exists check")
	}

	exists, ok := output.Data["exists"].(bool)
	if !ok {
		t.Fatal("Expected exists boolean in output data")
	}

	if !exists {
		t.Error("Expected file to exist")
	}

	// Test non-existing file
	input = interfaces.AgentInput{
		Type: "exists",
		Payload: map[string]interface{}{
			"filename": "nonexistent.txt",
		},
	}

	output, err = agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !output.Success {
		t.Error("Expected successful exists check for non-existent file")
	}

	exists, ok = output.Data["exists"].(bool)
	if !ok {
		t.Fatal("Expected exists boolean in output data")
	}

	if exists {
		t.Error("Expected file to not exist")
	}
}

func TestFileAgent_DeleteFile(t *testing.T) {
	tmpDir := t.TempDir()
	agent := NewFileAgent()

	config := map[string]interface{}{
		"working_directory": tmpDir,
	}
	err := agent.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test file
	testFile := filepath.Join(tmpDir, "delete.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	input := interfaces.AgentInput{
		Type: "delete",
		Payload: map[string]interface{}{
			"filename": "delete.txt",
		},
	}

	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !output.Success {
		t.Errorf("Expected successful delete, got error: %s", output.Error)
	}

	// Verify file was deleted
	_, err = os.Stat(testFile)
	if os.IsNotExist(err) {
		// Expected
	} else {
		t.Error("Expected file to be deleted")
	}
}

func TestFileAgent_FileInfo(t *testing.T) {
	tmpDir := t.TempDir()
	agent := NewFileAgent()

	config := map[string]interface{}{
		"working_directory": tmpDir,
	}
	err := agent.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test file
	testFile := filepath.Join(tmpDir, "info.txt")
	content := "Test content for info"
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	input := interfaces.AgentInput{
		Type: "info",
		Payload: map[string]interface{}{
			"filename": "info.txt",
		},
	}

	output, err := agent.Process(ctx, input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !output.Success {
		t.Errorf("Expected successful info, got error: %s", output.Error)
	}

	data := output.Data
	size, ok := data["size"].(int64)
	if !ok {
		t.Fatal("Expected size int64 in output data")
	}

	if size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), size)
	}

	filename, ok := data["filename"].(string)
	if !ok {
		t.Fatal("Expected filename string in output data")
	}

	if filename != "info.txt" {
		t.Errorf("Expected filename 'info.txt', got '%s'", filename)
	}
}

func TestFileAgent_Initialize(t *testing.T) {
	agent := NewFileAgent()

	// Test with config
	config := map[string]interface{}{
		"max_file_size":      1000,
		"allowed_extensions": []string{".txt", ".md"},
		"working_directory":  "/tmp",
	}

	err := agent.Initialize(config)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	if agent.maxFileSize != 1000 {
		t.Errorf("Expected maxFileSize 1000, got %d", agent.maxFileSize)
	}

	if len(agent.allowedExtensions) != 2 {
		t.Errorf("Expected 2 allowed extensions, got %d", len(agent.allowedExtensions))
	}

	if agent.workingDirectory != "/tmp" {
		t.Errorf("Expected working directory '/tmp', got '%s'", agent.workingDirectory)
	}
}

func TestFileAgent_HealthCheck(t *testing.T) {
	tmpDir := t.TempDir()
	agent := NewFileAgent()

	config := map[string]interface{}{
		"working_directory": tmpDir,
	}
	err := agent.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = agent.HealthCheck()
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// Test with non-existent directory
	agent2 := NewFileAgent()
	config2 := map[string]interface{}{
		"working_directory": "/non/existent/directory",
	}
	err = agent2.Initialize(config2)
	if err == nil {
		t.Fatal("Expected health check to fail with non-existent directory")
	}
}
