# xDS Service Governance for Yggdrasil

This package provides xDS (Envoy's data plane API) protocol integration for the Yggdrasil microservice framework, enabling advanced service mesh capabilities through integration with control planes like Istio Pilot.

## Supported xDS APIs

- **LDS** (Listener Discovery Service): Listener configuration
- **RDS** (Route Discovery Service): Route configuration
- **CDS** (Cluster Discovery Service): Cluster/service configuration
- **EDS** (Endpoint Discovery Service): Endpoint/instance discovery

### Service Discovery Flow

The xDS resolver follows the standard xDS discovery protocol chain:

1. **LDS (Listener Discovery Service)**: Your service connects to the xDS control plane and subscribes to LDS using the service name as the listener name
2. **RDS (Route Discovery Service)**: The listener configuration contains route information (either inline or as an RDS reference). If it's an RDS reference, the resolver subscribes to RDS to get the route configuration
3. **CDS (Cluster Discovery Service)**: The route configuration contains cluster references. The resolver extracts all cluster names and subscribes to CDS
4. **EDS (Endpoint Discovery Service)**: For each cluster, the resolver subscribes to EDS to get the actual endpoint addresses
5. **Configuration Update**: The resolver updates the framework's configuration with the discovered endpoints
6. **Load Balancing**: The balancer uses the updated endpoints for intelligent load balancing

```
Client Service Name
       ↓
   LDS Request (listener name = service name)
       ↓
   Listener Config
       ↓
   RDS Request (if route is referenced) OR Inline Route
       ↓
   Route Config (contains cluster names)
       ↓
   CDS Request (for discovered clusters)
       ↓
   Cluster Configs
       ↓
   EDS Request (for each cluster)
       ↓
   Endpoints (IP:Port)
```

This design allows for:
- **Dynamic routing**: Routes can be updated without restarting services
- **Traffic splitting**: Multiple clusters can be configured with weights
- **A/B testing**: Different versions can be routed based on headers
- **Canary deployments**: Gradual rollout of new versions


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
# Check if xDS control plane (e.g., Istio Pilot) is running
curl http://localhost:15010/ready

# Check xDS client logs
# Look for "xDS client connected" message
```

### No Endpoints Discovered

The resolver follows a chain: LDS → RDS → CDS → EDS. Check each step:

1. **Verify LDS subscription**:
   - Look for log: `xDS resolver subscribing to LDS (listener=<service-name>)`
   - Ensure the listener name matches your service name

2. **Check listener configuration**:
   - Look for log: `xDS resolver received listener`
   - Verify the listener contains route configuration

3. **Verify route discovery**:
   - For RDS: Look for `listener references RDS, subscribing`
   - For inline: Look for `listener has inline route config`
   - Check log: `xDS resolver received route configuration`

4. **Check cluster extraction**:
   - Look for log: `xDS resolver extracted clusters from route`
   - Verify cluster names are correct

5. **Verify CDS subscription**:
   - Look for log: `xDS resolver subscribing to CDS`
   - Check log: `xDS resolver received cluster`

6. **Check EDS subscription**:
   - Look for log: `xDS resolver subscribing to EDS for clusters`
   - Check log: `xDS resolver received endpoints`

7. **Verify final update**:
   - Look for log: `xDS resolver updated endpoints (count=N)`

**Common issues**:
- Service name doesn't match listener name in control plane
- Route configuration doesn't reference any clusters
- Cluster names in route don't exist in control plane
- No healthy endpoints available for the cluster

### TLS Errors

1. Verify certificate paths are correct
2. Check certificate validity
3. Ensure server name matches certificate CN
4. Verify CA certificate is trusted

### Debug Logging

Enable debug logging to see the full discovery chain:

```yaml
yggdrasil:
  logger:
    level: debug  # or info
```

Look for the sequence:
```
INFO  xDS resolver subscribing to LDS (listener=my-service)
INFO  xDS resolver received listener (listener=my-service)
INFO  listener references RDS, subscribing (route=my-route)
INFO  xDS resolver received route configuration (route=my-route)
INFO  xDS resolver extracted clusters from route (cluster_count=2, clusters=[cluster-v1, cluster-v2])
INFO  xDS resolver subscribing to CDS
INFO  xDS resolver received cluster (cluster=cluster-v1)
INFO  xDS resolver received cluster (cluster=cluster-v2)
INFO  xDS resolver subscribing to EDS for clusters (cluster_count=2)
INFO  xDS resolver received endpoints (cluster=cluster-v1, locality_count=1)
INFO  xDS resolver received endpoints (cluster=cluster-v2, locality_count=1)
INFO  xDS resolver updated endpoints (service=my-service, count=10)
```


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
