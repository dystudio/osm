package smi

import (
	target "github.com/deislabs/smi-sdk-go/pkg/apis/access/v1alpha1"
	spec "github.com/deislabs/smi-sdk-go/pkg/apis/specs/v1alpha1"
	split "github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/deislabs/smc/pkg/endpoint"
)

type friendlyName string

// InformerCollection is a struct of the Kubernetes informers used in SMC
type InformerCollection struct {
	Services      cache.SharedIndexInformer
	TrafficSplit  cache.SharedIndexInformer
	TrafficSpec   cache.SharedIndexInformer
	TrafficTarget cache.SharedIndexInformer
}

// CacheCollection is a struct of the Kubernetes caches used in SMC
type CacheCollection struct {
	Services      cache.Store
	TrafficSplit  cache.Store
	TrafficSpec   cache.Store
	TrafficTarget cache.Store
}

// Client is a struct for all components necessary to connect to and maintain state of a Kubernetes cluster.
type Client struct {
	caches        *CacheCollection
	cacheSynced   chan interface{}
	providerIdent string
	informers     *InformerCollection
	announcements chan interface{}
}

// ClientIdentity is the identity of an Envoy proxy connected to the Service Mesh Controller.
type ClientIdentity string

// MeshSpec is an interface declaring functions, which provide the specs for a service mesh declared with SMI.
type MeshSpec interface {
	// ListTrafficSplits lists TrafficSplit SMI resources.
	ListTrafficSplits() []*split.TrafficSplit

	// ListServices fetches all services declared with SMI Spec.
	ListServices() ([]endpoint.ServiceName, map[endpoint.ServiceName][]endpoint.ServiceName)

	// ListServiceAccounts fetches all service accounts declared with SMI Spec.
	ListServiceAccounts() []endpoint.ServiceAccount

	// GetService fetches a specific service declared in SMI.
	GetService(endpoint.ServiceName) (service *corev1.Service, exists bool, err error)

	// ListHTTPTrafficSpecs lists TrafficSpec SMI resources.
	ListHTTPTrafficSpecs() []*spec.HTTPRouteGroup

	// ListTrafficTargets lists TrafficTarget SMI resources.
	ListTrafficTargets() []*target.TrafficTarget

	// GetAnnouncementsChannel returns the channel on which SMI makes annoucements
	GetAnnouncementsChannel() <-chan interface{}
}