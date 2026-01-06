# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Tiered rate limiting with Redis-backed sliding window algorithm
  - Tier 1 (10 req/min): Expensive cryptography operations (POST /secrets, POST /secrets/:id/access)
  - Tier 2 (30 req/min): Moderate operations (GET /secrets/:id, DELETE /secrets/:id)
  - Tier 3 (60 req/min): Lightweight operations (GET /v2/config)
  - Health check (120 req/min): Monitoring endpoint (GET /health-check)
- Rate limit configuration settings with environment variable support
  - `RATE_LIMIT_ENABLED` (default: true, disabled in local development via `make services`)
  - `RATE_LIMIT_WINDOW_SECONDS` (default: 60)
  - `RATE_LIMIT_TIER1_REQUESTS_PER_WINDOW` (default: 10)
  - `RATE_LIMIT_TIER2_REQUESTS_PER_WINDOW` (default: 30)
  - `RATE_LIMIT_TIER3_REQUESTS_PER_WINDOW` (default: 60)
  - `RATE_LIMIT_HEALTH_CHECK_REQUESTS_PER_WINDOW` (default: 120)
- HTTP 429 Too Many Requests responses when rate limits are exceeded
- Standard rate limit HTTP headers on all responses
  - `X-RateLimit-Limit`: Maximum requests allowed in window
  - `X-RateLimit-Remaining`: Requests remaining in current window
  - `X-RateLimit-Reset`: Unix timestamp when window resets
  - `Retry-After`: Seconds until rate limit resets (on 429 responses only)
- RateLimitError type in pkg/errors for structured rate limit error handling
- Per-client IP rate limiting with support for X-Forwarded-For headers
- Fail-open behavior when Redis is unavailable (logs warning, allows requests)
- `/api/v2/config` endpoint for querying runtime configuration limits (maxFileSizeMB, maxAccessCount, maxExpirationSeconds)
- `APP_MAX_ACCESS_COUNT` configuration setting with default of 100 (minimum value: 1)
- `APP_MAX_EXPIRATION_SECONDS` configuration setting with default of 604800 seconds / 7 days (minimum value: 900 seconds / 15 minutes)
- Minimum value validation for all configuration settings to prevent misconfiguration
- Validation of `access_limit` parameter against `APP_MAX_ACCESS_COUNT` in CreateSecret command
- Validation of `expiration_epoch` parameter against `APP_MAX_EXPIRATION_SECONDS` in CreateSecret command
- HTTP 400 Bad Request responses when access_limit or expiration exceed configured maximums
- ValidationError type in pkg/errors for consistent error handling patterns
- FileTooLargeError type in pkg/errors for HTTP 413 responses when file size exceeds limits
- Centralized error handler middleware for automatic error-to-HTTP-status mapping
- Consistent error logging across all API endpoints

### Changed
- Refactored CreateSecret command to use ValidationError type instead of boolean return parameter
- Controllers now use IsValidationError() function instead of boolean flag for error type checking
- Improved error handling consistency across command and controller layers
- All controllers now use c.Error() pattern with middleware handling HTTP responses
- Parameter validation errors wrapped as ValidationError types for consistent handling
- HTTP error responses now centralized in middleware instead of scattered across controllers

## [3.3.1]

### Fixed
- Bug where filename was stored in Redis but not retrieved during secret access in v2 API
- AccessSecret command now correctly includes filename field in returned Secret struct
- Content-Disposition header now uses actual filename instead of fallback pattern for file downloads

### Added
- Unit test for filename retrieval in AccessSecret command to prevent regression

## [3.3.0]

### Added
- Filename storage and retrieval for v2 file uploads
- Filenames are now preserved through the upload-storage-retrieval cycle
- Original filename returned in CreateSecret response (v2 API)
- Actual filename used in AccessSecret Content-Disposition header (v2 API)
- Redis filename key storage with TTL matching other secret data
- Table-driven integration tests for filename storage scenarios

### Security
- Filename excluded from GetSecretMetadata response to prevent information leakage (filenames could reveal sensitive information about secret contents)

### Changed
- v2 CreateSecret response now includes `filename` field (optional, backward compatible)
- v2 AccessSecret uses actual filename in download headers instead of generic `cellar-{id}` pattern
- Test helpers updated with optional filename parameters using variadic arguments

**Note:** This release is fully backward compatible.
Old secrets without stored filenames use fallback pattern `cellar-{shortID}`.
Text secrets have empty filename field (only file uploads preserve filenames).
No breaking changes to HTTP API or configuration.

## [3.2.2]

### Fixed
- Fixed AWS KMS configuration variable from `CRYPTOGRAPHY_AWS_KMS_KEY_NAME` to `CRYPTOGRAPHY_AWS_KMS_KEY_ID` (correct AWS terminology)
- Fixed `APP_MAX_FILE_SIZE_MB` configuration variable to use snake_case (`APP_MAX_FILE_SIZE_MB` → reads `app.max_file_size_mb`)
- Added missing Docker secrets support (`_FILE` suffix) for Vault AWS IAM auth in docker-entrypoint.sh
- Added missing Docker secrets support (`_FILE` suffix) for Vault GCP IAM auth in docker-entrypoint.sh
- Added missing Docker secrets support (`_FILE` suffix) for Vault Kubernetes auth in docker-entrypoint.sh
- Added missing Docker secrets support (`_FILE` suffix) for AWS KMS region and key ID in docker-entrypoint.sh

**Note:** These are bug fixes for configuration variables that should have used correct naming from the start.
If you were using the incorrect variable names, update your configuration:
- `CRYPTOGRAPHY_AWS_KMS_KEY_NAME` → `CRYPTOGRAPHY_AWS_KMS_KEY_ID`
- Environment variable naming is now consistent (all use snake_case with underscores)

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

[Unreleased]: https://gitlab.com/cellar-app/cellar-api/-/compare/v3.3.1...main
[3.3.1]: https://gitlab.com/cellar-app/cellar-api/-/compare/v3.3.0...v3.3.1
[3.3.0]: https://gitlab.com/cellar-app/cellar-api/-/compare/v3.2.2...v3.3.0
[3.2.2]: https://gitlab.com/cellar-app/cellar-api/-/compare/v3.2.1...v3.2.2
[3.2.1]: https://gitlab.com/cellar-app/cellar-api/-/compare/v3.2.0...v3.2.1
[3.2.0]: https://gitlab.com/cellar-app/cellar-api/-/compare/v3.1.1...v3.2.0
[3.1.1]: https://gitlab.com/cellar-app/cellar-api/-/compare/v3.1.0...v3.1.1
[3.1.0]: https://gitlab.com/cellar-app/cellar-api/-/compare/v3.0.0...v3.1.0
[3.0.0]: https://gitlab.com/cellar-app/cellar-api/-/compare/v2.1.0...v3.0.0
[2.1.0]: https://gitlab.com/cellar-app/cellar-api/-/compare/v2.0.0...v2.1.0
[2.0.0]: https://gitlab.com/cellar-app/cellar-api/-/compare/v1.0.1...v2.0.0
[1.0.1]: https://gitlab.com/cellar-app/cellar-api/-/compare/v1.0.0...v1.0.1
[1.0.0]: https://gitlab.com/cellar-app/cellar-api/-/tags/v1.0.0