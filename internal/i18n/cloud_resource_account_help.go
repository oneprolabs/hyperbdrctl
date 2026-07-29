package i18n

const (
	cloudResourceFetchAccountAKBlockZH = `获取%[1]s块存储云账号的资源候选项。

建议按下面的顺序查询云同步网关和启动配置需要的资源。

查询可用区：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res zones \
    --output json

查询规格：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --fetch-res flavors \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

查询镜像和系统盘类型：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --flavor-id <flavor_id> \
    --fetch-res images,system_disk_types \
    --output json

查询网络和子网：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --fetch-res networks,subnets \
    --output json

如需查询 Windows 修复镜像，执行：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res win_hd_images \
    --output json

资源确认后，查看后续命令参数：
  hyperbdrctl cloud-sync-gateway create \
    --cloud-account-id <account_id> \
    --help
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help

表格输出时，` + "`--flavor-vcpus`" + ` 和 ` + "`--flavor-ram`" + ` 只过滤本地展示；JSON 输出保留原始 API 字段。`

	cloudResourceFetchAccountAliyunObjectZH = `获取%[1]s对象存储云账号的资源候选项。

建议按下面的顺序查询启动配置需要的资源。

查询可用区：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res zones \
    --output json

查询规格和操作系统类型：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --fetch-res flavors,os_types \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

查询卷类型：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --flavor-id <flavor_id> \
    --fetch-res system_volume_types,volume_types \
    --output json

查询网络：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res networks \
    --output json

查询子网和安全组：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --network-id <network_id> \
    --fetch-res subnets,security_groups \
    --output json

资源确认后，查看启动配置参数：
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help

表格输出时，` + "`--flavor-vcpus`" + ` 和 ` + "`--flavor-ram`" + ` 只过滤本地展示；JSON 输出保留原始 API 字段。`

	cloudResourceFetchAccountHuaweiObjectZH = `获取%[1]s对象存储云账号的资源候选项。

建议按下面的顺序查询启动配置需要的资源。

查询区域和可用区：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res regions,zones \
    --output json

查询规格和操作系统类型：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --fetch-res flavors,os_types \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

查询卷类型：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --fetch-res system_volume_types,volume_types \
    --output json

查询网络：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res networks \
    --output json

华为云子网不按可用区划分。查询子网和安全组：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --network-id <network_id> \
    --fetch-res subnets,security_groups \
    --output json

资源确认后，查看启动配置参数：
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help

表格输出时，` + "`--flavor-vcpus`" + ` 和 ` + "`--flavor-ram`" + ` 只过滤本地展示；JSON 输出保留原始 API 字段。`

	cloudResourceFetchAccountOpenStackBlockZH = `获取 OpenStack 块存储云账号的资源候选项。

建议按下面的顺序查询云同步网关和启动配置需要的资源。

查询区域、计算可用区和项目：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res regions,compute_zones,projects \
    --output json

查询规格和操作系统类型：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --fetch-res flavors,os_types \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

查询镜像：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --os-type linux \
    --fetch-res images \
    --output json

查询卷类型：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --flavor-id <flavor_id> \
    --fetch-res volume_types \
    --output json

查询网络：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --fetch-res networks \
    --output json

查询子网和安全组：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --network-id <network_id> \
    --fetch-res subnets,security_groups \
    --output json

资源确认后，查看后续命令参数：
  hyperbdrctl cloud-sync-gateway create \
    --cloud-account-id <account_id> \
    --help
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help`

	cloudResourceFetchAccountOpenStackObjectZH = `获取 OpenStack 对象存储云账号的资源候选项。

建议按下面的顺序查询启动配置需要的资源。

查询区域、计算可用区和项目：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res regions,compute_zones,projects \
    --output json

查询规格和操作系统类型：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --fetch-res flavors,os_types \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

查询卷类型：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --flavor-id <flavor_id> \
    --fetch-res system_volume_types,volume_types \
    --output json

查询网络：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --fetch-res networks \
    --output json

查询子网和安全组：
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --network-id <network_id> \
    --fetch-res subnets,security_groups \
    --output json

资源确认后，查看启动配置参数：
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help`

	cloudResourceFetchAccountGenericBlockZH = `获取块存储云账号的资源候选项。

先查询可用区、规格、镜像、磁盘类型、网络和子网；只有请求的资源需要时，才附加区域、可用区、规格或网络上下文。

资源确认后，查看后续命令参数：
  hyperbdrctl cloud-sync-gateway create \
    --cloud-account-id <account_id> \
    --help
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help`

	cloudResourceFetchAccountGenericObjectZH = `获取对象存储云账号的资源候选项。

先查询区域、可用区、规格、卷类型、网络、子网和安全组；只有请求的资源需要时，才附加对应上下文。

资源确认后，查看启动配置参数：
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help`

	cloudResourceFetchAccountAKBlockEN = `Fetch resource candidates for the %[1]s block-storage cloud account.

Query the resources needed by cloud sync gateway and boot configuration in this order.

Query availability zones:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res zones \
    --output json

Query flavors:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --fetch-res flavors \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

Query images and system disk types:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --flavor-id <flavor_id> \
    --fetch-res images,system_disk_types \
    --output json

Query networks and subnets:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --fetch-res networks,subnets \
    --output json

To query Windows repair images when needed, run:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res win_hd_images \
    --output json

After confirming the resources, view the downstream command flags:
  hyperbdrctl cloud-sync-gateway create \
    --cloud-account-id <account_id> \
    --help
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help

For table output, ` + "`--flavor-vcpus`" + ` and ` + "`--flavor-ram`" + ` only filter displayed rows. JSON output keeps raw API fields.`

	cloudResourceFetchAccountAliyunObjectEN = `Fetch resource candidates for the %[1]s object-storage cloud account.

Query the resources needed by boot configuration in this order.

Query availability zones:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res zones \
    --output json

Query flavors and OS types:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --fetch-res flavors,os_types \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

Query volume types:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --flavor-id <flavor_id> \
    --fetch-res system_volume_types,volume_types \
    --output json

Query networks:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res networks \
    --output json

Query subnets and security groups:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --network-id <network_id> \
    --fetch-res subnets,security_groups \
    --output json

After confirming the resources, view the boot configuration flags:
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help

For table output, ` + "`--flavor-vcpus`" + ` and ` + "`--flavor-ram`" + ` only filter displayed rows. JSON output keeps raw API fields.`

	cloudResourceFetchAccountHuaweiObjectEN = `Fetch resource candidates for the %[1]s object-storage cloud account.

Query the resources needed by boot configuration in this order.

Query regions and availability zones:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res regions,zones \
    --output json

Query flavors and OS types:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --fetch-res flavors,os_types \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

Query volume types:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --fetch-res system_volume_types,volume_types \
    --output json

Query networks:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res networks \
    --output json

Huawei Cloud subnets are not scoped by availability zone. Query subnets and security groups:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --network-id <network_id> \
    --fetch-res subnets,security_groups \
    --output json

After confirming the resources, view the boot configuration flags:
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help

For table output, ` + "`--flavor-vcpus`" + ` and ` + "`--flavor-ram`" + ` only filter displayed rows. JSON output keeps raw API fields.`

	cloudResourceFetchAccountOpenStackBlockEN = `Fetch resource candidates for the OpenStack block-storage cloud account.

Query the resources needed by cloud sync gateway and boot configuration in this order.

Query regions, compute zones, and projects:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res regions,compute_zones,projects \
    --output json

Query flavors and OS types:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --fetch-res flavors,os_types \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

Query images:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --os-type linux \
    --fetch-res images \
    --output json

Query volume types:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --flavor-id <flavor_id> \
    --fetch-res volume_types \
    --output json

Query networks:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --fetch-res networks \
    --output json

Query subnets and security groups:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --network-id <network_id> \
    --fetch-res subnets,security_groups \
    --output json

After confirming the resources, view the downstream command flags:
  hyperbdrctl cloud-sync-gateway create \
    --cloud-account-id <account_id> \
    --help
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help`

	cloudResourceFetchAccountOpenStackObjectEN = `Fetch resource candidates for the OpenStack object-storage cloud account.

Query the resources needed by boot configuration in this order.

Query regions, compute zones, and projects:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --fetch-res regions,compute_zones,projects \
    --output json

Query flavors and OS types:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --fetch-res flavors,os_types \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

Query volume types:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --flavor-id <flavor_id> \
    --fetch-res system_volume_types,volume_types \
    --output json

Query networks:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --fetch-res networks \
    --output json

Query subnets and security groups:
  hyperbdrctl cloud-resource fetch \
    --cloud-account-id <account_id> \
    --region-id <region_id> \
    --project-name <project_name> \
    --project-domain-id <project_domain_id> \
    --compute-zone-id <compute_zone_id> \
    --network-id <network_id> \
    --fetch-res subnets,security_groups \
    --output json

After confirming the resources, view the boot configuration flags:
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help`

	cloudResourceFetchAccountGenericBlockEN = `Fetch resource candidates for a block-storage cloud account.

Query availability zones, flavors, images, disk types, networks, and subnets first. Add region, zone, flavor, or network context only when the requested resource needs it.

After confirming the resources, view the downstream command flags:
  hyperbdrctl cloud-sync-gateway create \
    --cloud-account-id <account_id> \
    --help
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help`

	cloudResourceFetchAccountGenericObjectEN = `Fetch resource candidates for an object-storage cloud account.

Query regions, availability zones, flavors, volume types, networks, subnets, and security groups first. Add context only when the requested resource needs it.

After confirming the resources, view the boot configuration flags:
  hyperbdrctl boot-config apply \
    --cloud-account-id <account_id> \
    --help`

	cloudResourceFetchHuaweiObjectZH = `获取%[1]s对象存储只读资源。

直连凭证模式用于没有云账号时查询云资源；已有云账号时，优先执行：
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --help

参数来源：
  --access-key-id string
    使用华为云账号的 Access Key ID。

  --access-key-secret string
    使用华为云账号的 Access Key Secret。

先查询区域：
  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --fetch-res regions \
    --output json

再查询可用区和规格：
  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id> \
    --fetch-res zones \
    --output json

  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id> \
    --zone-id <zone_id> \
    --fetch-res flavors \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

按需查询网络、镜像和系统卷类型：
  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id> \
    --fetch-res networks,subnets \
    --output json

  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id> \
    --fetch-res images \
    --output json

  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id> \
    --zone-id <zone_id> \
    --fetch-res system_volume_types \
    --output json

查看云账号创建参数：
  hyperbdrctl cloud-account create --cloud-type %[2]s --storage-type object --help

如需脚本化处理原始字段，可附加：
  --output json`

	cloudResourceFetchHuaweiObjectEN = `Fetch read-only %[1]s object-storage resources.

Use direct-credential mode when no cloud account exists. With an existing cloud account, prefer:
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --help

Parameter Sources:
  --access-key-id string
    Use the Access Key ID from the Huawei Cloud account.

  --access-key-secret string
    Use the Access Key Secret from the Huawei Cloud account.

Query regions first:
  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --fetch-res regions \
    --output json

Then query availability zones and flavors:
  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id> \
    --fetch-res zones \
    --output json

  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id> \
    --zone-id <zone_id> \
    --fetch-res flavors \
    --flavor-vcpus 2 \
    --flavor-ram 4 \
    --output json

Query networks, images, and system volume types as needed:
  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id> \
    --fetch-res networks,subnets \
    --output json

  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id> \
    --fetch-res images \
    --output json

  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id> \
    --zone-id <zone_id> \
    --fetch-res system_volume_types \
    --output json

View cloud account creation flags:
  hyperbdrctl cloud-account create --cloud-type %[2]s --storage-type object --help

For scripting against raw fields, add:
  --output json`
)
