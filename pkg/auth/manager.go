package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"golang.org/x/crypto/bcrypt"
)

// UserManager handles secure user management with LevelDB
type UserManager struct {
	usersDB     *leveldb.DB
	apiKeysDB   *leveldb.DB
	accountsDir string
}

// User represents a user account
type User struct {
	UID          string     `json:"uid"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	PhoneNumber  string     `json:"phone_number,omitempty"`
	PasswordHash string     `json:"password_hash"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	IsActive     bool       `json:"is_active"`
	Roles        []string   `json:"roles,omitempty"`
}

// APIKey represents an API key
type APIKey struct {
	UID       string     `json:"uid"`
	KeyID     string     `json:"key_id"`
	KeyHash   string     `json:"key_hash"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
	IsActive  bool       `json:"is_active"`
	Scopes    []string   `json:"scopes,omitempty"`
}

// NewUserManager creates a new user manager
func NewUserManager(accountsDir string) (*UserManager, error) {
	if err := os.MkdirAll(accountsDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create accounts directory: %w", err)
	}

	// Open users database
	usersDBPath := filepath.Join(accountsDir, "users")
	usersDB, err := leveldb.OpenFile(usersDBPath, &opt.Options{
		WriteBuffer: 64 * 1024 * 1024, // 64MB write buffer
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open users database: %w", err)
	}

	// Open API keys database
	apiKeysDBPath := filepath.Join(accountsDir, "api_keys")
	apiKeysDB, err := leveldb.OpenFile(apiKeysDBPath, &opt.Options{
		WriteBuffer: 64 * 1024 * 1024, // 64MB write buffer
	})
	if err != nil {
		usersDB.Close()
		return nil, fmt.Errorf("failed to open API keys database: %w", err)
	}

	return &UserManager{
		usersDB:     usersDB,
		apiKeysDB:   apiKeysDB,
		accountsDir: accountsDir,
	}, nil
}

// Close closes the database connections
func (um *UserManager) Close() error {
	var errs []error

	if um.usersDB != nil {
		if err := um.usersDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close users DB: %w", err))
		}
	}

	if um.apiKeysDB != nil {
		if err := um.apiKeysDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close API keys DB: %w", err))
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

// CreateUser creates a new user account
func (um *UserManager) CreateUser(name, email, password string, phoneNumber *string) (*User, error) {
	// Check if user already exists by email
	if existingUser, err := um.GetUserByEmail(email); err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// Generate UID
	uid, err := um.generateUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UID: %w", err)
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		UID:          uid,
		Name:         name,
		Email:        email,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
		Roles:        []string{"user"},
	}

	if phoneNumber != nil {
		user.PhoneNumber = *phoneNumber
	}

	// Store user
	if err := um.storeUser(user); err != nil {
		return nil, fmt.Errorf("failed to store user: %w", err)
	}

	return user, nil
}

// AuthenticateUser authenticates a user with email and password
func (um *UserManager) AuthenticateUser(email, password string) (*User, error) {
	user, err := um.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	user.UpdatedAt = now

	if err := um.storeUser(user); err != nil {
		// Don't fail authentication if we can't update last login
		fmt.Printf("Warning: failed to update last login for user %s: %v\n", user.UID, err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (um *UserManager) GetUserByEmail(email string) (*User, error) {
	// Create email index key
	emailKey := []byte(fmt.Sprintf("email:%s", email))

	// Get UID by email
	uidBytes, err := um.usersDB.Get(emailKey, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Get user by UID
	return um.GetUserByUID(string(uidBytes))
}

// GetUserByUID retrieves a user by UID
func (um *UserManager) GetUserByUID(uid string) (*User, error) {
	userKey := []byte(fmt.Sprintf("user:%s", uid))

	data, err := um.usersDB.Get(userKey, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Deserialize user (simplified - in production, use proper JSON marshaling)
	user := &User{}
	if err := um.deserializeUser(data, user); err != nil {
		return nil, fmt.Errorf("failed to deserialize user: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (um *UserManager) UpdateUser(uid string, updates map[string]interface{}) (*User, error) {
	user, err := um.GetUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		user.Name = name
	}
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	if phoneNumber, ok := updates["phone_number"].(string); ok {
		user.PhoneNumber = phoneNumber
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		user.IsActive = isActive
	}
	if roles, ok := updates["roles"].([]string); ok {
		user.Roles = roles
	}

	user.UpdatedAt = time.Now()

	// Store updated user
	if err := um.storeUser(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user account
func (um *UserManager) DeleteUser(uid string) error {
	user, err := um.GetUserByUID(uid)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Start transaction batch
	batch := new(leveldb.Batch)

	// Delete user record
	userKey := []byte(fmt.Sprintf("user:%s", uid))
	batch.Delete(userKey)

	// Delete email index
	emailKey := []byte(fmt.Sprintf("email:%s", user.Email))
	batch.Delete(emailKey)

	// Delete all API keys for this user
	apiKeyPrefix := []byte(fmt.Sprintf("api_key:%s:", uid))
	iter := um.apiKeysDB.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Seek(apiKeyPrefix); iter.Valid() && strings.HasPrefix(string(iter.Key()), string(apiKeyPrefix)); iter.Next() {
		batch.Delete(iter.Key())
	}

	// Apply batch
	if err := um.usersDB.Write(batch, nil); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// CreateAPIKey creates a new API key for a user
func (um *UserManager) CreateAPIKey(uid, name string, expiresAt *time.Time, scopes []string) (*APIKey, string, error) {
	// Verify user exists
	if _, err := um.GetUserByUID(uid); err != nil {
		return nil, "", fmt.Errorf("user not found: %w", err)
	}

	// Generate API key
	apiKey, err := um.generateAPIKey()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Generate key ID
	keyID, err := um.generateUID()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate key ID: %w", err)
	}

	// Hash API key
	keyHash, err := um.hashAPIKey(apiKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash API key: %w", err)
	}

	// Create API key record
	apiKeyRecord := &APIKey{
		UID:       uid,
		KeyID:     keyID,
		KeyHash:   keyHash,
		Name:      name,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		IsActive:  true,
		Scopes:    scopes,
	}

	// Store API key
	if err := um.storeAPIKey(apiKeyRecord); err != nil {
		return nil, "", fmt.Errorf("failed to store API key: %w", err)
	}

	return apiKeyRecord, apiKey, nil
}

// ValidateAPIKey validates an API key and returns the associated user
func (um *UserManager) ValidateAPIKey(apiKey string) (*User, *APIKey, error) {
	// Hash the provided API key
	keyHash, err := um.hashAPIKey(apiKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash API key: %w", err)
	}

	// Search for API key by hash
	iter := um.apiKeysDB.NewIterator(nil, nil)
	defer iter.Release()

	var foundAPIKey *APIKey
	prefix := []byte("api_key:")

	for iter.Seek(prefix); iter.Valid() && strings.HasPrefix(string(iter.Key()), string(prefix)); iter.Next() {
		data := iter.Value()
		keyRecord := &APIKey{}
		if err := um.deserializeAPIKey(data, keyRecord); err != nil {
			continue
		}

		if keyRecord.KeyHash == keyHash && keyRecord.IsActive {
			// Check if key is expired
			if keyRecord.ExpiresAt != nil && time.Now().After(*keyRecord.ExpiresAt) {
				continue
			}

			foundAPIKey = keyRecord
			break
		}
	}

	if foundAPIKey == nil {
		return nil, nil, fmt.Errorf("invalid API key")
	}

	// Get user
	user, err := um.GetUserByUID(foundAPIKey.UID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, nil, fmt.Errorf("user account is inactive")
	}

	// Update last used
	now := time.Now()
	foundAPIKey.LastUsed = &now
	if err := um.storeAPIKey(foundAPIKey); err != nil {
		// Don't fail validation if we can't update last used
		fmt.Printf("Warning: failed to update last used for API key %s: %v\n", foundAPIKey.KeyID, err)
	}

	return user, foundAPIKey, nil
}

// Helper methods

func (um *UserManager) generateUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (um *UserManager) generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (um *UserManager) hashAPIKey(apiKey string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (um *UserManager) storeUser(user *User) error {
	// Serialize user (simplified)
	data := um.serializeUser(user)

	// Store user record
	userKey := []byte(fmt.Sprintf("user:%s", user.UID))
	if err := um.usersDB.Put(userKey, data, nil); err != nil {
		return fmt.Errorf("failed to store user: %w", err)
	}

	// Store email index
	emailKey := []byte(fmt.Sprintf("email:%s", user.Email))
	if err := um.usersDB.Put(emailKey, []byte(user.UID), nil); err != nil {
		return fmt.Errorf("failed to store email index: %w", err)
	}

	return nil
}

func (um *UserManager) storeAPIKey(apiKey *APIKey) error {
	// Serialize API key (simplified)
	data := um.serializeAPIKey(apiKey)

	// Store API key record
	keyRecordKey := []byte(fmt.Sprintf("api_key:%s:%s", apiKey.UID, apiKey.KeyID))
	if err := um.apiKeysDB.Put(keyRecordKey, data, nil); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	return nil
}

// Simplified serialization methods (in production, use proper JSON marshaling)
func (um *UserManager) serializeUser(user *User) []byte {
	return []byte(fmt.Sprintf(
		"uid:%s|name:%s|email:%s|phone:%s|hash:%s|created:%d|updated:%d|active:%t",
		user.UID, user.Name, user.Email, user.PhoneNumber, user.PasswordHash,
		user.CreatedAt.Unix(), user.UpdatedAt.Unix(), user.IsActive,
	))
}

func (um *UserManager) deserializeUser(data []byte, user *User) error {
	// Simplified deserialization (in production, use proper JSON unmarshaling)
	parts := strings.Split(string(data), "|")
	if len(parts) < 8 {
		return fmt.Errorf("invalid user data format")
	}

	user.UID = parts[0]
	user.Name = parts[1]
	user.Email = parts[2]
	user.PhoneNumber = parts[3]
	user.PasswordHash = parts[4]

	if created, err := strconv.ParseInt(parts[5], 10, 64); err == nil {
		user.CreatedAt = time.Unix(created, 0)
	}
	if updated, err := strconv.ParseInt(parts[6], 10, 64); err == nil {
		user.UpdatedAt = time.Unix(updated, 0)
	}
	if active, err := strconv.ParseBool(parts[7]); err == nil {
		user.IsActive = active
	}

	return nil
}

func (um *UserManager) serializeAPIKey(apiKey *APIKey) []byte {
	return []byte(fmt.Sprintf(
		"uid:%s|key_id:%s|hash:%s|name:%s|created:%d|expires:%d|active:%t",
		apiKey.UID, apiKey.KeyID, apiKey.KeyHash, apiKey.Name,
		apiKey.CreatedAt.Unix(), um.timeToUnix(apiKey.ExpiresAt), apiKey.IsActive,
	))
}

func (um *UserManager) deserializeAPIKey(data []byte, apiKey *APIKey) error {
	parts := strings.Split(string(data), "|")
	if len(parts) < 7 {
		return fmt.Errorf("invalid API key data format")
	}

	apiKey.UID = parts[0]
	apiKey.KeyID = parts[1]
	apiKey.KeyHash = parts[2]
	apiKey.Name = parts[3]

	if created, err := strconv.ParseInt(parts[4], 10, 64); err == nil {
		apiKey.CreatedAt = time.Unix(created, 0)
	}
	if expires, err := strconv.ParseInt(parts[5], 10, 64); err == nil && expires > 0 {
		t := time.Unix(expires, 0)
		apiKey.ExpiresAt = &t
	}
	if active, err := strconv.ParseBool(parts[6]); err == nil {
		apiKey.IsActive = active
	}

	return nil
}

func (um *UserManager) timeToUnix(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}
