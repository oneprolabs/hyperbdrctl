package cloudinfo

import "sort"

type ResourceEnvelope struct {
	Provider    string
	CloudType   string
	StorageType string
	Sections    []ResourceSection
}

type ResourceSection struct {
	Resource string
	TitleKey string
	Rows     []map[string]interface{}
	Data     interface{}
	Meta     map[string]interface{}
}

func BuildAuthResourceEnvelope(provider, cloudType, storageType string, data interface{}, requested []string, extraMeta map[string]interface{}) ResourceEnvelope {
	resources := requested
	if len(resources) == 0 {
		resources = authAutoDetectedResources(data)
	}

	sections := make([]ResourceSection, 0, len(resources))
	for _, resource := range resources {
		sections = append(sections, ResourceSection{
			Resource: resource,
			TitleKey: authResourceTitleKey(resource),
			Rows:     authResourceRows(data, resource),
			Data:     ResourceValue(data, resource),
			Meta: mergeSectionMeta(map[string]interface{}{
				"provider":     provider,
				"cloud_type":   cloudType,
				"storage_type": storageType,
			}, extraMeta),
		})
	}

	return ResourceEnvelope{
		Provider:    provider,
		CloudType:   cloudType,
		StorageType: storageType,
		Sections:    sections,
	}
}

func mergeSectionMeta(base, extra map[string]interface{}) map[string]interface{} {
	merged := map[string]interface{}{}
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range extra {
		merged[key] = value
	}
	return merged
}

func NormalizeHuaweiFlavorRows(rows []map[string]interface{}) []map[string]interface{} {
	leafRows := flattenHuaweiFlavorRows(rows)
	normalized := make([]map[string]interface{}, 0, len(leafRows))
	for _, row := range leafRows {
		normalized = append(normalized, map[string]interface{}{
			"id":           firstNonEmptyString(row["id"], row["flavor_id"], row["value"]),
			"name":         firstNonEmptyString(row["name"], row["flavor_name"], row["display_name"]),
			"vcpus":        firstNonEmptyIntValue(row["vcpus"], row["flavor_vcpus"]),
			"ram_gb":       firstNonEmptyIntValue(row["ram_GB"], row["ram"], row["flavor_ram"]),
			"zone_id":      firstNonEmptyString(row["zone_id"], row["zone"], row["az"]),
			"ghz":          firstNonEmptyString(row["GHz"], row["ghz"]),
			"quota_rate":   firstNonEmptyString(row["quota_rate"], row["quota_rate_name"]),
			"quota_pps":    firstNonEmptyString(row["quota_pps"], row["quota_pps_name"]),
			"max_nic_num":  firstNonEmptyIntValue(row["max_nic_num"]),
			"max_disk_num": firstNonEmptyIntValue(row["max_disk_num"]),
			"is_recommend": firstNonEmptyIntValue(row["is_recommend"]),
		})
	}
	return normalized
}

func flattenHuaweiFlavorRows(rows []map[string]interface{}) []map[string]interface{} {
	flattened := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		if hasHuaweiFlavorDetails(row) {
			flattened = append(flattened, row)
		}
		for _, child := range nestedChildRows(row) {
			flattened = append(flattened, flattenHuaweiFlavorRows([]map[string]interface{}{child})...)
		}
	}
	return flattened
}

func nestedChildRows(row map[string]interface{}) []map[string]interface{} {
	items, ok := row["children"].([]interface{})
	if !ok {
		return nil
	}
	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if child, ok := item.(map[string]interface{}); ok {
			rows = append(rows, child)
		}
	}
	return rows
}

func hasHuaweiFlavorDetails(row map[string]interface{}) bool {
	return firstNonEmptyIntValue(row["vcpus"], row["flavor_vcpus"], row["ram_GB"], row["ram"], row["flavor_ram"], row["max_nic_num"], row["max_disk_num"]) != nil ||
		firstNonEmptyString(row["zone_id"], row["quota_rate"], row["quota_pps"], row["GHz"], row["ghz"]) != ""
}

func authAutoDetectedResources(data interface{}) []string {
	ordered := []string{
		"regions",
		"projects",
		"compute_zones",
		"flavors",
		"images",
		"boot_loader_images",
		"boot_loader_flavors",
		"system_volume_types",
		"zones",
		"networks",
		"subnets",
	}
	detected := make([]string, 0, len(ordered))
	seen := map[string]bool{}
	for _, resource := range ordered {
		if authResourceHasData(data, resource) {
			seen[resource] = true
			detected = append(detected, resource)
		}
	}

	extras := authExtraResourceKeys(data, seen)
	sort.Strings(extras)
	return append(detected, extras...)
}

func authExtraResourceKeys(data interface{}, seen map[string]bool) []string {
	keys := make([]string, 0)
	appendKey := func(key string) {
		if key == "" || key == "cloud_info" || seen[key] {
			return
		}
		if ResourceValue(data, key) == nil {
			return
		}
		seen[key] = true
		keys = append(keys, key)
	}

	if m, ok := data.(map[string]interface{}); ok {
		for key := range m {
			appendKey(key)
		}
		if cloudInfo, ok := m["cloud_info"].(map[string]interface{}); ok {
			for key := range cloudInfo {
				appendKey(key)
			}
		}
	}
	return keys
}

func authResourceHasData(data interface{}, resource string) bool {
	switch resource {
	case "regions", "zones", "images", "boot_loader_images", "flavors", "projects", "compute_zones", "boot_loader_flavors", "system_volume_types", "networks", "subnets":
		return len(authResourceRows(data, resource)) > 0
	default:
		return ResourceValue(data, resource) != nil
	}
}

func authResourceRows(data interface{}, resource string) []map[string]interface{} {
	switch resource {
	case "regions":
		return RegionRows(data)
	case "zones":
		return ZoneRows(data)
	case "images":
		return ImageRows(data)
	case "system_volume_types":
		return SystemVolumeTypeRows(data)
	case "boot_loader_images":
		return NormalizeImageRows(ResourceRows(data, resource))
	case "flavors", "projects", "compute_zones", "boot_loader_flavors", "networks", "subnets":
		return ResourceRows(data, resource)
	default:
		return nil
	}
}

func authResourceTitleKey(resource string) string {
	switch resource {
	case "regions", "zones", "flavors", "images", "boot_loader_images", "compute_zones", "projects", "networks", "subnets":
		return "resource." + resource
	case "boot_loader_flavors":
		return "resource.boot_loader_flavors"
	case "system_volume_types":
		return "resource.volume_types"
	default:
		return resource
	}
}

func firstNonEmptyIntValue(values ...interface{}) interface{} {
	for _, value := range values {
		switch t := value.(type) {
		case int:
			return t
		case int32:
			return int(t)
		case int64:
			return int(t)
		case float64:
			if t == float64(int64(t)) {
				return int(t)
			}
			return t
		case float32:
			if t == float32(int64(t)) {
				return int(t)
			}
			return t
		}
	}
	return nil
}
