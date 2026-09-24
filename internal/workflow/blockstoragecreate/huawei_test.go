package blockstoragecreate

import (
	"reflect"
	"strings"
	"testing"
)

type huaweiFetcher struct {
	calls []string
}

func (f *huaweiFetcher) FetchGatewayCloudInfo(accountID string, resources []string, regionID, zoneID, flavorID, flavorVCPUs, flavorRAM, purpose string) (map[string]interface{}, error) {
	f.calls = append(f.calls, strings.Join(resources, ","))
	switch strings.Join(resources, ",") {
	case "regions,zones":
		return map[string]interface{}{
			"cloud_info": map[string]interface{}{
				"domain": map[string]interface{}{
					"regions": []interface{}{
						map[string]interface{}{"region_id": "cn-north-1", "region_name": "cn-north-1"},
					},
				},
				"zones": []interface{}{
					map[string]interface{}{"id": "cn-north-1a", "name": "cn-north-1a"},
				},
			},
		}, nil
	case "flavors":
		return map[string]interface{}{
			"cloud_info": map[string]interface{}{
				"flavors": []interface{}{
					map[string]interface{}{
						"id": "s2.large.2", "name": "s2.large.2", "flavor_vcpus": 2, "flavor_ram": 4,
						"GHz": "Intel E5-2680V4 2.4GHz", "quota_rate": "0.2 / 0.8 Gbit/s", "quota_pps": "100,000 PPS", "is_recommend": 1,
					},
				},
			},
		}, nil
	case "images,system_volume_types":
		return map[string]interface{}{
			"cloud_info": map[string]interface{}{
				"images": []interface{}{
					map[string]interface{}{"image_id": "image-1", "image_name": "Ubuntu 24.04 server 64bit", "os_type": "linux"},
				},
				"system_volume_types": []interface{}{
					map[string]interface{}{"id": "disk-type-1", "name": "SAS"},
				},
			},
		}, nil
	case "networks,subnets,abilities":
		return map[string]interface{}{
			"cloud_info": map[string]interface{}{
				"networks": []interface{}{
					map[string]interface{}{"id": "network-1", "name": "vpc-ray"},
				},
				"subnets": []interface{}{
					map[string]interface{}{"id": "subnet-1", "name": "subnet-ray", "network_id": "network-1"},
				},
			},
		}, nil
	default:
		return nil, nil
	}
}

func (f *huaweiFetcher) FetchGatewayTransitionImages(accountID, cloudType, regionID, zoneID, purpose, imageType, osType, bootMode string) (map[string]interface{}, error) {
	f.calls = append(f.calls, "transition-images")
	return map[string]interface{}{
		"cloud_info": map[string]interface{}{
			"images": []interface{}{
				map[string]interface{}{"image_id": "win-image-1", "image_name": "Windows repair image"},
			},
		},
	}, nil
}

func (f *huaweiFetcher) FetchOpenStackCloudInfo(accountID string) (map[string]interface{}, error) {
	return nil, nil
}

func TestBuildRequestHuaweiUsesDedicatedResourceFlowAndDefaults(t *testing.T) {
	fetcher := &huaweiFetcher{}
	_, body, err := BuildRequest(fetcher, Spec{
		CloudType:      "huawei_bs",
		CloudAccountID: "account-1",
		RegionID:       "cn-north-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	wantCalls := []string{"regions,zones", "flavors", "images,system_volume_types", "networks,subnets,abilities", "transition-images"}
	if !reflect.DeepEqual(fetcher.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", fetcher.calls, wantCalls)
	}

	createStorage := body["create_storage"].(map[string]interface{})
	metadata := createStorage["metadata"].(map[string]interface{})
	for key, want := range map[string]interface{}{
		"region_id": "cn-north-1", "region_name": "cn-north-1",
		"zone_id": "cn-north-1a", "zone_name": "cn-north-1a",
		"flavor_id": "s2.large.2", "flavor_name": "s2.large.2", "vcpu_id": "2 CPU", "ram_GB": "4 GiB",
		"GHz": "Intel E5-2680V4 2.4GHz", "quota_rate_name": "0.2 / 0.8 Gbit/s", "quota_pps_name": "100,000 PPS",
		"image_id": "image-1", "image_name": "Ubuntu 24.04 server 64bit",
		"system_disk_type_id": "disk-type-1", "system_disk_type_name": "SAS",
		"network_id": "network-1", "network_name": "vpc-ray", "subnet_id": "subnet-1", "subnet_name": "subnet-ray",
		"volume_proxy_type": "s3", "volume_proxy_type_name": "S3Block", "dest_device_type": "vbd",
		"hg_control_network": "floating_ip_without_proxy", "hg_data_network": "floating_ip_without_proxy",
		"bandwidth_size_type": "traffic", "hd_control_network": "floating_ip_with_hg_proxy",
		"win_hd_image_id": "win-image-1", "win_hd_image_name": "Windows repair image",
		"win_hd_access_type": "hg_password", "win_hd_access_type_name": "\u9ed8\u8ba4", "win_hd_username": "", "win_hd_password": "",
	} {
		if metadata[key] != want {
			t.Fatalf("metadata[%q] = %#v, want %#v; metadata=%#v", key, metadata[key], want, metadata)
		}
	}
	for _, key := range []string{
		"boot_types_id", "dest_device_type_name", "bandwidth_size_type_name",
		"system_disk_size", "bandwidth_size", "fixed_ip", "control_nat_ip", "data_nat_ip",
		"hg_control_network_name", "hg_data_network_name", "region_display_name", "zone_display_name",
	} {
		if _, ok := metadata[key]; ok {
			t.Fatalf("metadata must omit %q: %#v", key, metadata)
		}
	}
}

func TestBuildRequestHuaweiValidatesExplicitResourcesAndDynamicValueTypes(t *testing.T) {
	fetcher := &huaweiFetcher{}
	_, body, err := BuildRequest(fetcher, Spec{
		CloudType:          "huawei_bs",
		CloudAccountID:     "account-1",
		RegionID:           "cn-north-1",
		ZoneID:             "cn-north-1a",
		FlavorID:           "s2.large.2",
		ImageID:            "image-1",
		SystemDiskTypeID:   "disk-type-1",
		NetworkID:          "network-1",
		SubnetID:           "subnet-1",
		SystemDiskSize:     "40",
		BandwidthSize:      "300",
		FixedIP:            "192.0.2.8",
		ControlNATIP:       "192.0.2.10",
		DataNATIP:          "192.0.2.11",
		BootLoaderImageID:  "win-image-1",
		BootLoaderFlavorID: "boot-flavor-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	metadata := body["create_storage"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["system_disk_size"] != "40" || metadata["bandwidth_size"] != 300 {
		t.Fatalf("dynamic value types are wrong: %#v", metadata)
	}
	for key, want := range map[string]interface{}{
		"fixed_ip": "192.0.2.8", "control_nat_ip": "192.0.2.10", "data_nat_ip": "192.0.2.11",
		"boot_loader_image_id": "win-image-1", "boot_loader_flavor_id": "boot-flavor-1",
	} {
		if metadata[key] != want {
			t.Fatalf("metadata[%q] = %#v, want %#v", key, metadata[key], want)
		}
	}

	_, _, err = BuildRequest(&huaweiFetcher{}, Spec{
		CloudType: "huawei_bs", CloudAccountID: "account-1", RegionID: "cn-north-1", FlavorID: "missing",
	})
	if err == nil || !strings.Contains(err.Error(), `flavor-id: resource "missing" not found`) {
		t.Fatalf("err = %v", err)
	}
}
