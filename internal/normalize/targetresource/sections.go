package targetresource

import "strings"

type Section struct {
	Resource string
	TitleKey string
	Rows     []map[string]interface{}
	Value    interface{}
}

func NormalizeSections(data interface{}, requested []string) []Section {
	resources := requested
	if len(resources) == 0 {
		resources = autoDetectedResources(data)
	}
	sections := make([]Section, 0, len(resources))
	for _, resource := range resources {
		section := normalizeSection(data, resource)
		if section.Resource == "" {
			continue
		}
		if len(section.Rows) == 0 && isEmptyValue(section.Value) {
			continue
		}
		sections = append(sections, section)
	}
	return sections
}

func normalizeSection(data interface{}, resource string) Section {
	resource = normalizeResourceName(resource)
	switch resource {
	case "regions":
		rows := normalizeRegionRows(resourceRows(data, "regions"))
		return Section{Resource: resource, TitleKey: "resource.regions", Rows: rows, Value: rows}
	case "zones":
		rows := normalizeZoneRows(resourceRows(data, "zones"))
		return Section{Resource: resource, TitleKey: "resource.zones", Rows: rows, Value: rows}
	case "flavors":
		rows := resourceRows(data, "flavors")
		return Section{Resource: resource, TitleKey: "resource.flavors", Rows: rows, Value: rows}
	case "images":
		rows := normalizeImageRows(resourceRows(data, "images"))
		return Section{Resource: resource, TitleKey: "resource.images", Rows: rows, Value: rows}
	case "boot_loader_images":
		rows := normalizeImageRows(resourceRows(data, "boot_loader_images"))
		return Section{Resource: resource, TitleKey: "resource.boot_loader_images", Rows: rows, Value: rows}
	case "boot_loader_flavors":
		rows := resourceRows(data, "boot_loader_flavors")
		return Section{Resource: resource, TitleKey: "resource.boot_loader_flavors", Rows: rows, Value: rows}
	case "win_hd_images":
		rows := normalizeImageRows(resourceRows(data, "win_hd_images"))
		return Section{Resource: resource, TitleKey: "resource.win_hd_images", Rows: rows, Value: rows}
	case "linux_hd_images":
		rows := normalizeImageRows(resourceRows(data, "linux_hd_images"))
		return Section{Resource: resource, TitleKey: "resource.linux_hd_images", Rows: rows, Value: rows}
	case "compute_zones":
		rows := resourceRows(data, "compute_zones")
		return Section{Resource: resource, TitleKey: "resource.compute_zones", Rows: rows, Value: rows}
	case "projects":
		rows := resourceRows(data, "projects")
		return Section{Resource: resource, TitleKey: "resource.projects", Rows: rows, Value: rows}
	case "volume_types":
		rows := resourceRows(data, "volume_types")
		return Section{Resource: resource, TitleKey: "resource.volume_types", Rows: rows, Value: rows}
	case "system_volume_types":
		rows := normalizeSystemVolumeTypeRows(resourceRows(data, "system_volume_types"))
		return Section{Resource: resource, TitleKey: "resource.system_volume_types", Rows: rows, Value: rows}
	case "system_disk_types":
		rows := normalizeSystemDiskTypeRows(resourceRows(data, "system_disk_types"))
		return Section{Resource: resource, TitleKey: "resource.system_disk_types", Rows: rows, Value: rows}
	case "networks":
		rows := resourceRows(data, "networks")
		return Section{Resource: resource, TitleKey: "resource.networks", Rows: rows, Value: rows}
	case "subnets":
		rows := normalizeSubnetRows(resourceRows(data, "subnets"))
		return Section{Resource: resource, TitleKey: "resource.subnets", Rows: rows, Value: rows}
	case "security_groups":
		rows := resourceRows(data, "security_groups")
		return Section{Resource: resource, TitleKey: "resource.security_groups", Rows: rows, Value: rows}
	case "os_types":
		rows := resourceRows(data, "os_types")
		return Section{Resource: resource, TitleKey: "resource.os_types", Rows: rows, Value: rows}
	case "abilities":
		value := nestedMap(data, "cloud_info", "abilities")
		if value == nil {
			value = mapValue(data, "abilities")
		}
		return Section{Resource: resource, TitleKey: "resource.abilities", Value: value}
	default:
		value := resourceValue(data, resource)
		return Section{Resource: resource, TitleKey: resource, Value: value}
	}
}

func autoDetectedResources(data interface{}) []string {
	ordered := []string{
		"regions",
		"projects",
		"compute_zones",
		"flavors",
		"boot_loader_flavors",
		"images",
		"boot_loader_images",
		"zones",
		"win_hd_images",
		"linux_hd_images",
		"volume_types",
		"system_volume_types",
		"system_disk_types",
		"networks",
		"subnets",
		"security_groups",
		"os_types",
		"abilities",
	}
	out := make([]string, 0, len(ordered))
	for _, resource := range ordered {
		section := normalizeSection(data, resource)
		if len(section.Rows) > 0 || !isEmptyValue(section.Value) {
			out = append(out, resource)
		}
	}
	return out
}

func resourceRows(data interface{}, key string) []map[string]interface{} {
	if rows := listNestedMaps(data, "cloud_info", key); len(rows) > 0 {
		return rows
	}
	if key == "regions" {
		if rows := listNestedMaps(data, "cloud_info", "domain", "regions"); len(rows) > 0 {
			return rows
		}
	}
	if key == "zones" {
		if rows := listNestedMapsFromParentList(data, []string{"cloud_info", "regions"}, "zones"); len(rows) > 0 {
			return rows
		}
		if rows := listNestedMapsFromParentList(data, []string{"cloud_info", "domain", "regions"}, "zones"); len(rows) > 0 {
			return rows
		}
	}
	return listFromData(data, key)
}

func resourceValue(data interface{}, key string) interface{} {
	if m := mapFromData(data); m != nil {
		if cloudInfo, ok := m["cloud_info"].(map[string]interface{}); ok {
			if value, ok := cloudInfo[key]; ok {
				return value
			}
		}
		if value, ok := m[key]; ok {
			return value
		}
	}
	return nil
}

func normalizeRegionRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		row["region_id"] = firstNonEmptyString(row["region_id"], row["id"], row["value"])
		row["region_name"] = firstNonEmptyString(row["region_name"], row["display_name"], row["name"], row["local_name"], row["region_id"])
		row["local_name"] = firstNonEmptyString(row["local_name"], row["display_name"], row["name"], row["region_name"])
	}
	return rows
}

func normalizeZoneRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		row["id"] = firstNonEmptyString(row["id"], row["zone_id"], row["value"])
		row["display_name"] = firstNonEmptyString(row["display_name"], row["zone_name"], row["name"], row["local_name"], row["id"])
	}
	return rows
}

func normalizeImageRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		row["image_id"] = firstNonEmptyString(row["image_id"], row["id"], row["uuid"])
		row["image_name"] = firstNonEmptyString(row["image_name"], row["name"], row["display_name"], row["image_id"])
		row["boot_mode"] = firstNonEmptyString(row["boot_mode"], row["boot_firmware"])
	}
	return rows
}

func normalizeSystemVolumeTypeRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		if row["min_gb"] == nil {
			row["min_gb"] = row["min_GB"]
		}
		if row["max_gb"] == nil {
			row["max_gb"] = row["max_GB"]
		}
	}
	return rows
}

func normalizeSystemDiskTypeRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		if row["min_gb"] == nil {
			row["min_gb"] = nestedValue(row, "system_ability", "min_GB")
		}
		if row["max_gb"] == nil {
			row["max_gb"] = nestedValue(row, "system_ability", "max_GB")
		}
	}
	return rows
}

func normalizeSubnetRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		if firstNonEmptyString(row["cidr_block"]) == "" {
			row["cidr_block"] = firstNonEmptyString(row["cidr"])
		}
	}
	return rows
}

func listNestedMaps(data interface{}, path ...string) []map[string]interface{} {
	value := data
	for _, key := range path {
		m, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}
		value = m[key]
	}
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}
	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]interface{}); ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func listNestedMapsFromParentList(data interface{}, parentPath []string, childKey string) []map[string]interface{} {
	parents := listNestedMaps(data, parentPath...)
	if len(parents) == 0 {
		return nil
	}
	rows := make([]map[string]interface{}, 0)
	for _, parent := range parents {
		items, ok := parent[childKey].([]interface{})
		if !ok {
			continue
		}
		for _, item := range items {
			if row, ok := item.(map[string]interface{}); ok {
				rows = append(rows, row)
			}
		}
	}
	return rows
}

func listFromData(data interface{}, key string) []map[string]interface{} {
	if items, ok := data.([]interface{}); ok {
		rows := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			if row, ok := item.(map[string]interface{}); ok {
				rows = append(rows, row)
			}
		}
		return rows
	}
	m := mapFromData(data)
	if m == nil {
		return nil
	}
	items, ok := m[key].([]interface{})
	if !ok {
		return nil
	}
	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]interface{}); ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func mapFromData(data interface{}) map[string]interface{} {
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func mapValue(data interface{}, key string) map[string]interface{} {
	m := mapFromData(data)
	if m == nil {
		return nil
	}
	if value, ok := m[key].(map[string]interface{}); ok {
		return value
	}
	return nil
}

func nestedMap(data interface{}, path ...string) map[string]interface{} {
	value := data
	for _, key := range path {
		m, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}
		value = m[key]
	}
	if m, ok := value.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func nestedValue(data interface{}, path ...string) interface{} {
	value := data
	for _, key := range path {
		m, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}
		value = m[key]
	}
	return value
}

func firstNonEmptyString(values ...interface{}) string {
	for _, value := range values {
		if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
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

func isEmptyValue(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(v) == ""
	case []interface{}:
		return len(v) == 0
	case []map[string]interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}
