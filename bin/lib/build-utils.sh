#!/usr/bin/env bash

function _go_files() {
  find . -type f -name '*.go' -depth 1 -not -name '*_test.go' \
    | tr '\r\n' ' '
}

function _ldflags() {
  local version="${1:-development}"
  local package_name="main"
  local current_time
  current_time="$(date -u "+%Y-%m-%dT%H:%M:%SZ")" # -u is the UTC flag that works on Linux and Mac
  local git_commit
  git_commit="$(git rev-parse HEAD)"
  local git_tree_state
  if [[ "$(git status --short)" == "" ]]; then
    git_tree_state="Clean"
  else
    git_tree_state="Dirty"
  fi

  echo -n \
  -ldflags=" \
    -X '${package_name}.version=${version}' \
    -X '${package_name}.buildDate=${current_time}' \
    -X '${package_name}.gitCommit=${git_commit}' \
    -X '${package_name}.gitTreeState=${git_tree_state}' \
  "
}
