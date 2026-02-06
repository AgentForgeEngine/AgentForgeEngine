// agents/context-manager/main.go
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

type ContextManagerAgent struct {
	name string
}

func NewContextManagerAgent() *ContextManagerAgent {
	return &ContextManagerAgent{
		name: "context-manager",
	}
}

func (a *ContextManagerAgent) Name() string {
	return a.name
}

func (a *ContextManagerAgent) Initialize(config map[string]interface{}) error {
	return nil
}

func (a *ContextManagerAgent) Process(ctx context.Context, input interfaces.AgentInput) (interfaces.AgentOutput, error) {
	// Extract parameters from input
	path, _ := input.Payload["path"].(string)
	maxTokens, ok := input.Payload["max_tokens"].(float64)
	if !ok {
		maxTokens = 38400 // Default 15% of 256k tokens
	}

	// Process the context
	result, err := a.analyzeAndReportContext(path, int(maxTokens))
	if err != nil {
		return interfaces.AgentOutput{
			Success: false,
			Error:   fmt.Sprintf("Error processing context: %v", err),
		}, nil
	}

	return interfaces.AgentOutput{
		Success: true,
		Data: map[string]interface{}{
			"summary": result.summary,
			"token_count": result.tokenCount,
			"files_processed": result.filesProcessed,
		},
	}, nil
}

func (a *ContextManagerAgent) HealthCheck() error {
	return nil
}

type ContextResult struct {
	summary       string
	tokenCount    int
	filesProcessed int
}

func (a *ContextManagerAgent) analyzeAndReportContext(basePath string, maxTokens int) (*ContextResult, error) {
	var files []string
	var totalTokens int
	var summary strings.Builder
	var filesProcessed int

	// Walk through directory
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip binary files
		if a.isBinaryFile(path) {
			return nil
		}

		// Skip large files (larger than 1MB)
		if info.Size() > 1048576 {
			return nil
		}

		// Only process text files
		if a.isTextFile(path) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Process files in order of importance (smaller files first)
	// Sort by file size
	for _, file := range files {
		if totalTokens >= maxTokens {
			break
		}

		fileTokens, content, err := a.processFile(file, maxTokens-totalTokens)
		if err != nil {
			continue
		}

		if fileTokens > 0 {
			summary.WriteString(fmt.Sprintf("\n--- File: %s ---\n", file))
			summary.WriteString(content)
			totalTokens += fileTokens
			filesProcessed++
		}
	}

	return &ContextResult{
		summary:        summary.String(),
		tokenCount:     totalTokens,
		filesProcessed: filesProcessed,
	}, nil
}

func (a *ContextManagerAgent) isBinaryFile(filePath string) bool {
	// Try to detect binary files by looking for null bytes or common binary signatures
	file, err := os.Open(filePath)
	if err != nil {
		return true
	}
	defer file.Close()

	// Read first 1024 bytes
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return true
	}

	// Check for null bytes in the first 1024 bytes
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return true
		}
	}

	// Check for common binary file signatures
	content := string(buffer[:n])
	if strings.Contains(content, "\x00") {
		return true
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	binaryExtensions := []string{".exe", ".dll", ".so", ".dylib", ".class", ".jar", ".bin", ".dat", ".db", ".sqlite"}
	for _, binExt := range binaryExtensions {
		if ext == binExt {
			return true
		}
	}

	return false
}

func (a *ContextManagerAgent) isTextFile(filePath string) bool {
	// Check file extension for common text formats
	ext := strings.ToLower(filepath.Ext(filePath))
	textExtensions := []string{".txt", ".md", ".go", ".yaml", ".yml", ".json", ".xml", ".html", ".css", ".js", ".py", ".sh", ".rb", ".java", ".c", ".cpp", ".h", ".hpp"}
	
	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}

	// If no specific extension, try to detect by content
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 1024 bytes to check for text content
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	// Simple text detection: if there are mostly printable characters or common text patterns
	printableChars := 0
	for i := 0; i < n; i++ {
		if unicode.IsPrint(rune(buffer[i])) || buffer[i] == '\n' || buffer[i] == '\t' || buffer[i] == '\r' {
			printableChars++
		}
	}

	// If more than 70% of characters are printable, consider it text
	return float64(printableChars)/float64(n) > 0.7
}

func (a *ContextManagerAgent) processFile(filePath string, maxTokens int) (int, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, "", err
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return 0, "", err
	}

	// Estimate tokens (rough approximation)
	tokenCount := a.estimateTokens(string(content))

	// If file is too large, truncate to fit within token limit
	if tokenCount > maxTokens {
		// Truncate content to fit within token limit
		truncatedContent := a.truncateToTokens(string(content), maxTokens)
		return maxTokens, truncatedContent, nil
	}

	return tokenCount, string(content), nil
}

func (a *ContextManagerAgent) estimateTokens(content string) int {
	// Rough token estimation - 1 token ≈ 4 characters for English text
	// This is a very rough approximation and should be improved
	// More accurate methods would use actual tokenization libraries
	return len(content) / 4
}

func (a *ContextManagerAgent) truncateToTokens(content string, maxTokens int) string {
	// Truncate content to fit within token limit
	charsPerToken := 4
	maxChars := maxTokens * charsPerToken
	
	if len(content) <= maxChars {
		return content
	}

	// Truncate and add a note
	truncated := content[:maxChars]
	// Try to truncate at word boundary
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxChars/2 {
		truncated = truncated[:lastSpace]
	}
	
	return truncated + "\n... (truncated to fit token limit) ..."
}

func (a *ContextManagerAgent) HealthCheck() error {
	return nil
}

func (a *ContextManagerAgent) Shutdown() error {
	return nil
}

