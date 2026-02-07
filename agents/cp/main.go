package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type CpAgent struct {
	name string
}

func NewCpAgent() *CpAgent {
	return &CpAgent{name: "cp"}
}

func (a *CpAgent) Name() string {
	return a.name
}

func (a *CpAgent) Initialize(config map[string]interface{}) error {
	log.Printf("Initializing %s agent", a.name)
	return nil
}

func (a *CpAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract source and destination from input payload
	source, ok := input.Payload["source"].(string)
	if !ok || source == "" {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "Error: source parameter is required",
		}, nil
	}

	destination, ok := input.Payload["destination"].(string)
	if !ok || destination == "" {
		return interfaces.AgentOutput{
			Success: false,
			Error:   "Error: destination parameter is required",
		}, nil
	}

	// Check if source exists
	sourceInfo, err := os.Stat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return interfaces.AgentOutput{
				Success: false,
				Error:   fmt.Sprintf("Error: source %s does not exist", source),
			}, nil
		}
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error checking source %s: %v", source, err),
		}, nil
	}

	var copiedItems []string
	var totalSize int64

	if sourceInfo.IsDir() {
		// Copy directory recursively
		err = a.copyDirectory(source, destination, &copiedItems, &totalSize)
		if err != nil {
			return interfaces.AgentOutput{
				Success: false,
				Error:   fmt.Sprintf("Error copying directory %s to %s: %v", source, destination, err),
			}, nil
		}
	} else {
		// Copy single file
		err = a.copyFile(source, destination, &copiedItems, &totalSize)
		if err != nil {
			return interfaces.AgentOutput{
				Success: false,
				Error:   fmt.Sprintf("Error copying file %s to %s: %v", source, destination, err),
			}, nil
		}
	}

	// Get absolute paths for reporting
	absSource, _ := filepath.Abs(source)
	absDestination, _ := filepath.Abs(destination)

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"source":               source,
			"destination":          destination,
			"absolute_source":      absSource,
			"absolute_destination": absDestination,
			"type":                 map[bool]string{true: "directory", false: "file"}[sourceInfo.IsDir()],
			"copied_items":         copiedItems,
			"total_size":           totalSize,
			"success":              true,
		},
	}, nil
}

func (a *CpAgent) copyFile(src, dst string, copiedItems *[]string, totalSize *int64) error {
	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Get file info
	fileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	*copiedItems = append(*copiedItems, fmt.Sprintf("File: %s -> %s", src, dst))
	*totalSize += fileInfo.Size()

	return nil
}

func (a *CpAgent) copyDirectory(src, dst string, copiedItems *[]string, totalSize *int64) error {
	// Create destination directory
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			err = a.copyDirectory(srcPath, dstPath, copiedItems, totalSize)
			if err != nil {
				return err
			}
		} else {
			// Copy file
			err = a.copyFile(srcPath, dstPath, copiedItems, totalSize)
			if err != nil {
				return err
			}
		}
	}

	*copiedItems = append(*copiedItems, fmt.Sprintf("Directory: %s -> %s", src, dst))
	return nil
}

func (a *CpAgent) HealthCheck() error {
	return nil
}

func (a *CpAgent) Shutdown() error {
	log.Printf("Shutting down %s agent", a.name)
	return nil
}

// Export the agent for plugin loading
var Agent interfaces.Agent = NewCpAgent()
