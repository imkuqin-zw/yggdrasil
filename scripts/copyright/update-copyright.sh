#!/bin/bash -e

#
# Copyright 2022 The KodeRover Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Add the copyright header for all fo files if not.

CURRENT_PATH=$(dirname "$BASH_SOURCE")
BOILERPLATE=$CURRENT_PATH/boilerplate.go.txt

PROTECT_ROOT=$CURRENT_PATH
VENDOR_PATH=$PROTECT_ROOT/third_party

repeat() {
  local start=1
  local end=${1:-80}
  local str="${2:-=}"
  # shellcheck disable=SC2155
  local range=$(seq $start "$end")
  # shellcheck disable=SC2034
  for i in $range; do
    echo -n "${str}"
  done
}

init() {
  echo "init path"
  count=$(echo "${CURRENT_PATH}" | grep -o '/' | wc -l)
  relativePath=$(repeat "${count}" '/..')
  PROTECT_ROOT=$CURRENT_PATH$relativePath
  VENDOR_PATH=$PROTECT_ROOT/third_party
  REMOTE_PATH=$PROTECT_ROOT/pkg/remote
}

addCopyrightHeader() {
  echo "add copyright header"
  # shellcheck disable=SC2044
  for file in $(find "$PROTECT_ROOT" -not -path "$VENDOR_PATH/*" -not -path "$REMOTE_PATH/*" -not -name "zz_generated*" -not -name "escape.go" -not -name "*.pb.go" -not -name "signer.go" -type f -name \*.go); do
    if [[ $(grep -n "\/\/ Copyright" -m 1 "$file" | cut -f1 -d:) == 1 ]] || [[ $(grep -n "\/\*" -m 1 "$file" | cut -f1 -d:) == 1 && $(grep -n "Copyright" -m 1 "$file" | cut -f1 -d:) == 2 ]]; then
      # the file already has a copyright.
      continue
    else
      cat "$BOILERPLATE" >"$file".tmp
      echo "" >>"$file".tmp
      cat "$file" >>"$file".tmp
      mv "$file".tmp "$file"
    fi
  done
}

main() {
  init
  addCopyrightHeader
}

main
