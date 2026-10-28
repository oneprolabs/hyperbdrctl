package i18n

var cloud_sync_gateway_tencentZH = map[string]string{
	"cmd.cloud_sync_gateway.create.tencent.short": "创建腾讯云云同步网关",
	"help.cloud_sync_gateway.create.tencent": `创建腾讯云块存储云同步网关。

参数来源：
  --cloud-account-id string
    使用 cloud-account list 或 cloud-account detail 返回的腾讯云块存储云账号 ID。

资源获取：
  先查询可用区：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res zones \
      --output json

  再查询该可用区的云同步网关规格（最低要求为 2 vCPU、4 GiB 内存）：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res flavors \
      --purpose make_hg \
      --flavor-vcpus 2 \
      --flavor-ram 4 \
      --output json

  再查询该可用区的网络和子网：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res networks,subnets \
      --output json

  预览 Linux 镜像和系统盘类型解析结果：
    hyperbdrctl cloud-sync-gateway create \
      --cloud-account-id <account_id> \
      --preview-request \
      --output json

最小创建命令如下：
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --system-disk-type-id <system_disk_type_id>

覆盖 metadata 路径时，可重复传入：
  --set path=value
  --set-json path=<json>

路径相对于 create_storage.metadata。覆盖顺序为：
  --set-json < --set < 显式创建参数

创建前预览最终请求体：
  --preview-request

创建返回存储 ID 后，等待资源创建并查看网关详情：
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`,
	"cmd.target.cloud_sync_gateway.create.tencent.usage_notes": `创建腾讯云块存储云同步网关。

参数来源：
  --cloud-account-id string
    使用 cloud-account list 或 cloud-account detail 返回的腾讯云块存储云账号 ID。

资源获取：
  使用 cloud-resource fetch 查询可用区、规格、网络和子网。创建前预览 Linux 镜像和系统盘类型解析结果：
    hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> --preview-request --output json

最小创建命令如下：
  hyperbdrctl cloud-sync-gateway create --cloud-type tencent \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --system-disk-type-id <system_disk_type_id>

创建前附加 --preview-request，检查解析后的资源和最终请求体。

创建返回存储 ID 后，等待资源创建并查看网关详情：
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`,
}
