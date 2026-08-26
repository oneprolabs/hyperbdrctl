package cloudaccount

import (
	"fmt"
	"strings"

	"hyperbdr-client/internal/client"
	workflowcreate "hyperbdr-client/internal/workflow/cloudaccountcreate"
)

type FetchResourcesSpec struct {
	workflowcreate.Spec
	BootMode         string
	Purpose          string
	FetchRes         string
	ZoneID           string
	FlavorID         string
	FlavorVCPUs      string
	FlavorRAM        string
	ComputeZoneID    string
	BlockStoreZoneID string
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
	spec.Spec = workflowcreate.NormalizeSpec(spec.Spec)

	if spec.CloudType == "" {
		return client.APIResponse{}, fmt.Errorf("cloud-type is required")
	}
	if fetchResourcesNeedRegion(spec.FetchRes) && spec.RegionID == "" {
		return client.APIResponse{}, fmt.Errorf("region-id is required")
	}

	if spec.StorageType == "" {
		spec.StorageType = "objectstorage"
	}

	authType, err := workflowcreate.ResolveGenericAuthType(spec.Spec)
	if err != nil {
		return client.APIResponse{}, err
	}

	if shouldUseTargetAuthFetch(spec, authType) {
		return s.fetchTargetAuthResources(spec)
	}

	metadata, err := buildFetchGenericMetadata(spec, authType)
	if err != nil {
		return client.APIResponse{}, err
	}

	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      spec.CloudType,
			"cloud_auth_type": authType,
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
	if spec.ZoneID != "" {
		body["zone_id"] = spec.ZoneID
	}
	if spec.FlavorID != "" {
		body["flavor_id"] = spec.FlavorID
	}
	if spec.BootMode != "" {
		body["boot_mode"] = spec.BootMode
	}
	if strings.TrimSpace(spec.Purpose) != "" {
		body["purpose"] = spec.Purpose
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
	return s.fetchTargetAuthResources(FetchResourcesSpec{
		Spec: workflowcreate.Spec{
			CloudType:            "openstack",
			CloudAuthType:        "password",
			StorageType:          "objectstorage",
			AuthURL:              spec.AuthURL,
			CloudAccountUsername: spec.Username,
			CloudAccountPassword: spec.Password,
			UserDomainID:         spec.UserDomainID,
			RegionID:             spec.RegionID,
			ProjectID:            spec.ProjectID,
			ProjectDomainID:      spec.ProjectDomainID,
			ProjectName:          spec.ProjectName,
		},
		FetchRes:         spec.FetchRes,
		ComputeZoneID:    spec.ComputeZoneID,
		BlockStoreZoneID: spec.BlockStoreZoneID,
	})
}

func (s Service) fetchTargetAuthResources(spec FetchResourcesSpec) (client.APIResponse, error) {
	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      spec.CloudType,
			"cloud_auth_type": "password",
			"storage_type":    spec.StorageType,
			"metadata": map[string]interface{}{
				"cloud_type":     spec.CloudType,
				"auth_url":       spec.AuthURL,
				"user_domain_id": spec.UserDomainID,
				"username":       spec.CloudAccountUsername,
				"password":       spec.CloudAccountPassword,
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

func shouldUseTargetAuthFetch(spec FetchResourcesSpec, authType string) bool {
	return authType == "password" && spec.CloudType == "openstack"
}

func buildFetchGenericMetadata(spec FetchResourcesSpec, authType string) (map[string]interface{}, error) {
	metadata := map[string]interface{}{
		"cloud_type": spec.CloudType,
	}

	switch authType {
	case "aksk":
		switch {
		case spec.AccessKeyID != "" || spec.AccessKeySecret != "":
			if spec.AccessKeyID == "" {
				return nil, fmt.Errorf("access-key-id is required")
			}
			if spec.AccessKeySecret == "" {
				return nil, fmt.Errorf("access-key-secret is required")
			}
			metadata["access_key_id"] = spec.AccessKeyID
			metadata["access_key_secret"] = spec.AccessKeySecret
		default:
			if spec.AccessID == "" {
				return nil, fmt.Errorf("access-id is required")
			}
			if spec.AccessSecret == "" {
				return nil, fmt.Errorf("access-secret is required")
			}
			metadata["access_id"] = spec.AccessID
			metadata["access_secret"] = spec.AccessSecret
		}
	case "password":
		if spec.AuthURL == "" {
			return nil, fmt.Errorf("auth-url is required")
		}
		if spec.CloudAccountUsername == "" {
			return nil, fmt.Errorf("username is required")
		}
		if spec.CloudAccountPassword == "" {
			return nil, fmt.Errorf("password is required")
		}
		metadata["auth_url"] = spec.AuthURL
		metadata["username"] = spec.CloudAccountUsername
		metadata["password"] = spec.CloudAccountPassword
		if spec.UserDomainID != "" {
			metadata["user_domain_id"] = spec.UserDomainID
		}
	}

	if spec.RegionID != "" {
		metadata["region_type"] = "1"
		metadata["region_type_list"] = spec.RegionID
	}
	return metadata, nil
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
