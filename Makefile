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

REDOC_FILE ?= cellar-api-reference.html

VAULT_LOCAL_ADDR ?= http://127.0.0.1:8200
VAULT_ROOT_TOKEN ?= vault-admin
VAULT_TOKEN_NAME ?= cellar-key
VAULT_ROLE_NAME ?= cellar-testing

VAULT_REQUEST := @curl --header "X-Vault-Token: ${VAULT_ROOT_TOKEN}"


LOG := @sh -c '\
	   printf ${GREEN}; \
	   echo -e "\n> $$1\n"; \
	   printf ${NC}' VALUE

REDOC-CLI := @docker run \
				-u 1000:1000 \
				-v $${PWD}:/data \
				-w /data \
				node:alpine \
				npx redoc-cli

.PHONY: run build publish

swag-init:
	$(LOG) "Generating Swagger documentation"
	@cd cmd/cellar && swag init --parseDependency

swag-publish:
	$(LOG) "Generating redoc site"
	$(REDOC-CLI) bundle \
		-o ${REDOC_FILE} \
		--title "Cellar API ${APP_VERSION}" \
		cmd/cellar/docs/swagger.yaml

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

vault-configure: vault-enable-transit vault-enable-auth

vault-enable-transit:
	$(LOG) "Enabling the transit secrets engine with a single key"
	$(VAULT_REQUEST) -sX POST \
		--data '{"type": "transit"}' \
		${VAULT_LOCAL_ADDR}/v1/sys/mounts/transit
	$(VAULT_REQUEST) -sX POST \
		${VAULT_LOCAL_ADDR}/v1/transit/keys/${VAULT_TOKEN_NAME}

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

services: clean-services
	@[ -f ".env" ] && rm -f .env
	@touch .env
	$(LOG) "Starting API dependencies"
	@docker-compose pull
	@docker-compose up -d redis vault
	@make vault-configure
	@echo "VAULT_ROLE_ID=$$(make -s vault-role-id)" >> .env
	@echo "VAULT_SECRET_ID=$$(make -s vault-secret-id)" >> .env

clean-services:
	@[ -f ".env" ] || touch .env
	@docker-compose rm -svf
	@basename ${PWD} | xargs -I % docker volume rm -f %_redis_data
