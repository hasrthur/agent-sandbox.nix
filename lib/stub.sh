#!@bash@
# Pinned interpreter: macOS ships bash 3.2, which has no mapfile, so
# `/usr/bin/env bash` would silently assemble an empty command line.
# shellcheck shell=bash
#
# No `set -e`: the last command is the sandbox, and its exit status is the
# status of this script.
set -uo pipefail

DECLARED_ENV=()
UNRESOLVED=()
CREDENTIALS=()
UNMASKABLE=()

# The declared env values are runtime shell expressions; they expand here and
# never enter Python or touch disk. The expansion runs inside a command
# substitution so that `set -u` on an unset variable kills only the subshell,
# letting every failure be collected and reported against its env attribute.
declare_env() {
  local name=$1 expression=$2 value
  if value=$(eval "printf '%s' $expression" 2>/dev/null); then
    DECLARED_ENV+=("$name=$value")
  else
    UNRESOLVED+=("$name = $expression")
  fi
}

PHANTOM_ALPHABET='abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'

# A random string of exactly $1 bytes. Alphanumeric, so it carries no character
# with meaning in a header, a URL, a shell word or the JSON below. The entropy
# is what keeps a phantom from occurring by accident in content the proxy
# scans, where a collision would splice the real credential into whatever
# contained it.
#
# $SRANDOM, not $RANDOM, which is a seeded 15-bit PRNG. Pure bash, so this adds
# no dependency on a PATH the stub deliberately does not trust: every other
# external it runs is pinned to a store path. The modulo bias over 62 is about
# two parts in a hundred thousand, against a value tens of characters long.
mint_phantom() {
  local length=$1 out=''
  while ((${#out} < length)); do
    out+=${PHANTOM_ALPHABET:SRANDOM % ${#PHANTOM_ALPHABET}:1}
  done
  printf '%s' "$out"
}

# Masks one credential: the sandbox receives a phantom of the real value's
# exact byte length, which keeps every request's framing unchanged, and the
# proxy is told to swap the real value back in for the declared hosts.
#
# The real value stays in this shell and the proxy's inherited environment. It
# is never written to disk, never passed on a command line, and never added to
# DECLARED_ENV, which is the only thing that crosses into the sandbox.
mask_env() {
  local name=$1 expression=$2 hosts=$3 value phantom quoted_hosts
  if ! value=$(eval "printf '%s' $expression" 2>/dev/null); then
    UNRESOLVED+=("$name = $expression")
    return
  fi
  # Refused rather than escaped: a value carrying either would have to be
  # quoted into the JSON below, and a credential is the wrong place to find out
  # that the quoting was wrong.
  case $value in
  *\"* | *\\* | *$'\n'*)
    UNMASKABLE+=("$name")
    return
    ;;
  esac
  if ! phantom=$(mint_phantom "${#value}"); then
    UNMASKABLE+=("$name")
    return
  fi
  DECLARED_ENV+=("$name=$phantom")
  # Replaced in this shell too, so an env value declared from this name — an
  # authorization header computed from a token — is built from the phantom.
  export "$name=$phantom"
  quoted_hosts=$(
    IFS=,
    for host in $hosts; do
      printf '%s"%s"' "${sep-}" "$host"
      sep=,
    done
  )
  CREDENTIALS+=("{\"phantom\":\"$phantom\",\"real\":\"$value\",\"hosts\":[$quoted_hosts]}")
}

# shellcheck source=/dev/null
source "@envFragment@"

if ((${#UNRESOLVED[@]})); then
  {
    echo "@errorPrefix@ could not resolve these env values:"
    echo
    for entry in "${UNRESOLVED[@]}"; do
      echo "  $entry"
    done
    echo
    echo "Each value is a shell expression evaluated at launch; anything it"
    echo "references must be set in the shell you launch from."
  } >&2
  exit 1
fi

if ((${#UNMASKABLE[@]})); then
  {
    echo "@errorPrefix@ could not mask these credentials:"
    echo
    for entry in "${UNMASKABLE[@]}"; do
      echo "  $entry"
    done
    echo
    echo "A masked value must not contain a double quote, a backslash or a"
    echo "newline. The value itself is not shown, because this message is"
    echo "printed to a terminal."
  } >&2
  exit 1
fi

# Exported rather than declared, so it reaches the proxy through the
# environment the launcher hands it and never crosses into the sandbox, whose
# environment is rebuilt from DECLARED_ENV alone.
if ((${#CREDENTIALS[@]})); then
  SANDBOX_PROXY_CREDENTIALS="[$(
    IFS=,
    printf '%s' "${CREDENTIALS[*]}"
  )]"
  export SANDBOX_PROXY_CREDENTIALS
fi

# Exported so the entry point inside pasta's namespace inherits it too;
# `env -i` clears it before bubblewrap.
export PYTHONPATH=@launcher@

if ! SESSION_DIR=$("@python@" -P -s -S -m launcher.prepare "@spec@"); then
  exit 1
fi

# This shell does not exec, so it is the sandbox's parent until the session
# ends; its pid is how a later launch's prune tells a finished session from a
# running one.
echo $$ >"$SESSION_DIR/stub.pid"

# Armed only now: before this point there is no session to clean up, and
# prepare tears down its own failures. $? is captured first because it is the
# sandbox's exit status, and every command in the trap body overwrites it.
trap 'STATUS=$?; "@python@" -P -s -S -m launcher.cleanup "$SESSION_DIR" "$STATUS"' EXIT

mapfile -d '' ARGV_BEFORE_ENV < "$SESSION_DIR/argv-before-env"
mapfile -d '' ARGV_AFTER_ENV < "$SESSION_DIR/argv-after-env"

"${ARGV_BEFORE_ENV[@]}" "${DECLARED_ENV[@]}" "${ARGV_AFTER_ENV[@]}" "$@"
