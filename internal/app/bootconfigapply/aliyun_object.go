package bootconfigapply

var aliyunObjectApplyDriver = applyDriverFunc{
	name: "aliyun_object_storage",
	run: func(s Service, spec *applyDriverSpec) error {
		applyObjectStorageStorageDefaults(spec.Metadata, spec.StorageInfo)
		if err := s.enrichObjectStorageCloudInfoMetadata(spec.HostID, spec.Metadata, spec.HostInfo.Data); err != nil {
			return err
		}
		return s.enrichObjectStorageSubnetMetadata(spec.Metadata)
	},
}

func applyObjectStorageStorageDefaults(metadata map[string]interface{}, storageInfo storageDefaults) {
	if _, exists := metadata["storage_name"]; !exists {
		metadata["storage_name"] = storageInfo.StorageName
	}
	setMissingString(metadata, "storage_display_name", firstNonEmptyString(storageInfo.StorageDisplayName, storageInfo.StorageName, storageInfo.StorageID))
	setMissingString(metadata, "storage_type", "objectstorage")
	setMissingString(metadata, "network_addr_for_write_data", storageInfo.WriteNetwork)
	setMissingString(metadata, "network_addr_for_write_data_name", storageInfo.WriteNetworkName)
	setMissingString(metadata, "network_addr_for_read_data", storageInfo.ReadNetwork)
	setMissingString(metadata, "network_addr_for_read_data_name", storageInfo.ReadNetworkName)
	setMissingString(metadata, "default_pool_id", storageInfo.PoolID)
	setMissingString(metadata, "default_pool_name", firstNonEmptyString(storageInfo.PoolName, storageInfo.PoolID))
}
