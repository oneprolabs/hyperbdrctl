package commands

import (
	"fmt"

	"hyperbdr-client/catalog"
	"hyperbdr-client/internal/output"
)

func runCloudResourceCatalog(ctx *context, args []string) error {
	fs := newFlagSet("cloud-resource catalog")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if rest := fs.Args(); len(rest) > 0 {
		return errUnknown("cloud-resource catalog", rest[0])
	}

	grouped := map[string]interface{}{
		"block_clouds":  catalog.BlockClouds,
		"object_clouds": catalog.ObjectClouds,
	}
	if ctx.cfg.Output == "json" {
		return output.JSON(ctx.out, grouped)
	}

	if err := renderCloudResourceCatalogSection(ctx, ctx.loc.T("cmd.cloud_resource.catalog.block_title"), catalog.BlockClouds); err != nil {
		return err
	}
	fmt.Fprintln(ctx.out)
	return renderCloudResourceCatalogSection(ctx, ctx.loc.T("cmd.cloud_resource.catalog.object_title"), catalog.ObjectClouds)
}

func renderCloudResourceCatalogSection(ctx *context, title string, clouds []catalog.CloudEntry) error {
	if err := writeSectionTitle(ctx, title); err != nil {
		return err
	}
	return output.Table(ctx.out, ctx.loc, cloudCatalogRowsForDisplay(ctx, clouds), cloudResourceCatalogColumns())
}

func cloudCatalogRowsForDisplay(ctx *context, clouds []catalog.CloudEntry) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(clouds))
	for _, cloud := range clouds {
		name := cloud.NameEn
		if ctx.loc.Lang() == "zh_cn" {
			name = cloud.NameZhCN
		}
		rows = append(rows, map[string]interface{}{
			"provider": cloud.Provider,
			"name":     name,
		})
	}
	return rows
}

func cloudResourceCatalogColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.provider", Field: "provider"},
		{HeaderKey: "table.name", Field: "name"},
	}
}
