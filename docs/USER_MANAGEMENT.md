# AgentForge Engine User Management 🛡️

[![Security](https://img.shields.io/badge/Security-Enterprise-green.svg)](https://github.com/AgentForgeEngine/AgentForgeEngine)
[![LevelDB](https://img.shields.io/badge/LevelDB-Embedded-blue.svg)](https://github.com/syndtr/goleveldb)
[![bcrypt](https://img.shields.io/badge/bcrypt-Secure-orange.svg)](https://golang.org/x/crypto/bcrypt)

Enterprise-grade user management system for AgentForge Engine with LevelDB storage, bcrypt password hashing, and secure API key management.

## 🌟 Security Features

- **🔐 bcrypt Password Hashing**: Industry-standard password security
- **🗄️ LevelDB Storage**: High-performance embedded database
- **🔑 API Key Management**: Cryptographically secure key generation
- **📊 Audit Trail**: Comprehensive logging and tracking
- **🔒 Access Control**: Role-based permissions and scopes
- **🛡️ Secure File Permissions**: Proper directory and file permissions

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [User Management](#user-management)
- [API Key Management](#api-key-management)
- [Security Architecture](#security-architecture)
- [CLI Reference](#cli-reference)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)
- [API Reference](#api-reference)

## 🚀 Quick Start

### Create Your First User

```bash
# Create a user account
afe user create --name "John Doe" --email "john@example.com" --password "securepassword123"

✅ User created successfully!
📧 Email: john@example.com
🆔 UID: 00ef696711a48504d78561fe2cac2019
📅 Created: 2026-02-05 15:25:28
```

### Authenticate User

```bash
# Login with credentials
afe user login --email "john@example.com" --password "securepassword123"

✅ Authentication successful!
👤 Name: John Doe
📧 Email: john@example.com
🆔 UID: 00ef696711a48504d78561fe2cac2019
📅 Created: 2026-02-05 15:25:28
🔑 Last Login: 2026-02-05 15:30:45
```

### Create API Key

```bash
# Create an API key for the user
afe user api-key create --name "Production Key" --email "john@example.com"

✅ API key created successfully!
🔑 Key: 00ef696711a48504d78561fe2cac2019abc123def456ghi789jkl012mno345pqr678stu901vwx
📝 Name: Production Key
🆔 Key ID: 00ef696711a48504d78561fe2cac2019
📅 Created: 2026-02-05 15:35:00
🔒 Scopes: [read]
⚠️  Save this API key securely. It will not be shown again.
```

## 👥 User Management

### User Account Structure

```go
type User struct {
    UID          string    `json:"uid"`
    Name         string    `json:"name"`
    Email        string    `json:"email"`
    PhoneNumber  string    `json:"phone_number,omitempty"`
    PasswordHash string    `json:"password_hash"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    LastLogin    *time.Time `json:"last_login,omitempty"`
    IsActive     bool      `json:"is_active"`
    Roles        []string  `json:"roles,omitempty"`
}
```

### User Commands

#### Create User

```bash
afe user create --name "John Doe" --email "john@example.com" --password "secure123" [--phone "+1234567890"]
```

**Options:**
- `--name`: User name (required)
- `--email`: User email (required)
- `--password`: User password (required)
- `--phone`: Phone number (optional)

#### Authenticate User

```bash
afe user login --email "john@example.com" --password "secure123"
```

#### Update User

```bash
# Update user information (coming soon)
afe user update --uid "00ef696711a48504d78561fe2cac2019" --name "John Smith"
```

#### Delete User

```bash
# Delete user account (coming soon)
afe user delete --uid "00ef696711a48504d78561fe2cac2019"
```

### User Security Features

- **Password Hashing**: Uses bcrypt with proper salt
- **Account Status**: Active/inactive account management
- **Login Tracking**: Last login timestamp and audit trail
- **Role Management**: Role-based access control
- **Phone Support**: Optional phone number storage

## 🔑 API Key Management

### API Key Structure

```go
type APIKey struct {
    UID        string    `json:"uid"`
    KeyID      string    `json:"key_id"`
    KeyHash    string    `json:"key_hash"`
    Name       string    `json:"name"`
    CreatedAt  time.Time `json:"created_at"`
    ExpiresAt  *time.Time `json:"expires_at,omitempty"`
    LastUsed   *time.Time `json:"last_used,omitempty"`
    IsActive   bool      `json:"is_active"`
    Scopes     []string  `json:"scopes,omitempty"`
}
```

### API Key Commands

#### Create API Key

```bash
afe user api-key create --name "Production Key" --email "john@example.com" [--expires "2024-12-31"] [--scopes "read,write"]
```

**Options:**
- `--name`: API key name (required)
- `--email`: User email (required)
- `--expires`: Expiration date (optional, format: YYYY-MM-DD)
- `--scopes`: Comma-separated scopes (default: read)

#### List API Keys

```bash
afe user api-key list --email "john@example.com"
```

#### Revoke API Key

```bash
# Revoke API key (coming soon)
afe user api-key revoke --key-id "00ef696711a48504d78561fe2cac2019"
```

### API Key Security

- **Cryptographic Generation**: 32-byte random keys with hex encoding
- **Secure Hashing**: bcrypt hashing of API keys
- **Expiration Management**: Optional expiration dates
- **Usage Tracking**: Last used timestamp
- **Scope-Based Access**: Fine-grained permissions
- **Revocation Support**: Secure key deactivation

## 🏗️ Security Architecture

### Directory Structure

```
~/.afe/accounts/
├── users/                   # LevelDB user database
│   ├── 000002.ldb          # Main database file
│   ├── CURRENT             # Current database file
│   ├── LOG                 # Transaction log
│   └── LOCK                # File lock
└── api_keys/                # LevelDB API key database
    ├── 000003.ldb
    ├── CURRENT
    ├── LOG
    └── LOCK
```

### Security Measures

#### File Permissions

- **Directory Permissions**: `0700` (rwx for owner only)
- **Database Files**: Protected by LevelDB's internal encryption
- **Configuration Files**: Secure user-specific settings

#### Password Security

```go
// bcrypt password hashing with proper salt
passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
    return fmt.Errorf("failed to hash password: %w", err)
}
```

#### API Key Security

```go
// Cryptographically secure random key generation
bytes := make([]byte, 32)
if _, err := rand.Read(bytes); err != nil {
    return "", fmt.Errorf("failed to generate random bytes: %w", err)
}
apiKey := hex.EncodeToString(bytes)
```

### Data Protection

- **Encryption at Rest**: LevelDB's built-in encryption
- **Secure Hashing**: bcrypt for passwords and API keys
- **Access Control**: Role-based permissions
- **Audit Logging**: Comprehensive activity tracking
- **Secure Defaults**: Safe default configurations

## 🔧 Configuration

### User Directory Configuration

The user management system automatically creates secure directories:

```bash
$ ./afe init --verbose
✅ Creating ~/.afe/accounts/ directory with secure permissions
✅ User management system ready
```

### Database Configuration

LevelDB databases are configured with optimal settings:

```go
// Open users database with 64MB write buffer
usersDB, err := leveldb.OpenFile(usersDBPath, &opt.Options{
    WriteBuffer: 64 * 1024 * 1024, // 64MB write buffer
})
```

### Security Settings

```yaml
# ~/.afe/config/build_config.yaml
logging:
  build_log: "~/.afe/logs/build.log"
  cache_log: "~/.afe/logs/cache.log"
  verbose: false
  max_log_size_mb: 10
```

## 🐛 Troubleshooting

### Common Issues

#### User Creation Fails

```bash
$ afe user create --name "John" --email "john@example.com" --password "123"
❌ User creation failed: user with email john@example.com already exists
```

**Solution**: Use a different email address or delete the existing user first.

#### Authentication Fails

```bash
$ afe user login --email "john@example.com" --password "wrong"
❌ Authentication failed: user not found
```

**Solution**: Check the email address and password, or create a new user account.

#### API Key Issues

```bash
$ afe user api-key create --name "Test" --email "nonexistent@example.com"
❌ Failed to create API key: user not found
```

**Solution**: Ensure the user exists before creating API keys.

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
# Enable verbose output
export AFE_LOG_LEVEL=debug
./afe user create --name "Test" --email "test@example.com" --password "test123"
```

### Database Issues

#### Database Corruption

```bash
# Check database integrity
afe user validate

# Clean and recreate database (WARNING: This deletes all data)
rm -rf ~/.afe/accounts/
./afe init
```

#### Permission Issues

```bash
# Check directory permissions
ls -la ~/.afe/accounts/

# Fix permissions if needed
chmod 700 ~/.afe/accounts/
```

## 📚 API Reference

### User Management API

#### CreateUser
```go
userManager, err := auth.NewUserManager(accountsDir)
if err != nil {
    return nil, fmt.Errorf("failed to create user manager: %w", err)
}
defer userManager.Close()

user, err := userManager.CreateUser("John Doe", "john@example.com", "password123", nil)
if err != nil {
    return nil, fmt.Errorf("failed to create user: %w", err)
}
```

#### AuthenticateUser
```go
user, err := userManager.AuthenticateUser("john@example.com", "password123")
if err != nil {
    return nil, fmt.Errorf("authentication failed: %w", err)
}
```

#### GetUserByEmail
```go
user, err := userManager.GetUserByEmail("john@example.com")
if err != nil {
    return nil, fmt.Errorf("user not found: %w", err)
}
```

### API Key Management API

#### CreateAPIKey
```go
apiKeyRecord, apiKey, err := userManager.CreateAPIKey(
    user.UID,
    "Production Key",
    nil, // no expiration
    []string{"read", "write"},
)
if err != nil {
    return nil, fmt.Errorf("failed to create API key: %w", err)
}
```

#### ValidateAPIKey
```go
user, apiKeyRecord, err := userManager.ValidateAPIKey("00ef696711a48504d78561fe2cac2019abc123def456ghi789jkl012mno345pqr678stu901vwx")
if err != nil {
    return nil, fmt.Errorf("invalid API key: %w", err)
}
```

## 📊 Security Best Practices

### Password Security

- **Minimum Length**: Require passwords of at least 8 characters
- **Complexity Requirements**: Include uppercase, lowercase, numbers, and symbols
- **bcrypt Cost**: Use appropriate bcrypt cost factor (default is good)
- **Password Policies**: Implement password rotation policies

### API Key Security

- **Key Length**: Use 32-byte (256-bit) random keys
- **Hex Encoding**: Use hex encoding for easy transport
- **Expiration**: Set reasonable expiration dates for production keys
- **Scope Limitation**: Apply principle of least privilege
- **Regular Rotation**: Implement key rotation policies

### Database Security

- **File Permissions**: Restrict access to user directories
- **Backup Strategy**: Regular backups of LevelDB databases
- **Encryption**: Use LevelDB's built-in encryption features
- **Access Logging**: Log all database access attempts

### Operational Security

- **Audit Trail**: Log all user management operations
- **Rate Limiting**: Implement rate limiting for authentication attempts
- **Account Lockout**: Lock accounts after failed login attempts
- **Session Management**: Implement secure session handling

---

**Built with ❤️ by the AgentForge Engine security team**