GREEN := "\033[0;32m"
NC := "\033[0;0m"

IMAGE_NAME ?= cellar-api
IMAGE_TAG ?= local

APP_VERSION ?= 0.0.0

PACKAGE_TOKEN ?= ""
PACKAGE_ARCH ?= unknown
PACKAGE_REGISTRY_URL ?= localhost/projects/project-id/packages/generic/
PACKAGE_NAME ?= cellar-api
PACKAGE_ID := ${PACKAGE_NAME}-${APP_VERSION}-${PACKAGE_ARCH}
PACKAGE_URL := ${PACKAGE_REGISTRY_URL}/${PACKAGE_NAME}/${APP_VERSION}/${PACKAGE_ID}

RELEASE_TAG := v${APP_VERSION}
RELEASE_NAME := "Release ${PACKAGE_NAME} ${RELEASE_TAG}"


LOG := @sh -c '\
	   printf ${GREEN}; \
	   echo -e "\n> $$1\n"; \
	   printf ${NC}' VALUE

.PHONY: run build publish

swag-init:
	$(LOG) "Generating Swagger documentation"
	@cd cmd/cellar && swag init --parseDependency

generate-mocks:
	$(LOG) "Running go generate"
	@go generate ./...

test-unit:
	$(LOG) "Running unit tests"
	@go test ./...

test-integration:
	$(LOG) "Running integration tests"
	@go test -tags=integration testing/...

test-acceptance:
	$(LOG) "Running acceptance tests"
	@go test -tags=acceptance testing/...

run:
	$(LOG) "Running Cellar"
	@go run cellar/cmd/cellar

build:
	$(LOG) "Building all source files"
	go build ./...

publish:
	$(LOG) "Building cellar binary '${PACKAGE_ID}'"
	@go build -o ${PACKAGE_ID} -ldflags="-X main.version=${APP_VERSION}" cellar/cmd/cellar
	$(LOG) "Uploading cellar binary to ${PACKAGE_URL}"
	@curl \
		--header "JOB-TOKEN: ${PACKAGE_TOKEN}" \
		--upload-file ${PACKAGE_ID} \
		${PACKAGE_URL}

release:
	$(LOG) "Creating gitlab release '${RELEASE_NAME}'"
	@release-cli create \
		--name ${RELEASE_NAME}  \
		--tag-name ${RELEASE_TAG} \
		--assets-link '{"name": "${PACKAGE_ID}", "url":"${PACKAGE_URL}"}' \
		--assets-link '{"name": "${IMAGE_NAME}:${IMAGE_TAG}", "url":"https://${IMAGE_NAME}:${IMAGE_TAG}"}'

docker-build:
	$(LOG) "Building docker image '${IMAGE_NAME}:${IMAGE_TAG}"
	@docker build -t ${IMAGE_NAME}:${IMAGE_TAG} --build-arg APP_VERSION=${APP_VERSION} .

docker-run:
	$(LOG) "Running docker image '${IMAGE_NAME}:${IMAGE_TAG}"
	@docker run ${IMAGE_NAME}:${IMAGE_TAG}

docker-publish: docker-build
	$(LOG) "Pushing docker image '${IMAGE_NAME}:${IMAGE_TAG}"
	@docker push ${IMAGE_NAME}:${IMAGE_TAG}
