package commands

import (
	"fmt"

	"hyperbdr-client/catalog"
	"hyperbdr-client/internal/output"
)

func runTargetSupports(ctx *context, args []string) error {
	fs := newFlagSet("target supports")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if rest := fs.Args(); len(rest) > 0 {
		return errUnknown("target supports", rest[0])
	}

	grouped := map[string]interface{}{
		"block_clouds":  catalog.BlockClouds,
		"object_clouds": catalog.ObjectClouds,
	}
	if ctx.cfg.Output == "json" {
		return output.JSON(ctx.out, grouped)
	}

	if err := renderTargetSupportsSection(ctx, ctx.loc.T("cmd.target.supports.block_title"), catalog.BlockClouds); err != nil {
		return err
	}
	fmt.Fprintln(ctx.out)
	return renderTargetSupportsSection(ctx, ctx.loc.T("cmd.target.supports.object_title"), catalog.ObjectClouds)
}

func renderTargetSupportsSection(ctx *context, title string, clouds []catalog.CloudEntry) error {
	if err := writeSectionTitle(ctx, title); err != nil {
		return err
	}
	return output.Table(ctx.out, ctx.loc, cloudRowsForDisplay(ctx, clouds), targetSupportColumns())
}

func cloudRowsForDisplay(ctx *context, clouds []catalog.CloudEntry) []map[string]interface{} {
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

func targetSupportColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.provider", Field: "provider"},
		{HeaderKey: "table.name", Field: "name"},
	}
}
