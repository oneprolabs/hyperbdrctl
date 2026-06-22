package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTargetResourceHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "resource", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"Usage:", "\nFlags:\n", "\nCommands:\n", "block", "oss", "fetch", "Usage Notes:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestTargetResourceProviderGroupsShowProviders(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := [][]string{
		{"target", "resource", "block", "--help"},
		{"target", "resource", "oss", "--help"},
	}

	for _, args := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", args, err)
		}
		text := out.String()
		for _, want := range []string{"aliyun", "openstack", "Usage Notes:"} {
			if !strings.Contains(text, want) {
				t.Fatalf("args=%v missing %q: %q", args, want, text)
			}
		}
		assertNoHelpFooter(t, text)
	}
}

func TestTargetResourceFetchHelpRequiresCloudAccountID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"target", "resource", "fetch", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"--cloud-account-id", "--fetch-res", "Usage Notes:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
}

func TestTargetResourceDirectAuthRejectsCloudAccountID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "resource", "block", "aliyun",
		"--cloud-account-id", "account-1",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
	), &out, &errOut)
	if err == nil || err.Error() != "cloud-account-id cannot be used with target resource block aliyun" {
		t.Fatalf("err=%v", err)
	}
}

func TestTargetResourceBlockUsesDirectAuthEndpoint(t *testing.T) {
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
		"target", "resource", "block", "aliyun",
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

func TestTargetResourceOpenStackUsesTargetAuthEndpoint(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
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
		"target", "resource", "oss", "openstack",
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
}

func TestTargetResourceFetchUsesGetCloudInfoByDefault(t *testing.T) {
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
		"target", "resource", "fetch",
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

func TestTargetResourceFetchActionRuleUsesCloudAccountAction(t *testing.T) {
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
		"target", "resource", "fetch",
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

func TestTargetResourceFetchFlavorFiltersRenderTable(t *testing.T) {
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
		"target", "resource", "fetch",
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

func TestTargetResourceFetchFlavorFiltersDoNotAffectJSONOutput(t *testing.T) {
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
		"target", "resource", "fetch",
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
