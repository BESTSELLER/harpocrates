# Hexagonal Architecture Documentation

## Overview

Harpocrates has been refactored to follow the **Hexagonal Architecture** pattern (also known as Ports and Adapters). This architectural style separates the core business logic from external concerns, making the codebase more maintainable, testable, and extensible.

## Architecture Layers

### Domain Layer (`domain/`)

The domain layer contains the core business logic and is independent of any infrastructure concerns.

#### Ports (`domain/ports/`)

Ports are interfaces that define contracts for interacting with external systems. They act as boundaries between the domain and infrastructure layers.

- **`SecretFetcher`**: Interface for fetching secrets from a secret store
  - `ReadSecret(path string) (map[string]interface{}, error)`
  - `ReadSecretKey(path string, key string) (string, error)`

- **`SecretWriter`**: Interface for writing secrets to storage
  - `Write(output, fileName string, content interface{}, owner *int, append bool) error`
  - `Read(filePath string) (string, error)`

- **`Authenticator`**: Interface for authentication with secret stores
  - `Login() (*AuthResult, error)`
  - `IsTokenValid(token string, expiry time.Time) bool`

#### Services (`domain/services/`)

Services contain the business logic for orchestrating operations.

- **`SecretService`**: Orchestrates secret extraction and processing
  - `ExtractSecrets(input util.SecretJSON, appendToFile bool) ([]Outputs, error)`

### Adapter Layer (`adapters/`)

Adapters are concrete implementations of the ports, handling interactions with external systems.

#### Secondary Adapters (`adapters/secondary/`)

Secondary adapters are "driven" by the application (the application calls them).

- **`vault.Adapter`**: Implements `SecretFetcher` for HashiCorp Vault
- **`filesystem.Adapter`**: Implements `SecretWriter` for filesystem operations
- **`auth.Adapter`**: Implements `Authenticator` for various auth methods (Kubernetes, GCP, Token)

#### Primary Adapters (`adapters/primary/`)

Primary adapters "drive" the application (they call into the domain). Currently, the CLI commands in `cmd/` serve this role.

## Benefits of Hexagonal Architecture

1. **Separation of Concerns**: Business logic is isolated from infrastructure code
2. **Testability**: Easy to test with mock implementations of ports
3. **Flexibility**: Easy to swap implementations (e.g., use a different secret store)
4. **Maintainability**: Changes to infrastructure don't affect business logic
5. **Extensibility**: Easy to add new adapters (e.g., HTTP API, gRPC)

## Backward Compatibility

All existing CLI commands and functionality remain unchanged. The refactor is purely internal and maintains full backward compatibility with existing configurations and usage patterns.

## Testing

The new architecture enables easier testing:

```go
// Example: Testing with mock adapters
mockFetcher := &MockSecretFetcher{}
mockWriter := &MockSecretWriter{}
mockAuth := &MockAuthenticator{}

service := services.NewSecretService(mockFetcher, mockWriter, mockAuth)
// Test service without real Vault or filesystem
```

## Migration Path

Existing code gradually adopts the new architecture:

1. ✅ Created domain layer with ports and services
2. ✅ Created secondary adapters for vault, filesystem, and auth
3. ✅ Updated existing code to use adapters
4. ⏳ Future: Create primary CLI adapter for better separation
5. ⏳ Future: Add HTTP API adapter for REST endpoint support

## Adding New Adapters

To add a new secret store (e.g., AWS Secrets Manager):

1. Create a new adapter implementing `SecretFetcher`:
   ```go
   type AWSAdapter struct { ... }
   func (a *AWSAdapter) ReadSecret(path string) (map[string]interface{}, error) { ... }
   func (a *AWSAdapter) ReadSecretKey(path string, key string) (string, error) { ... }
   ```

2. Wire it up in the application:
   ```go
   awsAdapter := aws.NewAdapter(config)
   service := services.NewSecretService(awsAdapter, filesystemAdapter, authAdapter)
   ```

3. No changes needed to core domain logic (ports and services), but you may still need to update integration/wiring code (e.g., where adapters and services are constructed) until full dependency injection is in place.
