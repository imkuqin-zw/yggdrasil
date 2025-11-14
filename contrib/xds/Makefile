# XDS Contrib Component Makefile

.PHONY: build test clean example deps

# 默认目标
all: deps build test

# 安装依赖
deps:
	go mod tidy
	go mod download

# 构建
build:
	go build -v ./...

# 运行测试
test:
	go test -v ./...

# 清理
clean:
	go clean

# 运行基础示例
example-basic:
	cd example/basic && go run main.go

# 运行安全示例
example-security:
	cd example/security && go run main.go

# 运行所有示例
examples: example-basic example-security

# 代码格式化
fmt:
	go fmt ./...

# 代码检查
vet:
	go vet ./...

# 生成gRPC代码
proto:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

# 构建Docker镜像
docker-build:
	docker build -t yggdrasil-xds-example .

# 运行Docker容器
docker-run:
	docker run -p 50051:50051 -p 8080:8080 -p 8090:8090 yggdrasil-xds-example

# 开发模式 (自动重载)
dev:
	go install github.com/cosmtrek/air@latest
	air -c .air.toml