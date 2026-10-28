package i18n

var cloudAccountTencentZH = map[string]string{
	"help.cloud_resource.fetch.tencent.block": `获取腾讯云块存储只读资源。

没有云账号时使用直接凭据模式；已有云账号时使用：
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --fetch-res regions

参数来源：
  --access-key-id string
    使用腾讯云账号的 Access Key ID。

  --access-key-secret string
    使用腾讯云账号的 Access Key Secret。

  --app-id string
    使用腾讯云“账号信息”中的 App ID。

最小查询命令如下：
  hyperbdrctl cloud-resource fetch --cloud-type tencent --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --app-id <app_id> \
    --fetch-res regions

查询区域以外的资源时，按需附加 --region-id。

如需脚本化处理原始字段，可附加：
  --output json`,
	"cmd.cloud_accounts.create.block.tencent.short":                 "创建腾讯云块存储云账号",
	"cmd.cloud_accounts.create.block.tencent.long":                  "创建腾讯云块存储云账号。",
	"flag.cloud-account.tencent.app-id":                             "腾讯云 App ID",
	"cmd.cloud_accounts.create.block.tencent.examples":              "hyperbdrctl cloud-account create --cloud-type tencent --storage-type block --access-key-id <ak> --access-key-secret <sk> --app-id <app_id> --auth-region-id ap-beijing",
	"cmd.cloud_accounts.create.block.tencent.notes":                 "- 创建腾讯云块存储云账号。\n- 必须提供 Access Key ID、Access Key Secret、App ID 和认证地域 ID。",
	"cmd.cloud_accounts.create.block.tencent.workflow":              "1. 查询腾讯云地域列表。\n2. 确认认证地域 ID。\n3. 创建云账号并等待任务完成。",
	"cmd.cloud_accounts.create.block.tencent.minimum_flags":         "--access-key-id\n--access-key-secret\n--app-id\n--auth-region-id",
	"cmd.cloud_accounts.create.block.tencent.common_optional_flags": "--account-name",
	"cmd.cloud_accounts.create.block.tencent.related":               "hyperbdrctl cloud-resource fetch --cloud-type tencent --storage-type block --access-key-id <ak> --access-key-secret <sk> --app-id <app_id> --fetch-res regions --output json\nhyperbdrctl cloud-account list --storage-type block",
	"cmd.cloud_accounts.create.block.tencent.next_steps":            "创建完成后，可使用 `hyperbdrctl cloud-account detail --id <account_id>` 查看保存的账号信息。",
	"help.cloud_account.block.tencent": `创建腾讯云块存储云账号。

参数来源：
  --access-key-id string
    使用腾讯云账号的 Access Key ID。

  --access-key-secret string
    使用腾讯云账号的 Access Key Secret。

  --app-id string
    使用腾讯云“账号信息”中的 App ID。

  --auth-region-id string
    使用腾讯云认证地域 ID，例如 ap-beijing（北京）。

资源获取：
  先查询区域列表：
    hyperbdrctl cloud-resource fetch --cloud-type tencent --storage-type block \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --app-id <app_id> \
      --fetch-res regions \
      --output json

最小创建命令如下：
  hyperbdrctl cloud-account create --cloud-type tencent --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --app-id <app_id> \
    --auth-region-id <auth_region_id>

创建成功后，等待任务完成：
  hyperbdrctl cloud-account wait --id <account_id>

任务结束后，查看云账号详情：
  hyperbdrctl cloud-account detail --id <account_id>
  hyperbdrctl cloud-account list --storage-type block`,
}
