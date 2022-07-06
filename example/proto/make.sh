protoc \
  --go_out=../protogen --go_opt=paths=source_relative \
  --yggdrasil-rpc_out=../protogen --yggdrasil-rpc_opt=paths=source_relative \
  --yggdrasil-grpc_out=../protogen --yggdrasil-grpc_opt=paths=source_relative \
  --yggdrasil-error_out=../protogen --yggdrasil-error_opt=paths=source_relative \
  -I . -I ../../proto \
  ./*/*.proto
