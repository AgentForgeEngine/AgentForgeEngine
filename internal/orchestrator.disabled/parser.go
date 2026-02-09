package orchestrator

import (
	"fmt"
	"regexp"
	"strings"
)

// TodoPattern represents a pattern for matching todo items to agents
type TodoPattern struct {
	AgentName string
	Pattern   *regexp.Regexp
	Extractor func(matches []string) map[string]interface{}
}

// TodoParserImpl implements TodoParser interface
type TodoParserImpl struct {
	patterns []TodoPattern
}

// NewTodoParser creates a new todo parser with default patterns
func NewTodoParser() TodoParser {
	return &TodoParserImpl{
		patterns: getDefaultTodoPatterns(),
	}
}

// ParseTodo parses a single todo item into a ParsedTodo
func (tp *TodoParserImpl) ParseTodo(todo string) (*ParsedTodo, error) {
	// Remove checkbox prefix and clean
	cleaned := strings.TrimSpace(todo)
	if strings.HasPrefix(cleaned, "[") && strings.Contains(cleaned, "]") {
		if idx := strings.Index(cleaned, "]"); idx != -1 {
			cleaned = strings.TrimSpace(cleaned[idx+1:])
		}
	}

	// Try each pattern
	for _, pattern := range tp.patterns {
		if matches := pattern.Pattern.FindStringSubmatch(cleaned); matches != nil {
			args := pattern.Extractor(matches)

			return &ParsedTodo{
				Original:   todo,
				Cleaned:    cleaned,
				AgentName:  pattern.AgentName,
				Arguments:  args,
				Confidence: 1.0,
			}, nil
		}
	}

	// Fallback: treat as generic task
	return &ParsedTodo{
		Original:   todo,
		Cleaned:    cleaned,
		AgentName:  "unknown",
		Arguments:  map[string]interface{}{"task": cleaned},
		Confidence: 0.1,
	}, nil
}

// ParseMultiple parses multiple todo items
func (tp *TodoParserImpl) ParseMultiple(todos []string) ([]*ParsedTodo, error) {
	var parsed []*ParsedTodo

	for i, todo := range todos {
		parsedTodo, err := tp.ParseTodo(todo)
		if err != nil {
			return nil, fmt.Errorf("failed to parse todo %d: %w", i, err)
		}
		parsed = append(parsed, parsedTodo)
	}

	return parsed, nil
}

// getDefaultTodoPatterns returns the default todo patterns for common development tasks
func getDefaultTodoPatterns() []TodoPattern {
	return []TodoPattern{
		{
			AgentName: "ls-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:list|ls|show|display)\s+(?:the\s+)?(?:project\s+)?directory`),
			Extractor: func(matches []string) map[string]interface{} {
				return map[string]interface{}{
					"path": ".",
				}
			},
		},
		{
			AgentName: "ls-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:list|ls)\s+(.+?)\s*(?:directory|folder)?`),
			Extractor: func(matches []string) map[string]interface{} {
				return map[string]interface{}{
					"path": strings.TrimSpace(matches[1]),
				}
			},
		},
		{
			AgentName: "grep-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:grep|search)\s+(?:for\s+)?(['"]?[^'"]+['"]?)\s+(?:go\s+)?files?`),
			Extractor: func(matches []string) map[string]interface{} {
				pattern := strings.Trim(matches[1], `"' `)
				return map[string]interface{}{
					"pattern": pattern,
					"path":    ".",
				}
			},
		},
		{
			AgentName: "grep-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:find|grep|search)\s+(?:for\s+)?(['"]?[^'"]+['"]?)\s+in\s+(['"]?[^'"]+['"]?)`),
			Extractor: func(matches []string) map[string]interface{} {
				pattern := strings.Trim(matches[1], `"' `)
				path := strings.Trim(matches[2], `"' `)
				return map[string]interface{}{
					"pattern": pattern,
					"path":    path,
				}
			},
		},
		{
			AgentName: "touch-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:create|make|new)\s+(?:go\s+)?(?:mod\s+)?file\s+(.+)`),
			Extractor: func(matches []string) map[string]interface{} {
				filename := strings.TrimSpace(matches[1])
				if filename == "go mod" || filename == "go.mod" {
					filename = "go.mod"
				}
				return map[string]interface{}{
					"file": filename,
				}
			},
		},
		{
			AgentName: "touch-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:create|make|new|touch)\s+(.+)`),
			Extractor: func(matches []string) map[string]interface{} {
				return map[string]interface{}{
					"file": strings.TrimSpace(matches[1]),
				}
			},
		},
		{
			AgentName: "mkdir-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:create|make)\s+(?:new\s+)?(?:directory|folder|dir)\s+(.+)`),
			Extractor: func(matches []string) map[string]interface{} {
				return map[string]interface{}{
					"path": strings.TrimSpace(matches[1]),
				}
			},
		},
		{
			AgentName: "file-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:implement|write|code)\s+(.+)`),
			Extractor: func(matches []string) map[string]interface{} {
				target := strings.TrimSpace(matches[1])
				return map[string]interface{}{
					"operation": "write",
					"target":    target,
				}
			},
		},
		{
			AgentName: "file-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:edit|modify|update)\s+(.+)`),
			Extractor: func(matches []string) map[string]interface{} {
				target := strings.TrimSpace(matches[1])
				return map[string]interface{}{
					"operation": "edit",
					"target":    target,
				}
			},
		},
		{
			AgentName: "web-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:fetch|download|get)\s+(?:from\s+)?(https?://[^\s]+)`),
			Extractor: func(matches []string) map[string]interface{} {
				url := strings.TrimSpace(matches[1])
				return map[string]interface{}{
					"type":             "fetch",
					"url":              url,
					"extract_content":  true,
					"include_metadata": true,
				}
			},
		},
		{
			AgentName: "web-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:search|look up)\s+(?:for\s+)?(.+)\s+(?:online|web|internet)`),
			Extractor: func(matches []string) map[string]interface{} {
				query := strings.TrimSpace(matches[1])
				return map[string]interface{}{
					"type":   "search",
					"query":  query,
					"source": "web",
				}
			},
		},
		{
			AgentName: "echo-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:print|echo|display|show)\s+(.+?)\s*(?:to\s+console|log)?`),
			Extractor: func(matches []string) map[string]interface{} {
				message := strings.TrimSpace(matches[1])
				return map[string]interface{}{
					"message": message,
					"file":    "",
				}
			},
		},
		{
			AgentName: "chat-agent",
			Pattern:   regexp.MustCompile(`(?i)^(?:ask|chat|talk to|discuss)\s+(.+)`),
			Extractor: func(matches []string) map[string]interface{} {
				message := strings.TrimSpace(matches[1])
				return map[string]interface{}{
					"message": message,
				}
			},
		},
	}
}
