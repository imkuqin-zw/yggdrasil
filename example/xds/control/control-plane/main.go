package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	grpcPort = 15010
)

// Logger 实现 go-control-plane 的日志接口
type Logger struct{}

func (logger Logger) Debugf(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

func (logger Logger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func (logger Logger) Warnf(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

func (logger Logger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// makeCluster 创建集群配置
func makeCluster(clusterName string, endpoints []string) *clusterv3.Cluster {
	var lbEndpoints []*endpointv3.LbEndpoint

	for _, ep := range endpoints {
		host, port := ep, uint32(5678)
		lbEndpoints = append(lbEndpoints, &endpointv3.LbEndpoint{
			HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
				Endpoint: &endpointv3.Endpoint{
					Address: &corev3.Address{
						Address: &corev3.Address_SocketAddress{
							SocketAddress: &corev3.SocketAddress{
								Protocol: corev3.SocketAddress_TCP,
								Address:  host,
								PortSpecifier: &corev3.SocketAddress_PortValue{
									PortValue: port,
								},
							},
						},
					},
				},
			},
		})
	}

	return &clusterv3.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: clusterv3.Cluster_STRICT_DNS},
		LbPolicy:             clusterv3.Cluster_ROUND_ROBIN,
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []*endpointv3.LocalityLbEndpoints{{
				LbEndpoints: lbEndpoints,
			}},
		},
		DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
	}
}

// makeRoute 创建路由配置
func makeRoute(routeName string, clusterName string) *routev3.RouteConfiguration {
	return &routev3.RouteConfiguration{
		Name: routeName,
		VirtualHosts: []*routev3.VirtualHost{{
			Name:    "backend",
			Domains: []string{"*"},
			Routes: []*routev3.Route{{
				Match: &routev3.RouteMatch{
					PathSpecifier: &routev3.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &routev3.Route_Route{
					Route: &routev3.RouteAction{
						ClusterSpecifier: &routev3.RouteAction_Cluster{
							Cluster: clusterName,
						},
						Timeout: durationpb.New(0 * time.Second),
					},
				},
			}},
		}},
	}
}

// makeHTTPListener 创建 HTTP 监听器
func makeHTTPListener(listenerName string, routeName string, port uint32) *listenerv3.Listener {
	routerConfig, _ := anypb.New(&routerv3.Router{})

	manager := &hcmv3.HttpConnectionManager{
		CodecType:  hcmv3.HttpConnectionManager_AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcmv3.HttpConnectionManager_Rds{
			Rds: &hcmv3.Rds{
				ConfigSource: &corev3.ConfigSource{
					ResourceApiVersion: corev3.ApiVersion_V3,
					ConfigSourceSpecifier: &corev3.ConfigSource_Ads{
						Ads: &corev3.AggregatedConfigSource{},
					},
				},
				RouteConfigName: routeName,
			},
		},
		HttpFilters: []*hcmv3.HttpFilter{{
			Name: wellknown.Router,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{
				TypedConfig: routerConfig,
			},
		}},
	}

	pbst, _ := anypb.New(manager)

	return &listenerv3.Listener{
		Name: listenerName,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: corev3.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*listenerv3.FilterChain{{
			Filters: []*listenerv3.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listenerv3.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
}

// generateSnapshot 生成 xDS 配置快照
func generateSnapshot() *cachev3.Snapshot {
	// 创建集群
	backendCluster := makeCluster("backend_cluster", []string{"backend-service", "backend-service-2"})

	// 创建路由
	backendRoute := makeRoute("local_route", "backend_cluster")

	// 创建监听器
	httpListener := makeHTTPListener("listener_0", "local_route", 10000)

	snapshot, err := cachev3.NewSnapshot(
		"1",
		map[resourcev3.Type][]types.Resource{
			resourcev3.ClusterType:  {backendCluster},
			resourcev3.RouteType:    {backendRoute},
			resourcev3.ListenerType: {httpListener},
		},
	)
	if err != nil {
		log.Fatalf("Failed to create snapshot: %v", err)
	}

	return snapshot
}

// runServer 启动 xDS 服务器
func runServer(ctx context.Context, snapshotCache cachev3.SnapshotCache, port uint) {
	grpcServer := grpc.NewServer()

	cb := &serverv3.CallbackFuncs{
		StreamOpenFunc: func(ctx context.Context, id int64, typ string) error {
			log.Printf("Stream opened: ID=%d Type=%s", id, typ)
			return nil
		},
		StreamClosedFunc: func(id int64, node *corev3.Node) {
			log.Printf("Stream closed: ID=%d Node=%s", id, node.GetId())
		},
		StreamRequestFunc: func(id int64, req *discoverygrpc.DiscoveryRequest) error {
			log.Printf("Stream request: ID=%d Node=%s Type=%s Version=%s",
				id, req.GetNode().GetId(), req.GetTypeUrl(), req.GetVersionInfo())
			return nil
		},
		StreamResponseFunc: func(ctx context.Context, id int64, req *discoverygrpc.DiscoveryRequest, resp *discoverygrpc.DiscoveryResponse) {
			log.Printf("Stream response: ID=%d Type=%s Version=%s Resources=%d",
				id, resp.GetTypeUrl(), resp.GetVersionInfo(), len(resp.GetResources()))
		},
	}

	server := serverv3.NewServer(ctx, snapshotCache, cb)

	// 注册各个 Discovery Service
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, server)
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("xDS management server listening on :%d", port)
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	flag.Parse()

	ctx := context.Background()

	// 创建快照缓存，使用通配符匹配所有节点
	logger := Logger{}
	snapshotCache := cachev3.NewSnapshotCache(false, cachev3.IDHash{}, logger)

	// 生成初始配置快照
	snapshot := generateSnapshot()

	// 为所有可能的节点 ID 设置快照
	nodeIDs := []string{
		"one.default.svc.cluster.local",
		"envoy-proxy",
		"sidecar~10.0.0.1~envoy.default~default.svc.cluster.local",
	}

	for _, nodeID := range nodeIDs {
		if err := snapshotCache.SetSnapshot(ctx, nodeID, snapshot); err != nil {
			log.Printf("Failed to set snapshot for node %s: %v", nodeID, err)
		} else {
			log.Printf("Snapshot set for node: %s", nodeID)
		}
	}

	log.Printf("Snapshot created with version: 1")
	log.Printf("Clusters: 1")
	log.Printf("Routes: 1")
	log.Printf("Listeners: 1")

	// 启动 xDS 服务器
	runServer(ctx, snapshotCache, grpcPort)
}
