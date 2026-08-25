package i18n

// Cloud-account creation keeps provider-specific preparation and minimum-command guidance here.
// Context-dependent supplements are added by the command help renderer.
const (
	cloudAccountCreateHelpBacktick  = "\x60"
	cloudAccountCreateAliyunBlockZH = `创建阿里云块存储云账号。

参数来源：
  --access-key-id string
    使用阿里云账号的 Access Key ID。

  --access-key-secret string
    使用阿里云账号的 Access Key Secret。

  --auth-region-id string
    使用阿里云认证地域 ID。

资源获取：
  先查询可用认证地域：
    hyperbdrctl cloud-resource fetch --cloud-type aliyun --storage-type block \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --fetch-res regions \
      --output json

最小创建命令如下：
  hyperbdrctl cloud-account create --cloud-type aliyun --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --auth-region-id <auth_region_id>

创建成功后，等待任务完成：
  hyperbdrctl cloud-account wait --id <account_id>

任务结束后，查看云账号详情：
  hyperbdrctl cloud-account detail --id <account_id>
  hyperbdrctl cloud-account list --storage-type block`
	cloudAccountCreateHuaweiBlockZH = `创建华为云块存储云账号。

参数来源：
  --access-key-id string
    使用华为云账号的 Access Key ID。

  --access-key-secret string
    使用华为云账号的 Access Key Secret。

  --auth-region-id string
    创建云账号前需要先确认认证地域 ID。

资源获取：
  先查询区域列表：
    hyperbdrctl cloud-resource fetch --cloud-type huawei --storage-type block \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --fetch-res regions \
      --output json

最小创建命令如下：
  hyperbdrctl cloud-account create --cloud-type huawei --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --auth-region-id <auth_region_id>

如需保存云账号显示名称，可按需附加：
  --account-name <name>

创建成功后，等待任务完成：
  hyperbdrctl cloud-account wait --id <account_id>

任务结束后，查看云账号详情：
  hyperbdrctl cloud-account detail --id <account_id>`
	cloudAccountCreateOpenStackBlockZH = `创建 OpenStack 块存储云账号。

参数来源：
  --auth-url string
    登录平台后，点击右上角用户名，再点击【OpenStack RC 文件】，读取 OS_AUTH_URL。

  --username string
    通常使用平台右上角显示的用户名。

  --password string
    填写该用户名对应的登录密码。

  --user-domain-id string
    在 OpenStack 控制节点执行 ` + cloudAccountCreateHelpBacktick + `openstack user show <用户名>` + cloudAccountCreateHelpBacktick + `，读取 ` + cloudAccountCreateHelpBacktick + `domain_id` + cloudAccountCreateHelpBacktick + `；默认通常为 ` + cloudAccountCreateHelpBacktick + `default` + cloudAccountCreateHelpBacktick + `。

  --project-domain-id string
    在【OpenStack RC 文件】中读取 OS_PROJECT_DOMAIN_ID；默认通常为 ` + cloudAccountCreateHelpBacktick + `default` + cloudAccountCreateHelpBacktick + `。

  --project-name string
    通常与用户名相同。

  --region-name string
    在【OpenStack RC 文件】中读取 OS_REGION_NAME；默认通常为 ` + cloudAccountCreateHelpBacktick + `RegionOne` + cloudAccountCreateHelpBacktick + `。

资源获取：
  OpenStack 块存储云账号创建本身不依赖前置资源查询。
  准备好上面的鉴权与项目参数后，可直接创建云账号。

最小创建命令如下：
  hyperbdrctl cloud-account create --cloud-type openstack --storage-type block \
    --auth-url <auth_url> \
    --username <username> \
    --password <password> \
    --user-domain-id <user_domain_id> \
    --project-domain-id <project_domain_id> \
    --project-name <project_name> \
    --region-name <region_name>

创建成功后，等待任务完成：
  hyperbdrctl cloud-account wait --id <account_id>

任务结束后，查看云账号详情：
  hyperbdrctl cloud-account detail --id <account_id>`
	cloudAccountCreateAliyunObjectZH = `创建阿里云对象存储云账号。

参数来源：
  --access-key-id string
    使用阿里云账号的 Access Key ID。

  --access-key-secret string
    使用阿里云账号的 Access Key Secret。

  --region-id string
    使用目标区域 ID，通常来自区域列表查询结果。

资源获取：
  查询区域列表：
    hyperbdrctl cloud-resource fetch --cloud-type aliyun --storage-type object \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --fetch-res regions \
      --output json

最小创建命令如下：
  hyperbdrctl cloud-account create --cloud-type aliyun --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id>

创建成功后，等待任务完成：
  hyperbdrctl cloud-account wait --id <account_id>

任务结束后，查看云账号详情：
  hyperbdrctl cloud-account detail --id <account_id>`
	cloudAccountCreateHuaweiObjectZH = `创建华为云对象存储云账号。

参数来源：
  --access-key-id string
    使用华为云账号的 Access Key ID。

  --access-key-secret string
    使用华为云账号的 Access Key Secret。

  --region-id string
    创建云账号前需要先确认目标区域 ID。

资源获取：
  如需确认可用的 region-id，可先查询华为云地域列表：
    hyperbdrctl cloud-resource fetch --cloud-type huawei --storage-type object \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --fetch-res regions \
      --output json

最小创建命令如下：
  hyperbdrctl cloud-account create --cloud-type huawei --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id>

创建成功后，等待任务完成：
  hyperbdrctl cloud-account wait --id <account_id>

任务结束后，查看云账号详情：
  hyperbdrctl cloud-account detail --id <account_id>`
	cloudAccountCreateOpenStackObjectZH = `创建 OpenStack 对象存储云账号。

参数来源：
  --auth-url string
    登录平台后，点击右上角用户名，再点击【OpenStack RC 文件】，读取 OS_AUTH_URL。

  --username string
    通常使用平台右上角显示的用户名。

  --password string
    填写该用户名对应的登录密码。

  --user-domain-id string
    在 OpenStack 控制节点执行 ` + cloudAccountCreateHelpBacktick + `openstack user show <用户名>` + cloudAccountCreateHelpBacktick + `，读取 ` + cloudAccountCreateHelpBacktick + `domain_id` + cloudAccountCreateHelpBacktick + `；默认通常为 ` + cloudAccountCreateHelpBacktick + `default` + cloudAccountCreateHelpBacktick + `。

  --project-domain-id string
  --project-id string
  --project-name string
  --region-id string
    这些项目和地域字段通常来自资源获取结果；只有需要覆盖自动带入值时，才显式传入。

  --use-internal-ip string
    0 表示公网访问，1 表示内网访问。

资源获取：
  先查询云资源：
    hyperbdrctl cloud-resource fetch --cloud-type openstack --storage-type object \
      --auth-url <auth_url> \
      --username <username> \
      --password <password> \
      --user-domain-id <user_domain_id> \
      --output json

  重点关注返回结果中的：
    project_domain_id
    project_id
    project_name
    region_id
    region_name
    boot_loader_image_id
    boot_loader_image_name
    disk_bus_type_id
    disk_bus_type_name

最小创建命令如下：
  hyperbdrctl cloud-account create --cloud-type openstack --storage-type object \
    --auth-url <auth_url> \
    --username <username> \
    --password <password> \
    --user-domain-id <user_domain_id>

创建成功后，等待任务完成：
  hyperbdrctl cloud-account wait --id <account_id>

任务结束后，查看云账号详情：
  hyperbdrctl cloud-account detail --id <account_id>`
)
