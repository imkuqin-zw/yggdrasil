# xDS Example

This directory contains example applications demonstrating xDS integration with the Yggdrasil framework.

## Prerequisites

1. **Istio Pilot** running on `localhost:15010`
   ```bash
   docker run -d --name istio-pilot -p 15010:15010 -p 15011:15011 istio/pilot:1.19.0
   ```

2. **Go dependencies** installed
   ```bash
   cd ../../..
   go mod tidy
   ```

## Running the Examples

### Server Example

The server example demonstrates:
- xDS client initialization
- Service registration (if supported)
- Health checking
- Graceful shutdown

```bash
cd server
go run main.go
```

Expected output:
```
INFO  Starting xDS example server
INFO  xDS Configuration server=localhost:15010 cluster=example-cluster nodeId=example-service-instance-1 tls=false
INFO  xDS client connected server=localhost:15010 node=example-service-instance-1
INFO  Health server starting port=8080
INFO  Server health check status=healthy
```

### Client Example

The client example demonstrates:
- Service discovery via xDS EDS
- Load balancing with xDS CDS
- Endpoint selection
- Health-aware routing

```bash
cd client
go run main.go
```

Expected output:
```
INFO  Starting xDS example client
INFO  === Service Discovery Demo ===
INFO  Watching service via xDS service=example-service
INFO  xDS resolver updated endpoints service=example-service count=2
INFO  Discovered endpoints service=example-service endpoints=[...]
INFO  === Load Balancing Demo ===
INFO  Created xDS balancer service=example-service balancer=xds
INFO  Selected endpoint attempt=1 address=127.0.0.1:8080 protocol=grpc
INFO  Selected endpoint attempt=2 address=127.0.0.1:8081 protocol=grpc
...
```

## Configuration

Both examples use YAML configuration files:

- `server/config.yaml` - Server configuration
- `client/config.yaml` - Client configuration

Key configuration sections:

```yaml
yggdrasil:
  xds:
    server:
      address: "localhost:15010"  # Istio Pilot address
    node:
      cluster: "example-cluster"  # Cluster name
      id: "instance-id"           # Unique instance ID
    resources:
      cds: true                   # Cluster Discovery
      eds: true                   # Endpoint Discovery
```

## Testing with Istio

### 1. Configure Istio with Test Services

Create a test service configuration in Istio:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: example-service
spec:
  ports:
  - port: 8080
    name: grpc
---
apiVersion: v1
kind: Endpoints
metadata:
  name: example-service
subsets:
- addresses:
  - ip: 127.0.0.1
    port: 8080
  - ip: 127.0.0.1
    port: 8081
```

### 2. Verify xDS Responses

You can use `grpcurl` to inspect xDS responses:

```bash
# List available services
grpcurl -plaintext localhost:15010 list

# Get cluster configuration
grpcurl -plaintext -d '{"node":{"id":"test","cluster":"example-cluster"},"resource_names":["example-service"]}' \
  localhost:15010 envoy.service.discovery.v3.AggregatedDiscoveryService/StreamAggregatedResources
```

## Troubleshooting

### No Endpoints Discovered

1. Check Istio Pilot is running:
   ```bash
   curl http://localhost:15010/ready
   ```

2. Verify service is registered in Istio

3. Check xDS client logs for connection errors

### Connection Refused

1. Ensure Istio Pilot is listening on port 15010
2. Check firewall settings
3. Verify address in config.yaml

### TLS Errors

If using TLS (port 15011):
1. Verify certificate paths in config
2. Check certificate validity
3. Ensure CA certificate is correct

## Advanced Usage

### Using TLS

Update config.yaml:

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
```

### Custom Locality

Configure locality for geographic routing:

```yaml
yggdrasil:
  xds:
    node:
      locality:
        region: "us-west"
        zone: "us-west-1a"
        subZone: "rack-1"
```

### Load Balancing Policies

The balancer automatically uses the policy configured in CDS:
- Round Robin (default)
- Random
- Least Request

## Next Steps

1. Integrate with your actual gRPC services
2. Configure Istio for production use
3. Add monitoring and observability
4. Implement circuit breaking and retries
5. Set up multi-cluster routing

## References

- [Istio Documentation](https://istio.io/latest/docs/)
- [xDS Protocol](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol)
- [Yggdrasil Framework](../../README.md)
