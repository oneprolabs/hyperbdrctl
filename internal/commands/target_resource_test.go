package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func cloudResourceHelpFlags(text string) string {
	start := strings.Index(text, "\nFlags:\n")
	if start < 0 {
		start = strings.Index(text, "\n参数:\n")
	}
	end := strings.Index(text, "\nUsage Notes:\n")
	if end < 0 {
		end = strings.Index(text, "\n使用说明:\n")
	}
	if start < 0 || end < 0 || end <= start {
		return ""
	}
	return text[start:end]
}

func TestCloudResourceHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-resource", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"Usage:",
		"\nFlags:\n",
		"\nCommands:\n",
		"catalog",
		"fetch",
		"Usage Notes:",
		"hyperbdrctl cloud-resource catalog",
		"--cloud-account-id <account_id>",
		"--cloud-type aliyun",
		"--storage-type block",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	for _, unwanted := range []string{"\n  block", "\n  oss"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("help should not include %q: %q", unwanted, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestCloudResourceStorageHelpShowsProviders(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := [][]string{
		{"cloud-resource", "fetch", "--storage-type", "block", "--help"},
		{"cloud-resource", "fetch", "--storage-type", "object", "--help"},
	}

	for _, args := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", args, err)
		}
		text := out.String()
		for _, want := range []string{"--cloud-type", "--storage-type", "aliyun", "openstack", "Usage Notes:"} {
			if !strings.Contains(text, want) {
				t.Fatalf("args=%v missing %q: %q", args, want, text)
			}
		}
		flags := cloudResourceHelpFlags(text)
		for _, unwanted := range []string{"--cloud-account-id", "--fetch-res", "--access-key-id", "--auth-url", "--region-id"} {
			if strings.Contains(flags, unwanted) {
				t.Fatalf("args=%v flags should not contain %q: %q", args, unwanted, flags)
			}
		}
		if !strings.Contains(flags, "Storage type (required)") {
			t.Fatalf("args=%v storage-type should be required: %q", args, flags)
		}
		assertNoHelpFooter(t, text)
	}
}

func TestCloudResourceFetchHelpShowsGenericModes(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-resource", "fetch", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"--cloud-account-id",
		"--cloud-type",
		"--storage-type",
		"Usage Notes:",
		"With an existing cloud account",
		"--storage-type block --help",
		"--storage-type object --help",
		"Then view the direct-query flags for a provider",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	flags := cloudResourceHelpFlags(text)
	for _, unwanted := range []string{"--cloud-auth-type", "--fetch-res", "--region-id", "--flavor-id", "--auth-url", "--username", "--password"} {
		if strings.Contains(flags, unwanted) {
			t.Fatalf("generic flags should not contain %q: %q", unwanted, flags)
		}
	}
}

func TestCloudResourceDirectHelpShowsProviderProfile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args     []string
		want     []string
		unwanted []string
	}{
		{
			args: []string{"cloud-resource", "fetch", "--cloud-type", "aliyun", "--storage-type", "block", "--help"},
			want: []string{
				"Usage: hyperbdrctl cloud-resource fetch --cloud-type aliyun --storage-type block [flags]",
				"Fetch read-only Alibaba Cloud block-storage resources.",
				"--access-key-id <ak>",
				"system_disk_types",
				"cloud-sync-gateway create --cloud-account-id <account_id> --help",
			},
			unwanted: []string{"--cloud-auth-type string"},
		},
		{
			args: []string{"--lang", "zh_cn", "cloud-resource", "fetch", "--cloud-type", "huawei", "--storage-type", "block", "--help"},
			want: []string{
				"用法: hyperbdrctl cloud-resource fetch --cloud-type huawei --storage-type block [参数]",
				"获取华为云块存储只读资源。",
				"--access-key-secret <sk>",
				"system_disk_types",
			},
			unwanted: []string{"--cloud-auth-type string"},
		},
		{
			args: []string{"--lang", "zh_cn", "cloud-resource", "fetch", "--cloud-type", "openstack", "--storage-type", "block", "--help"},
			want: []string{
				"获取 OpenStack 块存储只读资源。",
				"--auth-url string",
				"云平台、源端或对象存储鉴权地址（必须）",
				"regions,compute_zones,projects",
				"cloud-sync-gateway create --cloud-account-id <account_id> --help",
			},
		},
		{
			args: []string{"cloud-resource", "fetch", "--cloud-type", "openstack", "--storage-type", "object", "--help"},
			want: []string{
				"Fetch read-only OpenStack object-storage resources.",
				"--auth-url",
				"--username",
				"--password",
				"--user-domain-id",
				"volume_types,subnets,security_groups",
				"cloud-account create --cloud-type openstack --storage-type object --help",
			},
		},
		{
			args: []string{"cloud-resource", "fetch", "--cloud-type", "aliyun", "--storage-type", "object", "--help"},
			want: []string{
				"Fetch read-only Alibaba Cloud object-storage resources.",
				"system_volume_types",
				"volume_types",
				"cloud-account create --cloud-type aliyun --storage-type object --help",
			},
			unwanted: []string{"cloud-sync-gateway create", "--cloud-auth-type string"},
		},
		{
			args: []string{"--lang", "zh_cn", "cloud-resource", "fetch", "--cloud-type", "huawei", "--storage-type", "object", "--help"},
			want: []string{
				"获取华为云对象存储只读资源。",
				"--os-type string",
				"--image-type string",
				"--purpose string",
				"system_volume_types",
				"cloud-account create --cloud-type huawei --storage-type object --help",
			},
			unwanted: []string{"云同步网关帮助", "--cloud-auth-type string"},
		},
	}
	for _, tt := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(tt.args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", tt.args, err)
		}
		text := out.String()
		for _, want := range tt.want {
			if !strings.Contains(text, want) {
				t.Fatalf("args=%v missing %q: %q", tt.args, want, text)
			}
		}
		for _, unwanted := range tt.unwanted {
			if strings.Contains(text, unwanted) {
				t.Fatalf("args=%v should not contain %q: %q", tt.args, unwanted, text)
			}
		}
		flags := cloudResourceHelpFlags(text)
		if strings.Contains(strings.Join(tt.args, " "), "openstack") {
			for _, want := range []string{"--auth-url", "--username", "--password", "--user-domain-id", "--network-id", "--os-type"} {
				if !strings.Contains(flags, want) {
					t.Fatalf("args=%v flags missing %q: %q", tt.args, want, flags)
				}
			}
			if strings.Contains(flags, "--access-key-id") {
				t.Fatalf("args=%v OpenStack flags should not contain AK credentials: %q", tt.args, flags)
			}
		} else {
			for _, want := range []string{"--access-key-id", "--access-key-secret", "--flavor-vcpus", "--flavor-ram", "--network-id"} {
				if !strings.Contains(flags, want) {
					t.Fatalf("args=%v flags missing %q: %q", tt.args, want, flags)
				}
			}
			for _, unwanted := range []string{"--cloud-auth-type", "--auth-url", "--username", "--password"} {
				if strings.Contains(flags, unwanted) {
					t.Fatalf("args=%v AK flags should not contain %q: %q", tt.args, unwanted, flags)
				}
			}
		}
		assertNoHelpFooter(t, text)
	}
}

func TestCloudResourceAccountHelpUsesProviderStorageProfile(t *testing.T) {
	cases := []struct {
		name        string
		lang        string
		cloudType   string
		storageType string
		want        []string
		unwanted    []string
		openstack   bool
	}{
		{name: "aliyun block", cloudType: "aliyun_bs", storageType: "HyperGate", want: []string{"Alibaba Cloud", "images,system_disk_types", "cloud-sync-gateway create"}},
		{name: "huawei block zh", lang: "zh_cn", cloudType: "huawei_bs", storageType: "HyperGate", want: []string{"华为云", "images,system_disk_types", "cloud-sync-gateway create"}},
		{name: "aliyun object", cloudType: "aliyun_obs", storageType: "objectstorage", want: []string{"Alibaba Cloud", "system_volume_types,volume_types", "--network-id <network_id>"}, unwanted: []string{"cloud-sync-gateway create"}},
		{name: "huawei object zh", lang: "zh_cn", cloudType: "huawei_obs", storageType: "objectstorage", want: []string{"华为云", "华为云子网不按可用区划分", "subnets,security_groups"}, unwanted: []string{"cloud-sync-gateway create"}},
		{name: "openstack block", cloudType: "openstack", storageType: "HyperGate", openstack: true, want: []string{"regions,compute_zones,projects", "--os-type linux", "cloud-sync-gateway create"}},
		{name: "openstack object zh", lang: "zh_cn", cloudType: "openstack", storageType: "objectstorage", openstack: true, want: []string{"regions,compute_zones,projects", "system_volume_types,volume_types", "--fetch-res networks", "boot-config apply"}, unwanted: []string{"cloud-sync-gateway create"}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)

			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": "00000000",
					"data": map[string]interface{}{
						"cloud_type":   tt.cloudType,
						"storage_type": tt.storageType,
					},
				})
			}))
			defer srv.Close()

			args := []string{"cloud-resource", "fetch", "--cloud-account-id", "account-1", "--help"}
			if tt.lang != "" {
				args = append([]string{"--lang", tt.lang}, args...)
			}
			var out, errOut bytes.Buffer
			if err := Execute(withHost(t, srv.URL, args...), &out, &errOut); err != nil {
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
			for _, unwanted := range append(tt.unwanted, "aliyun_bs", "huawei_bs", "HyperGate", "objectstorage") {
				if strings.Contains(text, unwanted) {
					t.Fatalf("help should not contain %q: %q", unwanted, text)
				}
			}
			flags := cloudResourceHelpFlags(text)
			for _, want := range []string{"--cloud-account-id", "--fetch-res", "--network-id"} {
				if !strings.Contains(flags, want) {
					t.Fatalf("flags missing %q: %q", want, flags)
				}
			}
			if tt.openstack {
				for _, want := range []string{"--project-id", "--project-domain-id", "--compute-zone-id", "--os-type"} {
					if !strings.Contains(flags, want) {
						t.Fatalf("OpenStack flags missing %q: %q", want, flags)
					}
				}
			} else if strings.Contains(flags, "--project-id") || strings.Contains(flags, "--os-type") {
				t.Fatalf("non-OpenStack flags contain OpenStack context: %q", flags)
			}
		})
	}
}

func TestCloudResourceRejectsCloudAccountIDWithDirectFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"cloud-resource", "fetch",
		"--cloud-type", "aliyun",
		"--storage-type", "block",
		"--cloud-account-id", "account-1",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
	), &out, &errOut)
	if err == nil || err.Error() != "cloud-account-id cannot be used with cloud-type, storage-type, or credential flags" {
		t.Fatalf("err=%v", err)
	}
}

func TestCloudResourceBlockUsesDirectAuthEndpoint(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"regions": []map[string]interface{}{
						{"id": "cn-beijing", "display_name": "Beijing"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"cloud-resource", "fetch",
		"--cloud-type", "aliyun",
		"--storage-type", "block",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--fetch-res", "regions",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v3/postCloudInfoForAuth" {
		t.Fatalf("path=%q", gotPath)
	}
	cloudAccount := gotBody["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "aliyun_bs" || cloudAccount["storage_type"] != "HyperGate" {
		t.Fatalf("cloud_account=%#v", cloudAccount)
	}
	if !strings.Contains(out.String(), "cn-beijing") {
		t.Fatalf("output=%q", out.String())
	}
}

func TestCloudResourceHuaweiObjectPurposeIsTopLevel(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/postCloudInfoForAuth" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"cloud_info": map[string]interface{}{
				"flavors": []map[string]interface{}{{"id": "flavor-1", "vcpus": 2, "ram_GB": 4}},
			}},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"cloud-resource", "fetch",
		"--cloud-type", "huawei",
		"--storage-type", "object",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-north-1",
		"--purpose", "make_image",
		"--fetch-res", "flavors",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["purpose"] != "make_image" {
		t.Fatalf("body=%+v", gotBody)
	}
	cloudAccount := gotBody["cloud_account"].(map[string]interface{})
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if _, ok := metadata["purpose"]; ok {
		t.Fatalf("metadata should not contain purpose: %+v", metadata)
	}
}

func TestCloudResourceHuaweiObjectKnownQueryFieldsAreTopLevel(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"cloud_info": map[string]interface{}{
				"images": []map[string]interface{}{{"id": "image-1", "name": "linux", "os_type": "linux"}},
			}},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"cloud-resource", "fetch",
		"--cloud-type", "huawei",
		"--storage-type", "object",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-north-1",
		"--zone-id", "cn-north-1a",
		"--flavor-id", "flavor-1",
		"--flavor-vcpus", "2",
		"--flavor-ram", "4",
		"--network-id", "network-1",
		"--os-type", "linux",
		"--image-type", "system",
		"--boot-mode", "bios",
		"--purpose", "make_image",
		"--provider-option", "custom",
		"--fetch-res", "images",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"fetch_res": "images", "region_id": "cn-north-1", "zone_id": "cn-north-1a",
		"flavor_id": "flavor-1", "flavor_vcpus": "2", "flavor_ram": "4",
		"network_id": "network-1", "os_type": "linux", "image_type": "system",
		"boot_mode": "bios", "purpose": "make_image",
	}
	for key, value := range want {
		if gotBody[key] != value {
			t.Fatalf("body[%q]=%v, want %q; body=%+v", key, gotBody[key], value, gotBody)
		}
	}
	metadata := gotBody["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	for key := range want {
		if _, ok := metadata[key]; ok {
			t.Fatalf("metadata should not contain query field %q: %+v", key, metadata)
		}
	}
	for _, key := range []string{"region_type", "region_type_list"} {
		if _, ok := metadata[key]; ok {
			t.Fatalf("metadata should not contain %q: %+v", key, metadata)
		}
	}
	if metadata["provider_option"] != "custom" {
		t.Fatalf("unmodeled provider field was not preserved: %+v", metadata)
	}
}

func TestCloudResourceFetchRejectsImageTypeUnderscoreFlag(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute([]string{
		"cloud-resource", "fetch",
		"--cloud-type", "huawei",
		"--storage-type", "object",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--image_type", "system",
		"--fetch-res", "images",
	}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --image_type; use --image-type") {
		t.Fatalf("err=%v", err)
	}
}

func TestCloudResourceOpenStackUsesTargetAuthEndpoint(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"projects": []map[string]interface{}{
						{"id": "project-1", "name": "demo"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"cloud-resource", "fetch",
		"--cloud-type", "openstack",
		"--storage-type", "object",
		"--auth-url", "http://identity:5000/v3",
		"--username", "demo",
		"--password", "secret",
		"--user-domain-id", "default",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/postTargetCloudInfoForAuth" {
		t.Fatalf("path=%q", gotPath)
	}
	cloudAccount := gotBody["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "openstack" || cloudAccount["storage_type"] != "objectstorage" {
		t.Fatalf("cloud_account=%#v", cloudAccount)
	}
}

func TestCloudResourceFetchUsesGetCloudInfoByDefault(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var paths []string
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		queries = append(queries, r.URL.RawQuery)
		switch r.URL.Path {
		case "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_type":   "aliyun_bs",
					"storage_type": "HyperGate",
					"region_id":    "cn-beijing",
				},
			})
		case "/api/v3/getCloudInfo":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"regions": []map[string]interface{}{
							{"id": "cn-beijing", "display_name": "Beijing"},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"cloud-resource", "fetch",
		"--cloud-account-id", "account-1",
		"--fetch-res", "regions",
		"--network-id", "network-1",
		"--os-type", "linux",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 || paths[1] != "/api/v3/getCloudInfo" {
		t.Fatalf("paths=%v", paths)
	}
	if !strings.Contains(queries[1], "cloud_account_id=account-1") ||
		!strings.Contains(queries[1], "rt_flatten=1") ||
		!strings.Contains(queries[1], "network_id=network-1") ||
		!strings.Contains(queries[1], "os_type=linux") {
		t.Fatalf("query=%q", queries[1])
	}
}

func TestCloudResourceFetchActionRuleUsesCloudAccountAction(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPostPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_type":   "aliyun_bs",
					"storage_type": "HyperGate",
				},
			})
		case "/hypermotion/v1/cloud_accounts/account-1/action":
			gotPostPath = r.URL.Path
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"abilities": map[string]interface{}{"supports_gateway": true},
					},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"cloud-resource", "fetch",
		"--cloud-account-id", "account-1",
		"--fetch-res", "abilities",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPostPath != "/hypermotion/v1/cloud_accounts/account-1/action" {
		t.Fatalf("path=%q", gotPostPath)
	}
}

func TestCloudResourceFetchFlavorFiltersRenderTable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		switch r.URL.Path {
		case "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_type":   "aliyun_obs",
					"storage_type": "objectstorage",
					"region_id":    "cn-beijing",
				},
			})
		case "/api/v3/getCloudInfo":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"flavors": []map[string]interface{}{
							{
								"id":          "ecs.u1-c1m2.large",
								"name":        "ecs.u1-c1m2.large(2C4G)",
								"vcpus":       2,
								"ram_GB":      4,
								"max_nic_num": 2,
							},
							{
								"id":          "ecs.g6.xlarge",
								"name":        "ecs.g6.xlarge(4C16G)",
								"vcpus":       4,
								"ram_GB":      16,
								"max_nic_num": 4,
							},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"cloud-resource", "fetch",
		"--cloud-account-id", "account-1",
		"--zone-id", "cn-beijing-h",
		"--fetch-res", "flavors",
		"--flavor-vcpus", "2",
		"--flavor-ram", "4",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(gotQuery, "flavor_vcpus=") || strings.Contains(gotQuery, "flavor_ram=") {
		t.Fatalf("query should not contain flavor filters: %q", gotQuery)
	}
	text := out.String()
	if !strings.Contains(text, "ecs.u1-c1m2.large") {
		t.Fatalf("output=%q, missing filtered row", text)
	}
	if strings.Contains(text, "ecs.g6.xlarge") {
		t.Fatalf("output=%q, should filter non-matching flavor", text)
	}
}

func TestCloudResourceFetchFlavorFiltersDoNotAffectJSONOutput(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_type":   "aliyun_obs",
					"storage_type": "objectstorage",
					"region_id":    "cn-beijing",
				},
			})
		case "/api/v3/getCloudInfo":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"flavors": []map[string]interface{}{
							{
								"id":          "ecs.u1-c1m2.large",
								"name":        "ecs.u1-c1m2.large(2C4G)",
								"vcpus":       2,
								"ram_GB":      4,
								"max_nic_num": 2,
							},
							{
								"id":          "ecs.g6.xlarge",
								"name":        "ecs.g6.xlarge(4C16G)",
								"vcpus":       4,
								"ram_GB":      16,
								"max_nic_num": 4,
							},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"cloud-resource", "fetch",
		"--cloud-account-id", "account-1",
		"--zone-id", "cn-beijing-h",
		"--fetch-res", "flavors",
		"--flavor-vcpus", "2",
		"--flavor-ram", "4",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"ecs.u1-c1m2.large", "ecs.g6.xlarge"} {
		if !strings.Contains(text, want) {
			t.Fatalf("json output=%q, missing %q", text, want)
		}
	}
}

func TestTargetResourceDirectAuthImageFilterKeepsRowsWithoutOSType(t *testing.T) {
	rows := []map[string]interface{}{
		{"id": "linux-1", "os_type": "linux"},
		{"id": "windows-1", "os_type": "windows"},
		{"id": "unknown-1"},
	}
	filtered := filterImageRows(rows, map[string]interface{}{"os_type": "linux"})
	if len(filtered) != 2 || filtered[0]["id"] != "linux-1" || filtered[1]["id"] != "unknown-1" {
		t.Fatalf("filtered=%#v", filtered)
	}
}

func TestRemovedTargetResourceCommandsReturnUnknown(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := [][]string{
		{"target", "resource"},
		{"target", "resource", "block"},
		{"target", "resource", "block", "aliyun"},
		{"target", "resource", "oss"},
		{"target", "resource", "oss", "openstack"},
		{"target", "resource", "fetch"},
		{"help", "target", "resource"},
		{"help", "target", "resource", "block"},
		{"help", "target", "resource", "oss"},
		{"help", "target", "resource", "fetch"},
	}
	for _, args := range cases {
		var out, errOut bytes.Buffer
		err := Execute(withHost(t, "https://example.invalid", args...), &out, &errOut)
		if err == nil || !strings.Contains(err.Error(), "unknown") {
			t.Fatalf("args=%v err=%v", args, err)
		}
	}
}

func TestCloudResourceFetchValidation(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want string
	}{
		{args: []string{"cloud-resource", "fetch"}, want: "storage-type is required"},
		{args: []string{"cloud-resource", "fetch", "aliyun"}, want: `unexpected argument "aliyun"`},
		{args: []string{"cloud-resource", "fetch", "--cloud-type", "aliyun"}, want: "storage-type is required"},
		{args: []string{"cloud-resource", "fetch", "--storage-type", "archive"}, want: "storage-type must be block or object"},
		{args: []string{"cloud-resource", "fetch", "--storage-type", "block_storage"}, want: "storage-type must be block or object"},
		{args: []string{"cloud-resource", "fetch", "--storage-type", "object_storage"}, want: "storage-type must be block or object"},
		{args: []string{"cloud-resource", "fetch", "--cloud-type", "vmware", "--storage-type", "block"}, want: `cloud-type "vmware" does not support storage-type block`},
	}
	for _, tt := range cases {
		var out, errOut bytes.Buffer
		err := Execute(withHost(t, "https://example.invalid", tt.args...), &out, &errOut)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Fatalf("args=%v err=%v, want %q", tt.args, err, tt.want)
		}
	}
}
