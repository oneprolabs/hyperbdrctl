package workflow

const (
	BlockStorageCreateFlow = "block_storage_create"
	CloudAccountCreateFlow = "cloud_account_create"
	BootConfigApplyFlow    = "boot_config_apply"

	HyperGateStorageType = "HyperGate"
	BlockStorageType     = "block"
)

type Key struct {
	Workflow    string
	CloudType   string
	StorageType string
}

func (k Key) String() string {
	return k.Workflow + "|" + k.CloudType + "|" + k.StorageType
}

func BlockStorageCreateKey(cloudType string) Key {
	return Key{
		Workflow:    BlockStorageCreateFlow,
		CloudType:   cloudType,
		StorageType: HyperGateStorageType,
	}
}

func CloudAccountCreateKey(cloudType, storageType string) Key {
	return Key{
		Workflow:    CloudAccountCreateFlow,
		CloudType:   cloudType,
		StorageType: normalizeCloudAccountStorageType(storageType),
	}
}

func BootConfigApplyKey(cloudType, storageType string) Key {
	return Key{
		Workflow:    BootConfigApplyFlow,
		CloudType:   cloudType,
		StorageType: normalizeBootConfigApplyStorageType(storageType),
	}
}

func normalizeCloudAccountStorageType(storageType string) string {
	switch storageType {
	case "", HyperGateStorageType:
		return BlockStorageType
	default:
		return storageType
	}
}

func normalizeBootConfigApplyStorageType(storageType string) string {
	switch storageType {
	case "", HyperGateStorageType, BlockStorageType:
		return HyperGateStorageType
	default:
		return storageType
	}
}
