protoc \
  --go_out=./ --go_opt=paths=source_relative \
  --yggdrasil-rpc_out=./ --yggdrasil-rpc_opt=paths=source_relative \
  --yggdrasil-reason_out=./ --yggdrasil-reason_opt=paths=source_relative \
  -I . \
  ./*.proto
