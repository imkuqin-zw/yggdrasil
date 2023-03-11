protoc \
  --go_out=../protogen --go_opt=paths=source_relative \
  --yggdrasil-rpc_out=../protogen --yggdrasil-rpc_opt=paths=source_relative \
  --go-grpc_out=../protogen --go-grpc_opt=paths=source_relative \
  --yggdrasil-reason_out=../protogen --yggdrasil-reason_opt=paths=source_relative \
  -I . -I ../helloword/proto \
  ./*/*.proto
