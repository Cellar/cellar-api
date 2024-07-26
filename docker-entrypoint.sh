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


if [ -z "$REDIS_PASSWORD" ]; then
  [ -z "$REDIS_PASSWORD_FILE" ] || export REDIS_PASSWORD=$(cat "$REDIS_PASSWORD_FILE")
fi

if [ -z "$VAULT_AUTH_APPROLE_ROLE_ID" ]; then
  [ -z "$VAULT_AUTH_APPROLE_ROLE_ID_FILE" ] || export VAULT_AUTH_APPROLE_ROLE_ID=$(cat "$VAULT_AUTH_APPROLE_ROLE_ID_FILE")
fi

if [ -z "$VAULT_AUTH_APPROLE_SECRET_ID" ]; then
  [ -z "$VAULT_AUTH_APPROLE_SECRET_ID_FILE" ] || export VAULT_AUTH_APPROLE_SECRET_ID=$(cat "$VAULT_AUTH_APPROLE_SECRET_ID_FILE")
fi

if [ -z "$VAULT_ENCRYPTION_TOKEN_NAME" ]; then
  [ -z "$VAULT_ENCRYPTION_TOKEN_NAME_FILE" ] || export VAULT_ENCRYPTION_TOKEN_NAME=$(cat "$VAULT_ENCRYPTION_TOKEN_NAME_FILE")
fi

verify_present "REDIS_HOST" "$REDIS_HOST"

verify_present "VAULT_ADDRESS" "$VAULT_ADDRESS"
verify_present "VAULT_AUTH_MOUNT_PATH" "$VAULT_AUTH_MOUNT_PATH"
verify_present "VAULT_ENCRYPTION_TOKEN_NAME" "$VAULT_ENCRYPTION_TOKEN_NAME"

exec /app/cellar $@

