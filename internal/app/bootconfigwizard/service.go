package bootconfigwizard

import (
	"fmt"
	"net/url"

	"hyperbdr-client/internal/client"
)

type API interface {
	Get(path string, q url.Values) (client.APIResponse, error)
}

type Service struct {
	api API
}

func NewService(api API) Service {
	return Service{api: api}
}

type StoragesSpec struct {
	Page        int
	PageSize    int
	StorageType string
	Status      string
	Query       url.Values
}

type StorageDetailSpec struct {
	StorageID string
	Query     url.Values
}

type TargetAccountsSpec struct {
	StorageType        string
	StorageTypeDefault string
	CloudType          string
	Status             string
	TaskStatus         string
	RTExtra            string
	Extra              string
	RTTree             string
	Page               int
	PageSize           int
	Query              url.Values
}

type TargetAuthInfoSpec struct {
	CloudAccountID        string
	CloudAccountCompat    string
	CloudType             string
	StorageType           string
	FetchRes              string
	HostID                string
	StorageID             string
	WriteNetwork          string
	ReadNetwork           string
	RegionID              string
	ZoneID                string
	CloudAccountUsername  string
	CloudAccountUsePublic string
	FlavorID              string
	BootLoaderFlavorID    string
	Arch                  string
	OSTypeID              string
	OSType                string
	Flavors               string
	FlavorVCPUs           string
	FlavorRAM             string
	MaxNICNum             string
	SystemVolumeTypeID    string
	VolumeTypeID          string
	DefaultVolumeTypeID   string
	DefaultPoolID         string
	DestBootMode          string
	NetworkID             string
	Query                 url.Values
}

type SubnetConfigSpec struct {
	CloudAccountID string
	CloudType      string
	RegionID       string
	ZoneID         string
	NetworkID      string
	SubnetID       string
	Query          url.Values
}

type HostProfileSpec struct {
	ID    string
	Query url.Values
}

type StrategiesSpec struct {
	Page     int
	PageSize int
	KW       string
	Status   string
	Query    url.Values
}

func (s Service) Storages(spec StoragesSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	addString(q, "type", spec.StorageType)
	q.Set("status", spec.Status)
	addInt(q, "page", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/api/v2/getStorages", q)
}

func (s Service) StorageDetail(spec StorageDetailSpec) (client.APIResponse, error) {
	if spec.StorageID == "" {
		return client.APIResponse{}, fmt.Errorf("storage-id is required")
	}
	q := cloneValues(spec.Query)
	q.Set("storage_id", spec.StorageID)
	return s.api.Get("/api/v2/getStorageDetailInfo", q)
}

func (s Service) TargetAccounts(spec TargetAccountsSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	storageType := spec.StorageType
	if storageType == "" {
		storageType = spec.StorageTypeDefault
	}
	q.Set("storage_type", storageType)
	q.Set("cloud_type", spec.CloudType)
	q.Set("status", spec.Status)
	q.Set("task_status", spec.TaskStatus)
	q.Set("rt_extra", spec.RTExtra)
	q.Set("extra", spec.Extra)
	q.Set("rt_tree", spec.RTTree)
	addInt(q, "page", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/api/v2/getCloudAccounts", q)
}

func (s Service) TargetAuthInfo(spec TargetAuthInfoSpec) (client.APIResponse, error) {
	accountID := spec.CloudAccountID
	if accountID == "" {
		accountID = spec.CloudAccountCompat
	}
	if accountID == "" {
		return client.APIResponse{}, fmt.Errorf("cloud-account is required")
	}
	q := cloneValues(spec.Query)
	q.Set("cloud_account_id", accountID)
	addString(q, "cloud_type", spec.CloudType)
	addString(q, "storage_type", spec.StorageType)
	addString(q, "fetch_res", spec.FetchRes)
	addString(q, "host_id", spec.HostID)
	addString(q, "storage_id", spec.StorageID)
	addString(q, "network_addr_for_write_data", spec.WriteNetwork)
	addString(q, "network_addr_for_read_data", spec.ReadNetwork)
	addString(q, "region_id", spec.RegionID)
	addString(q, "zone_id", spec.ZoneID)
	addString(q, "cloud_account_username", spec.CloudAccountUsername)
	addString(q, "cloud_account_use_public", spec.CloudAccountUsePublic)
	addString(q, "flavor_id", spec.FlavorID)
	addString(q, "boot_loader_flavor_id", spec.BootLoaderFlavorID)
	addString(q, "arch", spec.Arch)
	addString(q, "os_type_id", spec.OSTypeID)
	addString(q, "os_type", spec.OSType)
	addString(q, "flavors", spec.Flavors)
	addString(q, "flavor_vcpus", spec.FlavorVCPUs)
	addString(q, "flavor_ram", spec.FlavorRAM)
	addString(q, "max_nic_num", spec.MaxNICNum)
	addString(q, "system_volume_type_id", spec.SystemVolumeTypeID)
	addString(q, "volume_type_id", spec.VolumeTypeID)
	addString(q, "default_volume_type_id", spec.DefaultVolumeTypeID)
	addString(q, "default_pool_id", spec.DefaultPoolID)
	addString(q, "dest_boot_mode", spec.DestBootMode)
	addString(q, "network_id", spec.NetworkID)
	return s.api.Get("/api/v3/getCloudInfo", q)
}

func (s Service) SubnetConfig(spec SubnetConfigSpec) (client.APIResponse, error) {
	if spec.CloudAccountID == "" {
		return client.APIResponse{}, fmt.Errorf("cloud-account is required")
	}
	if spec.CloudType == "" {
		return client.APIResponse{}, fmt.Errorf("cloud-type is required")
	}
	if spec.RegionID == "" {
		return client.APIResponse{}, fmt.Errorf("region-id is required")
	}
	if spec.ZoneID == "" {
		return client.APIResponse{}, fmt.Errorf("zone-id is required")
	}
	if spec.NetworkID == "" {
		return client.APIResponse{}, fmt.Errorf("network-id is required")
	}
	q := cloneValues(spec.Query)
	q.Set("cloud_account_id", spec.CloudAccountID)
	q.Set("cloud_type", spec.CloudType)
	q.Set("region_id", spec.RegionID)
	q.Set("zone_id", spec.ZoneID)
	q.Set("network_id", spec.NetworkID)
	q.Set("subnet_id", spec.SubnetID)
	return s.api.Get("/api/v3/getSubnetConfig", q)
}

func (s Service) HostProfile(spec HostProfileSpec) (client.APIResponse, error) {
	if spec.ID == "" {
		return client.APIResponse{}, fmt.Errorf("id is required")
	}
	q := cloneValues(spec.Query)
	q.Set("host_id", spec.ID)
	return s.api.Get("/api/v2/getHostDetail", q)
}

func (s Service) Strategies(spec StrategiesSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	q.Set("kw", spec.KW)
	q.Set("status", spec.Status)
	addInt(q, "page", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/api/v2/getHostPolicyList", q)
}

func cloneValues(v url.Values) url.Values {
	cloned := url.Values{}
	for key, values := range v {
		for _, value := range values {
			cloned.Add(key, value)
		}
	}
	return cloned
}

func addInt(q url.Values, key string, value int) {
	if value > 0 {
		q.Set(key, fmt.Sprintf("%d", value))
	}
}

func addString(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}
