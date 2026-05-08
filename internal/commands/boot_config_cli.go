package commands

import appbootconfig "hyperbdr-client/internal/app/bootconfig"

func runBootConfigCLI(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("boot-config-cli", "")
	}
	service := appbootconfig.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "apply":
		fs := newFlagSet("boot-config-cli apply")
		id := fs.String("id", "", "")
		file := fs.String("file", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *id == "" {
			return missing(ctx, "error.missing_id")
		}
		metadata, err := appbootconfig.MetadataFile(*file)
		if err != nil {
			return err
		}
		result, err := service.Apply(*id, metadata)
		if err != nil {
			return err
		}
		return writeBootConfigMutationResponse(ctx, result)
	default:
		return errUnknown("boot-config-cli", args[0])
	}
}

func writeBootConfigMutationResponse(ctx *context, result appbootconfig.MutationResult) error {
	if ctx.cfg.Output == "json" {
		return writeResponse(ctx, result.Response, "", nil)
	}
	return writeValue(ctx, appbootconfig.MutationView(result))
}
