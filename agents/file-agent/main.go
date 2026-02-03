package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type FileAgent struct {
	name              string
	allowedExtensions []string
	maxFileSize       int64
	workingDirectory  string
}

func NewFileAgent() *FileAgent {
	return &FileAgent{
		name:              "file-agent",
		allowedExtensions: []string{".txt", ".md", ".go", ".yaml", ".json"},
		maxFileSize:       10 * 1024 * 1024, // 10MB
		workingDirectory:  ".",
	}
}

func (fa *FileAgent) Name() string {
	return fa.name
}

func (fa *FileAgent) Initialize(config map[string]interface{}) error {
	// Set allowed extensions from config
	if extensions, ok := config["allowed_extensions"].([]interface{}); ok {
		var exts []string
		for _, ext := range extensions {
			if extStr, ok := ext.(string); ok {
				exts = append(exts, extStr)
			}
		}
		if len(exts) > 0 {
			fa.allowedExtensions = exts
		}
	}

	// Set max file size from config
	if maxSize, ok := config["max_file_size"].(int); ok {
		fa.maxFileSize = int64(maxSize)
	}

	// Set working directory from config
	if workingDir, ok := config["working_directory"].(string); ok {
		fa.workingDirectory = workingDir
	}

	return nil
}

func (fa *FileAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	switch input.Type {
	case "read":
		return fa.readFile(ctx, input)
	case "write":
		return fa.writeFile(ctx, input)
	case "list":
		return fa.listFiles(ctx, input)
	case "exists":
		return fa.fileExists(ctx, input)
	case "delete":
		return fa.deleteFile(ctx, input)
	case "info":
		return fa.fileInfo(ctx, input)
	default:
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("unknown file operation: %s", input.Type),
		}, nil
	}
}

func (fa *FileAgent) HealthCheck() error {
	// Check if working directory exists and is accessible
	if _, err := os.Stat(fa.workingDirectory); os.IsNotExist(err) {
		return fmt.Errorf("working directory does not exist: %s", fa.workingDirectory)
	}

	// Test read/write permissions
	testFile := filepath.Join(fa.workingDirectory, ".health_check")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("cannot write to working directory: %w", err)
	}

	os.Remove(testFile)
	return nil
}

func (fa *FileAgent) Shutdown() error {
	// No cleanup needed
	return nil
}

func (fa *FileAgent) readFile(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	filename, ok := input.Payload["filename"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "filename not specified in payload",
		}, nil
	}

	filePath := filepath.Join(fa.workingDirectory, filename)

	// Check file extension
	if !fa.isAllowedExtension(filePath) {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("file extension not allowed: %s", filepath.Ext(filePath)),
		}, nil
	}

	// Check file size
	if info, err := os.Stat(filePath); err == nil {
		if info.Size() > fa.maxFileSize {
			return interfaces.AgentOutput{
				Success: false,
				Error:   fmt.Sprintf("file too large: %d bytes (max: %d)", info.Size(), fa.maxFileSize),
			}, nil
		}
	} else {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("file access error: %w", err),
		}, nil
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %w", err),
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"filename": filename,
			"content":  string(content),
			"size":     len(content),
		},
	}, nil
}

func (fa *FileAgent) writeFile(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	filename, ok := input.Payload["filename"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "filename not specified in payload",
		}, nil
	}

	content, ok := input.Payload["content"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "content not specified in payload",
		}, nil
	}

	filePath := filepath.Join(fa.workingDirectory, filename)

	// Check file extension
	if !fa.isAllowedExtension(filePath) {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("file extension not allowed: %s", filepath.Ext(filePath)),
		}, nil
	}

	// Check file size
	contentSize := int64(len(content))
	if contentSize > fa.maxFileSize {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("content too large: %d bytes (max: %d)", contentSize, fa.maxFileSize),
		}, nil
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %w", err),
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"filename": filename,
			"size":     contentSize,
			"written":  true,
		},
	}, nil
}

func (fa *FileAgent) listFiles(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	directory, ok := input.Payload["directory"].(string)
	if !ok {
		directory = fa.workingDirectory
	} else {
		directory = filepath.Join(fa.workingDirectory, directory)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to list directory: %w", err),
		}, nil
	}

	var files []map[string]interface{}
	var dirs []map[string]interface{}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		item := map[string]interface{}{
			"name":    info.Name(),
			"size":    info.Size(),
			"mode":    info.Mode().String(),
			"modtime": info.ModTime(),
		}

		if entry.IsDir() {
			dirs = append(dirs, item)
		} else {
			item["extension"] = filepath.Ext(info.Name())
			files = append(files, item)
		}
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"directory": directory,
			"files":     files,
			"dirs":      dirs,
			"total":     len(files) + len(dirs),
		},
	}, nil
}

func (fa *FileAgent) fileExists(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	filename, ok := input.Payload["filename"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "filename not specified in payload",
		}, nil
	}

	filePath := filepath.Join(fa.workingDirectory, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return interfaces.AgentOutput{
			Success: true,
			Data: map[string]interface{}{
				"filename": filename,
				"exists":   false,
			},
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"filename": filename,
			"exists":   true,
		},
	}, nil
}

func (fa *FileAgent) deleteFile(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	filename, ok := input.Payload["filename"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "filename not specified in payload",
		}, nil
	}

	filePath := filepath.Join(fa.workingDirectory, filename)

	if err := os.Remove(filePath); err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to delete file: %w", err),
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"filename": filename,
			"deleted":  true,
		},
	}, nil
}

func (fa *FileAgent) fileInfo(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	filename, ok := input.Payload["filename"].(string)
	if !ok {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "filename not specified in payload",
		}, nil
	}

	filePath := filepath.Join(fa.workingDirectory, filename)

	info, err := os.Stat(filePath)
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("failed to get file info: %w", err),
		}, nil
	}

	data := map[string]interface{}{
		"filename":  filename,
		"size":      info.Size(),
		"mode":      info.Mode().String(),
		"modtime":   info.ModTime(),
		"is_dir":    info.IsDir(),
		"extension": filepath.Ext(info.Name()),
	}

	return interfaces.AgentOutput{
		Success: true,
		Data:    data,
	}, nil
}

func (fa *FileAgent) isAllowedExtension(filePath string) bool {
	if len(fa.allowedExtensions) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	for _, allowedExt := range fa.allowedExtensions {
		if strings.ToLower(allowedExt) == ext {
			return true
		}
	}
	return false
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewFileAgent()
