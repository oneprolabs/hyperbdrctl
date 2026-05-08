package gateway

import (
	"fmt"

	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/normalize/cloudinfo"
)

func Data(resp client.APIResponse) interface{} {
	if payload, ok := resp.Data.(map[string]interface{}); ok {
		return payload
	}
	if resp.Raw != nil {
		return resp.Raw
	}
	return resp.Data
}

func CurrentDomainValue(data interface{}) (interface{}, bool) {
	currentDomain := nestedValue(mapFromData(data), "cloud_info", "current_domain")
	if currentDomain == nil {
		return nil, false
	}
	if currentMap, ok := currentDomain.(map[string]interface{}); ok {
		id := firstNonEmptyString(currentMap["region_id"], currentMap["id"], currentMap["value"])
		name := firstNonEmptyString(currentMap["region_name"], currentMap["display_name"], currentMap["name"], currentMap["local_name"])
		switch {
		case id != "" && name != "" && id != name:
			return fmt.Sprintf("%s (%s)", id, name), true
		case id != "":
			return id, true
		case name != "":
			return name, true
		default:
			return currentMap, true
		}
	}
	return currentDomain, true
}

func RegionRows(data interface{}) []map[string]interface{} {
	rows := ListNestedMaps(data, "cloud_info", "regions")
	if len(rows) == 0 {
		rows = ListNestedMaps(data, "cloud_info", "domain", "regions")
	}
	return cloudinfo.NormalizeRegionRows(rows)
}

func ZoneRows(data interface{}) []map[string]interface{} {
	rows := ListNestedMaps(data, "cloud_info", "zones")
	if len(rows) > 0 {
		return rows
	}
	rows = ListNestedMapsFromParentList(data, []string{"cloud_info", "regions"}, "zones")
	if len(rows) > 0 {
		return rows
	}
	return ListNestedMapsFromParentList(data, []string{"cloud_info", "domain", "regions"}, "zones")
}

func ImageRows(data interface{}) []map[string]interface{} {
	return cloudinfo.NormalizeImageRows(ListNestedMaps(data, "cloud_info", "images"))
}

func SystemDiskTypeRows(data interface{}) []map[string]interface{} {
	rows := ListNestedMaps(data, "cloud_info", "system_disk_types")
	for _, row := range rows {
		row["min_gb"] = nestedValue(row, "system_ability", "min_GB")
		row["max_gb"] = nestedValue(row, "system_ability", "max_GB")
	}
	return rows
}

func ComputeZoneRows(data interface{}) []map[string]interface{} {
	return ListNestedMaps(data, "cloud_info", "compute_zones")
}

func ProjectRows(data interface{}) []map[string]interface{} {
	return ListNestedMaps(data, "cloud_info", "projects")
}

func VolumeTypeRows(data interface{}) []map[string]interface{} {
	return ListNestedMaps(data, "cloud_info", "volume_types")
}

func SubnetRows(data interface{}) []map[string]interface{} {
	rows := ListNestedMaps(data, "cloud_info", "subnets")
	for _, row := range rows {
		if firstNonEmptyString(row["cidr_block"]) == "" {
			if cidr := firstNonEmptyString(row["cidr"]); cidr != "" {
				row["cidr_block"] = cidr
			}
		}
	}
	return rows
}

func ListNestedMaps(data interface{}, path ...string) []map[string]interface{} {
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

func ListNestedMapsFromParentList(data interface{}, parentPath []string, childKey string) []map[string]interface{} {
	parents := ListNestedMaps(data, parentPath...)
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

func NestedMap(data interface{}, path ...string) map[string]interface{} {
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

func nestedValue(m map[string]interface{}, path ...string) interface{} {
	var value interface{} = m
	for _, key := range path {
		next, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}
		value = next[key]
	}
	return value
}

func mapFromData(data interface{}) map[string]interface{} {
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func firstNonEmptyString(values ...interface{}) string {
	for _, value := range values {
		if s, ok := value.(string); ok && s != "" {
			return s
		}
	}
	return ""
}
