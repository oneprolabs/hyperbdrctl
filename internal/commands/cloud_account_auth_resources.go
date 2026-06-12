package commands

import (
	"sort"
	"strings"

	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/normalize/cloudinfo"
	"hyperbdr-client/internal/output"
)

func writeAuthResourcesResponse(ctx *context, resp client.APIResponse, fetchRes string) error {
	data := authResourcesResponseData(resp)
	if ctx.cfg.Output == "json" {
		return writeResponse(ctx, resp, "", nil)
	}

	resources := normalizeAuthRequestedResources(fetchRes)
	if len(resources) == 0 {
		resources = authAutoDetectedResources(data)
	}
	if len(resources) == 0 {
		return writeValue(ctx, data)
	}

	for i, resource := range resources {
		if i > 0 {
			if _, err := ctx.out.Write([]byte("\n")); err != nil {
				return err
			}
		}
		if err := writeSectionTitle(ctx, ctx.loc.T(authResourceTitleKey(resource))); err != nil {
			return err
		}
		if err := writeAuthResourceBody(ctx, data, resource); err != nil {
			return err
		}
	}
	return nil
}

func authResourcesResponseData(resp client.APIResponse) interface{} {
	if resp.Data != nil {
		return resp.Data
	}
	return resp.Raw
}

func normalizeAuthRequestedResources(fetchRes string) []string {
	if strings.TrimSpace(fetchRes) == "" {
		return nil
	}
	parts := strings.Split(fetchRes, ",")
	resources := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		resource := normalizeAuthResourceName(part)
		if resource == "" || seen[resource] {
			continue
		}
		seen[resource] = true
		resources = append(resources, resource)
	}
	return resources
}

func normalizeAuthResourceName(name string) string {
	resource := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(name, "-", "_")))
	switch resource {
	case "region":
		return "regions"
	case "flavor":
		return "flavors"
	default:
		return resource
	}
}

func authAutoDetectedResources(data interface{}) []string {
	ordered := []string{
		"regions",
		"projects",
		"compute_zones",
		"flavors",
		"images",
		"boot_loader_images",
		"zones",
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
		if cloudinfo.ResourceValue(data, key) == nil {
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
	case "regions":
		return len(cloudinfo.RegionRows(data)) > 0
	case "zones":
		return len(cloudinfo.ZoneRows(data)) > 0
	case "images":
		return len(cloudinfo.ImageRows(data)) > 0
	case "boot_loader_images":
		return len(cloudinfo.NormalizeImageRows(cloudinfo.ResourceRows(data, resource))) > 0
	case "flavors", "projects", "compute_zones":
		return len(cloudinfo.ResourceRows(data, resource)) > 0
	default:
		return cloudinfo.ResourceValue(data, resource) != nil
	}
}

func authResourceTitleKey(resource string) string {
	switch resource {
	case "regions", "zones", "flavors", "images", "boot_loader_images", "compute_zones", "projects":
		return gatewayResourceTitleKey(resource)
	default:
		return resource
	}
}

func writeAuthResourceBody(ctx *context, data interface{}, resource string) error {
	switch resource {
	case "regions":
		return writeAuthTable(ctx, cloudinfo.RegionRows(data), regionColumns())
	case "zones":
		return writeAuthTable(ctx, cloudinfo.ZoneRows(data), gatewayZoneColumns())
	case "images":
		rows := cloudinfo.ImageRows(data)
		return writeAuthTable(ctx, rows, cloudAccountImageColumns(rows))
	case "boot_loader_images":
		rows := cloudinfo.NormalizeImageRows(cloudinfo.ResourceRows(data, resource))
		return writeAuthTable(ctx, rows, cloudAccountImageColumns(rows))
	case "flavors":
		return writeAuthTable(ctx, cloudinfo.ResourceRows(data, resource), gatewayFlavorColumns())
	case "projects":
		return writeAuthTable(ctx, cloudinfo.ResourceRows(data, resource), gatewayProjectColumns())
	case "compute_zones":
		return writeAuthTable(ctx, cloudinfo.ResourceRows(data, resource), gatewayComputeZoneColumns())
	default:
		return writeValue(ctx, cloudinfo.ResourceValue(data, resource))
	}
}

func writeAuthTable(ctx *context, rows []map[string]interface{}, cols []output.Column) error {
	return output.Table(ctx.out, ctx.loc, rows, gatewayVisibleColumns(rows, cols))
}
