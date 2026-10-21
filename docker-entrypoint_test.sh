#!/usr/bin/env bash
set -uo pipefail

usage() {
  cat <<'EOF'
usage: docker-entrypoint_test.sh [--help]

Runs docker-entrypoint.sh against a stub cellar binary and checks how it
validates datastore settings, resolves *_FILE secrets, and forwards arguments.
Exits non-zero if any scenario fails.
EOF
}

if [[ "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

ENTRYPOINT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/docker-entrypoint.sh"
WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

STUB="$WORKDIR/cellar"
cat > "$STUB" <<'EOF'
#!/usr/bin/env sh
echo "ran"
for arg in "$@"; do echo "arg=$arg"; done
echo "valkey_password=${DATASTORE_VALKEY_PASSWORD:-}"
EOF
chmod +x "$STUB"

failures=0
passes=0

run_entrypoint() {
  env -i PATH="$PATH" CELLAR_BIN="$STUB" "$@" sh "$ENTRYPOINT" 2>&1
}

it() {
  local name=$1 expected=$2 actual=$3
  if [[ "$actual" == *"$expected"* ]]; then
    passes=$((passes + 1))
  else
    failures=$((failures + 1))
    printf 'FAIL: %s\n  expected to contain: %s\n  actual: %s\n' "$name" "$expected" "$actual"
  fi
}

it_not() {
  local name=$1 unexpected=$2 actual=$3
  if [[ "$actual" != *"$unexpected"* ]]; then
    passes=$((passes + 1))
  else
    failures=$((failures + 1))
    printf 'FAIL: %s\n  expected not to contain: %s\n  actual: %s\n' "$name" "$unexpected" "$actual"
  fi
}

for truthy in true TRUE True 1 t T; do
  out=$(run_entrypoint DATASTORE_REDIS_ENABLED="$truthy" DATASTORE_REDIS_HOST=redis)
  it "when redis is enabled with '$truthy', and the redis host is set, it should start cellar" "ran" "$out"

  out=$(run_entrypoint DATASTORE_REDIS_ENABLED="$truthy")
  it "when redis is enabled with '$truthy', and the redis host is missing, it should report the missing host" "DATASTORE_REDIS_HOST is required but not set" "$out"

  out=$(run_entrypoint DATASTORE_VALKEY_ENABLED="$truthy" DATASTORE_VALKEY_HOST=valkey)
  it "when valkey is enabled with '$truthy', and only the valkey host is set, it should start cellar" "ran" "$out"

  out=$(run_entrypoint DATASTORE_VALKEY_ENABLED="$truthy")
  it "when valkey is enabled with '$truthy', and the valkey host is missing, it should report the missing host" "DATASTORE_VALKEY_HOST is required but not set" "$out"
done

out=$(run_entrypoint DATASTORE_VALKEY_ENABLED=true DATASTORE_VALKEY_HOST=valkey)
it_not "when valkey is enabled, and the redis host is not set, it should not require the redis host" "DATASTORE_REDIS_HOST" "$out"

out=$(run_entrypoint DATASTORE_REDIS_ENABLED=false DATASTORE_VALKEY_ENABLED=false)
it "when no datastore is enabled, it should let cellar report the configuration error" "ran" "$out"

out=$(run_entrypoint DATASTORE_REDIS_ENABLED=true DATASTORE_REDIS_HOST=redis DATASTORE_VALKEY_ENABLED=true DATASTORE_VALKEY_HOST=valkey)
it "when both datastores are enabled, it should let cellar report the configuration error" "ran" "$out"

printf 'hunter2' > "$WORKDIR/valkey-password"
out=$(run_entrypoint DATASTORE_VALKEY_ENABLED=true DATASTORE_VALKEY_HOST=valkey DATASTORE_VALKEY_PASSWORD_FILE="$WORKDIR/valkey-password")
it "when valkey reads its password from a file, it should export the password" "valkey_password=hunter2" "$out"

printf 'correct horse * staple' > "$WORKDIR/spaced-password"
out=$(run_entrypoint DATASTORE_VALKEY_ENABLED=true DATASTORE_VALKEY_HOST=valkey DATASTORE_VALKEY_PASSWORD_FILE="$WORKDIR/spaced-password")
it "when a secret file contains spaces and globs, it should export the secret unchanged" "valkey_password=correct horse * staple" "$out"

out=$(run_entrypoint DATASTORE_VALKEY_ENABLED=true DATASTORE_VALKEY_HOST=valkey DATASTORE_VALKEY_PASSWORD_FILE="$WORKDIR/does-not-exist")
it "when a secret file cannot be read, it should report the unreadable file" "unable to read DATASTORE_VALKEY_PASSWORD_FILE" "$out"
it_not "when a secret file cannot be read, it should not start cellar" "ran" "$out"

out=$(run_entrypoint DATASTORE_VALKEY_ENABLED=true DATASTORE_VALKEY_HOST=valkey DATASTORE_VALKEY_PASSWORD=direct DATASTORE_VALKEY_PASSWORD_FILE="$WORKDIR/does-not-exist")
it "when a secret is set directly, it should ignore its file variable" "valkey_password=direct" "$out"

out=$(env -i PATH="$PATH" CELLAR_BIN="$STUB" DATASTORE_REDIS_ENABLED=true DATASTORE_REDIS_HOST=redis sh "$ENTRYPOINT" "two words" second 2>&1)
it "when arguments contain spaces, it should forward each argument intact" "arg=two words" "$out"

echo "${passes} passed, ${failures} failed"
[[ $failures -eq 0 ]]
