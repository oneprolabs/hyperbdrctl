package i18n

var cloudAccountTencentJA = map[string]string{
	"help.cloud_resource.fetch.tencent.block": `Tencent Cloud のブロックストレージ読み取り専用リソースを取得します。

クラウドアカウントがない場合は直接認証情報モードを使用します。既存のクラウドアカウントがある場合は、次を使用します。
  hyperbdrctl cloud-resource fetch --cloud-account-id <account_id> --fetch-res regions

パラメータソース:
  --access-key-id string
    Tencent Cloud アカウントの Access Key ID を使用します。

  --access-key-secret string
    Tencent Cloud アカウントの Access Key Secret を使用します。

  --app-id string
    Tencent Cloud のアカウント情報に表示される App ID を使用します。

最小限のクエリ:
  hyperbdrctl cloud-resource fetch --cloud-type tencent --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --app-id <app_id> \
    --fetch-res regions

リージョン以外のリソースをクエリする場合は、必要に応じて --region-id を追加します。

生のフィールドをスクリプトで処理するには、次を追加します。
  --output json`,
	"cmd.cloud_accounts.create.block.tencent.short":                 "Tencent Cloud ブロックストレージアカウントを作成",
	"cmd.cloud_accounts.create.block.tencent.long":                  "Tencent Cloud のブロックストレージアカウントを作成します。",
	"flag.cloud-account.tencent.app-id":                             "Tencent Cloud の App ID",
	"cmd.cloud_accounts.create.block.tencent.examples":              "hyperbdrctl cloud-account create --cloud-type tencent --storage-type block --access-key-id <ak> --access-key-secret <sk> --app-id <app_id> --auth-region-id ap-beijing",
	"cmd.cloud_accounts.create.block.tencent.notes":                 "- Tencent Cloud ブロックストレージアカウントを作成します。\n- Access Key ID、Access Key Secret、App ID、認証リージョン ID が必要です。",
	"cmd.cloud_accounts.create.block.tencent.workflow":              "1. Tencent Cloud のリージョン一覧をクエリします。\n2. 認証リージョン ID を確認します。\n3. クラウドアカウントを作成し、タスクの完了を待ちます。",
	"cmd.cloud_accounts.create.block.tencent.minimum_flags":         "--access-key-id\n--access-key-secret\n--app-id\n--auth-region-id",
	"cmd.cloud_accounts.create.block.tencent.common_optional_flags": "--account-name",
	"cmd.cloud_accounts.create.block.tencent.related":               "hyperbdrctl cloud-resource fetch --cloud-type tencent --storage-type block --access-key-id <ak> --access-key-secret <sk> --app-id <app_id> --fetch-res regions --output json\nhyperbdrctl cloud-account list --storage-type block",
	"cmd.cloud_accounts.create.block.tencent.next_steps":            "作成後、`hyperbdrctl cloud-account detail --id <account_id>` で保存されたアカウントを確認します。",
	"help.cloud_account.block.tencent": `Tencent Cloud のブロックストレージアカウントを作成します。

パラメータソース:
  --access-key-id string
    Tencent Cloud アカウントのアクセス キー ID を使用します。

  --access-key-secret string
    Tencent Cloud アカウントのアクセス キー シークレットを使用します。

  --app-id string
    Tencent Cloud のアカウント情報に表示される App ID を使用します。

  --auth-region-id string
    Tencent Cloud の認証リージョン ID を指定します。例: ap-beijing。

リソースの取得:
  最初にリージョン一覧をクエリします。
    hyperbdrctl cloud-resource fetch --cloud-type tencent --storage-type block \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --app-id <app_id> \
      --fetch-res regions \
      --output json

最小限の作成コマンドは次のとおりです。
  hyperbdrctl cloud-account create --cloud-type tencent --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --app-id <app_id> \
    --auth-region-id <auth_region_id>

作成が成功したら、タスクを待ちます。
  hyperbdrctl cloud-account wait --id <account_id>

タスクが完了したら、クラウド アカウントの詳細を表示します。
  hyperbdrctl cloud-account detail --id <account_id>
  hyperbdrctl cloud-account list --storage-type block`,
}
