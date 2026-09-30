package i18n

import "strings"

// japaneseOverrides contains the Japanese wording for the CLI vocabulary that
// appears in tables, errors, flags, and help headings.
var japaneseOverrides = map[string]string{
	"app.usage":                          "使い方: hyperbdrctl [グローバルオプション] <コマンド> [引数]",
	"app.commands":                       "コマンド: config, host, boot-config, production-site, agent, sync-proxy, license, cloud-account, cloud-resource, cloud-sync-gateway, oss",
	"config.saved":                       "設定を保存しました",
	"error.missing_host":                 "host は必須です",
	"error.missing_username":             "username は必須です",
	"error.missing_password":             "password は必須です",
	"error.missing_id":                   "id は必須です",
	"error.missing_task_id":              "task-id は必須です",
	"error.missing_source_id":            "source-id は必須です",
	"error.missing_storage_id":           "storage-id は必須です",
	"error.missing_source_type":          "type は必須です",
	"error.missing_cloud_type":           "cloud-type は必須です",
	"error.missing_region_id":            "region-id は必須です",
	"error.missing_zone_id":              "zone-id は必須です",
	"error.missing_network_id":           "network-id は必須です",
	"error.missing_connection_uuid":      "connection-uuid は必須です",
	"error.missing_connection_type":      "connection-type は必須です",
	"error.need_login":                   "認証に失敗しました。自動ログインに成功しませんでした",
	"error.output.summary":               "エラー",
	"error.output.details":               "詳細",
	"error.output.code":                  "コード",
	"error.output.trace_id":              "トレース ID",
	"error.output.auth_failed":           "認証に失敗しました",
	"error.output.resource_not_found":    "リソースが見つかりません",
	"error.output.internal_error":        "バックエンド内部エラー",
	"error.output.request_failed":        "リクエストに失敗しました",
	"table.id":                           "ID",
	"table.snapshot_id":                  "スナップショット ID",
	"table.migration_id":                 "移行 ID",
	"table.name":                         "名前",
	"table.name_en":                      "名前 (EN)",
	"table.status":                       "ステータス",
	"table.sync_mode":                    "同期方式",
	"table.available":                    "利用可能",
	"table.sync_capacity":                "同期容量",
	"table.boot_status":                  "起動ステータス",
	"table.health_status":                "健全性",
	"table.host_type":                    "ホスト種別",
	"table.os_type":                      "OS",
	"table.disk_count":                   "ディスク数",
	"table.disk_capacity":                "ディスク容量合計",
	"table.support_sync":                 "同期対応",
	"table.support_increment":            "増分対応",
	"table.storage_id":                   "ストレージ ID",
	"table.storage_name":                 "ストレージ名",
	"table.pool_id":                      "プール ID",
	"table.pool_name":                    "プール名",
	"table.created_at":                   "作成日時",
	"table.execution_time":               "実行時間",
	"table.version":                      "バージョン",
	"table.type":                         "種別",
	"table.uuid":                         "UUID",
	"table.node_ip":                      "ノード IP",
	"table.connections":                  "接続数",
	"table.source":                       "ソース",
	"table.source_id":                    "ソース ID",
	"table.source_type":                  "ソース種別",
	"table.task_id":                      "タスク ID",
	"table.state":                        "状態",
	"table.start_at":                     "開始日時",
	"table.end_at":                       "終了日時",
	"table.region":                       "リージョン",
	"table.current_domain":               "現在のドメイン",
	"table.cloud_type":                   "クラウド種別",
	"table.storage_type":                 "ストレージ種別",
	"table.auth_url":                     "認証 URL",
	"table.url":                          "URL",
	"table.message":                      "メッセージ",
	"table.step":                         "ステップ",
	"table.operation":                    "操作",
	"table.result":                       "結果",
	"table.elapsed_seconds":              "経過秒数",
	"table.error":                        "エラー",
	"table.total":                        "合計",
	"table.value":                        "値",
	"table.amount":                       "数量",
	"table.unused":                       "未使用",
	"table.used":                         "使用済み",
	"table.display_status":               "ステータス",
	"table.display_task_status":          "タスクステータス",
	"table.expire_at":                    "有効期限",
	"table.kkty":                         "登録コード",
	"table.platform":                     "プラットフォーム",
	"table.variant":                      "バリエーション",
	"table.command":                      "コマンド",
	"table.provider":                     "プロバイダー",
	"table.region_count":                 "リージョン数",
	"table.display_username":             "表示ユーザー名",
	"table.public_ip":                    "パブリック IP",
	"table.public_endpoint":              "パブリックエンドポイント",
	"table.internal_endpoint":            "内部エンドポイント",
	"table.protocol":                     "プロトコル",
	"table.bucket_lookup":                "バケット検索",
	"help.section_usage":                 "使い方",
	"help.section_deprecated":            "非推奨",
	"help.section_aliases":               "別名",
	"help.section_quick_start":           "クイックスタート",
	"help.section_automatic_behavior":    "自動動作",
	"help.section_common_flags":          "共通フラグ",
	"help.section_examples":              "例",
	"help.section_notes":                 "注意事項",
	"help.section_workflow":              "ワークフロー",
	"help.section_minimum_flags":         "最小フラグ",
	"help.section_common_optional_flags": "共通オプションフラグ",
	"help.section_commands":              "コマンド",
	"help.section_related_commands":      "関連コマンド",
	"help.section_next_steps":            "次の手順",
	"help.section_flags":                 "フラグ",
	"help.section_global_flags":          "グローバルフラグ",
	"help.section_usage_notes":           "使用上の注意",
	"help.note_required":                 " (必須)",
	"help.note_init_required":            " (初回初期化時に必須)",
	"help.note_required_unless_file":     " (--file を使わない場合は必須)",
	"help.inline_separator":              ", ",
	"help.inline_choices":                "使用可能な値",
	"help.inline_default":                "デフォルト",
	"flag.host":                          "HyperBDR または HyperMotion のエンドポイント（例: https://host:10443）",
	"flag.username":                      "ログインユーザー名",
	"flag.password":                      "ログインパスワード",
	"flag.scene":                         "プラットフォームシーン",
	"flag.timezone":                      "表示タイムゾーン。Local は OS の IANA 名に解決されます",
	"flag.lang":                          "表示言語、使用可能な値 en / zh_cn / ja、デフォルト en",
	"flag.insecure":                      "テスト環境で TLS 証明書の検証をスキップ",
	"flag.output":                        "出力形式、使用可能な値 table / json、デフォルト table",
	"flag.vertical":                      "一覧の各行を mysql 形式のブロックとして縦に表示",
	"flag.debug":                         "リクエストのデバッグログを出力",
	"flag.version":                       "CLI バージョン情報を表示",
	"flag.help":                          "ヘルプ情報を表示",
	"flag.method":                        "HTTP メソッド",
	"flag.path":                          "/ から始まるホスト相対 API パス",
	"flag.file":                          "JSON リクエスト本文またはメタデータをファイルから読み込む",
	"flag.body":                          "インライン JSON リクエスト本文",
	"flag.query":                         "key=value 形式の追加クエリ項目（繰り返し可能）",
	"flag.header":                        "key=value 形式の追加 HTTP ヘッダー（繰り返し可能）",
	"flag.set":                           "path=value 形式のメタデータ上書き（繰り返し可能）",
	"flag.set-json":                      "path=<json> 形式のメタデータ上書き（繰り返し可能）",
	"flag.preview-request":               "リクエストを送信せず本文を出力",
	"flag.show-secret":                   "保存済みパスワードを平文で表示",
	"flag.page":                          "ページ番号",
	"flag.page-size":                     "ページサイズ",
	"flag.status":                        "ステータスフィルター",
	"flag.boot-status":                   "起動ステータスフィルター",
	"flag.kw":                            "キーワードフィルター",
	"flag.cloud-type":                    "クラウドプラットフォーム種別",
	"flag.cloud-auth-type":               "クラウドアカウント認証種別",
	"flag.ids":                           "コンマ区切りの ID",
	"flag.macs":                          "コンマ区切りの MAC アドレス",
	"flag.id":                            "リソース ID",
	"flag.sync-detail":                   "スナップショット一覧に同期詳細を含める",
	"flag.mode":                          "同期モード",
	"flag.transfer-speed":                "転送速度制限（MB/s）",
	"flag.vm-id":                         "単一のソース VM ID",
	"flag.vm-ids":                        "コンマ区切りのソース VM ID",
	"flag.snapshot-id":                   "起動元のスナップショット ID",
	"flag.force":                         "強制操作",
	"flag.operation":                     "待機する操作",
	"flag.interval-seconds":              "ポーリング間隔（秒）",
	"flag.timeout-seconds":               "全体のタイムアウト（秒）",
	"flag.include-steps":                 "待機結果にステップ詳細を含める",
	"flag.type":                          "種別フィルター",
	"flag.storage-id":                    "ストレージ ID",
	"flag.storage-type":                  "ストレージ種別",
	"flag.cloud-account-id":              "クラウドアカウント ID",
	"flag.fetch-res":                     "要求するリソースグループ（コンマ区切り）",
	"flag.host-id":                       "ホスト ID",
	"resource.regions":                   "リージョン",
	"resource.zones":                     "ゾーン",
	"resource.flavors":                   "フレーバー",
	"resource.images":                    "イメージ",
	"resource.networks":                  "ネットワーク",
	"resource.subnets":                   "サブネット",
	"resource.security_groups":           "セキュリティグループ",
	"resource.volume_types":              "ボリューム種別",
	"resource.system_volume_types":       "システムボリューム種別",
	"resource.os_types":                  "OS 種別",
	"resource.projects":                  "プロジェクト",
	"resource.compute_zones":             "コンピュートゾーン",
}

func buildJapaneseCatalog(_ map[string]string) map[string]string {
	ja := make(map[string]string, len(japaneseGenerated)+len(japaneseSpecialHelp)+len(japaneseResidual)+len(japaneseOverrides))
	for key, value := range japaneseGenerated {
		ja[key] = value
	}
	for key, value := range japaneseSpecialHelp {
		ja[key] = value
	}
	for key, value := range japaneseResidual {
		ja[key] = value
	}
	for key, value := range japaneseOverrides {
		ja[key] = value
	}
	applyJapaneseTextCorrections(ja)
	return ja
}

var japaneseTextCorrections = map[string]string{
	"Create or update a boot configuration from cloud account ":                                                               "クラウドアカウント ",
	"Use the cloud sync gateway ID, usually from `hyperbdrctl cloud-sync-gateway list`.":                                      "通常は `hyperbdrctl cloud-sync-gateway list` から取得したクラウド同期ゲートウェイ ID を使用します。",
	"Run the minimum query like this:":                                                                                        "最小限のクエリは次のとおりです:",
	"`--fetch-res` is optional. When omitted, the CLI auto-detects the returned resource sections and renders them in order.": "`--fetch-res` は省略可能です。省略した場合、CLI は返されたリソースセクションを自動検出し、順番に表示します。",
	"When you need to continue the query with selected context, add:":                                                         "選択したコンテキストでクエリを続行する必要がある場合は、次を追加します:",
	"If the provider uses username/password authentication, start with:":                                                      "プロバイダーがユーザー名/パスワード認証を使用する場合は、次から開始します:",
	"The alias form is also accepted:":                                                                                        "エイリアス形式も使用できます:",
	"Include `--region-id` for region-scoped resources such as images.":                                                       "イメージなど、リージョンにスコープされたリソースには `--region-id` を指定します。",
	"Include `--zone-id` or other provider-specific dynamic flags only when the requested resource type needs them.":          "要求したリソース種別で必要な場合にのみ、`--zone-id` またはその他のプロバイダー固有の動的フラグを指定します。",
	"For scripting against the raw fields, add:":                                                                              "生のフィールドをスクリプトで扱う場合は、次を追加します:",
	"To continue overriding metadata, add `--set key=value` or `--set-json key=<json>`.":                                      "メタデータの上書きを続けるには、`--set key=value` または `--set-json key=<json>` を追加します。",
	"The override order is `--set-json < --set < explicit and dynamic flags`.":                                                "上書き順序は `--set-json < --set < 明示的フラグおよび動的フラグ` です。",
	"To inspect the final request body without sending a write request, add `--preview-request`.":                             "書き込みリクエストを送信せずに最終リクエスト本文を確認するには、`--preview-request` を追加します。",
	"`--volume-type-id`, and `--boot-loader-flavor-id`":                                                                       "`--volume-type-id`、および `--boot-loader-flavor-id`",
	"`hg-control-network=floating_ip_without_proxy`, and `hg-data-network=floating_ip_without_proxy`":                         "`hg-control-network=floating_ip_without_proxy`、および `hg-data-network=floating_ip_without_proxy`",
	"  2 つの vCPU という最小ゲートウェイ要件を満たす候補と、\n  4 GiB RAM。":                                                                         "  最低要件である 2 vCPU と 4 GiB RAM を満たす候補を選択します。",
}

func applyJapaneseTextCorrections(catalog map[string]string) {
	for key, value := range catalog {
		for english, japanese := range japaneseTextCorrections {
			value = strings.ReplaceAll(value, english, japanese)
		}
		catalog[key] = value
	}
}

func japaneseFragment(english map[string]string) map[string]string {
	fragment := make(map[string]string, len(english))
	for key := range english {
		fragment[key] = japaneseCatalog[key]
	}
	return fragment
}

func mergeJapaneseFragment(fragment map[string]string, overlays ...map[string]string) map[string]string {
	for _, overlay := range overlays {
		for key, value := range overlay {
			if _, ok := fragment[key]; ok {
				fragment[key] = value
			}
		}
	}
	return fragment
}
