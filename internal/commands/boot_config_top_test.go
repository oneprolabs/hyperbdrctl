package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTopLevelBootConfigRequiresSubcommand(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "boot-config"), &out, &errOut)
	if err == nil || err.Error() != "boot-config requires subcommand" {
		t.Fatalf("err = %v", err)
	}
}

func TestRemovedTopLevelBootConfigFetchCommandsReturnUnknownCommand(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := [][]string{
		{"boot-config", "fetch-block-resources"},
		{"boot-config", "fetch-oss-resources"},
	}

	for _, args := range cases {
		var out, errOut bytes.Buffer
		err := Execute(args, &out, &errOut)
		if err == nil || !strings.Contains(err.Error(), "unknown") {
			t.Fatalf("args=%v err=%v", args, err)
		}
	}
}

func TestTopLevelBootConfigApplyHelpUsesModernLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "boot-config", "apply"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"Usage:",
		"\nFlags:\n",
		"--id",
		"--cloud-account-id",
		"--set",
		"--set-json",
		"--preview-request",
		"Usage Notes:",
		"hyperbdrctl host list",
		"hyperbdrctl cloud-account list --storage-type block",
		"hyperbdrctl cloud-sync-gateway list",
		"hyperbdrctl oss list",
		"hyperbdrctl cloud-account list --storage-type object",
		"hyperbdrctl boot-config apply --cloud-account-id <account_id> --help",
		"For block storage, first view:",
		"For object storage, first view:",
		"automatically matches the cloud provider and storage type",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("help missing %q: %q", want, got)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nCommands:\n", "\nGlobal Flags:\n", "\nRelated Commands:\n", "\nAutomatic behavior:\n", "\nMinimum Flags:\n"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, got)
		}
	}
	if strings.Contains(got, "--file string") {
		t.Fatalf("help should hide file flag from parameter block: %q", got)
	}
	for _, unwanted := range []string{"fetch-block-resources", "fetch-oss-resources"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("help should not include retired fetch command %q: %q", unwanted, got)
		}
	}
	assertNoHelpFooter(t, got)
}

func TestTopLevelBootConfigApplyAccountHelpReadsAccountDetailForBlockProfile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_type":   "aliyun_bs",
				"storage_type": "HyperGate",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config", "apply",
		"--cloud-account-id", "account-1",
		"--help",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/hypermotion/v1/cloud_accounts/account-1" {
		t.Fatalf("path=%q", gotPath)
	}

	text := out.String()
	for _, want := range []string{
		"--id",
		"--cloud-account-id",
		"--storage-id",
		"--volume-type-id",
		"--security-group-id",
		"Create or update boot configuration for cloud account <account_id>",
		"provider `aliyun`, storage type `block`",
		"hyperbdrctl cloud-sync-gateway list",
		"--fetch-res networks,subnets",
		"--volume-type-id <volume_type_id>",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"--boot-loader-image-id", "--system-volume-type-id", "--file string"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestTopLevelBootConfigApplyAccountHelpShowsOpenStackObjectProfile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_type":   "openstack",
				"storage_type": "objectstorage",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config", "apply",
		"--cloud-account-id", "account-1",
		"--help",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"--project-id",
		"--project-domain-id",
		"--compute-zone-id",
		"--system-volume-type-id",
		"--boot-loader-image-id",
		"--boot-loader-flavor-id",
		"hyperbdrctl oss list",
		"provider `openstack`, storage type `object`",
		"regions,compute_zones,projects",
		"system_volume_types,volume_types",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestTopLevelBootConfigApplyAccountHelpRequiresResolvableContext(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"storage_type": "HyperGate",
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config", "apply",
		"--cloud-account-id", "account-1",
		"--help",
	), &out, &errOut)
	if err == nil || err.Error() != "cloud-type cannot be inferred from cloud-account-id" {
		t.Fatalf("err=%v", err)
	}
}

func TestTopLevelBootConfigHelpShowsGetSubcommand(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "boot-config"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"Commands:",
		"get",
		"apply",
		"Usage Notes:",
		"hyperbdrctl host list",
		"hyperbdrctl boot-config get --id <host_id>",
		"hyperbdrctl boot-config apply --help",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("help missing %q: %q", want, got)
		}
	}
	for _, unwanted := range []string{"fetch-block-resources", "fetch-oss-resources"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("help should hide %q: %q", unwanted, got)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nGlobal Flags:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, got)
		}
	}
	assertNoHelpFooter(t, got)
}

func TestTopLevelBootConfigGetHelpUsesModernLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "boot-config", "get"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"Usage:",
		"\nFlags:\n",
		"--id",
		"Usage Notes:",
		"hyperbdrctl boot-config get --id <host_id>",
		"hyperbdrctl --output json boot-config get --id <host_id>",
		"Read the host's current boot configuration.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("help missing %q: %q", want, got)
		}
	}
	for _, unwanted := range []string{"\nExamples:\n", "\nCommands:\n", "\nNotes:\n", "\nGlobal Flags:\n"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, got)
		}
	}
	assertNoHelpFooter(t, got)
}

func TestTopLevelBootConfigGetHelpZhCNMatchesArchiveGuidance(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--lang", "zh_cn", "help", "boot-config", "get"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "读取主机当前已有的启动配置。") {
		t.Fatalf("help missing archived guidance: %q", got)
	}
	if strings.Contains(got, "该命令用于读取主机当前已有的启动配置。") {
		t.Fatalf("help contains stale guidance: %q", got)
	}
}

func TestTopLevelBootConfigHelpZhCNMatchesArchiveGuidance(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--lang", "zh_cn", "help", "boot-config"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"主机启动配置",
		"查看主机启动配置",
		"查看或创建主机启动配置。",
		"hyperbdrctl host list",
		"hyperbdrctl boot-config get --id <host_id>",
		"hyperbdrctl boot-config apply --help",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("help missing %q: %q", want, got)
		}
	}
}

func TestTopLevelBootConfigGetAvailable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"id":             "host-1",
					"boot_config_id": "cfg-1",
					"boot_config":    map[string]interface{}{},
				},
			})
		case "/api/v2/batchGetBootConfigs":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_configs": []map[string]interface{}{
						{
							"id":           "cfg-1",
							"migration_id": "host-1",
							"storage_id":   "storage-1",
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "boot-config", "get", "--id", "host-1"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 2 || gotPaths[0] != "/api/v2/getHostDetail" || gotPaths[1] != "/api/v2/batchGetBootConfigs" {
		t.Fatalf("paths = %+v", gotPaths)
	}
	if !strings.Contains(out.String(), `"boot_configs"`) || !strings.Contains(out.String(), `"migration_id"`) {
		t.Fatalf("output = %q", out.String())
	}
}

func TestTopLevelBootConfigApplyPreviewRequestCreate(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-top-preview-create.json")
	if err := os.WriteFile(bodyPath, []byte(`{"cloud_type":"aliyun_bs","region_id":"cn-beijing"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"id": "host-1"},
			})
		case "/api/v2/batchBootConfigs", "/api/v2/batchUpdateBootConfigs":
			t.Fatalf("unexpected write path %q", r.URL.Path)
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config", "apply",
		"--id", "host-1",
		"--file", bodyPath,
		"--preview-request",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 1 || gotPaths[0] != "/api/v2/getHostDetail" {
		t.Fatalf("paths = %+v", gotPaths)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	rows := got["batch_create"].([]interface{})
	item := rows[0].(map[string]interface{})
	if item["migration_id"] != "host-1" {
		t.Fatalf("body = %+v", got)
	}
	meta := item["metadata"].(map[string]interface{})
	if meta["region_id"] != "cn-beijing" {
		t.Fatalf("body = %+v", got)
	}
}

func TestTopLevelBootConfigApplyPreviewRequestUpdate(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-top-preview-update.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1","repair_host_mapper":{"enable_dhcp_mode":"0"}}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getStorageDetailInfo":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"storage_type": "HyperGate"},
			})
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"boot_config": map[string]interface{}{"id": "cfg-1"}},
			})
		case "/api/v2/batchGetBootConfigs":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_configs": []map[string]interface{}{
						{
							"id": "cfg-1",
							"metadata": map[string]interface{}{
								"storage_id": "storage-1",
								"os_type":    "",
								"repair_host_mapper": map[string]interface{}{
									"enable_dhcp_mode":     "1",
									"enable_inject_driver": "1",
									"enable_repair_fs":     "1",
									"os_version":           "auto_check",
									"os_display_name":      "auto_check",
									"pre_script":           "",
									"post_script":          "",
								},
							},
						},
					},
				},
			})
		case "/api/v2/batchBootConfigs", "/api/v2/batchUpdateBootConfigs":
			t.Fatalf("unexpected write path %q", r.URL.Path)
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config", "apply",
		"--id", "host-1",
		"--file", bodyPath,
		"--preview-request=true",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 3 || gotPaths[0] != "/api/v2/getStorageDetailInfo" || gotPaths[1] != "/api/v2/getHostDetail" || gotPaths[2] != "/api/v2/batchGetBootConfigs" {
		t.Fatalf("paths = %+v", gotPaths)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	rows := got["batch_update"].([]interface{})
	item := rows[0].(map[string]interface{})
	if item["id"] != "cfg-1" || item["migration_id"] != "host-1" {
		t.Fatalf("body = %+v", got)
	}
	meta := item["metadata"].(map[string]interface{})
	if len(meta) != 1 {
		t.Fatalf("body = %+v", got)
	}
	if _, ok := meta["repair_host_mapper"]; !ok {
		t.Fatalf("body = %+v", got)
	}
}

func TestTopLevelBootConfigApplyCreatesWithIndependentOverrides(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-top-create.json")
	body := `{
		"region_id": "cn-beijing",
		"zone_id": "cn-beijing-h",
		"network_id": "vpc-old",
		"subnet_id": "vsw-old",
		"security_group_id": "sg-old",
		"dest_boot_mode": "uefi",
		"repair_host_mapper": {"enable_dhcp_mode": "1", "enable_inject_driver": "1"},
		"nics": [{"index":0,"subnet_id":"vsw-old","security_groups":[{"id":"sg-old"}]}],
		"instance_config_mapper": {},
		"disk_volume_mapper": [{"index":0,"volume_type_id":"cloud_efficiency"}]
	}`
	if err := os.WriteFile(bodyPath, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}

	var gotPaths []string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"id": "host-1"},
			})
		case "/api/v2/batchBootConfigs":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"status": "ok"},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config", "apply",
		"--id", "host-1",
		"--file", bodyPath,
		"--region-id", "cn-shanghai",
		"--zone-id", "cn-shanghai-b",
		"--network-id", "vpc-new",
		"--subnet-id", "vsw-new",
		"--security-group-id", "sg-new",
		"--dest-boot-mode", "bios",
		"--set", "repair_host_mapper.enable_dhcp_mode=0",
		"--set", "disk_volume_mapper[0].volume_type_id=cloud_essd",
		"--set-json", `nics=[{"index":0,"subnet_id":"vsw-json","security_groups":[{"id":"sg-json"}]}]`,
		"--set-json", `instance_config_mapper={"image_id":"img-1"}`,
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 2 || gotPaths[0] != "/api/v2/getHostDetail" || gotPaths[1] != "/api/v2/batchBootConfigs" {
		t.Fatalf("paths = %+v", gotPaths)
	}
	item := gotBody["batch_create"].([]interface{})[0].(map[string]interface{})
	meta := item["metadata"].(map[string]interface{})
	if meta["region_id"] != "cn-shanghai" || meta["zone_id"] != "cn-shanghai-b" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["network_id"] != "vpc-new" || meta["subnet_id"] != "vsw-new" || meta["security_group_id"] != "sg-new" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["dest_boot_mode"] != "bios" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["repair_host_mapper"].(map[string]interface{})["enable_dhcp_mode"] != float64(0) {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["disk_volume_mapper"].([]interface{})[0].(map[string]interface{})["volume_type_id"] != "cloud_essd" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["nics"].([]interface{})[0].(map[string]interface{})["subnet_id"] != "vsw-json" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["instance_config_mapper"].(map[string]interface{})["image_id"] != "img-1" {
		t.Fatalf("metadata = %+v", meta)
	}
	if !strings.Contains(out.String(), "create") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestTopLevelBootConfigApplyCreatesWithoutFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"id": "host-1"},
			})
		case "/api/v2/batchBootConfigs":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"status": "ok"},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config", "apply",
		"--id", "host-1",
		"--cloud-type", "aliyun_bs",
		"--region-id", "cn-beijing",
		"--set", "repair_host_mapper.enable_dhcp_mode=1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	item := gotBody["batch_create"].([]interface{})[0].(map[string]interface{})
	meta := item["metadata"].(map[string]interface{})
	if meta["cloud_type"] != "aliyun_bs" || meta["region_id"] != "cn-beijing" {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestTopLevelBootConfigApplyUpdatesWhenBootConfigExists(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-top-update.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1","repair_host_mapper":{"enable_dhcp_mode":"0"}}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/getStorageDetailInfo":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"storage_type": "HyperGate"},
			})
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"boot_config_id": "cfg-1"},
			})
		case "/api/v2/batchGetBootConfigs":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_configs": []map[string]interface{}{
						{
							"id": "cfg-1",
							"metadata": map[string]interface{}{
								"storage_id": "storage-1",
								"os_type":    "",
								"repair_host_mapper": map[string]interface{}{
									"enable_dhcp_mode":     "1",
									"enable_inject_driver": "1",
									"enable_repair_fs":     "1",
									"os_version":           "auto_check",
									"os_display_name":      "auto_check",
									"pre_script":           "",
									"post_script":          "",
								},
							},
						},
					},
				},
			})
		case "/api/v2/batchUpdateBootConfigs":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"status": "ok"},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "boot-config", "apply", "--id", "host-1", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	item := gotBody["batch_update"].([]interface{})[0].(map[string]interface{})
	if item["id"] != "cfg-1" || item["migration_id"] != "host-1" {
		t.Fatalf("body = %+v", gotBody)
	}
	meta := item["metadata"].(map[string]interface{})
	if len(meta) != 1 {
		t.Fatalf("body = %+v", gotBody)
	}
	if _, ok := meta["repair_host_mapper"]; !ok {
		t.Fatalf("body = %+v", gotBody)
	}
	if !strings.Contains(out.String(), "update") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestTopLevelBootConfigApplyReturnsJSONNoOpWhenNothingChanges(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "boot-config-top-noop.json")
	if err := os.WriteFile(bodyPath, []byte(`{"storage_id":"storage-1","repair_host_mapper":{"enable_dhcp_mode":"1"}}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getStorageDetailInfo":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"storage_type": "HyperGate"},
			})
		case "/api/v2/getHostDetail":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"boot_config_id": "cfg-1"},
			})
		case "/api/v2/batchGetBootConfigs":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"boot_configs": []map[string]interface{}{
						{
							"id": "cfg-1",
							"metadata": map[string]interface{}{
								"storage_id": "storage-1",
								"os_type":    "",
								"repair_host_mapper": map[string]interface{}{
									"enable_dhcp_mode":     "1",
									"enable_inject_driver": "1",
									"enable_repair_fs":     "1",
									"os_version":           "auto_check",
									"os_display_name":      "auto_check",
									"pre_script":           "",
									"post_script":          "",
								},
							},
						},
					},
				},
			})
		case "/api/v2/batchUpdateBootConfigs":
			t.Fatalf("unexpected write path %q", r.URL.Path)
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "boot-config", "apply", "--id", "host-1", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 3 || gotPaths[2] != "/api/v2/batchGetBootConfigs" {
		t.Fatalf("paths = %+v", gotPaths)
	}
	if !strings.Contains(out.String(), `"no_op": true`) || !strings.Contains(out.String(), `"operation": "update"`) || !strings.Contains(out.String(), `"boot_config_id": "cfg-1"`) {
		t.Fatalf("output = %q", out.String())
	}
}

func TestTopLevelBootConfigApplyValidationErrors(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	bodyPath := filepath.Join(dir, "boot-config-top-invalid.json")
	if err := os.WriteFile(bodyPath, []byte(`{"nics":[{"subnet_id":"vsw-1"}]}`), 0600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "boot-config", "apply", "--file", bodyPath), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}

	err = Execute(withHost(t, "https://example.invalid", "boot-config", "apply", "--id", "host-1"), &out, &errOut)
	if err == nil || err.Error() != "file is required when no metadata override flags are provided" {
		t.Fatalf("err = %v", err)
	}

	err = Execute(withHost(t, "https://example.invalid", "boot-config", "apply", "--id", "host-1", "--file", bodyPath, "--set", "bad"), &out, &errOut)
	if err == nil || err.Error() != "assignment must be in path=value format" {
		t.Fatalf("err = %v", err)
	}

	err = Execute(withHost(t, "https://example.invalid", "boot-config", "apply", "--id", "host-1", "--file", bodyPath, "--set-json", "nics={bad}"), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "invalid JSON for nics:") {
		t.Fatalf("err = %v", err)
	}

	err = Execute(withHost(t, "https://example.invalid", "boot-config", "apply", "--id", "host-1", "--file", bodyPath, "--set", "nics[2].subnet_id=vsw-2"), &out, &errOut)
	if err == nil || err.Error() != "path nics[2].subnet_id index 2 out of range" {
		t.Fatalf("err = %v", err)
	}

	err = Execute(withHost(t, "https://example.invalid", "boot-config", "apply", "--id", "--file", bodyPath), &out, &errOut)
	if err == nil || err.Error() != "--id requires value" {
		t.Fatalf("err = %v", err)
	}
}

func TestTopLevelBootConfigApplyRejectsInvalidMetadataFileShapes(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "array", raw: `[{}]`, want: "file must contain single metadata object"},
		{name: "scalar", raw: `"x"`, want: "file must contain metadata object"},
		{name: "batch create", raw: `{"batch_create":[]}`, want: "file must contain metadata object, not batch_create wrapper"},
		{name: "batch update", raw: `{"batch_update":[]}`, want: "file must contain metadata object, not batch_update wrapper"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)
			bodyPath := filepath.Join(dir, "boot-config-top-shape.json")
			if err := os.WriteFile(bodyPath, []byte(tc.raw), 0600); err != nil {
				t.Fatal(err)
			}
			var out, errOut bytes.Buffer
			err := Execute(withHost(t, "https://example.invalid", "boot-config", "apply", "--id", "host-1", "--file", bodyPath), &out, &errOut)
			if err == nil || err.Error() != tc.want {
				t.Fatalf("err = %v want %q", err, tc.want)
			}
		})
	}
}
