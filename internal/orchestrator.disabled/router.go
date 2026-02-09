package orchestrator

import (
	"fmt"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

// TaskRouterImpl implements TaskRouter interface
type TaskRouterImpl struct {
	agentMappings map[string]string
	pluginMgr     interfaces.PluginManager
}

// NewTaskRouter creates a new task router
func NewTaskRouter(pluginMgr interfaces.PluginManager) *TaskRouterImpl {
	agentMappings := map[string]string{
		"ls-agent":    "ls",
		"grep-agent":  "grep",
		"touch-agent": "touch",
		"mkdir-agent": "mkdir",
		"file-agent":  "file",
		"web-agent":   "web",
		"echo-agent":  "echo",
		"chat-agent":  "chat",
	}

	return &TaskRouterImpl{
		agentMappings: agentMappings,
		pluginMgr:     pluginMgr,
	}
}

// RouteTask routes a parsed todo to the appropriate agent
func (tr *TaskRouterImpl) RouteTask(todo *ParsedTodo) (string, map[string]interface{}, error) {
	if todo == nil {
		return "", nil, fmt.Errorf("todo is nil")
	}

	// Map agent name to actual agent registration name
	agentName, exists := tr.agentMappings[todo.AgentName]
	if !exists {
		// Try exact match first
		if tr.agentExists(todo.AgentName) {
			return todo.AgentName, todo.Arguments, nil
		}
		return "", nil, fmt.Errorf("unknown agent: %s", todo.AgentName)
	}

	return agentName, todo.Arguments, nil
}

// ListAgents returns list of available agents
func (tr *TaskRouterImpl) ListAgents() []string {
	if tr.pluginMgr == nil {
		return []string{}
	}

	return tr.pluginMgr.ListAgents()
}

// GetAgentCapabilities returns capabilities of an agent
func (tr *TaskRouterImpl) GetAgentCapabilities(agentName string) ([]string, error) {
	if tr.pluginMgr == nil {
		return nil, fmt.Errorf("plugin manager not available")
	}

	_, exists := tr.pluginMgr.GetAgent(agentName)
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentName)
	}

	// For now, return basic capabilities based on agent type
	switch agentName {
	case "ls":
		return []string{"list_files", "show_directory", "list_with_flags"}, nil
	case "grep":
		return []string{"search_files", "find_patterns", "search_in_path"}, nil
	case "touch":
		return []string{"create_file", "touch_file", "create_empty_file"}, nil
	case "mkdir":
		return []string{"create_directory", "make_directory", "create_folder"}, nil
	case "cat":
		return []string{"read_file", "display_file", "show_content"}, nil
	case "echo":
		return []string{"print_message", "display_text", "log_output"}, nil
	case "web":
		return []string{"fetch_url", "search_web", "extract_content"}, nil
	case "file":
		return []string{"write_file", "edit_file", "modify_content"}, nil
	case "chat":
		return []string{"chat", "ask_questions", "discuss"}, nil
	default:
		return []string{"execute", "process", "perform_task"}, nil
	}
}

// agentExists checks if an agent is available
func (tr *TaskRouterImpl) agentExists(agentName string) bool {
	_, exists := tr.pluginMgr.GetAgent(agentName)
	return exists
}
