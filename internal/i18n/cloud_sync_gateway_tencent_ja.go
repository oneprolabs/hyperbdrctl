package i18n

var cloud_sync_gateway_tencentJA = map[string]string{
	"cmd.cloud_sync_gateway.create.tencent.short": "Tencent Cloud クラウド同期ゲートウェイを作成する",
	"cmd.target.cloud_sync_gateway.create.tencent.usage_notes": `Tencent Cloud ブロックストレージのクラウド同期ゲートウェイを作成します。

パラメータの出所:
  --cloud-account-id string
    cloud-account list または cloud-account detail の Tencent Cloud ブロックストレージアカウント ID を使用します。

リソースの取得:
  cloud-resource fetch でゾーン、フレーバー、ネットワーク、サブネットを確認します。Linux イメージとシステムディスクタイプは作成前にプレビューします:
    hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> --preview-request --output json

最小作成コマンド:
  hyperbdrctl cloud-sync-gateway create --cloud-type tencent \
    --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --system-disk-type-id <system_disk_type_id>

作成前に --preview-request を追加して、解決済みリソースと最終リクエストボディを確認します。

作成後にストレージ ID が返されたら、プロビジョニングを待機します:
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`,
	"help.cloud_sync_gateway.create.tencent": `Tencent Cloud ブロックストレージのクラウド同期ゲートウェイを作成します。

パラメータの出所:
  --cloud-account-id string
    cloud-account list または cloud-account detail から取得した Tencent Cloud ブロックストレージアカウント ID を使用します。

リソースの取得:
  利用可能なゾーンを確認します:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --fetch-res zones \
      --output json

  ゲートウェイフレーバーを確認します (最小 2 vCPU / 4 GiB RAM):
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res flavors \
      --purpose make_hg \
      --flavor-vcpus 2 \
      --flavor-ram 4 \
      --output json

  同じゾーンのネットワークとサブネットを確認します:
    hyperbdrctl cloud-resource fetch \
      --cloud-account-id <account_id> \
      --zone-id <zone_id> \
      --fetch-res networks,subnets \
      --output json

  Linux イメージとシステムディスクタイプの解決結果をプレビューします:
    hyperbdrctl cloud-sync-gateway create \
      --cloud-account-id <account_id> \
      --preview-request \
      --output json

最小作成コマンド:
  hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> \
    --zone-id <zone_id> \
    --image-id <image_id> \
    --flavor-id <flavor_id> \
    --network-id <network_id> \
    --subnet-id <subnet_id> \
    --system-disk-type-id <system_disk_type_id>

作成前に最終リクエストボディを確認するには:
  --preview-request

作成後にストレージ ID が返されたら、プロビジョニングを待機して詳細を確認します:
  hyperbdrctl cloud-sync-gateway wait --id <storage_id>
  hyperbdrctl cloud-sync-gateway detail --id <storage_id>`,
}
