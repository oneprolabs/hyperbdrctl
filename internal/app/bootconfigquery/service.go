package bootconfigquery

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

type ResourcesSpec struct {
	CloudAccountID        string
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

func (s Service) Resources(spec ResourcesSpec) (client.APIResponse, error) {
	if spec.CloudAccountID == "" {
		return client.APIResponse{}, fmt.Errorf("cloud-account-id is required")
	}
	if spec.CloudType == "" {
		return client.APIResponse{}, fmt.Errorf("cloud-type is required")
	}
	if spec.StorageType == "" {
		return client.APIResponse{}, fmt.Errorf("storage-type is required")
	}

	q := cloneValues(spec.Query)
	if q.Get("rt_flatten") == "" {
		q.Set("rt_flatten", "1")
	}
	q.Set("cloud_account_id", spec.CloudAccountID)
	q.Set("cloud_type", spec.CloudType)
	q.Set("storage_type", spec.StorageType)
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

func cloneValues(v url.Values) url.Values {
	cloned := url.Values{}
	for key, values := range v {
		for _, value := range values {
			cloned.Add(key, value)
		}
	}
	return cloned
}

func addString(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}
