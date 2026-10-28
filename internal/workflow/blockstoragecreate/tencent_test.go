package blockstoragecreate

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type tencentGatewayFetcher struct {
	calls []string
}

func (f *tencentGatewayFetcher) FetchGatewayCloudInfo(_ string, resources []string, regionID, zoneID, flavorID, _, _, purpose string) (map[string]interface{}, error) {
	f.calls = append(f.calls, strings.Join(resources, ","))
	switch strings.Join(resources, ",") {
	case "regions,zones":
		return map[string]interface{}{"cloud_info": map[string]interface{}{
			"regions": []interface{}{map[string]interface{}{"id": "ap-beijing", "name": "华北地区(北京)"}},
			"zones":   []interface{}{map[string]interface{}{"id": "ap-beijing-3", "name": "北京三区"}},
		}}, nil
	case "flavors":
		if regionID != "ap-beijing" || zoneID != "ap-beijing-3" || purpose != "make_hg" {
			return nil, errTencentUnexpectedContext
		}
		return map[string]interface{}{"cloud_info": map[string]interface{}{
			"flavors": []interface{}{map[string]interface{}{
				"id": "S4.MEDIUM4", "name": "S4.MEDIUM4", "vcpus": 2, "ram_GB": 4,
				"is_recommend": 1,
			}},
		}}, nil
	case "images,system_disk_types":
		if flavorID != "S4.MEDIUM4" {
			return nil, errTencentUnexpectedContext
		}
		return map[string]interface{}{"cloud_info": map[string]interface{}{
			"images": []interface{}{map[string]interface{}{"id": "img-ubuntu", "name": "Ubuntu 24.04", "os_type": "linux"}},
			"system_disk_types": []interface{}{
				map[string]interface{}{"id": "CLOUD_HSSD", "display_name": "Enhanced SSD Cloud Disk"},
				map[string]interface{}{"id": "CLOUD_BSSD", "display_name": "Balanced SSD Cloud Disk"},
			},
		}}, nil
	case "networks,subnets,abilities":
		return map[string]interface{}{"cloud_info": map[string]interface{}{
			"networks": []interface{}{map[string]interface{}{"id": "vpc-1", "name": "vpc-beijing", "is_recommend": 1}},
			"subnets": []interface{}{
				map[string]interface{}{"id": "subnet-other-zone", "name": "other", "network_id": "vpc-1", "zone_id": "ap-beijing-4"},
				map[string]interface{}{"id": "subnet-3", "name": "BeijingZone3", "network_id": "vpc-1", "zone_id": "ap-beijing-3"},
			},
		}}, nil
	default:
		return nil, errTencentUnexpectedResources
	}
}

func (f *tencentGatewayFetcher) FetchGatewayTransitionImages(_ string, cloudType, regionID, zoneID, purpose, imageType, osType, bootMode string) (map[string]interface{}, error) {
	f.calls = append(f.calls, "transition-images")
	if cloudType != "tencent_bs" || regionID != "ap-beijing" || zoneID != "ap-beijing-3" || purpose != "rebuild_bcd" || imageType != "system" || osType != "windows" || bootMode != "bios" {
		return nil, errTencentUnexpectedContext
	}
	return map[string]interface{}{"cloud_info": map[string]interface{}{
		"images": []interface{}{map[string]interface{}{"image_id": "img-win", "image_name": "Windows Server 2019"}},
	}}, nil
}

func (*tencentGatewayFetcher) FetchOpenStackCloudInfo(string) (map[string]interface{}, error) {
	return nil, nil
}

var (
	errTencentUnexpectedContext   = errors.New("unexpected Tencent gateway resource context")
	errTencentUnexpectedResources = errors.New("unexpected Tencent gateway resources")
)

func TestBuildRequestTencentUsesCloudAccountResourcesAndUICompatibleDefaults(t *testing.T) {
	fetcher := &tencentGatewayFetcher{}
	path, body, err := BuildRequest(fetcher, Spec{
		CloudType:        "tencent_bs",
		CloudAccountID:   "account-tencent",
		RegionID:         "ap-beijing",
		HGControlNetwork: "floating_ip_without_proxy",
		HGDataNetwork:    "floating_ip_without_proxy",
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/hypermotion/v1/storages/action" {
		t.Fatalf("path = %q", path)
	}
	if want := []string{"regions,zones", "flavors", "images,system_disk_types", "networks,subnets,abilities", "transition-images"}; !reflect.DeepEqual(fetcher.calls, want) {
		t.Fatalf("resource calls = %#v, want %#v", fetcher.calls, want)
	}

	createStorage := body["create_storage"].(map[string]interface{})
	if createStorage["type"] != "HyperGate" || createStorage["cloud_type"] != "tencent_bs" || createStorage["cloud_account_uuid"] != "account-tencent" {
		t.Fatalf("create_storage = %#v", createStorage)
	}
	metadata := createStorage["metadata"].(map[string]interface{})
	for key, want := range map[string]interface{}{
		"region_id": "ap-beijing", "region_name": "华北地区(北京)",
		"zone_id": "ap-beijing-3", "zone_name": "北京三区",
		"flavor_id": "S4.MEDIUM4", "flavor_name": "S4.MEDIUM4", "vcpu_id": "2 CPU", "ram_GB": "4 GiB",
		"GHz": "暂时无法获取", "image_id": "img-ubuntu", "image_name": "Ubuntu 24.04",
		"system_disk_type_id": "CLOUD_BSSD", "system_disk_type_name": "Balanced SSD Cloud Disk", "system_disk_size": "50",
		"volume_proxy_type": "s3", "volume_proxy_type_name": "S3Block", "dest_host_boot_type": "image",
		"hg_control_network": "floating_ip_without_proxy", "hg_control_network_name": "公网",
		"hg_data_network": "floating_ip_without_proxy", "hg_data_network_name": "公网",
		"bandwidth_size": 100, "hd_control_network": "fixed_ip_with_hg_proxy", "hd_control_network_name": "内网网络并通过云同步网关代理",
		"network_id": "vpc-1", "network_name": "vpc-beijing", "subnet_id": "subnet-3", "subnet_name": "BeijingZone3",
		"fixed_ip": "", "win_hd_image_id": "img-win", "win_hd_image_name": "Windows Server 2019",
	} {
		if metadata[key] != want {
			t.Errorf("metadata[%q] = %#v, want %#v", key, metadata[key], want)
		}
	}
	if metadata["control_nat_ip"] != nil || metadata["data_nat_ip"] != nil {
		t.Errorf("NAT metadata = %#v", metadata)
	}
}

func TestBuildRequestTencentHonorsExplicitDriverNetworkMode(t *testing.T) {
	_, body, err := BuildRequest(&tencentGatewayFetcher{}, Spec{
		CloudType: "tencent_bs", CloudAccountID: "account-tencent", RegionID: "ap-beijing",
		HDControlNetwork: "floating_ip_with_hg_proxy", ExplicitMetadataKeys: []string{"hd_control_network"},
	})
	if err != nil {
		t.Fatal(err)
	}
	metadata := body["create_storage"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["hd_control_network"] != "floating_ip_with_hg_proxy" {
		t.Fatalf("hd_control_network = %#v", metadata["hd_control_network"])
	}
}

func TestBuildRequestTencentRequiresCloudAccountRegion(t *testing.T) {
	_, _, err := BuildRequest(&tencentGatewayFetcher{}, Spec{CloudType: "tencent_bs", CloudAccountID: "account-tencent"})
	if err == nil || !strings.Contains(err.Error(), "region-id is required in cloud account metadata") {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRequestTencentValidatesSelectedSubnetAgainstZone(t *testing.T) {
	_, _, err := BuildRequest(&tencentGatewayFetcher{}, Spec{
		CloudType: "tencent_bs", CloudAccountID: "account-tencent", RegionID: "ap-beijing",
		ZoneID: "ap-beijing-3", SubnetID: "subnet-other-zone",
	})
	if err == nil || !strings.Contains(err.Error(), "subnet-id: resource \"subnet-other-zone\" not found") {
		t.Fatalf("err = %v", err)
	}
}
