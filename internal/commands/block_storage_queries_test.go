package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBlockStoragesResourcesBuildsRegionsAndZonesRequest(t *testing.T) {
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
					"domain": map[string]interface{}{
						"regions": []map[string]interface{}{{"id": "cn-beijing", "display_name": "Beijing"}},
					},
					"zones": []map[string]interface{}{{"id": "cn-beijing-h", "display_name": "Zone H"}},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-beijing",
		"--fetch-res", "regions,zones",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/hypermotion/v1/cloud_accounts/account-1/action" {
		t.Fatalf("path = %q", gotPath)
	}
	getCloudInfo := gotBody["get_cloud_info"].(map[string]interface{})
	if getCloudInfo["domain"].(map[string]interface{})["region_id"] != "cn-beijing" {
		t.Fatalf("body = %+v", gotBody)
	}
	resourceOptions := getCloudInfo["resources_options"].(map[string]interface{})
	if resourceOptions["domain"].(map[string]interface{})["region_id"] != "cn-beijing" {
		t.Fatalf("body = %+v", gotBody)
	}
	if resourceOptions["zones"].(map[string]interface{})["region_id"] != "cn-beijing" {
		t.Fatalf("body = %+v", gotBody)
	}
	if !strings.Contains(out.String(), `"zones"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestBlockStoragesResourcesRegionsTableSupportsTopLevelRegionsAndCurrentDomain(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"current_domain": map[string]interface{}{
						"id":           "cn-beijing",
						"display_name": "Beijing",
					},
					"regions": []map[string]interface{}{
						{"id": "cn-beijing", "display_name": "Beijing", "local_name": "华北2（北京）"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-beijing",
		"--fetch-res", "regions",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"Current Domain", "cn-beijing (Beijing)", "== Regions ==", "-------------------------", "Region ID", "华北2（北京）"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBlockStoragesResourcesZonesTableSupportsNestedRegionZones(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/hypermotion/v1/cloud_accounts/account-1" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"region_id": "cn-beijing",
				},
			})
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"current_domain": "cn-beijing",
					"regions": []map[string]interface{}{
						{
							"id":           "cn-beijing",
							"display_name": "Beijing",
							"zones": []map[string]interface{}{
								{"id": "cn-beijing-l", "display_name": "Zone L"},
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
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
		"--fetch-res", "zones",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	getCloudInfo := gotBody["get_cloud_info"].(map[string]interface{})
	domain := getCloudInfo["domain"].(map[string]interface{})
	if domain["region_id"] != "cn-beijing" {
		t.Fatalf("domain = %#v", domain)
	}
	resourceOptions := getCloudInfo["resources_options"].(map[string]interface{})
	zoneOptions := resourceOptions["zones"].(map[string]interface{})
	if zoneOptions["region_id"] != "cn-beijing" {
		t.Fatalf("zones = %#v", zoneOptions)
	}
	text := out.String()
	for _, want := range []string{"Current Domain", "cn-beijing", "== Zones ==", "-------------------------", "Zone ID", "cn-beijing-l", "Zone L"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBlockStoragesResourcesMultiResourceHumanOutputUsesSections(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"current_domain": map[string]interface{}{
						"id":           "cn-beijing",
						"display_name": "Beijing",
					},
					"regions": []map[string]interface{}{
						{
							"id":           "cn-beijing",
							"display_name": "Beijing",
							"local_name":   "华北2（北京）",
							"zones": []map[string]interface{}{
								{"id": "cn-beijing-l", "display_name": "Zone L"},
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
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-beijing",
		"--fetch-res", "regions,zones",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"Current Domain",
		"cn-beijing (Beijing)",
		"== Regions ==",
		"-------------------------",
		"Region ID",
		"华北2（北京）",
		"== Zones ==",
		"-------------------------",
		"Zone ID",
		"cn-beijing-l",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBlockStoragesResourcesRequiresContextFields(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing zone", args: []string{"--cloud-account-id", "account-1", "--region-id", "cn-beijing", "--fetch-res", "flavors"}, want: "zone-id is required"},
		{name: "missing flavor", args: []string{"--cloud-account-id", "account-1", "--region-id", "cn-beijing", "--zone-id", "cn-beijing-h", "--fetch-res", "images"}, want: "flavor-id is required"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)

			var out, errOut bytes.Buffer
			args := append(withHost(t, "https://example.invalid", "target", "cloud-sync-gateway", "resources"), tc.args...)
			err := Execute(args, &out, &errOut)
			if err == nil || err.Error() != tc.want {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestBlockStoragesResourcesWithoutFetchResBuildsMinimalRequestAndAutoDetectsSections(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"region_id": "cn-beijing",
				},
			})
		case "/hypermotion/v1/cloud_accounts/account-1/action":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"current_domain": map[string]interface{}{
							"id":           "cn-beijing",
							"display_name": "Beijing",
						},
						"regions": []map[string]interface{}{
							{"id": "cn-beijing", "display_name": "Beijing", "local_name": "华北2（北京）"},
						},
						"projects": []map[string]interface{}{
							{"id": "project-1", "name": "autotest", "domain_id": "default"},
						},
						"compute_zones": []map[string]interface{}{
							{"id": "nova", "name": "nova"},
						},
						"flavors": []map[string]interface{}{
							{"id": "ecs.g1", "name": "ecs.g1", "vcpus": 2, "max_nic_num": 4},
						},
						"boot_loader_images": []map[string]interface{}{
							{"id": "boot-img-1", "name": "Windows_DriverFix_1", "os_type": "windows"},
						},
						"volume_types": []map[string]interface{}{
							{"id": "DEFAULT_VOLUME_TYPE", "name": "DEFAULT_VOLUME_TYPE"},
						},
						"networks": []map[string]interface{}{
							{"id": "net-1", "name": "public-net"},
						},
						"subnets": []map[string]interface{}{
							{"id": "subnet-1", "name": "public-subnet-10", "network_id": "net-1", "cidr": "192.168.0.0/20"},
						},
						"auth_info": map[string]interface{}{
							"project_id": "project-1",
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
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	getCloudInfo := gotBody["get_cloud_info"].(map[string]interface{})
	domain := getCloudInfo["domain"].(map[string]interface{})
	if domain["region_id"] != "cn-beijing" {
		t.Fatalf("domain = %#v", domain)
	}
	resourceOptions := getCloudInfo["resources_options"].(map[string]interface{})
	if len(resourceOptions) != 0 {
		t.Fatalf("resources_options = %#v", resourceOptions)
	}

	text := out.String()
	for _, want := range []string{
		"Current Domain",
		"cn-beijing (Beijing)",
		"== Regions ==",
		"华北2（北京）",
		"== Projects ==",
		"autotest",
		"== Compute Zones ==",
		"nova",
		"== Flavors ==",
		"ecs.g1",
		"== Windows Transition Images ==",
		"Windows_DriverFix_1",
		"== System Volume Types ==",
		"DEFAULT_VOLUME_TYPE",
		"== Networks ==",
		"public-net",
		"== Subnets ==",
		"192.168.0.0/20",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, "auth_info") {
		t.Fatalf("table output should not include raw auth_info fields: %q", text)
	}
	for _, hidden := range []string{"NVMe 支持", "内存 (GiB)", "操作系统版本", "CIDR 网段", "可用区 ID"} {
		if strings.Contains(text, hidden) {
			t.Fatalf("output = %q, should hide empty column %q", text, hidden)
		}
	}
}

func TestBlockStoragesResourcesWithoutFetchResJSONKeepsRawFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"region_id": "cn-beijing",
				},
			})
		case "/hypermotion/v1/cloud_accounts/account-1/action":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"auth_info": map[string]interface{}{
							"project_id": "project-1",
						},
						"networks": []map[string]interface{}{
							{"id": "net-1", "name": "public-net"},
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
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{`"auth_info"`, `"project_id": "project-1"`, `"networks"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBlockStoragesResourcesBuildsHelperImageRequest(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"win_hd_images": []map[string]interface{}{
						{"id": "win2016", "name": "win2016", "os_type": "windows", "os_version": "Windows Server 2016"},
					},
					"linux_hd_images": []map[string]interface{}{
						{"id": "linux1", "name": "linux1", "os_type": "linux", "os_version": "Linux"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-beijing",
		"--fetch-res", "win_hd_images,linux_hd_images",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	getCloudInfo := gotBody["get_cloud_info"].(map[string]interface{})
	resourceOptions := getCloudInfo["resources_options"].(map[string]interface{})
	winOptions := resourceOptions["win_hd_images"].(map[string]interface{})
	linuxOptions := resourceOptions["linux_hd_images"].(map[string]interface{})
	if winOptions["image_type"] != "system" {
		t.Fatalf("win_hd_images = %#v", winOptions)
	}
	if linuxOptions["image_type"] != "user_create" {
		t.Fatalf("linux_hd_images = %#v", linuxOptions)
	}
	text := out.String()
	for _, want := range []string{"== Windows Helper Images ==", "win2016", "== Linux Helper Images ==", "linux1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBlockStoragesResourcesShowsColumnsWhenAnyRowHasValue(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"images": []map[string]interface{}{
						{"id": "img-1", "name": "image-1", "os_type": "linux", "os_version": ""},
						{"id": "img-2", "name": "image-2", "os_type": "linux", "os_version": "24.04"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-beijing",
		"--zone-id", "cn-beijing-h",
		"--flavor-id", "flavor-1",
		"--fetch-res", "images",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"OS Version", "24.04"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBlockStoragesResourcesSubnetsHideZoneColumnWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"subnets": []map[string]interface{}{
						{"id": "subnet-1", "name": "subnet-1", "network_id": "net-1", "cidr": "192.168.0.0/20"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-beijing",
		"--zone-id", "cn-beijing-h",
		"--fetch-res", "subnets",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if strings.Contains(text, "Zone ID") {
		t.Fatalf("output = %q, should hide empty zone column", text)
	}
	if !strings.Contains(text, "CIDR Block") || !strings.Contains(text, "192.168.0.0/20") {
		t.Fatalf("output = %q", text)
	}
}

func TestBlockStoragesResourcesNetworksHideCIDRColumnWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"networks": []map[string]interface{}{
						{"id": "net-1", "name": "public-net"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-beijing",
		"--fetch-res", "networks",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if strings.Contains(text, "CIDR Block") {
		t.Fatalf("output = %q, should hide empty CIDR column", text)
	}
	for _, want := range []string{"ID", "Name", "public-net"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBlockStoragesResourcesOpenStackHelperImagesDoNotRequireRegion(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_type": "openstack",
				},
			})
		case "/hypermotion/v1/cloud_accounts/account-1/action":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"win_hd_images": []map[string]interface{}{
							{"id": "win2016", "name": "win2016"},
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
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
		"--fetch-res", "win_hd_images",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	getCloudInfo := gotBody["get_cloud_info"].(map[string]interface{})
	domain := getCloudInfo["domain"].(map[string]interface{})
	if _, ok := domain["region_id"]; ok {
		t.Fatalf("domain should omit region_id: %#v", domain)
	}
	text := out.String()
	for _, want := range []string{"== Windows Helper Images ==", "win2016"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBlockStoragesResourcesRejectsUnsupportedFetchRes(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "cloud-sync-gateway", "resources",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-beijing",
		"--fetch-res", "boot_loader_images",
	), &out, &errOut)
	if err == nil || err.Error() != `unsupported fetch-res value "boot_loader_images"` {
		t.Fatalf("err = %v", err)
	}
}

func TestBlockStoragesSubnetConfigBuildsQueryAndInfersCloudType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotCloudType, gotRegionID, gotZoneID, gotNetworkID, gotSubnetID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"cloud_type": "aliyun_bs"},
			})
		case "/api/v3/getSubnetConfig":
			gotCloudType = r.URL.Query().Get("cloud_type")
			gotRegionID = r.URL.Query().Get("region_id")
			gotZoneID = r.URL.Query().Get("zone_id")
			gotNetworkID = r.URL.Query().Get("network_id")
			gotSubnetID = r.URL.Query().Get("subnet_id")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"fixed_ip_limit": map[string]interface{}{
							"ranges": []map[string]interface{}{{"idx": 3, "range": "1-254"}},
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
		"target", "cloud-sync-gateway", "subnet-config",
		"--cloud-account-id", "account-1",
		"--region-id", "cn-beijing",
		"--zone-id", "cn-beijing-h",
		"--network-id", "vpc-1",
		"--subnet-id", "vsw-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotCloudType != "aliyun_bs" || gotRegionID != "cn-beijing" || gotZoneID != "cn-beijing-h" || gotNetworkID != "vpc-1" || gotSubnetID != "vsw-1" {
		t.Fatalf("query = cloud_type=%s region=%s zone=%s network=%s subnet=%s", gotCloudType, gotRegionID, gotZoneID, gotNetworkID, gotSubnetID)
	}
	if !strings.Contains(out.String(), `"fixed_ip_limit"`) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestBlockStoragesSubnetConfigUsesAccountRegionWhenRegionFlagIsOmitted(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotRegionID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_type": "aliyun_bs",
					"region_id":  "cn-beijing",
				},
			})
		case "/api/v3/getSubnetConfig":
			gotRegionID = r.URL.Query().Get("region_id")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "cloud-sync-gateway", "subnet-config",
		"--cloud-account-id", "account-1",
		"--zone-id", "cn-beijing-h",
		"--network-id", "vpc-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotRegionID != "cn-beijing" {
		t.Fatalf("region_id = %q", gotRegionID)
	}
}

func TestBlockStoragesSubnetConfigOmitsRegionWhenNeitherFlagNorAccountProvidesIt(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotRegionID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hypermotion/v1/cloud_accounts/account-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{
					"cloud_type": "aliyun_bs",
				},
			})
		case "/api/v3/getSubnetConfig":
			gotRegionID = r.URL.Query().Get("region_id")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "cloud-sync-gateway", "subnet-config",
		"--cloud-account-id", "account-1",
		"--zone-id", "cn-beijing-h",
		"--network-id", "vpc-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotRegionID != "" {
		t.Fatalf("region_id should be omitted, got %q", gotRegionID)
	}
}

func TestBlockStoragesListDefaultsToHyperGate(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"storages": []map[string]interface{}{}},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "cloud-sync-gateway", "list",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "type=HyperGate") {
		t.Fatalf("query = %q", gotQuery)
	}
}

func TestBlockStoragesListSupportsCloudAccountFilterAndGatewayColumns(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"storages": []map[string]interface{}{
					{
						"id":                         "storage-1",
						"display_name":               "CLOUD_SYNC_GATEWAY_1",
						"display_cloud_account_name": "aliyun-account",
						"display_status":             "Available",
						"region_name":                "China (Beijing)",
						"zone_name":                  "cn-beijing-h",
						"public_ip":                  "1.2.3.4",
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "cloud-sync-gateway", "list",
		"--cloud-account-id", "account-1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "cloud_account_uuid=account-1") {
		t.Fatalf("query = %q", gotQuery)
	}

	text := out.String()
	for _, want := range []string{
		"Storage Name",
		"Cloud Account",
		"Region Name",
		"Zone Name",
		"Public IP",
		"CLOUD_SYNC_GATEWAY_1",
		"aliyun-account",
		"1.2.3.4",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, "UUID") {
		t.Fatalf("output should not use object-storage columns: %q", text)
	}
}

func TestBlockStoragesListNormalizesGatewayFieldsFromFallbackSources(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"storages": []map[string]interface{}{
					{
						"uuid":                       "storage-uuid-1",
						"display_cloud_storage_name": "CLOUD_SYNC_GATEWAY_47.93.4.160",
						"status":                     "available",
						"metadata": map[string]interface{}{
							"region_name": "华北2（北京）",
							"zone_name":   "北京 可用区 H",
							"private_ip":  "10.0.0.8",
						},
					},
					{
						"id":             "storage-2",
						"display_name":   "CLOUD_SYNC_GATEWAY_2",
						"display_status": "创建失败",
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "cloud-sync-gateway", "list",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"storage-uuid-1",
		"CLOUD_SYNC_GATEWAY_47.93.4.160",
		"华北2（北京）",
		"北京 可用区 H",
		"10.0.0.8",
		"storage-2",
		"CLOUD_SYNC_GATEWAY_2",
		"创建失败",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestBlockStoragesDetailRequiresID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "target", "cloud-sync-gateway", "detail"), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestBlockStoragesCreateFlatPathRejected(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "cloud-sync-gateway", "create",
		"--cloud-account-id", "account-1",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "requires subcommand") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsGatewayCommandsRemoved(t *testing.T) {
	tests := []string{
		"gateway-resources",
		"gateway-subnet-config",
		"gateway-transition-images",
		"gateway-create",
	}

	for _, subcommand := range tests {
		t.Run(subcommand, func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)

			var out, errOut bytes.Buffer
			err := Execute(withHost(t, "https://example.invalid",
				"cloud-accounts", subcommand,
			), &out, &errOut)
			if err == nil || !strings.Contains(err.Error(), `unknown command "cloud-accounts"`) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}
