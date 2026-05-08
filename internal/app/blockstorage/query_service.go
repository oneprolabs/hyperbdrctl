package blockstorage

import (
	"fmt"
	"net/url"
	"strings"

	"hyperbdr-client/internal/client"
)

type ResourcesSpec struct {
	CloudAccountID string
	FetchRes       string
	RegionID       string
	ZoneID         string
	FlavorID       string
	FlavorVCPUs    string
	FlavorRAM      string
	Purpose        string
	ImageType      string
}

type ListSpec struct {
	Page        int
	PageSize    int
	StorageType string
	Query       url.Values
}

type DetailSpec struct {
	ID    string
	Query url.Values
}

type ResourcesResult struct {
	Response  client.APIResponse
	Resources []string
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

type TransitionImagesSpec struct {
	CloudAccountID string
	CloudType      string
	RegionID       string
	ZoneID         string
	Purpose        string
	ImageType      string
	OSType         string
	BootMode       string
	Query          url.Values
}

func (s Service) List(spec ListSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	q.Set("type", spec.StorageType)
	addInt(q, "page", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/api/v2/getStorages", q)
}

func (s Service) Detail(spec DetailSpec) (client.APIResponse, error) {
	if spec.ID == "" {
		return client.APIResponse{}, fmt.Errorf("id is required")
	}
	q := cloneValues(spec.Query)
	q.Set("storage_id", spec.ID)
	return s.api.Get("/api/v2/getStorageDetailInfo", q)
}

func (s Service) Resources(spec ResourcesSpec) (ResourcesResult, error) {
	if spec.CloudAccountID == "" {
		return ResourcesResult{}, fmt.Errorf("cloud-account-id is required")
	}

	resources, err := normalizeGatewayFetchResources(spec.FetchRes)
	if err != nil {
		return ResourcesResult{}, err
	}
	resolvedRegionID := spec.RegionID
	accountCloudType := ""
	if resolvedRegionID == "" {
		resp, err := s.api.Get("/hypermotion/v1/cloud_accounts/"+url.PathEscape(spec.CloudAccountID), url.Values{})
		if err != nil && gatewayResourcesNeedRegion(resources) {
			return ResourcesResult{}, err
		}
		data := responseMap(resp)
		if data != nil {
			resolvedRegionID = gatewayAccountRegionID(data)
			accountCloudType = gatewayAccountField(data, "cloud_type")
		}
	}
	if gatewayResourcesNeedRegion(resources) && resolvedRegionID == "" && !(isOpenStackCloudType(accountCloudType) && openStackResourcesAllowMissingRegion(resources)) {
		return ResourcesResult{}, fmt.Errorf("region-id is required")
	}
	if gatewayResourcesNeedZone(resources) && spec.ZoneID == "" {
		return ResourcesResult{}, fmt.Errorf("zone-id is required")
	}
	if gatewayResourcesNeedFlavor(resources) && spec.FlavorID == "" {
		return ResourcesResult{}, fmt.Errorf("flavor-id is required")
	}

	body, err := buildGatewayResourcesRequest(resources, resolvedRegionID, spec.ZoneID, spec.FlavorID, spec.FlavorVCPUs, spec.FlavorRAM, spec.Purpose, spec.ImageType)
	if err != nil {
		return ResourcesResult{}, err
	}

	resp, err := s.api.Post("/hypermotion/v1/cloud_accounts/"+url.PathEscape(spec.CloudAccountID)+"/action", body)
	if err != nil {
		return ResourcesResult{}, err
	}

	return ResourcesResult{Response: resp, Resources: resources}, nil
}

func (s Service) lookupGatewayAccountRegionID(accountID string) (string, error) {
	resp, err := s.api.Get("/hypermotion/v1/cloud_accounts/"+url.PathEscape(accountID), url.Values{})
	if err != nil {
		return "", err
	}
	data := responseMap(resp)
	if data == nil {
		return "", nil
	}
	return gatewayAccountRegionID(data), nil
}

func (s Service) SubnetConfig(spec SubnetConfigSpec) (client.APIResponse, error) {
	if spec.CloudAccountID == "" {
		return client.APIResponse{}, fmt.Errorf("cloud-account-id is required")
	}
	if spec.ZoneID == "" {
		return client.APIResponse{}, fmt.Errorf("zone-id is required")
	}
	if spec.NetworkID == "" {
		return client.APIResponse{}, fmt.Errorf("network-id is required")
	}

	resolvedRegionID := spec.RegionID
	if resolvedRegionID == "" {
		accountRegionID, err := s.lookupGatewayAccountRegionID(spec.CloudAccountID)
		if err != nil {
			return client.APIResponse{}, err
		}
		resolvedRegionID = accountRegionID
	}

	resolvedCloudType, err := s.gatewayAccountCloudType(spec.CloudAccountID, spec.CloudType)
	if err != nil {
		return client.APIResponse{}, err
	}
	if isOpenStackCloudType(resolvedCloudType) {
		return client.APIResponse{}, fmt.Errorf("subnet-config is not supported for openstack; use target cloud-sync-gateway create openstack --preview-request to inspect the resolved defaults")
	}

	q := cloneValues(spec.Query)
	q.Set("cloud_account_id", spec.CloudAccountID)
	q.Set("cloud_type", resolvedCloudType)
	if resolvedRegionID != "" {
		q.Set("region_id", resolvedRegionID)
	}
	q.Set("zone_id", spec.ZoneID)
	q.Set("network_id", spec.NetworkID)
	q.Set("subnet_id", spec.SubnetID)

	return s.api.Get("/api/v3/getSubnetConfig", q)
}

func (s Service) TransitionImages(spec TransitionImagesSpec) (client.APIResponse, error) {
	if spec.CloudAccountID == "" {
		return client.APIResponse{}, fmt.Errorf("cloud-account-id is required")
	}
	if spec.RegionID == "" {
		return client.APIResponse{}, fmt.Errorf("region-id is required")
	}
	if spec.ZoneID == "" {
		return client.APIResponse{}, fmt.Errorf("zone-id is required")
	}

	resolvedCloudType, err := s.gatewayAccountCloudType(spec.CloudAccountID, spec.CloudType)
	if err != nil {
		return client.APIResponse{}, err
	}

	q := cloneValues(spec.Query)
	q.Set("cloud_account_id", spec.CloudAccountID)
	q.Set("cloud_type", resolvedCloudType)
	q.Set("storage_type", "HyperGate")
	q.Set("region_id", spec.RegionID)
	q.Set("zone_id", spec.ZoneID)
	q.Set("purpose", spec.Purpose)
	q.Set("image_type", spec.ImageType)
	q.Set("os_type", spec.OSType)
	q.Set("boot_mode", spec.BootMode)
	q.Set("fetch_res", "images")
	q.Set("image_sources", "gold,private,shared")

	return s.api.Get("/api/v3/getCloudInfo", q)
}

func normalizeGatewayFetchResources(fetchRes string) ([]string, error) {
	if strings.TrimSpace(fetchRes) == "" {
		return []string{}, nil
	}
	parts := strings.Split(fetchRes, ",")
	normalized := make([]string, 0, len(parts))
	seen := map[string]bool{}
	allowed := map[string]bool{
		"regions":           true,
		"zones":             true,
		"flavors":           true,
		"images":            true,
		"win_hd_images":     true,
		"linux_hd_images":   true,
		"system_disk_types": true,
		"networks":          true,
		"subnets":           true,
		"abilities":         true,
	}
	for _, part := range parts {
		item := strings.TrimSpace(strings.ReplaceAll(part, "-", "_"))
		if item == "" {
			continue
		}
		if !allowed[item] {
			return nil, fmt.Errorf("unsupported fetch-res value %q", part)
		}
		if !seen[item] {
			seen[item] = true
			normalized = append(normalized, item)
		}
	}
	return normalized, nil
}

func gatewayResourcesNeedRegion(resources []string) bool {
	for _, resource := range resources {
		switch resource {
		case "regions", "zones":
			continue
		default:
			return true
		}
	}
	return false
}

func gatewayResourcesNeedZone(resources []string) bool {
	for _, resource := range resources {
		switch resource {
		case "flavors", "images", "system_disk_types", "subnets":
			return true
		}
	}
	return false
}

func gatewayResourcesNeedFlavor(resources []string) bool {
	for _, resource := range resources {
		switch resource {
		case "images", "system_disk_types":
			return true
		}
	}
	return false
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

func isOpenStackCloudType(cloudType string) bool {
	return strings.EqualFold(strings.TrimSpace(cloudType), "openstack")
}

func openStackResourcesAllowMissingRegion(resources []string) bool {
	for _, resource := range resources {
		switch resource {
		case "win_hd_images", "linux_hd_images":
			continue
		default:
			return false
		}
	}
	return len(resources) > 0
}
