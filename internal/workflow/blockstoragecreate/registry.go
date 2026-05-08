package blockstoragecreate

import "hyperbdr-client/internal/workflow"

type Fetcher interface {
	FetchGatewayCloudInfo(accountID string, resources []string, regionID, zoneID, flavorID, flavorVCPUs, flavorRAM, purpose string) (map[string]interface{}, error)
	FetchGatewayTransitionImages(accountID, cloudType, regionID, zoneID, purpose, imageType, osType, bootMode string) (map[string]interface{}, error)
	FetchOpenStackCloudInfo(accountID string) (map[string]interface{}, error)
}

type adapter struct {
	key   string
	build func(Fetcher, Spec) (string, map[string]interface{}, error)
}

var adapters = map[string]adapter{
	workflow.BlockStorageCreateKey("aliyun_bs").String(): {
		key:   workflow.BlockStorageCreateKey("aliyun_bs").String(),
		build: buildAliyun,
	},
	workflow.BlockStorageCreateKey("openstack").String(): {
		key:   workflow.BlockStorageCreateKey("openstack").String(),
		build: buildOpenStack,
	},
}

func HasRegisteredAdapter(cloudType string) bool {
	flowKey := workflow.BlockStorageCreateKey(cloudType)
	_, ok := adapters[flowKey.String()]
	return ok
}

func BuildRequest(fetcher Fetcher, spec Spec) (string, map[string]interface{}, error) {
	flowKey := workflow.BlockStorageCreateKey(spec.CloudType)
	if adapter, ok := adapters[flowKey.String()]; ok {
		return adapter.build(fetcher, spec)
	}
	return buildGeneric(spec)
}
