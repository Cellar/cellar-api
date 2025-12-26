GREEN := "\033[0;32m"
NC := "\033[0;0m"

IMAGE_NAME ?= cellar-api
IMAGE_TAG ?= local

APP_VERSION ?= 0.0.0

PID_FILE := /tmp/cellar-api.pid

PACKAGE_TOKEN ?= ""
PACKAGE_ARCH ?= unknown
PACKAGE_REGISTRY_URL ?= localhost/projects/project-id/packages/generic/
PACKAGE_NAME ?= cellar-api
PACKAGE_ID := ${PACKAGE_NAME}-${APP_VERSION}-${PACKAGE_ARCH}
PACKAGE_URL := ${PACKAGE_REGISTRY_URL}/${PACKAGE_NAME}/${APP_VERSION}/${PACKAGE_ID}

RELEASE_TAG := v${APP_VERSION}
RELEASE_NAME := "Release ${PACKAGE_NAME} ${RELEASE_TAG}"

REDOC_FILE ?= cellar-api-reference

VAULT_LOCAL_ADDR ?= http://127.0.0.1:8200
VAULT_ROOT_TOKEN ?= vault-admin
VAULT_ENCRYPTION_TOKEN_NAME ?= cellar-key
VAULT_ROLE_NAME ?= cellar-testing

VAULT_REQUEST := @curl --header "X-Vault-Token: ${VAULT_ROOT_TOKEN}"


LOG := @sh -c '\
	   printf ${GREEN}; \
	   echo -e "\n> $$1\n"; \
	   printf ${NC}' VALUE

-include .env

.PHONY: run build publish

targets:
	@awk -F'::?[[:space:]]*' '/^[a-zA-Z0-9][^$#\/\t=]*::?([^=]|$$)/ { \
		gsub(/^[[:space:]]+|[[:space:]]+$$/, "", $$1); \
		gsub(/^[[:space:]]+|[[:space:]]+$$/, "", $$2); \
		split($$1,A,/ /); \
		target=A[1]; \
		deps=$$2; \
		printf "%s", target; \
		if (deps) printf " → %s", deps; \
		print "" \
	}' $(MAKEFILE_LIST)

swag-init:
	$(LOG) "Generating Swagger documentation"
	@swag i --parseDependency -g main.go -dir pkg/controllers -o docs --ot json,go
	@curl -X 'POST' \
		'https://converter.swagger.io/api/convert' \
		-H 'accept: application/json' \
		-H 'Content-Type: application/json' \
		-d @docs/swagger.json | jq > docs/swagger3.json
	@mv docs/swagger3.json docs/swagger.json


redoc:
	$(LOG) "Generating redoc site"
	@npx @redocly/cli build-docs \
		--config .redocly.yaml \
		-o ${REDOC_FILE} \
		cellar

generate-mocks:
	$(LOG) "Running go generate"
	@go generate ./...

test: test-unit test-integration test-acceptance

test-unit:
	$(LOG) "Running unit tests"
	@go test ./...

test-integration:
	$(LOG) "Running integration tests"
	@CRYPTOGRAPHY_VAULT_ENABLED=true \
	 CRYPTOGRAPHY_VAULT_AUTH_MOUNT_PATH=approle \
	 CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID=${CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID} \
	 CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID=${CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID} \
	 CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME=${CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME} \
	 go test -tags=integration -race ./testing/integration/...

test-acceptance:
	$(LOG) "Running acceptance tests"
	@go test -tags=acceptance -race ./testing/acceptance/...

format:
	$(LOG) "Formatting Go code"
	@gofmt -w .

format-check:
	$(LOG) "Checking Go code formatting"
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$unformatted"; \
		echo ""; \
		echo "Run 'make format' to fix formatting"; \
		exit 1; \
	fi

lint:
	$(LOG) "Running linter"
	@golangci-lint run

run:
	$(LOG) "Running Cellar"
	@CRYPTOGRAPHY_VAULT_ENABLED=true \
	 CRYPTOGRAPHY_VAULT_AUTH_MOUNT_PATH=approle \
	 CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID=${CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID} \
	 CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID=${CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID} \
	 CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME=${CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME} \
	 go run cmd/cellar/main.go

run-daemon:
	$(LOG) "Starting Cellar"
	@go build -o cellar-bin cmd/cellar/main.go && chmod +x cellar-bin
	@nohup env \
		CRYPTOGRAPHY_VAULT_ENABLED=true \
		CRYPTOGRAPHY_VAULT_AUTH_MOUNT_PATH=approle \
		CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID=${CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID} \
		CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID=${CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID} \
		CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME=${CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME} \
		./cellar-bin > /tmp/cellar-api.log 2>&1 & echo $$! > ${PID_FILE}
	@sleep 2
	@echo "Cellar API started (PID: $$(cat ${PID_FILE}), logs: /tmp/cellar-api.log)"

stop-daemon:
	$(LOG) "Stopping Cellar"
	@if [ -f ${PID_FILE} ]; then \
		kill -s TERM $$(cat ${PID_FILE}) 2>/dev/null || true; \
		rm -f ${PID_FILE}; \
		rm -f cellar-bin; \
	else \
		echo "No PID file found at ${PID_FILE}"; \
	fi

build:
	$(LOG) "Building all source files"
	go build ./...

package:
	$(LOG) "Building cellar binary '${PACKAGE_ID}'"
	@go build -o ${PACKAGE_ID} -ldflags="-X main.version=${APP_VERSION}" cmd/cellar/main.go

package-lambda:
	$(LOG) "Building cellar binary for lambda '${PACKAGE_ID}'"
	@mkdir -p dist
	@rm -f dist/cellar-api.zip
	@go build -o dist/bootstrap -ldflags="-X main.version=${APP_VERSION}" cmd/cellar-lambda/main.go
	@cd dist && \
	 zip cellar-api.zip bootstrap
	@rm dist/bootstrap

publish: package
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

vault-configure: vault-enable-transit vault-enable-auth

vault-enable-transit:
	$(LOG) "Enabling the transit secrets engine with a single key"
	$(VAULT_REQUEST) -sX POST \
		--data '{"type": "transit"}' \
		${VAULT_LOCAL_ADDR}/v1/sys/mounts/transit
	$(VAULT_REQUEST) -sX POST \
		${VAULT_LOCAL_ADDR}/v1/transit/keys/${VAULT_ENCRYPTION_TOKEN_NAME}

vault-enable-auth:
	$(LOG) "Enabling approle authentication transit secrets engine"
	$(VAULT_REQUEST) -sX POST \
		--data '{"type": "approle"}' \
		${VAULT_LOCAL_ADDR}/v1/sys/auth/approle
	$(LOG) "Adding role ${VAULT_ROLE_NAME} with full access to transit engine"
	$(VAULT_REQUEST) -sX PUT \
		--data '{"name":"transit","policy":"path \"transit/*\" {\n  capabilities = [ \"create\", \"read\", \"update\", \"delete\", \"list\" ]\n}"}' \
		${VAULT_LOCAL_ADDR}/v1/sys/policy/transit
	$(VAULT_REQUEST) -sX POST \
		--data '{"policies": "transit"}' \
		${VAULT_LOCAL_ADDR}/v1/auth/approle/role/${VAULT_ROLE_NAME}

vault-role-id:
	$(VAULT_REQUEST) -sX GET \
		${VAULT_LOCAL_ADDR}/v1/auth/approle/role/${VAULT_ROLE_NAME}/role-id \
		| jq -r '.data.role_id'

vault-secret-id:
	$(VAULT_REQUEST) -sX POST \
		${VAULT_LOCAL_ADDR}/v1/auth/approle/role/${VAULT_ROLE_NAME}/secret-id \
		| jq -r '.data.secret_id'

services: clean-services services-api-dependencies services-vault-wait vault-configure services-env

services-api-dependencies:
	@[ -f ".env" ] && rm -f .env
	@touch .env
	$(LOG) "Starting API dependencies"
	@docker compose pull
	@docker compose up -d redis vault

services-vault-wait:
	@timeout 10 \
		sh -c "until [[ $$(docker compose ps --format=json vault | jq '.Status' ) =~ Up ]]; do echo \"waiting for vault\"; sleep 1; done;" || \
		{ echo "Timed out waiting for Vault to startup"; exit 1; }

services-env:
	@[ -f ".env" ] && rm -f .env
	@touch .env
	@echo "CRYPTOGRAPHY_VAULT_ENABLED=true" >> .env
	@echo "CRYPTOGRAPHY_VAULT_AUTH_MOUNT_PATH=approle" >> .env
	@echo "CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID=$$(make -s vault-role-id)" >> .env
	@echo "CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID=$$(make -s vault-secret-id)" >> .env
	@echo "CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME=${VAULT_ENCRYPTION_TOKEN_NAME}" >> .env

clean-services:
	@[ -f ".env" ] || touch .env
	@docker compose down
	@docker compose rm -svf
	@basename ${PWD} | xargs -I % docker volume rm -f %_redis_data
