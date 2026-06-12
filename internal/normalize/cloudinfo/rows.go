package cloudinfo

func RegionRows(data interface{}) []map[string]interface{} {
	rows := listNestedMaps(data, "cloud_info", "regions")
	if len(rows) == 0 {
		rows = listNestedMaps(data, "cloud_info", "domain", "regions")
	}
	if len(rows) == 0 {
		rows = listFromData(data, "regions")
	}
	return NormalizeRegionRows(rows)
}

func ImageRows(data interface{}) []map[string]interface{} {
	rows := listNestedMaps(data, "cloud_info", "images")
	if len(rows) == 0 {
		rows = listFromData(data, "images")
	}
	return NormalizeImageRows(rows)
}

func ZoneRows(data interface{}) []map[string]interface{} {
	rows := listNestedMaps(data, "cloud_info", "zones")
	if len(rows) == 0 {
		rows = listNestedMapsFromParentList(data, []string{"cloud_info", "regions"}, "zones")
	}
	if len(rows) == 0 {
		rows = listNestedMapsFromParentList(data, []string{"cloud_info", "domain", "regions"}, "zones")
	}
	if len(rows) == 0 {
		rows = listFromData(data, "zones")
	}
	return NormalizeZoneRows(rows)
}

func ResourceRows(data interface{}, key string) []map[string]interface{} {
	rows := listNestedMaps(data, "cloud_info", key)
	if len(rows) > 0 {
		return rows
	}
	if key == "regions" {
		rows = listNestedMaps(data, "cloud_info", "domain", "regions")
		if len(rows) > 0 {
			return rows
		}
	}
	return listFromData(data, key)
}

func ResourceValue(data interface{}, key string) interface{} {
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

func NormalizeRegionRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		row["region_id"] = firstNonEmptyString(row["region_id"], row["id"], row["value"])
		row["region_name"] = firstNonEmptyString(row["region_name"], row["display_name"], row["name"], row["local_name"], row["region_id"])
		row["local_name"] = firstNonEmptyString(row["local_name"], row["display_name"], row["name"], row["region_name"])
	}
	return rows
}

func NormalizeZoneRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		row["id"] = firstNonEmptyString(row["id"], row["zone_id"], row["value"])
		row["display_name"] = firstNonEmptyString(row["display_name"], row["zone_name"], row["name"], row["local_name"], row["id"])
	}
	return rows
}

func NormalizeImageRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		row["image_id"] = firstNonEmptyString(row["image_id"], row["id"], row["uuid"])
		row["image_name"] = firstNonEmptyString(row["image_name"], row["name"], row["display_name"], row["image_id"])
		row["boot_mode"] = firstNonEmptyString(row["boot_mode"], row["boot_firmware"])
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
	m, ok := data.(map[string]interface{})
	if !ok {
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

func firstNonEmptyString(values ...interface{}) string {
	for _, value := range values {
		if s, ok := value.(string); ok && s != "" {
			return s
		}
	}
	return ""
}
