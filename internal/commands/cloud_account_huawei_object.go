package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	appcloudaccount "hyperbdr-client/internal/app/cloudaccount"
	"hyperbdr-client/internal/normalize/cloudinfo"
	workflowcreate "hyperbdr-client/internal/workflow/cloudaccountcreate"
)

func enrichHuaweiObjectCloudAccountSpec(ctx *context, spec cloudAccountCreateSpec) (cloudAccountCreateSpec, error) {
	if spec.AccessKeyID == "" {
		return spec, fmt.Errorf("access-key-id is required")
	}
	if spec.AccessKeySecret == "" {
		return spec, fmt.Errorf("access-key-secret is required")
	}
	if spec.RegionID == "" {
		return spec, fmt.Errorf("region-id is required")
	}
	if err := validateHuaweiObjectControlFields(spec); err != nil {
		return spec, err
	}

	service := appcloudaccount.NewService(commandPosterAdapter{ctx: ctx})
	fetch := func(resources, zoneID, flavorID, bootMode string) (interface{}, error) {
		purpose := ""
		if resources == "flavors" || resources == "images" {
			purpose = "make_image"
		}
		resp, err := service.FetchResources(appcloudaccount.FetchResourcesSpec{
			Spec: workflowcreate.Spec{
				CloudType:       spec.CloudType,
				CloudAuthType:   "aksk",
				AccessKeyID:     spec.AccessKeyID,
				AccessKeySecret: spec.AccessKeySecret,
				StorageType:     spec.StorageType,
				RegionID:        spec.RegionID,
			},
			FetchRes: resources,
			Purpose:  purpose,
			ZoneID:   zoneID,
			FlavorID: flavorID,
			BootMode: bootMode,
		})
		if err != nil {
			return nil, fmt.Errorf("auto-resolve Huawei object resources %s: %w", resources, err)
		}
		return resp.Data, nil
	}

	regionData, err := fetch("regions,zones", "", "", "")
	if err != nil {
		return spec, err
	}
	region, err := selectHuaweiObjectRow(cloudinfo.RegionRows(regionData), spec.RegionID, "region_id", "id", "value")
	if err != nil {
		return spec, fmt.Errorf("region-id: %w", err)
	}
	if spec.RegionName == "" {
		spec.RegionName = preferredRegionLabel(ctx.loc.Lang(), region)
	}
	if spec.RegionName == "" {
		spec.RegionName = spec.RegionID
	}

	networkData, err := fetch("networks,subnets", spec.LinuxBootImageHostConfig.ZoneID, "", "")
	if err != nil {
		return spec, err
	}
	network, subnet, err := selectHuaweiObjectNetworkAndSubnet(
		cloudinfo.ResourceRows(networkData, "networks"),
		cloudinfo.ResourceRows(networkData, "subnets"),
		spec.LinuxBootImageHostConfig.NetworkID,
		spec.LinuxBootImageHostConfig.SubnetID,
		spec.LinuxBootImageHostConfig.ZoneID,
	)
	if err != nil {
		return spec, err
	}

	zones := cloudinfo.ZoneRows(regionData)
	wantedZoneID := spec.LinuxBootImageHostConfig.ZoneID
	if wantedZoneID == "" {
		wantedZoneID = mapString(subnet, "zone_id", "availability_zone", "az")
	}
	zone, err := selectHuaweiObjectRow(zones, wantedZoneID, "id", "zone_id", "value")
	if err != nil {
		return spec, fmt.Errorf("linux-boot-image-host-config-zone-id: %w", err)
	}
	zoneID := mapString(zone, "id", "zone_id", "value")
	if subnetZoneID := mapString(subnet, "zone_id", "availability_zone", "az"); subnetZoneID != "" && subnetZoneID != zoneID {
		return spec, fmt.Errorf("linux-boot-image-host-config-subnet-id: subnet %q is not available in zone %q", mapString(subnet, "id", "subnet_id"), zoneID)
	}

	flavorData, err := fetch("flavors", zoneID, "", "")
	if err != nil {
		return spec, err
	}
	flavor, flavorPath, err := selectHuaweiObjectFlavor(cloudinfo.ResourceRows(flavorData, "flavors"), spec.LinuxBootImageHostConfig.FlavorID)
	if err != nil {
		return spec, fmt.Errorf("linux-boot-image-host-config-flavor-id: %w", err)
	}
	flavorID := mapString(flavor, "id", "flavor_id", "value")

	imageData, err := fetch("images", zoneID, flavorID, "")
	if err != nil {
		return spec, err
	}
	image, err := selectHuaweiObjectImage(cloudinfo.ImageRows(imageData), spec.LinuxBootImageHostConfig.ImageID)
	if err != nil {
		return spec, fmt.Errorf("linux-boot-image-host-config-image-id: %w", err)
	}
	diskData, err := fetch("system_volume_types", zoneID, flavorID, "")
	if err != nil {
		return spec, err
	}
	diskType, err := selectHuaweiObjectRecommendedRow(cloudinfo.SystemVolumeTypeRows(diskData), spec.LinuxBootImageHostConfig.SystemDiskTypeID, "id", "system_disk_type_id", "value")
	if err != nil {
		return spec, fmt.Errorf("linux-boot-image-host-config-system-disk-type-id: %w", err)
	}

	transitionData, err := fetch("boot_loader_images", zoneID, flavorID, "bios")
	if err != nil {
		return spec, err
	}
	bootImage, err := selectHuaweiObjectRecommendedRow(cloudinfo.NormalizeImageRows(cloudinfo.ResourceRows(transitionData, "boot_loader_images")), spec.BootLoaderImageID, "image_id", "id", "uuid")
	if err != nil {
		return spec, fmt.Errorf("boot-loader-image-id: %w", err)
	}
	host := &spec.LinuxBootImageHostConfig
	host.ZoneID = zoneID
	host.ZoneName = firstNonEmptyString(host.ZoneName, mapString(zone, "zone_name", "display_name", "name"), zoneID)
	host.FlavorID = flavorID
	host.FlavorName = firstNonEmptyString(host.FlavorName, mapString(flavor, "name", "flavor_name", "display_name"), flavorID)
	host.FlavorIDArr = flavorPath
	host.NetworkID = mapString(network, "id", "network_id", "value")
	host.NetworkName = firstNonEmptyString(host.NetworkName, mapString(network, "name", "network_name", "display_name"), host.NetworkID)
	host.SubnetID = mapString(subnet, "id", "subnet_id", "value")
	host.SubnetName = firstNonEmptyString(host.SubnetName, mapString(subnet, "name", "subnet_name", "display_name"), host.SubnetID)
	host.ImageID = mapString(image, "image_id", "id", "uuid")
	host.ImageName = firstNonEmptyString(host.ImageName, mapString(image, "image_name", "name", "display_name"), host.ImageID)
	host.SystemDiskTypeID = mapString(diskType, "id", "system_disk_type_id", "value")
	host.SystemDiskTypeName = firstNonEmptyString(host.SystemDiskTypeName, mapString(diskType, "system_disk_type_name", "name", "display_name"), host.SystemDiskTypeID)
	spec.BootLoaderImageID = mapString(bootImage, "image_id", "id", "uuid")
	spec.BootLoaderImageName = firstNonEmptyString(spec.BootLoaderImageName, mapString(bootImage, "image_name", "name", "display_name"), spec.BootLoaderImageID)
	spec.AuthRegionID = firstNonEmptyString(spec.AuthRegionID, spec.RegionID)
	if spec.CustomName == "" {
		spec.CustomName = defaultHuaweiObjectCloudAccountName(ctx.loc.Lang(), spec.RegionName)
	}
	return spec, nil
}

func validateHuaweiObjectControlFields(spec cloudAccountCreateSpec) error {
	for _, value := range []struct {
		name  string
		value string
	}{{"use-internal-ip", spec.UseInternalIP}} {
		if value.value != "" && value.value != "0" && value.value != "1" {
			return fmt.Errorf("%s must be 0 or 1", value.name)
		}
	}
	if spec.ControlAccessIP != "" && net.ParseIP(spec.ControlAccessIP) == nil {
		return fmt.Errorf("control-access-ip must be a valid IP address")
	}
	return nil
}

func selectHuaweiObjectNetworkAndSubnet(networks, subnets []map[string]interface{}, networkID, subnetID, zoneID string) (map[string]interface{}, map[string]interface{}, error) {
	if len(networks) == 0 {
		return nil, nil, fmt.Errorf("linux-boot-image-host-config-network-id: no networks candidates returned")
	}
	if len(subnets) == 0 {
		return nil, nil, fmt.Errorf("linux-boot-image-host-config-subnet-id: no subnets candidates returned")
	}
	eligibleSubnets := make([]map[string]interface{}, 0, len(subnets))
	for _, subnet := range subnets {
		subnetZone := mapString(subnet, "zone_id", "availability_zone", "az")
		if zoneID == "" || subnetZone == "" || subnetZone == zoneID {
			eligibleSubnets = append(eligibleSubnets, subnet)
		}
	}
	if subnetID != "" {
		subnet, err := selectHuaweiObjectRow(eligibleSubnets, subnetID, "id", "subnet_id", "value")
		if err != nil {
			return nil, nil, fmt.Errorf("linux-boot-image-host-config-subnet-id: %w", err)
		}
		subnetNetworkID := mapString(subnet, "network_id", "vpc_id")
		if networkID != "" && subnetNetworkID != "" && subnetNetworkID != networkID {
			return nil, nil, fmt.Errorf("linux-boot-image-host-config-subnet-id: subnet %q belongs to network %q, not %q", subnetID, subnetNetworkID, networkID)
		}
		networkID = firstNonEmptyString(networkID, subnetNetworkID)
		network, err := selectHuaweiObjectRow(networks, networkID, "id", "network_id", "vpc_id", "value")
		if err != nil {
			return nil, nil, fmt.Errorf("linux-boot-image-host-config-network-id: %w", err)
		}
		return network, subnet, nil
	}

	if networkID != "" {
		network, err := selectHuaweiObjectRow(networks, networkID, "id", "network_id", "vpc_id", "value")
		if err != nil {
			return nil, nil, fmt.Errorf("linux-boot-image-host-config-network-id: %w", err)
		}
		for _, subnet := range eligibleSubnets {
			if mapString(subnet, "network_id", "vpc_id") == networkID {
				return network, subnet, nil
			}
		}
		return nil, nil, fmt.Errorf("linux-boot-image-host-config-subnet-id: network %q has no available subnet", networkID)
	}

	for _, recommended := range []bool{true, false} {
		for _, network := range networks {
			if recommended && !huaweiObjectRecommended(network) {
				continue
			}
			candidateID := mapString(network, "id", "network_id", "vpc_id", "value")
			for _, subnet := range eligibleSubnets {
				if mapString(subnet, "network_id", "vpc_id") == candidateID {
					return network, subnet, nil
				}
			}
		}
	}
	return nil, nil, fmt.Errorf("linux-boot-image-host-config-network-id: no network with an available subnet")
}

func selectHuaweiObjectFlavor(rows []map[string]interface{}, wantedID string) (map[string]interface{}, []string, error) {
	flatRows, paths := flattenHuaweiObjectFlavorRows(rows, nil)
	normalized := cloudinfo.NormalizeHuaweiFlavorRows(flatRows)
	if len(normalized) == 0 {
		return nil, nil, fmt.Errorf("no flavors candidates returned")
	}
	if wantedID != "" {
		for index, row := range normalized {
			if mapString(row, "id") == wantedID {
				return row, normalizedFlavorPath(paths[index], wantedID), nil
			}
		}
		return nil, nil, fmt.Errorf("resource %q not found", wantedID)
	}
	selected := -1
	for index, row := range normalized {
		if huaweiObjectFlavorValue(row, "vcpus") == 2 && huaweiObjectFlavorValue(row, "ram_gb") == 4 {
			if selected < 0 || huaweiObjectRecommended(row) {
				selected = index
			}
			if huaweiObjectRecommended(row) {
				break
			}
		}
	}
	if selected < 0 {
		return nil, nil, fmt.Errorf("no 2 vCPU / 4 GiB flavor candidates returned")
	}
	id := mapString(normalized[selected], "id")
	return normalized[selected], normalizedFlavorPath(paths[selected], id), nil
}

func flattenHuaweiObjectFlavorRows(rows []map[string]interface{}, parent []string) ([]map[string]interface{}, [][]string) {
	flat := make([]map[string]interface{}, 0)
	paths := make([][]string, 0)
	for _, row := range rows {
		id := mapString(row, "id", "flavor_id", "value")
		path := append([]string{}, parent...)
		if id != "" {
			path = append(path, id)
		}
		children, _ := row["children"].([]interface{})
		if len(children) == 0 {
			flat = append(flat, row)
			paths = append(paths, path)
			continue
		}
		childRows := make([]map[string]interface{}, 0, len(children))
		for _, child := range children {
			if childRow, ok := child.(map[string]interface{}); ok {
				childRows = append(childRows, childRow)
			}
		}
		childFlat, childPaths := flattenHuaweiObjectFlavorRows(childRows, path)
		flat = append(flat, childFlat...)
		paths = append(paths, childPaths...)
	}
	return flat, paths
}

func normalizedFlavorPath(path []string, flavorID string) []string {
	if len(path) == 0 {
		return []string{flavorID}
	}
	if path[len(path)-1] != flavorID {
		return append(path, flavorID)
	}
	return path
}

func selectHuaweiObjectImage(rows []map[string]interface{}, wantedID string) (map[string]interface{}, error) {
	if wantedID != "" {
		return selectHuaweiObjectRow(rows, wantedID, "image_id", "id", "uuid")
	}
	for _, recommended := range []bool{true, false} {
		for _, row := range rows {
			osType := strings.ToLower(mapString(row, "os_type", "__os_type", "platform"))
			if osType != "" && osType != "linux" {
				continue
			}
			if !recommended || huaweiObjectRecommended(row) {
				return row, nil
			}
		}
	}
	return nil, fmt.Errorf("no Linux images candidates returned")
}

func selectHuaweiObjectRecommendedRow(rows []map[string]interface{}, wantedID string, idKeys ...string) (map[string]interface{}, error) {
	if wantedID != "" {
		return selectHuaweiObjectRow(rows, wantedID, idKeys...)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no candidates returned")
	}
	for _, row := range rows {
		if huaweiObjectRecommended(row) {
			return row, nil
		}
	}
	return rows[0], nil
}

func selectHuaweiObjectRow(rows []map[string]interface{}, wantedID string, idKeys ...string) (map[string]interface{}, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("no candidates returned")
	}
	if wantedID == "" {
		return rows[0], nil
	}
	for _, row := range rows {
		if mapString(row, idKeys...) == wantedID {
			return row, nil
		}
	}
	return nil, fmt.Errorf("resource %q not found", wantedID)
}

func huaweiObjectRecommended(row map[string]interface{}) bool {
	value, ok := row["is_recommend"]
	if !ok {
		value = row["recommended"]
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case float64:
		return typed == 1
	case int:
		return typed == 1
	case string:
		return typed == "1" || strings.EqualFold(typed, "true")
	default:
		return false
	}
}

func huaweiObjectFlavorValue(row map[string]interface{}, key string) int {
	value := row[key]
	switch typed := value.(type) {
	case int:
		return typed
	case float64:
		return int(typed)
	case string:
		parsed, _ := strconv.Atoi(typed)
		return parsed
	default:
		return 0
	}
}

func defaultHuaweiObjectCloudAccountName(lang, regionLabel string) string {
	if lang == "zh_cn" {
		return "华为云(推荐使用，SDK v3.1.86)-" + regionLabel
	}
	return "Huawei Cloud(Recommended, SDK v3.1.86)-" + regionLabel
}
