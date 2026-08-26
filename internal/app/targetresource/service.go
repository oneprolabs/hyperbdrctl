package targetresource

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"hyperbdr-client/catalog"
	"hyperbdr-client/internal/client"
)

type API interface {
	Get(path string, q url.Values) (client.APIResponse, error)
	Post(path string, body interface{}) (client.APIResponse, error)
}

type Service struct {
	api API
}

func NewService(api API) Service {
	return Service{api: api}
}

type DirectAuthSpec struct {
	CloudType     string
	StorageType   string
	CloudAuthType string
	FetchRes      string
	RegionID      string
	ZoneID        string
	BootMode      string
	DynamicFields map[string]string
}

type AccountFetchSpec struct {
	CloudAccountID string
	FetchRes       string
	RegionID       string
	ZoneID         string
	FlavorID       string
	Query          url.Values
}

type Result struct {
	Response    client.APIResponse
	CloudType   string
	StorageType string
	Resources   []string
	Route       string
}

type CloudAccountContext struct {
	CloudAccountID string
	CloudType      string
	StorageType    string
	RegionID       string
}

func (s Service) CloudAccountContext(accountID string) (CloudAccountContext, error) {
	return s.loadCloudAccountContext(accountID)
}

type accountFetchRoute string

const (
	routeDirectAuthGeneric   = "/api/v3/postCloudInfoForAuth"
	routeDirectAuthOpenStack = "/api/v2/postTargetCloudInfoForAuth"
	routeAccountGetCloudInfo = "/api/v3/getCloudInfo"
	routeAccountAction       = "/hypermotion/v1/cloud_accounts/%s/action"
	routeAccountDetail       = "/hypermotion/v1/cloud_accounts/%s"

	accountRouteGetCloudInfo accountFetchRoute = "get_cloud_info"
	accountRouteAction       accountFetchRoute = "cloud_account_action"
)

type accountFetchRule func(ctx CloudAccountContext, resources []string, q url.Values) bool

var accountFetchActionRules = []accountFetchRule{
	func(ctx CloudAccountContext, resources []string, _ url.Values) bool {
		if !isBlockStorageType(ctx.StorageType) || len(resources) == 0 {
			return false
		}
		allowed := map[string]bool{
			"abilities":       true,
			"win_hd_images":   true,
			"linux_hd_images": true,
		}
		for _, resource := range resources {
			if !allowed[resource] {
				return false
			}
		}
		return true
	},
}

func (s Service) DirectAuth(spec DirectAuthSpec) (Result, error) {
	spec.DynamicFields = cloneStringMap(spec.DynamicFields)
	spec.CloudType = strings.TrimSpace(spec.CloudType)
	spec.StorageType = normalizeStorageType(spec.StorageType)
	if spec.CloudType == "" {
		return Result{}, fmt.Errorf("cloud-type is required")
	}
	if spec.StorageType == "" {
		return Result{}, fmt.Errorf("storage-type is required")
	}

	resources := NormalizeRequestedResources(spec.FetchRes)

	authType, err := resolveAuthType(spec.CloudAuthType, spec.DynamicFields)
	if err != nil {
		return Result{}, err
	}

	if shouldUseOpenStackDirectAuth(spec.CloudType, authType) {
		// The OpenStack target-auth endpoint cannot return flattened flavor
		// candidates. Keep a compatibility path for object-storage flavor
		// queries through the generic auth endpoint, which accepts the
		// anonymous account scope used by the resource API.
		if shouldUseOpenStackObjectFlavorCompat(spec.CloudType, spec.StorageType, resources) {
			compatSpec := spec
			compatSpec.FetchRes = "flavors"
			body, err := buildGenericDirectAuthBody(compatSpec, authType)
			if err != nil {
				return Result{}, err
			}
			cloudAccount := body["cloud_account"].(map[string]interface{})
			cloudAccount["cloud_account_id"] = "anonymous"
			body["rt_flatten"] = 1
			resp, err := s.api.Post(routeDirectAuthGeneric, body)
			if err != nil {
				return Result{}, err
			}
			return Result{
				Response:    resp,
				CloudType:   spec.CloudType,
				StorageType: spec.StorageType,
				Resources:   resources,
				Route:       routeDirectAuthGeneric,
			}, nil
		}
		body, err := buildOpenStackDirectAuthBody(spec)
		if err != nil {
			return Result{}, err
		}
		resp, err := s.api.Post(routeDirectAuthOpenStack, body)
		if err != nil {
			return Result{}, err
		}
		return Result{
			Response:    resp,
			CloudType:   spec.CloudType,
			StorageType: spec.StorageType,
			Resources:   resources,
			Route:       routeDirectAuthOpenStack,
		}, nil
	}

	body, err := buildGenericDirectAuthBody(spec, authType)
	if err != nil {
		return Result{}, err
	}
	resp, err := s.api.Post(routeDirectAuthGeneric, body)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Response:    resp,
		CloudType:   spec.CloudType,
		StorageType: spec.StorageType,
		Resources:   resources,
		Route:       routeDirectAuthGeneric,
	}, nil
}

func (s Service) Fetch(spec AccountFetchSpec) (Result, error) {
	if strings.TrimSpace(spec.CloudAccountID) == "" {
		return Result{}, fmt.Errorf("cloud-account-id is required")
	}

	ctx, err := s.loadCloudAccountContext(spec.CloudAccountID)
	if err != nil {
		return Result{}, err
	}
	if ctx.CloudType == "" {
		return Result{}, fmt.Errorf("cloud-type cannot be inferred from cloud-account-id")
	}
	if ctx.StorageType == "" {
		return Result{}, fmt.Errorf("storage-type cannot be inferred from cloud-account-id")
	}

	resources := NormalizeRequestedResources(spec.FetchRes)
	if spec.RegionID == "" {
		spec.RegionID = ctx.RegionID
	}

	route := resolveAccountFetchRoute(ctx, resources, spec.Query)
	switch route {
	case accountRouteAction:
		body, err := buildAccountActionBody(resources, spec.RegionID, spec.ZoneID, spec.FlavorID, spec.Query)
		if err != nil {
			return Result{}, err
		}
		resp, err := s.api.Post(fmt.Sprintf(routeAccountAction, url.PathEscape(spec.CloudAccountID)), body)
		if err != nil {
			return Result{}, err
		}
		return Result{
			Response:    resp,
			CloudType:   ctx.CloudType,
			StorageType: ctx.StorageType,
			Resources:   resources,
			Route:       string(accountRouteAction),
		}, nil
	default:
		q := buildGetCloudInfoQuery(spec, ctx)
		resp, err := s.api.Get(routeAccountGetCloudInfo, q)
		if err != nil {
			return Result{}, err
		}
		return Result{
			Response:    resp,
			CloudType:   ctx.CloudType,
			StorageType: ctx.StorageType,
			Resources:   resources,
			Route:       string(accountRouteGetCloudInfo),
		}, nil
	}
}

func (s Service) loadCloudAccountContext(accountID string) (CloudAccountContext, error) {
	resp, err := s.api.Get(fmt.Sprintf(routeAccountDetail, url.PathEscape(accountID)), url.Values{})
	if err != nil {
		return CloudAccountContext{}, err
	}
	data := responseMap(resp)
	if data == nil {
		return CloudAccountContext{}, nil
	}

	ctx := CloudAccountContext{
		CloudAccountID: accountID,
		CloudType: firstNonEmptyString(
			data["cloud_type"],
			nestedAccountValue(data, "cloud_type"),
			nestedMetadataValue(data, "cloud_type"),
		),
		StorageType: normalizeStorageType(firstNonEmptyString(
			data["storage_type"],
			nestedAccountValue(data, "storage_type"),
			nestedMetadataValue(data, "storage_type"),
		)),
		RegionID: firstNonEmptyString(
			data["region_id"],
			data["auth_region_id"],
			nestedAccountValue(data, "region_id"),
			nestedAccountValue(data, "region_type_list"),
			nestedAccountValue(data, "auth_region_id"),
			nestedMetadataValue(data, "region_type_list"),
			nestedMetadataValue(data, "auth_region_id"),
		),
	}
	if ctx.StorageType == "" {
		switch {
		case isKnownBlockCloud(ctx.CloudType):
			ctx.StorageType = "HyperGate"
		case isKnownObjectCloud(ctx.CloudType):
			ctx.StorageType = "objectstorage"
		}
	}
	return ctx, nil
}

func resolveAccountFetchRoute(ctx CloudAccountContext, resources []string, q url.Values) accountFetchRoute {
	if isSharedBlockResourceQuery(ctx.StorageType, resources) {
		return accountRouteGetCloudInfo
	}
	for _, rule := range accountFetchActionRules {
		if rule(ctx, resources, q) {
			return accountRouteAction
		}
	}
	return accountRouteGetCloudInfo
}

func buildGetCloudInfoQuery(spec AccountFetchSpec, ctx CloudAccountContext) url.Values {
	q := cloneValues(spec.Query)
	if q.Get("rt_flatten") == "" {
		q.Set("rt_flatten", "1")
	}
	q.Set("cloud_account_id", spec.CloudAccountID)
	q.Set("cloud_type", ctx.CloudType)
	q.Set("storage_type", ctx.StorageType)
	addString(q, "fetch_res", spec.FetchRes)
	addString(q, "region_id", spec.RegionID)
	addString(q, "zone_id", spec.ZoneID)
	addString(q, "flavor_id", spec.FlavorID)
	return q
}

func buildGenericDirectAuthBody(spec DirectAuthSpec, authType string) (map[string]interface{}, error) {
	metadata, err := buildGenericMetadata(spec, authType)
	if err != nil {
		return nil, err
	}
	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      spec.CloudType,
			"cloud_auth_type": authType,
			"storage_type":    spec.StorageType,
			"metadata":        metadata,
		},
	}
	addStringToBody(body, "fetch_res", spec.FetchRes)
	addStringToBody(body, "region_id", spec.RegionID)
	addStringToBody(body, "zone_id", spec.ZoneID)
	addStringToBody(body, "boot_mode", spec.BootMode)
	return body, nil
}

func buildOpenStackDirectAuthBody(spec DirectAuthSpec) (map[string]interface{}, error) {
	fields := spec.DynamicFields
	authURL := strings.TrimSpace(fields["auth_url"])
	username := strings.TrimSpace(fields["username"])
	password := strings.TrimSpace(fields["password"])
	userDomainID := strings.TrimSpace(fields["user_domain_id"])
	if authURL == "" {
		return nil, fmt.Errorf("auth-url is required")
	}
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if userDomainID == "" {
		return nil, fmt.Errorf("user-domain-id is required")
	}

	body := map[string]interface{}{
		"cloud_account": map[string]interface{}{
			"cloud_type":      spec.CloudType,
			"cloud_auth_type": "password",
			"storage_type":    spec.StorageType,
			"metadata": map[string]interface{}{
				"cloud_type":     spec.CloudType,
				"auth_url":       authURL,
				"user_domain_id": userDomainID,
				"username":       username,
				"password":       password,
			},
		},
		"fetch_scene":         "gateway",
		"rt_tree":             0,
		"region_id":           nilString(spec.RegionID),
		"project_id":          nilString(fields["project_id"]),
		"project_domain_id":   nilString(fields["project_domain_id"]),
		"project_name":        nilString(fields["project_name"]),
		"compute_zone_id":     nilString(fields["compute_zone_id"]),
		"block_store_zone_id": nilString(fields["block_store_zone_id"]),
	}
	addStringToBody(body, "fetch_res", spec.FetchRes)
	return body, nil
}

func buildGenericMetadata(spec DirectAuthSpec, authType string) (map[string]interface{}, error) {
	metadata := map[string]interface{}{
		"cloud_type": spec.CloudType,
	}

	fields := spec.DynamicFields
	switch authType {
	case "aksk":
		switch {
		case strings.TrimSpace(fields["access_key_id"]) != "" || strings.TrimSpace(fields["access_key_secret"]) != "":
			if strings.TrimSpace(fields["access_key_id"]) == "" {
				return nil, fmt.Errorf("access-key-id is required")
			}
			if strings.TrimSpace(fields["access_key_secret"]) == "" {
				return nil, fmt.Errorf("access-key-secret is required")
			}
			metadata["access_key_id"] = fields["access_key_id"]
			metadata["access_key_secret"] = fields["access_key_secret"]
		default:
			if strings.TrimSpace(fields["access_id"]) == "" {
				return nil, fmt.Errorf("access-id is required")
			}
			if strings.TrimSpace(fields["access_secret"]) == "" {
				return nil, fmt.Errorf("access-secret is required")
			}
			metadata["access_id"] = fields["access_id"]
			metadata["access_secret"] = fields["access_secret"]
		}
	case "password":
		if strings.TrimSpace(fields["auth_url"]) == "" {
			return nil, fmt.Errorf("auth-url is required")
		}
		if strings.TrimSpace(fields["username"]) == "" {
			return nil, fmt.Errorf("username is required")
		}
		if strings.TrimSpace(fields["password"]) == "" {
			return nil, fmt.Errorf("password is required")
		}
		metadata["auth_url"] = fields["auth_url"]
		metadata["username"] = fields["username"]
		metadata["password"] = fields["password"]
		if strings.TrimSpace(fields["user_domain_id"]) != "" {
			metadata["user_domain_id"] = fields["user_domain_id"]
		}
	}

	if spec.RegionID != "" {
		metadata["region_type"] = "1"
		metadata["region_type_list"] = spec.RegionID
	}

	reserved := map[string]bool{
		"access_key_id":     true,
		"access_key_secret": true,
		"access_id":         true,
		"access_secret":     true,
		"auth_url":          true,
		"username":          true,
		"password":          true,
		"user_domain_id":    true,
		"cloud_auth_type":   true,
		"fetch_res":         true,
		"region_id":         true,
		"zone_id":           true,
		"boot_mode":         true,
		"cloud_account_id":  true,
	}
	for key, value := range fields {
		if reserved[key] || strings.TrimSpace(value) == "" {
			continue
		}
		metadata[key] = value
	}
	return metadata, nil
}

func buildAccountActionBody(resources []string, regionID, zoneID, flavorID string, q url.Values) (map[string]interface{}, error) {
	resourceOptions := map[string]interface{}{}
	domain := map[string]interface{}{}
	if regionID != "" {
		domain["region_id"] = regionID
	}
	body := map[string]interface{}{
		"get_cloud_info": map[string]interface{}{
			"domain":            domain,
			"resources_options": resourceOptions,
		},
	}

	flavorVCPUs := q.Get("flavor_vcpus")
	flavorRAM := q.Get("flavor_ram")
	purpose := q.Get("purpose")
	if purpose == "" {
		purpose = "make_hg"
	}
	imageType := q.Get("image_type")

	for _, resource := range resources {
		switch resource {
		case "regions":
			resourceOptions["domain"] = actionRegionOption(regionID)
		case "zones":
			resourceOptions["zones"] = actionRegionOption(regionID)
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
				"image_type": actionResourceImageType(resource, imageType),
				"purpose":    purpose,
			}
		case "win_hd_images":
			resourceOptions["win_hd_images"] = map[string]interface{}{
				"image_type": actionResourceImageType(resource, imageType),
			}
		case "linux_hd_images":
			resourceOptions["linux_hd_images"] = map[string]interface{}{
				"image_type": actionResourceImageType(resource, imageType),
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
	return body, nil
}

func resolveAuthType(explicit string, fields map[string]string) (string, error) {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" && explicit != "aksk" && explicit != "password" {
		return "", fmt.Errorf("cloud-auth-type must be one of aksk or password")
	}

	hasAKSK := (strings.TrimSpace(fields["access_key_id"]) != "" || strings.TrimSpace(fields["access_key_secret"]) != "") ||
		(strings.TrimSpace(fields["access_id"]) != "" || strings.TrimSpace(fields["access_secret"]) != "")
	hasPassword := strings.TrimSpace(fields["username"]) != "" || strings.TrimSpace(fields["password"]) != ""

	if explicit != "" {
		return explicit, nil
	}
	switch {
	case hasAKSK && hasPassword:
		return "", fmt.Errorf("multiple credential styles provided; pass --cloud-auth-type explicitly")
	case hasAKSK:
		return "aksk", nil
	case hasPassword:
		return "password", nil
	default:
		return "", fmt.Errorf("cloud-auth-type is required")
	}
}

func shouldUseOpenStackDirectAuth(cloudType, authType string) bool {
	return strings.EqualFold(strings.TrimSpace(cloudType), "openstack") && authType == "password"
}

func shouldUseOpenStackObjectFlavorCompat(cloudType, storageType string, resources []string) bool {
	if !strings.EqualFold(strings.TrimSpace(cloudType), "openstack") || normalizeStorageType(storageType) != "objectstorage" {
		return false
	}
	for _, resource := range resources {
		if resource == "flavors" {
			return true
		}
	}
	return false
}

func NormalizeRequestedResources(fetchRes string) []string {
	if strings.TrimSpace(fetchRes) == "" {
		return nil
	}
	parts := strings.Split(fetchRes, ",")
	resources := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		resource := normalizeResourceName(part)
		if resource == "" || seen[resource] {
			continue
		}
		seen[resource] = true
		resources = append(resources, resource)
	}
	return resources
}

func normalizeResourceName(name string) string {
	resource := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(name, "-", "_")))
	switch resource {
	case "region":
		return "regions"
	case "zone":
		return "zones"
	case "flavor":
		return "flavors"
	default:
		return resource
	}
}

func resourcesNeedRegion(resources []string) bool {
	for _, resource := range resources {
		switch resource {
		case "", "regions", "zones":
			continue
		default:
			return true
		}
	}
	return false
}

func isSharedBlockResourceQuery(storageType string, resources []string) bool {
	if !isBlockStorageType(storageType) || len(resources) == 0 {
		return false
	}
	shared := map[string]bool{
		"regions":  true,
		"zones":    true,
		"flavors":  true,
		"images":   true,
		"networks": true,
		"subnets":  true,
	}
	for _, resource := range resources {
		if !shared[resource] {
			return false
		}
	}
	return true
}

func normalizeStorageType(storageType string) string {
	switch strings.ToLower(strings.TrimSpace(storageType)) {
	case "block", "blockstorage", "hypergate":
		return "HyperGate"
	case "object", "oss", "objectstorage":
		return "objectstorage"
	default:
		return strings.TrimSpace(storageType)
	}
}

func isBlockStorageType(storageType string) bool {
	return normalizeStorageType(storageType) == "HyperGate"
}

func isKnownBlockCloud(cloudType string) bool {
	_, ok := catalog.FindBlockCloud(cloudType)
	return ok
}

func isKnownObjectCloud(cloudType string) bool {
	_, ok := catalog.FindObjectCloud(cloudType)
	return ok
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

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func addString(q url.Values, key, value string) {
	if strings.TrimSpace(value) != "" {
		q.Set(key, value)
	}
}

func addStringToBody(body map[string]interface{}, key, value string) {
	if strings.TrimSpace(value) != "" {
		body[key] = value
	}
}

func nilString(value string) interface{} {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func actionRegionOption(regionID string) map[string]interface{} {
	if regionID == "" {
		return map[string]interface{}{}
	}
	return map[string]interface{}{"region_id": regionID}
}

func actionResourceImageType(resource, imageType string) string {
	if imageType != "" {
		return imageType
	}
	if resource == "linux_hd_images" {
		return "user_create"
	}
	return "system"
}

func responseMap(resp client.APIResponse) map[string]interface{} {
	if m, ok := resp.Data.(map[string]interface{}); ok {
		return m
	}
	if resp.Raw != nil {
		return resp.Raw
	}
	return nil
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

func firstNonEmptyString(values ...interface{}) string {
	for _, value := range values {
		if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}
