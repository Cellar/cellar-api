#!/usr/bin/env sh

verify_present() {
    name=$1
    val=$2
    if [ -z "$val" ]; then
        echo "$name is required but not set"
        exit 1
    fi
}

is_true() {
    case "$1" in
        1|t|T|true|TRUE|True) return 0 ;;
        *) return 1 ;;
    esac
}

load_secret_file() {
    name=$1
    eval "value=\${$name:-}"
    eval "file=\${${name}_FILE:-}"

    if [ -n "$value" ] || [ -z "$file" ]; then
        return 0
    fi

    if ! content=$(cat "$file" 2>/dev/null); then
        echo "unable to read ${name}_FILE ($file)"
        exit 1
    fi
    export "$name=$content"
}

for secret in \
    DATASTORE_REDIS_PASSWORD \
    DATASTORE_VALKEY_PASSWORD \
    CRYPTOGRAPHY_VAULT_AUTH_APPROLE_ROLE_ID \
    CRYPTOGRAPHY_VAULT_AUTH_APPROLE_SECRET_ID \
    CRYPTOGRAPHY_VAULT_AUTH_AWSIAM_ROLE \
    CRYPTOGRAPHY_VAULT_AUTH_GCPIAM_ROLE \
    CRYPTOGRAPHY_VAULT_AUTH_KUBERNETES_ROLE \
    CRYPTOGRAPHY_VAULT_ENCRYPTION_TOKEN_NAME \
    CRYPTOGRAPHY_AWS_REGION \
    CRYPTOGRAPHY_AWS_KMS_KEY_ID
do
    load_secret_file "$secret"
done

if is_true "$DATASTORE_REDIS_ENABLED"; then
    verify_present "DATASTORE_REDIS_HOST" "$DATASTORE_REDIS_HOST"
fi

if is_true "$DATASTORE_VALKEY_ENABLED"; then
    verify_present "DATASTORE_VALKEY_HOST" "$DATASTORE_VALKEY_HOST"
fi

exec "${CELLAR_BIN:-/app/cellar}" "$@"
