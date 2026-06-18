package commands

import (
	"fmt"
	applicense "hyperbdr-client/internal/app/license"
	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/output"
)

func runLicenses(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("license", "")
	}
	service := applicense.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "list":
		fs := newFlagSet("license list")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 10, "")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.List(applicense.ListSpec{
			Page:     *page,
			PageSize: *pageSize,
			Query:    q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "pages", licenseColumns())
	case "reg-code":
		fs := newFlagSet("license reg-code")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.RegCode(applicense.RegCodeSpec{Query: q})
		if err != nil {
			return err
		}
		return writeLicenseRegCodeResponse(ctx, resp)
	case "activate":
		fs := newFlagSet("license activate")
		kkty := fs.String("kkty", "", "")
		ddty := fs.String("ddty", "", "")
		file := fs.String("file", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		body, err := service.ActivateBody(applicense.ActivateSpec{
			KKTY: *kkty,
			DDTY: *ddty,
			File: *file,
		})
		if err != nil {
			return err
		}
		if body["kkty"] == "" {
			return missing(ctx, "error.missing_kkty")
		}
		if body["ddty"] == "" {
			return missing(ctx, "error.missing_ddty")
		}
		resp, err := service.Activate(body)
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	default:
		return errUnknown("license", args[0])
	}
}

func licenseColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.amount", Field: "amount"},
		{HeaderKey: "table.unused", Field: "unused"},
		{HeaderKey: "table.used", Field: "used"},
		{HeaderKey: "table.status", Field: "status"},
		{HeaderKey: "table.display_status", Field: "display_status"},
		{HeaderKey: "table.start_at", Field: "start_at"},
		{HeaderKey: "table.expire_at", Field: "expire_at"},
	}
}

func writeLicenseRegCodeResponse(ctx *context, resp client.APIResponse) error {
	if ctx.cfg.Output == "json" {
		return writeResponse(ctx, resp, "", nil)
	}

	code := licenseRegCodeValue(resp)
	if err := writeSectionTitle(ctx, ctx.loc.T("table.kkty")); err != nil {
		return err
	}
	_, err := fmt.Fprintln(ctx.out, code)
	return err
}

func licenseRegCodeValue(resp client.APIResponse) string {
	if data := mapFromData(resp.Data); data != nil {
		if code := mapString(data, "kkty"); code != "" {
			return code
		}
	}
	if resp.Raw != nil {
		if data := nestedMap(resp.Raw, "data"); data != nil {
			return mapString(data, "kkty")
		}
	}
	return ""
}
