package bootconfigapply

import (
	"strings"

	"hyperbdr-client/internal/workflow"
)

type applyDriver interface {
	Name() string
	Enrich(Service, *applyDriverSpec) error
}

type applyDriverFunc struct {
	name string
	run  func(Service, *applyDriverSpec) error
}

func (d applyDriverFunc) Name() string {
	return d.name
}

func (d applyDriverFunc) Enrich(s Service, spec *applyDriverSpec) error {
	if d.run == nil {
		return nil
	}
	return d.run(s, spec)
}

var (
	genericApplyDriver = applyDriverFunc{name: "generic"}
	applyDrivers       = map[string]applyDriver{
		workflow.BootConfigApplyKey("aliyun_obs", "objectstorage").String():              aliyunObjectApplyDriver,
		workflow.BootConfigApplyKey("aliyun_bs", workflow.HyperGateStorageType).String(): aliyunBlockApplyDriver,
	}
)

func resolveApplyDriver(metadata map[string]interface{}, storageInfo storageDefaults) applyDriver {
	key := workflow.BootConfigApplyKey(
		strings.ToLower(mapStringValue(metadata["cloud_type"])),
		effectiveStorageType(metadata, storageInfo),
	)
	if driver, ok := applyDrivers[key.String()]; ok {
		return driver
	}
	return genericApplyDriver
}

func effectiveStorageType(metadata map[string]interface{}, storageInfo storageDefaults) string {
	if storageInfo.IsObjectStorage {
		return "objectstorage"
	}
	if storageInfo.IsBlockStorage {
		return workflow.HyperGateStorageType
	}
	switch normalizeStorageType(mapStringValue(metadata["storage_type"])) {
	case "objectstorage":
		return "objectstorage"
	case "hypergate", "blockstorage", "block":
		return workflow.HyperGateStorageType
	default:
		return mapStringValue(metadata["storage_type"])
	}
}
