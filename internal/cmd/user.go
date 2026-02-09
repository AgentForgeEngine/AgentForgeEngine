package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/auth"
	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/userdirs"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// userCmd represents the user command group
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage user accounts",
	Long: `Manage AgentForge user accounts with secure authentication.
This command provides utilities for creating, authenticating, and managing
user accounts stored in LevelDB with bcrypt password hashing.`,
}

// userCreateCmd represents the 'afe user create' command
var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user account",
	Long: `Create a new user account with email and password.
The password will be hashed using bcrypt for secure storage.`,
	RunE: runUserCreate,
}

// userLoginCmd represents the 'afe user login' command
var userLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate a user",
	Long: `Authenticate a user with email and password.
This command validates credentials and returns user information.`,
	RunE: runUserLogin,
}

// userApiKeyCmd represents the 'afe user api-key' command
var userApiKeyCmd = &cobra.Command{
	Use:   "api-key",
	Short: "Manage API keys",
	Long: `Manage API keys for user authentication.
Create, list, and revoke API keys for secure API access.`,
}

// apiKeyCreateCmd represents the 'afe user api-key create' command
var apiKeyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key",
	Long: `Create a new API key for a user account.
API keys can be used for secure API access without passwords.`,
	RunE: runAPIKeyCreate,
}

// apiKeyListCmd represents the 'afe user api-key list' command
var apiKeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	Long: `List all API keys for a user account.
Shows key details including creation date and expiration.`,
	RunE: runAPIKeyList,
}

var (
	userName      string
	userEmail     string
	userPassword  string
	userPhone     string
	apiKeyName    string
	apiKeyExpires string
	apiKeyScopes  []string
)

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userLoginCmd)
	userCmd.AddCommand(userApiKeyCmd)
	userApiKeyCmd.AddCommand(apiKeyCreateCmd)
	userApiKeyCmd.AddCommand(apiKeyListCmd)

	// User create flags
	userCreateCmd.Flags().StringVar(&userName, "name", "", "User name (required)")
	userCreateCmd.Flags().StringVar(&userEmail, "email", "", "User email (required)")
	userCreateCmd.Flags().StringVar(&userPhone, "phone", "", "User phone number (optional)")

	// User login flags
	userLoginCmd.Flags().StringVar(&userEmail, "email", "", "User email (required)")
	userLoginCmd.Flags().StringVar(&userPassword, "password", "", "User password (required)")

	// API key create flags
	apiKeyCreateCmd.Flags().StringVar(&apiKeyName, "name", "", "API key name (required)")
	apiKeyCreateCmd.Flags().StringVar(&apiKeyExpires, "expires", "", "Expiration date (optional, format: 2024-12-31)")
	apiKeyCreateCmd.Flags().StringSliceVar(&apiKeyScopes, "scopes", []string{"read"}, "API key scopes")
	apiKeyCreateCmd.Flags().StringVar(&userEmail, "email", "", "User email (required)")
}

// readPassword reads password from terminal without echoing, or from stdin if piped
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Check if stdin is a terminal (interactive)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		// Interactive mode - read password without echoing
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // Add newline after password input
		return string(password), nil
	} else {
		// Non-interactive mode - read from stdin
		reader := bufio.NewReader(os.Stdin)
		passwordStr, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}
		return strings.TrimSpace(passwordStr), nil
	}
}

// runUserCreate creates a new user account
func runUserCreate(cmd *cobra.Command, args []string) error {
	if userName == "" || userEmail == "" {
		return fmt.Errorf("name and email are required")
	}

	// Read password interactively
	password, err := readPassword("Enter password: ")
	if err != nil {
		return err
	}

	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}

	// Confirm password
	confirmPassword, err := readPassword("Confirm password: ")
	if err != nil {
		return err
	}

	if len(confirmPassword) == 0 {
		return fmt.Errorf("password confirmation cannot be empty")
	}

	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	userPassword = password

	// Initialize user directories
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return fmt.Errorf("failed to create user directories: %w", err)
	}

	// Create accounts directory if it doesn't exist
	accountsDir := filepath.Join(userDirs.AFEDir, "accounts")
	if err := os.MkdirAll(accountsDir, 0700); err != nil {
		return fmt.Errorf("failed to create accounts directory: %w", err)
	}

	// Create user manager
	userManager, err := auth.NewUserManager(accountsDir)
	if err != nil {
		return fmt.Errorf("failed to create user manager: %w", err)
	}
	defer userManager.Close()

	// Create user
	var phoneNumber *string
	if userPhone != "" {
		phoneNumber = &userPhone
	}

	user, err := userManager.CreateUser(userName, userEmail, userPassword, phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("✅ User created successfully!\n")
	fmt.Printf("📧 Email: %s\n", user.Email)
	fmt.Printf("🆔 UID: %s\n", user.UID)
	fmt.Printf("📅 Created: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

// runUserLogin authenticates a user
func runUserLogin(cmd *cobra.Command, args []string) error {
	if userEmail == "" || userPassword == "" {
		return fmt.Errorf("email and password are required")
	}

	// Initialize user directories
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return fmt.Errorf("failed to create user directories: %w", err)
	}

	// Create accounts directory if it doesn't exist
	accountsDir := filepath.Join(userDirs.AFEDir, "accounts")
	if err := os.MkdirAll(accountsDir, 0700); err != nil {
		return fmt.Errorf("failed to create accounts directory: %w", err)
	}

	// Create user manager
	userManager, err := auth.NewUserManager(accountsDir)
	if err != nil {
		return fmt.Errorf("failed to create user manager: %w", err)
	}
	defer userManager.Close()

	// Authenticate user
	user, err := userManager.AuthenticateUser(userEmail, userPassword)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Printf("✅ Authentication successful!\n")
	fmt.Printf("👤 Name: %s\n", user.Name)
	fmt.Printf("📧 Email: %s\n", user.Email)
	fmt.Printf("🆔 UID: %s\n", user.UID)
	fmt.Printf("📅 Created: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
	if user.LastLogin != nil {
		fmt.Printf("🔑 Last Login: %s\n", user.LastLogin.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// runAPIKeyCreate creates a new API key
func runAPIKeyCreate(cmd *cobra.Command, args []string) error {
	if apiKeyName == "" || userEmail == "" {
		return fmt.Errorf("API key name and user email are required")
	}

	// Parse expiration date
	var expiresAt *time.Time
	if apiKeyExpires != "" {
		if parsed, err := time.Parse("2006-01-02", apiKeyExpires); err == nil {
			expiresAt = &parsed
		} else {
			return fmt.Errorf("invalid expiration date format, use YYYY-MM-DD")
		}
	}

	// Initialize user directories
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return fmt.Errorf("failed to create user directories: %w", err)
	}

	// Create accounts directory if it doesn't exist
	accountsDir := filepath.Join(userDirs.AFEDir, "accounts")
	if err := os.MkdirAll(accountsDir, 0700); err != nil {
		return fmt.Errorf("failed to create accounts directory: %w", err)
	}

	// Create user manager
	userManager, err := auth.NewUserManager(accountsDir)
	if err != nil {
		return fmt.Errorf("failed to create user manager: %w", err)
	}
	defer userManager.Close()

	// Get user by email
	user, err := userManager.GetUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Create API key
	apiKeyRecord, apiKey, err := userManager.CreateAPIKey(user.UID, apiKeyName, expiresAt, apiKeyScopes)
	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	fmt.Printf("✅ API key created successfully!\n")
	fmt.Printf("🔑 Key: %s\n", apiKey)
	fmt.Printf("📝 Name: %s\n", apiKeyRecord.Name)
	fmt.Printf("🆔 Key ID: %s\n", apiKeyRecord.KeyID)
	fmt.Printf("📅 Created: %s\n", apiKeyRecord.CreatedAt.Format("2006-01-02 15:04:05"))
	if apiKeyRecord.ExpiresAt != nil {
		fmt.Printf("⏰ Expires: %s\n", apiKeyRecord.ExpiresAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("🔒 Scopes: %v\n", apiKeyRecord.Scopes)

	fmt.Println("\n⚠️  Save this API key securely. It will not be shown again.")

	return nil
}

// runAPIKeyList lists API keys for a user
func runAPIKeyList(cmd *cobra.Command, args []string) error {
	if userEmail == "" {
		return fmt.Errorf("user email is required")
	}

	// Initialize user directories
	userDirs, err := userdirs.NewUserDirectories()
	if err != nil {
		return fmt.Errorf("failed to create user directories: %w", err)
	}

	// Create accounts directory if it doesn't exist
	accountsDir := filepath.Join(userDirs.AFEDir, "accounts")
	if err := os.MkdirAll(accountsDir, 0700); err != nil {
		return fmt.Errorf("failed to create accounts directory: %w", err)
	}

	// Create user manager
	userManager, err := auth.NewUserManager(accountsDir)
	if err != nil {
		return fmt.Errorf("failed to create user manager: %w", err)
	}
	defer userManager.Close()

	// Get user by email
	user, err := userManager.GetUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	fmt.Printf("🔑 API Keys for %s (%s)\n", user.Name, user.Email)
	fmt.Println(strings.Repeat("=", 50))

	// TODO: Implement API key listing
	// This would require adding a method to list API keys for a user
	fmt.Println("📝 API key listing not yet implemented")

	return nil
}
