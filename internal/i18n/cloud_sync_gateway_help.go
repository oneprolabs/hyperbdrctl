package i18n

const (
	cloudSyncGatewayCreateAliyunZH = `创建阿里云云同步网关。

参数来源：
  --cloud-account-id string
    来自下面命令的返回结果：
      hyperbdrctl cloud-account create --cloud-type aliyun --storage-type block

资源获取：
  先查询可用区：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res zones

  再查询规格：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res flavors \
      --purpose make_hg \
      --flavor-vcpus 2 \
      --flavor-ram 4

  再查询镜像和系统盘类型：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --flavor-id <flavor_id> \
      --purpose make_hg \
      --image_type=system \
      --fetch-res images,system_volume_types

  再查询网络和子网：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res networks,subnets

最小创建命令如下：
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --system-disk-type-id <system_disk_type_id>

创建返回云同步网关 ID 后，等待任务完成：
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>

任务结束后，查看云同步网关详情：
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`
	cloudSyncGatewayCreateHuaweiZH = `创建华为云云同步网关。

参数来源：
  --cloud-account-id string
    来自下面命令的返回结果：
      hyperbdrctl cloud-account create --cloud-type huawei --storage-type block

    云账号中保存的区域会自动用于创建请求，无需传入 --region-id。

资源获取：
  先查询可用区：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res zones

  再查询规格：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res flavors \
      --flavor-vcpus 2 \
      --flavor-ram 4

  再查询镜像和系统盘类型：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --flavor-id <flavor_id> \
      --fetch-res images,system_disk_types

  再查询网络和子网：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res networks,subnets

最小创建命令如下：
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --system-disk-type-id <system_disk_type_id>

省略 --system-disk-size 时，系统盘大小默认使用 40 GiB。

创建返回云同步网关 ID 后，等待任务完成：
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>

任务结束后，查看云同步网关详情：
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`
	cloudSyncGatewayCreateOpenstackZH = `创建 OpenStack 云同步网关。

参数来源：
  --cloud-account-id string
    来自下面命令的返回结果：
      hyperbdrctl cloud-account create --cloud-type openstack --storage-type block

资源获取：
  先查询项目、区域和计算可用区：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res regions,compute_zones,projects

  再查询规格、网络和镜像：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res flavors,networks,images \
      --flavor-vcpus 2 \
      --flavor-ram 4

  再查询卷类型、子网和安全组：
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res volume_types,subnets,security_groups \
      --flavor-id <flavor_id> \
      --network-id <network_id>

最小创建命令如下：
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id>

按流程补充关键字段后，可执行：
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> \
    --project-id <project_id> \
    --region-id <region_id> \
    --compute-zone-id <compute_zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --volume-type-id <volume_type_id> \
    --block-store-zone-id <block_store_zone_id>

创建返回云同步网关 ID 后，等待任务完成：
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>

任务结束后，查看云同步网关详情：
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`
	cloudSyncGatewayCreateAliyunEN = `Create an Alibaba Cloud cloud sync gateway.

Parameter Sources:
  --cloud-account-id string
    Comes from the result of:
      hyperbdrctl cloud-account create --cloud-type aliyun --storage-type block

Resource Discovery:
  Query zones first:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res zones

  Then query flavors:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res flavors \
      --purpose make_hg \
      --flavor-vcpus 2 \
      --flavor-ram 4

  Then query images and system disk types:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --flavor-id <flavor_id> \
      --purpose make_hg \
      --image_type=system \
      --fetch-res images,system_volume_types

  Then query networks and subnets:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res networks,subnets

The minimum create command is:
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --system-disk-type-id <system_disk_type_id>

After create returns a cloud sync gateway ID, wait for the task to complete:
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>

After the task finishes, view the cloud sync gateway details:
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`
	cloudSyncGatewayCreateHuaweiEN = `Create a Huawei Cloud cloud sync gateway.

Parameter Sources:
  --cloud-account-id string
    Comes from the result of:
      hyperbdrctl cloud-account create --cloud-type huawei --storage-type block

    The saved cloud-account region is used automatically; do not pass --region-id.

Resource Discovery:
  Query zones first:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res zones

  Then query flavors:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res flavors \
      --flavor-vcpus 2 \
      --flavor-ram 4

  Then query images and system disk types:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --flavor-id <flavor_id> \
      --fetch-res images,system_disk_types

  Then query networks and subnets:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res networks,subnets

The minimum create command is:
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --system-disk-type-id <system_disk_type_id>

The system disk size defaults to 40 GiB when omitted.

After create returns a cloud sync gateway ID, wait for the task to complete:
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>

After the task finishes, view the cloud sync gateway details:
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`
	cloudSyncGatewayCreateOpenstackEN = `Create an OpenStack cloud sync gateway.

Parameter Sources:
  --cloud-account-id string
    Comes from the result of:
      hyperbdrctl cloud-account create --cloud-type openstack --storage-type block

Resource Discovery:
  Query projects, regions, and compute zones first:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res regions,compute_zones,projects

  Then query flavors, networks, and images:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res flavors,networks,images \
      --flavor-vcpus 2 \
      --flavor-ram 4

  Then query volume types, subnets, and security groups:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res volume_types,subnets,security_groups \
      --flavor-id <flavor_id> \
      --network-id <network_id>

The minimum create command is:
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id>

After adding the key fields from the resource flow, run:
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> \
    --project-id <project_id> \
    --region-id <region_id> \
    --compute-zone-id <compute_zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --volume-type-id <volume_type_id> \
    --block-store-zone-id <block_store_zone_id>

After create returns a cloud sync gateway ID, wait for the task to complete:
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>

After the task finishes, view the cloud sync gateway details:
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`
)
