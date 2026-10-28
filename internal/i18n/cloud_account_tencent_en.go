package i18n

var cloudAccountTencentEN = map[string]string{
	"help.cloud_resource.fetch.tencent.block": `Fetch Tencent Cloud block-storage resources.

Use direct-credential mode when no cloud account exists. With an existing cloud account, use:
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --fetch-res regions

Parameter Sources:
  --access-key-id string
    Use the Access Key ID from the Tencent Cloud account.

  --access-key-secret string
    Use the Access Key Secret from the Tencent Cloud account.

  --app-id string
    Use the App ID from Tencent Cloud Account Information.

The minimum query is:
  hyperbdrctl cloud-resource fetch --cloud-type tencent --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --app-id <app_id> \
    --fetch-res regions

For queries other than regions, add --region-id when required.

For scripting against raw fields, add:
  --output json`,
	"cmd.cloud_accounts.create.block.tencent.short":                 "Create Tencent Cloud block-storage account",
	"cmd.cloud_accounts.create.block.tencent.long":                  "Create a Tencent Cloud block-storage account.",
	"flag.cloud-account.tencent.app-id":                             "Tencent Cloud App ID",
	"cmd.cloud_accounts.create.block.tencent.examples":              "hyperbdrctl cloud-account create --cloud-type tencent --storage-type block --access-key-id <ak> --access-key-secret <sk> --app-id <app_id> --auth-region-id ap-beijing",
	"cmd.cloud_accounts.create.block.tencent.notes":                 "- Tencent Cloud block-storage account creation.\n- Requires Access Key ID, Access Key Secret, App ID, and authentication region ID.",
	"cmd.cloud_accounts.create.block.tencent.workflow":              "1. Query the Tencent Cloud region list.\n2. Select the authentication region ID.\n3. Create the cloud account and wait for the task to finish.",
	"cmd.cloud_accounts.create.block.tencent.minimum_flags":         "--access-key-id\n--access-key-secret\n--app-id\n--auth-region-id",
	"cmd.cloud_accounts.create.block.tencent.common_optional_flags": "--account-name",
	"cmd.cloud_accounts.create.block.tencent.related":               "hyperbdrctl cloud-resource fetch --cloud-type tencent --storage-type block --access-key-id <ak> --access-key-secret <sk> --app-id <app_id> --fetch-res regions --output json\nhyperbdrctl cloud-account list --storage-type block",
	"cmd.cloud_accounts.create.block.tencent.next_steps":            "After creation, use `hyperbdrctl cloud-account detail --id <account_id>` to inspect the saved account.",
	"help.cloud_account.block.tencent": `Create a Tencent Cloud block-storage cloud account.

Parameter Sources:
  --access-key-id string
    Use the Access Key ID from the Tencent Cloud account.

  --access-key-secret string
    Use the Access Key Secret from the Tencent Cloud account.

  --app-id string
    Use the App ID from Tencent Cloud Account Information.

  --auth-region-id string
    Use the Tencent Cloud authentication region ID, for example ap-beijing.

Resource Retrieval:
  Query the region list first:
    hyperbdrctl cloud-resource fetch --cloud-type tencent --storage-type block \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --app-id <app_id> \
      --fetch-res regions \
      --output json

The minimum create command is:
  hyperbdrctl cloud-account create --cloud-type tencent --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --app-id <app_id> \
    --auth-region-id <auth_region_id>

After creation succeeds, wait for the task:
  hyperbdrctl cloud-account wait --id <account_id>

After the task finishes, view the cloud account details:
  hyperbdrctl cloud-account detail --id <account_id>
  hyperbdrctl cloud-account list --storage-type block`,
}
