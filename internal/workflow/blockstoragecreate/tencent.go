package blockstoragecreate

import (
	"fmt"
	"strconv"
	"strings"
)

// buildTencent builds the Tencent Cloud block-storage gateway payload from the
// same read-only cloud-info resources used by the Alibaba and Huawei adapters.
func buildTencent(fetcher Fetcher, spec Spec) (string, map[string]interface{}, error) {
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
	regionRow := selectGatewayRegionRow(regionsAndZones, spec.RegionID)
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
	flavorRow, err := selectGatewayFlavorRow(listNestedMaps(flavorInfo, "cloud_info", "flavors"), spec.FlavorID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["id"], item["flavor_id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("flavor-id: %w", err)
	}
	flavorID := firstNonEmptyString(flavorRow["id"], flavorRow["flavor_id"])

	imageAndDiskInfo, err := fetcher.FetchGatewayCloudInfo(spec.CloudAccountID, []string{"images", "system_disk_types"}, spec.RegionID, zoneID, flavorID, "", "", "make_hg")
	if err != nil {
		return "", nil, err
	}
	imageRow, err := selectGatewayImageRow(listNestedMaps(imageAndDiskInfo, "cloud_info", "images"), spec.ImageID)
	if err != nil {
		return "", nil, fmt.Errorf("image-id: %w", err)
	}
	diskTypeRow, err := selectTencentSystemDiskType(listNestedMaps(imageAndDiskInfo, "cloud_info", "system_disk_types"), spec.SystemDiskTypeID)
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
	subnetCandidates := make([]map[string]interface{}, 0, len(subnetRows))
	for _, row := range subnetRows {
		rowZoneID := firstNonEmptyString(row["zone_id"])
		if rowZoneID == "" || rowZoneID == zoneID {
			subnetCandidates = append(subnetCandidates, row)
		}
	}
	subnetRow, err := selectGatewaySubnetRow(subnetCandidates, networkID, spec.SubnetID)
	if err != nil {
		return "", nil, fmt.Errorf("subnet-id: %w", err)
	}

	transitionImages, err := fetcher.FetchGatewayTransitionImages(spec.CloudAccountID, spec.CloudType, spec.RegionID, zoneID, "rebuild_bcd", "system", "windows", "bios")
	if err != nil {
		return "", nil, err
	}
	windowsImageRow, err := selectGatewayNamedRow(gatewayImageRows(transitionImages), spec.BootLoaderImageID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["image_id"], item["id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("boot-loader-image-id: %w", err)
	}

	hgControlNetwork := firstNonEmptyString(spec.HGControlNetwork, "floating_ip_without_proxy")
	hgDataNetwork := firstNonEmptyString(spec.HGDataNetwork, "floating_ip_without_proxy")
	hdControlNetwork := spec.HDControlNetwork
	if hdControlNetwork == "" || (hdControlNetwork == "floating_ip_with_hg_proxy" && !containsMetadataKey(spec.ExplicitMetadataKeys, "hd_control_network")) {
		hdControlNetwork = "fixed_ip_with_hg_proxy"
	}
	systemDiskSize := firstNonEmptyString(spec.SystemDiskSize, "50")
	if _, err := strconv.Atoi(systemDiskSize); err != nil {
		return "", nil, fmt.Errorf("invalid system-disk-size %q", systemDiskSize)
	}
	defaultBandwidthSize := "50"
	if strings.HasPrefix(hgControlNetwork, "floating_ip") || strings.HasPrefix(hgDataNetwork, "floating_ip") {
		defaultBandwidthSize = "100"
	}
	bandwidthSize, err := strconv.Atoi(firstNonEmptyString(spec.BandwidthSize, defaultBandwidthSize))
	if err != nil {
		return "", nil, fmt.Errorf("invalid bandwidth-size %q", spec.BandwidthSize)
	}

	metadata := map[string]interface{}{
		"region_id":               spec.RegionID,
		"region_name":             firstNonEmptyString(regionRow["region_name"], regionRow["name"], regionRow["display_name"], regionRow["id"], spec.RegionID),
		"zone_id":                 zoneID,
		"zone_name":               firstNonEmptyString(zoneRow["zone_name"], zoneRow["name"], zoneRow["display_name"], zoneID),
		"flavor_id":               flavorID,
		"flavor_name":             firstNonEmptyString(flavorRow["name"], flavorRow["flavor_name"], flavorID),
		"vcpu_id":                 formatAliyunGatewayVCPU(flavorRow["vcpus"]),
		"ram_GB":                  formatAliyunGatewayRAM(firstNonEmpty(flavorRow["ram_GB"], flavorRow["ram"])),
		"GHz":                     firstNonEmptyString(flavorRow["GHz"], flavorRow["ghz"], "暂时无法获取"),
		"quota_rate_name":         firstNonEmptyString(flavorRow["quota_rate_name"], flavorRow["quota_rate"]),
		"quota_pps_name":          firstNonEmptyString(flavorRow["quota_pps_name"], flavorRow["quota_pps"]),
		"image_id":                firstNonEmptyString(imageRow["image_id"], imageRow["id"]),
		"image_name":              firstNonEmptyString(imageRow["image_name"], imageRow["name"], imageRow["id"]),
		"system_disk_type_id":     firstNonEmptyString(diskTypeRow["id"], diskTypeRow["system_disk_type_id"]),
		"system_disk_type_name":   firstNonEmptyString(diskTypeRow["system_disk_type_name"], diskTypeRow["display_name"], diskTypeRow["name"], diskTypeRow["id"]),
		"system_disk_size":        systemDiskSize,
		"volume_proxy_type":       defaultVolumeProxyType,
		"volume_proxy_type_name":  defaultVolumeProxyTypeName,
		"dest_host_boot_type":     "image",
		"hg_control_network":      hgControlNetwork,
		"hg_control_network_name": aliyunGatewayControlNetworkName(hgControlNetwork),
		"control_nat_ip":          valueOrNil(spec.ControlNATIP),
		"hg_data_network":         hgDataNetwork,
		"hg_data_network_name":    aliyunGatewayControlNetworkName(hgDataNetwork),
		"data_nat_ip":             valueOrNil(spec.DataNATIP),
		"bandwidth_size":          bandwidthSize,
		"hd_control_network":      hdControlNetwork,
		"hd_control_network_name": aliyunDriverAdaptNetworkModeName(hdControlNetwork),
		"network_id":              networkID,
		"network_name":            firstNonEmptyString(networkRow["name"], networkRow["display_name"], networkID),
		"subnet_id":               firstNonEmptyString(subnetRow["id"], subnetRow["subnet_id"]),
		"subnet_name":             firstNonEmptyString(subnetRow["name"], subnetRow["display_name"], subnetRow["id"]),
		"fixed_ip":                spec.FixedIP,
		"win_hd_image_id":         firstNonEmptyString(windowsImageRow["image_id"], windowsImageRow["id"]),
		"win_hd_image_name":       firstNonEmptyString(windowsImageRow["image_name"], windowsImageRow["name"], windowsImageRow["id"]),
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

func firstNonEmpty(values ...interface{}) interface{} {
	for _, value := range values {
		if value == nil {
			continue
		}
		if text, ok := value.(string); ok && strings.TrimSpace(text) == "" {
			continue
		}
		return value
	}
	return nil
}

func selectTencentSystemDiskType(items []map[string]interface{}, want string) (map[string]interface{}, error) {
	id := func(item map[string]interface{}) string {
		return firstNonEmptyString(item["id"], item["system_disk_type_id"])
	}
	if want != "" {
		return selectGatewayNamedRow(items, want, id)
	}
	// The Tencent console defaults to Balanced SSD when it is available.
	for _, item := range items {
		if id(item) == "CLOUD_BSSD" {
			return item, nil
		}
	}
	return selectGatewayRecommendedRow(items, "", id)
}

func containsMetadataKey(keys []string, want string) bool {
	for _, key := range keys {
		if key == want {
			return true
		}
	}
	return false
}
