// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	rest "k8s.io/client-go/rest"
)

type HwameistorV1alpha1Interface interface {
	RESTClient() rest.Interface
	LocalDisksGetter
	LocalDiskClaimsGetter
	LocalDiskNodesGetter
	LocalDiskVolumesGetter
	LocalStorageNodesGetter
	LocalVolumesGetter
	LocalVolumeConvertsGetter
	LocalVolumeExpandsGetter
	LocalVolumeGroupsGetter
	LocalVolumeMigratesGetter
	LocalVolumeReplicasGetter
}

// HwameistorV1alpha1Client is used to interact with features provided by the hwameistor.io group.
type HwameistorV1alpha1Client struct {
	restClient rest.Interface
}

func (c *HwameistorV1alpha1Client) LocalDisks() LocalDiskInterface {
	return newLocalDisks(c)
}

func (c *HwameistorV1alpha1Client) LocalDiskClaims() LocalDiskClaimInterface {
	return newLocalDiskClaims(c)
}

func (c *HwameistorV1alpha1Client) LocalDiskNodes() LocalDiskNodeInterface {
	return newLocalDiskNodes(c)
}

func (c *HwameistorV1alpha1Client) LocalDiskVolumes() LocalDiskVolumeInterface {
	return newLocalDiskVolumes(c)
}

func (c *HwameistorV1alpha1Client) LocalStorageNodes() LocalStorageNodeInterface {
	return newLocalStorageNodes(c)
}

func (c *HwameistorV1alpha1Client) LocalVolumes() LocalVolumeInterface {
	return newLocalVolumes(c)
}

func (c *HwameistorV1alpha1Client) LocalVolumeConverts() LocalVolumeConvertInterface {
	return newLocalVolumeConverts(c)
}

func (c *HwameistorV1alpha1Client) LocalVolumeExpands() LocalVolumeExpandInterface {
	return newLocalVolumeExpands(c)
}

func (c *HwameistorV1alpha1Client) LocalVolumeGroups() LocalVolumeGroupInterface {
	return newLocalVolumeGroups(c)
}

func (c *HwameistorV1alpha1Client) LocalVolumeMigrates() LocalVolumeMigrateInterface {
	return newLocalVolumeMigrates(c)
}

func (c *HwameistorV1alpha1Client) LocalVolumeReplicas() LocalVolumeReplicaInterface {
	return newLocalVolumeReplicas(c)
}

// NewForConfig creates a new HwameistorV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*HwameistorV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &HwameistorV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new HwameistorV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *HwameistorV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new HwameistorV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *HwameistorV1alpha1Client {
	return &HwameistorV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *HwameistorV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
