package commands

import (
	"fmt"
	"strconv"
	"strings"

	"hyperbdr-client/catalog"
	appbootconfigapply "hyperbdr-client/internal/app/bootconfigapply"
	apptargetresource "hyperbdr-client/internal/app/targetresource"
	"hyperbdr-client/internal/config"
	"hyperbdr-client/internal/output"

	"github.com/spf13/cobra"
)

const bootConfigApplyHelpProfileAnnotation = "boot-config-apply-help-profile"

type bootConfigApplyHelpSelection struct {
	CloudAccountID string
}

type bootConfigApplyHelpProfile struct {
	Entry    catalog.CloudEntry
	Provider string
	Kind     string
}

func newTopLevelBootConfigCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "boot-config", "cmd.boot_config_top.short", "cmd.boot_config_top.long", "cmd.boot_config_top.examples", "cmd.boot_config_top.notes", "boot-config")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.boot_config_top.usage_line")
	addUsageNotes(cmd, ctx, "cmd.boot_config_top.usage_notes")

	getCmd := newRawLeafCommand(ctx, "get", "cmd.boot_config.get.short", "cmd.boot_config.get.long", "cmd.boot_config.get.examples", "cmd.boot_config.get.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
	}, func(args []string) error {
		return runBootConfig(ctx, append([]string{"get"}, args...))
	})
	addHelpLayout(getCmd, helpLayoutFourSection)
	addUsageLine(getCmd, ctx, "cmd.boot_config_top.get.usage_line")
	addUsageNotes(getCmd, ctx, "cmd.boot_config_top.get.usage_notes")

	applyCmd := &cobra.Command{
		Use:                "apply",
		Short:              ctx.loc.T("cmd.boot_config_top.apply.short"),
		Long:               ctx.loc.T("cmd.boot_config_top.apply.long"),
		Example:            strings.TrimSpace(ctx.loc.T("cmd.boot_config_top.apply.examples")),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			selection, err := parseBootConfigApplyHelpSelection(args)
			if err != nil {
				return err
			}
			if rawArgsHelp(cmd, args) {
				return renderBootConfigApplyHelp(ctx, cmd, selection)
			}
			return runTopLevelBootConfig(ctx, append([]string{"apply"}, args...))
		},
	}
	addNotes(applyCmd, ctx, "cmd.boot_config_top.apply.notes")
	addFlagString(applyCmd, ctx, "id")
	for _, name := range []string{
		"cloud-account-id",
		"storage-id",
		"region-id",
		"zone-id",
		"project-id",
		"project-domain-id",
		"compute-zone-id",
		"image-id",
		"flavor-id",
		"system-volume-type-id",
		"volume-type-id",
		"network-id",
		"subnet-id",
		"security-group-id",
		"boot-loader-image-id",
		"boot-loader-flavor-id",
	} {
		addFlagString(applyCmd, ctx, name)
	}
	addFlagString(applyCmd, ctx, "file")
	applyCmd.Flags().StringArray("set", nil, ctx.loc.T("flag.set"))
	applyCmd.Flags().StringArray("set-json", nil, ctx.loc.T("flag.set-json"))
	applyCmd.Flags().Bool("preview-request", false, ctx.loc.T("flag.preview-request"))
	addHelpLayout(applyCmd, helpLayoutFourSection)
	addUsageLine(applyCmd, ctx, "cmd.boot_config_top.apply.usage_line")
	addUsageNotes(applyCmd, ctx, "cmd.boot_config_top.apply.usage_notes")

	cmd.AddCommand(
		getCmd,
		applyCmd,
	)
	return cmd
}

func runTopLevelBootConfig(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("boot-config", "")
	}
	switch args[0] {
	case "apply":
		input, err := parseTopLevelBootConfigApplyArgs(args[1:])
		if err != nil {
			return err
		}
		if input.HostID == "" {
			return missing(ctx, "error.missing_id")
		}
		service := appbootconfigapply.NewService(commandAPIAdapter{ctx: ctx})
		if input.PreviewRequest {
			prepared, err := service.PrepareRequest(input)
			if err != nil {
				return err
			}
			return output.JSON(ctx.out, prepared.Body)
		}
		result, err := service.Apply(input)
		if err != nil {
			return err
		}
		if ctx.cfg.Output == "json" {
			if result.NoOp {
				return writeValue(ctx, appbootconfigapply.MutationView(result))
			}
			return writeResponse(ctx, result.Response, "", nil)
		}
		return writeValue(ctx, appbootconfigapply.MutationView(result))
	default:
		return errUnknown("boot-config", args[0])
	}
}

func parseBootConfigApplyHelpSelection(args []string) (bootConfigApplyHelpSelection, error) {
	selection := bootConfigApplyHelpSelection{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--help" || arg == "-h" {
			continue
		}
		if !strings.HasPrefix(arg, "--") {
			return selection, fmt.Errorf("unexpected argument %q", arg)
		}
		name, value, hasInline := splitFlag(arg)
		switch name {
		case "--cloud-account-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return selection, err
			}
			selection.CloudAccountID = v
			i = next
		default:
			if flagConsumesValue(name) && !hasInline && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				i++
			}
		}
	}
	return selection, nil
}

func renderBootConfigApplyHelp(ctx *context, cmd *cobra.Command, selection bootConfigApplyHelpSelection) error {
	profile := "generic"
	if strings.TrimSpace(selection.CloudAccountID) != "" {
		cfg, err := config.Resolve(ctx.flags)
		if err != nil {
			return err
		}
		ctx.cfg = cfg
		accountCtx, err := apptargetresource.NewService(commandAPIAdapter{ctx: ctx}).CloudAccountContext(selection.CloudAccountID)
		if err != nil {
			return err
		}
		resolved, err := resolveBootConfigApplyAccountProfile(accountCtx)
		if err != nil {
			return err
		}
		profile = resolved.Kind
		addAnnotationValue(cmd, usageLineAnnotation, ctx.loc.T("cmd.boot_config_top.apply.account.usage_line"))
		addAnnotationValue(cmd, helpDescriptionAnnotation, ctx.loc.T("cmd.boot_config_top.apply.account.short"))
		addAnnotationValue(cmd, usageNotesAnnotation, bootConfigApplyAccountUsageNotes(ctx, accountCtx, resolved))
	}
	addAnnotationValue(cmd, bootConfigApplyHelpProfileAnnotation, profile)
	return renderHelp(cmd, ctx)
}

func resolveBootConfigApplyAccountProfile(accountCtx apptargetresource.CloudAccountContext) (bootConfigApplyHelpProfile, error) {
	if strings.TrimSpace(accountCtx.CloudType) == "" {
		return bootConfigApplyHelpProfile{}, fmt.Errorf("cloud-type cannot be inferred from cloud-account-id")
	}
	kind, err := bootConfigApplyStorageKind(accountCtx.StorageType)
	if err != nil {
		return bootConfigApplyHelpProfile{}, err
	}

	var (
		entry catalog.CloudEntry
		ok    bool
	)
	if kind == "block" {
		entry, ok = catalog.FindBlockCloud(accountCtx.CloudType)
	} else {
		entry, ok = catalog.FindObjectCloud(accountCtx.CloudType)
	}
	if !ok {
		return bootConfigApplyHelpProfile{}, fmt.Errorf("cloud-type %q does not support boot-config apply help", accountCtx.CloudType)
	}

	profile := "account|" + kind
	if entry.Provider == "openstack" {
		profile = "account|openstack|" + kind
	}
	return bootConfigApplyHelpProfile{
		Entry:    entry,
		Provider: entry.Provider,
		Kind:     profile,
	}, nil
}

func bootConfigApplyStorageKind(storageType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(storageType)) {
	case "hypergate", "blockstorage", "block", "block_storage":
		return "block", nil
	case "objectstorage", "object", "object_storage":
		return "object", nil
	case "":
		return "", fmt.Errorf("storage-type cannot be inferred from cloud-account-id")
	default:
		return "", fmt.Errorf("storage-type %q does not support boot-config apply help", storageType)
	}
}

func bootConfigApplyAccountUsageNotes(ctx *context, accountCtx apptargetresource.CloudAccountContext, profile bootConfigApplyHelpProfile) string {
	storageKind := "object"
	if strings.HasSuffix(profile.Kind, "|block") {
		storageKind = "block"
	}
	if profile.Provider == "aliyun" || profile.Provider == "openstack" || (profile.Provider == "huawei" && storageKind == "object") {
		return ctx.loc.T(fmt.Sprintf("help.boot_config.apply.%s.%s", profile.Provider, storageKind))
	}
	switch profile.Kind {
	case "account|openstack|block":
		return fmt.Sprintf(ctx.loc.T("cmd.boot_config_top.apply.account.openstack.block.usage_notes"), accountCtx.CloudAccountID, accountCtx.CloudType, accountCtx.StorageType, profile.Provider)
	case "account|openstack|object":
		return fmt.Sprintf(ctx.loc.T("cmd.boot_config_top.apply.account.openstack.object.usage_notes"), accountCtx.CloudAccountID, accountCtx.CloudType, accountCtx.StorageType, profile.Provider)
	case "account|block":
		return fmt.Sprintf(ctx.loc.T("cmd.boot_config_top.apply.account.block.usage_notes"), accountCtx.CloudAccountID, accountCtx.CloudType, accountCtx.StorageType, profile.Provider)
	default:
		return fmt.Sprintf(ctx.loc.T("cmd.boot_config_top.apply.account.object.usage_notes"), accountCtx.CloudAccountID, accountCtx.CloudType, accountCtx.StorageType, profile.Provider)
	}
}

func parseTopLevelBootConfigApplyArgs(args []string) (appbootconfigapply.ApplyInput, error) {
	input := appbootconfigapply.ApplyInput{Dynamic: map[string]string{}}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return input, fmt.Errorf("unexpected argument %q", arg)
		}
		name, value, hasInline := splitFlag(arg)
		name = strings.TrimPrefix(name, "--")
		switch name {
		case "preview-request":
			if hasInline {
				enabled, err := strconv.ParseBool(value)
				if err != nil {
					return input, fmt.Errorf("%s requires boolean value", arg)
				}
				input.PreviewRequest = enabled
				continue
			}
			input.PreviewRequest = true
		case "id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return input, err
			}
			input.HostID = v
			i = next
		case "file":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return input, err
			}
			input.File = v
			i = next
		case "set":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return input, err
			}
			input.Sets = append(input.Sets, v)
			i = next
		case "set-json":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return input, err
			}
			input.SetJSONs = append(input.SetJSONs, v)
			i = next
		case "help":
			return input, fmt.Errorf("unexpected argument %q", arg)
		default:
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return input, err
			}
			input.Dynamic[strings.ReplaceAll(name, "-", "_")] = v
			i = next
		}
	}
	return input, nil
}

func strictFlagValue(args []string, idx int, inline string, hasInline bool) (string, int, error) {
	if hasInline {
		return inline, idx, nil
	}
	if idx+1 >= len(args) {
		return "", idx, fmt.Errorf("%s requires value", args[idx])
	}
	if strings.HasPrefix(args[idx+1], "--") {
		return "", idx, fmt.Errorf("%s requires value", args[idx])
	}
	return args[idx+1], idx + 1, nil
}
