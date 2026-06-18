package commands

import (
	"fmt"
	"strings"

	apptargetresource "hyperbdr-client/internal/app/targetresource"
	normalizetargetresource "hyperbdr-client/internal/normalize/targetresource"
	"hyperbdr-client/internal/output"
)

func writeTargetResourceResponse(ctx *context, result apptargetresource.Result) error {
	if ctx.cfg.Output == "json" {
		return writeResponse(ctx, result.Response, "", nil)
	}

	data := result.Response.Data
	if data == nil && result.Response.Raw != nil {
		data = result.Response.Raw
	}
	sections := normalizetargetresource.NormalizeSections(data, result.Resources)
	if len(sections) == 0 {
		return writeValue(ctx, data)
	}

	for i, section := range sections {
		if i > 0 {
			if _, err := fmt.Fprintln(ctx.out); err != nil {
				return err
			}
		}
		if err := writeSectionTitle(ctx, ctx.loc.T(section.TitleKey)); err != nil {
			return err
		}
		if err := writeTargetResourceSection(ctx, section); err != nil {
			return err
		}
	}
	return nil
}

func writeTargetResourceSection(ctx *context, section normalizetargetresource.Section) error {
	if len(section.Rows) == 0 {
		return writeValue(ctx, section.Value)
	}
	return output.Table(ctx.out, ctx.loc, section.Rows, visibleColumns(section.Rows, targetResourceColumns(section.Resource)))
}

func targetResourceColumns(resource string) []output.Column {
	switch resource {
	case "regions":
		return []output.Column{
			{HeaderKey: "table.region_id", Field: "region_id"},
			{HeaderKey: "table.region_name", Field: "region_name"},
			{HeaderKey: "table.local_name", Field: "local_name"},
		}
	case "zones":
		return []output.Column{
			{HeaderKey: "table.zone_id", Field: "id"},
			{HeaderKey: "table.zone_name", Field: "display_name"},
		}
	case "flavors", "boot_loader_flavors":
		return []output.Column{
			{HeaderKey: "table.flavor_id", Field: "id"},
			{HeaderKey: "table.flavor_name", Field: "name"},
			{HeaderKey: "table.vcpus", Field: "vcpus"},
			{HeaderKey: "table.ram_gb", Field: "ram_GB"},
			{HeaderKey: "table.zone_id", Field: "zone_id"},
			{HeaderKey: "table.max_nic_num", Field: "max_nic_num"},
		}
	case "images", "boot_loader_images", "win_hd_images", "linux_hd_images":
		return []output.Column{
			{HeaderKey: "table.image_id", Field: "image_id"},
			{HeaderKey: "table.image_name", Field: "image_name"},
			{HeaderKey: "table.os_type", Field: "os_type"},
			{HeaderKey: "table.os_version", Field: "os_version"},
			{HeaderKey: "table.boot_mode", Field: "boot_mode"},
		}
	case "compute_zones":
		return []output.Column{
			{HeaderKey: "table.compute_zone_id", Field: "id"},
			{HeaderKey: "table.compute_zone_name", Field: "name"},
		}
	case "projects":
		return []output.Column{
			{HeaderKey: "table.project_id", Field: "id"},
			{HeaderKey: "table.project_name", Field: "name"},
			{HeaderKey: "table.domain_id", Field: "domain_id"},
		}
	case "volume_types":
		return []output.Column{
			{HeaderKey: "table.id", Field: "id"},
			{HeaderKey: "table.name", Field: "name"},
		}
	case "system_volume_types", "system_disk_types":
		return []output.Column{
			{HeaderKey: "table.disk_type", Field: "id"},
			{HeaderKey: "table.display_name", Field: "display_name"},
			{HeaderKey: "table.min_gb", Field: "min_gb"},
			{HeaderKey: "table.max_gb", Field: "max_gb"},
		}
	case "networks":
		return []output.Column{
			{HeaderKey: "table.id", Field: "id"},
			{HeaderKey: "table.name", Field: "name"},
			{HeaderKey: "table.cidr_block", Field: "cidr_block"},
		}
	case "subnets":
		return []output.Column{
			{HeaderKey: "table.id", Field: "id"},
			{HeaderKey: "table.name", Field: "name"},
			{HeaderKey: "table.network_id", Field: "network_id"},
			{HeaderKey: "table.zone_id", Field: "zone_id"},
			{HeaderKey: "table.cidr_block", Field: "cidr_block"},
		}
	case "security_groups":
		return []output.Column{
			{HeaderKey: "table.id", Field: "id"},
			{HeaderKey: "table.name", Field: "name"},
			{HeaderKey: "table.display_name", Field: "display_name"},
		}
	case "os_types":
		return []output.Column{
			{HeaderKey: "table.id", Field: "id"},
			{HeaderKey: "table.name", Field: "name"},
			{HeaderKey: "table.os_type", Field: "os_type"},
		}
	default:
		return []output.Column{
			{HeaderKey: "table.id", Field: "id"},
			{HeaderKey: "table.name", Field: "name"},
		}
	}
}

func visibleColumns(rows []map[string]interface{}, cols []output.Column) []output.Column {
	visible := make([]output.Column, 0, len(cols))
	for _, col := range cols {
		if columnHasData(rows, col.Field) {
			visible = append(visible, col)
		}
	}
	return visible
}

func columnHasData(rows []map[string]interface{}, field string) bool {
	for _, row := range rows {
		if valueHasData(row[field]) {
			return true
		}
	}
	return false
}

func valueHasData(v interface{}) bool {
	switch value := v.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(value) != ""
	case []interface{}:
		return len(value) > 0
	case []string:
		return len(value) > 0
	case map[string]interface{}:
		return len(value) > 0
	default:
		return true
	}
}
