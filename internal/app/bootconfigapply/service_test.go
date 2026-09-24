package bootconfigapply

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hyperbdr-client/internal/client"
)

type fakeAPI struct {
	gets     []func(string, url.Values) (client.APIResponse, error)
	posts    []func(string, interface{}) (client.APIResponse, error)
	getPath  string
	postPath string
	postBody interface{}
}

func (f *fakeAPI) Get(path string, q url.Values) (client.APIResponse, error) {
	f.getPath = path
	if len(f.gets) == 0 {
		return client.APIResponse{}, nil
	}
	fn := f.gets[0]
	f.gets = f.gets[1:]
	return fn(path, q)
}

func (f *fakeAPI) Post(path string, body interface{}) (client.APIResponse, error) {
	f.postPath = path
	f.postBody = body
	if len(f.posts) == 0 {
		return client.APIResponse{}, nil
	}
	fn := f.posts[0]
	f.posts = f.posts[1:]
	return fn(path, body)
}

func TestResolveApplyDriver(t *testing.T) {
	tests := []struct {
		name        string
		metadata    map[string]interface{}
		storageInfo storageDefaults
		want        string
	}{
		{
			name:        "aliyun object storage",
			metadata:    map[string]interface{}{"cloud_type": "aliyun_obs", "storage_type": "objectstorage"},
			storageInfo: storageDefaults{IsObjectStorage: true},
			want:        "aliyun_object_storage",
		},
		{
			name:        "aliyun block storage",
			metadata:    map[string]interface{}{"cloud_type": "aliyun_bs", "storage_type": "HyperGate"},
			storageInfo: storageDefaults{IsBlockStorage: true},
			want:        "aliyun_block_storage",
		},
		{
			name:        "fallback generic",
			metadata:    map[string]interface{}{"cloud_type": "openstack", "storage_type": "objectstorage"},
			storageInfo: storageDefaults{IsObjectStorage: true},
			want:        "generic",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			driver := resolveApplyDriver(tc.metadata, tc.storageInfo)
			if driver.Name() != tc.want {
				t.Fatalf("driver = %q, want %q", driver.Name(), tc.want)
			}
		})
	}
}

func TestApplyCreatesWithOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(path, []byte(`{"region_id":"cn-beijing","repair_host_mapper":{"enable_dhcp_mode":"1"},"nics":[{"subnet_id":"vsw-old","security_groups":[{"id":"sg-old"}]}],"disk_volume_mapper":[{"volume_type_id":"cloud_efficiency"}]}`), 0600); err != nil {
		t.Fatal(err)
	}

	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v2/getHostDetail" {
					t.Fatalf("unexpected get path %q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID:   "host-1",
		File:     path,
		Dynamic:  map[string]string{"region_id": "cn-shanghai", "zone_id": "cn-shanghai-b"},
		Sets:     []string{"repair_host_mapper.enable_dhcp_mode=0", "nics[0].subnet_id=vsw-new", "disk_volume_mapper[0].volume_type_id=cloud_essd"},
		SetJSONs: []string{`nics=[{"subnet_id":"vsw-json","security_groups":[{"id":"sg-json"}]}]`},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Operation != "create" || api.postPath != "/api/v2/batchBootConfigs" {
		t.Fatalf("result=%+v postPath=%q", result, api.postPath)
	}
	item := api.postBody.(map[string]interface{})["batch_create"].([]map[string]interface{})[0]
	meta := item["metadata"].(map[string]interface{})
	if meta["region_id"] != "cn-shanghai" || meta["zone_id"] != "cn-shanghai-b" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["nics"].([]interface{})[0].(map[string]interface{})["subnet_id"] != "vsw-json" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["repair_host_mapper"].(map[string]interface{})["enable_dhcp_mode"] != 0 {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestApplyAutofillsCloudAccountAndObjectStorageMetadata(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/hypermotion/v1/cloud_accounts/account-1" {
					t.Fatalf("unexpected path %q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_type":              "aliyun_obs",
					"cloud_type_display_name": "validated-cloud-display",
					"cloud_account": map[string]interface{}{
						"name":       "aliyun-account",
						"username":   "ak-value",
						"region_id":  "cn-beijing",
						"cloud_type": "aliyun_obs",
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v2/getStorageDetailInfo" || q.Get("storage_id") != "storage-1" {
					t.Fatalf("unexpected path=%q query=%v", path, q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"type":         "objectstorage",
					"name":         "storage-name",
					"display_name": "storage-display",
					"pool_id":      "pool-1",
					"pool_name":    "pool-name",
					"network_addrs_for_write_data": []interface{}{
						map[string]interface{}{"name": "public_endpoint", "display_name": "object-storage.example.invalid:9000(Public Network)", "value": "public_endpoint"},
					},
					"network_addrs_for_read_data": []interface{}{
						map[string]interface{}{"name": "internal_endpoint", "display_name": "object-storage.example.invalid:9000(Private Network)", "value": "internal_endpoint"},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v2/getHostDetail" {
					t.Fatalf("unexpected path %q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"id": "host-1",
				}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_account_id": "account-1",
			"storage_id":       "storage-1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	meta := result.Metadata
	if meta["cloud_type"] != "aliyun_obs" || meta["cloud_type_name"] != "validated-cloud-display" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["cloud_type_display_name"] != "validated-cloud-display" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["cloud_account_name"] != "aliyun-account" || meta["cloud_account_username"] != "ak-value" || meta["region_id"] != "cn-beijing" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["storage_name"] != "storage-name" || meta["storage_display_name"] != "storage-display" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["network_addr_for_write_data"] != "public_endpoint" || meta["network_addr_for_write_data_name"] != "object-storage.example.invalid:9000(Public Network)" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["network_addr_for_read_data"] != "internal_endpoint" || meta["network_addr_for_read_data_name"] != "object-storage.example.invalid:9000(Private Network)" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["default_pool_id"] != "pool-1" || meta["default_pool_name"] != "pool-name" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["storage_type"] != "objectstorage" {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestApplyAutofillsObjectStoragePoolFromNestedStoragePools(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_type": "aliyun_obs",
					"cloud_account": map[string]interface{}{
						"name":      "aliyun-account",
						"username":  "ak-value",
						"region_id": "cn-beijing",
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v2/getStorageDetailInfo" || q.Get("storage_id") != "storage-1" {
					t.Fatalf("unexpected path=%q query=%v", path, q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"storages": []interface{}{
						map[string]interface{}{
							"type":         "objectstorage",
							"name":         "storage-name",
							"display_name": "storage-display",
							"storage_pools": []interface{}{
								map[string]interface{}{"uuid": "pool-1", "name": "pool-name"},
								map[string]interface{}{"uuid": "pool-2", "name": "pool-name-2"},
							},
							"network_addrs_for_write_data": []interface{}{
								map[string]interface{}{"name": "public_endpoint", "display_name": "public", "value": "public_endpoint"},
							},
							"network_addrs_for_read_data": []interface{}{
								map[string]interface{}{"name": "internal_endpoint", "display_name": "private", "value": "internal_endpoint"},
							},
						},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"id": "host-1",
					"disks": []interface{}{
						map[string]interface{}{"index": 0, "disk_id": "disk-boot", "is_boot_disk": true},
					},
				}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_account_id": "account-1",
			"storage_id":       "storage-1",
			"volume_type_id":   "cloud_essd_entry",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	meta := result.Metadata
	if meta["default_pool_id"] != "pool-1" || meta["default_pool_name"] != "pool-name" {
		t.Fatalf("metadata = %+v", meta)
	}
	rows := meta["disk_volume_mapper"].([]interface{})
	row := rows[0].(map[string]interface{})
	if row["pool_id"] != "pool-1" || row["pool_name"] != "pool-name" {
		t.Fatalf("row = %+v", row)
	}
}

func TestApplyAutofillsObjectStorageMetadataFromStorageWrapperAndConfig(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Raw: map[string]interface{}{
					"cloud_account": map[string]interface{}{
						"cloud_type":   "aliyun_obs",
						"custom_name":  "validated-cloud-display-region-1",
						"region_name":  "region-1",
						"username":     "ak-value",
						"account_name": "aliyun_obs_20260601134445",
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"storage": map[string]interface{}{
						"type":         "objectstorage",
						"display_name": "阿里云-华北2（北京）",
						"name":         "pool uuid",
						"config": map[string]interface{}{
							"public_endpoint":   "oss-cn-beijing.aliyuncs.com",
							"internal_endpoint": "oss-cn-beijing-internal.aliyuncs.com",
						},
						"storage_pools": []interface{}{
							map[string]interface{}{"uuid": "pool-1", "name": "bucket-1", "display_name": "bucket-1"},
						},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_account_id": "account-1",
			"storage_id":       "storage-1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	meta := result.Metadata
	if meta["cloud_type_display_name"] != "validated-cloud-display" || meta["cloud_type_name"] != "validated-cloud-display" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["storage_name"] != "" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["storage_display_name"] != "阿里云-华北2（北京）(bucket-1)" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["network_addr_for_write_data"] != "public_endpoint" || meta["network_addr_for_write_data_name"] != "oss-cn-beijing.aliyuncs.com" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["network_addr_for_read_data"] != "internal_endpoint" || meta["network_addr_for_read_data_name"] != "oss-cn-beijing-internal.aliyuncs.com" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["default_pool_id"] != "pool-1" || meta["default_pool_name"] != "bucket-1" {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestApplyKeepsExplicitDefaultPoolValues(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"storages": []interface{}{
						map[string]interface{}{
							"type": "objectstorage",
							"storage_pools": []interface{}{
								map[string]interface{}{"uuid": "pool-1", "name": "pool-name"},
							},
						},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"id": "host-1",
					"disks": []interface{}{
						map[string]interface{}{"index": 0, "disk_id": "disk-boot", "is_boot_disk": true},
					},
				}}, nil
			},
		},
	}

	service := NewService(api)
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata.json")
	body := `{"storage_id":"storage-1","volume_type_id":"cloud_essd_entry","default_pool_id":"custom-pool","default_pool_name":"custom-pool-name","disk_volume_mapper":[{"pool_id":"row-pool","pool_name":"row-pool-name"}]}`
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := service.Apply(ApplyInput{HostID: "host-1", File: path})
	if err != nil {
		t.Fatal(err)
	}
	meta := result.Metadata
	if meta["default_pool_id"] != "custom-pool" || meta["default_pool_name"] != "custom-pool-name" {
		t.Fatalf("metadata = %+v", meta)
	}
	row := meta["disk_volume_mapper"].([]interface{})[0].(map[string]interface{})
	if row["pool_id"] != "row-pool" || row["pool_name"] != "row-pool-name" {
		t.Fatalf("row = %+v", row)
	}
}

func TestApplyAutofillOnlyFillsMissingFields(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_type": "aliyun_obs",
					"cloud_account": map[string]interface{}{
						"name":      "server-name",
						"username":  "server-ak",
						"region_id": "cn-beijing",
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"id": "host-1",
				}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_account_id":       "account-1",
			"cloud_type":             "custom-cloud",
			"cloud_account_name":     "custom-name",
			"cloud_account_username": "custom-user",
			"region_id":              "cn-shanghai",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	meta := result.Metadata
	if meta["cloud_type"] != "custom-cloud" || meta["cloud_account_name"] != "custom-name" || meta["cloud_account_username"] != "custom-user" || meta["region_id"] != "cn-shanghai" {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestApplyGeneratesDiskVolumeMapperForAllDisks(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_account": map[string]interface{}{
						"cloud_type": "aliyun_obs",
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"type":      "objectstorage",
					"pool_id":   "pool-1",
					"pool_name": "pool-name",
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"id": "host-1",
					"disks": []interface{}{
						map[string]interface{}{"index": 0, "disk_id": "disk-boot", "is_boot_disk": true},
						map[string]interface{}{"index": 1, "disk_id": "disk-data", "is_boot_disk": false},
					},
				}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_account_id":                 "account-1",
			"storage_id":                       "storage-1",
			"system_volume_type_id":            "cloud_boot",
			"system_volume_type_name":          "cloud_boot",
			"system_volume_type_display_name":  "Boot Disk",
			"volume_type_id":                   "cloud_data",
			"volume_type_name":                 "cloud_data",
			"volume_type_display_name":         "Data Disk",
			"default_volume_type_id":           "cloud_default",
			"default_volume_type_name":         "cloud_default",
			"default_volume_type_display_name": "Default Disk",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	rows := result.Metadata["disk_volume_mapper"].([]interface{})
	if len(rows) != 2 {
		t.Fatalf("rows = %+v", rows)
	}
	bootRow := rows[0].(map[string]interface{})
	dataRow := rows[1].(map[string]interface{})
	if bootRow["disk_id"] != "disk-boot" || bootRow["volume_type_id"] != "cloud_boot" || bootRow["volume_type_display_name"] != "Boot Disk" {
		t.Fatalf("boot row = %+v", bootRow)
	}
	if dataRow["disk_id"] != "disk-data" || dataRow["volume_type_id"] != "cloud_data" || dataRow["volume_type_display_name"] != "Data Disk" {
		t.Fatalf("data row = %+v", dataRow)
	}
	if bootRow["pool_id"] != "pool-1" || dataRow["pool_name"] != "pool-name" {
		t.Fatalf("rows = %+v", rows)
	}
	if dataRow["default_volume_type_id"] != "cloud_default" {
		t.Fatalf("data row = %+v", dataRow)
	}
}

func TestApplyGeneratesBlockStorageDiskVolumeMapperPoolsByVolumeType(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"storage_type": "HyperGate",
					"cloud_type":   "aliyun_bs",
					"storage_pools": []interface{}{
						map[string]interface{}{"name": "cloud_efficiency", "uuid": "pool-eff", "display_name": "高效云盘"},
						map[string]interface{}{"name": "cloud_ssd", "uuid": "pool-ssd", "display_name": "SSD 云盘"},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"id": "host-1",
					"disks": []interface{}{
						map[string]interface{}{"index": 0, "disk_id": "disk-boot", "is_boot_disk": true},
						map[string]interface{}{"index": 1, "disk_id": "disk-data-1", "is_boot_disk": false},
					},
				}}, nil
			},
		},
	}

	result, err := NewService(api).Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_type":              "aliyun_bs",
			"storage_id":              "storage-1",
			"system_volume_type_id":   "cloud_efficiency",
			"system_volume_type_name": "cloud_efficiency",
			"volume_type_id":          "cloud_ssd",
			"volume_type_name":        "cloud_ssd",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	rows := result.Metadata["disk_volume_mapper"].([]interface{})
	bootRow := rows[0].(map[string]interface{})
	dataRow := rows[1].(map[string]interface{})
	if bootRow["pool_id"] != "pool-eff" || bootRow["pool_name"] != "cloud_efficiency" {
		t.Fatalf("boot row = %+v", bootRow)
	}
	if dataRow["pool_id"] != "pool-ssd" || dataRow["pool_name"] != "cloud_ssd" {
		t.Fatalf("data row = %+v", dataRow)
	}
}

func TestApplyBlockStorageRowVolumeTypeOverridesTopLevelPoolSelection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata.json")
	body := `{"cloud_type":"aliyun_bs","storage_id":"storage-1","volume_type_id":"cloud_efficiency","disk_volume_mapper":[{"volume_type_id":"cloud_ssd"}]}`
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}

	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"storage_type": "HyperGate",
					"cloud_type":   "aliyun_bs",
					"storage_pools": []interface{}{
						map[string]interface{}{"name": "cloud_efficiency", "uuid": "pool-eff"},
						map[string]interface{}{"name": "cloud_ssd", "uuid": "pool-ssd"},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"id": "host-1",
					"disks": []interface{}{
						map[string]interface{}{"index": 0, "disk_id": "disk-boot", "is_boot_disk": true},
					},
				}}, nil
			},
		},
	}

	result, err := NewService(api).Apply(ApplyInput{HostID: "host-1", File: path})
	if err != nil {
		t.Fatal(err)
	}
	row := result.Metadata["disk_volume_mapper"].([]interface{})[0].(map[string]interface{})
	if row["pool_id"] != "pool-ssd" || row["pool_name"] != "cloud_ssd" {
		t.Fatalf("row = %+v", row)
	}
}

func TestApplyBlockStorageKeepsExplicitPoolFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata.json")
	body := `{"cloud_type":"aliyun_bs","storage_id":"storage-1","volume_type_id":"cloud_efficiency","disk_volume_mapper":[{"pool_id":"custom-pool","pool_name":"custom-name"}]}`
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}

	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"storage_type": "HyperGate",
					"cloud_type":   "aliyun_bs",
					"storage_pools": []interface{}{
						map[string]interface{}{"name": "cloud_efficiency", "uuid": "pool-eff"},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"id": "host-1",
					"disks": []interface{}{
						map[string]interface{}{"index": 0, "disk_id": "disk-boot", "is_boot_disk": true},
					},
				}}, nil
			},
		},
	}

	result, err := NewService(api).Apply(ApplyInput{HostID: "host-1", File: path})
	if err != nil {
		t.Fatal(err)
	}
	row := result.Metadata["disk_volume_mapper"].([]interface{})[0].(map[string]interface{})
	if row["pool_id"] != "custom-pool" || row["pool_name"] != "custom-name" {
		t.Fatalf("row = %+v", row)
	}
}

func TestApplyBlockStorageErrorsWhenVolumeTypeHasNoMatchingPool(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"storage_type": "HyperGate",
					"cloud_type":   "aliyun_bs",
					"storage_pools": []interface{}{
						map[string]interface{}{"name": "cloud_efficiency", "uuid": "pool-eff"},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"id": "host-1",
					"disks": []interface{}{
						map[string]interface{}{"index": 0, "disk_id": "disk-boot", "is_boot_disk": true},
					},
				}}, nil
			},
		},
	}

	_, err := NewService(api).Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_type":     "aliyun_bs",
			"storage_id":     "storage-1",
			"volume_type_id": "cloud_ssd",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "disk_volume_mapper[0]") || !strings.Contains(err.Error(), "cloud_ssd") || !strings.Contains(err.Error(), "storage-1") {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyObjectStorageCloudInfoAutofillAndFallbacks(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_type": "aliyun_obs",
					"cloud_account": map[string]interface{}{
						"name":      "account-name",
						"username":  "account-ak",
						"region_id": "cn-beijing",
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"type": "objectstorage",
					"name": "storage-name",
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"os_name":   "Ubuntu Linux",
					"boot_mode": "uefi",
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v3/getCloudInfo" || q.Get("fetch_res") != "regions,zones" {
					t.Fatalf("unexpected path=%q query=%v", path, q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"regions": []interface{}{
							map[string]interface{}{"id": "cn-beijing", "display_name": "China (Beijing)"},
						},
						"zones": []interface{}{},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if q.Get("fetch_res") != "flavors" {
					t.Fatalf("unexpected query=%v", q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"flavors": []interface{}{
							map[string]interface{}{
								"id": "group-1",
								"children": []interface{}{
									map[string]interface{}{
										"id": "group-2",
										"children": []interface{}{
											map[string]interface{}{
												"id":          "ecs.u1-c1m2.large",
												"name":        "ecs.u1-c1m2.large(2C4G)",
												"value":       "id-ecs.u1-c1m2.large",
												"vcpus":       2,
												"ram_GB":      4,
												"max_nic_num": 2,
											},
										},
									},
								},
							},
						},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				switch q.Get("fetch_res") {
				case "regions,zones":
					return client.APIResponse{Data: map[string]interface{}{
						"cloud_info": map[string]interface{}{
							"regions": []interface{}{
								map[string]interface{}{"id": "cn-beijing", "display_name": "China (Beijing)"},
							},
							"zones": []interface{}{},
						},
					}}, nil
				case "os_types":
					return client.APIResponse{Data: map[string]interface{}{
						"cloud_info": map[string]interface{}{
							"os_types": []interface{}{
								map[string]interface{}{"id": "Linux", "name": "Linux", "os_type": "Linux"},
							},
						},
					}}, nil
				default:
					t.Fatalf("unexpected query=%v", q)
					return client.APIResponse{}, nil
				}
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if q.Get("fetch_res") != "system_volume_types,volume_types" {
					t.Fatalf("unexpected query=%v", q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"volume_types": []interface{}{
							map[string]interface{}{"id": "cloud_essd_entry", "display_name": "ESSD Entry Disk"},
						},
						"system_volume_types": []interface{}{
							map[string]interface{}{"id": "cloud_essd_entry", "display_name": "ESSD Entry Disk"},
						},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if q.Get("fetch_res") != "networks,security_groups" || q.Get("network_id") != "vpc-1" || q.Get("volume_type_id") != "cloud_essd_entry" {
					t.Fatalf("unexpected query=%v", q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"networks": []interface{}{
							map[string]interface{}{"id": "vpc-1", "name": "terraform-vpc", "display_name": "terraform-vpc (198.51.100.0/24)"},
						},
						"security_groups": []interface{}{},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v3/getSubnetConfig" {
					t.Fatalf("unexpected path=%q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"subnets": []interface{}{
						map[string]interface{}{"id": "vsw-1", "name": "terraform-vswitch", "display_name": "terraform-vswitch (198.51.100.0/24)"},
					},
				}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_account_id":  "account-1",
			"storage_id":        "storage-1",
			"region_id":         "cn-beijing",
			"zone_id":           "cn-beijing-h",
			"flavor_id":         "ecs.u1-c1m2.large",
			"volume_type_id":    "cloud_essd_entry",
			"network_id":        "vpc-1",
			"subnet_id":         "vsw-1",
			"security_group_id": "sg-1",
			"bandwidth_size":    "100",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	meta := result.Metadata
	if meta["region_name"] != "China (Beijing)" || meta["zone_name"] != "cn-beijing-h" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["flavors"] != "id-ecs.u1-c1m2.large" || meta["flavor_name"] != "ecs.u1-c1m2.large(2C4G)" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["flavor_vcpus"] != 2 || meta["flavor_ram"] != 4 || meta["max_nic_num"] != 2 {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["boot_loader_flavor_id"] != "ecs.u1-c1m2.large" || meta["boot_loader_flavor_name"] != "ecs.u1-c1m2.large(2C4G)" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["os_type_id"] != "id-Linux" || meta["os_type_name"] != "Auto Match (Linux)" || meta["os_type"] != "Linux" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["system_volume_type_id"] != "cloud_essd_entry" || meta["default_volume_type_id"] != "cloud_essd_entry" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["volume_type_name"] != "cloud_essd_entry" || meta["system_volume_type_name"] != "cloud_essd_entry" || meta["default_volume_type_name"] != "cloud_essd_entry" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["network_name"] != "terraform-vpc" || meta["network_display_name"] != "terraform-vpc (198.51.100.0/24)" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["subnet_name"] != "terraform-vswitch" || meta["subnet_display_name"] != "terraform-vswitch (198.51.100.0/24)" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["security_group_name"] != "sg-1" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["security_group_display_name"] != "sg-1" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["dest_boot_mode"] != "uefi" || meta["dest_boot_mode_name"] != "UEFI" {
		t.Fatalf("metadata = %+v", meta)
	}
	nics, ok := meta["nics"].([]interface{})
	if !ok || len(nics) != 1 {
		t.Fatalf("metadata = %+v", meta)
	}
	nic := nics[0].(map[string]interface{})
	if nic["index"] != 0 || nic["network_id"] != "vpc-1" || nic["subnet_id"] != "vsw-1" {
		t.Fatalf("nic = %+v", nic)
	}
	if nic["network_name"] != "terraform-vpc" || nic["network_display_name"] != "terraform-vpc (198.51.100.0/24)" {
		t.Fatalf("nic = %+v", nic)
	}
	if nic["subnet_name"] != "terraform-vswitch" || nic["subnet_display_name"] != "terraform-vswitch (198.51.100.0/24)" {
		t.Fatalf("nic = %+v", nic)
	}
	if nic["is_fixed_ip_input"] != "false" || nic["is_fixed_ip_select"] != "false" || nic["bandwidth_size"] != "100" {
		t.Fatalf("nic = %+v", nic)
	}
	sgs := nic["security_groups"].([]interface{})
	if len(sgs) != 1 || sgs[0].(map[string]interface{})["id"] != "sg-1" || sgs[0].(map[string]interface{})["name"] != "sg-1" {
		t.Fatalf("nic = %+v", nic)
	}
}

func TestApplyGeneratesDefaultNICWithoutExistingScaffold(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"network_id":                  "vpc-1",
			"network_name":                "terraform-vpc",
			"network_display_name":        "terraform-vpc (198.51.100.0/24)",
			"subnet_id":                   "vsw-1",
			"subnet_name":                 "terraform-vswitch",
			"subnet_display_name":         "terraform-vswitch (198.51.100.0/24)",
			"security_group_id":           "sg-1",
			"security_group_name":         "t-1",
			"security_group_display_name": "t-1",
			"bandwidth_size":              "100",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	nics := result.Metadata["nics"].([]interface{})
	if len(nics) != 1 {
		t.Fatalf("metadata = %+v", result.Metadata)
	}
	nic := nics[0].(map[string]interface{})
	if nic["fixed_ip"] != "" || nic["floating_ip"] != "" || nic["bandwidth_id"] != "" || nic["bandwidth_name"] != "" {
		t.Fatalf("nic = %+v", nic)
	}
	if nic["security_groups"].([]interface{})[0].(map[string]interface{})["name"] != "t-1" {
		t.Fatalf("nic = %+v", nic)
	}
}

func TestApplyObjectStorageDerivesDefaultVolumeTypesWhenUserDidNotPassThem(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_type": "aliyun_obs",
					"cloud_account": map[string]interface{}{
						"name":      "account-name",
						"username":  "account-ak",
						"region_id": "cn-beijing",
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"type": "objectstorage",
					"name": "storage-name",
					"storage_pools": []interface{}{
						map[string]interface{}{"uuid": "pool-1", "name": "bucket-1"},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"boot_mode": "bios",
					"disks": []interface{}{
						map[string]interface{}{"index": 0, "disk_id": "disk-boot", "is_boot_disk": true},
						map[string]interface{}{"index": 1, "disk_id": "disk-data", "is_boot_disk": false},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if q.Get("fetch_res") != "regions,zones" {
					t.Fatalf("unexpected query=%v", q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"regions": []interface{}{
							map[string]interface{}{"id": "cn-beijing", "display_name": "China (Beijing)"},
						},
						"zones": []interface{}{},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if q.Get("fetch_res") != "system_volume_types,volume_types" {
					t.Fatalf("unexpected query=%v", q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"volume_types": []interface{}{
							map[string]interface{}{"id": "cloud_data", "display_name": "Data Disk", "is_recommend": 1},
						},
						"system_volume_types": []interface{}{
							map[string]interface{}{"id": "cloud_system", "display_name": "System Disk", "is_recommend": 1},
						},
					},
				}}, nil
			},
		},
	}

	result, err := NewService(api).Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_account_id": "account-1",
			"storage_id":       "storage-1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	meta := result.Metadata
	if meta["volume_type_id"] != "cloud_data" || meta["system_volume_type_id"] != "cloud_system" || meta["default_volume_type_id"] != "cloud_data" {
		t.Fatalf("metadata = %+v", meta)
	}
	rows := meta["disk_volume_mapper"].([]interface{})
	bootRow := rows[0].(map[string]interface{})
	dataRow := rows[1].(map[string]interface{})
	if bootRow["volume_type_id"] != "cloud_system" || bootRow["volume_type_display_name"] != "System Disk" {
		t.Fatalf("bootRow = %+v", bootRow)
	}
	if dataRow["volume_type_id"] != "cloud_data" || dataRow["volume_type_display_name"] != "Data Disk" {
		t.Fatalf("dataRow = %+v", dataRow)
	}
}

func TestApplyStorageAutofillOnlyAppliesObjectStorageFieldsForObjectStorage(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"storage_type": "HyperGate",
					"name":         "block-storage-name",
					"pool_id":      "pool-1",
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID:  "host-1",
		Dynamic: map[string]string{"storage_id": "storage-1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	meta := result.Metadata
	if _, ok := meta["storage_name"]; ok {
		t.Fatalf("metadata = %+v", meta)
	}
	if _, ok := meta["network_addr_for_write_data"]; ok {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestApplyAutofillsZoneIDFromAliyunBlockStorageDetail(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v2/getStorageDetailInfo" || q.Get("storage_id") != "storage-1" {
					t.Fatalf("unexpected path=%q query=%v", path, q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"storage_type": "HyperGate",
					"cloud_type":   "aliyun_bs",
					"zone_id":      "cn-beijing-h",
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_type": "aliyun_bs",
			"storage_id": "storage-1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Metadata["zone_id"] != "cn-beijing-h" {
		t.Fatalf("metadata = %+v", result.Metadata)
	}
}

func TestApplyBlockStorageSubnetConstrainsNetworkAndSecurityGroup(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/hypermotion/v1/cloud_accounts/account-1" {
					t.Fatalf("unexpected path %q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_type": "aliyun_bs",
					"cloud_account": map[string]interface{}{
						"region_id": "cn-beijing",
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v2/getStorageDetailInfo" {
					t.Fatalf("unexpected path %q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"storage_type": "HyperGate",
					"cloud_type":   "aliyun_bs",
					"zone_id":      "cn-beijing-h",
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v2/getHostDetail" {
					t.Fatalf("unexpected path %q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v3/getSubnetConfig" || q.Get("zone_id") != "cn-beijing-h" || q.Get("subnet_id") != "vsw-1" {
					t.Fatalf("unexpected subnet query path=%q query=%v", path, q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"subnets": []interface{}{
						map[string]interface{}{
							"id":           "vsw-1",
							"name":         "subnet-a",
							"display_name": "subnet-a (198.51.100.1/24)",
							"network_id":   "vpc-right",
							"zone_id":      "cn-beijing-h",
						},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v3/getCloudInfo" || q.Get("storage_type") != "HyperGate" || q.Get("network_id") != "vpc-right" {
					t.Fatalf("unexpected cloud info path=%q query=%v", path, q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"networks": []interface{}{
							map[string]interface{}{"id": "vpc-right", "name": "vpc-a", "display_name": "vpc-a (198.51.100.0/24)"},
						},
						"security_groups": []interface{}{
							map[string]interface{}{"id": "sg-1", "name": "sg-a", "display_name": "sg-a", "network_id": "vpc-right"},
						},
					},
				}}, nil
			},
		},
	}

	result, err := NewService(api).Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_account_id":  "account-1",
			"cloud_type":        "aliyun_bs",
			"storage_id":        "storage-1",
			"subnet_id":         "vsw-1",
			"security_group_id": "sg-1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	meta := result.Metadata
	if meta["zone_id"] != "cn-beijing-h" || meta["network_id"] != "vpc-right" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["network_name"] != "vpc-a" || meta["subnet_name"] != "subnet-a" || meta["security_group_name"] != "sg-a" {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestApplyBlockStorageKeepsExplicitSubnetAndNetworkInputs(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_type": "aliyun_bs",
					"cloud_account": map[string]interface{}{
						"region_id": "cn-beijing",
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"storage_type": "HyperGate",
					"cloud_type":   "aliyun_bs",
					"zone_id":      "cn-beijing-h",
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v3/getSubnetConfig" {
					t.Fatalf("unexpected path %q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"subnets": []interface{}{
						map[string]interface{}{"id": "vsw-1", "name": "subnet-a", "network_id": "vpc-right", "zone_id": "cn-beijing-h"},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v3/getCloudInfo" {
					t.Fatalf("unexpected path %q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"networks": []interface{}{
							map[string]interface{}{"id": "vpc-wrong", "name": "vpc-wrong"},
						},
						"security_groups": []interface{}{
							map[string]interface{}{"id": "sg-1", "name": "sg-a", "network_id": "vpc-wrong"},
						},
					},
				}}, nil
			},
		},
	}

	result, err := NewService(api).Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_account_id":  "account-1",
			"cloud_type":        "aliyun_bs",
			"storage_id":        "storage-1",
			"network_id":        "vpc-wrong",
			"subnet_id":         "vsw-1",
			"security_group_id": "sg-1",
		},
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if result.Metadata["network_id"] != "vpc-wrong" || result.Metadata["subnet_id"] != "vsw-1" {
		t.Fatalf("metadata = %+v", result.Metadata)
	}
}

func TestApplyBlockStorageKeepsExplicitSecurityGroupInput(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_type": "aliyun_bs",
					"cloud_account": map[string]interface{}{
						"region_id": "cn-beijing",
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"storage_type": "HyperGate",
					"cloud_type":   "aliyun_bs",
					"zone_id":      "cn-beijing-h",
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"subnets": []interface{}{
						map[string]interface{}{"id": "vsw-1", "network_id": "vpc-right", "zone_id": "cn-beijing-h"},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"networks": []interface{}{
							map[string]interface{}{"id": "vpc-right", "name": "vpc-a"},
						},
						"security_groups": []interface{}{
							map[string]interface{}{"id": "sg-1", "name": "sg-a", "network_id": "vpc-other"},
						},
					},
				}}, nil
			},
		},
	}

	result, err := NewService(api).Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_account_id":  "account-1",
			"cloud_type":        "aliyun_bs",
			"storage_id":        "storage-1",
			"subnet_id":         "vsw-1",
			"security_group_id": "sg-1",
		},
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if result.Metadata["security_group_id"] != "sg-1" || result.Metadata["network_id"] != "vpc-right" {
		t.Fatalf("metadata = %+v", result.Metadata)
	}
}

func TestApplyBlockStorageAutoSelectsCompatibleNetworkSubnetAndSecurityGroup(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_type": "aliyun_bs",
					"cloud_account": map[string]interface{}{
						"region_id": "cn-beijing",
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"storage_type": "HyperGate",
					"cloud_type":   "aliyun_bs",
					"zone_id":      "cn-beijing-h",
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v3/getSubnetConfig" || q.Get("zone_id") != "cn-beijing-h" {
					t.Fatalf("unexpected path=%q query=%v", path, q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"subnets": []interface{}{
						map[string]interface{}{"id": "vsw-1", "name": "subnet-a", "display_name": "subnet-a", "network_id": "vpc-right", "zone_id": "cn-beijing-h"},
					},
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v3/getCloudInfo" || q.Get("network_id") != "vpc-right" {
					t.Fatalf("unexpected path=%q query=%v", path, q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"networks": []interface{}{
							map[string]interface{}{"id": "vpc-right", "name": "vpc-a", "display_name": "vpc-a"},
						},
						"security_groups": []interface{}{
							map[string]interface{}{"id": "sg-1", "name": "sg-a", "display_name": "sg-a", "network_id": "vpc-right"},
						},
					},
				}}, nil
			},
		},
	}

	result, err := NewService(api).Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"cloud_account_id": "account-1",
			"cloud_type":       "aliyun_bs",
			"storage_id":       "storage-1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	meta := result.Metadata
	if meta["subnet_id"] != "vsw-1" || meta["network_id"] != "vpc-right" || meta["security_group_id"] != "sg-1" {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestApplyAutofillsRepairHostMapperDefaults(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID:  "host-1",
		Dynamic: map[string]string{"cloud_type": "aliyun_obs"},
	})
	if err != nil {
		t.Fatal(err)
	}
	mapper := result.Metadata["repair_host_mapper"].(map[string]interface{})
	if mapper["enable_dhcp_mode"] != "1" || mapper["enable_inject_driver"] != "1" || mapper["enable_repair_fs"] != "1" {
		t.Fatalf("mapper = %+v", mapper)
	}
	if mapper["os_version"] != "auto_check" || mapper["os_display_name"] != "auto_check" {
		t.Fatalf("mapper = %+v", mapper)
	}
}

func TestApplyGeneratesDiskVolumeMapperFromVolumeTypeOnly(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"id": "host-1",
					"disks": []interface{}{
						map[string]interface{}{"disk_id": "disk-1"},
						map[string]interface{}{"disk_id": "disk-2"},
					},
				}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID: "host-1",
		Dynamic: map[string]string{
			"volume_type_id": "cloud_essd_entry",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	rows := result.Metadata["disk_volume_mapper"].([]interface{})
	for _, item := range rows {
		row := item.(map[string]interface{})
		if row["volume_type_id"] != "cloud_essd_entry" {
			t.Fatalf("row = %+v", row)
		}
	}
}

func TestApplyKeepsExistingDiskMapperFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata.json")
	body := `{"volume_type_id":"cloud_data","disk_volume_mapper":[{"index":0,"disk_id":"custom-boot","volume_type_id":"custom-row"},{"index":1}]}`
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}

	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"id": "host-1",
					"disks": []interface{}{
						map[string]interface{}{"disk_id": "disk-boot", "is_boot_disk": true},
						map[string]interface{}{"disk_id": "disk-data", "is_boot_disk": false},
					},
				}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{HostID: "host-1", File: path})
	if err != nil {
		t.Fatal(err)
	}
	rows := result.Metadata["disk_volume_mapper"].([]interface{})
	if rows[0].(map[string]interface{})["disk_id"] != "custom-boot" || rows[0].(map[string]interface{})["volume_type_id"] != "custom-row" {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[1].(map[string]interface{})["disk_id"] != "disk-data" || rows[1].(map[string]interface{})["volume_type_id"] != "cloud_data" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestApplyUpdatesWhenBootConfigExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(path, []byte(`{"storage_id":"storage-1","repair_host_mapper":{"enable_dhcp_mode":"0","enable_inject_driver":"1"}}`), 0600); err != nil {
		t.Fatal(err)
	}

	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"storage_type": "HyperGate"}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"boot_config": map[string]interface{}{"id": "cfg-1"}}}, nil
			},
		},
		posts: []func(string, interface{}) (client.APIResponse, error){
			func(path string, body interface{}) (client.APIResponse, error) {
				if path != "/api/v2/batchGetBootConfigs" {
					t.Fatalf("unexpected post path %q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"boot_configs": []interface{}{
						map[string]interface{}{
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
				}}, nil
			},
			func(path string, body interface{}) (client.APIResponse, error) {
				if path != "/api/v2/batchUpdateBootConfigs" {
					t.Fatalf("unexpected post path %q", path)
				}
				return client.APIResponse{}, nil
			},
		},
	}
	service := NewService(api)
	result, err := service.Apply(ApplyInput{HostID: "host-1", File: path})
	if err != nil {
		t.Fatal(err)
	}
	if result.Operation != "update" || api.postPath != "/api/v2/batchUpdateBootConfigs" {
		t.Fatalf("result=%+v postPath=%q", result, api.postPath)
	}
	item := api.postBody.(map[string]interface{})["batch_update"].([]map[string]interface{})[0]
	if item["id"] != "cfg-1" || item["migration_id"] != "host-1" {
		t.Fatalf("body = %+v", api.postBody)
	}
	meta := item["metadata"].(map[string]interface{})
	if len(meta) != 1 {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["repair_host_mapper"].(map[string]interface{})["enable_dhcp_mode"] != "0" {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestPrepareRequestUpdatePreviewUsesTopLevelDiff(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(path, []byte(`{"storage_id":"storage-1","repair_host_mapper":{"enable_dhcp_mode":"0","enable_inject_driver":"1"}}`), 0600); err != nil {
		t.Fatal(err)
	}

	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"storage_type": "HyperGate"}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"boot_config": map[string]interface{}{"id": "cfg-1"}}}, nil
			},
		},
		posts: []func(string, interface{}) (client.APIResponse, error){
			func(path string, body interface{}) (client.APIResponse, error) {
				if path != "/api/v2/batchGetBootConfigs" {
					t.Fatalf("unexpected post path %q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"boot_configs": []interface{}{
						map[string]interface{}{
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
				}}, nil
			},
		},
	}

	prepared, err := NewService(api).PrepareRequest(ApplyInput{HostID: "host-1", File: path})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.NoOp {
		t.Fatalf("prepared = %+v", prepared)
	}
	item := prepared.Body["batch_update"].([]map[string]interface{})[0]
	meta := item["metadata"].(map[string]interface{})
	if len(meta) != 1 {
		t.Fatalf("metadata = %+v", meta)
	}
	if _, ok := meta["repair_host_mapper"]; !ok {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestApplyReturnsNoOpWhenMetadataMatchesExistingBootConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(path, []byte(`{"storage_id":"storage-1","repair_host_mapper":{"enable_dhcp_mode":"1","enable_inject_driver":"1"}}`), 0600); err != nil {
		t.Fatal(err)
	}

	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"storage_type": "HyperGate"}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"boot_config_id": "cfg-1"}}, nil
			},
		},
		posts: []func(string, interface{}) (client.APIResponse, error){
			func(path string, body interface{}) (client.APIResponse, error) {
				if path != "/api/v2/batchGetBootConfigs" {
					t.Fatalf("unexpected post path %q", path)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"boot_configs": []interface{}{
						map[string]interface{}{
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
				}}, nil
			},
		},
	}

	result, err := NewService(api).Apply(ApplyInput{HostID: "host-1", File: path})
	if err != nil {
		t.Fatal(err)
	}
	if !result.NoOp || result.Operation != "update" {
		t.Fatalf("result = %+v", result)
	}
	if api.postPath != "/api/v2/batchGetBootConfigs" {
		t.Fatalf("postPath = %q", api.postPath)
	}
}

func TestApplyRejectsInvalidMetadataShapes(t *testing.T) {
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
			path := filepath.Join(dir, "metadata.json")
			if err := os.WriteFile(path, []byte(tc.raw), 0600); err != nil {
				t.Fatal(err)
			}
			service := NewService(&fakeAPI{})
			_, err := service.Apply(ApplyInput{HostID: "host-1", File: path})
			if err == nil || err.Error() != tc.want {
				t.Fatalf("err = %v want %q", err, tc.want)
			}
		})
	}
}

func TestApplyCreatesWithoutFileWhenOverridesProvided(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
		},
	}

	service := NewService(api)
	result, err := service.Apply(ApplyInput{
		HostID:  "host-1",
		Dynamic: map[string]string{"cloud_type": "aliyun_bs", "region_id": "cn-beijing"},
		Sets:    []string{"repair_host_mapper.enable_dhcp_mode=1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Operation != "create" || api.postPath != "/api/v2/batchBootConfigs" {
		t.Fatalf("result=%+v postPath=%q", result, api.postPath)
	}
	item := api.postBody.(map[string]interface{})["batch_create"].([]map[string]interface{})[0]
	meta := item["metadata"].(map[string]interface{})
	if meta["cloud_type"] != "aliyun_bs" || meta["region_id"] != "cn-beijing" {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta["repair_host_mapper"].(map[string]interface{})["enable_dhcp_mode"] != 1 {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestApplyRejectsMissingMetadataSources(t *testing.T) {
	service := NewService(&fakeAPI{})
	_, err := service.Apply(ApplyInput{HostID: "host-1"})
	if err == nil || err.Error() != "file is required when no metadata override flags are provided" {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyAcceptsUTF8BOMMetadataFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata-bom.json")
	body := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"cloud_type":"aliyun_bs"}`)...)
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}

	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"id": "host-1"}}, nil
			},
		},
	}
	service := NewService(api)
	if _, err := service.Apply(ApplyInput{HostID: "host-1", File: path}); err != nil {
		t.Fatal(err)
	}
}

func TestApplyRejectsBadPathAndJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(path, []byte(`{"nics":[{"subnet_id":"vsw-1"}],"repair_host_mapper":{}}`), 0600); err != nil {
		t.Fatal(err)
	}

	service := NewService(&fakeAPI{})
	_, err := service.Apply(ApplyInput{HostID: "host-1", File: path, Sets: []string{"nics[2].subnet_id=vsw-2"}})
	if err == nil || err.Error() != "path nics[2].subnet_id index 2 out of range" {
		t.Fatalf("err = %v", err)
	}

	_, err = service.Apply(ApplyInput{HostID: "host-1", File: path, Sets: []string{"repair_host_mapper[0].foo=1"}})
	if err == nil || err.Error() != "path repair_host_mapper[0].foo expects array at repair_host_mapper" {
		t.Fatalf("err = %v", err)
	}

	_, err = service.Apply(ApplyInput{HostID: "host-1", File: path, SetJSONs: []string{`repair_host_mapper={bad}`}})
	if err == nil || !strings.HasPrefix(err.Error(), "invalid JSON for repair_host_mapper:") {
		t.Fatalf("err = %v", err)
	}
}

func TestMutationViewAddsOperation(t *testing.T) {
	view := MutationView(MutationResult{
		Operation: "create",
		Response:  client.APIResponse{Data: map[string]interface{}{"status": "ok"}},
	})
	if view["operation"] != "create" || view["status"] != "ok" {
		t.Fatalf("view = %+v", view)
	}
}

func TestMutationViewAddsNoOpMetadata(t *testing.T) {
	view := MutationView(MutationResult{
		Operation:    "update",
		HostID:       "host-1",
		BootConfigID: "cfg-1",
		NoOp:         true,
	})
	if view["operation"] != "update" || view["no_op"] != true || view["migration_id"] != "host-1" || view["boot_config_id"] != "cfg-1" {
		t.Fatalf("view = %+v", view)
	}
}
