package blockstoragecreate

import "fmt"

func buildOpenStack(fetcher Fetcher, spec Spec) (string, map[string]interface{}, error) {
	if spec.CloudAccountID == "" {
		return "", nil, fmt.Errorf("cloud-account-id is required")
	}

	cloudInfo, err := fetcher.FetchOpenStackCloudInfo(spec.CloudAccountID)
	if err != nil {
		return "", nil, err
	}

	authInfo := nestedMap(cloudInfo, "auth_info")

	project, err := selectOpenStackProject(cloudInfo, authInfo, spec.ProjectID)
	if err != nil {
		return "", nil, err
	}
	region, err := selectOpenStackRegion(cloudInfo, authInfo, spec.RegionID)
	if err != nil {
		return "", nil, err
	}
	computeZone, err := selectOpenStackNamedResource(openStackMapList(cloudInfo, "compute_zones"), spec.ComputeZoneID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["id"], item["name"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("compute-zone-id: %w", err)
	}
	image, err := selectOpenStackImage(cloudInfo, spec.ImageID)
	if err != nil {
		return "", nil, err
	}
	flavor, err := selectOpenStackFlavorRow(openStackMapList(cloudInfo, "flavors"), spec.FlavorID, func(item map[string]interface{}) string {
		return stringValue(item["id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("flavor-id: %w", err)
	}
	network, err := selectOpenStackRecommendedResource(openStackMapList(cloudInfo, "networks"), spec.NetworkID, func(item map[string]interface{}) string {
		return stringValue(item["id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("network-id: %w", err)
	}
	subnet, err := selectOpenStackSubnet(cloudInfo, stringValue(network["id"]), spec.SubnetID)
	if err != nil {
		return "", nil, err
	}
	volumeType, err := selectOpenStackNamedResource(openStackMapList(cloudInfo, "volume_types"), spec.VolumeTypeID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["id"], item["name"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("volume-type-id: %w", err)
	}
	bootLoaderImage, err := selectOpenStackNamedResource(openStackMapList(cloudInfo, "boot_loader_images"), spec.BootLoaderImageID, func(item map[string]interface{}) string {
		return stringValue(item["id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("boot-loader-image-id: %w", err)
	}
	bootLoaderFlavor, err := selectOpenStackRecommendedResource(openStackMapList(cloudInfo, "boot_loader_flavors"), spec.BootLoaderFlavorID, func(item map[string]interface{}) string {
		return stringValue(item["id"])
	})
	if err != nil {
		return "", nil, fmt.Errorf("boot-loader-flavor-id: %w", err)
	}

	computeZoneID := firstNonEmptyString(computeZone["id"], computeZone["name"])
	bootTypesID := firstNonEmptyString(spec.BootTypesID, "boot_from_volume")
	volumeProxyType := firstNonEmptyString(spec.VolumeProxyType, "s3")
	hgControlNetwork := firstNonEmptyString(spec.HGControlNetwork, "floating_ip_without_proxy")
	hgDataNetwork := firstNonEmptyString(spec.HGDataNetwork, "floating_ip_without_proxy")

	metadata := map[string]interface{}{
		"project_id":              stringValue(project["id"]),
		"project_name":            firstNonEmptyString(project["name"], authInfo["project_name"]),
		"region_id":               firstNonEmptyString(region["id"], region["name"]),
		"region_name":             firstNonEmptyString(region["name"], region["id"]),
		"compute_zone_id":         computeZoneID,
		"compute_zone_name":       firstNonEmptyString(computeZone["name"], computeZone["id"]),
		"image_id":                stringValue(image["id"]),
		"image_name":              firstNonEmptyString(image["name"], image["display_name"]),
		"flavor_id":               stringValue(flavor["id"]),
		"flavor_name":             firstNonEmptyString(flavor["name"], flavor["id"]),
		"network_id":              stringValue(network["id"]),
		"network_name":            firstNonEmptyString(network["display_name"], network["name"], network["id"]),
		"subnet_id":               stringValue(subnet["id"]),
		"subnet_name":             openStackSubnetName(subnet),
		"fixed_ip":                spec.FixedIP,
		"volume_type_id":          firstNonEmptyString(volumeType["id"], volumeType["name"]),
		"volume_type_name":        firstNonEmptyString(volumeType["name"], volumeType["id"]),
		"system_disk_size":        firstNonEmptyString(spec.SystemDiskSize, "50"),
		"block_store_zone_id":     firstNonEmptyString(spec.BlockStoreZoneID, computeZoneID),
		"boot_loader_image_id":    stringValue(bootLoaderImage["id"]),
		"boot_loader_image_name":  firstNonEmptyString(bootLoaderImage["name"], bootLoaderImage["id"]),
		"boot_types_id":           bootTypesID,
		"boot_types_id_name":      openStackBootTypeName(bootTypesID),
		"volume_proxy_type":       volumeProxyType,
		"volume_proxy_type_name":  openStackVolumeProxyTypeName(volumeProxyType),
		"hg_control_network":      hgControlNetwork,
		"hg_control_network_name": openStackGatewayNetworkName(hgControlNetwork),
		"control_nat_ip":          valueOrNil(spec.ControlNATIP),
		"hg_data_network":         hgDataNetwork,
		"hg_data_network_name":    openStackGatewayNetworkName(hgDataNetwork),
		"data_nat_ip":             valueOrNil(spec.DataNATIP),
		"project_domain_id":       firstNonEmptyString(spec.ProjectDomainID, authInfo["project_domain_id"]),
		"boot_loader_flavor_id":   stringValue(bootLoaderFlavor["id"]),
		"boot_loader_flavor_name": firstNonEmptyString(bootLoaderFlavor["name"], bootLoaderFlavor["id"]),
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
