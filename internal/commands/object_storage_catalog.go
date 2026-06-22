package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"hyperbdr-client/internal/output"
)

type objectStorageCatalogProvider struct {
	ID                   string                       `json:"id"`
	Name                 string                       `json:"name"`
	NameEn               string                       `json:"name_en"`
	Regions              []objectStorageCatalogRegion `json:"regions"`
	ObjectStorageSupport bool                         `json:"object_storage_support"`
}

type objectStorageCatalogRegion struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	NameEn           string `json:"name_en"`
	AuthURL          string `json:"auth_url"`
	ExternalEndpoint string `json:"external_endpoint"`
	InternalEndpoint string `json:"internal_endpoint"`
	Protocol         string `json:"protocol"`
	BucketLookup     string `json:"bucket_lookup"`
}

func loadObjectStorageCatalog(ctx *context) ([]objectStorageCatalogProvider, error) {
	body, err := loadObjectStorageCatalogBody(ctx)
	if err != nil {
		return nil, err
	}
	var providers []objectStorageCatalogProvider
	if err := json.Unmarshal(body, &providers); err != nil {
		return nil, err
	}
	return providers, nil
}

func loadObjectStorageCatalogRaw(ctx *context) ([]map[string]interface{}, error) {
	body, err := loadObjectStorageCatalogBody(ctx)
	if err != nil {
		return nil, err
	}
	var providers []map[string]interface{}
	if err := json.Unmarshal(body, &providers); err != nil {
		return nil, err
	}
	return providers, nil
}

func loadObjectStorageCatalogBody(ctx *context) ([]byte, error) {
	c, err := ensureClient(ctx)
	if err != nil {
		return nil, err
	}
	body, err := c.GetRaw("/static/json/s3.json", nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func findObjectStorageCatalogProvider(providers []objectStorageCatalogProvider, provider string) (*objectStorageCatalogProvider, bool) {
	for i := range providers {
		if strings.EqualFold(providers[i].ID, provider) {
			return &providers[i], true
		}
	}
	return nil, false
}

func findObjectStorageCatalogRegion(provider objectStorageCatalogProvider, regionID string) (*objectStorageCatalogRegion, bool) {
	for i := range provider.Regions {
		if strings.EqualFold(provider.Regions[i].ID, regionID) {
			return &provider.Regions[i], true
		}
	}
	return nil, false
}

func findObjectStorageCatalogProviderRaw(providers []map[string]interface{}, provider string) (map[string]interface{}, bool) {
	for _, item := range providers {
		if strings.EqualFold(mapString(item, "id"), provider) {
			return item, true
		}
	}
	return nil, false
}

func objectStorageCatalogProviderRows(providers []objectStorageCatalogProvider, lang string) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(providers))
	for _, provider := range providers {
		rows = append(rows, map[string]interface{}{
			"id":           provider.ID,
			"name":         localizedObjectStorageCatalogName(provider.Name, provider.NameEn, lang),
			"region_count": len(provider.Regions),
		})
	}
	return rows
}

func objectStorageCatalogRegionRows(provider objectStorageCatalogProvider, lang string) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(provider.Regions))
	for _, region := range provider.Regions {
		rows = append(rows, map[string]interface{}{
			"id":                region.ID,
			"name":              localizedObjectStorageCatalogName(region.Name, region.NameEn, lang),
			"auth_url":          region.AuthURL,
			"public_endpoint":   region.ExternalEndpoint,
			"internal_endpoint": region.InternalEndpoint,
			"protocol":          region.Protocol,
			"bucket_lookup":     region.BucketLookup,
		})
	}
	return rows
}

func objectStorageCatalogProviderColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.region_count", Field: "region_count"},
	}
}

func objectStorageCatalogRegionColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.auth_url", Field: "auth_url"},
		{HeaderKey: "table.public_endpoint", Field: "public_endpoint"},
		{HeaderKey: "table.internal_endpoint", Field: "internal_endpoint"},
		{HeaderKey: "table.protocol", Field: "protocol"},
		{HeaderKey: "table.bucket_lookup", Field: "bucket_lookup"},
	}
}

func localizedObjectStorageCatalogName(name, nameEn, lang string) string {
	if strings.EqualFold(lang, "zh_cn") {
		if strings.TrimSpace(name) != "" {
			return name
		}
		return nameEn
	}
	if strings.TrimSpace(nameEn) != "" {
		return nameEn
	}
	return name
}

func defaultObjectStorageCatalogDisplayName(provider objectStorageCatalogProvider, region objectStorageCatalogRegion, lang string) string {
	providerName := strings.TrimSpace(localizedObjectStorageCatalogName(provider.Name, provider.NameEn, lang))
	regionName := strings.TrimSpace(localizedObjectStorageCatalogName(region.Name, region.NameEn, lang))
	switch {
	case providerName == "":
		return regionName
	case regionName == "":
		return providerName
	default:
		return providerName + "-" + regionName
	}
}

func resolveObjectStorageCatalogRegion(ctx *context, providerID, regionID string) (objectStorageCatalogProvider, objectStorageCatalogRegion, error) {
	providers, err := loadObjectStorageCatalog(ctx)
	if err != nil {
		return objectStorageCatalogProvider{}, objectStorageCatalogRegion{}, err
	}
	provider, ok := findObjectStorageCatalogProvider(providers, providerID)
	if !ok {
		return objectStorageCatalogProvider{}, objectStorageCatalogRegion{}, fmt.Errorf(ctx.loc.T("error.object_storage_catalog_provider_not_found"), providerID)
	}
	region, ok := findObjectStorageCatalogRegion(*provider, regionID)
	if !ok {
		return objectStorageCatalogProvider{}, objectStorageCatalogRegion{}, fmt.Errorf(ctx.loc.T("error.object_storage_catalog_region_not_found"), providerID, regionID)
	}
	return *provider, *region, nil
}
