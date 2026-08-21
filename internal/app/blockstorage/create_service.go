package blockstorage

import (
	"fmt"
	"net/url"
	"strconv"

	"hyperbdr-client/internal/client"
	workflowcreate "hyperbdr-client/internal/workflow/blockstoragecreate"
)

type CreateSpec = workflowcreate.Spec

type PreparedCreateRequest struct {
	Path string
	Body map[string]interface{}
}

func (s Service) Create(spec CreateSpec) (client.APIResponse, error) {
	prepared, err := s.PrepareCreate(spec)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.api.Post(prepared.Path, prepared.Body)
}

func (s Service) PrepareCreate(spec CreateSpec) (PreparedCreateRequest, error) {
	if spec.CloudAccountID == "" {
		return PreparedCreateRequest{}, fmt.Errorf("cloud-account-id is required")
	}
	defaults, err := s.gatewayAccountDefaults(spec.CloudAccountID, spec.CloudType, spec.RegionID)
	if err != nil {
		return PreparedCreateRequest{}, err
	}
	spec.CloudType = defaults.CloudType
	spec.RegionID = defaults.RegionID
	if spec.CloudType == "aliyun_bs" && spec.RegionID == "" {
		return PreparedCreateRequest{}, fmt.Errorf("region-id is required")
	}

	path, body, err := workflowcreate.BuildRequest(runtimeAdapter{api: s.api}, spec)
	if err != nil {
		return PreparedCreateRequest{}, err
	}
	return PreparedCreateRequest{
		Path: path,
		Body: body,
	}, nil
}

type gatewayAccountDefaults struct {
	CloudType string
	RegionID  string
}

func (s Service) gatewayAccountDefaults(accountID, explicitCloudType, explicitRegionID string) (gatewayAccountDefaults, error) {
	defaults := gatewayAccountDefaults{
		CloudType: explicitCloudType,
		RegionID:  explicitRegionID,
	}
	needAccountLookup := defaults.CloudType == "" || (defaults.RegionID == "" && gatewayCloudTypeUsesAccountRegion(defaults.CloudType))
	if !needAccountLookup {
		return defaults, nil
	}
	resp, err := s.api.Get("/hypermotion/v1/cloud_accounts/"+url.PathEscape(accountID), url.Values{})
	if err != nil {
		return gatewayAccountDefaults{}, err
	}
	data := responseMap(resp)
	if data == nil {
		return gatewayAccountDefaults{}, fmt.Errorf("cloud-type is required")
	}
	if defaults.CloudType == "" {
		defaults.CloudType = gatewayAccountField(data, "cloud_type")
		if defaults.CloudType == "" {
			return gatewayAccountDefaults{}, fmt.Errorf("cloud-type is required")
		}
	}
	if defaults.RegionID == "" && gatewayCloudTypeUsesAccountRegion(defaults.CloudType) {
		defaults.RegionID = gatewayAccountRegionID(data)
	}
	return defaults, nil
}

func gatewayCloudTypeUsesAccountRegion(cloudType string) bool {
	switch cloudType {
	case "aliyun_bs", "huawei_bs":
		return true
	default:
		return false
	}
}

func gatewayAccountField(data map[string]interface{}, key string) string {
	if value := firstNonEmptyString(data[key]); value != "" {
		return value
	}
	if account, ok := data["cloud_account"].(map[string]interface{}); ok {
		if value := firstNonEmptyString(account[key]); value != "" {
			return value
		}
	}
	return ""
}

func gatewayAccountRegionID(data map[string]interface{}) string {
	return firstNonEmptyString(
		data["region_id"],
		data["auth_region_id"],
		nestedAccountValue(data, "region_id"),
		nestedAccountValue(data, "region_type_list"),
		nestedAccountValue(data, "auth_region_id"),
		nestedMetadataValue(data, "region_type_list"),
		nestedMetadataValue(data, "auth_region_id"),
	)
}

func nestedAccountValue(data map[string]interface{}, key string) interface{} {
	if account, ok := data["cloud_account"].(map[string]interface{}); ok {
		return account[key]
	}
	return nil
}

func nestedMetadataValue(data map[string]interface{}, key string) interface{} {
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		if value := metadata[key]; value != nil {
			return value
		}
	}
	if account, ok := data["cloud_account"].(map[string]interface{}); ok {
		if metadata, ok := account["metadata"].(map[string]interface{}); ok {
			return metadata[key]
		}
	}
	return nil
}

func (s Service) gatewayAccountCloudType(accountID, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	defaults, err := s.gatewayAccountDefaults(accountID, explicit, "")
	if err != nil {
		return "", err
	}
	return defaults.CloudType, nil
}

type runtimeAdapter struct {
	api API
}

func (r runtimeAdapter) FetchGatewayCloudInfo(accountID string, resources []string, regionID, zoneID, flavorID, flavorVCPUs, flavorRAM, purpose string) (map[string]interface{}, error) {
	body, err := buildGatewayResourcesRequest(resources, regionID, zoneID, flavorID, flavorVCPUs, flavorRAM, purpose, "system")
	if err != nil {
		return nil, err
	}
	resp, err := r.api.Post("/hypermotion/v1/cloud_accounts/"+url.PathEscape(accountID)+"/action", body)
	if err != nil {
		return nil, err
	}
	data := responseMap(resp)
	if data == nil || nestedMap(data, "cloud_info") == nil {
		return nil, fmt.Errorf("cloud_info is missing from gateway resource response")
	}
	return data, nil
}

func (r runtimeAdapter) FetchGatewayTransitionImages(accountID, cloudType, regionID, zoneID, purpose, imageType, osType, bootMode string) (map[string]interface{}, error) {
	q := url.Values{}
	q.Set("cloud_account_id", accountID)
	q.Set("cloud_type", cloudType)
	q.Set("storage_type", "HyperGate")
	q.Set("region_id", regionID)
	q.Set("zone_id", zoneID)
	q.Set("purpose", purpose)
	q.Set("image_type", imageType)
	q.Set("os_type", osType)
	q.Set("boot_mode", bootMode)
	q.Set("fetch_res", "images")
	q.Set("image_sources", "gold,private,shared")

	resp, err := r.api.Get("/api/v3/getCloudInfo", q)
	if err != nil {
		return nil, err
	}
	data := responseMap(resp)
	if data == nil || nestedMap(data, "cloud_info") == nil {
		return nil, fmt.Errorf("cloud_info is missing from transition image response")
	}
	return data, nil
}

func (r runtimeAdapter) FetchOpenStackCloudInfo(accountID string) (map[string]interface{}, error) {
	resp, err := r.api.Post("/hypermotion/v1/cloud_accounts/"+url.PathEscape(accountID)+"/action", map[string]interface{}{
		"get_cloud_info": map[string]interface{}{},
	})
	if err != nil {
		return nil, err
	}
	cloudInfo := nestedMap(responseData(resp), "cloud_info")
	if cloudInfo == nil {
		return nil, fmt.Errorf("cloud_info is missing from OpenStack gateway resource response")
	}
	return cloudInfo, nil
}

func responseData(resp client.APIResponse) interface{} {
	if payload, ok := resp.Data.(map[string]interface{}); ok {
		return payload
	}
	if resp.Raw != nil {
		return resp.Raw
	}
	return resp.Data
}

func responseMap(resp client.APIResponse) map[string]interface{} {
	payload, _ := responseData(resp).(map[string]interface{})
	return payload
}

func nestedMap(data interface{}, path ...string) map[string]interface{} {
	value := data
	for _, key := range path {
		next, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}
		value = next[key]
	}
	m, _ := value.(map[string]interface{})
	return m
}

func firstNonEmptyString(values ...interface{}) string {
	for _, value := range values {
		if s, ok := value.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func buildGatewayResourcesRequest(resources []string, regionID, zoneID, flavorID, flavorVCPUs, flavorRAM, purpose, imageType string) (map[string]interface{}, error) {
	resourceOptions := map[string]interface{}{}
	domain := map[string]interface{}{}
	if regionID != "" {
		domain["region_id"] = regionID
	}
	request := map[string]interface{}{
		"get_cloud_info": map[string]interface{}{
			"domain":            domain,
			"resources_options": resourceOptions,
		},
	}

	for _, resource := range resources {
		switch resource {
		case "regions":
			resourceOptions["domain"] = gatewayRegionOption(regionID)
		case "zones":
			resourceOptions["zones"] = gatewayRegionOption(regionID)
		case "flavors":
			flavorOptions := map[string]interface{}{"zone_id": zoneID, "purpose": purpose}
			if flavorVCPUs != "" {
				vcpus, err := strconv.Atoi(flavorVCPUs)
				if err != nil {
					return nil, fmt.Errorf("invalid flavor-vcpus %q", flavorVCPUs)
				}
				flavorOptions["vcpus"] = vcpus
			}
			if flavorRAM != "" {
				ramGB, err := strconv.ParseFloat(flavorRAM, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid flavor-ram %q", flavorRAM)
				}
				flavorOptions["ram_GB"] = ramGB
			}
			resourceOptions["flavors"] = flavorOptions
		case "images":
			resourceOptions["images"] = map[string]interface{}{
				"zone_id":    zoneID,
				"flavor_id":  flavorID,
				"image_type": gatewayResourceImageType(resource, imageType),
				"purpose":    purpose,
			}
		case "win_hd_images":
			resourceOptions["win_hd_images"] = map[string]interface{}{
				"image_type": gatewayResourceImageType(resource, imageType),
			}
		case "linux_hd_images":
			resourceOptions["linux_hd_images"] = map[string]interface{}{
				"image_type": gatewayResourceImageType(resource, imageType),
			}
		case "system_disk_types":
			resourceOptions["system_disk_types"] = map[string]interface{}{
				"zone_id":   zoneID,
				"flavor_id": flavorID,
				"purpose":   purpose,
				"support":   true,
			}
		case "networks":
			resourceOptions["networks"] = map[string]interface{}{}
		case "subnets":
			resourceOptions["subnets"] = map[string]interface{}{"zone_id": zoneID}
		case "abilities":
			resourceOptions["abilities"] = map[string]interface{}{}
		}
	}

	return request, nil
}

func gatewayRegionOption(regionID string) map[string]interface{} {
	if regionID == "" {
		return map[string]interface{}{}
	}
	return map[string]interface{}{"region_id": regionID}
}

func gatewayResourceImageType(resource, imageType string) string {
	if imageType != "" {
		return imageType
	}
	if resource == "linux_hd_images" {
		return "user_create"
	}
	return "system"
}
