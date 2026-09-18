#!/usr/bin/env bash
# Masked credentials (shared across platforms).
#
# maskedCredentials hands the sandbox a phantom of the real value's byte
# length and tells the proxy to swap the real value back in, for the declared
# hosts only. What the upstream received is the only honest witness, so every
# assertion below reads go-httpbin's echo of the request it got rather than
# anything inside the sandbox.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

source "$SCRIPT_DIR/../lib.sh"

echo "=== Masked credential tests (shared) ==="
echo

LOCAL_HTTPBIN_PORT=18920
if nc -z 127.0.0.1 "$LOCAL_HTTPBIN_PORT" 2>/dev/null; then
	echo "FAIL: test setup — 127.0.0.1:$LOCAL_HTTPBIN_PORT already in use" >&2
	exit 1
fi
HTTPBIN_BIN=$(build_host_pkg go-httpbin)/bin/go-httpbin
"$HTTPBIN_BIN" -host 127.0.0.1 -port "$LOCAL_HTTPBIN_PORT" >/tmp/sandbox-httpbin-masked.log 2>&1 &
HTTPBIN_PID=$!
trap 'kill "$HTTPBIN_PID" 2>/dev/null || true' EXIT
_httpbin_ready=0
for _ in 1 2 3 4 5 6 7 8 9 10; do
	if nc -z 127.0.0.1 "$LOCAL_HTTPBIN_PORT" 2>/dev/null; then
		_httpbin_ready=1
		break
	fi
	sleep 0.2
done
if [ "$_httpbin_ready" -ne 1 ]; then
	echo "FAIL: test setup — go-httpbin never came up on 127.0.0.1:$LOCAL_HTTPBIN_PORT" >&2
	exit 1
fi

SANDBOXED=$(build_fixture masked-credentials.nix --argstr httpbinPort "$LOCAL_HTTPBIN_PORT")
SHELL_BIN="$SANDBOXED/bin/sandboxed-bash-masked"

# Shaped like a real token and deliberately not a round multiple of three, so
# the base64 cases below do not pass on a lucky alignment.
export TEST_TOKEN="ghp_realvalue_never_reaches_the_sandbox_01"

echo "--- what the sandbox itself can read ---"

capture "$SHELL_BIN" --norc --noprofile -c 'printf "%s" "$TEST_TOKEN"'
assert_exit_code "launch succeeds with a masked credential declared" 0
assert_output_not_contains "the real value is not readable inside the sandbox" \
	"ghp_realvalue_never_reaches_the_sandbox_01"

capture "$SHELL_BIN" --norc --noprofile -c 'printf "%s" "${#TEST_TOKEN}"'
assert_output_equals "the phantom has the real value's byte length" "42"

# Matched in bash rather than with grep, which the fixture does not put on the
# sandbox's PATH: a missing binary would make this assertion pass on nothing.
capture "$SHELL_BIN" --norc --noprofile -c \
	'e=$(env); if [[ $e == *ghp_realvalue* ]]; then printf found; else printf none; fi'
assert_output_equals "no environment entry carries the real value" "none"

capture "$SHELL_BIN" --norc --noprofile -c 'printf "%s" "${SANDBOX_PROXY_CREDENTIALS:-unset}"'
assert_output_equals "the credential declaration itself does not cross inward" "unset"

# Minted per launch, so a phantom captured from one session is worth nothing in
# the next.
capture "$SHELL_BIN" --norc --noprofile -c 'printf "%s" "$TEST_TOKEN"'
FIRST_PHANTOM="$CAP_OUT"
capture "$SHELL_BIN" --norc --noprofile -c 'printf "%s" "$TEST_TOKEN"'
if [ -n "$FIRST_PHANTOM" ] && [ "$FIRST_PHANTOM" != "$CAP_OUT" ]; then
	echo "PASS: the phantom is regenerated per session"
	PASS=$((PASS + 1))
else
	echo "FAIL: two sessions were handed the same phantom"
	FAIL=$((FAIL + 1))
fi

echo
echo "--- an env value derived from a masked credential ---"

capture "$SHELL_BIN" --norc --noprofile -c 'printf "%s" "$TEST_TOKEN_HEADER"'
assert_output_not_contains "a derived env value does not carry the real value" \
	"ghp_realvalue_never_reaches_the_sandbox_01"

# Compared inside the sandbox, which is the only place both names are readable.
capture "$SHELL_BIN" --norc --noprofile -c \
	'if [[ $TEST_TOKEN_HEADER == "Bearer $TEST_TOKEN" ]]; then printf phantom; else printf something-else; fi'
assert_output_equals "a derived env value carries this session's phantom" "phantom"

capture "$SHELL_BIN" --norc --noprofile -c \
	'curl -sf --max-time 10 -H "Authorization: $TEST_TOKEN_HEADER" https://httpbin.test/headers'
assert_exit_code "a request carrying the derived name succeeds" 0
assert_output_contains "the declared host received the real value through the derived name" \
	"Bearer ghp_realvalue_never_reaches_the_sandbox_01"

echo
echo "--- what the upstream received ---"

# httpbin's /headers echoes the request headers it was sent.
capture "$SHELL_BIN" --norc --noprofile -c \
	'curl -sf --max-time 10 -H "Authorization: Bearer $TEST_TOKEN" https://httpbin.test/headers'
assert_exit_code "a request to the declared host succeeds" 0
assert_output_contains "the declared host received the real value" \
	"Bearer ghp_realvalue_never_reaches_the_sandbox_01"

capture "$SHELL_BIN" --norc --noprofile -c \
	'curl -sf --max-time 10 -H "Authorization: Bearer $TEST_TOKEN" https://pie.test/headers'
assert_exit_code "a request to the undeclared host succeeds" 0
assert_output_not_contains "the undeclared host did not receive the real value" \
	"ghp_realvalue_never_reaches_the_sandbox_01"

# The phantom travels on to a host the credential does not name, so the request
# is inert there rather than refused. Compared inside the sandbox, because the
# phantom is minted per session and the harness never learns it.
capture "$SHELL_BIN" --norc --noprofile -c \
	'r=$(curl -sf --max-time 10 -H "Authorization: Bearer $TEST_TOKEN" https://pie.test/headers)
	 if [[ $r == *"$TEST_TOKEN"* ]]; then printf phantom; else printf something-else; fi'
assert_output_equals "the undeclared host received the phantom instead" "phantom"

echo
echo "--- the encoded form, which no configuration names ---"

# curl -u builds "Basic base64(user:token)", in which the token appears nowhere
# literally. Nothing in the fixture mentions basic auth or this username.
capture "$SHELL_BIN" --norc --noprofile -c \
	'curl -sf --max-time 10 -u "x-access-token:$TEST_TOKEN" https://httpbin.test/basic-auth-echo 2>/dev/null ||
	 curl -sf --max-time 10 -u "x-access-token:$TEST_TOKEN" https://httpbin.test/headers'
assert_exit_code "a basic-auth request to the declared host succeeds" 0
EXPECTED_BASIC=$(printf 'x-access-token:%s' "$TEST_TOKEN" | base64 | tr -d '\n')
assert_output_contains "the declared host received the real value, base64-encoded" \
	"$EXPECTED_BASIC"

# The same encoding with a username whose length does not divide by three: the
# case a pre-encoded registration would silently miss.
capture "$SHELL_BIN" --norc --noprofile -c \
	'curl -sf --max-time 10 -u "oauth2:$TEST_TOKEN" https://httpbin.test/headers'
EXPECTED_OAUTH2=$(printf 'oauth2:%s' "$TEST_TOKEN" | base64 | tr -d '\n')
assert_output_contains "a username whose length does not divide by three also substitutes" \
	"$EXPECTED_OAUTH2"

# The token in the username slot, with no password, as several APIs take it.
capture "$SHELL_BIN" --norc --noprofile -c \
	'curl -sf --max-time 10 -u "$TEST_TOKEN:" https://httpbin.test/headers'
EXPECTED_USERSLOT=$(printf '%s:' "$TEST_TOKEN" | base64 | tr -d '\n')
assert_output_contains "the token in the username slot also substitutes" \
	"$EXPECTED_USERSLOT"

capture "$SHELL_BIN" --norc --noprofile -c \
	'curl -sf --max-time 10 -u "x-access-token:$TEST_TOKEN" https://pie.test/headers'
assert_output_not_contains "the undeclared host received no encoded credential either" \
	"$EXPECTED_BASIC"

echo
echo "--- the proxy log ---"

PROXY_LOG=$(find "${AGENT_SANDBOX_SESSIONS_ROOT:-$HOME/.local/state/agent-sandbox}" \
	-name proxy.log -newer "$SANDBOXED" 2>/dev/null | head -1 || true)
if [ -n "$PROXY_LOG" ]; then
	if grep -q "ghp_realvalue\|Authorization" "$PROXY_LOG"; then
		echo "FAIL: the proxy log carries a credential or a header value"
		FAIL=$((FAIL + 1))
	else
		echo "PASS: the proxy log carries neither a credential nor a header value"
		PASS=$((PASS + 1))
	fi
else
	echo "SKIP: no proxy.log found to check"
fi

print_results
exit_status
