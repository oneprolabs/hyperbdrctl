package commands

import (
	"fmt"

	appbootconfig "hyperbdr-client/internal/app/bootconfig"
)

func runBatchBootConfig(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("batch-boot-config", "")
	}
	service := appbootconfig.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "create":
		fs := newFlagSet("batch-boot-config create")
		file := fs.String("file", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *file == "" {
			return fmt.Errorf("file is required")
		}
		body, err := readObjectFile(*file)
		if err != nil {
			return err
		}
		resp, err := service.BatchCreate(body)
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	case "get":
		fs := newFlagSet("batch-boot-config get")
		ids := fs.String("ids", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		hostIDs := splitCSV(*ids)
		if len(hostIDs) == 0 {
			return fmt.Errorf("ids is required")
		}
		resp, err := service.BatchGet(hostIDs)
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "boot_configs", drConfigColumns())
	case "update":
		fs := newFlagSet("batch-boot-config update")
		file := fs.String("file", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *file == "" {
			return fmt.Errorf("file is required")
		}
		body, err := readObjectFile(*file)
		if err != nil {
			return err
		}
		resp, err := service.BatchUpdate(body)
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	default:
		return errUnknown("batch-boot-config", args[0])
	}
}
