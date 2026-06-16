package commands

import (
	"fmt"
	"strconv"
	"strings"

	appbootconfigapply "hyperbdr-client/internal/app/bootconfigapply"
	"hyperbdr-client/internal/output"

	"github.com/spf13/cobra"
)

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

	applyCmd := newRawLeafCommand(ctx, "apply", "cmd.boot_config_top.apply.short", "cmd.boot_config_top.apply.long", "cmd.boot_config_top.apply.examples", "cmd.boot_config_top.apply.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
		addFlagString(cmd, ctx, "file")
		cmd.Flags().StringArray("set", nil, ctx.loc.T("flag.set"))
		cmd.Flags().StringArray("set-json", nil, ctx.loc.T("flag.set-json"))
		cmd.Flags().Bool("preview-request", false, ctx.loc.T("flag.preview-request"))
	}, func(args []string) error {
		return runTopLevelBootConfig(ctx, append([]string{"apply"}, args...))
	})
	addAutomaticBehavior(applyCmd, ctx, "cmd.boot_config_top.apply.automatic_behavior")
	addMinimumFlags(applyCmd, ctx, "cmd.boot_config_top.apply.minimum_flags")
	addRelatedCommands(applyCmd, ctx, "cmd.boot_config_top.apply.related")
	cmd.AddCommand(
		getCmd,
		applyCmd,
		newBootConfigFetchBlockResourcesCommand(ctx),
		newBootConfigFetchOSSResourcesCommand(ctx),
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
