package cloudaccountcreate

type Spec struct {
	CloudType                string
	CloudAuthType            string
	StorageType              string
	HasDirectAKSKStyle       bool
	HasDirectPasswordStyle   bool
	AccessKeyID              string
	AccessKeySecret          string
	AccessID                 string
	AccessSecret             string
	RegionID                 string
	RegionName               string
	AccountName              string
	AuthRegionID             string
	AuthProjectID            string
	AuthURL                  string
	CloudAccountUsername     string
	CloudAccountPassword     string
	UserDomainID             string
	ProjectDomainID          string
	ProjectID                string
	ProjectName              string
	UseInternalIP            string
	ControlAccessIP          string
	BootLoaderImageID        string
	BootLoaderImageName      string
	BootLoaderFlavorID       string
	LinuxBootImageID         string
	WindowsBootImageID       string
	LinuxUEFIBootImageID     string
	WindowsUEFIBootImageID   string
	CustomName               string
	DiskBusTypeID            string
	DiskBusTypeName          string
	SSHPort                  string
	SSHPass                  string
	LinuxHDUsername          string
	LinuxHDPassword          string
	LinuxHDPort              string
	AutoUploadImages         *int
	UploadUEFIImage          *int
	OnlyVerify               *bool
	MetadataOverrides        map[string]interface{}
	RequestOverrides         map[string]interface{}
	LinuxBootImageHostConfig LinuxBootImageHostConfigSpec
}

type LinuxBootImageHostConfigSpec struct {
	ZoneID             string
	ZoneName           string
	FlavorID           string
	FlavorName         string
	FlavorIDArr        []string
	NetworkID          string
	NetworkName        string
	SubnetID           string
	SubnetName         string
	ImageID            string
	ImageName          string
	SystemDiskTypeID   string
	SystemDiskTypeName string
}
