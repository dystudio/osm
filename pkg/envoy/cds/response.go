package cds

import (
	"context"

	xds "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes"

	"github.com/deislabs/smc/pkg/catalog"
	"github.com/deislabs/smc/pkg/envoy"
	"github.com/deislabs/smc/pkg/log/level"
	"github.com/deislabs/smc/pkg/smi"
)

const (
	serverName = "CDS"
)

// NewResponse creates a new Cluster Discovery Response.
func NewResponse(ctx context.Context, catalog catalog.MeshCataloger, meshSpec smi.MeshSpec, proxy *envoy.Proxy) (*xds.DiscoveryResponse, error) {
	allServices, err := catalog.ListEndpoints("TBD")
	if err != nil {
		glog.Errorf("[%s] Failed listing endpoints: %+v", serverName, err)
		return nil, err
	}
	glog.Infof("[%s] WeightedServices: %+v", serverName, allServices)
	resp := &xds.DiscoveryResponse{
		TypeUrl: string(envoy.TypeCDS),
	}

	clusterFactories := []*xds.Cluster{}
	for targetedServiceName, weightedServices := range allServices {
		remoteService := envoy.GetServiceCluster(string(targetedServiceName), string(targetedServiceName))
		clusterFactories = append(clusterFactories, remoteService)
		for _, localservice := range weightedServices {
			clusterFactories = append(clusterFactories, getServiceClusterLocal(string(localservice.ServiceName)))
		}
	}

	for _, cluster := range clusterFactories {
		glog.V(level.Trace).Infof("[%s] Constructed ClusterConfiguration: %+v", serverName, cluster)
		marshalledClusters, err := ptypes.MarshalAny(cluster)
		if err != nil {
			glog.Errorf("[%s] Failed to marshal cluster for proxy %s: %v", serverName, proxy.GetCommonName(), err)
			return nil, err
		}
		resp.Resources = append(resp.Resources, marshalledClusters)
	}
	return resp, nil
}