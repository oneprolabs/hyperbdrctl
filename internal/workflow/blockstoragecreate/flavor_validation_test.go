package blockstoragecreate

import (
	"strings"
	"testing"
)

type minimumFlavorAliyunFetcher struct {
	flavors []map[string]interface{}
}

func (f minimumFlavorAliyunFetcher) FetchGatewayCloudInfo(accountID string, resources []string, regionID, zoneID, flavorID, flavorVCPUs, flavorRAM, purpose string) (map[string]interface{}, error) {
	switch {
	case len(resources) == 2 && resources[0] == "regions":
		return map[string]interface{}{
			"cloud_info": map[string]interface{}{
				"domain": map[string]interface{}{
					"regions": []interface{}{
						map[string]interface{}{"region_id": "cn-beijing", "region_name": "North China 2 (Beijing)"},
					},
				},
				"zones": []interface{}{
					map[string]interface{}{"id": "cn-beijing-h", "display_name": "Beijing Zone H"},
				},
			},
		}, nil
	case len(resources) == 1 && resources[0] == "flavors":
		rows := make([]interface{}, 0, len(f.flavors))
		for _, flavor := range f.flavors {
			rows = append(rows, flavor)
		}
		return map[string]interface{}{
			"cloud_info": map[string]interface{}{
				"flavors": rows,
			},
		}, nil
	case len(resources) == 2 && resources[0] == "images":
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
	case len(resources) == 3 && resources[0] == "networks":
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
	default:
		return nil, nil
	}
}

func (minimumFlavorAliyunFetcher) FetchGatewayTransitionImages(accountID, cloudType, regionID, zoneID, purpose, imageType, osType, bootMode string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"cloud_info": map[string]interface{}{
			"images": []interface{}{
				map[string]interface{}{"image_id": "boot-img-1", "image_name": "Windows 2016"},
			},
		},
	}, nil
}

func (minimumFlavorAliyunFetcher) FetchOpenStackCloudInfo(accountID string) (map[string]interface{}, error) {
	return nil, nil
}

type minimumFlavorOpenStackFetcher struct {
	flavors []map[string]interface{}
}

func (minimumFlavorOpenStackFetcher) FetchGatewayCloudInfo(accountID string, resources []string, regionID, zoneID, flavorID, flavorVCPUs, flavorRAM, purpose string) (map[string]interface{}, error) {
	return nil, nil
}

func (minimumFlavorOpenStackFetcher) FetchGatewayTransitionImages(accountID, cloudType, regionID, zoneID, purpose, imageType, osType, bootMode string) (map[string]interface{}, error) {
	return nil, nil
}

func (f minimumFlavorOpenStackFetcher) FetchOpenStackCloudInfo(accountID string) (map[string]interface{}, error) {
	rows := make([]interface{}, 0, len(f.flavors))
	for _, flavor := range f.flavors {
		rows = append(rows, flavor)
	}
	return map[string]interface{}{
		"auth_info": map[string]interface{}{
			"project_domain_id": "default",
			"project_id":        "project-1",
			"project_name":      "autotest",
			"region_id":         "RegionOne",
			"region_name":       "RegionOne",
		},
		"projects": []interface{}{
			map[string]interface{}{"id": "project-1", "name": "autotest"},
		},
		"regions": []interface{}{
			map[string]interface{}{"id": "RegionOne", "name": "RegionOne"},
		},
		"compute_zones": []interface{}{
			map[string]interface{}{"id": "nova", "name": "nova"},
		},
		"images": []interface{}{
			map[string]interface{}{"id": "img-1", "name": "ubuntu24.04", "os_type": "linux"},
		},
		"flavors": rows,
		"networks": []interface{}{
			map[string]interface{}{"id": "net-1", "name": "public-network-10", "display_name": "public-network-10"},
		},
		"subnets": []interface{}{
			map[string]interface{}{"id": "", "name": "default", "network_id": "net-1"},
		},
		"volume_types": []interface{}{
			map[string]interface{}{"id": "DEFAULT_VOLUME_TYPE", "name": "DEFAULT_VOLUME_TYPE"},
		},
		"boot_loader_images": []interface{}{
			map[string]interface{}{"id": "boot-img-1", "name": "Windows_DriverFix"},
		},
		"boot_loader_flavors": []interface{}{
			map[string]interface{}{"id": "boot-flavor-1", "name": "2C4G", "is_recommend": 1},
		},
	}, nil
}

func TestBuildRequestAliyunSkipsUndersizedFlavorCandidates(t *testing.T) {
	_, body, err := BuildRequest(minimumFlavorAliyunFetcher{
		flavors: []map[string]interface{}{
			{"id": "ecs.small", "name": "ecs.small(1C2G)", "vcpus": 1, "ram_GB": 2, "is_recommend": 1},
			{"id": "ecs.medium", "name": "ecs.medium(2C4G)", "vcpus": 2, "ram_GB": 4},
		},
	}, Spec{
		CloudType:      "aliyun_bs",
		CloudAccountID: "account-1",
		RegionID:       "cn-beijing",
		ZoneID:         "cn-beijing-h",
	})
	if err != nil {
		t.Fatal(err)
	}

	createStorage := body["create_storage"].(map[string]interface{})
	metadata := createStorage["metadata"].(map[string]interface{})
	if metadata["flavor_id"] != "ecs.medium" {
		t.Fatalf("flavor_id = %#v", metadata["flavor_id"])
	}
}

func TestBuildRequestAliyunRejectsExplicitUndersizedFlavor(t *testing.T) {
	_, _, err := BuildRequest(minimumFlavorAliyunFetcher{
		flavors: []map[string]interface{}{
			{"id": "ecs.small", "name": "ecs.small(1C2G)", "vcpus": 1, "ram_GB": 2},
			{"id": "ecs.medium", "name": "ecs.medium(2C4G)", "vcpus": 2, "ram_GB": 4},
		},
	}, Spec{
		CloudType:      "aliyun_bs",
		CloudAccountID: "account-1",
		RegionID:       "cn-beijing",
		ZoneID:         "cn-beijing-h",
		FlavorID:       "ecs.small",
	})
	if err == nil || !strings.Contains(err.Error(), "at least 2 vCPUs and 4 GiB RAM") {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRequestOpenStackRejectsUndersizedFlavorByName(t *testing.T) {
	_, _, err := BuildRequest(minimumFlavorOpenStackFetcher{
		flavors: []map[string]interface{}{
			{"id": "flavor-small", "name": "1C2G"},
			{"id": "flavor-medium", "name": "2C_4G_40G(2C4G)", "is_recommend": 1},
		},
	}, Spec{
		CloudType:         "openstack",
		CloudAccountID:    "account-1",
		BootLoaderImageID: "boot-img-1",
		FlavorID:          "flavor-small",
	})
	if err == nil || !strings.Contains(err.Error(), "at least 2 vCPUs and 4 GiB RAM") {
		t.Fatalf("err = %v", err)
	}
}
