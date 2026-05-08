package blockstoragecreate

type Spec struct {
	CloudAccountID     string
	CloudType          string
	ProjectID          string
	RegionID           string
	ZoneID             string
	ComputeZoneID      string
	ImageID            string
	FlavorID           string
	NetworkID          string
	SubnetID           string
	FixedIP            string
	SystemDiskTypeID   string
	VolumeTypeID       string
	SystemDiskSize     string
	BlockStoreZoneID   string
	BootLoaderImageID  string
	BootLoaderFlavorID string
	ProjectDomainID    string
	BootTypesID        string
	VolumeProxyType    string
	HGControlNetwork   string
	ControlNATIP       string
	HGDataNetwork      string
	DataNATIP          string
	BandwidthSize      string
	HDControlNetwork   string
}
