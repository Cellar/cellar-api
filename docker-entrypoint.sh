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

if [ -z "$VAULT_APPROLE_ROLE_ID" ]; then
  [ -z "$VAULT_APPROLE_ROLE_ID_FILE" ] || export VAULT_APPROLE_ROLE_ID=$(cat "$VAULT_APPROLE_ROLE_ID_FILE")
fi

if [ -z "$VAULT_APPROLE_SECRET_ID" ]; then
  [ -z "$VAULT_APPROLE_SECRET_ID_FILE" ] || export VAULT_APPROLE_SECRET_ID=$(cat "$VAULT_APPROLE_SECRET_ID_FILE")
fi

if [ -z "$VAULT_TOKEN_NAME" ]; then
  [ -z "$VAULT_TOKEN_NAME_FILE" ] || export VAULT_TOKEN_NAME=$(cat "$VAULT_TOKEN_NAME_FILE")
fi

verify_present "REDIS_HOST" "$REDIS_HOST"

verify_present "VAULT_ADDRESS" "$VAULT_ADDRESS"
verify_present "VAULT_AUTH_BACKEND" "$VAULT_AUTH_BACKEND"
verify_present "VAULT_APPROLE_ROLE_ID" "$VAULT_ROLE_ID"
verify_present "VAULT_APPROLE_SECRET_ID" "$VAULT_SECRET_ID"
verify_present "VAULT_TOKEN_NAME" "$VAULT_TOKEN_NAME"

exec /app/cellar $@

