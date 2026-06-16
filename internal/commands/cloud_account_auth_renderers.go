package commands

import (
	"fmt"
	"strconv"
	"strings"

	"hyperbdr-client/internal/normalize/cloudinfo"
	"hyperbdr-client/internal/output"
)

type authResourceRenderer interface {
	Normalize(section cloudinfo.ResourceSection) ([]map[string]interface{}, error)
	Columns(rows []map[string]interface{}) []output.Column
}

type authTableRenderer struct {
	normalize func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error)
	columns   func(rows []map[string]interface{}) []output.Column
}

func (r authTableRenderer) Normalize(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) {
	return r.normalize(section)
}

func (r authTableRenderer) Columns(rows []map[string]interface{}) []output.Column {
	return r.columns(rows)
}

func resolveAuthResourceRenderer(provider, cloudType, storageType string, section cloudinfo.ResourceSection) authResourceRenderer {
	keys := []string{
		authRendererKey(strings.ToLower(cloudType), section.Resource),
		authRendererKey(authProviderFamily(provider, cloudType), section.Resource),
		authRendererKey(strings.ToLower(storageType), section.Resource),
		authRendererKey("generic", section.Resource),
	}
	for _, key := range keys {
		if renderer, ok := authResourceRenderers[key]; ok {
			return renderer
		}
	}
	return nil
}

func authRendererKey(scope, resource string) string {
	scope = strings.TrimSpace(strings.ToLower(scope))
	if scope == "" {
		return ""
	}
	return scope + "/" + resource
}

func authProviderFamily(provider, cloudType string) string {
	switch strings.ToLower(provider) {
	case "huawei", "huaweicloud":
		return "huawei"
	}
	switch {
	case strings.HasPrefix(strings.ToLower(cloudType), "huawei"),
		strings.HasPrefix(strings.ToLower(cloudType), "huaweicloud"):
		return "huawei"
	default:
		return strings.ToLower(provider)
	}
}

var authResourceRenderers = map[string]authResourceRenderer{
	authRendererKey("generic", "regions"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return regionColumns() },
	},
	authRendererKey("generic", "zones"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return gatewayZoneColumns() },
	},
	authRendererKey("generic", "images"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return cloudAccountImageColumns(rows) },
	},
	authRendererKey("generic", "boot_loader_images"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return cloudAccountImageColumns(rows) },
	},
	authRendererKey("generic", "flavors"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return gatewayFlavorColumns() },
	},
	authRendererKey("generic", "os_types"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return bootConfigOSTypeColumns() },
	},
	authRendererKey("generic", "projects"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return gatewayProjectColumns() },
	},
	authRendererKey("generic", "compute_zones"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return gatewayComputeZoneColumns() },
	},
	authRendererKey("generic", "system_volume_types"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return gatewaySystemDiskTypeColumns() },
	},
	authRendererKey("generic", "volume_types"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return gatewayVolumeTypeColumns() },
	},
	authRendererKey("generic", "networks"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return gatewayNetworkColumns() },
	},
	authRendererKey("generic", "subnets"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return gatewaySubnetColumns() },
	},
	authRendererKey("generic", "security_groups"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) { return section.Rows, nil },
		columns:   func(rows []map[string]interface{}) []output.Column { return bootConfigSecurityGroupColumns() },
	},
	authRendererKey("huawei", "flavors"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) {
			return filterHuaweiFlavorRows(cloudinfo.NormalizeHuaweiFlavorRows(section.Rows), section.Meta)
		},
		columns: func(rows []map[string]interface{}) []output.Column {
			return huaweiFlavorColumns()
		},
	},
	authRendererKey("huawei", "boot_loader_flavors"): authTableRenderer{
		normalize: func(section cloudinfo.ResourceSection) ([]map[string]interface{}, error) {
			return cloudinfo.NormalizeHuaweiFlavorRows(section.Rows), nil
		},
		columns: func(rows []map[string]interface{}) []output.Column {
			return huaweiFlavorColumns()
		},
	},
}

func huaweiFlavorColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.flavor_id", Field: "id"},
		{HeaderKey: "table.flavor_name", Field: "name"},
		{HeaderKey: "table.vcpus", Field: "vcpus"},
		{HeaderKey: "table.ram_gb", Field: "ram_gb"},
		{HeaderKey: "table.zone_id", Field: "zone_id"},
		{HeaderKey: "table.ghz", Field: "ghz"},
		{HeaderKey: "table.quota_rate", Field: "quota_rate"},
		{HeaderKey: "table.quota_pps", Field: "quota_pps"},
		{HeaderKey: "table.max_nic_num", Field: "max_nic_num"},
		{HeaderKey: "table.max_disk_num", Field: "max_disk_num"},
	}
}

func bootConfigOSTypeColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.os_type", Field: "os_type"},
		{HeaderKey: "table.display_name", Field: "display_name"},
	}
}

func bootConfigSecurityGroupColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.network_id", Field: "network_id"},
	}
}

func filterHuaweiFlavorRows(rows []map[string]interface{}, meta map[string]interface{}) ([]map[string]interface{}, error) {
	wantVCPUs, hasVCPUs, err := optionalIntFilter(meta["flavor_vcpus"], "flavor-vcpus")
	if err != nil {
		return nil, err
	}
	wantRAM, hasRAM, err := optionalIntFilter(meta["flavor_ram"], "flavor-ram")
	if err != nil {
		return nil, err
	}
	if !hasVCPUs && !hasRAM {
		return rows, nil
	}

	filtered := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		if hasVCPUs {
			got, ok := scalarInt(row["vcpus"])
			if !ok || got != wantVCPUs {
				continue
			}
		}
		if hasRAM {
			got, ok := scalarInt(row["ram_gb"])
			if !ok || got != wantRAM {
				continue
			}
		}
		filtered = append(filtered, row)
	}
	return filtered, nil
}

func optionalIntFilter(v interface{}, flagName string) (int, bool, error) {
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return 0, false, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, false, fmt.Errorf("invalid %s %q", flagName, s)
	}
	return parsed, true, nil
}

func scalarInt(v interface{}) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int32:
		return int(t), true
	case int64:
		return int(t), true
	case float64:
		if t == float64(int64(t)) {
			return int(t), true
		}
	case float32:
		if t == float32(int64(t)) {
			return int(t), true
		}
	}
	return 0, false
}
