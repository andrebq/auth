#!/usr/bin/env bash

declare -r test_script="${0}"

declare http_cli=http
[[ -n "$HTTPIE" ]] && {
	http_pie="${HTTPIE}"
}
readonly http_cli

declare jq=jq
[[ -n "$JQ" ]] && {
	jq="${JQ}"
}
readonly jq

declare -r test_user="test"
declare -r test_password="test"

set -eou pipefail

function echo-err {
	echo "[ERR]: ${@}" 1>&2
}

function echo-ok {
	echo "[OK]: ${@}" 1>&2
}

function echo-info {
	echo "[INFO]: ${@}" 1>&2
}

which "${http_cli}" > /dev/null || {
	echo-err "Tool ${http_cli} not found, either install httpie (https://httpie.io/cli) or override $$HTTPIE variable with the correct path"
	exit 1
}

which "${jq}" > /dev/null || {
	echo-err "Tool ${jq} not found, either install "${jq}" or override $$"${jq}" variable with the correct path"
	exit 1
}

echo-ok "${http_cli} found"
echo-ok "${jq} found"

echo-info "Trying to login with ${test_user}:${test_password}"
"${http_cli}" --quiet --check-status POST "$AUTH_ENDPOINT/auth/login" login="${test_user}" password="${test_password}"
echo-ok "Authentication was a success"

echo-info "Starting a new session with ${test_user}:${test_password}"
declare token
token=$("${http_cli}" --body --check-status POST "$AUTH_ENDPOINT/session" login="${test_user}" password="${test_password}" | "${jq}" -r '.token')
readonly token
echo-ok "Session created with token ${token}"

echo-info "Authenticating using token from previous session"
"${http_cli}" --quiet --check-status POST "$AUTH_ENDPOINT/auth/token" token="${token}"
echo-ok "Token authentication was a success"


echo-info "Trying to authenticate with invalid credentials not-a-user:not-a-password"
"${http_cli}" -qq --check-status POST "$AUTH_ENDPOINT/auth/login" login="not-a-user" password="not-a-password" 2> /dev/null && {
	echo-err "Invalid credentials got a valid response!"
	exit 1
} || {
	echo-ok "Server rejected invalid credentials for login"
}

echo-info "Trying to authenticate with invalid token"
"${http_cli}" -qq --check-status POST "$AUTH_ENDPOINT/auth/token" token="not-a-token:asdf" 2> /dev/null && {
	echo-err "Invalid token got a valid response!"
	exit 1
} || {
	echo-ok "Server rejected invalid token for login"
}

echo-info "Trying to authenticate with invalid token"
"${http_cli}" -qq --check-status POST "$AUTH_ENDPOINT/session" login="not-a-user" password="not-a-password" 2> /dev/null && {
	echo-err "Invalid token got a valid response!"
	exit 1
} || {
	echo-ok "Server rejected invalid token for login"
}

echo-ok "TEST PASSED: ${test_script}"
