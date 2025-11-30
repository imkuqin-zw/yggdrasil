# xDS Service Governance for Yggdrasil

This package provides xDS (Envoy's data plane API) protocol integration for the Yggdrasil microservice framework, enabling advanced service mesh capabilities through integration with control planes like Istio Pilot.

## Features

- **Service Discovery**: Automatic service endpoint discovery via EDS (Endpoint Discovery Service)
- **Load Balancing**: Intelligent load balancing with CDS (Cluster Discovery Service) supporting:
  - Round Robin
  - Random
  - Least Request
- **Health Checking**: Endpoint health status tracking and filtering
- **Locality Awareness**: Support for region/zone/subzone based routing
- **Dynamic Configuration**: Real-time updates from xDS control plane
- **TLS Support**: Secure connections to xDS server with mutual TLS

## Supported xDS APIs

- **LDS** (Listener Discovery Service): Listener configuration
- **RDS** (Route Discovery Service): Route configuration
- **CDS** (Cluster Discovery Service): Cluster/service configuration
- **EDS** (Endpoint Discovery Service): Endpoint/instance discovery

## Configuration

### Basic Configuration (gRPC)

```yaml
yggdrasil:
  xds:
    server:
      address: "localhost:15010"
      useTLS: false
    node:
      cluster: "my-cluster"
      id: "my-service-instance-1"
    resources:
      lds: true
      rds: true
      cds: true
      eds: true
```

### TLS Configuration

```yaml
yggdrasil:
  xds:
    server:
      address: "localhost"
      useTLS: true
      tlsPort: 15011
    tls:
      enabled: true
      caCert: "/path/to/ca.crt"
      clientCert: "/path/to/client.crt"
      clientKey: "/path/to/client.key"
      serverName: "istiod.istio-system.svc"
    node:
      cluster: "my-cluster"
      id: "my-service-instance-1"
```

### Advanced Configuration

```yaml
yggdrasil:
  xds:
    server:
      address: "localhost:15010"
    node:
      cluster: "production-cluster"
      id: "service-v1-abc123"
      locality:
        region: "us-west"
        zone: "us-west-1a"
        subZone: "rack-1"
      metadata:
        version: "1.0.0"
        environment: "production"
    resources:
      lds: true
      rds: true
      cds: true
      eds: true
      clusterNames:
        - "my-service"
        - "other-service"
    timeout: 10s
    retryInterval: 5s
    maxRetries: 3
    useADS: true
```

## Usage

### Using xDS Resolver

```go
import (
    "github.com/imkuqin-zw/yggdrasil"
    _ "github.com/imkuqin-zw/yggdrasil/contrib/xds"
)

func main() {
    // The xDS resolver is automatically registered
    // Configure your client to use xDS for service discovery
    yggdrasil.Run("my-service",
        yggdrasil.WithResolver("xds"),
        // ... other options
    )
}
```

### Using xDS Balancer

```go
import (
    "github.com/imkuqin-zw/yggdrasil"
    _ "github.com/imkuqin-zw/yggdrasil/contrib/xds"
)

func main() {
    yggdrasil.Run("my-service",
        yggdrasil.WithBalancer("xds"),
        // ... other options
    )
}
```

### Using xDS Registry

```go
import (
    "github.com/imkuqin-zw/yggdrasil"
    _ "github.com/imkuqin-zw/yggdrasil/contrib/xds"
)

func main() {
    // Note: xDS registry is primarily for discovery
    // Registration is typically handled by the control plane
    yggdrasil.Run("my-service",
        yggdrasil.WithRegistry("xds"),
        // ... other options
    )
}
```

## Integration with Istio

This package is designed to work with Istio Pilot 1.19.0 and later versions.

### Prerequisites

1. **Install Istio** in your Kubernetes cluster or run Istio Pilot standalone
2. **Configure Pilot** to expose xDS endpoints:
   - gRPC: port 15010 (default)
   - gRPC with TLS: port 15011

### Docker Setup

```bash
# Run Istio Pilot in Docker
docker run -d \
  --name istio-pilot \
  -p 15010:15010 \
  -p 15011:15011 \
  istio/pilot:1.19.0
```

### Service Discovery Flow

1. Your service starts and connects to Istio Pilot via xDS
2. The xDS client subscribes to EDS for service endpoints
3. Istio Pilot pushes endpoint updates to your service
4. The resolver updates the framework's configuration
5. The balancer uses the updated endpoints for load balancing

## Load Balancing Policies

The balancer supports the following policies configured via CDS:

- **Round Robin**: Distributes requests evenly across endpoints
- **Random**: Randomly selects an endpoint for each request
- **Least Request**: Routes to the endpoint with the fewest active requests

## Health Checking

Endpoints are automatically filtered based on health status:

- `HEALTHY`: Endpoint is healthy and receives traffic
- `UNHEALTHY`: Endpoint is unhealthy and excluded from load balancing
- `DRAINING`: Endpoint is being drained, no new connections
- `TIMEOUT`: Health check timed out
- `DEGRADED`: Endpoint is degraded but may receive traffic
- `UNKNOWN`: Health status unknown, treated as healthy

## Locality-Aware Load Balancing

The implementation supports locality-aware routing:

- Endpoints are grouped by region/zone/subzone
- Priority-based routing (lower priority preferred)
- Weighted load balancing within localities

## Error Handling

The client includes robust error handling:

- Automatic reconnection on connection loss
- Configurable retry intervals and max retries
- Graceful degradation when xDS is unavailable
- Detailed error logging

## Troubleshooting

### Connection Issues

```bash
# Check if Istio Pilot is running
curl http://localhost:15010/ready

# Check xDS client logs
# Look for "xDS client connected" message
```

### No Endpoints Discovered

1. Verify service is registered in Istio
2. Check cluster name matches service name
3. Ensure EDS subscription is active
4. Review Istio Pilot logs for errors

### TLS Errors

1. Verify certificate paths are correct
2. Check certificate validity
3. Ensure server name matches certificate CN
4. Verify CA certificate is trusted

## Performance Considerations

- **Connection Pooling**: The xDS client maintains a single connection to the control plane
- **Update Efficiency**: Only subscribed resources are watched
- **Memory Usage**: Endpoints are cached and updated incrementally
- **CPU Usage**: Minimal overhead for ADS stream processing

## Limitations

- **Service Registration**: xDS is primarily for discovery; registration is typically handled by the control plane (e.g., Kubernetes)
- **LDS/RDS**: Full listener and route configuration support is available but may require additional implementation for specific use cases
- **Control Plane Compatibility**: Tested with Istio Pilot 1.19.0; other xDS servers may require adjustments

## References

- [xDS Protocol](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol)
- [Istio Architecture](https://istio.io/latest/docs/ops/deployment/architecture/)
- [Envoy API](https://www.envoyproxy.io/docs/envoy/latest/api/api)

## License

Copyright 2022 The imkuqin-zw Authors. Licensed under the Apache License, Version 2.0.
