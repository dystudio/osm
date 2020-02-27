package sds

import (
	"context"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes"

	"github.com/deislabs/smc/pkg/catalog"
	"github.com/deislabs/smc/pkg/certificate"
	"github.com/deislabs/smc/pkg/envoy"
	"github.com/deislabs/smc/pkg/smi"
)

const (
	serverName = "SDS"
)

// NewResponse creates a new Secrets Discovery Response.
func NewResponse(ctx context.Context, catalog catalog.MeshCataloger, meshSpec smi.MeshSpec, proxy *envoy.Proxy) (*v2.DiscoveryResponse, error) {
	glog.Infof("[%s] Composing SDS Discovery Response for proxy: %s", serverName, proxy.GetCommonName())
	cert, err := catalog.GetCertificateForService(proxy.GetService())
	if err != nil {
		glog.Errorf("[%s] Error obtaining a certificate for client %s: %s", serverName, proxy.GetCommonName(), err)
		return nil, err
	}

	resp := &v2.DiscoveryResponse{
		TypeUrl: string(envoy.TypeSDS),
	}

	allServices, err := catalog.ListEndpoints("TBD")
	if err != nil {
		glog.Errorf("[%s] Failed listing endpoints: %+v", serverName, err)
		return nil, err
	}

	for targetedServiceName, weightedServices := range allServices {
		secret := newSecret(cert, string(targetedServiceName))
		marshalledSecret, err := ptypes.MarshalAny(secret)
		if err != nil {
			glog.Errorf("[%s] Failed to marshal secret for proxy %s: %v", serverName, proxy.GetCommonName(), err)
			return nil, err
		}
		resp.Resources = append(resp.Resources, marshalledSecret)
		for _, localservice := range weightedServices {
			secret := newSecret(cert, string(localservice.ServiceName))
			marshalledSecret, err := ptypes.MarshalAny(secret)
			if err != nil {
				glog.Errorf("[%s] Failed to marshal secret for proxy %s: %v", serverName, proxy.GetCommonName(), err)
				return nil, err
			}
			resp.Resources = append(resp.Resources, marshalledSecret)
		}
	}
	//Add server_cert
	serverSecret := newSecret(cert, envoy.CertificateName)
	marshalledSecret, err := ptypes.MarshalAny(serverSecret)
	if err != nil {
		glog.Errorf("[%s] Failed to marshal secret for proxy %s: %v", serverName, proxy.GetCommonName(), err)
		return nil, err
	}
	resp.Resources = append(resp.Resources, marshalledSecret)
	return resp, nil
}

func newSecret(cert certificate.Certificater, serviceName string) *auth.Secret {
	secret := &auth.Secret{
		// The Name field must match the tls_context.common_tls_context.tls_certificate_sds_secret_configs.name in the Envoy yaml config
		Name: serviceName, // cert.GetName(),
		Type: &auth.Secret_TlsCertificate{
			TlsCertificate: &auth.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{
						InlineBytes: cert.GetCertificateChain(),
					},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{
						InlineBytes: cert.GetPrivateKey(),
					},
				},
			},
		},
	}
	return secret
}