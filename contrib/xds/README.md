# XDS Contrib Component

Yggdrasil框架的XDS (eXtended Discovery Service) 协议支持组件，实现了基于Envoy XDS协议的proxyless微服务架构。

## 概述

该组件让Yggdrasil框架能够作为数据平面直接与Istio、Envoy等控制平面通信，实现无需sidecar proxy的微服务治理。

> **⚠️ 当前实现状态**: 本组件提供了完整的XDS集成框架，支持基础的服务发现和负载均衡功能。完整的XDS协议实现正在开发中，当前版本提供模拟资源生成用于开发测试。

### 核心特性

- **服务发现**: 基于XDS协议的动态服务发现
- **负载均衡**: 支持多种负载均衡策略（轮询、随机、最少请求、一致性哈希等）
- **安全认证**: 基于XDS配置的mTLS、JWT认证
- **流量管理**: 支持路由规则、限流、熔断等
- **可观测性**: 集成OpenTelemetry链路追踪和监控
- **框架集成**: 与yggdrasil框架无缝集成

## 快速开始

### 1. 添加依赖

```go
import (
	_ "github.com/imkuqin-zw/yggdrasil/contrib/xds"
)
```

### 2. 配置XDS

```yaml
yggdrasil:
  application:
    namespace: "default"

  # 使用XDS解析器和负载均衡器
  client:
    service-name:
      resolver: "xds"
      balancer: "xds"

  # 配置XDS拦截器
  interceptor:
    unaryServer: ["xds"]
    streamServer: ["xds"]
    unaryClient: ["xds"]
    streamClient: ["xds"]

xds:
  management_server_addresses:
    - "istiod.istio-system:15010"

  node:
    id: "yggdrasil-client"
    cluster: "yggdrasil-cluster"
    metadata:
      app: "my-app"
      version: "v1.0.0"

  xds_resources:
    ads: true  # 使用聚合发现服务
    listener_names: ["*"]
    route_config_names: ["*"]
    cluster_names: ["*"]

  security:
    tls_enabled: true
    ca_cert_file: "/etc/certs/ca.crt"
    cert_file: "/etc/certs/tls.crt"
    key_file: "/etc/certs/tls.key"
    server_name: "istiod.istio-system"

  backoff:
    base_interval: 500ms
    max_interval: 30s
    max_retries: 10

  initial_load_timeout: 15s
```

### 3. 初始化XDS客户端

```go
package main

import (
	"context"
	"log"

	"github.com/imkuqin-zw/yggdrasil"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/xds"
)

func main() {
	ctx := context.Background()

	// 初始化XDS客户端
	if err := xds.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize XDS: %v", err)
	}

	// 启动Yggdrasil服务
	if err := yggdrasil.Run("my-service",
		yggdrasil.WithServiceDesc(&myServiceDesc, &myServiceImpl{}),
	); err != nil {
		log.Fatalf("Failed to run service: %v", err)
	}
}
```

### 4. 使用XDS客户端

```go
package main

import (
	"context"
	"fmt"

	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/contrib/xds"
)

func main() {
	// 检查XDS是否就绪
	if xds.IsInitialized() {
		fmt.Println("XDS client is ready")
	}

	// 获取XDS客户端
	client := xds.GetClient()

	// 创建客户端连接
	conn, err := yggdrasil.NewClient("external-service")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer conn.Close()

	// 使用客户端调用服务
	client := NewExternalServiceClient(conn)
	resp, err := client.CallMethod(ctx, &Request{})
	if err != nil {
		log.Fatalf("Call failed: %v", err)
	}
}
```

## 配置详解

### 基础配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `management_server_addresses` | []string | `["xds-server:15010"]` | XDS管理服务器地址 |
| `initial_load_timeout` | Duration | `15s` | 初始资源加载超时时间 |

### 节点配置

| 配置项 | 类型 | 说明 |
|--------|------|------|
| `id` | string | 节点唯一标识 |
| `cluster` | string | 节点所属集群 |
| `metadata` | map[string]string | 节点元数据 |

### 资源配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `ads` | bool | `true` | 是否启用聚合发现服务 |
| `listener_names` | []string | `["*"]` | 监听的Listener名称 |
| `route_config_names` | []string | `["*"]` | 监听的路由配置名称 |
| `cluster_names` | []string | `["*"]` | 监听的集群名称 |

### 安全配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `tls_enabled` | bool | `true` | 是否启用TLS |
| `ca_cert_file` | string | - | CA证书文件路径 |
| `cert_file` | string | - | 客户端证书文件路径 |
| `key_file` | string | - | 客户端私钥文件路径 |
| `server_name` | string | - | TLS服务器名称 |
| `insecure_skip_verify` | bool | `false` | 是否跳过TLS验证 |

### 重试配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `base_interval` | Duration | `500ms` | 基础退避间隔 |
| `max_interval` | Duration | `30s` | 最大退避间隔 |
| `max_retries` | int | `10` | 最大重试次数 |

## 负载均衡策略

支持以下负载均衡策略：

### 1. 轮询 (Round Robin)
```yaml
# 在XDS集群配置中设置
load_balancing_policy:
  round_robin: {}
```

### 2. 随机 (Random)
```yaml
load_balancing_policy:
  random: {}
```

### 3. 最少请求 (Least Request)
```yaml
load_balancing_policy:
  least_request: {}
```

### 4. 一致性哈希 (Consistent Hash)
```yaml
load_balancing_policy:
  ring_hash:
    hash_function: "MURMUR_HASH_2"
    minimum_ring_size: 1024
```

### 5. Maglev哈希
```yaml
load_balancing_policy:
  maglev: {}
```

## 安全特性

### mTLS认证
```yaml
xds:
  security:
    tls_enabled: true
    ca_cert_file: "/etc/certs/ca.crt"
    cert_file: "/etc/certs/tls.crt"
    key_file: "/etc/certs/tls.key"
```

### JWT认证
```yaml
# 在XDS JWT过滤器配置中设置
jwt_auth:
  providers:
    my-provider:
      issuer: "https://my-issuer.com"
      audiences: ["my-service"]
      jwks_uri: "https://my-issuer.com/.well-known/jwks.json"
```

## 可观测性

### 链路追踪
XDS组件自动集成OpenTelemetry链路追踪：

```go
import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// 追踪会自动添加到所有gRPC调用
ctx := context.Background()
span := trace.SpanFromContext(ctx)  // 获取当前span
```

### 监控指标
支持以下监控指标：
- xds_client_connected - XDS客户端连接状态
- xds_config_updates_received - 配置更新接收计数
- xds_config_updates_applied - 配置更新应用计数
- xds_rpc_duration_seconds - RPC调用耗时

## 高级用法

### 自定义拦截器
```go
func main() {
	// 创建自定义XDS拦截器
	interceptor := xds.NewXDSInterceptor(xdsClient)

	// 应用到客户端
	conn, err := grpc.Dial("service:80",
		grpc.WithUnaryInterceptor(interceptor.UnaryClientInterceptor()),
	)
}
```

### 监听配置变更
```go
client := xds.GetClient().(*xds.xdsClient)
client.SetConfigUpdateCallback(func(resource types.Resource) {
	logger.InfoField("config updated",
		logger.String("type", fmt.Sprintf("%T", resource)))
})
```

### 多环境支持
```yaml
# 开发环境
xds:
  management_server_addresses:
    - "istiod-dev:15010"
  node:
    metadata:
      env: "development"

# 生产环境
xds:
  management_server_addresses:
    - "istiod-prod:15010"
  node:
    metadata:
      env: "production"
```

## 故障排查

### 常见问题

1. **连接XDS服务器失败**
   ```
   ERROR: failed to connect to management server
   ```
   - 检查网络连接
   - 验证TLS证书配置
   - 确认服务器地址正确

2. **配置加载超时**
   ```
   ERROR: timeout waiting for initial resources
   ```
   - 增加`initial_load_timeout`配置
   - 检查XDS服务器是否有对应配置

3. **服务发现失败**
   ```
   WARN: cluster not found for service: my-service
   ```
   - 确认XDS中配置了对应集群
   - 检查集群名称匹配规则

### 调试模式
```yaml
yggdrasil:
  logger:
    level: "debug"
```

启用调试日志可以查看详细的XDS通信过程。

## 示例

参考[example](./example)目录下的完整示例：

- [basic](./example/basic) - 基础XDS客户端使用
- [security](./example/security) - mTLS认证示例

## 版本兼容性

### 当前支持的版本

- **go-control-plane**: v0.12.x (推荐)
- **Istio**: 1.18+
- **Kubernetes**: 1.25+

详细兼容性信息请参考 [COMPATIBILITY.md](./COMPATIBILITY.md)

## 开发状态

### ✅ 已完成
- 完整的框架集成
- XDS客户端基础架构
- 多种负载均衡算法
- TLS/mTLS安全支持
- OpenTelemetry集成

### 🔄 开发中
- 完整的LDS/RDS/CDS/EDS协议实现
- 动态配置更新
- 生产级错误处理

### ⏳ 计划中
- ADS聚合发现服务
- 高级路由规则
- 性能优化

## 贡献

欢迎提交Issue和Pull Request来改进这个组件。

### 贡献指南
1. 查看版本兼容性要求
2. 遵循现有代码架构
3. 添加适当的测试
4. 更新相关文档

## 许可证

Apache License 2.0