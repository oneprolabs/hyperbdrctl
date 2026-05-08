package workflow

import "testing"

func TestBlockStorageCreateKey(t *testing.T) {
	key := BlockStorageCreateKey("aliyun_bs")
	if key.Workflow != BlockStorageCreateFlow || key.CloudType != "aliyun_bs" || key.StorageType != HyperGateStorageType {
		t.Fatalf("key = %+v", key)
	}
	if key.String() != "block_storage_create|aliyun_bs|HyperGate" {
		t.Fatalf("key.String() = %q", key.String())
	}
}

func TestCloudAccountCreateKeyNormalizesBlockStorageType(t *testing.T) {
	tests := []struct {
		name        string
		storageType string
		want        string
	}{
		{name: "empty", storageType: "", want: BlockStorageType},
		{name: "hypergate", storageType: HyperGateStorageType, want: BlockStorageType},
		{name: "objectstorage", storageType: "objectstorage", want: "objectstorage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := CloudAccountCreateKey("openstack", tt.storageType)
			if key.Workflow != CloudAccountCreateFlow || key.CloudType != "openstack" || key.StorageType != tt.want {
				t.Fatalf("key = %+v", key)
			}
		})
	}
}
