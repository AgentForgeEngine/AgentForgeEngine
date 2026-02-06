package status

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// StatusInfo contains detailed status information
type StatusInfo struct {
	PID         int       `json:"pid"`
	StartTime   time.Time `json:"start_time"`
	Version     string    `json:"version"`
	Uptime      string    `json:"uptime"`
	Status      string    `json:"status"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	ModelsCount int       `json:"models_count"`
	AgentsCount int       `json:"agents_count"`
}

// Manager handles PID file and Unix socket for status tracking
type Manager struct {
	pidFile  string
	sockFile string
	listener net.Listener
}

// NewManager creates a new status manager
func NewManager(afeDir string) *Manager {
	return &Manager{
		pidFile:  filepath.Join(afeDir, "afe.pid"),
		sockFile: filepath.Join(afeDir, "afe.sock"),
	}
}

// WritePID writes the current process PID to file
func (m *Manager) WritePID() error {
	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)

	if err := os.WriteFile(m.pidFile, []byte(pidStr), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// ReadPID reads the PID from file
func (m *Manager) ReadPID() (int, error) {
	data, err := os.ReadFile(m.pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("PID file not found")
		}
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID format: %w", err)
	}

	return pid, nil
}

// IsRunning checks if the process with stored PID is still running
func (m *Manager) IsRunning() bool {
	pid, err := m.ReadPID()
	if err != nil {
		return false
	}

	// Check if process is running by sending signal 0
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// Cleanup removes PID file and socket
func (m *Manager) Cleanup() error {
	var errors []string

	// Remove PID file
	if err := os.Remove(m.pidFile); err != nil && !os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("PID file: %v", err))
	}

	// Remove socket file
	if err := os.Remove(m.sockFile); err != nil && !os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("socket file: %v", err))
	}

	// Close listener if open
	if m.listener != nil {
		m.listener.Close()
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %s", errors)
	}

	return nil
}

// StartSocketServer starts the Unix socket server for status queries
func (m *Manager) StartSocketServer(statusInfo *StatusInfo) error {
	// Remove existing socket file
	os.Remove(m.sockFile)

	listener, err := net.Listen("unix", m.sockFile)
	if err != nil {
		return fmt.Errorf("failed to create socket listener: %w", err)
	}

	m.listener = listener

	// Set socket permissions
	if err := os.Chmod(m.sockFile, 0755); err != nil {
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	// Start serving status requests
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// Listener closed, exit
				return
			}

			go m.handleConnection(conn, statusInfo)
		}
	}()

	return nil
}

// handleConnection handles individual socket connections
func (m *Manager) handleConnection(conn net.Conn, statusInfo *StatusInfo) {
	defer conn.Close()

	// Update uptime
	if statusInfo.StartTime.IsZero() {
		statusInfo.StartTime = time.Now()
	}
	statusInfo.Uptime = time.Since(statusInfo.StartTime).String()

	// Encode status info
	data, err := json.Marshal(statusInfo)
	if err != nil {
		conn.Write([]byte(`{"error": "failed to encode status"}`))
		return
	}

	// Send status
	conn.Write(data)
}

// GetStatusViaSocket attempts to get detailed status via Unix socket
func (m *Manager) GetStatusViaSocket() (*StatusInfo, error) {
	conn, err := net.Dial("unix", m.sockFile)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to socket: %w", err)
	}
	defer conn.Close()

	// Read response
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read from socket: %w", err)
	}

	// Parse response
	var statusInfo StatusInfo
	if err := json.Unmarshal(buffer[:n], &statusInfo); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}

	return &statusInfo, nil
}

// GetBasicStatus returns basic status using PID file only
func (m *Manager) GetBasicStatus() *StatusInfo {
	if !m.IsRunning() {
		return &StatusInfo{
			Status: "STOPPED",
		}
	}

	pid, _ := m.ReadPID()
	return &StatusInfo{
		PID:    pid,
		Status: "RUNNING",
	}
}

// GetPIDFile returns the PID file path
func (m *Manager) GetPIDFile() string {
	return m.pidFile
}

// GetSocketFile returns the socket file path
func (m *Manager) GetSocketFile() string {
	return m.sockFile
}
