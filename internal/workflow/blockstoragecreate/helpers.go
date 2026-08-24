package blockstoragecreate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	minGatewayFlavorVCPUs  = 2
	minGatewayFlavorRAMGiB = 4
)

var compactFlavorSpecPattern = regexp.MustCompile(`(\d+)C[_-]?(\d+(?:\.\d+)?)G`)

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

func nestedMap(data interface{}, path ...string) map[string]interface{} {
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

func firstNonEmptyString(values ...interface{}) string {
	for _, value := range values {
		if s, ok := value.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func stringValue(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func valueOrNil(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}

func gatewayRegionRows(data interface{}) []map[string]interface{} {
	rows := listNestedMaps(data, "cloud_info", "regions")
	if len(rows) == 0 {
		rows = listNestedMaps(data, "cloud_info", "domain", "regions")
	}
	return rows
}

func gatewayZoneRows(data interface{}) []map[string]interface{} {
	rows := listNestedMaps(data, "cloud_info", "zones")
	if len(rows) > 0 {
		return rows
	}
	rows = listNestedMapsFromParentList(data, []string{"cloud_info", "regions"}, "zones")
	if len(rows) > 0 {
		return rows
	}
	return listNestedMapsFromParentList(data, []string{"cloud_info", "domain", "regions"}, "zones")
}

func gatewayImageRows(data interface{}) []map[string]interface{} {
	return listNestedMaps(data, "cloud_info", "images")
}

func selectGatewayRegionRow(data map[string]interface{}, regionID string) map[string]interface{} {
	for _, item := range gatewayRegionRows(data) {
		if firstNonEmptyString(item["region_id"], item["id"]) == regionID {
			return item
		}
	}
	return map[string]interface{}{
		"region_id":   regionID,
		"region_name": regionID,
	}
}

func selectGatewayImageRow(items []map[string]interface{}, imageID string) (map[string]interface{}, error) {
	if imageID != "" {
		return selectGatewayNamedRow(items, imageID, func(item map[string]interface{}) string {
			return firstNonEmptyString(item["image_id"], item["id"])
		})
	}
	for _, item := range items {
		if firstNonEmptyString(item["os_type"], item["__os_type"]) == "linux" {
			return item, nil
		}
	}
	if len(items) > 0 {
		return items[0], nil
	}
	return nil, fmt.Errorf("no candidates available")
}

func selectGatewaySubnetRow(items []map[string]interface{}, networkID, subnetID string) (map[string]interface{}, error) {
	filtered := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if networkID == "" || firstNonEmptyString(item["network_id"]) == networkID {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}
	return selectGatewayNamedRow(filtered, subnetID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["id"], item["subnet_id"])
	})
}

func selectGatewayNetworkRow(items, subnets []map[string]interface{}, want string, id func(map[string]interface{}) string) (map[string]interface{}, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}
	if want != "" {
		return selectGatewayNamedRow(items, want, id)
	}
	if len(subnets) > 0 {
		networksWithSubnets := make(map[string]struct{}, len(subnets))
		for _, subnet := range subnets {
			networkID := firstNonEmptyString(subnet["network_id"])
			if networkID != "" {
				networksWithSubnets[networkID] = struct{}{}
			}
		}
		for _, item := range items {
			itemID := id(item)
			if _, ok := networksWithSubnets[itemID]; ok {
				switch item["is_recommend"] {
				case float64(1), 1, "1", true:
					return item, nil
				}
			}
		}
		for _, item := range items {
			itemID := id(item)
			if _, ok := networksWithSubnets[itemID]; ok {
				return item, nil
			}
		}
	}
	return selectGatewayRecommendedRow(items, want, id)
}

func selectGatewayNamedRow(items []map[string]interface{}, want string, id func(map[string]interface{}) string) (map[string]interface{}, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}
	if want == "" {
		return items[0], nil
	}
	for _, item := range items {
		if id(item) == want {
			return item, nil
		}
	}
	return nil, fmt.Errorf("resource %q not found", want)
}

func selectGatewayRecommendedRow(items []map[string]interface{}, want string, id func(map[string]interface{}) string) (map[string]interface{}, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}
	if want != "" {
		return selectGatewayNamedRow(items, want, id)
	}
	for _, item := range items {
		switch item["is_recommend"] {
		case float64(1), 1, "1", true:
			return item, nil
		}
	}
	return items[0], nil
}

func selectGatewayFlavorRow(items []map[string]interface{}, want string, id func(map[string]interface{}) string) (map[string]interface{}, error) {
	return selectMinimumGatewayFlavorRow(items, want, id)
}

func selectOpenStackFlavorRow(items []map[string]interface{}, want string, id func(map[string]interface{}) string) (map[string]interface{}, error) {
	return selectMinimumGatewayFlavorRow(items, want, id)
}

func selectMinimumGatewayFlavorRow(items []map[string]interface{}, want string, id func(map[string]interface{}) string) (map[string]interface{}, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}
	if want != "" {
		row, err := selectGatewayNamedRow(items, want, id)
		if err != nil {
			return nil, err
		}
		if !gatewayFlavorMeetsMinimum(row) {
			return nil, fmt.Errorf("resource %q does not meet minimum requirement: at least %d vCPUs and %d GiB RAM", want, minGatewayFlavorVCPUs, minGatewayFlavorRAMGiB)
		}
		return row, nil
	}
	eligible := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if gatewayFlavorMeetsMinimum(item) {
			eligible = append(eligible, item)
		}
	}
	if len(eligible) == 0 {
		return nil, fmt.Errorf("no candidates available with minimum requirement: at least %d vCPUs and %d GiB RAM", minGatewayFlavorVCPUs, minGatewayFlavorRAMGiB)
	}
	return selectGatewayRecommendedRow(eligible, "", id)
}

func gatewayFlavorMeetsMinimum(item map[string]interface{}) bool {
	vcpus, ramGiB, ok := gatewayFlavorSpec(item)
	return ok && vcpus >= minGatewayFlavorVCPUs && ramGiB >= minGatewayFlavorRAMGiB
}

func gatewayFlavorSpec(item map[string]interface{}) (int, float64, bool) {
	vcpus, hasVCPUs := numericInt(
		item["vcpus"],
		item["cpu"],
	)
	ramGiB, hasRAM := numericFloat(
		item["ram_GB"],
		item["ram_gb"],
		item["memory_gb"],
	)
	if !hasRAM {
		if ramValue, ok := numericFloat(item["ram"]); ok {
			ramGiB = ramValue
			if ramGiB > 256 {
				ramGiB = ramGiB / 1024
			}
			hasRAM = true
		}
	}
	if hasVCPUs && hasRAM {
		return vcpus, ramGiB, true
	}
	name := strings.ToUpper(firstNonEmptyString(item["name"], item["display_name"], item["flavor_name"], item["id"]))
	match := compactFlavorSpecPattern.FindStringSubmatch(name)
	if len(match) == 3 {
		parsedVCPUs, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, 0, false
		}
		parsedRAM, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			return 0, 0, false
		}
		return parsedVCPUs, parsedRAM, true
	}
	return 0, 0, false
}

func numericInt(values ...interface{}) (int, bool) {
	for _, value := range values {
		switch v := value.(type) {
		case int:
			return v, true
		case int32:
			return int(v), true
		case int64:
			return int(v), true
		case float64:
			return int(v), true
		case string:
			if v == "" {
				continue
			}
			parsed, err := strconv.Atoi(v)
			if err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func numericFloat(values ...interface{}) (float64, bool) {
	for _, value := range values {
		switch v := value.(type) {
		case float64:
			return v, true
		case float32:
			return float64(v), true
		case int:
			return float64(v), true
		case int32:
			return float64(v), true
		case int64:
			return float64(v), true
		case string:
			if v == "" {
				continue
			}
			parsed, err := strconv.ParseFloat(v, 64)
			if err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func formatAliyunGatewayVCPU(value interface{}) string {
	switch v := value.(type) {
	case string:
		if v == "" {
			return ""
		}
		return v + " CPU"
	case float64:
		return strconv.FormatInt(int64(v), 10) + " CPU"
	case int:
		return strconv.Itoa(v) + " CPU"
	default:
		return ""
	}
}

func formatAliyunGatewayRAM(value interface{}) string {
	switch v := value.(type) {
	case string:
		if v == "" {
			return ""
		}
		return v + " GiB"
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64) + " GiB"
	case int:
		return strconv.Itoa(v) + " GiB"
	default:
		return ""
	}
}

func aliyunGatewayControlNetworkName(mode string) string {
	switch mode {
	case "floating_ip_without_proxy", "floating_ip_with_hg_proxy":
		return "\u516c\u7f51"
	case "fixed_ip_without_proxy", "fixed_ip_with_hg_proxy":
		return "\u5185\u7f51"
	default:
		return mode
	}
}

func aliyunDriverAdaptNetworkModeName(mode string) string {
	switch mode {
	case "floating_ip_without_proxy":
		return "\u516c\u7f51\u7f51\u7edc\u4e0d\u4f7f\u7528\u4ee3\u7406"
	case "fixed_ip_without_proxy":
		return "\u5185\u7f51\u7f51\u7edc\u4e0d\u4f7f\u7528\u4ee3\u7406"
	case "floating_ip_with_hg_proxy":
		return "\u516c\u7f51\u7f51\u7edc\u5e76\u901a\u8fc7\u4e91\u540c\u6b65\u7f51\u5173\u4ee3\u7406"
	case "fixed_ip_with_hg_proxy":
		return "\u5185\u7f51\u7f51\u7edc\u5e76\u901a\u8fc7\u4e91\u540c\u6b65\u7f51\u5173\u4ee3\u7406"
	default:
		return mode
	}
}

func selectOpenStackProject(cloudInfo, authInfo map[string]interface{}, projectID string) (map[string]interface{}, error) {
	defaultID := firstNonEmptyString(projectID, authInfo["project_id"])
	projects := openStackMapList(cloudInfo, "projects")
	project, err := selectOpenStackNamedResource(projects, defaultID, func(item map[string]interface{}) string {
		return stringValue(item["id"])
	})
	if err == nil {
		return project, nil
	}
	if defaultID != "" {
		return nil, fmt.Errorf("project-id: %w", err)
	}
	return nil, err
}

func selectOpenStackRegion(cloudInfo, authInfo map[string]interface{}, regionID string) (map[string]interface{}, error) {
	defaultID := firstNonEmptyString(regionID, authInfo["region_id"])
	regions := openStackMapList(cloudInfo, "regions")
	region, err := selectOpenStackNamedResource(regions, defaultID, func(item map[string]interface{}) string {
		return firstNonEmptyString(item["id"], item["name"])
	})
	if err == nil {
		return region, nil
	}
	if defaultID != "" {
		return nil, fmt.Errorf("region-id: %w", err)
	}
	return nil, err
}

func selectOpenStackImage(cloudInfo map[string]interface{}, imageID string) (map[string]interface{}, error) {
	images := openStackMapList(cloudInfo, "images")
	if imageID != "" {
		return selectOpenStackNamedResource(images, imageID, func(item map[string]interface{}) string {
			return stringValue(item["id"])
		})
	}
	for _, item := range images {
		if firstNonEmptyString(item["os_type"], item["__os_type"]) == "linux" {
			return item, nil
		}
	}
	if len(images) > 0 {
		return images[0], nil
	}
	return nil, fmt.Errorf("no candidate image returned from get_cloud_info")
}

func selectOpenStackSubnet(cloudInfo map[string]interface{}, networkID, subnetID string) (map[string]interface{}, error) {
	subnets := openStackMapList(cloudInfo, "subnets")
	filtered := make([]map[string]interface{}, 0, len(subnets))
	for _, subnet := range subnets {
		if networkID == "" || stringValue(subnet["network_id"]) == networkID {
			filtered = append(filtered, subnet)
		}
	}
	if subnetID != "" {
		return selectOpenStackNamedResource(filtered, subnetID, func(item map[string]interface{}) string {
			return stringValue(item["id"])
		})
	}
	for _, subnet := range filtered {
		if stringValue(subnet["id"]) == "" {
			return subnet, nil
		}
	}
	if len(filtered) > 0 {
		return filtered[0], nil
	}
	return nil, fmt.Errorf("no candidate subnet returned for network %s", networkID)
}

func openStackMapList(data map[string]interface{}, key string) []map[string]interface{} {
	items, _ := data[key].([]interface{})
	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]interface{}); ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func selectOpenStackNamedResource(items []map[string]interface{}, want string, id func(map[string]interface{}) string) (map[string]interface{}, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}
	if want == "" {
		return items[0], nil
	}
	for _, item := range items {
		if id(item) == want {
			return item, nil
		}
	}
	return nil, fmt.Errorf("resource %q not found", want)
}

func selectOpenStackRecommendedResource(items []map[string]interface{}, want string, id func(map[string]interface{}) string) (map[string]interface{}, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}
	if want != "" {
		return selectOpenStackNamedResource(items, want, id)
	}
	for _, item := range items {
		switch item["is_recommend"] {
		case float64(1), 1, "1", true:
			return item, nil
		}
	}
	return items[0], nil
}

func openStackSubnetName(subnet map[string]interface{}) string {
	return firstNonEmptyString(subnet["display_name"], subnet["name"], subnet["id"])
}

func openStackBootTypeName(bootTypeID string) string {
	switch bootTypeID {
	case "boot_from_volume":
		return "\u5377\u542f\u52a8"
	case "boot_from_image":
		return "\u955c\u50cf\u542f\u52a8"
	default:
		return bootTypeID
	}
}

func openStackGatewayNetworkName(networkMode string) string {
	switch networkMode {
	case "floating_ip_without_proxy":
		return "\u516c\u7f51"
	case "fixed_ip_without_proxy":
		return "\u5185\u7f51"
	case "floating_ip_with_proxy":
		return "\u516c\u7f51(\u4ee3\u7406)"
	case "fixed_ip_with_proxy":
		return "\u5185\u7f51(\u4ee3\u7406)"
	default:
		return networkMode
	}
}
