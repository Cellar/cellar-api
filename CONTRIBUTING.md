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
Naturally, you will need the go compiler.

This project does use go modules, so most of the dependencies can be installed using `go get ./...`.
However, to generate swagger docs, you will need the `swag` command.
It can be installed using `go get -u github.com/swaggo/swag/cmd/swag`.
For more information see the [gin-swagger readme][gin-swagger].
You will also need the `mockgen` command for generating mocks for unit testing.
It can be installed using `GO111MODULE=on go get github.com/golang/mock/mockgen@v{VERSION}`
(replacing {VERSION} with the version of `github.com/golang/mock` found in the [go.mod][go-mod] file.)
For more information see the [go-mock readme][go-mock].

This project also makes extensive use of [**GNU Make**][gnu-make].
Make can generally be installed from your linux distros package manager or on Mac using brew.
For Windows there are multiple ports of Make from which to choose or installing make on Windows subsystem for Linux will likely also work.

[Curl][curl] is used within Makefiles for making http requests.
It can generally be installed from your linux distros package manager or on Mac using brew.
On Windows curl is usually aliased automatically to another windows native http client.

Finally, working on this project relies heavily on [Docker][docker] and [docker-compose][docker-compose].


### Getting started

Once you have all the above installed, you are ready to start the dependencies.
This is done through make by running the `services` target of the [Makefile][makefile]:

```shell
make services
```

This will startup any dependencies required by the API and bootstrap their configuration.
It will also output any relevant secrets into a file called `.env` from which the Cellar API will read secrets.

You are now ready to run the API.
This can either be done from your IDE or using the `run` target of the [Makefile][makefile]:

```shell
make run
```

> Note: If you choose to run without `make` you will need to load all values from the .env file into your environment

Now in another terminal or in your IDE you can run the tests:

```shell
make test-unit
make test-integration
make test-acceptance
```

> Note: The unit tests do not require the services to be running,
> and only the acceptance tests actually require the API to be running.

If all the tests pass, you are ready to begin work!
You can stop the API anytime by terminating the process,
and the running dependencies can be stopped with `make clean-services`.


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


#### DataStore and Cryptography Interfaces

Within `pkg/datastore` and `pkg/cryptography` there is an interface for datastores and for cryptography engines.
Adding a new datastore or cryptography engine is as simple as implementing one of those interfaces.
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

This project uses [semantic versioning][semver].
To update the version, change the `APP_VERSION` variable in [`.gitlab-ci.yml`][gitlab-ci].
Then make sure add a list of changes to the [CHANGELOG.md][changelog].
Tagging and release will be handled automatically through the [CI/CD pipelines][pipelines] in GitLab.


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
[go-mock]: https://github.com/golang/mock
[gin-swagger]: https://github.com/swaggo/gin-swagger

[vault]: https://www.vaultproject.io/
[redis]: https://redis.io/
[golang]: https://golang.org/
[gnu-make]: https://www.gnu.org/software/make/
[docker]: https://www.docker.com/
[docker-compose]: https://docs.docker.com/compose/
[curl]: https://curl.se/
[semver]: https://semver.org/
[swagger]: https://swagger.io/
