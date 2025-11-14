# 多阶段构建Dockerfile

# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o example/basic/main ./example/basic

# 运行阶段
FROM gcr.io/distroless/static:nonroot

# 从构建阶段复制CA证书
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 从构建阶段复制时区信息
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 创建非root用户
USER nonroot:nonroot

# 复制构建的二进制文件
COPY --from=builder /app/example/basic/main /app/main

# 设置工作目录
WORKDIR /app

# 暴露端口
EXPOSE 50051 8080 8090

# 运行应用
ENTRYPOINT ["/app/main"]