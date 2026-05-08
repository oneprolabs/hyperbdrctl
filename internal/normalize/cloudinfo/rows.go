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

func NormalizeRegionRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		row["region_id"] = firstNonEmptyString(row["region_id"], row["id"], row["value"])
		row["region_name"] = firstNonEmptyString(row["region_name"], row["display_name"], row["name"], row["local_name"], row["region_id"])
		row["local_name"] = firstNonEmptyString(row["local_name"], row["display_name"], row["name"], row["region_name"])
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

func firstNonEmptyString(values ...interface{}) string {
	for _, value := range values {
		if s, ok := value.(string); ok && s != "" {
			return s
		}
	}
	return ""
}
