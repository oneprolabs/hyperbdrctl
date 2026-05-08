package bootconfigapply

type applyDriverSpec struct {
	HostID      string
	Metadata    map[string]interface{}
	HostInfo    hostDetail
	StorageInfo storageDefaults
	Dynamic     map[string]string
}
