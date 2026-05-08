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

func TestCloudAccountsCreateHelpShowsRawBodyFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "account", "create", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"--file", "--body", "--preview-request"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"aliyun_bs_block", "create block", "create oss", "--cloud-type string"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
}

func TestCloudAccountsCreateBlockHelpShowsEnabledProviders(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "account", "create-block", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"aliyun", "openstack"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	if strings.Contains(text, "huaweicloud") {
		t.Fatalf("help should not include disabled provider: %q", text)
	}
}

func TestCloudAccountsCreateOSSHelpShowsEnabledProviders(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "account", "create-oss", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"aliyun", "openstack", "vmware"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	if strings.Contains(text, "aws_obs") {
		t.Fatalf("help should not include disabled provider: %q", text)
	}
}

func TestCloudAccountsCreateBlockAliyunUsesValidatedWorkflow(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"target", "account", "create-block", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-qingdao",
		"--region-name", "North China 1 (Qingdao)",
		"--auth-region-id", "cn-qingdao",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "aliyun_bs" || cloudAccount["cloud_auth_type"] != "aksk" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
}

func TestCloudAccountsCreateBlockOpenStackUsesValidatedWorkflow(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"target", "account", "create-block", "openstack",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--cloud-account-username", "autotest",
		"--cloud-account-password", "0b33333d1f0f3533",
		"--user-domain-id", "default",
		"--project-domain-id", "default",
		"--project-name", "autotest",
		"--region-name", "RegionOne",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "openstack" || cloudAccount["cloud_auth_type"] != "password" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
}

func TestCloudAccountsCreateBlockOpenStackHelpShowsRefinedFlagGuidance(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"--lang", "zh_cn", "target", "account", "create-block", "openstack", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"用法:", "\n参数:\n", "\n使用说明:\n"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\n流程:\n", "\n最小必填参数:\n", "\n常用可选参数:\n", "\n相关命令:\n", "\n下一步:\n"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	for _, want := range []string{
		"OS_AUTH_URL",
		"OpenStack RC 文件",
		"openstack user show",
		"domain_id",
		"OS_PROJECT_DOMAIN_ID",
		"OS_REGION_NAME",
		"最小创建命令如下",
		"--only-verify",
		"--preview-request",
		"target account detail --id <account_id> --output json",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
}

func TestCloudAccountsCreateProviderHelpsUseFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		name      string
		args      []string
		want      []string
		unwanted  []string
		orderWant []string
	}{
		{
			name: "block aliyun",
			args: []string{"target", "account", "create-block", "aliyun", "--help"},
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "fetch-block-resources aliyun", "--only-verify"},
			unwanted: []string{
				"\nExamples:\n",
				"\nNotes:\n",
				"\nWorkflow:\n",
				"\nMinimum Flags:\n",
				"\nCommon Optional Flags:\n",
				"\nRelated Commands:\n",
				"\nNext Steps:\n",
			},
			orderWant: []string{"Usage:", "\nFlags:\n", "Usage Notes:"},
		},
		{
			name: "block huawei generic",
			args: []string{"target", "account", "create-block", "huawei", "--help"},
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "--cloud-auth-type <aksk|password>", "create-block huawei"},
			unwanted: []string{
				"\nExamples:\n",
				"\nNotes:\n",
				"\nWorkflow:\n",
				"\nRelated Commands:\n",
			},
			orderWant: []string{"Usage:", "\nFlags:\n", "Usage Notes:"},
		},
		{
			name: "oss aliyun",
			args: []string{"target", "account", "create-oss", "aliyun", "--help"},
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "boot_loader_images", "fetch-oss-resources aliyun"},
			unwanted: []string{
				"\nExamples:\n",
				"\nNotes:\n",
				"\nWorkflow:\n",
				"\nQuick Start:\n",
				"\nAutomatic Behavior:\n",
				"\nMinimum Flags:\n",
				"\nCommon Optional Flags:\n",
				"\nRelated Commands:\n",
				"\nNext Steps:\n",
			},
			orderWant: []string{"Usage:", "\nFlags:\n", "Usage Notes:"},
		},
		{
			name: "oss openstack",
			args: []string{"target", "account", "create-oss", "openstack", "--help"},
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "Parameter sources:", "OpenStack RC file", "fetch-oss-resources openstack"},
			unwanted: []string{
				"\nExamples:\n",
				"\nNotes:\n",
				"\nWorkflow:\n",
				"\nMinimum Flags:\n",
				"\nCommon Optional Flags:\n",
				"\nRelated Commands:\n",
				"\nNext Steps:\n",
				"--only-verify",
				"--boot-loader-flavor-id",
			},
			orderWant: []string{"Usage:", "\nFlags:\n", "Usage Notes:"},
		},
		{
			name: "oss huawei generic",
			args: []string{"target", "account", "create-oss", "huawei", "--help"},
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "--cloud-auth-type <aksk|password>", "create-oss huawei"},
			unwanted: []string{
				"\nExamples:\n",
				"\nNotes:\n",
				"\nWorkflow:\n",
				"\nRelated Commands:\n",
			},
			orderWant: []string{"Usage:", "\nFlags:\n", "Usage Notes:"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if err := Execute(tt.args, &out, &errOut); err != nil {
				t.Fatal(err)
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
			last := -1
			for _, marker := range tt.orderWant {
				idx := strings.Index(text, marker)
				if idx < 0 {
					t.Fatalf("help missing order marker %q: %q", marker, text)
				}
				if idx <= last {
					t.Fatalf("help order mismatch around %q: %q", marker, text)
				}
				last = idx
			}
			assertNoHelpFooter(t, text)
		})
	}
}

func TestCloudAccountsCreateOSSAliyunUsesValidatedWorkflow(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"target", "account", "create-oss", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-qingdao",
		"--boot-loader-image-id", "win2016_1607_x64_dtc_zh-cn_40G_alibase_20260513.vhd",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "aliyun_obs" || cloudAccount["cloud_auth_type"] != "aksk" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
}

func TestCloudAccountsCreateOSSOpenStackUsesValidatedWorkflow(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"target", "account", "create-oss", "openstack",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--cloud-account-username", "autotest",
		"--cloud-account-password", "0b33333d1f0f3533",
		"--user-domain-id", "default",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "openstack" || cloudAccount["cloud_auth_type"] != "password" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
	if _, ok := body["only_verify"]; ok {
		t.Fatalf("body should not contain only_verify: %+v", body)
	}

	metadata := cloudAccount["metadata"].(map[string]interface{})
	for key, want := range map[string]string{
		"project_domain_id":      "default",
		"project_id":             "8090da32ef324861858a9aad102b8e62",
		"project_name":           "autotest",
		"region_id":              "RegionOne",
		"region_name":            "RegionOne",
		"boot_loader_image_id":   "boot-img-openstack-1",
		"boot_loader_image_name": "Windows_DriverFix_8584ece1",
		"boot_loader_flavor_id":  "flavor-recommend-1",
		"disk_bus_type_id":       "virtio",
		"disk_bus_type_name":     "virtio",
		"custom_name":            "OpenStackCommunity(Juno+)-RegionOne",
	} {
		if metadata[key] != want {
			t.Fatalf("metadata[%s] = %v, want %q; metadata=%+v", key, metadata[key], want, metadata)
		}
	}
}

func TestCloudAccountsCreateOSSOpenStackRejectsOnlyVerify(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "account", "create-oss", "openstack",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--cloud-account-username", "autotest",
		"--cloud-account-password", "autotest",
		"--user-domain-id", "default",
		"--only-verify",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "only-verify cannot be used with target account create-oss openstack") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateBlockGenericProviderFallsBackToGenericBuilder(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"target", "account", "create-block", "huawei",
		"--cloud-auth-type", "aksk",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-north-1",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "huawei_bs" || cloudAccount["cloud_auth_type"] != "aksk" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
}

func TestCloudAccountsCreateOSSGenericProviderFallsBackToGenericBuilder(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"target", "account", "create-oss", "vmware",
		"--cloud-auth-type", "password",
		"--auth-url", "https://vc.example.invalid",
		"--cloud-account-username", "admin",
		"--cloud-account-password", "secret",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "vmware_obs" || cloudAccount["cloud_auth_type"] != "password" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
}

func TestCloudAccountsCreateRawFileRoutesBlockPayload(t *testing.T) {
	dir := t.TempDir()
	bodyPath := filepath.Join(dir, "cloud-account.json")
	if err := os.WriteFile(bodyPath, []byte(`{"cloud_account":{"storage_type":"HyperGate","cloud_type":"huawei_bs"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	path, _ := executeCloudAccountCreateAtPath(t, []string{
		"target", "account", "create",
		"--file", bodyPath,
	})
	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
}

func TestCloudAccountsCreateRawBodyRoutesObjectPayload(t *testing.T) {
	path, _ := executeCloudAccountCreateAtPath(t, []string{
		"target", "account", "create",
		"--body", `{"cloud_account":{"storage_type":"objectstorage","cloud_type":"vmware_obs"}}`,
	})
	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
}

func TestCloudAccountsCreateRawRejectsInvalidStorageType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "account", "create",
		"--body", `{"cloud_account":{"storage_type":"archive"}}`,
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "storage_type must be HyperGate, block, or objectstorage") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateLegacyFlagsShowMigration(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "account", "create",
		"--cloud-type", "aliyun_bs",
		"--access-key-id", "ak",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "target account create --cloud-type ... has been removed") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateLegacyKeyShowsProviderMigration(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "account", "create", "aliyun_bs_block",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "use target account create-block aliyun") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateLegacyBlockVendorShowsProviderMigration(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "account", "create", "block", "aliyun",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "use target account create-block aliyun") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateLegacyOSSVendorShowsProviderMigration(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "account", "create", "oss", "openstack",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "use target account create-oss openstack") {
		t.Fatalf("err = %v", err)
	}
}

func executeCloudAccountCreateAtPath(t *testing.T, commandArgs []string) (string, map[string]interface{}) {
	t.Helper()

	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotQuery string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/postCloudInfoForAuth" {
			var requestBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				t.Fatal(err)
			}
			fetchRes, _ := requestBody["fetch_res"].(string)
			cloudInfo := map[string]interface{}{}
			switch fetchRes {
			case "boot_loader_images":
				cloudInfo["boot_loader_images"] = []map[string]interface{}{
					{"id": "boot-img-auto-1", "name": "Windows DriverFix Auto 1"},
					{"id": "boot-img-auto-2", "name": "Windows DriverFix Auto 2"},
				}
			default:
				cloudInfo["regions"] = []map[string]interface{}{
					{"id": "cn-beijing", "display_name": "North China 2 (Beijing)", "local_name": "华北2（北京）"},
					{"id": "cn-qingdao", "display_name": "North China 1 (Qingdao)", "local_name": "华北1（青岛）"},
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"cloud_info": cloudInfo},
			})
			return
		}
		if r.URL.Path == "/api/v2/postTargetCloudInfoForAuth" {
			var requestBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				t.Fatal(err)
			}
			fetchRes, _ := requestBody["fetch_res"].(string)
			cloudInfo := map[string]interface{}{
				"auth_info": map[string]interface{}{
					"project_domain_id": "default",
					"project_id":        "8090da32ef324861858a9aad102b8e62",
					"project_name":      "autotest",
					"region_id":         "RegionOne",
					"region_name":       "RegionOne",
					"tenant_name":       "autotest",
				},
			}
			switch fetchRes {
			case "flavor":
				cloudInfo["boot_loader_flavors"] = []map[string]interface{}{
					{"id": "flavor-recommend-1", "name": "2C_4G_40G(2C4G)", "is_recommend": 1},
				}
				cloudInfo["boot_loader_images"] = []map[string]interface{}{
					{"id": "boot-img-openstack-1", "name": "Windows_DriverFix_8584ece1"},
					{"id": "boot-img-openstack-2", "name": "Windows_DriverFix_c5ed1912"},
				}
				cloudInfo["disk_bus_types"] = []map[string]interface{}{
					{"id": "virtio", "name": "virtio"},
					{"id": "scsi", "name": "scsi"},
				}
			default:
				cloudInfo["projects"] = []map[string]interface{}{
					{"id": "8090da32ef324861858a9aad102b8e62", "name": "autotest"},
				}
				cloudInfo["regions"] = []map[string]interface{}{
					{"id": "RegionOne", "display_name": "RegionOne"},
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"cloud_info": cloudInfo},
			})
			return
		}

		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"id": "account-1"},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	args := append(withHost(t, srv.URL, "--output", "json"), commandArgs...)
	if err := Execute(args, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	fullPath := gotPath
	if gotQuery != "" {
		fullPath += "?" + gotQuery
	}
	return fullPath, gotBody
}
