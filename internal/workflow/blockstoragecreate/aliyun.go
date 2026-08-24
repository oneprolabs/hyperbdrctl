package blockstoragecreate

import (
	"fmt"
	"strconv"
	"strings"
)

func buildAliyun(fetcher Fetcher, spec Spec) (string, map[string]interface{}, error) {
	if spec.CloudAccountID == "" {
		return "", nil, fmt.Errorf("cloud-account-id is required")
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
	systemDiskTypeRow, err := selectGatewayNamedRow(listNestedMaps(imageAndDiskInfo, "cloud_info", "system_disk_types"), spec.SystemDiskTypeID, func(item map[string]interface{}) string {
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

	hgControlNetwork := firstNonEmptyString(spec.HGControlNetwork, "floating_ip_without_proxy")
	hgDataNetwork := firstNonEmptyString(spec.HGDataNetwork, "floating_ip_without_proxy")
	hdControlNetwork := firstNonEmptyString(spec.HDControlNetwork, "floating_ip_with_hg_proxy")
	systemDiskSize := firstNonEmptyString(spec.SystemDiskSize, "40")
	defaultBandwidthSize := "50"
	if spec.BandwidthSize == "" && (strings.HasPrefix(hgControlNetwork, "floating_ip") || strings.HasPrefix(hgDataNetwork, "floating_ip")) {
		defaultBandwidthSize = "100"
	}
	bandwidthSize, err := strconv.Atoi(firstNonEmptyString(spec.BandwidthSize, defaultBandwidthSize))
	if err != nil {
		return "", nil, fmt.Errorf("invalid bandwidth-size %q", spec.BandwidthSize)
	}

	metadata := map[string]interface{}{
		"region_id":               spec.RegionID,
		"region_name":             firstNonEmptyString(regionRow["region_name"], regionRow["display_name"], regionRow["name"], spec.RegionID),
		"zone_id":                 zoneID,
		"zone_name":               firstNonEmptyString(zoneRow["display_name"], zoneRow["name"], zoneID),
		"flavor_name":             firstNonEmptyString(flavorRow["name"], flavorID),
		"vcpu_id":                 formatAliyunGatewayVCPU(flavorRow["vcpus"]),
		"ram_GB":                  formatAliyunGatewayRAM(flavorRow["ram_GB"]),
		"GHz":                     firstNonEmptyString(flavorRow["GHz"], flavorRow["ghz"]),
		"quota_rate_name":         firstNonEmptyString(flavorRow["quota_rate_name"]),
		"quota_pps_name":          firstNonEmptyString(flavorRow["quota_pps_name"]),
		"flavor_id":               flavorID,
		"image_id":                firstNonEmptyString(imageRow["image_id"], imageRow["id"]),
		"image_name":              firstNonEmptyString(imageRow["image_name"], imageRow["name"], imageRow["id"]),
		"system_disk_type_id":     firstNonEmptyString(systemDiskTypeRow["id"], systemDiskTypeRow["system_disk_type_id"]),
		"system_disk_type_name":   firstNonEmptyString(systemDiskTypeRow["id"], systemDiskTypeRow["system_disk_type_id"], systemDiskTypeRow["display_name"], systemDiskTypeRow["name"]),
		"system_disk_size":        systemDiskSize,
		"volume_proxy_type":       defaultVolumeProxyType,
		"volume_proxy_type_name":  defaultVolumeProxyTypeName,
		"dest_device_type":        "vbd",
		"dest_device_type_name":   "23",
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
		"win_hd_image_id":         firstNonEmptyString(winHDImageRow["image_id"], winHDImageRow["id"]),
		"win_hd_image_name":       firstNonEmptyString(winHDImageRow["image_name"], winHDImageRow["name"], winHDImageRow["id"]),
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
