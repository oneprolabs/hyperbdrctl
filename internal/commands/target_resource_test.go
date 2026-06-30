package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCloudResourceHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-resource", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"Usage:", "\nFlags:\n", "\nCommands:\n", "fetch", "Usage Notes:"} {
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
	for _, want := range []string{"--cloud-account-id", "--cloud-type", "--storage-type", "--fetch-res", "Usage Notes:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
}

func TestCloudResourceDirectHelpShowsProviderProfile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"cloud-resource", "fetch", "--cloud-type", "aliyun", "--storage-type", "block", "--help"},
			want: []string{"--access-key-id", "--access-key-secret", "block", "Usage Notes:"},
		},
		{
			args: []string{"cloud-resource", "fetch", "--cloud-type", "openstack", "--storage-type", "object", "--help"},
			want: []string{"--auth-url", "--username", "--password", "--user-domain-id", "object", "Usage Notes:"},
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
		assertNoHelpFooter(t, text)
	}
}

func TestCloudResourceAccountHelpReadsAccountDetail(t *testing.T) {
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
		"cloud-resource", "fetch",
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
	for _, want := range []string{"--cloud-account-id", "aliyun_bs", "HyperGate", "Usage Notes:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
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
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 || paths[1] != "/api/v3/getCloudInfo" {
		t.Fatalf("paths=%v", paths)
	}
	if !strings.Contains(queries[1], "cloud_account_id=account-1") || !strings.Contains(queries[1], "rt_flatten=1") {
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
