# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum* ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY main.go ./

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o xds-server .

# 运行阶段
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/xds-server .

EXPOSE 15010

CMD ["./xds-server"]
