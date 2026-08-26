package objectstoragecreate

import (
	"fmt"
	"strings"
)

const (
	ModeCustom  = "custom"
	ModeCatalog = "catalog"
)

// Profile describes the selected object-storage create path. Catalog profiles
// carry the defaults resolved from /static/json/s3.json; custom profiles do not
// depend on that catalog.
type Profile struct {
	Mode               string
	ProviderID         string
	ProviderName       string
	RegionID           string
	RegionName         string
	AuthURL            string
	PublicEndpoint     string
	InternalEndpoint   string
	Protocol           string
	BucketLookup       string
	DefaultDisplayName string
}

func CustomProfile(defaultDisplayName string) Profile {
	return Profile{
		Mode:               ModeCustom,
		ProviderID:         ModeCustom,
		DefaultDisplayName: defaultDisplayName,
	}
}

func (p Profile) IsCustom() bool {
	return strings.EqualFold(strings.TrimSpace(p.Mode), ModeCustom)
}

// Spec contains user input for one create request. ExplicitFields records
// flags that were passed even when their value was empty, preserving the
// command-line-over-catalog precedence used by the existing command.
type Spec struct {
	DisplayName      string
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
	AppID            string
	ExplicitFields   map[string]bool
}

type PreparedRequest struct {
	Path string
	Body map[string]interface{}
}

type adapter struct {
	build func(Profile, Spec) (PreparedRequest, error)
}

// providerAdapters is intentionally empty in the first version. A provider
// should be registered only after its createStorage payload is verified to
// differ from the shared catalog shape.
var providerAdapters = map[string]adapter{}

var modeAdapters = map[string]adapter{
	ModeCustom:  {build: buildCustom},
	ModeCatalog: {build: buildCatalog},
}

func BuildRequest(profile Profile, spec Spec) (PreparedRequest, error) {
	providerID := strings.ToLower(strings.TrimSpace(profile.ProviderID))
	if registered, ok := providerAdapters[providerID]; ok {
		return registered.build(profile, spec)
	}
	mode := strings.ToLower(strings.TrimSpace(profile.Mode))
	registered, ok := modeAdapters[mode]
	if !ok {
		return PreparedRequest{}, fmt.Errorf("unsupported object storage create profile mode %q", profile.Mode)
	}
	return registered.build(profile, spec)
}

func buildCustom(profile Profile, spec Spec) (PreparedRequest, error) {
	profile.Mode = ModeCustom
	profile.ProviderID = ModeCustom
	return buildCommon(profile, spec)
}

func buildCatalog(profile Profile, spec Spec) (PreparedRequest, error) {
	if strings.TrimSpace(profile.ProviderID) == "" {
		return PreparedRequest{}, fmt.Errorf("provider is required")
	}
	if strings.TrimSpace(profile.RegionID) == "" {
		return PreparedRequest{}, fmt.Errorf("region-id is required")
	}
	return buildCommon(profile, spec)
}

func buildCommon(profile Profile, spec Spec) (PreparedRequest, error) {
	applyProfileDefaults(profile, &spec)

	if spec.AuthURL == "" {
		return PreparedRequest{}, fmt.Errorf("auth-url is required")
	}
	if spec.RegionID == "" && !profile.IsCustom() {
		return PreparedRequest{}, fmt.Errorf("region-id is required")
	}
	if spec.AccessKeyID == "" {
		return PreparedRequest{}, fmt.Errorf("access-key-id is required")
	}
	if spec.AccessKeySecret == "" {
		return PreparedRequest{}, fmt.Errorf("access-key-secret is required")
	}
	if spec.BucketName == "" {
		return PreparedRequest{}, fmt.Errorf("bucket-name is required")
	}

	mode, err := normalizeBucketMode(spec.BucketMode)
	if err != nil {
		return PreparedRequest{}, err
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
	if spec.DisplayName == "" {
		spec.DisplayName = technicalDisplayName(profile.ProviderID, spec.RegionID)
	}

	cloudType := strings.TrimSpace(profile.ProviderID)
	cloudTypeSelect := cloudType + "," + spec.RegionID
	if profile.IsCustom() {
		cloudType = ModeCustom
		cloudTypeSelect = ModeCustom
	}

	return PreparedRequest{
		Path: "/api/v2/createStorage",
		Body: map[string]interface{}{
			"display_name": spec.DisplayName,
			"cloud_type":   cloudType,
			"type":         "objectstorage",
			"config": map[string]interface{}{
				"need_creation":     mode == "new",
				"bucket_name":       spec.BucketName,
				"auth_type":         "aksk",
				"auth_key":          spec.AccessKeyID,
				"auth_cert":         spec.AccessKeySecret,
				"auth_url":          spec.AuthURL,
				"protocol":          spec.Protocol,
				"bucket_lookup":     NormalizeBucketLookup(spec.BucketLookup),
				"use_tls":           spec.UseTLS,
				"region_id":         spec.RegionID,
				"public_endpoint":   spec.PublicEndpoint,
				"internal_endpoint": spec.InternalEndpoint,
			},
			"metadata": map[string]interface{}{
				"cloud_type_select": cloudTypeSelect,
				"app_id":            spec.AppID,
			},
		},
	}, nil
}

func applyProfileDefaults(profile Profile, spec *Spec) {
	if spec.RegionID == "" && !explicit(spec, "region-id") {
		spec.RegionID = profile.RegionID
	}
	if spec.AuthURL == "" && !explicit(spec, "auth-url") {
		spec.AuthURL = profile.AuthURL
	}
	if spec.PublicEndpoint == "" && !explicit(spec, "public-endpoint") {
		spec.PublicEndpoint = profile.PublicEndpoint
	}
	if spec.InternalEndpoint == "" && !explicit(spec, "internal-endpoint") {
		spec.InternalEndpoint = profile.InternalEndpoint
	}
	if spec.Protocol == "" && !explicit(spec, "protocol") {
		spec.Protocol = profile.Protocol
	}
	if spec.BucketLookup == "" && !explicit(spec, "bucket-lookup") {
		spec.BucketLookup = profile.BucketLookup
	}
	if spec.DisplayName == "" && !explicit(spec, "display-name") {
		spec.DisplayName = profile.DefaultDisplayName
	}
}

func explicit(spec *Spec, name string) bool {
	return spec.ExplicitFields != nil && spec.ExplicitFields[name]
}

func technicalDisplayName(providerID, regionID string) string {
	providerID = strings.TrimSpace(providerID)
	regionID = strings.TrimSpace(regionID)
	if regionID == "" {
		return providerID
	}
	return providerID + "-" + regionID
}

func NormalizeBucketLookup(value string) string {
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
