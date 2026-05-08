package cloudaccountcreate

import (
	"fmt"

	"hyperbdr-client/internal/workflow"
)

type adapter struct {
	key   string
	build func(Spec) (string, map[string]interface{}, error)
}

var adapters = map[string]adapter{
	workflow.CloudAccountCreateKey("aliyun_bs", "").String(): {
		key:   workflow.CloudAccountCreateKey("aliyun_bs", "").String(),
		build: buildAliyunBlock,
	},
	workflow.CloudAccountCreateKey("openstack", "").String(): {
		key:   workflow.CloudAccountCreateKey("openstack", "").String(),
		build: buildOpenStackBlock,
	},
	workflow.CloudAccountCreateKey("aliyun_obs", "objectstorage").String(): {
		key:   workflow.CloudAccountCreateKey("aliyun_obs", "objectstorage").String(),
		build: buildAliyunObject,
	},
	workflow.CloudAccountCreateKey("openstack", "objectstorage").String(): {
		key:   workflow.CloudAccountCreateKey("openstack", "objectstorage").String(),
		build: buildOpenStackObject,
	},
}

func HasRegisteredAdapter(cloudType, storageType string) bool {
	flowKey := workflow.CloudAccountCreateKey(cloudType, storageType)
	_, ok := adapters[flowKey.String()]
	return ok
}

func BuildRequest(spec Spec) (string, map[string]interface{}, error) {
	flowKey := workflow.CloudAccountCreateKey(spec.CloudType, spec.StorageType)
	adapter, ok := adapters[flowKey.String()]
	if ok {
		return adapter.build(spec)
	}
	switch flowKey.StorageType {
	case workflow.BlockStorageType:
		return buildGenericBlock(spec)
	case "objectstorage":
		return buildGenericObject(spec)
	default:
		return "", nil, fmt.Errorf("unsupported target account create flow for %s", flowKey.String())
	}
}
