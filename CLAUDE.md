# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Cellar API is a RESTful API for secure secret sharing, written in Go using the Gin web framework.
The API supports two versions (v1 and v2) and can be deployed as a standalone binary, Docker container, or AWS Lambda function.

## Requirements

**Go Version:** Go 1.25+ required (currently using Go 1.25.5)
- Minimum Go 1.21+ due to Gin v1.11.0 requirement
- Uses Go modules for dependency management

**Key Dependencies:**
- Gin Web Framework v1.11.0+ (requires Go 1.21+)
- AWS SDK for Go v2 (migrated from deprecated v1)
- Viper v1.21.0+ for configuration
- HashiCorp Vault API v1.22.0+

**AWS SDK Migration:**
- This project uses **aws-sdk-go-v2**, not the deprecated aws-sdk-go v1
- All AWS operations require `context.Context` parameter
- Configuration uses `config.LoadDefaultConfig()` pattern
- Affected packages: `pkg/cryptography/aws`, `pkg/aws`

## Development Commands

### Local Development Setup

Start Redis and Vault services locally and configure Vault:
```bash
make services
```

This creates a `.env` file with Vault credentials needed for local development.

Run the API server locally:
```bash
make run
```

Run as daemon (background process):
```bash
make run-daemon
make stop-daemon
```

### Building

Build all source files:
```bash
go build ./...
# or
make build
```

Build standalone binary:
```bash
make package APP_VERSION=1.0.0
```

Build AWS Lambda deployment package:
```bash
make package-lambda APP_VERSION=1.0.0
```

### Testing

Run unit tests (no external dependencies):
```bash
go test ./...
# or
make test-unit
```

Run integration tests (requires Redis + Vault services):
```bash
make services  # Start dependencies first
make test-integration
```

Run acceptance tests (requires running API server):
```bash
make run-daemon  # Start API in background
make test-acceptance
make stop-daemon
```

Run a single test:
```bash
go test ./pkg/models -run TestSecretModel
```

Run a specific test within a package with verbose output:
```bash
go test -v ./testing/acceptance/v2 -run TestCreateSecret
```

### Docker

Build Docker image:
```bash
make docker-build IMAGE_NAME=cellar-api IMAGE_TAG=dev
```

Run Docker image:
```bash
make docker-run IMAGE_NAME=cellar-api IMAGE_TAG=dev
```

### Code Generation

Generate mocks for testing (uses uber/mock):
```bash
make generate-mocks
```

Generate Swagger documentation:
```bash
make swag-init
```

Generate ReDocly API reference:
```bash
make redoc
```

### Vault Operations

Get role ID for testing:
```bash
make vault-role-id
```

Get secret ID for testing:
```bash
make vault-secret-id
```

## Architecture

### Project Structure

```
cmd/
├── cellar/          # HTTP server entry point
└── cellar-lambda/   # AWS Lambda handler entry point

pkg/
├── commands/        # Business logic (CreateSecret, AccessSecret, DeleteSecret)
├── controllers/     # HTTP handlers
│   ├── v1/         # API version 1 endpoints
│   └── v2/         # API version 2 endpoints
├── models/          # Data structures (Secret, SecretMetadata)
├── cryptography/    # Encryption abstraction and implementations
│   ├── vault/      # HashiCorp Vault backend
│   └── aws/        # AWS KMS backend
├── datastore/       # Storage abstraction and implementations
│   └── redis/      # Redis implementation
├── middleware/      # HTTP middleware (logging, DI, error handling)
├── settings/        # Configuration management
└── mocks/           # Generated test mocks

testing/
├── acceptance/      # E2E tests (tagged with //go:build acceptance)
├── integration/     # Integration tests (tagged with //go:build integration)
└── testhelpers/     # Shared test utilities
```

### Dependency Injection Pattern

The codebase uses middleware-based dependency injection:
1. Middleware injects configuration, encryption client, and datastore into Gin context
2. Controllers retrieve dependencies from context using `c.MustGet()`
3. Controllers pass dependencies to command handlers in `pkg/commands`

When adding new dependencies:
- Inject them in `pkg/middleware`
- Retrieve them in controllers
- Pass them to command functions

### Interface Abstractions

Key interfaces allow pluggable implementations:

**IConfiguration** - Configuration management (Viper-based)
- `App()` - application settings
- `Datastore()` - Redis configuration
- `Encryption()` - cryptography backend selection
- `Logging()` - logging configuration

**DataStore** - Storage backend (Redis implementation provided)
- `Save()` - persist secret metadata
- `Get()` - retrieve secret metadata
- `Delete()` - remove secret metadata

**Encryption** - Cryptography backend (Vault and AWS KMS implementations provided)
- `Encrypt()` - encrypt plaintext
- `Decrypt()` - decrypt ciphertext

### API Versioning

Two independent API versions with different characteristics:

**v1** (`pkg/controllers/v1`):
- JSON-based secret content
- Simpler metadata structure
- Endpoints: POST `/api/v1/secret`, GET `/api/v1/secret/:id`, DELETE `/api/v1/secret/:id`

**v2** (`pkg/controllers/v2`):
- Multipart form data with file upload support
- Enhanced metadata including `content_type`
- Endpoints: POST `/api/v2/secret`, GET `/api/v2/secret/:id`, DELETE `/api/v2/secret/:id`

When adding features, consider whether they should be v2-only or backported to v1.

## Configuration

Configuration uses Viper with environment variables.
Dots in config keys are converted to underscores for environment variables.

### Required Configuration

**Redis (always required):**
```bash
DATASTORE_REDIS_HOST=localhost
DATASTORE_REDIS_PORT=6379
DATASTORE_REDIS_PASSWORD=
DATASTORE_REDIS_DB=0
```

**Cryptography Backend (choose one):**

HashiCorp Vault:
```bash
CRYPTOGRAPHY_VAULT_ENABLED=true
CRYPTOGRAPHY_VAULT_ADDRESS=http://127.0.0.1:8200
CRYPTOGRAPHY_VAULT_AUTH_MOUNT_PATH=approle
CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID=<role-id>
CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID=<secret-id>
CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME=cellar-key
```

AWS KMS:
```bash
CRYPTOGRAPHY_AWS_ENABLED=true
CRYPTOGRAPHY_AWS_KEY_ID=<kms-key-id>
```

### Optional Configuration

```bash
APP_BIND_ADDRESS=:8080
LOGGING_LEVEL=INFO          # DEBUG, INFO, WARN, ERROR
LOGGING_FORMAT=text         # text or json
GIN_MODE=release            # release or debug
DISABLE_SWAGGER=false       # set to true to disable /swagger endpoint
```

## Testing Practices

### Test Organization

**Unit tests** - Located adjacent to source files (`*_test.go`):
- No external dependencies
- Fast execution
- Use custom assertion helpers (`testing/testhelpers`)

**Integration tests** - Located in `testing/integration/`:
- Tagged with `//go:build integration`
- Test real Vault and Redis interactions
- Require running services

**Acceptance tests** - Located in `testing/acceptance/`:
- Tagged with `//go:build acceptance`
- Make real HTTP requests to running API
- Test end-to-end workflows

### Test Helpers

Use shared utilities from `testing/testhelpers`:
- `CreateSecretV1()` / `CreateSecretV2()` - HTTP request builders
- `LoadConfig()` - test configuration setup
- `EqualsF()`, `NotEqualsF()`, `AssertF()` - custom assertions

### Running Specific Test Types

Unit tests only (no tags needed):
```bash
go test ./pkg/...
```

Integration tests only:
```bash
go test -tags=integration ./testing/integration/...
```

Acceptance tests only:
```bash
go test -tags=acceptance ./testing/acceptance/...
```

## Common Development Patterns

### Adding a New API Endpoint

1. Add handler function to appropriate controller (`pkg/controllers/v1` or `v2`)
2. Add Swagger annotations for documentation
3. Add route in controller's `SetupRoutes()` function
4. Implement business logic in `pkg/commands` if complex
5. Add acceptance test in `testing/acceptance/v1` or `v2`
6. Regenerate Swagger docs: `make swag-init`

### Adding a New Cryptography Backend

1. Implement `Encryption` interface in new package under `pkg/cryptography/`
2. Add configuration in `pkg/settings/cryptography/`
3. Add initialization logic in `cmd/cellar/main.go` and `cmd/cellar-lambda/main.go`
4. Add integration tests in `testing/integration/cryptography/`

### Adding a New Datastore Backend

1. Implement `DataStore` interface in new package under `pkg/datastore/`
2. Add configuration in `pkg/settings/datastore/`
3. Add initialization logic in `cmd/cellar/main.go` and `cmd/cellar-lambda/main.go`
4. Add integration tests in `testing/integration/datastore/`

## Deployment

### Standalone Binary

```bash
make package APP_VERSION=1.0.0
./cellar-api-1.0.0-amd64
```

### Docker

```bash
make docker-build IMAGE_TAG=1.0.0
docker run -p 8080:8080 --env-file .env cellar-api:1.0.0
```

### AWS Lambda

```bash
make package-lambda APP_VERSION=1.0.0
# Upload dist/cellar-api.zip to Lambda
# Configure Lambda with required environment variables
```

## Git Workflow

This project uses GitLab as the primary repository.
Push to the `gitlab` remote (not `origin` or `upstream` unless explicitly instructed).
