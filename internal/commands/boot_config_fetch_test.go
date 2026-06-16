package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBootConfigHelpShowsFetchCommands(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"help", "boot-config"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"fetch-block-resources", "fetch-oss-resources", "Usage Notes:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	assertNoHelpFooter(t, text)
}

func TestBootConfigFetchResourcesHelpUsesGroupLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"boot-config", "fetch-block-resources", "--help"},
			want: []string{"\nCommands:\n", "aliyun", "openstack", "Usage Notes:"},
		},
		{
			args: []string{"boot-config", "fetch-oss-resources", "--help"},
			want: []string{"\nCommands:\n", "aliyun", "openstack", "Usage Notes:"},
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
				t.Fatalf("args=%v help missing %q: %q", tt.args, want, text)
			}
		}
		assertNoHelpFooter(t, text)
	}
}

func TestBootConfigFetchOSSAliyunHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"boot-config", "fetch-oss-resources", "aliyun", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"Usage Notes:", "--cloud-account-id", "--zone-id", "--flavor-id", "regions,zones", "boot-config-wizard subnet-config"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	if strings.Contains(text, "\nCommands:\n") {
		t.Fatalf("provider help should not include commands section: %q", text)
	}
	assertNoHelpFooter(t, text)
}

func TestBootConfigFetchOSSHuaweiHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"boot-config", "fetch-oss-resources", "huawei", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"Usage Notes:", "--cloud-account-id", "--zone-id", "--flavor-vcpus", "regions,zones", "boot-config-wizard subnet-config"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
	if strings.Contains(text, "\nCommands:\n") {
		t.Fatalf("provider help should not include commands section: %q", text)
	}
	assertNoHelpFooter(t, text)
}

func TestBootConfigFetchOSSAliyunRegionsZonesDoesNotRequireRegionID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"regions": []map[string]interface{}{
						{
							"id":           "cn-beijing",
							"display_name": "China (Beijing)",
							"zones": []map[string]interface{}{
								{"id": "cn-beijing-h", "display_name": "Beijing H"},
							},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config", "fetch-oss-resources", "aliyun",
		"--cloud-account-id", "account-1",
		"--fetch-res", "regions,zones",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v3/getCloudInfo" {
		t.Fatalf("path = %q", gotPath)
	}
	for _, want := range []string{
		"cloud_account_id=account-1",
		"cloud_type=aliyun_obs",
		"storage_type=objectstorage",
		"fetch_res=regions%2Czones",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %q, want contains %q", gotQuery, want)
		}
	}
	if strings.Contains(gotQuery, "region_id=") {
		t.Fatalf("query must not contain region_id: %q", gotQuery)
	}
	text := out.String()
	for _, want := range []string{"== Regions ==", "== Zones ==", "cn-beijing", "cn-beijing-h"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBootConfigFetchOSSAliyunFlavorsRequireZoneID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"boot-config", "fetch-oss-resources", "aliyun",
		"--cloud-account-id", "account-1",
		"--fetch-res", "flavors,os_types",
	), &out, &errOut)
	if err == nil || err.Error() != "zone-id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestBootConfigFetchOSSHuaweiRegionsZonesDoesNotRequireRegionID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"regions": []map[string]interface{}{
						{
							"id":           "cn-north-1",
							"display_name": "CN North 1",
							"zones": []map[string]interface{}{
								{"id": "cn-north-1a", "display_name": "CN North 1A"},
							},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config", "fetch-oss-resources", "huawei",
		"--cloud-account-id", "account-1",
		"--fetch-res", "regions,zones",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v3/getCloudInfo" {
		t.Fatalf("path = %q", gotPath)
	}
	for _, want := range []string{
		"cloud_account_id=account-1",
		"cloud_type=huawei_obs",
		"storage_type=objectstorage",
		"fetch_res=regions%2Czones",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %q, want contains %q", gotQuery, want)
		}
	}
	if strings.Contains(gotQuery, "region_id=") {
		t.Fatalf("query must not contain region_id: %q", gotQuery)
	}
	text := out.String()
	for _, want := range []string{"== Regions ==", "== Zones ==", "cn-north-1", "cn-north-1a"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBootConfigFetchOSSHuaweiFlavorsRequireZoneID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"boot-config", "fetch-oss-resources", "huawei",
		"--cloud-account-id", "account-1",
		"--fetch-res", "flavors,os_types",
	), &out, &errOut)
	if err == nil || err.Error() != "zone-id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestBootConfigFetchOSSHuaweiVolumeTypesRequireZoneID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"boot-config", "fetch-oss-resources", "huawei",
		"--cloud-account-id", "account-1",
		"--fetch-res", "system_volume_types,volume_types",
	), &out, &errOut)
	if err == nil || err.Error() != "zone-id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestBootConfigFetchOSSAliyunVolumeTypesRequireFlavorID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"boot-config", "fetch-oss-resources", "aliyun",
		"--cloud-account-id", "account-1",
		"--fetch-res", "system_volume_types,volume_types",
		"--zone-id", "cn-beijing-h",
	), &out, &errOut)
	if err == nil || err.Error() != "flavor-id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestBootConfigFetchBlockResourcesGenericProviderUsesFixedCloudTypeAndPassthrough(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"projects": []map[string]interface{}{
						{"id": "project-1", "name": "Project One"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"boot-config", "fetch-block-resources", "openstack",
		"--cloud-account-id", "account-1",
		"--fetch-res", "projects",
		"--project-id", "project-1",
		"--foo-bar", "baz",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"cloud_account_id=account-1",
		"cloud_type=openstack",
		"storage_type=HyperGate",
		"fetch_res=projects",
		"project_id=project-1",
		"foo_bar=baz",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %q, want contains %q", gotQuery, want)
		}
	}
	if !strings.Contains(out.String(), "\"projects\"") {
		t.Fatalf("json output = %q", out.String())
	}
}

func TestBootConfigFetchOSSHuaweiFlavorFiltersRenderTable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"flavors": []map[string]interface{}{
						{
							"id":   "u-2",
							"name": "2C",
							"children": []interface{}{
								map[string]interface{}{
									"id":   "u-2-m-4",
									"name": "4GB",
									"children": []interface{}{
										map[string]interface{}{
											"id":           "c3.large.2",
											"name":         "c3.large.2",
											"vcpus":        2,
											"ram_GB":       4,
											"zone_id":      "cn-north-1a",
											"quota_rate":   "0.6 / 1.5 Gbit/s",
											"quota_pps":    "300,000 PPS",
											"max_nic_num":  12,
											"max_disk_num": 24,
										},
										map[string]interface{}{
											"id":           "c3.xlarge.2",
											"name":         "c3.xlarge.2",
											"vcpus":        4,
											"ram_GB":       8,
											"zone_id":      "cn-north-1a",
											"quota_rate":   "1 / 3 Gbit/s",
											"quota_pps":    "500,000 PPS",
											"max_nic_num":  12,
											"max_disk_num": 24,
										},
									},
								},
							},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config", "fetch-oss-resources", "huawei",
		"--cloud-account-id", "account-1",
		"--zone-id", "cn-north-1a",
		"--fetch-res", "flavors",
		"--flavor-vcpus", "2",
		"--flavor-ram", "4",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if !strings.Contains(text, "c3.large.2") {
		t.Fatalf("output = %q, missing filtered row", text)
	}
	if strings.Contains(text, "c3.xlarge.2") {
		t.Fatalf("output = %q, should filter non-matching flavor", text)
	}
	for _, want := range []string{"== Flavors ==", "Flavor ID", "RAM (GiB)", "Zone ID"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBootConfigFetchOSSAliyunSecurityGroupsRenderTable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"networks": []map[string]interface{}{
						{"id": "vpc-1", "name": "vpc-one", "cidr_block": "192.168.0.0/16"},
					},
					"subnets": []map[string]interface{}{
						{"id": "vsw-1", "name": "vsw-one", "network_id": "vpc-1", "cidr_block": "192.168.0.0/24"},
					},
					"security_groups": []map[string]interface{}{
						{"id": "sg-1", "name": "sg-one", "network_id": "vpc-1"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"boot-config", "fetch-oss-resources", "aliyun",
		"--cloud-account-id", "account-1",
		"--fetch-res", "networks,subnets,security_groups",
		"--zone-id", "cn-beijing-h",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"== Networks ==", "== Subnets ==", "== Security Groups ==", "sg-1", "sg-one", "vpc-1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}
