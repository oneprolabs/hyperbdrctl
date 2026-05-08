package blockstoragecreate

import (
	"strings"
	"testing"
)

type noopFetcher struct{}

func (noopFetcher) FetchGatewayCloudInfo(accountID string, resources []string, regionID, zoneID, flavorID, flavorVCPUs, flavorRAM, purpose string) (map[string]interface{}, error) {
	return nil, nil
}

func (noopFetcher) FetchGatewayTransitionImages(accountID, cloudType, regionID, zoneID, purpose, imageType, osType, bootMode string) (map[string]interface{}, error) {
	return nil, nil
}

func (noopFetcher) FetchOpenStackCloudInfo(accountID string) (map[string]interface{}, error) {
	return nil, nil
}

func TestBuildRequestGenericProviderFallsBackToGenericBuilder(t *testing.T) {
	path, body, err := BuildRequest(noopFetcher{}, Spec{
		CloudType:          "huawei_bs",
		CloudAccountID:     "account-1",
		RegionID:           "cn-north-4",
		NetworkID:          "network-1",
		BootTypesID:        "boot_from_volume",
		VolumeProxyType:    "s3",
		HGControlNetwork:   "floating_ip_without_proxy",
		HGDataNetwork:      "floating_ip_without_proxy",
		HDControlNetwork:   "floating_ip_with_hg_proxy",
		BootLoaderImageID:  "boot-image-1",
		BootLoaderFlavorID: "boot-flavor-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/hypermotion/v1/storages/action" {
		t.Fatalf("path = %q", path)
	}
	createStorage := body["create_storage"].(map[string]interface{})
	if createStorage["cloud_type"] != "huawei_bs" || createStorage["cloud_account_uuid"] != "account-1" || createStorage["type"] != "HyperGate" {
		t.Fatalf("create_storage = %#v", createStorage)
	}
	metadata := createStorage["metadata"].(map[string]interface{})
	for key, want := range map[string]string{
		"region_id":             "cn-north-4",
		"network_id":            "network-1",
		"boot_types_id":         "boot_from_volume",
		"volume_proxy_type":     "s3",
		"hg_control_network":    "floating_ip_without_proxy",
		"hg_data_network":       "floating_ip_without_proxy",
		"hd_control_network":    "floating_ip_with_hg_proxy",
		"boot_loader_image_id":  "boot-image-1",
		"boot_loader_flavor_id": "boot-flavor-1",
	} {
		if metadata[key] != want {
			t.Fatalf("metadata[%q] = %#v, want %q", key, metadata[key], want)
		}
	}
}

func TestBuildRequestGenericRequiresCloudType(t *testing.T) {
	_, _, err := BuildRequest(noopFetcher{}, Spec{CloudAccountID: "account-1"})
	if err == nil || !strings.Contains(err.Error(), "cloud-type is required") {
		t.Fatalf("err = %v", err)
	}
}

type aliyunFetcher struct{}

func (aliyunFetcher) FetchGatewayCloudInfo(accountID string, resources []string, regionID, zoneID, flavorID, flavorVCPUs, flavorRAM, purpose string) (map[string]interface{}, error) {
	switch len(resources) {
	case 2:
		if resources[0] == "regions" {
			return map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"domain": map[string]interface{}{
						"regions": []interface{}{
							map[string]interface{}{"region_id": "cn-beijing", "region_name": "North China 2 (Beijing)"},
						},
					},
					"zones": []interface{}{
						map[string]interface{}{"id": "cn-beijing-l", "display_name": "Beijing Zone L"},
					},
				},
			}, nil
		}
		if resources[0] == "images" {
			return map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"images": []interface{}{
						map[string]interface{}{"id": "img-1", "name": "ubuntu", "os_type": "linux"},
					},
					"system_disk_types": []interface{}{
						map[string]interface{}{"id": "cloud_essd_entry", "display_name": "ESSD Entry"},
					},
				},
			}, nil
		}
	}
	if len(resources) == 1 && resources[0] == "flavors" {
		return map[string]interface{}{
			"cloud_info": map[string]interface{}{
				"flavors": []interface{}{
					map[string]interface{}{"id": "ecs.e-c1m2.large", "name": "2C4G", "is_recommend": 1},
				},
			},
		}, nil
	}
	if len(resources) == 3 && resources[0] == "networks" {
		return map[string]interface{}{
			"cloud_info": map[string]interface{}{
				"networks": []interface{}{
					map[string]interface{}{"id": "net-1", "name": "vpc-a"},
				},
				"subnets": []interface{}{
					map[string]interface{}{"id": "subnet-1", "name": "subnet-a", "network_id": "net-1"},
				},
			},
		}, nil
	}
	return nil, nil
}

func (aliyunFetcher) FetchGatewayTransitionImages(accountID, cloudType, regionID, zoneID, purpose, imageType, osType, bootMode string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"cloud_info": map[string]interface{}{
			"images": []interface{}{
				map[string]interface{}{"image_id": "boot-img-1", "image_name": "Windows 2016"},
			},
		},
	}, nil
}

func (aliyunFetcher) FetchOpenStackCloudInfo(accountID string) (map[string]interface{}, error) {
	return nil, nil
}

func TestBuildRequestAliyunUsesDefaultsForBootLoaderAndFixedNetworkBandwidth(t *testing.T) {
	_, body, err := BuildRequest(aliyunFetcher{}, Spec{
		CloudType:        "aliyun_bs",
		CloudAccountID:   "account-1",
		RegionID:         "cn-beijing",
		HGControlNetwork: "fixed_ip_without_proxy",
		HGDataNetwork:    "fixed_ip_without_proxy",
	})
	if err != nil {
		t.Fatal(err)
	}

	createStorage := body["create_storage"].(map[string]interface{})
	metadata := createStorage["metadata"].(map[string]interface{})
	if metadata["win_hd_image_id"] != "boot-img-1" {
		t.Fatalf("win_hd_image_id = %#v", metadata["win_hd_image_id"])
	}
	if metadata["bandwidth_size"] != 50 {
		t.Fatalf("bandwidth_size = %#v", metadata["bandwidth_size"])
	}
}
