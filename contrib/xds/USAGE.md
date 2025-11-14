# XDS Contrib Component 使用指南

## 概述

Yggdrasil XDS组件实现了基于Envoy XDS协议的proxyless微服务架构，允许应用直接与Istio等控制平面通信，无需sidecar代理。

## 架构图

```
┌─────────────────┐    XDS API    ┌─────────────────┐
│   Yggdrasil     │◄─────────────►│   Istio/Envoy   │
│   Application   │               │   Control Plane │
│                 │               │                 │
│ ┌─────────────┐ │               │ ┌─────────────┐ │
│ │   gRPC      │ │               │ │  Pilot      │ │
│ │  Server     │ │               │ │  (LDS/RDS)  │ │
│ └─────────────┘ │               │ └─────────────┘ │
│                 │               │                 │
│ ┌─────────────┐ │               │ ┌─────────────┐ │
│ │   gRPC      │ │               │ │  Citadel    │ │
│ │  Client     │ │               │ │  (SDS)      │ │
│ └─────────────┘ │               │ └─────────────┘ │
└─────────────────┘               └─────────────────┘
```

## 核心组件

### 1. XDS Client
- 与控制平面建立gRPC连接
- 订阅配置变更（Listener, Route, Cluster, Endpoint）
- 管理配置缓存和生命周期

### 2. Resolver
- 实现服务发现接口
- 从XDS Endpoint资源中解析服务实例
- 支持实时服务实例更新

### 3. Balancer
- 实现负载均衡接口
- 支持多种负载均衡策略
- 基于XDS Cluster配置动态调整

### 4. Interceptor
- 集成认证、授权、限流等功能
- 支持OpenTelemetry追踪
- 提供请求/响应拦截

## 快速集成

### 1. 添加依赖
```go
import (
    _ "github.com/imkuqin-zw/yggdrasil/contrib/xds"
)
```

### 2. 基础配置
```yaml
# config.yaml
xds:
  management_server_addresses:
    - "istiod.istio-system:15010"
  node:
    id: "my-service"
    cluster: "Kubernetes"
```

### 3. 启动服务
```go
func main() {
    // 初始化XDS客户端
    if err := xds.Initialize(context.Background()); err != nil {
        log.Fatal(err)
    }

    // 启动Yggdrasil服务
    yggdrasil.Run("my-service",
        yggdrasil.WithServiceDesc(&MyServiceDesc, &MyService{}),
    )
}
```

## 高级配置

### Istio集成
```yaml
xds:
  management_server_addresses:
    - "istiod.istio-system:15010"
  node:
    id: "my-service-v1"
    cluster: "Kubernetes"
    metadata:
      app: "my-app"
      version: "v1"
      namespace: "default"
  security:
    tls_enabled: true
    ca_cert_file: "/var/run/secrets/istio/root-cert.pem"
    cert_file: "/var/run/secrets/istio/tls.crt"
    key_file: "/var/run/secrets/istio/tls.key"
```

### 多控制平面支持
```yaml
xds:
  management_server_addresses:
    - "primary-controlplane:15010"
    - "backup-controlplane:15010"
  initial_load_timeout: 30s
  backoff:
    max_retries: 20
```

## 负载均衡策略

### 配置示例
```yaml
# Istio VirtualService配置示例
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: my-service
spec:
  host: my-service
  trafficPolicy:
    loadBalancer:
      simple: LEAST_CONN  # 最少连接
    # 或者使用一致性哈希
    # loadBalancer:
    #   consistentHash:
    #     httpHeaderName: "user-id"
```

### 支持的策略
1. **Round Robin** - 轮询
2. **Random** - 随机
3. **Least Request** - 最少请求
4. **Ring Hash** - 环形哈希
5. **Maglev** - Maglev哈希

## 安全配置

### mTLS配置
```yaml
xds:
  security:
    tls_enabled: true
    ca_cert_file: "/etc/certs/ca.crt"
    cert_file: "/etc/certs/tls.crt"
    key_file: "/etc/certs/tls.key"
    server_name: "istiod.istio-system"
```

### JWT认证
```yaml
# Istio RequestAuthentication
apiVersion: security.istio.io/v1beta1
kind: RequestAuthentication
metadata:
  name: my-service
spec:
  selector:
    matchLabels:
      app: my-service
  jwtRules:
  - issuer: "https://my-issuer.com"
    jwksUri: "https://my-issuer.com/.well-known/jwks.json"
```

## 可观测性

### 指标收集
```yaml
yggdrasil:
  stats:
    server: "otel"
    client: "otel"
    config:
      otel:
        enableMetrics: true
```

### 链路追踪
```go
// 自动集成OpenTelemetry
ctx, span := tracer.Start(ctx, "my-operation")
defer span.End()

// 在gRPC调用中自动传播trace context
response, err := client.CallMethod(ctx, request)
```

## 故障排查

### 常见问题

#### 1. 连接控制平面失败
**症状**: `failed to connect to management server`
**解决方案**:
```bash
# 检查网络连接
kubectl exec -it my-pod -- nc -zv istiod.istio-system 15010

# 检查证书
kubectl get secret istio-ca-secret -n istio-system
```

#### 2. 配置加载超时
**症状**: `timeout waiting for initial resources`
**解决方案**:
```yaml
xds:
  initial_load_timeout: 60s  # 增加超时时间
```

#### 3. 服务发现失败
**症状**: `no endpoints found for service`
**解决方案**:
```yaml
# 确认服务注册正确
kubectl get destinationrule -l app=my-service
```

### 调试模式
```yaml
yggdrasil:
  logger:
    level: "debug"
    encoder: "json"
```

### 健康检查
```go
// 检查XDS状态
if xds.IsInitialized() {
    fmt.Println("XDS is ready")
}

// 获取客户端信息
client := xds.GetClient()
fmt.Printf("Node ID: %s\n", client.GetNode().ID)
```

## 性能调优

### 1. 连接优化
```yaml
xds:
  backoff:
    base_interval: 100ms    # 减少重试间隔
    max_interval: 10s       # 减少最大间隔
```

### 2. 缓存优化
- 使用内存缓存减少XDS调用
- 配置合适的资源刷新间隔

### 3. 负载均衡优化
- 根据服务特性选择合适的LB策略
- 配置健康检查权重

## 最佳实践

### 1. 优雅启动
```go
func main() {
    ctx := context.Background()

    // 先初始化XDS
    if err := xds.Initialize(ctx); err != nil {
        log.Fatal(err)
    }

    // 等待配置就绪
    for i := 0; i < 30; i++ {
        if xds.IsInitialized() {
            break
        }
        time.Sleep(1 * time.Second)
    }

    // 再启动服务
    yggdrasil.Run("my-service", ...)
}
```

### 2. 优雅关闭
```go
// 注册关闭钩子
defers.Register(func() error {
    return xds.Shutdown()
})
```

### 3. 错误处理
```go
client := yggdrasil.NewClient("external-service",
    yggdrasil.WithUnaryInterceptor(func(ctx context.Context, req interface{},
        info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // 添加重试逻辑
        for i := 0; i < 3; i++ {
            resp, err := handler(ctx, req)
            if err == nil {
                return resp, nil
            }
            time.Sleep(time.Second * time.Duration(i+1))
        }
        return nil, err
    }),
)
```

## 示例项目

查看example目录下的完整示例：

- [基础示例](./example/basic) - 简单的XDS集成
- [安全示例](./example/security) - mTLS认证
- [可观测性示例](./example/observability) - 监控和追踪

## 更多资源

- [Envoy XDS协议文档](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol)
- [Istio官方文档](https://istio.io/docs/)
- [Yggdrasil框架文档](https://github.com/imkuqin-zw/yggdrasil)