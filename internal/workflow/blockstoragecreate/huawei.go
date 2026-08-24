package blockstoragecreate

import (
	"fmt"
	"strconv"
)

func fetchHuaweiGatewayCloudInfo(fetcher Fetcher, accountID string, resources []string, regionID, zoneID, flavorID, purpose string) (map[string]interface{}, error) {
	if huaweiFetcher, ok := fetcher.(HuaweiGatewayCloudInfoFetcher); ok {
		return huaweiFetcher.FetchHuaweiGatewayCloudInfo(accountID, resources, regionID, zoneID, flavorID, purpose)
	}
	return fetcher.FetchGatewayCloudInfo(accountID, resources, regionID, zoneID, flavorID, "", "", purpose)
}

func buildHuawei(fetcher Fetcher, spec Spec) (string, map[string]interface{}, error) {
	if spec.CloudAccountID == "" {
		return "", nil, fmt.Errorf("cloud-account-id is required")
	}
	if spec.RegionID == "" {
		return "", nil, fmt.Errorf("region-id is required in cloud account metadata")
	}

	regionsAndZones, err := fetcher.FetchGatewayCloudInfo(spec.CloudAccountID, []string{"regions", "zones"}, spec.RegionID, "", "", "", "", "make_hg")
	if err != nil {
		return "", nil, err
	}
	regionRow, err := selectGatewayNamedRow(gatewayRegionRows(regionsAndZones), spec.RegionID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["region_id"], item["id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("region-id: %w", err)
	}
	zoneRow, err := selectGatewayNamedRow(gatewayZoneRows(regionsAndZones), spec.ZoneID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["id"], item["zone_id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("zone-id: %w", err)
	}
	zoneID := firstNonEmptyString(zoneRow["id"], zoneRow["zone_id"])

	flavorInfo, err := fetcher.FetchGatewayCloudInfo(spec.CloudAccountID, []string{"flavors"}, spec.RegionID, zoneID, "", "", "", "make_hg")
	if err != nil {
		return "", nil, err
	}
	flavorRows := normalizeHuaweiGatewayFlavorRows(listNestedMaps(flavorInfo, "cloud_info", "flavors"))
	flavorRow, err := selectGatewayFlavorRow(flavorRows, spec.FlavorID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["id"], item["flavor_id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("flavor-id: %w", err)
	}
	flavorID := firstNonEmptyString(flavorRow["id"], flavorRow["flavor_id"])

	imageAndDiskInfo, err := fetchHuaweiGatewayCloudInfo(fetcher, spec.CloudAccountID, []string{"images", "system_volume_types"}, spec.RegionID, zoneID, flavorID, "make_hg")
	if err != nil {
		return "", nil, err
	}
	imageRow, err := selectGatewayImageRow(listNestedMaps(imageAndDiskInfo, "cloud_info", "images"), spec.ImageID)
	if err != nil {
		return "", nil, fmt.Errorf("image-id: %w", err)
	}
	systemDiskTypeRow, err := selectGatewayNamedRow(listNestedMaps(imageAndDiskInfo, "cloud_info", "system_volume_types"), spec.SystemDiskTypeID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["id"], item["system_disk_type_id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("system-disk-type-id: %w", err)
	}

	networkInfo, err := fetcher.FetchGatewayCloudInfo(spec.CloudAccountID, []string{"networks", "subnets", "abilities"}, spec.RegionID, zoneID, "", "", "", "make_hg")
	if err != nil {
		return "", nil, err
	}
	networkRows := listNestedMaps(networkInfo, "cloud_info", "networks")
	subnetRows := listNestedMaps(networkInfo, "cloud_info", "subnets")
	networkRow, err := selectGatewayNetworkRow(networkRows, subnetRows, spec.NetworkID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["id"], item["network_id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("network-id: %w", err)
	}
	networkID := firstNonEmptyString(networkRow["id"], networkRow["network_id"])
	subnetRow, err := selectGatewaySubnetRow(subnetRows, networkID, spec.SubnetID)
	if err != nil {
		return "", nil, fmt.Errorf("subnet-id: %w", err)
	}

	transitionImages, err := fetcher.FetchGatewayTransitionImages(spec.CloudAccountID, spec.CloudType, spec.RegionID, zoneID, "rebuild_bcd", "system", "windows", "bios")
	if err != nil {
		return "", nil, err
	}
	winHDImageRow, err := selectGatewayNamedRow(gatewayImageRows(transitionImages), spec.BootLoaderImageID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["image_id"], item["id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("boot-loader-image-id: %w", err)
	}

	metadata := map[string]interface{}{
		"region_id":               spec.RegionID,
		"region_name":             firstNonEmptyString(regionRow["region_name"], regionRow["name"], regionRow["display_name"], regionRow["region_id"], regionRow["id"], spec.RegionID),
		"zone_id":                 zoneID,
		"zone_name":               firstNonEmptyString(zoneRow["zone_name"], zoneRow["name"], zoneRow["display_name"], zoneRow["zone_id"], zoneRow["id"], zoneID),
		"flavor_name":             firstNonEmptyString(flavorRow["name"], flavorRow["flavor_name"], flavorRow["flavor_id"], flavorRow["id"], flavorID),
		"vcpu_id":                 formatAliyunGatewayVCPU(flavorRow["vcpus"]),
		"ram_GB":                  formatAliyunGatewayRAM(flavorRow["ram_GB"]),
		"GHz":                     firstNonEmptyString(flavorRow["GHz"], flavorRow["ghz"]),
		"quota_rate_name":         firstNonEmptyString(flavorRow["quota_rate_name"], flavorRow["quota_rate"]),
		"quota_pps_name":          firstNonEmptyString(flavorRow["quota_pps_name"], flavorRow["quota_pps"]),
		"flavor_id":               flavorID,
		"image_id":                firstNonEmptyString(imageRow["image_id"], imageRow["id"]),
		"image_name":              firstNonEmptyString(imageRow["image_name"], imageRow["name"], imageRow["image_id"], imageRow["id"]),
		"system_disk_type_id":     firstNonEmptyString(systemDiskTypeRow["id"], systemDiskTypeRow["system_disk_type_id"]),
		"system_disk_type_name":   firstNonEmptyString(systemDiskTypeRow["system_disk_type_name"], systemDiskTypeRow["name"], systemDiskTypeRow["display_name"], systemDiskTypeRow["system_disk_type_id"], systemDiskTypeRow["id"]),
		"volume_proxy_type":       defaultVolumeProxyType,
		"volume_proxy_type_name":  defaultVolumeProxyTypeName,
		"dest_device_type":        "vbd",
		"hg_control_network":      firstNonEmptyString(spec.HGControlNetwork, "floating_ip_without_proxy"),
		"hg_data_network":         firstNonEmptyString(spec.HGDataNetwork, "floating_ip_without_proxy"),
		"bandwidth_size_type":     "traffic",
		"hd_control_network":      firstNonEmptyString(spec.HDControlNetwork, "floating_ip_with_hg_proxy"),
		"network_id":              networkID,
		"network_name":            firstNonEmptyString(networkRow["name"], networkRow["network_name"], networkRow["display_name"], networkRow["network_id"], networkRow["id"], networkID),
		"subnet_id":               firstNonEmptyString(subnetRow["id"], subnetRow["subnet_id"]),
		"subnet_name":             firstNonEmptyString(subnetRow["name"], subnetRow["subnet_name"], subnetRow["display_name"], subnetRow["subnet_id"], subnetRow["id"]),
		"win_hd_image_id":         firstNonEmptyString(winHDImageRow["image_id"], winHDImageRow["id"]),
		"win_hd_image_name":       firstNonEmptyString(winHDImageRow["image_name"], winHDImageRow["name"], winHDImageRow["image_id"], winHDImageRow["id"]),
		"win_hd_access_type":      "hg_password",
		"win_hd_access_type_name": "\u9ed8\u8ba4",
		"win_hd_username":         "",
		"win_hd_password":         "",
	}
	if spec.SystemDiskSize != "" {
		if _, err := strconv.Atoi(spec.SystemDiskSize); err != nil {
			return "", nil, fmt.Errorf("invalid system-disk-size %q", spec.SystemDiskSize)
		}
		metadata["system_disk_size"] = spec.SystemDiskSize
	}
	putMetadataString(metadata, "fixed_ip", spec.FixedIP)
	putMetadataString(metadata, "control_nat_ip", spec.ControlNATIP)
	putMetadataString(metadata, "data_nat_ip", spec.DataNATIP)
	putMetadataString(metadata, "boot_loader_image_id", spec.BootLoaderImageID)
	putMetadataString(metadata, "boot_loader_flavor_id", spec.BootLoaderFlavorID)
	if spec.BandwidthSize != "" {
		bandwidthSize, err := strconv.Atoi(spec.BandwidthSize)
		if err != nil {
			return "", nil, fmt.Errorf("invalid bandwidth-size %q", spec.BandwidthSize)
		}
		metadata["bandwidth_size"] = bandwidthSize
	}

	return "/hypermotion/v1/storages/action", map[string]interface{}{
		"create_storage": map[string]interface{}{
			"type":               "HyperGate",
			"cloud_type":         spec.CloudType,
			"cloud_account_uuid": spec.CloudAccountID,
			"metadata":           metadata,
		},
	}, nil
}

func normalizeHuaweiGatewayFlavorRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		if row["vcpus"] == nil {
			row["vcpus"] = row["flavor_vcpus"]
		}
		if row["ram_GB"] == nil {
			row["ram_GB"] = row["flavor_ram"]
		}
	}
	return rows
}
