package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	chatMessage    string
	chatModel      string
	chatVerbosity  int
	chatOutputFile string
	chatQuiet      bool
	chatTimeout    int
)

// chatCmd represents the 'afe chat' command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Chat with AgentForgeEngine",
	Long:  "Send messages to AgentForgeEngine and get responses with optional verbose agent call logging",
	RunE:  runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)

	// Chat-specific flags
	chatCmd.Flags().StringVarP(&chatMessage, "message", "m", "", "Single message to send (non-interactive mode)")
	chatCmd.Flags().StringVar(&chatModel, "model", "", "Model to use (default: from config)")
	chatCmd.Flags().CountVarP(&chatVerbosity, "verbose", "v", "Verbosity level (-v, -vv, -vvv)")
	chatCmd.Flags().StringVarP(&chatOutputFile, "output", "o", "", "Save conversation to file")
	chatCmd.Flags().BoolVar(&chatQuiet, "quiet", false, "Only show final response, hide agent calls")
	chatCmd.Flags().IntVar(&chatTimeout, "timeout", 30, "Timeout in seconds for completion detection")
}

func runChat(cmd *cobra.Command, args []string) error {
	// Validate arguments
	if chatMessage == "" {
		return fmt.Errorf("interactive mode not implemented yet. Use -m \"your message\" to send a single message")
	}

	// Set verbosity level based on flag count
	verbosityLevel := getVerbosityLevel(chatVerbosity)

	if chatVerbosity > 0 {
		fmt.Printf("🤖 Starting chat with verbosity level: %s\n", verbosityLevel)
		fmt.Printf("📝 Message: %q\n", chatMessage)
		if chatModel != "" {
			fmt.Printf("🧠 Model: %s\n", chatModel)
		}
		if chatOutputFile != "" {
			fmt.Printf("💾 Output file: %s\n", chatOutputFile)
		}
		fmt.Println()
	}

	// Call the chat API
	response, err := callChatAPI(chatMessage, chatModel, chatVerbosity, chatTimeout)
	if err != nil {
		return fmt.Errorf("failed to call chat API: %w", err)
	}

	// Display the response
	displayChatResponse(response, chatVerbosity)

	// Save to file if requested
	if chatOutputFile != "" {
		if err := saveConversationToFile(response, chatMessage, chatOutputFile); err != nil {
			return fmt.Errorf("failed to save conversation: %w", err)
		}
	}

	return nil
}

// Chat API response structure
type ChatAPIResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Message       string        `json:"message"`
		FunctionCalls []interface{} `json:"function_calls"`
		Completed     bool          `json:"completed"`
		Timestamp     time.Time     `json:"timestamp"`
		Duration      string        `json:"duration"`
	} `json:"data"`
	Error string `json:"error,omitempty"`
}

// callChatAPI sends a request to the chat API
func callChatAPI(message, model string, verbosity, timeout int) (*ChatAPIResponse, error) {
	// Get server configuration (hardcoded for now, will read from config later)
	apiURL := "http://localhost:8082/api/v1/chat"

	// Prepare request payload
	payload := map[string]interface{}{
		"message":   message,
		"verbosity": verbosity,
		"timeout":   timeout,
	}

	if model != "" {
		payload["model"] = model
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var chatResp ChatAPIResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !chatResp.Success {
		return nil, fmt.Errorf("API error: %s", chatResp.Error)
	}

	return &chatResp, nil
}

// displayChatResponse formats and displays the chat response
func displayChatResponse(response *ChatAPIResponse, verbosity int) {
	if verbosity == 0 {
		// Quiet mode - just show the message
		fmt.Println(response.Data.Message)
		return
	}

	// Verbose mode
	fmt.Printf("🤖 Response: %s\n", response.Data.Message)
	fmt.Printf("⏱️  Duration: %s\n", response.Data.Duration)
	fmt.Printf("✅ Completed: %t\n", response.Data.Completed)

	if verbosity > 1 {
		fmt.Printf("🕐 Timestamp: %s\n", response.Data.Timestamp.Format("2006-01-02 15:04:05"))
	}

	if len(response.Data.FunctionCalls) > 0 {
		fmt.Printf("🔧 Function Calls: %d\n", len(response.Data.FunctionCalls))
		if verbosity > 1 {
			for i, call := range response.Data.FunctionCalls {
				fmt.Printf("  %d. %+v\n", i+1, call)
			}
		}
	}
}

// saveConversationToFile saves the conversation to a file
func saveConversationToFile(response *ChatAPIResponse, userMessage, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Write conversation header
	fmt.Fprintf(file, "# AgentForgeEngine Chat Conversation\n")
	fmt.Fprintf(file, "# Generated: %s\n\n", timestamp)

	// Write user message
	fmt.Fprintf(file, "## User Message\n%s\n\n", userMessage)

	// Write AI response
	fmt.Fprintf(file, "## AI Response\n%s\n\n", response.Data.Message)

	// Write metadata
	fmt.Fprintf(file, "## Metadata\n")
	fmt.Fprintf(file, "- Duration: %s\n", response.Data.Duration)
	fmt.Fprintf(file, "- Completed: %t\n", response.Data.Completed)
	fmt.Fprintf(file, "- Timestamp: %s\n", response.Data.Timestamp.Format(time.RFC3339))

	if len(response.Data.FunctionCalls) > 0 {
		fmt.Fprintf(file, "- Function Calls: %d\n", len(response.Data.FunctionCalls))
	}

	return nil
}

// getVerbosityLevel converts verbosity count to string
func getVerbosityLevel(count int) string {
	switch count {
	case 0:
		return "normal"
	case 1:
		return "basic (-v)"
	case 2:
		return "detailed (-vv)"
	default:
		return "debug (-vvv)"
	}
}

// saveConversation saves the conversation to a file
func saveConversation(conversation string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(file, "# AgentForgeEngine Chat Conversation\n")
	_, err = fmt.Fprintf(file, "# Generated: %s\n\n", timestamp)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	_, err = file.WriteString(conversation)
	if err != nil {
		return fmt.Errorf("failed to write conversation: %w", err)
	}

	return nil
}
