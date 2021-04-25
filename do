#!/usr/bin/env bash

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

APP_NAME="kubectl-nearby"
BUILD_DIRECTORY="bin"

function _go_files() {
  find . -type f -name '*.go' -not -name '*_test.go' \
    | tr '\r\n' ' '
}

# DOCUMENTATION: Show help
function help() {
  local function_name
  while IFS= read -r line; do
    function_name="$(echo "$line" | sed -e 's/^function //' -e 's/() {$//')"
    echo "${function_name}"
    while IFS= read -r nearby_line; do
      if echo "${nearby_line}" | grep -q -E '^# DOCUMENTATION:\s?'; then
        echo "${nearby_line}" | sed -e 's/# DOCUMENTATION: /      /' -e 's/# DOCUMENTATION:/      /'
      else
        break
      fi
    done <<< $(grep --before-context 10 "${line}" "${0}" | tail -r | grep -v "${line}") \
      | tail -r
    echo
  done <<< $(grep -E '^function .* {' "${0}" | grep -v -E 'function _.*' | sort)
}

# DOCUMENTATION: Build the binary
function build() {
  local binary_path="${BUILD_DIRECTORY}/${APP_NAME}"
  echo "Building ${binary_path}"
  mkdir -p "${BUILD_DIRECTORY}"
  go build -o "${binary_path}" $(_go_files)
}

# DOCUMENTATION: Run the binary.
# DOCUMENTATION: Arguments and options will be passed to the binary.
# DOCUMENTATION: EXAMPLE:
# DOCUMENTATION:   ./do run pods --namespace staging
function run() {
  go run $(_go_files) "$@"
}

if [[ $# -lt 1 ]]; then
  help
else
  command="${1}"
  shift
  "${command}" "${@}"
fi
