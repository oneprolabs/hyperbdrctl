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
	if err := Execute([]string{"cloud-account", "create", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"--preview-request"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"aliyun_bs_block", "create block", "create oss", "--access-key-id", "--auth-url", "--file string", "--body string"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
}

func TestParseCloudAccountCreateRawArgsSupportsBody(t *testing.T) {
	parsed, err := parseCloudAccountCreateRawArgs([]string{
		"--body", `{"cloud_account":{"storage_type":"objectstorage","cloud_type":"openstack","metadata":{"account_name":"body-account"}}}`,
		"--preview-request",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.previewRequest {
		t.Fatalf("previewRequest = %v", parsed.previewRequest)
	}
	cloudAccount := parsed.body["cloud_account"].(map[string]interface{})
	if cloudAccount["storage_type"] != "objectstorage" || cloudAccount["cloud_type"] != "openstack" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if metadata["account_name"] != "body-account" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestCloudAccountsCreateBlockHelpShowsEnabledProviders(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-account", "create", "--storage-type", "block", "--help"}, &out, &errOut); err != nil {
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
	if err := Execute([]string{"cloud-account", "create", "--storage-type", "object", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"aliyun", "huawei", "openstack"} {
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
		"cloud-account", "create", "--cloud-type", "aliyun", "--storage-type", "block",
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
		"cloud-account", "create", "--cloud-type", "openstack", "--storage-type", "block",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--username", "autotest",
		"--password", "0b33333d1f0f3533",
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
	if err := Execute([]string{"--lang", "zh_cn", "cloud-account", "create", "--cloud-type", "openstack", "--storage-type", "block", "--help"}, &out, &errOut); err != nil {
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
		"--set stringArray",
		"--set-json stringArray",
		"--foo-bar <value>",
		"--preview-request",
		"cloud-account detail --id <account_id> --output json",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"--only-verify", "--file string"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include removed flag %q: %q", unwanted, text)
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
			args: []string{"cloud-account", "create", "--cloud-type", "aliyun", "--storage-type", "block", "--help"},
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "cloud-resource fetch --cloud-type aliyun --storage-type block", "--set stringArray", "--set-json stringArray", "--foo-bar <value>"},
			unwanted: []string{
				"\nExamples:\n",
				"\nNotes:\n",
				"\nWorkflow:\n",
				"\nMinimum Flags:\n",
				"\nCommon Optional Flags:\n",
				"\nRelated Commands:\n",
				"\nNext Steps:\n",
				"--only-verify",
				"--file string",
			},
			orderWant: []string{"Usage:", "\nFlags:\n", "Usage Notes:"},
		},
		{
			name: "block huawei generic",
			args: []string{"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block", "--help"},
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "--access-key-id string", "--access-key-secret string", "--account-name string", "--set stringArray", "--set-json stringArray", "cloud-account create --cloud-type huawei --storage-type block", "--foo-bar <value>", "--access-id <ak>", "--access-secret <sk>"},
			unwanted: []string{
				"\nExamples:\n",
				"\nNotes:\n",
				"\nWorkflow:\n",
				"\nRelated Commands:\n",
				"--only-verify",
				"--file string",
				"--cloud-auth-type <aksk|password>",
				"--auth-url",
				"--username <username>",
				"--password <password>",
				"If both AK/SK-style and username/password-style flags are present",
			},
			orderWant: []string{"Usage:", "\nFlags:\n", "Usage Notes:"},
		},
		{
			name: "oss aliyun",
			args: []string{"cloud-account", "create", "--cloud-type", "aliyun", "--storage-type", "object", "--help"},
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "boot_loader_images", "cloud-resource fetch --cloud-type aliyun --storage-type object", "--set stringArray", "--set-json stringArray"},
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
				"--file string",
			},
			orderWant: []string{"Usage:", "\nFlags:\n", "Usage Notes:"},
		},
		{
			name: "oss openstack",
			args: []string{"cloud-account", "create", "--cloud-type", "openstack", "--storage-type", "object", "--help"},
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "Parameter sources:", "OpenStack RC file", "cloud-resource fetch --cloud-type openstack --storage-type object", "--set stringArray", "--set-json stringArray"},
			unwanted: []string{
				"\nExamples:\n",
				"\nNotes:\n",
				"\nWorkflow:\n",
				"\nMinimum Flags:\n",
				"\nCommon Optional Flags:\n",
				"\nRelated Commands:\n",
				"\nNext Steps:\n",
				"--only-verify",
				"--file string",
			},
			orderWant: []string{"Usage:", "\nFlags:\n", "Usage Notes:"},
		},
		{
			name: "oss huawei generic",
			args: []string{"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object", "--help"},
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "--cloud-auth-type <aksk|password>", "cloud-account create --cloud-type huawei --storage-type object", "--set stringArray", "--set-json stringArray", "--foo-bar <value>", "--access-id + --access-secret => aksk", "If both AK/SK-style and username/password-style flags are present"},
			unwanted: []string{
				"\nExamples:\n",
				"\nNotes:\n",
				"\nWorkflow:\n",
				"\nRelated Commands:\n",
				"--auto-upload-images",
				"--only-verify",
				"--file string",
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
		"cloud-account", "create", "--cloud-type", "aliyun", "--storage-type", "object",
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
		"cloud-account", "create", "--cloud-type", "openstack", "--storage-type", "object",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--username", "autotest",
		"--password", "0b33333d1f0f3533",
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

func TestCloudAccountsCreateOSSOpenStackRejectsRemovedRootFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create", "--cloud-type", "openstack", "--storage-type", "object",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--username", "autotest",
		"--password", "autotest",
		"--user-domain-id", "default",
		"--only-verify",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --only-verify") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateBlockGenericProviderFallsBackToGenericBuilder(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block",
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

func TestCloudAccountsCreateBlockGenericProviderInfersAKSKFromAliasFlags(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block",
		"--access-id", "ak",
		"--access-secret", "sk",
		"--region-id", "cn-north-1",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "huawei_bs" || cloudAccount["cloud_auth_type"] != "aksk" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}

	metadata := cloudAccount["metadata"].(map[string]interface{})
	if metadata["access_id"] != "ak" || metadata["access_secret"] != "sk" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if _, ok := metadata["access_key_id"]; ok {
		t.Fatalf("metadata should not contain access_key_id: %+v", metadata)
	}
	if _, ok := metadata["access_key_secret"]; ok {
		t.Fatalf("metadata should not contain access_key_secret: %+v", metadata)
	}
}

func TestCloudAccountsCreateBlockRejectsRemovedRootFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create", "--cloud-type", "openstack", "--storage-type", "block",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--username", "autotest",
		"--password", "autotest",
		"--user-domain-id", "default",
		"--project-domain-id", "default",
		"--project-name", "autotest",
		"--region-name", "RegionOne",
		"--only-verify",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --only-verify") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateBlockRejectsLegacyCredentialFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create", "--cloud-type", "openstack", "--storage-type", "block",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--cloud-account-username", "autotest",
		"--cloud-account-password", "autotest",
		"--user-domain-id", "default",
		"--project-domain-id", "default",
		"--project-name", "autotest",
		"--region-name", "RegionOne",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --cloud-account-username") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateHuaweiBlockRejectsCloudAuthType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block",
		"--cloud-auth-type", "aksk",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "cloud-auth-type cannot be used") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateHuaweiBlockRejectsPasswordStyleFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block",
		"--auth-url", "https://iam.example.invalid/v3",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --auth-url") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateBlockGenericFileSetAndFlagOverrides(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(filePath, []byte(`{"project_id":"file-project","account_name":"file-name","nested":{"ssh_port":"22"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block",
		"--file", filePath,
		"--set-json", `nested={"ssh_port":"2200","ssh_pass":"json-pass"}`,
		"--set", "project_id=set-project",
		"--account-name", "flag-name",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-north-1",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["project_id"] != "set-project" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if metadata["account_name"] != "flag-name" {
		t.Fatalf("metadata = %+v", metadata)
	}
	nested := metadata["nested"].(map[string]interface{})
	if nested["ssh_port"] != "2200" || nested["ssh_pass"] != "json-pass" {
		t.Fatalf("nested = %+v", nested)
	}
}

func TestCloudAccountsCreateBlockOpenStackSupportsFileSetAndDynamicMetadata(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(filePath, []byte(`{"project_domain_id":"default","project_name":"file-project","region_name":"RegionOne"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "openstack", "--storage-type", "block",
		"--file", filePath,
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--username", "autotest",
		"--password", "0b33333d1f0f3533",
		"--user-domain-id", "default",
		"--project-domain-id", "flag-domain",
		"--region-name", "RegionOne",
		"--set", "project_name=set-project",
		"--custom-ssh-label", "ops",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["project_domain_id"] != "flag-domain" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if metadata["project_name"] != "set-project" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if metadata["custom_ssh_label"] != "ops" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestCloudAccountsCreateBlockFileRejectsWrapperObject(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	filePath := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(filePath, []byte(`{"cloud_account":{"metadata":{"region_id":"cn-north-1"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block",
		"--file", filePath,
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "file must contain metadata object, not cloud_account wrapper") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateOSSGenericProviderFallsBackToGenericBuilder(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--auth-url", "https://vc.example.invalid",
		"--username", "admin",
		"--password", "secret",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "huawei_obs" || cloudAccount["cloud_auth_type"] != "password" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
}

func TestCloudAccountsCreateGenericProviderMixedCredentialStylesRequireCloudAuthType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--username", "admin",
		"--password", "secret",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "multiple credential styles provided; pass --cloud-auth-type explicitly") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateGenericProviderExplicitCloudAuthTypeAllowsMixedCredentialStyles(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--cloud-auth-type", "password",
		"--auth-url", "https://vc.example.invalid",
		"--username", "admin",
		"--password", "secret",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}

	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_auth_type"] != "password" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if metadata["access_key_id"] != "ak" || metadata["access_key_secret"] != "sk" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestCloudAccountsCreateGenericProviderDirectCredentialStyleOverridesFileInference(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(filePath, []byte(`{"access_key_id":"file-ak","access_key_secret":"file-sk"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--file", filePath,
		"--auth-url", "https://vc.example.invalid",
		"--username", "admin",
		"--password", "secret",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_auth_type"] != "password" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
}

func TestCloudAccountsCreateGenericProvidersAllowFormerLegacyConfigFlagsAsMetadata(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantKey string
		want    interface{}
	}{
		{
			name: "block host",
			args: []string{
				"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block",
				"--access-key-id", "ak",
				"--access-key-secret", "sk",
				"--host", "https://legacy.invalid",
			},
			wantKey: "host",
			want:    "https://legacy.invalid",
		},
		{
			name: "oss scene",
			args: []string{
				"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
				"--cloud-auth-type", "password",
				"--auth-url", "https://vc.example.invalid",
				"--username", "admin",
				"--password", "secret",
				"--scene", "migration",
			},
			wantKey: "scene",
			want:    "migration",
		},
		{
			name: "oss insecure",
			args: []string{
				"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
				"--cloud-auth-type", "password",
				"--auth-url", "https://vc.example.invalid",
				"--username", "admin",
				"--password", "secret",
				"--insecure", "true",
			},
			wantKey: "insecure",
			want:    "true",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, body := executeCloudAccountCreateAtPath(t, tc.args)
			metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
			if metadata[tc.wantKey] != tc.want {
				t.Fatalf("args=%v metadata=%+v want %s=%v", tc.args, metadata, tc.wantKey, tc.want)
			}
		})
	}
}

func TestCloudAccountsCreateOSSHuaweiAcceptsDynamicMetadataFlags(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-north-4",
		"--project-domain-id", "domain-1",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if metadata["access_key_id"] != "ak" || metadata["project_domain_id"] != "domain-1" || metadata["region_id"] != "cn-north-4" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestCloudAccountsCreateOSSHuaweiAutoGeneratesCustomName(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-north-1",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}

	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["custom_name"] != "Huawei Cloud(Recommended, SDK v3.1.86)-cn-north-1" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestCloudAccountsCreateOSSOpenStackKeepsAccessAliasAsDynamicMetadata(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "openstack", "--storage-type", "object",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--username", "autotest",
		"--password", "0b33333d1f0f3533",
		"--user-domain-id", "default",
		"--access-id", "ak",
		"--access-secret", "sk",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_auth_type"] != "password" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if metadata["access_id"] != "ak" || metadata["access_secret"] != "sk" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestCloudAccountsCreateBlockGenericAliasValidationKeepsFieldErrors(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block",
		"--access-id", "ak",
		"--region-id", "cn-north-1",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "access-secret is required") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateOSSGenericFileSetAndFlagOverrides(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(filePath, []byte(`{"project_id":"file-project","custom_name":"file-name","nested":{"disk_bus_type_id":"file-bus"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--cloud-auth-type", "password",
		"--file", filePath,
		"--set-json", `nested={"disk_bus_type_id":"json-bus","disk_bus_type_name":"virtio"}`,
		"--set", "project_id=set-project",
		"--custom-name", "flag-name",
		"--auth-url", "https://vc.example.invalid",
		"--username", "admin",
		"--password", "secret",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["project_id"] != "set-project" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if metadata["custom_name"] != "flag-name" {
		t.Fatalf("metadata = %+v", metadata)
	}
	nested := metadata["nested"].(map[string]interface{})
	if nested["disk_bus_type_id"] != "json-bus" || nested["disk_bus_type_name"] != "virtio" {
		t.Fatalf("nested = %+v", nested)
	}
}

func TestCloudAccountsCreateOSSGenericMakeImageKeepsAutoUploadImagesEnabled(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--auth-url", "https://vc.example.invalid",
		"--username", "admin",
		"--password", "secret",
		"--set", "linux_boot_image_id=make_image",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	if body["auto_upload_images"] != float64(1) && body["auto_upload_images"] != 1 {
		t.Fatalf("body = %+v", body)
	}

	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["linux_boot_image_id"] != "make_image" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestCloudAccountsCreateOSSRejectsLegacyCredentialFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create", "--cloud-type", "openstack", "--storage-type", "object",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--cloud-account-username", "autotest",
		"--cloud-account-password", "autotest",
		"--user-domain-id", "default",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --cloud-account-username") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateOSSOpenStackSupportsFileSetAndDynamicMetadata(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(filePath, []byte(`{"project_domain_id":"default","project_id":"file-project","project_name":"file-name","region_id":"RegionOne","region_name":"RegionOne"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "openstack", "--storage-type", "object",
		"--file", filePath,
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--username", "autotest",
		"--password", "0b33333d1f0f3533",
		"--user-domain-id", "default",
		"--set", "project_id=set-project",
		"--boot-loader-flavor-id", "manual-flavor",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["project_id"] != "set-project" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if metadata["boot_loader_flavor_id"] != "manual-flavor" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestCloudAccountsCreateRequiresStorageType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create",
		"--cloud-type", "aliyun",
		"--access-key-id", "ak",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "storage-type is required") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateRequiresCloudType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create",
		"--storage-type", "block",
		"--access-key-id", "ak",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "cloud-type is required") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateRejectsInvalidPublicStorageType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []string{"archive", "block_storage", "object_storage"}
	for _, storageType := range cases {
		var out, errOut bytes.Buffer
		err := Execute(withHost(t, "https://example.invalid",
			"cloud-account", "create",
			"--cloud-type", "aliyun",
			"--storage-type", storageType,
		), &out, &errOut)
		if err == nil || !strings.Contains(err.Error(), "storage-type must be block or object") {
			t.Fatalf("storageType=%q err=%v", storageType, err)
		}
	}
}

func TestCloudAccountsCreateRejectsUnknownProfile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create",
		"--cloud-type", "vmware",
		"--storage-type", "block",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), `cloud-type "vmware" does not support storage-type block`) {
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
