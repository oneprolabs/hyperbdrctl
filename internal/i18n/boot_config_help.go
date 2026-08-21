package i18n

const (
	bootConfigApplyAliyunBlockZH = `基于云账号 ` + "`<account_id>`" + ` 创建或更新启动配置；对应云厂商为 ` + "`aliyun`" + `，存储类型为 ` + "`block`" + `。

参数来源：
  --storage-id string
    块存储场景使用云同步网关 ID，通常来自：
      hyperbdrctl cloud-sync-gateway list

建议按下面顺序准备：
  1. 确认基础资源：
    hyperbdrctl host list --status host_register_done
    hyperbdrctl cloud-account detail --id <account_id>
    hyperbdrctl cloud-sync-gateway list

  2. 查询云资源：
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

    再查询卷类型：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --flavor-id <flavor_id> \
        --fetch-res system_volume_types,volume_types

    再查询网络和子网：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --fetch-res networks,subnets

    再查询安全组：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --network-id <network_id> \
        --fetch-res security_groups

最小应用命令如下：
  hyperbdrctl boot-config apply \
    --id <host_id> \
    --cloud-account-id <account_id> \
    --storage-id <storage_id> \
    --zone-id <zone_id> \
    --flavor-id <flavor_id> \
    --volume-type-id <volume_type_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --security-group-id <security_group_id>

如需继续覆盖 metadata，可附加下面的参数：
  --set key=value
  --set-json key=<json>

覆盖顺序固定为：
  --set-json < --set < 显式参数和动态参数

如需只检查最终请求体而不发送写请求，可附加下面的参数：
  --preview-request`
)

const bootConfigApplyAliyunObjectZH = `基于云账号 ` + "`<account_id>`" + ` 创建或更新启动配置；对应云厂商为 ` + "`aliyun`" + `，存储类型为 ` + "`object`" + `。

参数来源：
  --storage-id string
    对象存储场景使用目标对象存储 ID，通常来自：
      hyperbdrctl oss list

  --system-volume-type-id string
    使用系统盘卷类型 ID。

  --volume-type-id string
    使用数据盘卷类型 ID。

建议按下面顺序准备：
  1. 确认基础资源：
    hyperbdrctl host list --status host_register_done
    hyperbdrctl cloud-account detail --id <account_id>
    hyperbdrctl oss list

  2. 查询云资源：
    先查询可用区：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --fetch-res zones

    再查询规格和操作系统类型：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --fetch-res flavors,os_types \
        --flavor-vcpus 2 \
        --flavor-ram 4

    再查询卷类型：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --flavor-id <flavor_id> \
        --fetch-res system_volume_types,volume_types

    再查询网络：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --fetch-res networks

    最后查询子网和安全组：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --network-id <network_id> \
        --fetch-res subnets,security_groups

最小应用命令如下：
  hyperbdrctl boot-config apply \
    --id <host_id> \
    --cloud-account-id <account_id> \
    --storage-id <storage_id> \
    --zone-id <zone_id> \
    --flavor-id <flavor_id> \
    --volume-type-id <volume_type_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --security-group-id <security_group_id> \
    --bandwidth-size 100

如需继续覆盖 metadata，可附加下面的参数：
  --set key=value
  --set-json key=<json>

覆盖顺序固定为：
  --set-json < --set < 显式参数和动态参数

如需只检查最终请求体而不发送写请求，可附加下面的参数：
  --preview-request`

const bootConfigApplyHuaweiObjectZH = `基于云账号 ` + "`<account_id>`" + ` 创建或更新启动配置；对应云厂商为 ` + "`huawei`" + `，存储类型为 ` + "`object`" + `。

参数来源：
  --storage-id string
    对象存储场景使用目标对象存储 ID，通常来自：
      hyperbdrctl oss list

  --system-volume-type-id string
    使用系统盘卷类型 ID。

  --volume-type-id string
    使用数据盘卷类型 ID。

建议按下面顺序准备：
  1. 确认基础资源：
    hyperbdrctl host list --status host_register_done
    hyperbdrctl cloud-account detail --id <account_id>
    hyperbdrctl oss list

  2. 查询云资源：
    先查询地域和可用区：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --fetch-res regions,zones

    再查询规格和操作系统类型：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --fetch-res flavors,os_types \
        --flavor-vcpus 2 \
        --flavor-ram 4

    再查询卷类型：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --zone-id <zone_id> \
        --fetch-res system_volume_types,volume_types

    再查询网络：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --fetch-res networks

    最后查询子网和安全组：
      hyperbdrctl cloud-resource fetch \
        --cloud-account-id <account_id> \
        --fetch-res subnets,security_groups

最小应用命令如下：
  hyperbdrctl boot-config apply \
    --id <host_id> \
    --cloud-account-id <account_id> \
    --storage-id <storage_id> \
    --zone-id <zone_id> \
    --flavor-id <flavor_id> \
    --volume-type-id <volume_type_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --security-group-id <security_group_id> \
    --bandwidth-size 100

如需继续覆盖 metadata，可附加下面的参数：
  --set key=value
  --set-json key=<json>

覆盖顺序固定为：
  --set-json < --set < 显式参数和动态参数

如需只检查最终请求体而不发送写请求，可附加下面的参数：
  --preview-request`
