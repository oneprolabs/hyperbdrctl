package blockstorage

import (
	"net/url"
	"strings"
	"testing"

	"hyperbdr-client/internal/client"
	workflowcreate "hyperbdr-client/internal/workflow/blockstoragecreate"
)

type createFakeAPI struct {
	getPath          string
	getQuery         url.Values
	sawAccountDetail bool
	postCalls        int
	postPath         string
	postBody         interface{}
	cloudType        string
	regionID         string
}

func (f *createFakeAPI) Get(path string, q url.Values) (client.APIResponse, error) {
	f.getPath = path
	f.getQuery = q
	if path == "/hypermotion/v1/cloud_accounts/account-1" {
		f.sawAccountDetail = true
		cloudType := f.cloudType
		if cloudType == "" {
			cloudType = "aliyun_bs"
		}
		regionID := f.regionID
		if regionID == "" {
			regionID = "cn-beijing"
		}
		return client.APIResponse{
			Raw: map[string]interface{}{
				"cloud_account": map[string]interface{}{
					"cloud_type":       cloudType,
					"region_type_list": regionID,
					"auth_region_id":   regionID,
				},
			},
		}, nil
	}
	if path == "/api/v3/getCloudInfo" {
		cloudInfo := map[string]interface{}{
			"images": []interface{}{
				map[string]interface{}{"image_id": "boot-img-1", "image_name": "Windows 2016"},
			},
		}
		if f.cloudType == "huawei_bs" && strings.Contains(q.Get("fetch_res"), "system_volume_types") {
			cloudInfo["images"] = []interface{}{map[string]interface{}{"id": "img-1", "name": "ubuntu", "os_type": "linux"}}
			cloudInfo["system_volume_types"] = []interface{}{map[string]interface{}{"id": "cloud_essd_entry"}}
		}
		return client.APIResponse{Data: map[string]interface{}{"cloud_info": cloudInfo}}, nil
	}
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func (f *createFakeAPI) Post(path string, body interface{}) (client.APIResponse, error) {
	f.postCalls++
	f.postPath = path
	f.postBody = body
	if path == "/hypermotion/v1/cloud_accounts/account-1/action" {
		switch f.postCalls {
		case 1:
			regionID := f.regionID
			if regionID == "" {
				regionID = "cn-beijing"
			}
			return client.APIResponse{Data: map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"domain": map[string]interface{}{
						"regions": []interface{}{
							map[string]interface{}{"region_id": regionID, "region_name": regionID},
						},
					},
					"zones": []interface{}{
						map[string]interface{}{"id": regionID + "a", "display_name": regionID + "a"},
					},
				},
			}}, nil
		case 2:
			return client.APIResponse{Data: map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"flavors": []interface{}{
						map[string]interface{}{"id": "ecs.e-c1m2.large", "name": "2C4G", "vcpus": 2, "ram_GB": 4, "is_recommend": 1},
					},
				},
			}}, nil
		case 3:
			if f.cloudType == "huawei_bs" {
				return client.APIResponse{Data: map[string]interface{}{
					"cloud_info": map[string]interface{}{
						"networks": []interface{}{map[string]interface{}{"id": "net-1", "name": "vpc-a"}},
						"subnets":  []interface{}{map[string]interface{}{"id": "subnet-1", "name": "subnet-a", "network_id": "net-1"}},
					},
				}}, nil
			}
			diskKey := "system_disk_types"
			if f.cloudType == "huawei_bs" {
				diskKey = "system_volume_types"
			}
			return client.APIResponse{Data: map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"images": []interface{}{
						map[string]interface{}{"id": "img-1", "name": "ubuntu", "os_type": "linux"},
					},
					diskKey: []interface{}{map[string]interface{}{"id": "cloud_essd_entry"}},
				},
			}}, nil
		case 4:
			return client.APIResponse{Data: map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"networks": []interface{}{
						map[string]interface{}{"id": "net-1", "name": "vpc-a"},
					},
					"subnets": []interface{}{
						map[string]interface{}{"id": "subnet-1", "name": "subnet-a", "network_id": "net-1"},
					},
				},
			}}, nil
		}
	}
	if path == "/hypermotion/v1/storages/action" {
		return client.APIResponse{Data: map[string]interface{}{"uuid": "storage-1"}}, nil
	}
	return client.APIResponse{Data: map[string]interface{}{"uuid": "storage-1"}}, nil
}

func TestServiceCreateUsesRawCloudAccountDetailForAliyunDefaults(t *testing.T) {
	api := &createFakeAPI{}
	service := NewService(api)

	_, err := service.Create(CreateSpec{
		CloudAccountID: "account-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !api.sawAccountDetail {
		t.Fatalf("account detail was not queried; last get path = %q", api.getPath)
	}
	if api.postPath != "/hypermotion/v1/storages/action" {
		t.Fatalf("post path = %q", api.postPath)
	}
}

func TestServicePrepareCreateUsesCloudAccountRegionForHuaweiDefaults(t *testing.T) {
	api := &createFakeAPI{cloudType: "huawei_bs", regionID: "cn-north-1"}
	service := NewService(api)

	prepared, err := service.PrepareCreate(CreateSpec{CloudAccountID: "account-1"})
	if err != nil {
		t.Fatal(err)
	}
	if !api.sawAccountDetail {
		t.Fatalf("account detail was not queried; last get path = %q", api.getPath)
	}
	metadata := prepared.Body["create_storage"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["region_id"] != "cn-north-1" {
		t.Fatalf("region_id = %#v, want %q", metadata["region_id"], "cn-north-1")
	}
}

func TestApplyCreateMetadataOverridesHonorsPrecedenceAndPaths(t *testing.T) {
	body := map[string]interface{}{
		"create_storage": map[string]interface{}{
			"type":               "HyperGate",
			"cloud_type":         "huawei_bs",
			"cloud_account_uuid": "account-1",
			"metadata": map[string]interface{}{
				"system_disk_size": "40",
				"nics":             []interface{}{map[string]interface{}{"subnet_id": "subnet-1"}},
			},
		},
	}
	spec := CreateSpec{
		JSONMetadataOverrides: []workflowcreate.MetadataOverride{
			{Path: "system_disk_size", Value: "20"},
			{Path: "nics", Value: []interface{}{map[string]interface{}{"subnet_id": "subnet-json"}}},
		},
		MetadataOverrides: []workflowcreate.MetadataOverride{
			{Path: "system_disk_size", Value: 30},
			{Path: "enabled", Value: true},
			{Path: "nics[0].subnet_id", Value: "subnet-set"},
		},
		ExplicitMetadataKeys: []string{"system_disk_size"},
	}
	if err := applyCreateMetadataOverrides(body, spec); err != nil {
		t.Fatal(err)
	}
	createStorage := body["create_storage"].(map[string]interface{})
	metadata := createStorage["metadata"].(map[string]interface{})
	if metadata["system_disk_size"] != "40" || metadata["enabled"] != true {
		t.Fatalf("metadata = %#v", metadata)
	}
	nics := metadata["nics"].([]interface{})
	if nics[0].(map[string]interface{})["subnet_id"] != "subnet-set" {
		t.Fatalf("metadata nics = %#v", nics)
	}
	if createStorage["type"] != "HyperGate" || createStorage["cloud_type"] != "huawei_bs" || createStorage["cloud_account_uuid"] != "account-1" {
		t.Fatalf("create storage root changed: %#v", createStorage)
	}

	spec.MetadataOverrides = []workflowcreate.MetadataOverride{{Path: "nics[1].subnet_id", Value: "invalid"}}
	if err := applyCreateMetadataOverrides(body, spec); err == nil {
		t.Fatal("out-of-range array path should fail")
	}
}
