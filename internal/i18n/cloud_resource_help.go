package i18n

const (
	cloudResourceFetchAKBlockZH = `获取%[1]s块存储只读资源。

直连凭证模式用于没有云账号时查询云资源；已有云账号时，优先执行：
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --help

最小查询命令如下：
  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --fetch-res regions

常见资源组包括：
  regions
  zones
  flavors
  images
  networks,subnets
  system_disk_types

查看云账号创建帮助：
  hyperbdrctl cloud-account create --cloud-type %[2]s --storage-type block --help

查看云同步网关帮助：
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> --help

查看启动配置帮助：
  hyperbdrctl boot-config apply --help

如需脚本化处理原始字段，可附加：
  --output json`

	cloudResourceFetchAKObjectZH = `获取%[1]s对象存储只读资源。

直连凭证模式用于没有云账号时查询云资源；已有云账号时，优先执行：
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --help

最小查询命令如下：
  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --fetch-res regions

常见资源组包括：
  regions
  zones
  flavors
  images
  networks,subnets
  system_volume_types
  volume_types

查看云账号创建帮助：
  hyperbdrctl cloud-account create --cloud-type %[2]s --storage-type object --help

查看启动配置帮助：
  hyperbdrctl boot-config apply --help

如需脚本化处理原始字段，可附加：
  --output json`

	cloudResourceFetchOpenstackBlockZH = `获取 OpenStack 块存储只读资源。

直连凭证模式用于没有云账号时查询云资源；已有云账号时，优先执行：
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --help

参数来源：
  --auth-url string
    登录平台后，点击右上角用户名，再点击【OpenStack RC 文件】，读取 OS_AUTH_URL。

  --username string
    通常使用平台右上角显示的用户名。

  --password string
    填写该用户名对应的登录密码。

  --user-domain-id string
    在 OpenStack 控制节点执行 ` + "`openstack user show <用户名>`" + `，读取 ` + "`domain_id`" + `；默认通常为 ` + "`default`" + `。

最小查询命令如下：
  hyperbdrctl cloud-resource fetch --cloud-type openstack --storage-type block \
    --auth-url <auth_url> \
    --username <username> \
    --password <password> \
    --user-domain-id <user_domain_id>

常见资源组包括：
  regions,compute_zones,projects
  flavors,networks,images
  volume_types,subnets,security_groups

查看云账号创建帮助：
  hyperbdrctl cloud-account create --cloud-type openstack --storage-type block --help

查看云同步网关帮助：
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> --help

查看启动配置帮助：
  hyperbdrctl boot-config apply --help

如需脚本化处理原始字段，可附加：
  --output json`

	cloudResourceFetchOpenstackObjectZH = `获取 OpenStack 对象存储只读资源。

直连凭证模式用于没有云账号时查询云资源；已有云账号时，优先执行：
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --help

参数来源：
  --auth-url string
    登录平台后，点击右上角用户名，再点击【OpenStack RC 文件】，读取 OS_AUTH_URL。

  --username string
    通常使用平台右上角显示的用户名。

  --password string
    填写该用户名对应的登录密码。

  --user-domain-id string
    在 OpenStack 控制节点执行 ` + "`openstack user show <用户名>`" + `，读取 ` + "`domain_id`" + `；默认通常为 ` + "`default`" + `。

最小查询命令如下：
  hyperbdrctl cloud-resource fetch --cloud-type openstack --storage-type object \
    --auth-url <auth_url> \
    --username <username> \
    --password <password> \
    --user-domain-id <user_domain_id>

常见资源组包括：
  regions,compute_zones,projects
  flavors,networks,images
  volume_types,subnets,security_groups

查看云账号创建帮助：
  hyperbdrctl cloud-account create --cloud-type openstack --storage-type object --help

查看启动配置帮助：
  hyperbdrctl boot-config apply --help

如需脚本化处理原始字段，可附加：
  --output json`

	cloudResourceFetchAKBlockEN = `Fetch read-only %[1]s block-storage resources.

Use direct-credential mode to query cloud resources when no cloud account exists. With an existing cloud account, prefer:
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --help

The minimum query is:
  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --fetch-res regions

Common resource groups include:
  regions
  zones
  flavors
  images
  networks,subnets
  system_disk_types

View cloud account creation help:
  hyperbdrctl cloud-account create --cloud-type %[2]s --storage-type block --help

View cloud sync gateway help:
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> --help

View boot configuration help:
  hyperbdrctl boot-config apply --help

For scripting against raw fields, add:
  --output json`

	cloudResourceFetchAKObjectEN = `Fetch read-only %[1]s object-storage resources.

Use direct-credential mode to query cloud resources when no cloud account exists. With an existing cloud account, prefer:
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --help

The minimum query is:
  hyperbdrctl cloud-resource fetch --cloud-type %[2]s --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --fetch-res regions

Common resource groups include:
  regions
  zones
  flavors
  images
  networks,subnets
  system_volume_types
  volume_types

View cloud account creation help:
  hyperbdrctl cloud-account create --cloud-type %[2]s --storage-type object --help

View boot configuration help:
  hyperbdrctl boot-config apply --help

For scripting against raw fields, add:
  --output json`

	cloudResourceFetchOpenstackBlockEN = `Fetch read-only OpenStack block-storage resources.

Use direct-credential mode to query cloud resources when no cloud account exists. With an existing cloud account, prefer:
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --help

Parameter Sources:
  --auth-url string
    In the platform UI, click the username in the upper-right corner, click OpenStack RC File, and read OS_AUTH_URL.

  --username string
    Usually use the username shown in the upper-right corner of the platform UI.

  --password string
    Enter the login password for that username.

  --user-domain-id string
    On the OpenStack control node, run ` + "`openstack user show <username>`" + ` and read ` + "`domain_id`" + `; the default is usually ` + "`default`" + `.

The minimum query is:
  hyperbdrctl cloud-resource fetch --cloud-type openstack --storage-type block \
    --auth-url <auth_url> \
    --username <username> \
    --password <password> \
    --user-domain-id <user_domain_id>

Common resource groups include:
  regions,compute_zones,projects
  flavors,networks,images
  volume_types,subnets,security_groups

View cloud account creation help:
  hyperbdrctl cloud-account create --cloud-type openstack --storage-type block --help

View cloud sync gateway help:
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> --help

View boot configuration help:
  hyperbdrctl boot-config apply --help

For scripting against raw fields, add:
  --output json`

	cloudResourceFetchOpenstackObjectEN = `Fetch read-only OpenStack object-storage resources.

Use direct-credential mode to query cloud resources when no cloud account exists. With an existing cloud account, prefer:
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --help

Parameter Sources:
  --auth-url string
    In the platform UI, click the username in the upper-right corner, click OpenStack RC File, and read OS_AUTH_URL.

  --username string
    Usually use the username shown in the upper-right corner of the platform UI.

  --password string
    Enter the login password for that username.

  --user-domain-id string
    On the OpenStack control node, run ` + "`openstack user show <username>`" + ` and read ` + "`domain_id`" + `; the default is usually ` + "`default`" + `.

The minimum query is:
  hyperbdrctl cloud-resource fetch --cloud-type openstack --storage-type object \
    --auth-url <auth_url> \
    --username <username> \
    --password <password> \
    --user-domain-id <user_domain_id>

Common resource groups include:
  regions,compute_zones,projects
  flavors,networks,images
  volume_types,subnets,security_groups

View cloud account creation help:
  hyperbdrctl cloud-account create --cloud-type openstack --storage-type object --help

View boot configuration help:
  hyperbdrctl boot-config apply --help

For scripting against raw fields, add:
  --output json`
)
