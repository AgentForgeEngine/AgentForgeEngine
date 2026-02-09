package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/loader"
	"github.com/AgentForgeEngine/AgentForgeEngine/internal/models"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"

	"github.com/AgentForgeEngine/AgentForgeEngine/internal/response"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/status"
	"github.com/gorilla/websocket"
)

// Server represents the HTTP API server
type Server struct {
	port       int
	host       string
	router     *http.ServeMux
	wsUpgrader websocket.Upgrader
	wsClients  map[*websocket.Conn]bool
	wsMutex    sync.RWMutex

	// AFE components
	statusManager *status.Manager
	pluginManager *loader.Manager
	modelManager  *models.Manager
	// orchestratorManager *orchestrator.Manager // Disabled for now
	formatter *response.XMLFormatter
}

// NewServer creates a new API server instance
func NewServer(host string, port int) *Server {
	return &Server{
		host:   host,
		port:   port,
		router: http.NewServeMux(),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow same origin for now
			},
		},
		wsClients: make(map[*websocket.Conn]bool),
		formatter: response.NewXMLFormatter(),
	}
}

// SetComponents sets the AFE components for the server
func (s *Server) SetComponents(statusMgr *status.Manager, pluginMgr *loader.Manager, modelMgr *models.Manager) {
	s.statusManager = statusMgr
	s.pluginManager = pluginMgr
	s.modelManager = modelMgr
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Status endpoints
	s.router.HandleFunc("/api/v1/status", s.handleStatus)
	s.router.HandleFunc("/api/v1/health", s.handleHealth)

	// Chat endpoints
	s.router.HandleFunc("/api/v1/chat", s.handleChat)

	// Agent endpoints
	s.router.HandleFunc("/api/v1/agents", s.handleListAgents)
	s.router.HandleFunc("/api/v1/agents/", s.handleCallAgent)

	// Log endpoints
	s.router.HandleFunc("/api/v1/logs", s.handleGetLogs)

	// System control endpoints
	s.router.HandleFunc("/api/v1/start", s.handleStart)
	s.router.HandleFunc("/api/v1/stop", s.handleStop)

	// WebSocket endpoint for real-time events
	s.router.HandleFunc("/api/v1/events", s.handleWebSocket)
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs all requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("API Request: %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("API Response: %s %s - %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// wrapHandler adds CORS and logging to handlers
func (s *Server) wrapHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Log request
		start := time.Now()
		log.Printf("API Request: %s %s", r.Method, r.URL.Path)

		// Call handler
		handler(w, r)

		log.Printf("API Response: %s %s - %v", r.Method, r.URL.Path, time.Since(start))
	}
}

// wrapHandlers creates a new router with all handlers wrapped
func (s *Server) wrapHandlers() http.Handler {
	wrappedRouter := http.NewServeMux()

	// Wrap all handlers
	wrappedRouter.HandleFunc("/api/v1/status", s.wrapHandler(s.handleStatus))
	wrappedRouter.HandleFunc("/api/v1/health", s.wrapHandler(s.handleHealth))
	wrappedRouter.HandleFunc("/api/v1/chat", s.wrapHandler(s.handleChat))
	wrappedRouter.HandleFunc("/api/v1/agents", s.wrapHandler(s.handleListAgents))
	wrappedRouter.HandleFunc("/api/v1/agents/", s.wrapHandler(s.handleCallAgent))
	wrappedRouter.HandleFunc("/api/v1/logs", s.wrapHandler(s.handleGetLogs))
	wrappedRouter.HandleFunc("/api/v1/start", s.wrapHandler(s.handleStart))
	wrappedRouter.HandleFunc("/api/v1/stop", s.wrapHandler(s.handleStop))
	wrappedRouter.HandleFunc("/api/v1/events", s.handleWebSocket)

	return wrappedRouter
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.setupRoutes()

	// Wrap all handlers with middleware
	wrappedRouter := s.wrapHandlers()

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	server := &http.Server{
		Addr:    addr,
		Handler: wrappedRouter,
	}

	log.Printf("API Server starting on %s", addr)

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("API Server error: %v", err)
		}
	}()

	// Handle shutdown
	<-ctx.Done()
	log.Println("Shutting down API Server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}

// BroadcastWebSocket sends a message to all connected WebSocket clients
func (s *Server) BroadcastWebSocket(message interface{}) {
	s.wsMutex.RLock()
	defer s.wsMutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}

	for client := range s.wsClients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			delete(s.wsClients, client)
		}
	}
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	s.wsMutex.Lock()
	s.wsClients[conn] = true
	s.wsMutex.Unlock()

	log.Printf("WebSocket client connected: %s", conn.RemoteAddr())

	// Send welcome message
	s.sendToClient(conn, map[string]interface{}{
		"type":      "welcome",
		"message":   "Connected to AgentForgeEngine API",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket client disconnected: %v", err)
			break
		}
	}

	s.wsMutex.Lock()
	delete(s.wsClients, conn)
	s.wsMutex.Unlock()
}

// sendToClient sends a message to a specific WebSocket client
func (s *Server) sendToClient(conn *websocket.Conn, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("WebSocket write error: %v", err)
	}
}

// API Response helpers
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (s *Server) sendJSON(w http.ResponseWriter, status int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func (s *Server) sendSuccess(w http.ResponseWriter, data interface{}) {
	s.sendJSON(w, http.StatusOK, APIResponse{Success: true, Data: data})
}

func (s *Server) sendError(w http.ResponseWriter, status int, message string) {
	s.sendJSON(w, status, APIResponse{Success: false, Error: message})
}

// Chat request/response structures
type ChatRequest struct {
	Message   string                 `json:"message"`
	Model     string                 `json:"model,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
	Verbosity int                    `json:"verbosity,omitempty"`
	Timeout   int                    `json:"timeout,omitempty"`
}

type ChatResponse struct {
	Message       string         `json:"message"`
	FunctionCalls []FunctionCall `json:"function_calls,omitempty"`
	Completed     bool           `json:"completed"`
	Timestamp     time.Time      `json:"timestamp"`
	Duration      string         `json:"duration"`
}

type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	Response  *FunctionResponse      `json:"response,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  string                 `json:"duration"`
}

type FunctionResponse struct {
	Name        string                 `json:"name"`
	Data        map[string]interface{} `json:"data"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	RawResponse string                 `json:"raw_response,omitempty"`
}

// API Handler Methods

// handleStatus returns the current engine status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if s.statusManager == nil {
		s.sendError(w, http.StatusInternalServerError, "Status manager not initialized")
		return
	}

	// Try to get detailed status via socket
	statusInfo, err := s.statusManager.GetStatusViaSocket()
	if err != nil {
		// Fallback to basic status
		statusInfo = s.statusManager.GetBasicStatus()
	}

	s.sendSuccess(w, statusInfo)
}

// handleHealth performs a health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.sendSuccess(w, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}

// handleChat processes chat messages
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.sendError(w, http.StatusMethodNotAllowed, "Only POST method allowed")
		return
	}

	// Parse request body
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	// Validate request
	if req.Message == "" {
		s.sendError(w, http.StatusBadRequest, "Message field is required")
		return
	}

	// Use model manager for real model integration
	startTime := time.Now()

	// Broadcast chat start event
	s.BroadcastWebSocket(map[string]interface{}{
		"type":      "chat_start",
		"message":   req.Message,
		"model":     req.Model,
		"timestamp": startTime,
	})

	// Check if model manager is available
	if s.modelManager == nil {
		s.sendError(w, http.StatusInternalServerError, "Model manager not initialized")
		return
	}

	// Default to llamacpp model, or use request model
	modelName := req.Model
	if modelName == "" {
		modelName = "llamacpp"
	}

	// Create generation request
	genReq := interfaces.GenerationRequest{
		Prompt:      req.Message,
		MaxTokens:   8000,
		Temperature: 0.7,
		Stream:      false,
	}

	// Call the model
	modelResponse, err := s.modelManager.Generate(r.Context(), modelName, genReq)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Model generation failed: %v", err))
		return
	}

	// Parse function calls from response
	var functionCalls []FunctionCall
	if modelResponse.Text != "" && strings.Contains(modelResponse.Text, "<function_call") {
		calls, err := s.parseFunctionCalls(modelResponse.Text)
		if err == nil {
			// Execute function calls with safety check
			s.executeFunctionCalls(calls)
			functionCalls = calls
		}
	}

	// Create response
	response := ChatResponse{
		Message:       modelResponse.Text,
		FunctionCalls: functionCalls,
		Completed:     modelResponse.Finished,
		Timestamp:     time.Now(),
		Duration:      time.Since(startTime).String(),
	}

	// Broadcast completion event
	s.BroadcastWebSocket(map[string]interface{}{
		"type":      "chat_complete",
		"message":   response.Message,
		"completed": response.Completed,
		"timestamp": response.Timestamp,
	})

	s.sendSuccess(w, response)
}

// parseFunctionCalls parses function calls from model response text
func (s *Server) parseFunctionCalls(text string) ([]FunctionCall, error) {
	var calls []FunctionCall
	pattern := regexp.MustCompile(`<function_call name="([^"]+)">(.*?)</function_call>`)
	matches := pattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(match[2]), &args); err == nil {
				calls = append(calls, FunctionCall{
					Name:      match[1],
					Arguments: args,
					Timestamp: time.Now(),
				})
			}
		}
	}
	return calls, nil
}

// executeFunctionCalls executes parsed function calls via agents
func (s *Server) executeFunctionCalls(functionCalls []FunctionCall) {
	if s.pluginManager == nil {
		return
	}

	for i := range functionCalls {
		call := &functionCalls[i]

		// Safety check - only allow safe commands
		if !s.isSafeCommand(call.Name, call.Arguments) {
			call.Response = &FunctionResponse{
				Name:    call.Name,
				Success: false,
				Error:   "Command not allowed for safety reasons",
			}
			continue
		}

		start := time.Now()
		agent, exists := s.pluginManager.GetAgent(call.Name)
		if !exists {
			call.Response = &FunctionResponse{
				Name:    call.Name,
				Success: false,
				Error:   fmt.Sprintf("Agent %s not found", call.Name),
			}
			call.Duration = time.Since(start).String()
			continue
		}

		// Execute agent
		agentInput := interfaces.AgentInput{
			Type:    "execute",
			Payload: call.Arguments,
		}

		output, err := agent.Process(context.Background(), agentInput)
		call.Duration = time.Since(start).String()

		if err != nil {
			call.Response = &FunctionResponse{
				Name:    call.Name,
				Success: false,
				Error:   err.Error(),
			}
		} else {
			call.Response = &FunctionResponse{
				Name:    call.Name,
				Success: output.Success,
				Data:    output.Data,
				Error:   output.Error,
			}
		}
	}
}

// isSafeCommand checks if a command is safe to execute
func (s *Server) isSafeCommand(agentName string, args map[string]interface{}) bool {
	// Whitelist of safe commands as requested
	safeCommands := map[string]bool{
		"ls":     true,
		"cat":    true,
		"pwd":    true,
		"whoami": true,
		"df":     true,
		"uname":  true,
	}

	return safeCommands[agentName]
}

// handleListAgents lists available agents
func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	if s.pluginManager == nil {
		s.sendError(w, http.StatusInternalServerError, "Plugin manager not initialized")
		return
	}

	agents := s.pluginManager.ListAgents()
	s.sendSuccess(w, map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	})
}

// handleCallAgent calls a specific agent (placeholder for now)
func (s *Server) handleCallAgent(w http.ResponseWriter, r *http.Request) {
	// Placeholder - we'll implement this later
	s.sendError(w, http.StatusNotImplemented, "Agent call endpoint not yet implemented")
}

// handleGetLogs retrieves system logs (placeholder for now)
func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	// Placeholder - we'll implement this later
	s.sendError(w, http.StatusNotImplemented, "Logs endpoint not yet implemented")
}

// handleStart starts the engine (placeholder for now)
func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	// Placeholder - we'll implement this later
	s.sendError(w, http.StatusNotImplemented, "Start endpoint not yet implemented")
}

// handleStop stops the engine (placeholder for now)
func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	// Placeholder - we'll implement this later
	s.sendError(w, http.StatusNotImplemented, "Stop endpoint not yet implemented")
}
