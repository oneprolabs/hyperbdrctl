package i18n

var cloud_sync_gateway_tencentEN = map[string]string{
	"cmd.cloud_sync_gateway.create.tencent.short": "Create a Tencent Cloud cloud sync gateway",
	"help.cloud_sync_gateway.create.tencent": `Create a Tencent Cloud block-storage cloud sync gateway.

Parameter Sources:
  --cloud-account-id string
    Use the Tencent Cloud block-storage account ID from cloud-account list or cloud-account detail.

Resource Discovery:
  Query available zones:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res zones \
      --output json

  Then query gateway flavors for the selected zone (minimum 2 vCPU and 4 GiB RAM):
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res flavors \
      --purpose make_hg \
      --flavor-vcpus 2 \
      --flavor-ram 4 \
      --output json

  Then query networks and subnets in that zone:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res networks,subnets \
      --output json

  Preview the resolved Linux image and system-disk type:
    hyperbdrctl cloud-sync-gateway create \
      --cloud-account-id <account_id> \
      --preview-request \
      --output json

Minimum Create Command:
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --system-disk-type-id <system_disk_type_id>

To override metadata paths, repeat either flag:
  --set path=value
  --set-json path=<json>

Paths are relative to create_storage.metadata. Precedence is:
  --set-json < --set < explicit create flags

To inspect the final request body before create, add:
  --preview-request

After create returns a storage ID, wait for provisioning and inspect the gateway:
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`,
	"cmd.target.cloud_sync_gateway.create.tencent.usage_notes": `Create a Tencent Cloud block-storage cloud sync gateway.

Parameter Sources:
  --cloud-account-id string
    Use the Tencent Cloud block-storage account ID from cloud-account list or cloud-account detail.

Resource Discovery:
  Query zones, flavors, networks, and subnets with cloud-resource fetch. Preview to resolve the Linux image and system-disk type:
    hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> --preview-request --output json

Minimum Create Command:
  hyperbdrctl cloud-sync-gateway create --cloud-type tencent \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --system-disk-type-id <system_disk_type_id>

To inspect the resolved resources and final request body before create, add:
  --preview-request

After create returns a storage ID, wait for provisioning and inspect the gateway:
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`,
}
