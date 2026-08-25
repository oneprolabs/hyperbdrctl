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
	for _, unwanted := range []string{"aliyun_bs_block", "create block", "create oss", "--access-key-id", "--auth-url", "--file string", "--body string", "--preview-request"} {
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

func TestCloudAccountsCreateProviderPreviewRequestRemainsAvailable(t *testing.T) {
	var out, errOut bytes.Buffer
	err := Execute([]string{
		"--output", "json",
		"cloud-account", "create", "--cloud-type", "aliyun", "--storage-type", "block",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--auth-region-id", "cn-beijing",
		"--preview-request",
	}, &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "\"cloud_account\"") {
		t.Fatalf("preview output = %q", out.String())
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
		"--auth-region-id", "cn-qingdao",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "aliyun_bs" || cloudAccount["cloud_auth_type"] != "aksk" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if len(metadata) != 3 || metadata["access_key_id"] != "ak" || metadata["access_key_secret"] != "sk" || metadata["auth_region_id"] != "cn-qingdao" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestCloudAccountsCreateBlockAliyunRequiresAuthRegionID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create", "--cloud-type", "aliyun", "--storage-type", "block",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
	), &out, &errOut)
	if err == nil || err.Error() != "auth-region-id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateBlockAliyunRejectsLegacyRegionFlags(t *testing.T) {
	for _, flag := range []string{"region-id", "region-name"} {
		t.Run(flag, func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)

			var out, errOut bytes.Buffer
			err := Execute(withHost(t, "https://example.invalid",
				"cloud-account", "create", "--cloud-type", "aliyun", "--storage-type", "block",
				"--access-key-id", "ak",
				"--access-key-secret", "sk",
				"--auth-region-id", "cn-beijing",
				"--"+flag, "legacy-region",
			), &out, &errOut)
			if err == nil || !strings.Contains(err.Error(), "use --auth-region-id") {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestCloudAccountsCreateBlockAliyunHelpUsesAuthRegionID(t *testing.T) {
	for _, args := range [][]string{
		{"cloud-account", "create", "--cloud-type", "aliyun", "--storage-type", "block", "--help"},
		{"--lang", "zh_cn", "cloud-account", "create", "--cloud-type", "aliyun", "--storage-type", "block", "--help"},
	} {
		var out, errOut bytes.Buffer
		if err := Execute(args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", args, err)
		}

		text := out.String()
		if !strings.Contains(text, "--auth-region-id string") {
			t.Fatalf("args=%v help missing auth region flag: %q", args, text)
		}
		for _, unwanted := range []string{"--region-id string", "--region-name string"} {
			if strings.Contains(text, unwanted) {
				t.Fatalf("args=%v help should not include %q: %q", args, unwanted, text)
			}
		}
	}
}

func TestCloudAccountsCreateBlockHuaweiUsesAuthRegionAndProjectID(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--auth-region-id", "cn-north-4",
		"--auth-project-id", "project-1",
	})
	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["auth_region_id"] != "cn-north-4" || metadata["auth_project_id"] != "project-1" {
		t.Fatalf("metadata = %+v", metadata)
	}
	for _, key := range []string{"region_id", "region_name"} {
		if _, ok := metadata[key]; ok {
			t.Fatalf("metadata must not contain %s: %+v", key, metadata)
		}
	}
}

func TestCloudAccountsCreateBlockHuaweiRejectsLegacyRegionFlags(t *testing.T) {
	for _, flag := range []string{"region-id", "region-name"} {
		t.Run(flag, func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)

			var out, errOut bytes.Buffer
			err := Execute(withHost(t, "https://example.invalid",
				"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block",
				"--access-key-id", "ak",
				"--access-key-secret", "sk",
				"--auth-region-id", "cn-north-4",
				"--"+flag, "legacy-region",
			), &out, &errOut)
			if err == nil || !strings.Contains(err.Error(), "use --auth-region-id") {
				t.Fatalf("err = %v", err)
			}
		})
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
		"--preview-request",
		"cloud-account detail --id <account_id>",
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
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "cloud-resource fetch --cloud-type aliyun --storage-type block", "--set stringArray", "--set-json stringArray", "cloud-account wait --id <account_id>"},
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
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "--access-key-id string", "--access-key-secret string", "--auth-region-id string", "Optional dynamic parameters:", "--account-name <name>", "Cloud account name", "--auth-project-id <project_id>", "Subproject ID", "--set stringArray", "--set-json stringArray", "cloud-resource fetch --cloud-type huawei --storage-type block", "cloud-account wait --id <account_id>"},
			unwanted: []string{
				"\nExamples:\n",
				"\nNotes:\n",
				"\nWorkflow:\n",
				"\nRelated Commands:\n",
				"--only-verify",
				"--file string",
				"--account-name string",
				"--cloud-auth-type <aksk|password>",
				"--auth-url",
				"--username <username>",
				"--password <password>",
				"--access-id <ak>",
				"--access-secret <sk>",
				"--region-id string",
				"--region-name string",
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
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "Parameter Sources:", "OpenStack RC File", "cloud-resource fetch --cloud-type openstack --storage-type object", "--set stringArray", "--set-json stringArray"},
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
			want: []string{"Usage:", "\nFlags:\n", "Usage Notes:", "--access-key-id string", "--access-key-secret string", "--region-id string", "Resource Retrieval:", "--fetch-res regions", "Optional dynamic parameters:", "--custom-name <name>", "--set stringArray", "--set-json stringArray"},
			unwanted: []string{
				"\nExamples:\n",
				"\nNotes:\n",
				"\nWorkflow:\n",
				"\nRelated Commands:\n",
				"--auto-upload-images",
				"--only-verify",
				"--file string",
				"--cloud-auth-type",
				"--foo-bar <value>",
				"--access-id + --access-secret => aksk",
				"If both AK/SK-style and username/password-style flags are present",
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

func TestCloudAccountCreateHuaweiObjectDynamicParameterHelp(t *testing.T) {
	cases := []struct {
		name       string
		lang       string
		title      string
		regionText string
		nameText   string
	}{
		{
			name:       "english",
			lang:       "en",
			title:      "Optional dynamic parameters:",
			regionText: "Query by --region-id and fill it automatically; fall back to the region ID if no display name is returned.",
			nameText:   "When omitted, generate `Huawei Cloud(Recommended, SDK v3.1.86)-<region display name or region ID>` automatically.",
		},
		{
			name:       "chinese",
			lang:       "zh_cn",
			title:      "可按需补充以下动态参数：",
			regionText: "根据 --region-id 查询并回填；查询不到地域名称时使用区域 ID。",
			nameText:   "省略时按 `华为云(推荐使用，SDK v3.1.86)-<地域显示名称或区域 ID>` 自动生成。",
		},
	}

	forbiddenFlags := []string{
		"linux-boot-image-id",
		"windows-boot-image-id",
		"linux-uefi-boot-image-id",
		"windows-uefi-boot-image-id",
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if err := Execute([]string{
				"--lang", tt.lang,
				"cloud-account", "create",
				"--cloud-type", "huawei",
				"--storage-type", "object",
				"--help",
			}, &out, &errOut); err != nil {
				t.Fatal(err)
			}

			text := out.String()
			for _, want := range []string{
				tt.title,
				"--account-name <name>",
				"--region-name <region_name>",
				"--custom-name <name>",
				tt.regionText,
				tt.nameText,
				"--use-internal-ip <mode>",
				"--control-access-ip <ip>",
				"--linux-boot-image-host-config-zone-id <zone_id>",
				"--linux-boot-image-host-config-flavor-id <flavor_id>",
				"--linux-boot-image-host-config-network-id <network_id>",
				"--linux-boot-image-host-config-subnet-id <subnet_id>",
				"--linux-boot-image-host-config-image-id <image_id>",
				"--linux-boot-image-host-config-system-disk-type-id <type_id>",
				"--boot-loader-image-id <image_id>",
				"--boot-loader-flavor-id <flavor_id>",
				"boot_loader_images",
				"--fetch-res zones",
				"--network-id <network_id>",
				"--fetch-res subnets",
			} {
				if !strings.Contains(text, want) {
					t.Fatalf("help missing %q: %q", want, text)
				}
			}
			for _, flag := range []string{
				"account-name", "region-name", "custom-name", "auth-project-id", "use-internal-ip", "control-access-ip",
				"linux-boot-image-host-config-zone-id", "linux-boot-image-host-config-flavor-id", "linux-boot-image-host-config-network-id",
				"linux-boot-image-host-config-subnet-id", "linux-boot-image-host-config-image-id", "linux-boot-image-host-config-system-disk-type-id",
				"boot-loader-image-id", "boot-loader-image-name", "boot-loader-flavor-id",
			} {
				if strings.Contains(text, "--"+flag+" string") {
					t.Fatalf("dynamic parameter --%s must not be rendered in Flags: %q", flag, text)
				}
			}
			for _, flag := range forbiddenFlags {
				if strings.Contains(text, "--"+flag) {
					t.Fatalf("Huawei object help must not advertise advanced parameter --%s: %q", flag, text)
				}
			}
			for _, forbidden := range []string{
				"--control-access-ip-radio", "--linux-boot-image-host-config-bandwidth-id", "--linux-boot-image-host-config-bandwidth-name",
				"--linux-boot-image-host-config-zone-name", "--linux-boot-image-host-config-flavor-name",
				"--linux-boot-image-host-config-network-name", "--linux-boot-image-host-config-subnet-name",
				"--linux-boot-image-host-config-image-name", "--linux-boot-image-host-config-system-disk-type-name",
				"--boot-loader-image-name", "boot_loader_flavors",
			} {
				if strings.Contains(text, forbidden) {
					t.Fatalf("Huawei object help must not contain %q: %q", forbidden, text)
				}
			}
			if strings.Count(text, "--preview-request") != 2 {
				t.Fatalf("preview-request must appear once in Flags and once in Usage Notes: %q", text)
			}
			for _, removed := range []string{"Resource selection flow:", "资源选择流程："} {
				if strings.Contains(text, removed) {
					t.Fatalf("resource selection flow must not be rendered: %q", text)
				}
			}
			if strings.Count(text, tt.title) != 1 {
				t.Fatalf("dynamic parameter title must be rendered once: %q", text)
			}
			transitionTitle := "Huawei Cloud temporary transition-host image build:"
			if tt.lang == "zh_cn" {
				transitionTitle = "华为云临时过渡主机镜像构建："
			}
			assertContainsInOrder(t, text,
				"--account-name <name>",
				"--region-name <region_name>",
				"--custom-name <name>",
				"--use-internal-ip <mode>",
				"--boot-loader-image-id <image_id>",
				"--boot-loader-flavor-id <flavor_id>",
				transitionTitle,
				"--linux-boot-image-host-config-zone-id <zone_id>",
				"--preview-request",
			)

			dynamicStart := strings.Index(text, tt.title)
			dynamicEnd := strings.Index(text[dynamicStart:], "--preview-request")
			if dynamicEnd < 0 {
				t.Fatalf("preview guidance missing after dynamic parameter section: %q", text)
			}
			dynamicText := text[dynamicStart : dynamicStart+dynamicEnd]
			if strings.Contains(dynamicText, "\n\n\n    --") {
				t.Fatalf("dynamic parameters must not have multiple blank lines between entries: %q", text)
			}
		})
	}
}

func TestCloudAccountCreateAliyunObjectDynamicParameterHelp(t *testing.T) {
	flags := []string{
		"region-name",
		"custom-name",
		"use-internal-ip",
		"boot-loader-image-id",
		"boot-loader-image-name",
		"boot-loader-flavor-id",
		"linux-boot-image-id",
		"windows-boot-image-id",
		"linux-uefi-boot-image-id",
		"windows-uefi-boot-image-id",
	}

	cases := []struct {
		name string
		lang string
		want []string
	}{
		{
			name: "english",
			lang: "en",
			want: []string{
				"Optional dynamic parameters:",
				"--account-name <name>",
				"--region-name <region_name>",
				"Region display name",
				"--custom-name <name>",
				"Object-storage cloud-account display name",
				"--use-internal-ip <mode>",
				"Default: 0",
				"Allowed values: 0 / 1",
				"Value 0 uses public access",
				"Value 1 uses internal access",
				"--boot-loader-image-id <image_id>",
				"--boot-loader-flavor-id <flavor_id>",
				"--linux-boot-image-id <image_id>",
				"Default: auto_upload",
				"--windows-boot-image-id <image_id>",
				"--linux-uefi-boot-image-id <image_id>",
				"--windows-uefi-boot-image-id <image_id>",
				"boot_loader_images",
				"--flavor-vcpus 2",
				"--flavor-ram 4",
				"--os-type linux",
				"--os-type windows",
				"--fetch-res images",
			},
		},
		{
			name: "chinese",
			lang: "zh_cn",
			want: []string{
				"可按需补充以下动态参数：",
				"--account-name <name>",
				"--region-name <region_name>",
				"地域显示名称",
				"--custom-name <name>",
				"对象存储云账号显示名称",
				"--use-internal-ip <mode>",
				"默认值：0",
				"可选值：0 / 1",
				"值 0 使用公网",
				"值 1 使用内网",
				"--boot-loader-image-id <image_id>",
				"--boot-loader-flavor-id <flavor_id>",
				"--linux-boot-image-id <image_id>",
				"默认值：auto_upload",
				"--windows-boot-image-id <image_id>",
				"--linux-uefi-boot-image-id <image_id>",
				"--windows-uefi-boot-image-id <image_id>",
				"boot_loader_images",
				"--flavor-vcpus 2",
				"--flavor-ram 4",
				"--os-type linux",
				"--os-type windows",
				"--fetch-res images",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			err := Execute([]string{
				"--lang", tt.lang,
				"cloud-account", "create",
				"--cloud-type", "aliyun",
				"--storage-type", "object",
				"--help",
			}, &out, &errOut)
			if err != nil {
				t.Fatal(err)
			}

			text := out.String()
			for _, want := range tt.want {
				if !strings.Contains(text, want) {
					t.Fatalf("help missing %q: %q", want, text)
				}
			}
			for _, flag := range flags {
				if strings.Contains(text, "--"+flag+" string") {
					t.Fatalf("dynamic parameter --%s must not be rendered in Flags: %q", flag, text)
				}
			}
			if strings.Count(text, tt.want[0]) != 1 {
				t.Fatalf("dynamic parameter title must be rendered once: %q", text)
			}
			dynamicStart := strings.Index(text, tt.want[0])
			dynamicEnd := strings.Index(text[dynamicStart:], "--preview-request")
			if dynamicEnd < 0 {
				t.Fatalf("preview guidance missing after dynamic parameter section: %q", text)
			}
			dynamicText := text[dynamicStart : dynamicStart+dynamicEnd]
			if got := strings.Count(dynamicText, "\n\n    --"); got != len(flags) {
				t.Fatalf("dynamic parameters should have exactly one blank line between entries: separators=%d help=%q", got, text)
			}
			if strings.Contains(dynamicText, "\n\n\n    --") {
				t.Fatalf("dynamic parameters must not have multiple blank lines between entries: %q", text)
			}
			assertContainsInOrder(t, text,
				"--account-name <name>",
				"--region-name <region_name>",
				"--custom-name <name>",
				"--use-internal-ip <mode>",
				"--boot-loader-image-id <image_id>",
				"--boot-loader-image-name <image_name>",
				"--boot-loader-flavor-id <flavor_id>",
				"--linux-boot-image-id <image_id>",
				"--windows-boot-image-id <image_id>",
				"--linux-uefi-boot-image-id <image_id>",
				"--windows-uefi-boot-image-id <image_id>",
				"--preview-request",
			)
		})
	}

	var out, errOut bytes.Buffer
	if err := Execute([]string{
		"--lang", "en",
		"cloud-account", "create",
		"--cloud-type", "openstack",
		"--storage-type", "object",
		"--help",
	}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if text := out.String(); strings.Contains(text, "--linux-uefi-boot-image-id <image_id>") {
		t.Fatalf("OpenStack object help must not include Aliyun-only UEFI parameter: %q", text)
	}
}

func TestCloudAccountsCreateHelpDynamicallyAddsSupplementalFlags(t *testing.T) {
	profiles := [][]string{
		{"aliyun", "block"},
		{"huawei", "block"},
		{"openstack", "block"},
		{"aliyun", "object"},
		{"huawei", "object"},
		{"openstack", "object"},
	}
	labels := map[string]struct {
		overrides string
		order     string
		preview   string
		next      string
	}{
		"en": {
			overrides: "To add more fields, repeat either flag as needed:",
			order:     "--set-json < --set < explicit create flags",
			preview:   "To inspect the final request body first, add:",
			next:      "After creation succeeds",
		},
		"zh_cn": {
			overrides: "如需更多字段，可重复附加：",
			order:     "--set-json < --set < 显式创建参数",
			preview:   "如需先检查最终请求体，可附加下面的参数：",
			next:      "创建成功后",
		},
	}

	for lang, label := range labels {
		for _, profile := range profiles {
			name := lang + "/" + profile[0] + "/" + profile[1]
			t.Run(name, func(t *testing.T) {
				var out, errOut bytes.Buffer
				args := []string{"--lang", lang, "cloud-account", "create", "--cloud-type", profile[0], "--storage-type", profile[1], "--help"}
				if err := Execute(args, &out, &errOut); err != nil {
					t.Fatal(err)
				}

				text := out.String()
				for _, want := range []string{label.overrides, label.order, label.preview} {
					if strings.Count(text, want) != 1 {
						t.Fatalf("help should contain %q exactly once: %q", want, text)
					}
				}
				overrideIndex := strings.Index(text, label.overrides)
				previewIndex := strings.Index(text, label.preview)
				nextIndex := strings.Index(text, label.next)
				if nextIndex < 0 || overrideIndex <= nextIndex || previewIndex <= overrideIndex {
					t.Fatalf("supplemental help order is invalid: %q", text)
				}
				if strings.Contains(text, "Optional dynamic parameters:") || strings.Contains(text, "可按需补充以下动态参数：") {
					dynamicIndex := strings.Index(text, "Optional dynamic parameters:")
					if dynamicIndex < 0 {
						dynamicIndex = strings.Index(text, "可按需补充以下动态参数：")
					}
					if dynamicIndex <= overrideIndex || dynamicIndex >= previewIndex {
						t.Fatalf("dynamic parameter help must appear before preview: %q", text)
					}
				}
			})
		}
	}
}

func TestCloudAccountsCreateSelectorHelpOmitsPreviewRequest(t *testing.T) {
	for _, args := range [][]string{
		{"cloud-account", "create", "--help"},
		{"cloud-account", "create", "--storage-type", "block", "--help"},
		{"--lang", "zh_cn", "cloud-account", "create", "--help"},
		{"--lang", "zh_cn", "cloud-account", "create", "--storage-type", "object", "--help"},
	} {
		var out, errOut bytes.Buffer
		if err := Execute(args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", args, err)
		}
		if strings.Contains(out.String(), "--preview-request") {
			t.Fatalf("selector help should not include preview-request: %q", out.String())
		}
	}
}

func TestCloudAccountCreateAtomyDynamicParameterHelp(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	atomyCases := []struct {
		name        string
		cloudType   string
		storageType string
	}{
		{name: "aliyun block", cloudType: "aliyun", storageType: "block"},
		{name: "huawei block", cloudType: "huawei", storageType: "block"},
		{name: "aliyun object", cloudType: "aliyun", storageType: "object"},
		{name: "huawei object", cloudType: "huawei", storageType: "object"},
	}
	for _, tt := range atomyCases {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			err := Execute([]string{
				"--lang", "en", "cloud-account", "create",
				"--cloud-type", tt.cloudType,
				"--storage-type", tt.storageType,
				"--help",
			}, &out, &errOut)
			if err != nil {
				t.Fatal(err)
			}

			text := out.String()
			for _, want := range []string{
				"Optional dynamic parameters:",
				"--account-name <name>",
				"Cloud account name",
			} {
				if !strings.Contains(text, want) {
					t.Fatalf("help missing %q: %q", want, text)
				}
			}
			if strings.Contains(text, "--account-name string") {
				t.Fatalf("dynamic parameter must not be rendered in Flags: %q", text)
			}
			if strings.Contains(text, "metadata.account_name") {
				t.Fatalf("dynamic parameter description must not expose metadata path: %q", text)
			}
			if strings.Count(text, "--account-name <name>") != 1 {
				t.Fatalf("dynamic parameter should be rendered once: %q", text)
			}
		})
	}

	t.Run("localized", func(t *testing.T) {
		var out, errOut bytes.Buffer
		err := Execute([]string{
			"--lang", "zh_cn", "cloud-account", "create",
			"--cloud-type", "huawei",
			"--storage-type", "block",
			"--help",
		}, &out, &errOut)
		if err != nil {
			t.Fatal(err)
		}
		text := out.String()
		for _, want := range []string{
			"可按需补充以下动态参数：",
			"--account-name <name>",
			"云账号名称",
			"--auth-project-id <project_id>",
			"子项目 ID",
		} {
			if !strings.Contains(text, want) {
				t.Fatalf("help missing %q: %q", want, text)
			}
		}
		if strings.Contains(text, "--account-name string") || strings.Count(text, "--account-name <name>") != 1 {
			t.Fatalf("localized dynamic parameter placement is invalid: %q", text)
		}
		if strings.Contains(text, "metadata.account_name") {
			t.Fatalf("localized dynamic parameter description must not expose metadata path: %q", text)
		}
	})

	for _, storageType := range []string{"block", "object"} {
		t.Run("not atomy "+storageType, func(t *testing.T) {
			var out, errOut bytes.Buffer
			err := Execute([]string{
				"--lang", "en", "cloud-account", "create",
				"--cloud-type", "openstack",
				"--storage-type", storageType,
				"--help",
			}, &out, &errOut)
			if err != nil {
				t.Fatal(err)
			}
			if text := out.String(); strings.Contains(text, "--account-name <name>") || strings.Contains(text, "--account-name string") {
				t.Fatalf("non-Atomy help must not include the Atomy account-name parameter: %q", text)
			}
		})
	}
}

func TestCloudAccountCreateOpenStackBlockImageAccessDynamicParameterHelp(t *testing.T) {
	cases := []struct {
		name     string
		language string
		want     []string
	}{
		{
			name:     "english",
			language: "en",
			want: []string{
				"Optional dynamic parameters:",
				"--ssh-port <port>", "Cloud-sync-gateway image SSH port", "Default: 22",
				"--ssh-pass <password>", "Cloud-sync-gateway image SSH root password",
				"--linux-hd-username <username>", "Transition host image username",
				"--linux-hd-password <password>", "Transition host image password",
				"--linux-hd-port <port>", "Transition host image communication port", "Default: 10729",
			},
		},
		{
			name:     "chinese",
			language: "zh_cn",
			want: []string{
				"可按需补充以下动态参数：",
				"--ssh-port <port>", "云同步网关镜像 SSH 端口", "默认值：22",
				"--ssh-pass <password>", "云同步网关镜像 SSH root 密码",
				"--linux-hd-username <username>", "过渡主机镜像用户名",
				"--linux-hd-password <password>", "过渡主机镜像密码",
				"--linux-hd-port <port>", "过渡主机镜像通讯端口", "默认值：10729",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			err := Execute([]string{
				"--lang", tt.language, "cloud-account", "create",
				"--cloud-type", "openstack",
				"--storage-type", "block",
				"--help",
			}, &out, &errOut)
			if err != nil {
				t.Fatal(err)
			}

			text := out.String()
			for _, want := range tt.want {
				if !strings.Contains(text, want) {
					t.Fatalf("help missing %q: %q", want, text)
				}
			}
			for _, flag := range []string{"ssh-port", "ssh-pass", "linux-hd-username", "linux-hd-password", "linux-hd-port"} {
				if strings.Contains(text, "--"+flag+" string") {
					t.Fatalf("dynamic parameter --%s must not be rendered in Flags: %q", flag, text)
				}
			}
			if strings.Count(text, tt.want[0]) != 1 {
				t.Fatalf("dynamic parameter title must be rendered once: %q", text)
			}
			assertContainsInOrder(t, text,
				"--ssh-port <port>",
				"--ssh-pass <password>",
				"--linux-hd-username <username>",
				"--linux-hd-password <password>",
				"--linux-hd-port <port>",
				"--preview-request",
			)
		})
	}
}

func TestDynamicParameterHelpAttachmentMatchesAllSelectors(t *testing.T) {
	attachment := dynamicParameterHelpAttachment{
		Command:      "cloud-account-create",
		Provider:     "huawei",
		CloudType:    "huawei_bs",
		Architecture: "AtomyV2",
		StorageType:  "block",
	}
	match := dynamicParameterHelpContext{
		Command:      "CLOUD-ACCOUNT-CREATE",
		Provider:     "Huawei",
		CloudType:    "HUAWEI_BS",
		Architecture: "atomyv2",
		StorageType:  "BLOCK",
	}
	if !dynamicParameterHelpAttachmentMatches(attachment, match) {
		t.Fatal("attachment should match every selector case-insensitively")
	}

	for name, context := range map[string]dynamicParameterHelpContext{
		"command":      {Command: "cloud-sync-gateway-create", Provider: "huawei", CloudType: "huawei_bs", Architecture: "AtomyV2", StorageType: "block"},
		"provider":     {Command: "cloud-account-create", Provider: "aliyun", CloudType: "huawei_bs", Architecture: "AtomyV2", StorageType: "block"},
		"cloud type":   {Command: "cloud-account-create", Provider: "huawei", CloudType: "huawei_obs", Architecture: "AtomyV2", StorageType: "block"},
		"architecture": {Command: "cloud-account-create", Provider: "huawei", CloudType: "huawei_bs", Architecture: "NotAtomy", StorageType: "block"},
		"storage type": {Command: "cloud-account-create", Provider: "huawei", CloudType: "huawei_bs", Architecture: "AtomyV2", StorageType: "objectstorage"},
	} {
		t.Run(name, func(t *testing.T) {
			if dynamicParameterHelpAttachmentMatches(attachment, context) {
				t.Fatalf("attachment unexpectedly matched context: %+v", context)
			}
		})
	}

	if !dynamicParameterHelpAttachmentMatches(dynamicParameterHelpAttachment{Command: "cloud-account-create"}, dynamicParameterHelpContext{Command: "cloud-account-create"}) {
		t.Fatal("empty selectors should act as wildcards")
	}
}

func TestCloudAccountArchivedHelpCopyIsLocalized(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		name   string
		path   []string
		zh     []string
		en     []string
		omitZh []string
		omitEn []string
	}{
		{
			name: "root",
			zh:   []string{"云账号管理", "list                     列出云账号", "查看已有云账号", "创建后等待任务完成"},
			en:   []string{"Cloud account management", "list                     List cloud accounts", "View existing cloud accounts", "Wait for the task after creation"},
		},
		{
			name: "list",
			path: []string{"list"},
			zh:   []string{"列出云账号", "分页列出云账号", "如未传入 --storage-type", "--storage-type object"},
			en:   []string{"List cloud accounts", "List cloud accounts with pagination", "When --storage-type is omitted", "--storage-type object"},
		},
		{
			name: "detail",
			path: []string{"detail"},
			zh:   []string{"查看云账号详情", "账号 ID 通常来自", "cloud-account detail --id <account_id>"},
			en:   []string{"Show cloud account detail", "The account ID usually comes from", "cloud-account detail --id <account_id>"},
		},
		{
			name: "wait",
			path: []string{"wait"},
			zh:   []string{"等待云账号创建完成", "--interval-seconds 15", "--timeout-seconds 1800"},
			en:   []string{"Wait for cloud account creation", "--interval-seconds 15", "--timeout-seconds 1800"},
		},
		{
			name:   "delete",
			path:   []string{"delete"},
			zh:     []string{"删除云账号", "最小删除命令如下", "未被其它配置引用"},
			en:     []string{"Delete cloud account", "The minimum delete command is", "not referenced by other configurations"},
			omitZh: []string{"默认值 false"},
			omitEn: []string{"default false"},
		},
		{
			name:   "create guide",
			path:   []string{"create"},
			zh:     []string{"创建云账号", "统一的云账号创建引导入口", "具体创建参数请进入对应云厂商帮助页查看"},
			en:     []string{"Create cloud account", "unified cloud account creation guide", "provider-specific help page"},
			omitZh: []string{"--preview-request"},
			omitEn: []string{"--preview-request"},
		},
		{
			name: "block guide",
			path: []string{"create", "--storage-type", "block"},
			zh:   []string{"创建块存储云账号", "先从下面选择云厂商", "--cloud-type huawei --storage-type block", "云厂商:"},
			en:   []string{"Create block-storage cloud account", "Choose a cloud provider below", "--cloud-type huawei --storage-type block", "Providers:"},
		},
		{
			name: "object guide",
			path: []string{"create", "--storage-type", "object"},
			zh:   []string{"创建对象存储云账号", "先从下面选择云厂商", "--cloud-type huawei --storage-type object", "云厂商:"},
			en:   []string{"Create object-storage cloud account", "Choose a cloud provider below", "--cloud-type huawei --storage-type object", "Providers:"},
		},
		{
			name: "aliyun block",
			path: []string{"create", "--cloud-type", "aliyun", "--storage-type", "block"},
			zh:   []string{"创建阿里云块存储账号", "参数来源：", "资源获取：", "--set-json < --set < 显式创建参数"},
			en:   []string{"Create Alibaba Cloud block-storage account", "Parameter Sources:", "Resource Retrieval:", "--set-json < --set < explicit create flags"},
		},
		{
			name: "huawei block",
			path: []string{"create", "--cloud-type", "huawei", "--storage-type", "block"},
			zh:   []string{"创建华为云块存储账号", "认证地域 ID（必须）", "cloud-resource fetch --cloud-type huawei --storage-type block", "--auth-project-id <project_id>"},
			en:   []string{"Create Huawei Cloud block-storage account", "Authentication region ID (required)", "cloud-resource fetch --cloud-type huawei --storage-type block", "--auth-project-id <project_id>"},
		},
		{
			name: "openstack block",
			path: []string{"create", "--cloud-type", "openstack", "--storage-type", "block"},
			zh:   []string{"创建 OpenStack 块存储账号", "OpenStack RC 文件", "创建本身不依赖前置资源查询", "--linux-hd-port <port>"},
			en:   []string{"Create OpenStack block-storage account", "OpenStack RC File", "does not require a resource query first", "--linux-hd-port <port>"},
		},
		{
			name: "aliyun object",
			path: []string{"create", "--cloud-type", "aliyun", "--storage-type", "object"},
			zh:   []string{"创建阿里云对象存储账号", "控制台访问方式", "默认值：0", "可选值：", "boot_loader_images", "--fetch-res images"},
			en:   []string{"Create Alibaba Cloud object-storage account", "Console access mode", "Default: 0", "Allowed values:", "boot_loader_images", "--fetch-res images"},
		},
		{
			name: "huawei object",
			path: []string{"create", "--cloud-type", "huawei", "--storage-type", "object"},
			zh:   []string{"创建华为云对象存储账号", "区域 ID（必须）", "资源获取：", "--fetch-res regions", "华为云临时过渡主机镜像构建："},
			en:   []string{"Create Huawei Cloud object-storage account", "Region ID (required)", "Resource Retrieval:", "--fetch-res regions", "Huawei Cloud temporary transition-host image build:"},
		},
		{
			name: "openstack object",
			path: []string{"create", "--cloud-type", "openstack", "--storage-type", "object"},
			zh:   []string{"创建 OpenStack 对象存储账号", "控制台访问方式", "重点关注返回结果中的", "disk_bus_type_name"},
			en:   []string{"Create OpenStack object-storage account", "Console access method", "Pay particular attention to these fields", "disk_bus_type_name"},
		},
	}

	for _, tt := range cases {
		for _, lang := range []string{"zh_cn", "en"} {
			t.Run(tt.name+"/"+lang, func(t *testing.T) {
				args := append([]string{"--lang", lang, "cloud-account"}, tt.path...)
				args = append(args, "--help")
				var out, errOut bytes.Buffer
				if err := Execute(args, &out, &errOut); err != nil {
					t.Fatalf("args=%v err=%v", args, err)
				}

				want, omit := tt.en, tt.omitEn
				if lang == "zh_cn" {
					want, omit = tt.zh, tt.omitZh
				}
				text := out.String()
				for _, value := range want {
					if !strings.Contains(text, value) {
						t.Fatalf("args=%v help missing %q: %q", args, value, text)
					}
				}
				for _, value := range omit {
					if strings.Contains(text, value) {
						t.Fatalf("args=%v help should not contain %q: %q", args, value, text)
					}
				}
				assertNoHelpFooter(t, text)
			})
		}
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
		"--auth-region-id", "cn-north-1",
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
		"--auth-region-id", "cn-north-1",
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

func TestCloudAccountsCreateHuaweiBlockRequiresAuthRegionID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "block",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
	), &out, &errOut)
	if err == nil || err.Error() != "auth-region-id is required" {
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
		"--auth-region-id", "cn-north-1",
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

func TestCloudAccountsCreateOSSHuaweiUsesDedicatedAKSKBuilder(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-north-1",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "huawei_obs" || cloudAccount["cloud_auth_type"] != "aksk" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if metadata["access_id"] != "ak" || metadata["access_secret"] != "sk" {
		t.Fatalf("metadata = %+v", metadata)
	}
	for _, forbidden := range []string{"access_key_id", "access_key_secret", "boot_image_source", "skip_driver_fix", "windows_boot_image_id", "linux_uefi_boot_image_id", "windows_uefi_boot_image_id", "upload_uefi_image", "region_type", "region_type_list", "control_access_ip_radio", "control_access_ip", "boot_loader_flavor_id"} {
		if _, ok := metadata[forbidden]; ok {
			t.Fatalf("metadata must not contain %s: %+v", forbidden, metadata)
		}
	}
	host := metadata["linux_boot_image_host_config"].(map[string]interface{})
	for _, forbidden := range []string{"bandwidth_id", "bandwidth_name"} {
		if _, ok := host[forbidden]; ok {
			t.Fatalf("host config must not contain %s: %+v", forbidden, host)
		}
	}
}

func TestCloudAccountsCreateOSSHuaweiSubmitsExplicitControlIPAndBootLoaderFlavor(t *testing.T) {
	_, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-north-1",
		"--control-access-ip", "2001:db8::10",
		"--boot-loader-flavor-id", "manual-flavor",
	})

	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["control_access_ip"] != "2001:db8::10" || metadata["boot_loader_flavor_id"] != "manual-flavor" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestCloudAccountsCreateOSSHuaweiRejectsRemovedFlags(t *testing.T) {
	for _, flag := range []string{
		"control-access-ip-radio",
		"linux-boot-image-host-config-bandwidth-id",
		"linux-boot-image-host-config-bandwidth-name",
	} {
		t.Run(flag, func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)
			var out, errOut bytes.Buffer
			err := Execute(withHost(t, "https://example.invalid",
				"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
				"--access-key-id", "ak", "--access-key-secret", "sk", "--region-id", "cn-north-1",
				"--"+flag, "value",
			), &out, &errOut)
			if err == nil || !strings.Contains(err.Error(), "unknown flag: --"+flag) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestCloudAccountsCreateOSSHuaweiRequiresRegionAndValidControlIP(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing region",
			args: []string{"--access-key-id", "ak", "--access-key-secret", "sk"},
			want: "region-id is required",
		},
		{
			name: "invalid control IP",
			args: []string{"--access-key-id", "ak", "--access-key-secret", "sk", "--region-id", "cn-north-1", "--control-access-ip", "not-an-ip"},
			want: "control-access-ip must be a valid IP address",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)
			args := []string{"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object"}
			args = append(args, tc.args...)
			var out, errOut bytes.Buffer
			err := Execute(withHost(t, "https://example.invalid", args...), &out, &errOut)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v", err)
			}
		})
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
				"--auth-region-id", "cn-north-1",
				"--host", "https://legacy.invalid",
			},
			wantKey: "host",
			want:    "https://legacy.invalid",
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
		"--operator-label", "ops",
	})

	if path != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", path)
	}
	cloudAccount := body["cloud_account"].(map[string]interface{})
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if metadata["access_id"] != "ak" || metadata["operator_label"] != "ops" || metadata["region_id"] != "cn-north-4" || metadata["region_name"] != "North China - Beijing 4" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestCloudAccountsCreateOSSHuaweiFallsBackToSpecifiedRegionIDForName(t *testing.T) {
	_, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-test-empty",
	})

	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["region_id"] != "cn-test-empty" || metadata["region_name"] != "cn-test-empty" {
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
	if metadata["custom_name"] != "Huawei Cloud(Recommended, SDK v3.1.86)-North China - Beijing 1" {
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
		"--auth-region-id", "cn-north-1",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "access-secret is required") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsCreateOSSHuaweiFileSetAndFlagOverrides(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(filePath, []byte(`{"project_id":"file-project","custom_name":"file-name","nested":{"disk_bus_type_id":"file-bus"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--file", filePath,
		"--set-json", `nested={"disk_bus_type_id":"json-bus","disk_bus_type_name":"virtio"}`,
		"--set", "project_id=set-project",
		"--custom-name", "flag-name",
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
	if metadata["custom_name"] != "flag-name" {
		t.Fatalf("metadata = %+v", metadata)
	}
	nested := metadata["nested"].(map[string]interface{})
	if nested["disk_bus_type_id"] != "json-bus" || nested["disk_bus_type_name"] != "virtio" {
		t.Fatalf("nested = %+v", nested)
	}
}

func TestCloudAccountsCreateOSSHuaweiAlwaysMakesSingleLinuxImage(t *testing.T) {
	path, body := executeCloudAccountCreateAtPath(t, []string{
		"cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-north-1",
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
	host := metadata["linux_boot_image_host_config"].(map[string]interface{})
	if host["zone_id"] != "cn-north-1a" || host["flavor_id"] != "c3.large.2" || host["network_id"] != "network-1" || host["subnet_id"] != "subnet-1" || host["image_id"] != "image-linux-1" || host["system_disk_type_id"] != "disk-type-1" {
		t.Fatalf("host = %+v", host)
	}
	flavorIDs := host["flavor_id_arr"].([]interface{})
	if len(flavorIDs) != 3 || flavorIDs[0] != "u-2" || flavorIDs[1] != "u-2-m-4" || flavorIDs[2] != "c3.large.2" {
		t.Fatalf("flavor_id_arr = %+v", flavorIDs)
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
			cloudType := ""
			if cloudAccount, ok := requestBody["cloud_account"].(map[string]interface{}); ok {
				cloudType, _ = cloudAccount["cloud_type"].(string)
			}
			cloudInfo := map[string]interface{}{}
			if cloudType == "huawei_obs" {
				switch fetchRes {
				case "regions,zones":
					cloudInfo["regions"] = []map[string]interface{}{
						{"id": "cn-north-1", "display_name": "North China - Beijing 1", "local_name": "华北-北京一"},
						{"id": "cn-north-4", "display_name": "North China - Beijing 4", "local_name": "华北-北京四"},
						{"id": "cn-test-empty"},
					}
					cloudInfo["zones"] = []map[string]interface{}{
						{"id": "cn-north-1a", "name": "Availability Zone 1", "local_name": "可用区1"},
						{"id": "cn-north-4a", "name": "Availability Zone 1", "local_name": "可用区1"},
					}
				case "networks,subnets":
					cloudInfo["networks"] = []map[string]interface{}{{"id": "network-1", "name": "vpc-ray", "is_recommend": 1}}
					cloudInfo["subnets"] = []map[string]interface{}{{"id": "subnet-1", "name": "subnet-ray", "network_id": "network-1"}}
				case "flavors":
					cloudInfo["flavors"] = []map[string]interface{}{{
						"value": "u-2", "children": []interface{}{map[string]interface{}{
							"value": "u-2-m-4", "children": []interface{}{map[string]interface{}{
								"id": "c3.large.2", "name": "c3.large.2", "vcpus": 2, "ram_GB": 4, "is_recommend": 1,
							}},
						}},
					}}
				case "images,system_volume_types":
					cloudInfo["images"] = []map[string]interface{}{{"id": "image-linux-1", "name": "Ubuntu 24.04 server 64bit", "os_type": "linux", "is_recommend": 1}}
					cloudInfo["system_volume_types"] = []map[string]interface{}{{"id": "disk-type-1", "name": "General Purpose SSD", "is_recommend": 1}}
				case "boot_loader_images":
					cloudInfo["boot_loader_images"] = []map[string]interface{}{{"id": "boot-image-1", "name": "Windows Server 2016 Standard 64bit English", "is_recommend": 1}}
				}
				if strings.Contains(fetchRes, "boot_loader_flavors") || strings.Contains(fetchRes, "bandwidths") {
					t.Errorf("unexpected Huawei object resource query %q", fetchRes)
				}
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": "00000000",
					"data": map[string]interface{}{"cloud_info": cloudInfo},
				})
				return
			}
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
