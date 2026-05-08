package blockstoragecreate

import "fmt"

func buildGeneric(spec Spec) (string, map[string]interface{}, error) {
	if spec.CloudAccountID == "" {
		return "", nil, fmt.Errorf("cloud-account-id is required")
	}
	if spec.CloudType == "" {
		return "", nil, fmt.Errorf("cloud-type is required")
	}

	metadata := map[string]interface{}{}
	putMetadataString(metadata, "project_id", spec.ProjectID)
	putMetadataString(metadata, "region_id", spec.RegionID)
	putMetadataString(metadata, "zone_id", spec.ZoneID)
	putMetadataString(metadata, "compute_zone_id", spec.ComputeZoneID)
	putMetadataString(metadata, "image_id", spec.ImageID)
	putMetadataString(metadata, "flavor_id", spec.FlavorID)
	putMetadataString(metadata, "network_id", spec.NetworkID)
	putMetadataString(metadata, "subnet_id", spec.SubnetID)
	putMetadataString(metadata, "fixed_ip", spec.FixedIP)
	putMetadataString(metadata, "system_disk_type_id", spec.SystemDiskTypeID)
	putMetadataString(metadata, "volume_type_id", spec.VolumeTypeID)
	putMetadataString(metadata, "system_disk_size", spec.SystemDiskSize)
	putMetadataString(metadata, "block_store_zone_id", spec.BlockStoreZoneID)
	putMetadataString(metadata, "boot_loader_image_id", spec.BootLoaderImageID)
	putMetadataString(metadata, "boot_loader_flavor_id", spec.BootLoaderFlavorID)
	putMetadataString(metadata, "project_domain_id", spec.ProjectDomainID)
	putMetadataString(metadata, "boot_types_id", spec.BootTypesID)
	putMetadataString(metadata, "volume_proxy_type", spec.VolumeProxyType)
	putMetadataString(metadata, "hg_control_network", spec.HGControlNetwork)
	putMetadataString(metadata, "control_nat_ip", spec.ControlNATIP)
	putMetadataString(metadata, "hg_data_network", spec.HGDataNetwork)
	putMetadataString(metadata, "data_nat_ip", spec.DataNATIP)
	putMetadataString(metadata, "bandwidth_size", spec.BandwidthSize)
	putMetadataString(metadata, "hd_control_network", spec.HDControlNetwork)

	return "/hypermotion/v1/storages/action", map[string]interface{}{
		"create_storage": map[string]interface{}{
			"type":               "HyperGate",
			"cloud_type":         spec.CloudType,
			"cloud_account_uuid": spec.CloudAccountID,
			"metadata":           metadata,
		},
	}, nil
}

func putMetadataString(metadata map[string]interface{}, key, value string) {
	if value == "" {
		return
	}
	metadata[key] = value
}
