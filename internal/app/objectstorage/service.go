package objectstorage

import (
	"fmt"
	"net/url"
	"strings"

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

type BucketsSpec struct {
	AuthURL         string
	RegionID        string
	AccessKeyID     string
	AccessKeySecret string
	Protocol        string
	BucketLookup    string
	UseTLS          bool
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

type CreateSpec struct {
	RawBody          interface{}
	DisplayName      string
	CloudType        string
	AuthURL          string
	RegionID         string
	AccessKeyID      string
	AccessKeySecret  string
	Protocol         string
	BucketLookup     string
	UseTLS           bool
	BucketMode       string
	BucketName       string
	PublicEndpoint   string
	InternalEndpoint string
	CloudTypeSelect  string
	AppID            string
}

type PreparedCreateRequest struct {
	Path string
	Body interface{}
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

func (s Service) Buckets(spec BucketsSpec) (client.APIResponse, error) {
	if spec.AuthURL == "" {
		return client.APIResponse{}, fmt.Errorf("auth-url is required")
	}
	if spec.AccessKeyID == "" {
		return client.APIResponse{}, fmt.Errorf("access-key-id is required")
	}
	if spec.AccessKeySecret == "" {
		return client.APIResponse{}, fmt.Errorf("access-key-secret is required")
	}
	if spec.Protocol == "" {
		spec.Protocol = "s3"
	}

	return s.api.Post("/api/v2/objectStorageBuckets", map[string]interface{}{
		"auth_type":     "aksk",
		"auth_url":      spec.AuthURL,
		"region_id":     spec.RegionID,
		"auth_key":      spec.AccessKeyID,
		"auth_cert":     spec.AccessKeySecret,
		"use_tls":       spec.UseTLS,
		"protocol":      spec.Protocol,
		"bucket_lookup": normalizeBucketLookup(spec.BucketLookup),
	})
}

func (s Service) Create(spec CreateSpec) (client.APIResponse, error) {
	prepared, err := s.PrepareCreate(spec)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.api.Post(prepared.Path, prepared.Body)
}

func (s Service) PrepareCreate(spec CreateSpec) (PreparedCreateRequest, error) {
	if spec.RawBody != nil {
		return PreparedCreateRequest{
			Path: "/api/v2/createStorage",
			Body: spec.RawBody,
		}, nil
	}
	spec.CloudType = normalizeObjectStorageCloudType(spec.CloudType, spec.AuthURL)
	if spec.AuthURL == "" {
		return PreparedCreateRequest{}, fmt.Errorf("auth-url is required")
	}
	if spec.RegionID == "" {
		return PreparedCreateRequest{}, fmt.Errorf("region-id is required")
	}
	if spec.AccessKeyID == "" {
		return PreparedCreateRequest{}, fmt.Errorf("access-key-id is required")
	}
	if spec.AccessKeySecret == "" {
		return PreparedCreateRequest{}, fmt.Errorf("access-key-secret is required")
	}
	if spec.BucketName == "" {
		return PreparedCreateRequest{}, fmt.Errorf("bucket-name is required")
	}
	if spec.DisplayName == "" {
		spec.DisplayName = defaultObjectStorageDisplayName(spec.CloudType, spec.RegionID)
	}

	normalizedMode, err := normalizeBucketMode(spec.BucketMode)
	if err != nil {
		return PreparedCreateRequest{}, err
	}
	if spec.Protocol == "" {
		spec.Protocol = "s3"
	}
	if spec.PublicEndpoint == "" {
		spec.PublicEndpoint = spec.AuthURL
	}
	if spec.InternalEndpoint == "" {
		spec.InternalEndpoint = spec.AuthURL
	}
	if spec.CloudTypeSelect == "" {
		spec.CloudTypeSelect = spec.CloudType + "," + spec.RegionID
	}

	return PreparedCreateRequest{
		Path: "/api/v2/createStorage",
		Body: map[string]interface{}{
			"display_name": spec.DisplayName,
			"cloud_type":   spec.CloudType,
			"type":         "objectstorage",
			"config": map[string]interface{}{
				"need_creation":     normalizedMode == "new",
				"bucket_name":       spec.BucketName,
				"auth_type":         "aksk",
				"auth_key":          spec.AccessKeyID,
				"auth_cert":         spec.AccessKeySecret,
				"auth_url":          spec.AuthURL,
				"protocol":          spec.Protocol,
				"bucket_lookup":     normalizeBucketLookup(spec.BucketLookup),
				"use_tls":           spec.UseTLS,
				"region_id":         spec.RegionID,
				"public_endpoint":   spec.PublicEndpoint,
				"internal_endpoint": spec.InternalEndpoint,
			},
			"metadata": map[string]interface{}{
				"cloud_type_select": spec.CloudTypeSelect,
				"app_id":            spec.AppID,
			},
		},
	}, nil
}

func normalizeObjectStorageCloudType(value, authURL string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized != "" {
		return normalized
	}
	authURL = strings.TrimSpace(strings.ToLower(authURL))
	switch {
	case strings.Contains(authURL, "myhuaweicloud.com"):
		return "huawei"
	case strings.Contains(authURL, "aliyuncs.com"):
		return "aliyun"
	default:
		return "aliyun"
	}
}

func defaultObjectStorageDisplayName(cloudType, regionID string) string {
	cloudType = strings.TrimSpace(cloudType)
	regionID = strings.TrimSpace(regionID)
	if regionID == "" {
		return cloudType
	}
	return cloudType + "-" + regionID
}

func normalizeBucketLookup(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "", "virtual-hosted-style", "virtual_hosted_style", "virtualhostedstyle":
		return "dns"
	default:
		return normalized
	}
}

func normalizeBucketMode(value string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "existing", "exist", "existing-bucket":
		return "existing", nil
	case "new", "new-bucket":
		return "new", nil
	default:
		return "", fmt.Errorf("bucket-mode must be existing or new")
	}
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
