#!/usr/bin/env sh

variable_not_found() {
    echo "$1 is not set but is required"
    exit 1
}

verify_present() {
    name=$1
    val=$2
    if [ -z "$val" ]; then
        echo "$name is required but not set"
        exit 1
    fi
}


if [ -z "$DATASTORE_REDIS_PASSWORD" ]; then
  [ -z "$DATASTORE_REDIS_PASSWORD_FILE" ] || export DATASTORE_REDIS_PASSWORD=$(cat "$DATASTORE_REDIS_PASSWORD_FILE")
fi

if [ -z "$CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID" ]; then
  [ -z "$CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID_FILE" ] || export CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID=$(cat "$CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID_FILE")
fi

if [ -z "$CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID" ]; then
  [ -z "$CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID_FILE" ] || export CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID=$(cat "$CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID_FILE")
fi

if [ -z "$CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME" ]; then
  [ -z "$CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME_FILE" ] || export CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME=$(cat "$CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME_FILE")
fi

verify_present "DATASTORE_REDIS_HOST" "$DATASTORE_REDIS_HOST"

exec /app/cellar $@

