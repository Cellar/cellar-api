# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- File size limits for v2 file uploads with `APP_MAX_FILE_SIZE_MB` configuration (default: 8 MB)
- Filename sanitization utility to prevent path traversal attacks
- Empty file validation (rejects 0-byte files)
- Gin multipart memory limits that scale with configured file size
- `targets` command to Makefile to list all available targets
- `test` command to Makefile to run all test types
- `format` and `lint` commands to Makefile

### Security
- Secure download headers for file secrets: X-Content-Type-Options, Content-Security-Policy, X-Frame-Options, Cache-Control
- File size validation prevents DoS attacks via large uploads
- Filename sanitization prevents directory traversal and injection attacks

### Changed
- Improved Makefile with proper Vault wait logic using polling instead of fixed sleep
- Updated CONTRIBUTING.md with better testing documentation and code quality guidelines
- Restructured Makefile `services` target into logical subtargets

### Fixed
- `make run-daemon` now properly runs in background and returns immediately
- Daemon process output redirected to `/tmp/cellar-api.log` for troubleshooting
- Binary no longer removed while daemon is running

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