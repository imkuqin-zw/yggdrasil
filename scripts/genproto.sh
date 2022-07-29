#!/usr/bin/env bash

# shellcheck disable=SC2128
CURRENT_PATH=$(dirname "$BASH_SOURCE")
ROOT_PATH=$(dirname "$CURRENT_PATH")
API_PATH="$ROOT_PATH/api"
GEN_PATH="$ROOT_PATH/pkg/genproto"

if [ ! -d "$GEN_PATH" ]; then
        mkdir -p "$GEN_PATH"
fi

protoc -I=. -I="$ROOT_PATH" --go_out="$GEN_PATH" --go_opt=paths=source_relative \
  --yggdrasil-rpc_out="$GEN_PATH" --yggdrasil-rpc_opt=paths=source_relative \
  --yggdrasil-grpc_out="$GEN_PATH" --yggdrasil-grpc_opt=paths=source_relative \
  --yggdrasil-error_out="$GEN_PATH" --yggdrasil-error_opt=paths=source_relative \
  "$API_PATH"/*.proto
echo "proto build complete"