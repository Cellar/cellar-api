# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [3.2.1]

### Fixed
- Critical: Added type assertion checks in Vault encryption to prevent runtime panics
- Fixed incorrect log message in AWS KMS encryption (said "vault" instead of "aws kms")
- Fixed stale comment in Lambda handler function

### Changed
- Migrated test suite to use testify/assert and testify/require exclusively
- Removed custom test assertion helpers in favor of testify standard patterns
- Replaced deprecated io/ioutil package with modern io and os equivalents (Go 1.16+)
- Organized imports following Go conventions (stdlib, external, internal) across all files
- Simplified RedisKey.buildKey() to remove unnecessary closure pattern
- Removed redundant type specifications from variable declarations

### Added
- Godoc comments for exported functions: FileToBytes, HandleError, SanitizeFilename, GetGcpIamRequestInfo
- HTTP status constant usage (http.StatusOK) instead of magic number 200

**Note:** This is a patch release with code quality improvements and critical bug fixes.
No breaking changes to HTTP API or configuration.
All changes improve code safety, maintainability, and follow modern Go best practices.

## [3.2.0]

### Added
- File size limits for v2 file uploads with `APP_MAX_FILE_SIZE_MB` configuration (default: 8 MB)
- Filename sanitization utility to prevent path traversal attacks
- Empty file validation (rejects 0-byte files)
- Gin multipart memory limits that scale with configured file size
- `targets` command to Makefile to list all available targets
- `test` command to Makefile to run all test types
- `format` and `lint` commands to Makefile
- Custom context error handling package (pkg/errors) for proper request cancellation support
- Explicit context cancellation checks at the start of all command functions
- Comprehensive context cancellation tests for all command layer functions

### Security
- Secure download headers for file secrets: X-Content-Type-Options, Content-Security-Policy, X-Frame-Options, Cache-Control
- File size validation prevents DoS attacks via large uploads
- Filename sanitization prevents directory traversal and injection attacks
- Fixed fmt.Sprintf format string security issues for Go 1.25 compatibility

### Changed
- Upgraded from Go 1.22.0 to Go 1.25.5 (**requires Go 1.21+ minimum for building from source**)
- Migrated from deprecated aws-sdk-go v1 to aws-sdk-go-v2 (internal implementation change)
- Updated all Go dependencies to latest versions (Gin v1.11.0, Viper v1.21.0, Vault API v1.22.0, etc.)
- Updated Docker images: golang:1.25-alpine, golang:1.25-bookworm, redis:8-alpine, hashicorp/vault:1.21
- AWS KMS operations now use context.Context parameter (internal implementation change)
- AWS SDK clients now use config.LoadDefaultConfig for credential management (internal implementation change)
- Improved Makefile with proper Vault wait logic using polling instead of fixed sleep
- Updated CONTRIBUTING.md with better testing documentation and code quality guidelines
- Restructured Makefile `services` target into logical subtargets
- Upgraded go-redis from v7 to v9 for native context support and improved performance (internal change)
- Replaced all context.TODO() with proper context propagation from HTTP requests (internal change)
- v2 API endpoints now return HTTP 408 Request Timeout for cancelled or timed-out requests
- All internal interfaces and functions now accept context.Context for proper cancellation support

### Fixed
- `make run-daemon` now properly runs in background and returns immediately
- Daemon process output redirected to `/tmp/cellar-api.log` for troubleshooting
- Binary no longer removed while daemon is running
- Long-running v2 operations now properly respect client disconnections
- Redis and Vault operations can be cancelled mid-flight when clients disconnect
- AWS SDK v2 operations now properly receive and respect request context

**Note:** This release contains no breaking changes to the HTTP API or configuration.
v1 endpoints maintain backward compatibility using background context.
v2 endpoints gain request cancellation support (returns HTTP 408 on timeout).
All changes are internal implementation details for self-hosted deployments.
Pre-built binaries and Docker images work as drop-in replacements.

## [3.1.1]

### Fixed
- Bug with Docker expecting old configuration keys

## [3.1.0]

### Added
- Logging setting to allow either text or JSON formatted logging

## [3.0.0]

### Added
- Support for [AWS KMS](https://aws.amazon.com/kms/) as cryptography engine
- Enabled property for cryptography engines (only one can be enabled)

### Changed
- Restructured configuration to have sub-levels for both datastore and cryptography

## [2.1.0]

### Changed
- Updated Go to version 1.23
- Updated all dependencies

## [2.0.0]

### Added
- Support for [Vault AWS IAM authentication](https://www.vaultproject.io/docs/auth/aws.html)
- Support for [Vault Kubernetes authentication](https://www.vaultproject.io/docs/auth/kubernetes)
- Support for [Google Cloud IAM authentication](https://www.vaultproject.io/docs/auth/gcp)

### Changed
- Vault AppRole auth is now optional (other auth methods can be specified)
- Restructured Vault configuration with sub-levels for authentication
- Mount point of auth backend must be specified as `VAULT_AUTH_MOUNT_PATH`

### Removed
- Docker container verification of auth configuration (except mount path)

## [1.0.1]

### Added
- Application version to the `/health-check` endpoint

## [1.0.0]

### Added
- Initial open source release