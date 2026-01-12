# Contributing to the Cellar API

First of all, thank you for your desire to contribute!
It is sincerely appreciated!

> Note: The primary location for contributing to the project is on GitLab.
> It is mirrored to other locations for visibility.
> If you would like to contribute, start by navigating to [this document on GitLab][contributing-gitlab].

## Reporting Issues

Reporting issues is a great way to contribute!
We appreciate detailed issue reports.
However, before reporting an issue, please make sure it doesn't already exist in the [issues list][issues-list].

If an issue does exist, please refrain from commenting "+1" or similar comments.
That said, if you have addditional context, such as new ways to reproduce an issue, please to leave a comment.

When reporting an issue, make sure to follow the [bug issue template][issues-bug].
Make sure not to include any sensitive information.
You may replace any sensitive information that appears in logs you share with the word "REDACTED".


## Requesting New Datastores or Cryptography Engines

Cellar is designed to integrate with the systems you already trust.
That means that Cellar does handle encryption or data storage itself.
It relies on other trusted infrastructure that you deploy, such as [Hashicorp Vault][vault] or [Redis][redis].

If Cellar does not currently support your favorite [Datastore][docs-datastore] or your favorite [cryptography engine][docs-cryptography],
request it by creating an [engine request issue][issues-engine-request].
Make sure to follow the template and provide as much information about your use case as possible to help with implementation.


## Requesting Other Features

For any other feature requests, please create a [feature request issue][issues-feature-request].
Explain in as much detail as possible why you need the feature and how it would work.
Feature requests will be evaluated on a case by case basis.

Keep in mind that Cellar is built on the principle of minimalism.
It aspires to be a lightweight layer to facilitate secret sharing through existing secure platforms.
It is not intended to replace or replicate the features of a password manager or other feature rich, account based secret sharing platforms.


## Developing the Cellar API

The Cellar API is a RESTful API written in Go.
Before working on the Cellar API, make sure you are familiar with the [purpose of Cellar][docs-home] and the [components that make up Cellar][docs-application-structure]


### Local Developer Dependencies

The Cellar API is a RESTful API written in [Go][golang] using the [Gin Web Framework][gin].
**Requires Go 1.21 or later** (currently developed with Go 1.25.5).

This project does use go modules, so most of the dependencies can be installed using `go get ./...`.
However, to generate swagger docs, you will need the `swag` command.
For more information see the [gin-swagger readme][gin-swagger].
You will also need the `mockgen` command for generating mocks for unit testing.
For more information see the [uber-go mock readme][go-mock].
For linting, you will need [golangci-lint][golangci-lint] installed.

This project also makes extensive use of [**GNU Make**][gnu-make].
Make can generally be installed from your linux distros package manager or on Mac using brew.
For Windows there are multiple ports of Make from which to choose or installing make on Windows subsystem for Linux will likely also work.

[Curl][curl] is used within Makefiles for making http requests.
It can generally be installed from your linux distros package manager or on Mac using brew.
On Windows curl is usually aliased automatically to another windows native http client.

Finally, working on this project relies heavily on [Docker][docker] and [docker-compose][docker-compose].


### Getting started

Once you have all the above installed, you are ready to start the dependencies.
This is done through make by running `make services`.
This will startup any dependencies required by the API and bootstrap their configuration.
It will also output any relevant secrets into a file called `.env` from which the Cellar API will read secrets.

You are now ready to run the API.
This can either be done from your IDE or using make:

```shell
make run
```

> Note: If you choose to run without `make` you will need to load all values from the .env file into your environment.

Now in another terminal or in your IDE you can run the tests:

```shell
make test           # Run all tests (unit, integration, acceptance)
make test-unit      # Run only unit tests
make test-integration  # Run only integration tests
make test-acceptance   # Run only acceptance tests
```

> Note: The unit tests do not require the services to be running.
> The integration tests require Redis and Vault to be running.
> The acceptance tests require the API to be running.

If all the tests pass, you are ready to begin work!
You can stop the API anytime by terminating the process,
and the running dependencies can be stopped:

```shell
make clean-services
```

> **Tip:** Run `make targets` to see all available make commands.


### Project Structure

This project is structured in three main parts:

- The `cmd` folder contains a package for the main entrypoint to the program.
  If more binaries are ever added to Cellar, they will be placed in their own package here.
- The `pkg` folder contains the majority of the source of this project.
  Basically, all implementation code is found here along with related unit tests.
- The `testing` folder contains the integration and acceptance tests, each within their own folder.
  (Unit tests are left next to their packages in the `pkg` folder.


#### Testing

Tests are split into three types: unit, integration, and acceptance.
For the sake of this project, the differences are as follows:

**Unit tests** verify isolated, public parts of packages.
They are found next to the packages they test and make use of mocking and similar techniques to isolate their targets.

**Integration tests** verify the integration between the cellar API and external dependencies.
For example, the integration between the Cellar API redis client implemenation and a running instance of Redis.

**Acceptance tests** verify end to end functionality of the API.
For example, verifying that a given RESTful endpoint behaves as expected.
In some cases it may be necessary to use multiple endpoints for a single test.


#### Rate Limiting in Development

Rate limiting is **disabled by default** in local development environments to ensure tests run reliably without interference.
When you run `make services`, the generated `.env` file automatically sets `RATE_LIMIT_ENABLED=false`.

**Why disabled in local development:**
- Prevents test failures from rate limit collisions when all tests use the same IP (127.0.0.1)
- Ensures tests remain fast and idempotent
- Allows focus on business logic rather than rate limiting behavior
- Acceptance tests verify API functionality without rate limiting interference

**Rate limiting is thoroughly tested:**
- **Unit tests** (`pkg/ratelimit/redis_test.go`) - Test rate limiter logic with mocks
- **Integration tests** (`testing/integration/middleware/ratelimit_test.go`) - Test with real Redis using unique IPs per test

**To test rate limiting behavior:**
```shell
# Integration tests include rate limiting tests
make test-integration
```

**To enable rate limiting in local development:**
```shell
# Edit .env file
RATE_LIMIT_ENABLED=true

# Then restart the API
make stop-daemon
make run-daemon
```

**Production deployment:**
Rate limiting is enabled by default in production.
Configure via environment variables:
```shell
RATE_LIMIT_ENABLED=true  # Default in production
RATE_LIMIT_TIER1_REQUESTS_PER_WINDOW=10  # Cryptography operations
RATE_LIMIT_TIER2_REQUESTS_PER_WINDOW=30  # Moderate operations
RATE_LIMIT_TIER3_REQUESTS_PER_WINDOW=60  # Lightweight operations
```


#### Code Quality

Before committing code, make sure to format and lint your changes:

```shell
make format    # Format code with gofmt
make lint      # Run golangci-lint
```

All code changes should pass linting without warnings or errors.


#### DataStore and Cryptography Interfaces

Within `pkg/datastore` and `pkg/cryptography` there is an interface for datastores and for cryptography engines.
Adding a new datastore or cryptography engine is as simple as implementing one of those interfaces in a new package.
Then simply expose the new implementations using settings `pkg/settings`.

If you need to re-generate mocks for any reason, you can do so with the `generate-mocks` target:

```shell
make generate-mocks
```

Any changes to the mocks should be checked in to source control along with other changes.


#### Swagger

This project uses [swagger][swagger] (specifically [gin-swagger][gin-swagger]) for documentation and to make manually testing the API easier.

With the API running, you can load the swagger page at http://127.0.0.1:8080/swagger/index.html.

To regenerate the swagger documentation, use the `swag-init` target:


```shell
make swag init
```



### Versioning

Cellar uses two separate versioning schemes that serve different purposes:

#### 1. Application Versioning (Semantic Versioning)

This project uses [semantic versioning][semver] for application releases.
The version number indicates user-facing changes and deployment impacts.

**Version Format:** `MAJOR.MINOR.PATCH` (e.g., 3.2.0)

**MAJOR version** increments when:
- HTTP API contracts break (adding new endpoint version, removing old endpoint version)
- Configuration format requires migration
- Deployment process changes fundamentally
- Minimum Go version increases by major version

**MINOR version** increments when:
- New features are added to existing endpoints
- Existing functionality is enhanced
- Dependencies are upgraded (minor/patch versions)
- Internal improvements are made
- New optional configuration is added

**PATCH version** increments when:
- Bugs are fixed in existing functionality
- Security vulnerabilities are patched
- Documentation is updated

**How to update application version:**
1. Update `APP_VERSION` in `.gitlab-ci.yml`
2. Document changes in `CHANGELOG.md` under appropriate version
3. Create merge request
4. Tagging and releasing happen automatically through [CI/CD pipelines][pipelines]

#### 2. API Endpoint Versioning

Cellar has versioned HTTP endpoints (`/api/v1/`, `/api/v2/`) that represent distinct API contracts.

**Rolling Window Policy:** Cellar maintains **current + previous** endpoint versions (n and n-1).

**Current state:**
- **v2** (current): File upload support, enhanced metadata, request cancellation
- **v1** (previous): JSON-only secrets, simple metadata, legacy support

**When v3 is created:**
- **v3** becomes current version
- **v2** becomes previous version
- **v1** is removed entirely (code, tests, documentation)
- Users have one full endpoint version cycle to migrate from v1 → v2

#### When to Add New Endpoint Version

**Only when making breaking changes** to the current endpoint's HTTP contract.

Examples of breaking changes requiring new version:
- Changing required fields in requests/responses
- Modifying field types (string → object, int → string)
- Removing fields that clients depend on
- Changing HTTP status codes for existing behaviors
- Changing authentication/authorization mechanisms

Examples that do NOT require new version:
- Adding new optional fields (backward compatible)
- Adding new features while maintaining existing behavior
- Internal implementation changes (encryption, storage)
- Bug fixes that restore documented behavior
- Performance improvements

**Decision tree for changes:**

```
Is the HTTP contract changing?
├─ No → Add to v2 (current), maybe backport to v1 (minor/patch bump)
└─ Yes → Is it backward compatible?
    ├─ Yes → Add to v2 as optional feature (minor bump)
    └─ No → Must break contract?
        ├─ No → Redesign to be backward compatible
        └─ Yes → Create v3, remove v1 (MAJOR bump)
```

#### Process for Adding New Endpoint Version

If you must create v3 (happens rarely, years apart):

**1. Planning Phase**
- Document why v2 contract must break
- Design v3 contract with long-term stability in mind
- Create migration guide from v2 → v3
- Announce deprecation timeline for v1 in advance (if possible)

**2. Implementation Phase**
- Create new controller package: `pkg/controllers/v3/`
- Implement new handlers with improved contract
- Keep v2 functional (now becomes "previous")
- Remove all v1 code:
  - Delete `pkg/controllers/v1/`
  - Delete `testing/acceptance/v1/`
  - Remove v1 routes from main
  - Remove v1 from Swagger docs
- Update Swagger to show v2 and v3
- Add acceptance tests in `testing/acceptance/v3/`

**3. Documentation Phase**
- Update `CHANGELOG.md` as MAJOR version (e.g., 3.x → 4.0.0)
- Document v1 removal
- Document v2 → v3 migration path
- Update README and API documentation
- Update CLAUDE.md with new current/previous versions

**4. Release Phase**
- MAJOR version bump (removes v1, adds v3)
- Clear release notes about removed v1 support
- Migration examples for v2 → v3

#### Example Changelog Entry for v3

```markdown
## [4.0.0] - BREAKING CHANGES

### Added
- API v3 endpoints with [describe improvements]

### Changed (BREAKING)
- Removed API v1 endpoints (deprecated since version X.X.X)
- v2 endpoints remain fully supported (now previous version)
- v3 endpoints are now current version

### Migration
- Users on v1: Migrate to v2 first, then v3
- Users on v2: See v2 → v3 migration guide
```

#### Endpoint Lifecycle Example

```
Cellar 2.0.0: Adds v2, maintains v1 (current: v2, previous: v1)
Cellar 3.0.0: Maintains v2 and v1 (current: v2, previous: v1)
Cellar 3.1.0: Maintains v2 and v1 (current: v2, previous: v1)
Cellar 3.2.0: Maintains v2 and v1 (current: v2, previous: v1)
Cellar 4.0.0: Adds v3, removes v1 (current: v3, previous: v2)
Cellar 4.1.0: Maintains v3 and v2 (current: v3, previous: v2)
Cellar 5.0.0: Adds v4, removes v2 (current: v4, previous: v3)
```

Users get **one full endpoint version cycle** as a migration window.

#### Current Development Guidelines

**For v2 (current):**
- Add new features here
- This is the primary development target
- Use proper request context and cancellation
- Return HTTP 408 for context errors

**For v1 (previous):**
- Only backport critical security fixes
- Only backport fixes for showstopper bugs
- Use background context for backward compatibility
- Do not add new features
- Will be removed when v3 is created

**Before creating v3:**
- Ensure v2 has been stable for significant time (months/years)
- Ensure there's strong justification for breaking v2 contract
- Ensure v3 contract is designed for long-term stability
- Document comprehensive migration path from v2 → v3


### Final Thoughts

- Documentation is hard and keeping it up to date is harder.
  Contributions that help keep this and other documentation clear, concise, and up to date are both welcome and appreciated.
- Tests are mandatory. Code changes will not be accepted without new or updated tests nor without all tests passing.


[makefile]: Makefile
[go-mod]: go.mod
[gitlab-ci]: .gitlab-ci.yml
[changelog]: CHANGELOG.md
[contributing-gitlab]: https://gitlab.com/cellar-app/cellar-api/-/blob/main/CONTRIBUTING.md

[docs-datastore]: https://cellar-app.io/basics/application-structure/#datastore
[docs-cryptography]: https://cellar-app.io/basics/application-structure/#cryptography
[docs-application-structure]: https://cellar-app.io/basics/application-structure/
[docs-home]: https://cellar-app.io/

[issues-list]: https://gitlab.com/cellar-app/cellar-api/-/issues
[issues-bug]: https://gitlab.com/cellar-app/cellar-api/-/issues/new
[issues-engine-request]: https://gitlab.com/cellar-app/cellar-api/-/issues/new
[issues-feature-request]: https://gitlab.com/cellar-app/cellar-api/-/issues/new

[pipelines]: https://gitlab.com/cellar-app/cellar-api/-/pipelines

[gin]:  https://github.com/gin-gonic/gin
[go-mock]: https://github.com/uber-go/mock
[gin-swagger]: https://github.com/swaggo/gin-swagger
[golangci-lint]: https://golangci-lint.run/

[vault]: https://www.vaultproject.io/
[redis]: https://redis.io/
[golang]: https://golang.org/
[gnu-make]: https://www.gnu.org/software/make/
[docker]: https://www.docker.com/
[docker-compose]: https://docs.docker.com/compose/
[curl]: https://curl.se/
[semver]: https://semver.org/
[swagger]: https://swagger.io/
