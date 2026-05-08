package commands

import (
	"fmt"
	"strings"

	appblockstorage "hyperbdr-client/internal/app/blockstorage"
	"hyperbdr-client/internal/client"
	normalizecloudinfo "hyperbdr-client/internal/normalize/cloudinfo"
	normalizegateway "hyperbdr-client/internal/normalize/gateway"
	"hyperbdr-client/internal/output"
)

func executeGatewayResources(ctx *context, accountID, fetchRes, regionID, zoneID, flavorID, flavorVCPUs, flavorRAM, purpose, imageType string) error {
	service := appblockstorage.NewService(commandAPIAdapter{ctx: ctx})
	result, err := service.Resources(appblockstorage.ResourcesSpec{
		CloudAccountID: accountID,
		FetchRes:       fetchRes,
		RegionID:       regionID,
		ZoneID:         zoneID,
		FlavorID:       flavorID,
		FlavorVCPUs:    flavorVCPUs,
		FlavorRAM:      flavorRAM,
		Purpose:        purpose,
		ImageType:      imageType,
	})
	if err != nil {
		return err
	}
	return writeGatewayResourcesResponse(ctx, result.Response, result.Resources)
}

func writeGatewayResourcesResponse(ctx *context, resp client.APIResponse, resources []string) error {
	data := gatewayResponseData(resp)
	if ctx.cfg.Output == "json" {
		return writeValue(ctx, data)
	}
	if len(resources) == 0 {
		return writeGatewayAutoDetectedResourceSections(ctx, data)
	}
	if len(resources) != 1 {
		return writeGatewayResourceSections(ctx, data, resources)
	}

	return writeGatewaySingleResourceSection(ctx, data, resources[0])
}

func writeGatewayAutoDetectedResourceSections(ctx *context, data interface{}) error {
	resources := gatewayAutoDetectedResources(data)
	if len(resources) == 0 {
		if currentDomain, ok := gatewayCurrentDomainValue(data); ok {
			return writeGatewayCurrentDomain(ctx, currentDomain)
		}
		return writeValue(ctx, data)
	}
	return writeGatewayResourceSections(ctx, data, resources)
}

func writeGatewayResourceSections(ctx *context, data interface{}, resources []string) error {
	if currentDomain, ok := gatewayCurrentDomainValue(data); ok {
		if err := writeGatewayCurrentDomain(ctx, currentDomain); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(ctx.out); err != nil {
			return err
		}
	}

	for i, resource := range resources {
		if i > 0 {
			if _, err := fmt.Fprintln(ctx.out); err != nil {
				return err
			}
		}
		if err := writeGatewaySectionTitle(ctx, resource); err != nil {
			return err
		}
		if err := writeGatewayResourceBody(ctx, data, resource); err != nil {
			return err
		}
	}
	return nil
}

func writeGatewaySingleResourceSection(ctx *context, data interface{}, resource string) error {
	if currentDomain, ok := gatewayCurrentDomainValue(data); ok {
		if err := writeGatewayCurrentDomain(ctx, currentDomain); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(ctx.out); err != nil {
			return err
		}
	}
	if err := writeGatewaySectionTitle(ctx, resource); err != nil {
		return err
	}
	return writeGatewayResourceBody(ctx, data, resource)
}

func writeGatewaySectionTitle(ctx *context, resource string) error {
	return writeSectionTitle(ctx, ctx.loc.T(gatewayResourceTitleKey(resource)))
}

func writeGatewayResourceBody(ctx *context, data interface{}, resource string) error {
	switch resource {
	case "regions":
		return writeGatewayTable(ctx, normalizegateway.RegionRows(data), regionColumns())
	case "zones":
		return writeGatewayTable(ctx, normalizegateway.ZoneRows(data), gatewayZoneColumns())
	case "flavors":
		return writeGatewayTable(ctx, normalizegateway.ListNestedMaps(data, "cloud_info", "flavors"), gatewayFlavorColumns())
	case "images":
		return writeGatewayTable(ctx, normalizegateway.ImageRows(data), imageColumns())
	case "boot_loader_images":
		return writeGatewayTable(ctx, normalizeGatewayImageRows(data, resource), imageColumns())
	case "compute_zones":
		return writeGatewayTable(ctx, normalizegateway.ComputeZoneRows(data), gatewayComputeZoneColumns())
	case "projects":
		return writeGatewayTable(ctx, normalizegateway.ProjectRows(data), gatewayProjectColumns())
	case "win_hd_images":
		return writeGatewayTable(ctx, normalizeGatewayImageRows(data, resource), imageColumns())
	case "linux_hd_images":
		return writeGatewayTable(ctx, normalizeGatewayImageRows(data, resource), imageColumns())
	case "volume_types":
		return writeGatewayTable(ctx, normalizegateway.VolumeTypeRows(data), gatewayVolumeTypeColumns())
	case "system_disk_types":
		return writeGatewayTable(ctx, normalizegateway.SystemDiskTypeRows(data), gatewaySystemDiskTypeColumns())
	case "networks":
		return writeGatewayTable(ctx, normalizegateway.ListNestedMaps(data, "cloud_info", "networks"), gatewayNetworkColumns())
	case "subnets":
		return writeGatewayTable(ctx, normalizegateway.SubnetRows(data), gatewaySubnetColumns())
	case "abilities":
		if abilities := normalizegateway.NestedMap(data, "cloud_info", "abilities"); abilities != nil {
			return writeValue(ctx, abilities)
		}
		return writeValue(ctx, map[string]interface{}{})
	default:
		return writeValue(ctx, data)
	}
}

func writeGatewayTable(ctx *context, rows []map[string]interface{}, cols []output.Column) error {
	return output.Table(ctx.out, ctx.loc, rows, gatewayVisibleColumns(rows, cols))
}

func gatewayVisibleColumns(rows []map[string]interface{}, cols []output.Column) []output.Column {
	visible := make([]output.Column, 0, len(cols))
	for _, col := range cols {
		if gatewayColumnHasData(rows, col.Field) {
			visible = append(visible, col)
		}
	}
	return visible
}

func gatewayColumnHasData(rows []map[string]interface{}, field string) bool {
	for _, row := range rows {
		if gatewayValueHasData(row[field]) {
			return true
		}
	}
	return false
}

func gatewayValueHasData(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(t) != ""
	case []interface{}:
		return len(t) > 0
	case []string:
		return len(t) > 0
	case map[string]interface{}:
		return len(t) > 0
	default:
		return true
	}
}

func writeGatewayCurrentDomain(ctx *context, currentDomain interface{}) error {
	return output.OrderedKeyValue(ctx.out, ctx.loc, []output.KeyValueRow{
		{Key: ctx.loc.T("table.current_domain"), Value: currentDomain},
	})
}

func gatewayAutoDetectedResources(data interface{}) []string {
	ordered := []string{
		"regions",
		"projects",
		"compute_zones",
		"flavors",
		"images",
		"boot_loader_images",
		"zones",
		"win_hd_images",
		"linux_hd_images",
		"volume_types",
		"system_disk_types",
		"networks",
		"subnets",
		"abilities",
	}
	detected := make([]string, 0, len(ordered))
	for _, resource := range ordered {
		if gatewayResourceHasData(data, resource) {
			detected = append(detected, resource)
		}
	}
	return detected
}

func gatewayResourceHasData(data interface{}, resource string) bool {
	switch resource {
	case "regions":
		return len(normalizegateway.RegionRows(data)) > 0
	case "zones":
		return len(normalizegateway.ZoneRows(data)) > 0
	case "flavors":
		return len(normalizegateway.ListNestedMaps(data, "cloud_info", "flavors")) > 0
	case "images":
		return len(normalizegateway.ImageRows(data)) > 0
	case "boot_loader_images", "win_hd_images", "linux_hd_images":
		return len(normalizeGatewayImageRows(data, resource)) > 0
	case "compute_zones":
		return len(normalizegateway.ComputeZoneRows(data)) > 0
	case "projects":
		return len(normalizegateway.ProjectRows(data)) > 0
	case "volume_types":
		return len(normalizegateway.VolumeTypeRows(data)) > 0
	case "system_disk_types":
		return len(normalizegateway.SystemDiskTypeRows(data)) > 0
	case "networks":
		return len(normalizegateway.ListNestedMaps(data, "cloud_info", "networks")) > 0
	case "subnets":
		return len(normalizegateway.SubnetRows(data)) > 0
	case "abilities":
		return normalizegateway.NestedMap(data, "cloud_info", "abilities") != nil
	default:
		return false
	}
}

func gatewayResourceTitleKey(resource string) string {
	switch resource {
	case "regions":
		return "resource.regions"
	case "zones":
		return "resource.zones"
	case "flavors":
		return "resource.flavors"
	case "images":
		return "resource.images"
	case "boot_loader_images":
		return "resource.boot_loader_images"
	case "compute_zones":
		return "resource.compute_zones"
	case "projects":
		return "resource.projects"
	case "win_hd_images":
		return "resource.win_hd_images"
	case "linux_hd_images":
		return "resource.linux_hd_images"
	case "volume_types":
		return "resource.volume_types"
	case "system_disk_types":
		return "resource.system_disk_types"
	case "networks":
		return "resource.networks"
	case "subnets":
		return "resource.subnets"
	case "abilities":
		return "resource.abilities"
	default:
		return resource
	}
}

func gatewayCurrentDomainValue(data interface{}) (interface{}, bool) {
	return normalizegateway.CurrentDomainValue(data)
}

func gatewayZoneColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.zone_id", Field: "id"},
		{HeaderKey: "table.zone_name", Field: "display_name"},
	}
}

func gatewayFlavorColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.flavor_id", Field: "id"},
		{HeaderKey: "table.flavor_name", Field: "name"},
		{HeaderKey: "table.vcpus", Field: "vcpus"},
		{HeaderKey: "table.ram_gb", Field: "ram_GB"},
		{HeaderKey: "table.max_nic_num", Field: "max_nic_num"},
		{HeaderKey: "table.nvme_support", Field: "nvme_support"},
	}
}

func gatewayComputeZoneColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.compute_zone_id", Field: "id"},
		{HeaderKey: "table.compute_zone_name", Field: "name"},
	}
}

func gatewayProjectColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.project_id", Field: "id"},
		{HeaderKey: "table.project_name", Field: "name"},
		{HeaderKey: "table.domain_id", Field: "domain_id"},
	}
}

func gatewayVolumeTypeColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
	}
}

func gatewaySystemDiskTypeColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.disk_type", Field: "id"},
		{HeaderKey: "table.display_name", Field: "display_name"},
		{HeaderKey: "table.min_gb", Field: "min_gb"},
		{HeaderKey: "table.max_gb", Field: "max_gb"},
	}
}

func gatewayNetworkColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.cidr_block", Field: "cidr_block"},
	}
}

func gatewaySubnetColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.network_id", Field: "network_id"},
		{HeaderKey: "table.zone_id", Field: "zone_id"},
		{HeaderKey: "table.cidr_block", Field: "cidr_block"},
	}
}

func gatewayStorageColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.storage_name", Field: "display_name"},
		{HeaderKey: "table.cloud_account_name", Field: "display_cloud_account_name"},
		{HeaderKey: "table.display_status", Field: "display_status"},
		{HeaderKey: "table.region_name", Field: "region_name"},
		{HeaderKey: "table.zone_name", Field: "zone_name"},
		{HeaderKey: "table.public_ip", Field: "public_ip"},
	}
}

func normalizeGatewayImageRows(data interface{}, key string) []map[string]interface{} {
	return normalizecloudinfo.NormalizeImageRows(normalizegateway.ListNestedMaps(data, "cloud_info", key))
}

func normalizeGatewayStorageRows(data interface{}) []map[string]interface{} {
	rows := listFromData(data, "storages")
	normalized := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		normalized = append(normalized, normalizeGatewayStorageRow(row))
	}
	return normalized
}

func normalizeGatewayStorageRow(row map[string]interface{}) map[string]interface{} {
	metadata, _ := row["metadata"].(map[string]interface{})
	return map[string]interface{}{
		"id": firstNonEmptyInterface(
			row["id"],
			row["uuid"],
			row["storage_uuid"],
		),
		"display_name": firstNonEmptyInterface(
			row["display_name"],
			row["display_cloud_storage_name"],
			row["storage_name"],
			row["name"],
		),
		"display_cloud_account_name": firstNonEmptyInterface(
			row["display_cloud_account_name"],
			row["cloud_account_name"],
			row["display_cloud_type"],
			row["cloud_account_uuid"],
		),
		"display_status": firstNonEmptyInterface(
			row["display_status"],
			row["status"],
		),
		"region_name": firstNonEmptyInterface(
			row["region_name"],
			row["display_region_name"],
			row["region_display_name"],
			metadata["region_name"],
			row["region_id"],
			metadata["region_id"],
		),
		"zone_name": firstNonEmptyInterface(
			row["zone_name"],
			metadata["zone_name"],
			row["zone_id"],
			metadata["zone_id"],
		),
		"public_ip": firstNonEmptyInterface(
			row["public_ip"],
			metadata["public_ip"],
			row["fixed_ip"],
			metadata["private_ip"],
		),
	}
}

func firstNonEmptyInterface(values ...interface{}) interface{} {
	for _, value := range values {
		if s, ok := value.(string); ok && s != "" {
			return s
		}
	}
	return ""
}
