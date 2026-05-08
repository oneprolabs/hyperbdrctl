package bootconfigapply

var aliyunBlockApplyDriver = applyDriverFunc{
	name: "aliyun_block_storage",
	run: func(s Service, spec *applyDriverSpec) error {
		applyAliyunBlockStorageDefaults(spec.Metadata, spec.StorageInfo)
		return s.enrichBlockStorageNetworkMetadata(spec.HostID, spec.Metadata, spec.Dynamic)
	},
}

func applyAliyunBlockStorageDefaults(metadata map[string]interface{}, storageInfo storageDefaults) {
	setMissingString(metadata, "zone_id", storageInfo.ZoneID)
}
