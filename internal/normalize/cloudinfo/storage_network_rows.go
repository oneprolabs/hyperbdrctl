package cloudinfo

import "fmt"

func StorageNetworkRows(data interface{}) []map[string]interface{} {
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}
	flattened := flattenMaps(m)
	rows := make([]map[string]interface{}, 0)
	appendRows := func(kind string, value interface{}) {
		switch v := value.(type) {
		case []interface{}:
			for _, item := range v {
				rows = append(rows, map[string]interface{}{
					"type":  kind,
					"name":  networkRowName(item),
					"value": item,
				})
			}
		case []string:
			for _, item := range v {
				rows = append(rows, map[string]interface{}{
					"type":  kind,
					"name":  item,
					"value": item,
				})
			}
		case string:
			if v != "" {
				rows = append(rows, map[string]interface{}{
					"type":  kind,
					"name":  v,
					"value": v,
				})
			}
		}
	}
	for _, candidate := range []struct {
		key  string
		kind string
	}{
		{key: "network_addr_for_write_data", kind: "write"},
		{key: "network_addrs_for_write_data", kind: "write"},
		{key: "network_addr_for_read_data", kind: "read"},
		{key: "network_addrs_for_read_data", kind: "read"},
		{key: "write_networks", kind: "write"},
		{key: "read_networks", kind: "read"},
		{key: "networks", kind: "network"},
	} {
		if value, exists := flattened[candidate.key]; exists {
			appendRows(candidate.kind, value)
		}
	}
	return rows
}

func flattenMaps(root map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	var walk func(map[string]interface{})
	walk = func(current map[string]interface{}) {
		for key, value := range current {
			if _, exists := out[key]; !exists {
				out[key] = value
			}
			if nested, ok := value.(map[string]interface{}); ok {
				walk(nested)
			}
		}
	}
	walk(root)
	return out
}

func networkRowName(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case map[string]interface{}:
		for _, key := range []string{"name", "label", "address", "value", "network_addr"} {
			if s, ok := v[key].(string); ok && s != "" {
				return s
			}
		}
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
