package commands

import (
	appupgrade "hyperbdr-client/internal/app/upgrade"
	"hyperbdr-client/internal/output"
)

func runUpgrade(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("upgrade", "")
	}
	service := appupgrade.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "host":
		fs := newFlagSet("upgrade host")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 10, "")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.HostList(appupgrade.HostListSpec{
			Page:     *page,
			PageSize: *pageSize,
			Query:    q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "hosts", upgradeHostColumns())
	default:
		return errUnknown("upgrade", args[0])
	}
}

func upgradeHostColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.status", Field: "status"},
		{HeaderKey: "table.os_type", Field: "os_type"},
	}
}
