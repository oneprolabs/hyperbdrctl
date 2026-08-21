package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestBlockStoragesHelpShowsGuidedSections(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-sync-gateway", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage:",
		"\nFlags:\n",
		"\nCommands:\n",
		"Usage Notes:",
		"create",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
}

func TestBlockStoragesCreateHelpShowsCloudAccountGuide(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute([]string{"cloud-sync-gateway", "create", "--help"}, &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if strings.Contains(text, "Read JSON request body or metadata object from file") {
		t.Fatalf("help should not expose hidden legacy flags: %q", text)
	}
	for _, want := range []string{
		"Usage:",
		"\nFlags:\n",
		"--cloud-account-id",
		"Usage Notes:",
		"hyperbdrctl cloud-account list --storage-type block",
		"hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> --help",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"--cloud-type", "Providers:", "aliyun", "openstack", "huawei", "\nCommands:\n", "\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
}

func TestBlockStoragesCreateGenericProviderHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-sync-gateway", "create", "--cloud-type", "huawei", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage:",
		"\nFlags:\n",
		"Usage Notes:",
		"Parameter Sources:",
		"--cloud-account-id",
		"--boot-types-id string",
		"--preview-request",
		"--cloud-type",
		"cloud-sync-gateway create --cloud-type huawei",
		"cloud-sync-gateway detail --id <storage_id>",
		"cloud-sync-gateway wait --id <storage_id>",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nCommands:\n", "\nWorkflow:\n", "\nMinimum Flags:\n", "\nCommon Optional Flags:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("generic provider help should not include %q: %q", unwanted, text)
		}
	}
	for _, unwanted := range []string{"target cloud-sync-gateway", "cloud-sync-gateway create huawei"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("generic provider help should not include legacy command %q: %q", unwanted, text)
		}
	}
}

func TestBlockStoragesCreateHelpInfersProviderFromCloudAccount(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		name     string
		args     []string
		account  map[string]interface{}
		want     []string
		unwanted []string
	}{
		{
			name: "aliyun account wins over explicit cloud-type",
			args: []string{
				"cloud-sync-gateway", "create",
				"--cloud-account-id", "account-1",
				"--cloud-type", "openstack",
				"--help",
			},
			account: map[string]interface{}{
				"cloud_type":   "aliyun_bs",
				"storage_type": "HyperGate",
			},
			want: []string{
				"Create an Alibaba Cloud cloud sync gateway",
				"Usage: hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> [flags]",
				"--cloud-account-id",
				"Cloud account ID (required)",
				"Zone ID (required)",
				"Image ID (required)",
				"Flavor ID (required)",
				"Network ID (required)",
				"Subnet ID (required)",
				"System disk type ID (required)",
				"System disk size in GiB, default 40",
				"--bandwidth-size",
				"--hd-control-network",
				"Create an Alibaba Cloud cloud sync gateway.",
				"--purpose make_hg",
				"--image_type=system",
				"--fetch-res images,system_volume_types",
				"cloud-resource fetch",
			},
			unwanted: []string{
				"--region-id string",
				"--boot-loader-image-id string    引导加载器镜像 ID（必须）",
				"--cloud-type string",
			},
		},
		{
			name: "openstack account shows openstack profile",
			args: []string{
				"cloud-sync-gateway", "create",
				"--cloud-account-id", "account-1",
				"--help",
			},
			account: map[string]interface{}{
				"cloud_type":   "openstack",
				"storage_type": "HyperGate",
			},
			want: []string{
				"--cloud-account-id",
				"Boot loader image ID",
				"System disk size in GiB, default 50",
				"--boot-types-id string",
				"Usage: hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> [flags]",
				"--fetch-res regions,compute_zones,projects",
				"cloud-sync-gateway wait --id <storage_id>",
			},
			unwanted: []string{
				"Boot loader image ID (required)",
				"--cloud-type string",
			},
		},
		{
			name: "huawei account shows huawei profile in chinese",
			args: []string{
				"--lang", "zh_cn",
				"cloud-sync-gateway", "create",
				"--cloud-account-id", "account-1",
				"--help",
			},
			account: map[string]interface{}{
				"cloud_type":   "huawei_bs",
				"storage_type": "HyperGate",
			},
			want: []string{
				"用法: hyperbdrctl cloud-sync-gateway create --cloud-account-id <account_id> [参数]",
				"--cloud-account-id string        云账号 ID（必须）",
				"创建华为云云同步网关。",
				"--fetch-res images,system_disk_types",
				"cloud-sync-gateway detail --id <storage_id>",
			},
			unwanted: []string{
				"--cloud-type string",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": "00000000",
					"data": tt.account,
				})
			}))
			defer srv.Close()

			var out, errOut bytes.Buffer
			if err := Execute(withHost(t, srv.URL, tt.args...), &out, &errOut); err != nil {
				t.Fatal(err)
			}
			if gotPath != "/hypermotion/v1/cloud_accounts/account-1" {
				t.Fatalf("path=%q", gotPath)
			}

			text := out.String()
			for _, want := range tt.want {
				if !strings.Contains(text, want) {
					t.Fatalf("help missing %q: %q", want, text)
				}
			}
			for _, unwanted := range tt.unwanted {
				if strings.Contains(text, unwanted) {
					t.Fatalf("help should not include %q: %q", unwanted, text)
				}
			}
		})
	}
}

func TestBlockStoragesCreateRejectsBackendCloudType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute([]string{
		"cloud-sync-gateway", "create",
		"--cloud-type", "huawei_bs",
	}, &out, &errOut)
	if err == nil || err.Error() != `cloud-type "huawei_bs" does not support cloud-sync-gateway create` {
		t.Fatalf("err = %v", err)
	}
}

func TestBlockStoragesCreateGenericProviderPreviewRequestBuildsBody(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_type":   "huawei_bs",
				"storage_type": "HyperGate",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"cloud-sync-gateway", "create", "--cloud-type", "huawei",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-north-4",
		"--network-id", "network-1",
		"--boot-loader-image-id", "boot-image-1",
		"--preview-request",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	createStorage := body["create_storage"].(map[string]interface{})
	if createStorage["cloud_type"] != "huawei_bs" || createStorage["cloud_account_uuid"] != "account-1" || createStorage["type"] != "HyperGate" {
		t.Fatalf("create_storage = %#v", createStorage)
	}
	metadata := createStorage["metadata"].(map[string]interface{})
	for key, want := range map[string]string{
		"region_id":            "cn-north-4",
		"network_id":           "network-1",
		"boot_loader_image_id": "boot-image-1",
		"boot_types_id":        "boot_from_volume",
		"volume_proxy_type":    "s3",
		"hg_control_network":   "floating_ip_without_proxy",
		"hg_data_network":      "floating_ip_without_proxy",
		"hd_control_network":   "floating_ip_with_hg_proxy",
	} {
		if metadata[key] != want {
			t.Fatalf("metadata[%q] = %#v, want %q", key, metadata[key], want)
		}
	}
	if _, ok := metadata["cloud_type"]; ok {
		t.Fatalf("metadata should not duplicate cloud_type: %#v", metadata)
	}
}

func TestBlockStoragesCreatePreviewRequestInfersCloudTypeFromCloudAccount(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_type":   "huawei_bs",
				"storage_type": "HyperGate",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"cloud-sync-gateway", "create",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-north-4",
		"--network-id", "network-1",
		"--boot-loader-image-id", "boot-image-1",
		"--preview-request",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/hypermotion/v1/cloud_accounts/account-1" {
		t.Fatalf("path=%q", gotPath)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	createStorage := body["create_storage"].(map[string]interface{})
	if createStorage["cloud_type"] != "huawei_bs" || createStorage["cloud_account_uuid"] != "account-1" || createStorage["type"] != "HyperGate" {
		t.Fatalf("create_storage = %#v", createStorage)
	}
}

func TestBlockStoragesCreatePreviewRequestPrefersCloudAccountOverExplicitCloudType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_type":   "huawei_bs",
				"storage_type": "HyperGate",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"cloud-sync-gateway", "create",
		"--cloud-account-id", "account-1",
		"--cloud-type", "openstack",
		"--region-id", "cn-north-4",
		"--network-id", "network-1",
		"--boot-loader-image-id", "boot-image-1",
		"--preview-request",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	createStorage := body["create_storage"].(map[string]interface{})
	if createStorage["cloud_type"] != "huawei_bs" {
		t.Fatalf("create_storage = %#v", createStorage)
	}
}

func TestBlockStoragesCreateRequiresCloudTypeWithoutCloudAccount(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute([]string{
		"cloud-sync-gateway", "create",
	}, &out, &errOut)
	if err == nil || err.Error() != "cloud-type is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestBlockStoragesCreateAliyunHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-sync-gateway", "create", "--cloud-type", "aliyun", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage:",
		"\nFlags:\n",
		"Usage Notes:",
		"Parameter Sources:",
		"--cloud-account-id",
		"--region-id",
		"--boot-loader-image-id",
		"cloud-resource fetch",
		"cloud-resource fetch --help",
		"cloud-sync-gateway wait --id <storage_id>",
		"floating_ip_without_proxy",
		"floating_ip_with_hg_proxy",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\nCommands:\n", "\nQuick Start:\n", "\nAutomatic behavior:\n", "\nWorkflow:\n", "\nCommon Optional Flags:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	for _, unwanted := range []string{"target cloud-sync-gateway", "cloud-sync-gateway create aliyun"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("aliyun create help should not include legacy command %q: %q", unwanted, text)
		}
	}
}

func TestBlockStoragesCreateOpenStackHelpShowsFourSectionWorkflow(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-sync-gateway", "create", "--cloud-type", "openstack", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage:",
		"\nFlags:\n",
		"Usage Notes:",
		"Parameter Sources:",
		"--cloud-account-id",
		"--boot-loader-image-id",
		"--boot-types-id string",
		"boot_from_volume",
		"default s3",
		"floating_ip_without_proxy",
		"cloud-resource fetch",
		"--preview-request",
		"cloud-sync-gateway detail --id <storage_id> --output json",
		"cloud-sync-gateway wait --id <storage_id>",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	if strings.Contains(text, "cloud-sync-gateway subnet-config") {
		t.Fatalf("help should not advertise subnet-config for openstack: %q", text)
	}
	for _, unwanted := range []string{"target cloud-sync-gateway", "cloud-sync-gateway create openstack"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("openstack create help should not include legacy command %q: %q", unwanted, text)
		}
	}
	for _, unwanted := range []string{"\nWorkflow:\n", "\nMinimum Flags:\n", "\nCommon Optional Flags:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
}

func TestBlockStoragesCreateAliyunValidatedPayloadWithAccountDefaults(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var actionBodies []map[string]interface{}
	var transitionQuery url.Values
	var gotCreateBody map[string]interface{}
	var gotAccountDetailPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1":
			gotAccountDetailPath = r.URL.Path
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_account": map[string]interface{}{
						"cloud_type": "aliyun_bs",
						"metadata": map[string]interface{}{
							"region_type_list": "cn-beijing",
							"auth_region_id":   "cn-qingdao",
						},
					},
				},
			})
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1/action":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			actionBodies = append(actionBodies, body)
			switch len(actionBodies) {
			case 1:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": "00000000",
					"data": map[string]interface{}{
						"cloud_info": map[string]interface{}{
							"domain": map[string]interface{}{
								"regions": []map[string]interface{}{
									{"region_id": "cn-beijing", "region_name": "鍗庡寳2锛堝寳浜級"},
								},
							},
							"zones": []map[string]interface{}{
								{"id": "cn-beijing-l", "display_name": "鍖椾含 鍙敤鍖?L"},
							},
						},
					},
				})
			case 2:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": "00000000",
					"data": map[string]interface{}{
						"cloud_info": map[string]interface{}{
							"flavors": []map[string]interface{}{
								{"id": "ecs.e-c1m2.large", "name": "ecs.e-c1m2.large(2C4G)", "vcpus": 2, "ram_GB": 4, "GHz": "2.5GHz", "quota_rate_name": "", "quota_pps_name": "", "is_recommend": 1},
							},
						},
					},
				})
			case 3:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": "00000000",
					"data": map[string]interface{}{
						"cloud_info": map[string]interface{}{
							"images": []map[string]interface{}{
								{"id": "ubuntu_24_04_x64_20G_alibase_20260506.vhd", "name": "ubuntu_24_04_x64_20G_alibase_20260506.vhd", "os_type": "linux"},
							},
							"system_disk_types": []map[string]interface{}{
								{"id": "cloud_essd_entry", "display_name": "ESSD Entry 浜戠洏"},
							},
						},
					},
				})
			case 4:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": "00000000",
					"data": map[string]interface{}{
						"cloud_info": map[string]interface{}{
							"networks": []map[string]interface{}{
								{"id": "vpc-2ze1neh93b6yjts6g0wja", "name": "wangka2"},
							},
							"subnets": []map[string]interface{}{
								{"id": "vsw-2ze4l4iau8q3qfnhqfhjx", "name": "wangka2-1", "network_id": "vpc-2ze1neh93b6yjts6g0wja"},
							},
							"abilities": map[string]interface{}{},
						},
					},
				})
			default:
				t.Fatalf("unexpected action call %d", len(actionBodies))
			}
		case r.URL.Path == "/api/v3/getCloudInfo":
			transitionQuery = r.URL.Query()
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"images": []map[string]interface{}{
							{"image_id": "win2016_1607_x64_dtc_zh-cn_40G_alibase_20260513.vhd", "image_name": "win2016_1607_x64_dtc_zh-cn_40G_alibase_20260513.vhd"},
						},
					},
				},
			})
		case r.URL.Path == "/hypermotion/v1/storages/action":
			if err := json.NewDecoder(r.Body).Decode(&gotCreateBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"uuid": "storage-1",
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	args := withHost(t, srv.URL,
		"--output", "json",
		"cloud-sync-gateway", "create", "--cloud-type", "aliyun",
		"--cloud-account-id", "account-1",
	)
	if err := Execute(args, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	if gotAccountDetailPath != "/hypermotion/v1/cloud_accounts/account-1" {
		t.Fatalf("account detail path = %q", gotAccountDetailPath)
	}
	if len(actionBodies) != 4 {
		t.Fatalf("action calls = %d", len(actionBodies))
	}
	if transitionQuery.Get("cloud_type") != "aliyun_bs" || transitionQuery.Get("storage_type") != "HyperGate" {
		t.Fatalf("transition query = %+v", transitionQuery)
	}

	createStorage := gotCreateBody["create_storage"].(map[string]interface{})
	if createStorage["type"] != "HyperGate" || createStorage["cloud_type"] != "aliyun_bs" || createStorage["cloud_account_uuid"] != "account-1" {
		t.Fatalf("create_storage = %+v", createStorage)
	}
	metadata := createStorage["metadata"].(map[string]interface{})
	assertEqual := func(key string, want interface{}) {
		if metadata[key] != want {
			t.Fatalf("metadata[%q] = %#v want %#v; metadata=%+v", key, metadata[key], want, metadata)
		}
	}
	assertEqual("region_id", "cn-beijing")
	assertEqual("region_name", "鍗庡寳2锛堝寳浜級")
	assertEqual("zone_id", "cn-beijing-l")
	assertEqual("zone_name", "鍖椾含 鍙敤鍖?L")
	assertEqual("flavor_id", "ecs.e-c1m2.large")
	assertEqual("flavor_name", "ecs.e-c1m2.large(2C4G)")
	assertEqual("vcpu_id", "2 CPU")
	assertEqual("ram_GB", "4 GiB")
	assertEqual("GHz", "2.5GHz")
	assertEqual("image_id", "ubuntu_24_04_x64_20G_alibase_20260506.vhd")
	assertEqual("image_name", "ubuntu_24_04_x64_20G_alibase_20260506.vhd")
	assertEqual("system_disk_type_id", "cloud_essd_entry")
	assertEqual("system_disk_type_name", "cloud_essd_entry")
	assertEqual("system_disk_size", "40")
	assertEqual("volume_proxy_type", "s3")
	assertEqual("volume_proxy_type_name", "S3Block")
	assertEqual("dest_device_type", "vbd")
	assertEqual("dest_device_type_name", "23")
	assertEqual("hg_control_network", "floating_ip_without_proxy")
	assertEqual("hg_control_network_name", "\u516c\u7f51")
	assertEqual("hg_data_network", "floating_ip_without_proxy")
	assertEqual("hg_data_network_name", "\u516c\u7f51")
	assertEqual("bandwidth_size", float64(100))
	assertEqual("hd_control_network", "floating_ip_with_hg_proxy")
	assertEqual("hd_control_network_name", "\u516c\u7f51\u7f51\u7edc\u5e76\u901a\u8fc7\u4e91\u540c\u6b65\u7f51\u5173\u4ee3\u7406")
	assertEqual("network_id", "vpc-2ze1neh93b6yjts6g0wja")
	assertEqual("network_name", "wangka2")
	assertEqual("subnet_id", "vsw-2ze4l4iau8q3qfnhqfhjx")
	assertEqual("subnet_name", "wangka2-1")
	assertEqual("fixed_ip", "")
	assertEqual("win_hd_image_id", "win2016_1607_x64_dtc_zh-cn_40G_alibase_20260513.vhd")
	assertEqual("win_hd_image_name", "win2016_1607_x64_dtc_zh-cn_40G_alibase_20260513.vhd")
	if metadata["control_nat_ip"] != nil || metadata["data_nat_ip"] != nil {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestBlockStoragesCreateAliyunAcceptsRawCloudInfoResponses(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var actionCalls int
	var gotCreateBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_type":   "aliyun_bs",
					"storage_type": "HyperGate",
				},
			})
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1/action":
			actionCalls++
			switch actionCalls {
			case 1:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"domain": map[string]interface{}{
							"regions": []map[string]interface{}{
								{"region_id": "cn-beijing", "region_name": "North China 2 (Beijing)"},
							},
						},
						"zones": []map[string]interface{}{
							{"id": "cn-beijing-l", "display_name": "Beijing Zone L"},
						},
					},
				})
			case 2:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"flavors": []map[string]interface{}{
							{"id": "ecs.e-c1m2.large", "name": "ecs.e-c1m2.large(2C4G)", "vcpus": 2, "ram_GB": 4, "is_recommend": 1},
						},
					},
				})
			case 3:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"images": []map[string]interface{}{
							{"id": "ubuntu_24_04_x64_20G_alibase_20260506.vhd", "name": "ubuntu_24_04_x64_20G_alibase_20260506.vhd", "os_type": "linux"},
						},
						"system_disk_types": []map[string]interface{}{
							{"id": "cloud_essd_entry", "display_name": "ESSD Entry"},
						},
					},
				})
			case 4:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"networks": []map[string]interface{}{
							{"id": "vpc-2ze1neh93b6yjts6g0wja", "name": "vpc-a"},
						},
						"subnets": []map[string]interface{}{
							{"id": "vsw-2ze4l4iau8q3qfnhqfhjx", "name": "subnet-a", "network_id": "vpc-2ze1neh93b6yjts6g0wja"},
						},
					},
				})
			default:
				t.Fatalf("unexpected action call %d", actionCalls)
			}
		case r.URL.Path == "/api/v3/getCloudInfo":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"images": []map[string]interface{}{
						{"image_id": "win2016_1607_x64_dtc_zh-cn_40G_alibase_20260513.vhd", "image_name": "win2016_1607_x64_dtc_zh-cn_40G_alibase_20260513.vhd"},
					},
				},
			})
		case r.URL.Path == "/hypermotion/v1/storages/action":
			if err := json.NewDecoder(r.Body).Decode(&gotCreateBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid": "storage-1",
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"cloud-sync-gateway", "create", "--cloud-type", "aliyun",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-beijing",
		"--boot-loader-image-id", "win2016_1607_x64_dtc_zh-cn_40G_alibase_20260513.vhd",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	if actionCalls != 4 {
		t.Fatalf("actionCalls = %d", actionCalls)
	}
	createStorage := gotCreateBody["create_storage"].(map[string]interface{})
	metadata := createStorage["metadata"].(map[string]interface{})
	if metadata["region_id"] != "cn-beijing" || metadata["zone_id"] != "cn-beijing-l" || metadata["win_hd_image_id"] != "win2016_1607_x64_dtc_zh-cn_40G_alibase_20260513.vhd" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestBlockStoragesCreateAliyunPrefersNetworkThatHasReturnedSubnets(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var actionCalls int
	var gotCreateBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_type":   "aliyun_bs",
					"storage_type": "HyperGate",
				},
			})
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1/action":
			actionCalls++
			switch actionCalls {
			case 1:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"domain": map[string]interface{}{
							"regions": []map[string]interface{}{
								{"region_id": "cn-beijing", "region_name": "North China 2 (Beijing)"},
							},
						},
						"zones": []map[string]interface{}{
							{"id": "cn-beijing-h", "display_name": "Beijing Zone H"},
						},
					},
				})
			case 2:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"flavors": []map[string]interface{}{
							{"id": "ecs.t6-c4m1.large", "name": "ecs.t6-c4m1.large", "vcpus": 4, "ram_GB": 16},
						},
					},
				})
			case 3:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"images": []map[string]interface{}{
							{"id": "ubuntu_24_04_x64_20G_alibase_20260506.vhd", "name": "ubuntu_24_04_x64_20G_alibase_20260506.vhd", "os_type": "linux"},
						},
						"system_disk_types": []map[string]interface{}{
							{"id": "cloud_essd_entry", "display_name": "ESSD Entry"},
						},
					},
				})
			case 4:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"networks": []map[string]interface{}{
							{"id": "vpc-without-subnet", "name": "wrong-first-network"},
							{"id": "vpc-with-subnet", "name": "right-network"},
						},
						"subnets": []map[string]interface{}{
							{"id": "vsw-1", "name": "subnet-1", "network_id": "vpc-with-subnet", "zone_id": "cn-beijing-h"},
						},
					},
				})
			default:
				t.Fatalf("unexpected action call %d", actionCalls)
			}
		case r.URL.Path == "/api/v3/getCloudInfo":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"images": []map[string]interface{}{
						{"image_id": "win-img-1", "image_name": "win-img-1"},
					},
				},
			})
		case r.URL.Path == "/hypermotion/v1/storages/action":
			if err := json.NewDecoder(r.Body).Decode(&gotCreateBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid": "storage-1",
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"cloud-sync-gateway", "create", "--cloud-type", "aliyun",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-beijing",
		"--zone-id", "cn-beijing-h",
		"--boot-loader-image-id", "win-img-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	createStorage := gotCreateBody["create_storage"].(map[string]interface{})
	metadata := createStorage["metadata"].(map[string]interface{})
	if metadata["network_id"] != "vpc-with-subnet" || metadata["subnet_id"] != "vsw-1" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestBlockStoragesCreateAliyunPreviewRequestPrintsRequestBody(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var actionBodies []map[string]interface{}
	var createCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_account": map[string]interface{}{
						"cloud_type":   "aliyun_bs",
						"storage_type": "HyperGate",
						"metadata": map[string]interface{}{
							"region_type_list": "cn-beijing",
						},
					},
				},
			})
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1/action":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			actionBodies = append(actionBodies, body)
			switch len(actionBodies) {
			case 1:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"domain": map[string]interface{}{
							"regions": []map[string]interface{}{
								{"region_id": "cn-beijing", "region_name": "North China 2 (Beijing)"},
							},
						},
						"zones": []map[string]interface{}{
							{"id": "cn-beijing-l", "display_name": "Beijing Zone L"},
						},
					},
				})
			case 2:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"flavors": []map[string]interface{}{
							{"id": "ecs.e-c1m2.large", "name": "ecs.e-c1m2.large(2C4G)", "vcpus": 2, "ram_GB": 4, "is_recommend": 1},
						},
					},
				})
			case 3:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"images": []map[string]interface{}{
							{"id": "ubuntu_24_04_x64_20G_alibase_20260506.vhd", "name": "ubuntu_24_04_x64_20G_alibase_20260506.vhd", "os_type": "linux"},
						},
						"system_disk_types": []map[string]interface{}{
							{"id": "cloud_essd_entry", "display_name": "ESSD Entry"},
						},
					},
				})
			case 4:
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"networks": []map[string]interface{}{
							{"id": "vpc-2ze1neh93b6yjts6g0wja", "name": "vpc-a"},
						},
						"subnets": []map[string]interface{}{
							{"id": "vsw-2ze4l4iau8q3qfnhqfhjx", "name": "subnet-a", "network_id": "vpc-2ze1neh93b6yjts6g0wja"},
						},
					},
				})
			default:
				t.Fatalf("unexpected action call %d", len(actionBodies))
			}
		case r.URL.Path == "/api/v3/getCloudInfo":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"images": []map[string]interface{}{
						{"image_id": "win2016_1607_x64_dtc_zh-cn_40G_alibase_20260513.vhd", "image_name": "win2016_1607_x64_dtc_zh-cn_40G_alibase_20260513.vhd"},
					},
				},
			})
		case r.URL.Path == "/hypermotion/v1/storages/action":
			createCalled = true
			t.Fatalf("preview-request should skip final create request")
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"cloud-sync-gateway", "create", "--cloud-type", "aliyun",
		"--cloud-account-id", "account-1",
		"--preview-request",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if createCalled {
		t.Fatal("create request should not be sent")
	}
	if len(actionBodies) != 4 {
		t.Fatalf("action calls = %d", len(actionBodies))
	}

	var gotBody map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &gotBody); err != nil {
		t.Fatal(err)
	}
	createStorage := gotBody["create_storage"].(map[string]interface{})
	if createStorage["type"] != "HyperGate" || createStorage["cloud_account_uuid"] != "account-1" {
		t.Fatalf("create_storage = %+v", createStorage)
	}
	metadata := createStorage["metadata"].(map[string]interface{})
	if metadata["region_id"] != "cn-beijing" || metadata["zone_id"] != "cn-beijing-l" || metadata["network_id"] != "vpc-2ze1neh93b6yjts6g0wja" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestBlockStoragesCreateOpenStackValidatedPayload(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotCloudInfoPath string
	var gotCloudInfoBody map[string]interface{}
	var gotCreatePath string
	var gotCreateBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_type":   "openstack",
					"storage_type": "HyperGate",
				},
			})
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1/action":
			gotCloudInfoPath = r.URL.Path
			if err := json.NewDecoder(r.Body).Decode(&gotCloudInfoBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"auth_info": map[string]interface{}{
							"project_domain_id": "default",
							"project_id":        "8090da32ef324861858a9aad102b8e62",
							"project_name":      "autotest",
							"region_id":         "RegionOne",
							"region_name":       "RegionOne",
						},
						"projects": []map[string]interface{}{
							{"id": "8090da32ef324861858a9aad102b8e62", "name": "autotest"},
						},
						"regions": []map[string]interface{}{
							{"id": "RegionOne", "name": "RegionOne"},
						},
						"compute_zones": []map[string]interface{}{
							{"id": "nova", "name": "nova"},
						},
						"images": []map[string]interface{}{
							{"id": "cfacf9a2-718a-4cc4-ac76-f25d2467fbc1", "name": "ubuntu24.04-server (Max disks: 20) ", "os_type": "linux"},
						},
						"flavors": []map[string]interface{}{
							{"id": "bdefeffc-57b5-44e6-a64e-8af5548ac8e9", "name": "2C_4G_40G(2C4G)", "is_recommend": 1},
						},
						"networks": []map[string]interface{}{
							{"id": "64882465-93bf-40bd-83b6-9c6a3a970a26", "name": "public-network-10", "display_name": "public-network-10"},
						},
						"subnets": []map[string]interface{}{
							{"id": "", "name": "榛樿", "network_id": "64882465-93bf-40bd-83b6-9c6a3a970a26"},
						},
						"volume_types": []map[string]interface{}{
							{"id": "DEFAULT_VOLUME_TYPE", "name": "DEFAULT_VOLUME_TYPE"},
						},
						"boot_loader_images": []map[string]interface{}{
							{"id": "e85c097f-8a11-4218-98ea-3039282cbee2", "name": "Windows_DriverFix_c5ed1912"},
						},
						"boot_loader_flavors": []map[string]interface{}{
							{"id": "bdefeffc-57b5-44e6-a64e-8af5548ac8e9", "name": "2C_4G_40G(2C4G)", "is_recommend": 1},
						},
					},
				},
			})
		case r.URL.Path == "/hypermotion/v1/storages/action":
			gotCreatePath = r.URL.Path
			if err := json.NewDecoder(r.Body).Decode(&gotCreateBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"uuid": "storage-1",
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	args := withHost(t, srv.URL,
		"--output", "json",
		"cloud-sync-gateway", "create", "--cloud-type", "openstack",
		"--cloud-account-id", "account-1",
		"--boot-loader-image-id", "e85c097f-8a11-4218-98ea-3039282cbee2",
	)
	if err := Execute(args, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	if gotCloudInfoPath != "/hypermotion/v1/cloud_accounts/account-1/action" {
		t.Fatalf("cloud info path = %q", gotCloudInfoPath)
	}
	if gotCreatePath != "/hypermotion/v1/storages/action" {
		t.Fatalf("create path = %q", gotCreatePath)
	}
	if gotCloudInfoBody["get_cloud_info"] == nil {
		t.Fatalf("cloud info body = %+v", gotCloudInfoBody)
	}

	createStorage := gotCreateBody["create_storage"].(map[string]interface{})
	if createStorage["type"] != "HyperGate" || createStorage["cloud_type"] != "openstack" || createStorage["cloud_account_uuid"] != "account-1" {
		t.Fatalf("create_storage = %+v", createStorage)
	}
	metadata := createStorage["metadata"].(map[string]interface{})
	assertEqual := func(key string, want interface{}) {
		if metadata[key] != want {
			t.Fatalf("metadata[%q] = %#v want %#v; metadata=%+v", key, metadata[key], want, metadata)
		}
	}
	assertEqual("project_id", "8090da32ef324861858a9aad102b8e62")
	assertEqual("project_name", "autotest")
	assertEqual("region_id", "RegionOne")
	assertEqual("compute_zone_id", "nova")
	assertEqual("image_id", "cfacf9a2-718a-4cc4-ac76-f25d2467fbc1")
	assertEqual("flavor_id", "bdefeffc-57b5-44e6-a64e-8af5548ac8e9")
	assertEqual("network_id", "64882465-93bf-40bd-83b6-9c6a3a970a26")
	assertEqual("subnet_id", "")
	assertEqual("subnet_name", "榛樿")
	assertEqual("volume_type_id", "DEFAULT_VOLUME_TYPE")
	assertEqual("system_disk_size", "50")
	assertEqual("block_store_zone_id", "nova")
	assertEqual("boot_loader_image_id", "e85c097f-8a11-4218-98ea-3039282cbee2")
	assertEqual("boot_loader_image_name", "Windows_DriverFix_c5ed1912")
	assertEqual("boot_loader_flavor_id", "bdefeffc-57b5-44e6-a64e-8af5548ac8e9")
	assertEqual("boot_types_id", "boot_from_volume")
	assertEqual("boot_types_id_name", "\u5377\u542f\u52a8")
	assertEqual("volume_proxy_type", "s3")
	assertEqual("volume_proxy_type_name", "S3Block")
	assertEqual("hg_control_network", "floating_ip_without_proxy")
	assertEqual("hg_control_network_name", "\u516c\u7f51")
	assertEqual("hg_data_network", "floating_ip_without_proxy")
	assertEqual("hg_data_network_name", "\u516c\u7f51")
	if metadata["control_nat_ip"] != nil || metadata["data_nat_ip"] != nil {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestBlockStoragesCreateOpenStackAcceptsRawCloudInfoResponse(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotCreateBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_type":   "openstack",
					"storage_type": "HyperGate",
				},
			})
		case r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1/action":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"auth_info": map[string]interface{}{
						"project_domain_id": "default",
						"project_id":        "project-1",
						"project_name":      "autotest",
						"region_id":         "RegionOne",
						"region_name":       "RegionOne",
					},
					"projects": []map[string]interface{}{
						{"id": "project-1", "name": "autotest"},
					},
					"regions": []map[string]interface{}{
						{"id": "RegionOne", "name": "RegionOne"},
					},
					"compute_zones": []map[string]interface{}{
						{"id": "nova", "name": "nova"},
					},
					"images": []map[string]interface{}{
						{"id": "img-1", "name": "ubuntu24.04", "os_type": "linux"},
					},
					"flavors": []map[string]interface{}{
						{"id": "flavor-1", "name": "2C4G", "is_recommend": 1},
					},
					"networks": []map[string]interface{}{
						{"id": "net-1", "name": "public-network-10", "display_name": "public-network-10"},
					},
					"subnets": []map[string]interface{}{
						{"id": "", "name": "default", "network_id": "net-1"},
					},
					"volume_types": []map[string]interface{}{
						{"id": "DEFAULT_VOLUME_TYPE", "name": "DEFAULT_VOLUME_TYPE"},
					},
					"boot_loader_images": []map[string]interface{}{
						{"id": "boot-img-1", "name": "Windows_DriverFix"},
					},
					"boot_loader_flavors": []map[string]interface{}{
						{"id": "flavor-1", "name": "2C4G", "is_recommend": 1},
					},
				},
			})
		case r.URL.Path == "/hypermotion/v1/storages/action":
			if err := json.NewDecoder(r.Body).Decode(&gotCreateBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid": "storage-1",
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"cloud-sync-gateway", "create", "--cloud-type", "openstack",
		"--cloud-account-id", "account-1",
		"--boot-loader-image-id", "boot-img-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	createStorage := gotCreateBody["create_storage"].(map[string]interface{})
	metadata := createStorage["metadata"].(map[string]interface{})
	if metadata["project_id"] != "project-1" || metadata["region_id"] != "RegionOne" || metadata["boot_loader_image_id"] != "boot-img-1" {
		t.Fatalf("metadata = %+v", metadata)
	}
}
