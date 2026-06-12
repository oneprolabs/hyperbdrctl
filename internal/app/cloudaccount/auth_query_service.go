package cloudaccount

import (
	"fmt"
	"strings"

	"hyperbdr-client/internal/client"
)

type FetchResourcesSpec struct {
	CloudType       string
	AccessKeyID     string
	AccessKeySecret string
	StorageType     string
	RegionID        string
	BootMode        string
	FetchRes        string
}

type FetchOpenStackObjectResourcesSpec struct {
	AuthURL          string
	Username         string
	Password         string
	UserDomainID     string
	FetchRes         string
	RegionID         string
	ProjectID        string
	ProjectDomainID  string
	ProjectName      string
	ComputeZoneID    string
	BlockStoreZoneID string
}

func (s Service) FetchResources(spec FetchResourcesSpec) (client.APIResponse, error) {
	if spec.CloudType == "" {
		return client.APIResponse{}, fmt.Errorf("cloud-type is required")
	}
	if spec.AccessKeyID == "" {
		return client.APIResponse{}, fmt.Errorf("access-key-id is required")
	}
	if spec.AccessKeySecret == "" {
		return client.APIResponse{}, fmt.Errorf("access-key-secret is required")
	}
	if fetchResourcesNeedRegion(spec.FetchRes) && spec.RegionID == "" {
		return client.APIResponse{}, fmt.Errorf("region-id is required")
	}

	if spec.StorageType == "" {
		spec.StorageType = "objectstorage"
	}

	metadata := map[string]interface{}{
		"access_key_id":     spec.AccessKeyID,
		"access_key_secret": spec.AccessKeySecret,
		"cloud_type":        spec.CloudType,
	}
	if spec.RegionID != "" {
		metadata["region_type"] = "1"
		metadata["region_type_list"] = spec.RegionID
	}

	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      spec.CloudType,
			"cloud_auth_type": "aksk",
			"storage_type":    spec.StorageType,
			"metadata":        metadata,
		},
	}
	if strings.TrimSpace(spec.FetchRes) != "" {
		body["fetch_res"] = spec.FetchRes
	}
	if spec.RegionID != "" {
		body["region_id"] = spec.RegionID
	}
	if spec.BootMode != "" {
		body["boot_mode"] = spec.BootMode
	}

	return s.api.Post("/api/v3/postCloudInfoForAuth", body)
}

func (s Service) FetchOpenStackObjectResources(spec FetchOpenStackObjectResourcesSpec) (client.APIResponse, error) {
	if spec.AuthURL == "" {
		return client.APIResponse{}, fmt.Errorf("auth-url is required")
	}
	if spec.Username == "" {
		return client.APIResponse{}, fmt.Errorf("cloud-account-username is required")
	}
	if spec.Password == "" {
		return client.APIResponse{}, fmt.Errorf("cloud-account-password is required")
	}
	if spec.UserDomainID == "" {
		return client.APIResponse{}, fmt.Errorf("user-domain-id is required")
	}
	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      "openstack",
			"cloud_auth_type": "password",
			"storage_type":    "objectstorage",
			"metadata": map[string]interface{}{
				"cloud_type":     "openstack",
				"auth_url":       spec.AuthURL,
				"user_domain_id": spec.UserDomainID,
				"username":       spec.Username,
				"password":       spec.Password,
			},
		},
		"fetch_scene":         "gateway",
		"rt_tree":             0,
		"region_id":           nilString(spec.RegionID),
		"project_id":          nilString(spec.ProjectID),
		"project_domain_id":   nilString(spec.ProjectDomainID),
		"project_name":        nilString(spec.ProjectName),
		"compute_zone_id":     nilString(spec.ComputeZoneID),
		"block_store_zone_id": nilString(spec.BlockStoreZoneID),
	}
	if strings.TrimSpace(spec.FetchRes) != "" {
		body["fetch_res"] = spec.FetchRes
	}

	return s.api.Post("/api/v2/postTargetCloudInfoForAuth", body)
}

func fetchResourcesNeedRegion(fetchRes string) bool {
	if strings.TrimSpace(fetchRes) == "" {
		return false
	}
	for _, raw := range strings.Split(fetchRes, ",") {
		resource := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(raw, "-", "_")))
		switch resource {
		case "", "region", "regions", "zones":
			continue
		default:
			return true
		}
	}
	return false
}

func nilString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
