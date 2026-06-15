package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCloudAccountsFetchResourcesRegionsNormalizesTableFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/postCloudInfoForAuth" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"regions": []map[string]interface{}{
						{
							"id":           "cn-beijing",
							"display_name": "North China 2 (Beijing)",
							"local_name":   "华北2（北京）",
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-block-resources", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--fetch-res", "regions",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	cloudAccount := gotBody["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "aliyun_bs" || cloudAccount["storage_type"] != "HyperGate" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
	if _, ok := gotBody["region_id"]; ok {
		t.Fatalf("regions request should not include region_id: %+v", gotBody)
	}
	if _, ok := gotBody["boot_mode"]; ok {
		t.Fatalf("regions request should not include boot_mode: %+v", gotBody)
	}

	text := out.String()
	for _, want := range []string{"cn-beijing", "North China 2 (Beijing)", "华北2（北京）"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestCloudAccountsFetchResourcesImagesNormalizesTableFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/postCloudInfoForAuth" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"images": []map[string]interface{}{
						{
							"id":         "img-1",
							"name":       "Alibaba Linux 3",
							"os_type":    "linux",
							"os_version": "3.2104 LTS",
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-oss-resources", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-beijing",
		"--fetch-res", "images",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := gotBody["boot_mode"]; ok {
		t.Fatalf("images request should not include boot_mode by default: %+v", gotBody)
	}

	text := out.String()
	for _, want := range []string{"img-1", "Alibaba Linux 3", "linux", "3.2104 LTS"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, "Boot Mode") {
		t.Fatalf("output should not include boot mode column when no boot field is returned: %q", text)
	}
}

func TestCloudAccountsFetchResourcesImagesShowsBootModeFromBootFirmware(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/postCloudInfoForAuth" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"images": []map[string]interface{}{
						{
							"id":            "img-uefi-1",
							"name":          "Windows Server 2022",
							"os_type":       "windows",
							"os_version":    "2022",
							"boot_firmware": "UEFI",
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-oss-resources", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-beijing",
		"--fetch-res", "images",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"Boot Mode", "UEFI", "img-uefi-1", "Windows Server 2022"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestCloudAccountsFetchResourcesImagesHideEmptyColumns(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/postCloudInfoForAuth" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"images": []map[string]interface{}{
						{
							"id":      "img-1",
							"name":    "Alibaba Linux 3",
							"os_type": "linux",
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-oss-resources", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-beijing",
		"--fetch-res", "images",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if strings.Contains(text, "OS Version") {
		t.Fatalf("output should hide empty columns: %q", text)
	}
}

func TestCloudAccountsFetchResourcesZonesRenderTable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/postCloudInfoForAuth" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"regions": []map[string]interface{}{
						{
							"id": "cn-north-1",
							"zones": []map[string]interface{}{
								{"id": "cn-north-1a", "display_name": "AZ1"},
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
		"target", "account", "fetch-oss-resources", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-north-1",
		"--fetch-res", "zones",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"== Zones ==", "Zone ID", "cn-north-1a", "AZ1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestCloudAccountsFetchResourcesMultipleSectionsRenderInOrder(t *testing.T) {
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
					"regions": []map[string]interface{}{
						{"id": "cn-beijing", "display_name": "Beijing"},
					},
					"zones": []map[string]interface{}{
						{"id": "cn-beijing-h", "display_name": "Zone H"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-block-resources", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--fetch-res", "regions,zones",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["fetch_res"] != "regions,zones" {
		t.Fatalf("body = %+v", gotBody)
	}

	text := out.String()
	regions := strings.Index(text, "== Regions ==")
	zones := strings.Index(text, "== Zones ==")
	if regions < 0 || zones < 0 || regions >= zones {
		t.Fatalf("section order mismatch: %q", text)
	}
}

func TestCloudAccountsFetchResourcesAutoDetectsSectionsWhenFetchResOmitted(t *testing.T) {
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
					"regions": []map[string]interface{}{
						{"id": "cn-beijing", "display_name": "Beijing"},
					},
					"zones": []map[string]interface{}{
						{"id": "cn-beijing-h", "display_name": "Zone H"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-block-resources", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := gotBody["fetch_res"]; ok {
		t.Fatalf("body should not contain fetch_res: %+v", gotBody)
	}

	text := out.String()
	for _, want := range []string{"== Regions ==", "== Zones ==", "cn-beijing", "cn-beijing-h"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestCloudAccountsFetchResourcesBootLoaderImagesAndFlavorsRenderTables(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"boot_loader_images": []map[string]interface{}{
						{"id": "boot-img-1", "name": "Windows Driver", "os_type": "windows"},
					},
					"flavors": []map[string]interface{}{
						{"id": "ecs.g6.large", "name": "2C4G", "vcpus": 2},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-oss-resources", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--region-id", "cn-beijing",
		"--fetch-res", "boot_loader_images,flavors",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"== Windows Transition Images ==", "Windows Driver", "== Flavors ==", "ecs.g6.large", "Flavor ID"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	for _, unwanted := range []string{"Quota Rate", "Quota PPS", "Zone ID"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("output = %q, should not contain %q for generic flavors", text, unwanted)
		}
	}
}

func TestCloudAccountsFetchResourcesHuaweiFlavorsUseProviderSpecificColumns(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"boot_loader_images": []map[string]interface{}{
						{"id": "boot-img-1", "name": "Windows Driver", "os_type": "windows"},
					},
					"flavors": []map[string]interface{}{
						{
							"id":           "s3.large.2",
							"name":         "s3.large.2",
							"vcpus":        2,
							"ram":          4,
							"zone_id":      "cn-north-1a",
							"GHz":          "Intel SkyLake 6161 2.2GHz",
							"quota_rate":   "0.2 / 0.8 Gbit/s",
							"quota_pps":    "100,000 PPS",
							"max_nic_num":  12,
							"max_disk_num": 24,
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-oss-resources", "huawei",
		"--access-id", "ak",
		"--access-secret", "sk",
		"--region-id", "cn-north-1",
		"--fetch-res", "boot_loader_images,flavors",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	bootImages := strings.Index(text, "== Windows Transition Images ==")
	flavors := strings.Index(text, "== Flavors ==")
	if bootImages < 0 || flavors < 0 || bootImages >= flavors {
		t.Fatalf("section order mismatch: %q", text)
	}
	for _, want := range []string{
		"Flavor ID",
		"Flavor Name",
		"RAM (GiB)",
		"Zone ID",
		"GHz",
		"Quota Rate",
		"Quota PPS",
		"Max NICs",
		"Max Disk Num",
		"s3.large.2",
		"cn-north-1a",
		"0.2 / 0.8 Gbit/s",
		"100,000 PPS",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestCloudAccountsFetchResourcesHuaweiBootLoaderFlavorsRenderTable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"boot_loader_flavors": []map[string]interface{}{
						{
							"id":           "boot-flavor-1",
							"name":         "boot-flavor-1",
							"vcpus":        2,
							"ram_GB":       8,
							"zone_id":      "cn-north-1b",
							"quota_rate":   "0.4 / 1.5 Gbit/s",
							"quota_pps":    "150,000 PPS",
							"max_nic_num":  12,
							"max_disk_num": 24,
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-block-resources", "huawei",
		"--access-id", "ak",
		"--access-secret", "sk",
		"--region-id", "cn-north-1",
		"--fetch-res", "boot_loader_flavors",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"== Windows Transition Flavors ==",
		"Flavor ID",
		"RAM (GiB)",
		"Zone ID",
		"Quota Rate",
		"Quota PPS",
		"boot-flavor-1",
		"cn-north-1b",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, "\"boot_loader_flavors\"") {
		t.Fatalf("output should render a table instead of raw JSON: %q", text)
	}
}

func TestCloudAccountsFetchResourcesHuaweiFlavorsPassesZoneIDAndFiltersInCLI(t *testing.T) {
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
		"target", "account", "fetch-block-resources", "huawei",
		"--access-id", "ak",
		"--access-secret", "sk",
		"--region-id", "cn-north-1",
		"--zone-id", "cn-north-1a",
		"--fetch-res", "flavors",
		"--flavor-vcpus", "2",
		"--flavor-ram", "4",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	if gotBody["zone_id"] != "cn-north-1a" {
		t.Fatalf("body = %+v", gotBody)
	}
	if _, ok := gotBody["flavor_vcpus"]; ok {
		t.Fatalf("body should not contain flavor_vcpus: %+v", gotBody)
	}
	if _, ok := gotBody["flavor_ram"]; ok {
		t.Fatalf("body should not contain flavor_ram: %+v", gotBody)
	}

	text := out.String()
	if !strings.Contains(text, "c3.large.2") {
		t.Fatalf("output = %q, missing filtered row", text)
	}
	if strings.Contains(text, "c3.xlarge.2") {
		t.Fatalf("output = %q, should filter non-matching flavor", text)
	}
}

func TestCloudAccountsFetchResourcesNetworksAndSubnetsRenderTables(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"networks": []map[string]interface{}{
						{
							"id":           "24e0103e-1fbc-4b4d-8b37-916cc41843f0",
							"name":         "vpc-ray",
							"cidr_block":   "10.0.0.0/8",
							"display_name": "vpc-ray(10.0.0.0/8)",
						},
					},
					"subnets": []map[string]interface{}{
						{
							"id":           "6abbb2f0-d351-4a0f-9b33-9d2babcad652",
							"name":         "subnet-ray",
							"network_id":   "24e0103e-1fbc-4b4d-8b37-916cc41843f0",
							"zone_id":      "cn-north-1c",
							"cidr_block":   "10.0.0.0/24",
							"display_name": "subnet-ray(10.0.0.0/24)",
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-oss-resources", "huawei",
		"--access-id", "ak",
		"--access-secret", "sk",
		"--region-id", "cn-north-1",
		"--fetch-res", "networks,subnets",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"== Networks ==",
		"== Subnets ==",
		"vpc-ray",
		"10.0.0.0/8",
		"subnet-ray",
		"24e0103e-1fbc-4b4d-8b37-916cc41843f0",
		"cn-north-1c",
		"10.0.0.0/24",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, "\"networks\"") || strings.Contains(text, "\"subnets\"") {
		t.Fatalf("output should render tables instead of raw JSON: %q", text)
	}
}

func TestCloudAccountsFetchResourcesSystemVolumeTypesRenderTable(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"system_volume_types": []map[string]interface{}{
						{
							"id":           "09edebca-2128-49b8-9aa3-4c69c3c6e7f6",
							"display_name": "通用型SSD",
							"min_GB":       40,
							"max_GB":       32768,
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-oss-resources", "huawei",
		"--access-id", "ak",
		"--access-secret", "sk",
		"--region-id", "cn-north-1",
		"--zone-id", "cn-north-1a",
		"--fetch-res", "system_volume_types",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"== System Volume Types ==",
		"Disk Type",
		"Display Name",
		"Min GB",
		"Max GB",
		"09edebca-2128-49b8-9aa3-4c69c3c6e7f6",
		"通用型SSD",
		"40",
		"32768",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, "\"system_volume_types\"") {
		t.Fatalf("output should render a table instead of raw JSON: %q", text)
	}
}

func TestCloudAccountsFetchOSSResourcesOpenStackBuildsValidatedGatewayAuthRequest(t *testing.T) {
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
						{"id": "RegionOne", "display_name": "RegionOne"},
					},
					"projects": []map[string]interface{}{
						{"id": "project-1", "name": "autotest"},
					},
					"compute_zones": []map[string]interface{}{
						{"id": "nova", "name": "nova"},
					},
					"boot_loader_images": []map[string]interface{}{
						{"id": "boot-img-1", "name": "Windows_DriverFix_c5ed1912", "os_type": "windows"},
					},
					"flavors": []map[string]interface{}{
						{"id": "flavor-1", "name": "2C_4G_40G", "vcpus": 2, "ram_GB": 4},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-oss-resources", "openstack",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--username", "autotest",
		"--password", "autotest",
		"--user-domain-id", "default",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	if gotPath != "/api/v2/postTargetCloudInfoForAuth" {
		t.Fatalf("path = %q", gotPath)
	}
	cloudAccount := gotBody["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "openstack" || cloudAccount["cloud_auth_type"] != "password" || cloudAccount["storage_type"] != "objectstorage" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if metadata["auth_url"] != "http://192.168.10.201:5000/v3" || metadata["user_domain_id"] != "default" || metadata["username"] != "autotest" || metadata["password"] != "autotest" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if gotBody["fetch_scene"] != "gateway" || gotBody["rt_tree"].(float64) != 0 {
		t.Fatalf("body = %+v", gotBody)
	}
	if _, ok := gotBody["fetch_res"]; ok {
		t.Fatalf("fetch_res should be omitted by default: %+v", gotBody["fetch_res"])
	}
	for _, key := range []string{"region_id", "project_id", "project_domain_id", "project_name", "compute_zone_id", "block_store_zone_id"} {
		if gotBody[key] != nil {
			t.Fatalf("%s = %#v, want nil", key, gotBody[key])
		}
	}

	text := out.String()
	for _, want := range []string{"== Regions ==", "RegionOne", "== Projects ==", "autotest", "== Compute Zones ==", "nova", "== Windows Transition Images ==", "Windows_DriverFix_c5ed1912"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestCloudAccountsFetchOSSResourcesOpenStackJSONKeepsRawFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"projects": []map[string]interface{}{
						{"id": "project-1", "name": "autotest"},
					},
					"boot_loader_images": []map[string]interface{}{
						{"id": "boot-img-1", "name": "Windows_DriverFix_c5ed1912"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"target", "account", "fetch-oss-resources", "openstack",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--username", "autotest",
		"--password", "autotest",
		"--user-domain-id", "default",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{`"projects"`, `"boot_loader_images"`, `"Windows_DriverFix_c5ed1912"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestCloudAccountsFetchResourcesSupportsAccessIDAlias(t *testing.T) {
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
					"regions": []map[string]interface{}{
						{"id": "cn-north-1", "display_name": "CN North 1"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-oss-resources", "huawei",
		"--access-id", "ak",
		"--access-secret", "sk",
		"--fetch-res", "regions",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	cloudAccount := gotBody["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_auth_type"] != "aksk" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
	metadata := cloudAccount["metadata"].(map[string]interface{})
	if metadata["access_id"] != "ak" || metadata["access_secret"] != "sk" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if _, ok := metadata["access_key_id"]; ok {
		t.Fatalf("metadata should preserve alias keys: %+v", metadata)
	}
}

func TestCloudAccountsFetchResourcesRejectsMixedCredentialStylesWithoutExplicitAuthType(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "account", "fetch-block-resources", "huawei",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--username", "demo",
		"--password", "secret",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "multiple credential styles provided; pass --cloud-auth-type explicitly") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsFetchBlockResourcesOpenStackUsesPasswordPath(t *testing.T) {
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
						{"id": "RegionOne", "display_name": "RegionOne"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-block-resources", "openstack",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--username", "autotest",
		"--password", "autotest",
		"--user-domain-id", "default",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	if gotPath != "/api/v2/postTargetCloudInfoForAuth" {
		t.Fatalf("path = %q", gotPath)
	}
	cloudAccount := gotBody["cloud_account"].(map[string]interface{})
	if cloudAccount["cloud_type"] != "openstack" || cloudAccount["cloud_auth_type"] != "password" || cloudAccount["storage_type"] != "HyperGate" {
		t.Fatalf("cloud_account = %+v", cloudAccount)
	}
}

func TestCloudAccountsFetchResourcesRejectsLegacyOpenStackCredentialFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "account", "fetch-oss-resources", "openstack",
		"--auth-url", "http://192.168.10.201:5000/v3",
		"--cloud-account-username", "autotest",
		"--cloud-account-password", "autotest",
		"--user-domain-id", "default",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --cloud-account-username") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsFetchResourcesRegionsFallsBackToLegacyTopLevelRows(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/postCloudInfoForAuth" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"regions": []map[string]interface{}{
					{
						"id":           "cn-qingdao",
						"display_name": "North China 1 (Qingdao)",
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "fetch-block-resources", "aliyun",
		"--access-key-id", "ak",
		"--access-key-secret", "sk",
		"--fetch-res", "regions",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"cn-qingdao", "North China 1 (Qingdao)"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestCloudAccountsLegacyFetchResourcesShowsMigration(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid",
		"target", "account", "fetch-resources",
		"--cloud-type", "aliyun_bs",
		"--storage-type", "HyperGate",
	), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "use target account fetch-block-resources <provider> or target account fetch-oss-resources <provider>") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsListBuildsQuery(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_accounts": []map[string]interface{}{
					{"id": "account-1", "name": "aliyun-account"},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "list",
		"--page", "2",
		"--page-size", "50",
		"--storage-type", "objectstorage",
		"--custom-filter", "x",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"page=2", "page_size=50", "storage_type=objectstorage", "custom_filter=x"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %q, missing %q", gotQuery, want)
		}
	}
}

func TestCloudAccountsListOmitsStorageTypeWhenNotProvided(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_accounts": []map[string]interface{}{
					{"id": "account-1", "name": "mixed-account"},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "list",
		"--page", "1",
		"--page-size", "10",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(gotQuery, "storage_type=") {
		t.Fatalf("query should omit storage_type by default: %q", gotQuery)
	}
	for _, want := range []string{"page=1", "page_size=10"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %q, missing %q", gotQuery, want)
		}
	}
}

func TestCloudAccountsListNormalizesBlockStorageFilterToHyperGate(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"cloud_accounts": []map[string]interface{}{
					{"id": "account-1", "name": "openstack-account"},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "list",
		"--storage-type", "blockstorage",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "storage_type=HyperGate") {
		t.Fatalf("query = %q", gotQuery)
	}
}

func TestCloudAccountsDetailRequiresID(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "target", "account", "detail"), &out, &errOut)
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestCloudAccountsDeleteBuildsForcePathAndBody(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotMethod, gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.RequestURI()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"deleted": true}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"target", "account", "delete",
		"--id", "account-1",
		"--storage-type", "HyperGate",
		"--force",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/hypermotion/v1/cloud_accounts/account-1?force=true" {
		t.Fatalf("method=%q path=%q", gotMethod, gotPath)
	}
	if gotBody["id"] != "account-1" || gotBody["storage_type"] != "HyperGate" {
		t.Fatalf("body = %+v", gotBody)
	}
}
