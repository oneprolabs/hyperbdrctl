package commands

import (
	appbootconfig "hyperbdr-client/internal/app/bootconfig"
	"hyperbdr-client/internal/output"
)

func runBootConfig(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("boot-config", "")
	}
	service := appbootconfig.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "get":
		fs := newFlagSet("boot-config get")
		id := fs.String("id", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *id == "" {
			return missing(ctx, "error.missing_id")
		}
		resp, err := service.Get(*id)
		if err != nil {
			return err
		}
		if ctx.cfg.Output == "json" {
			return writeResponse(ctx, resp, "boot_configs", drConfigColumns())
		}
		rows := listFromData(resp.Data, "boot_configs")
		if len(rows) == 0 {
			return writeResponse(ctx, resp, "boot_configs", drConfigColumns())
		}
		fields := appbootconfig.DetailFields(*id, rows[0])
		view := make([]output.KeyValueRow, 0, len(fields))
		for _, field := range fields {
			view = append(view, output.KeyValueRow{Key: field.Key, Value: field.Value})
		}
		return output.OrderedKeyValue(ctx.out, ctx.loc, view)
	default:
		return errUnknown("boot-config", args[0])
	}
}

func drConfigColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.migration_id", Field: "migration_id"},
		{HeaderKey: "table.storage_id", Field: "storage_id"},
		{HeaderKey: "table.storage_name", Field: "storage_name"},
		{HeaderKey: "table.pool_id", Field: "pool_id"},
		{HeaderKey: "table.pool_name", Field: "pool_name"},
	}
}
