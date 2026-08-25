package commands

import (
	"fmt"
	"strings"

	"hyperbdr-client/internal/config"

	"github.com/spf13/cobra"
)

const productionSiteCreateHelpProfileAnnotation = "production-site-create-help-profile"

type productionSiteCreateSelection struct {
	Type string
}

func newProductionSiteCreateCommand(ctx *context) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "create",
		Short:              ctx.loc.T("cmd.production_site.create.short"),
		Long:               ctx.loc.T("cmd.production_site.create.long"),
		Example:            strings.TrimSpace(ctx.loc.T("cmd.production_site.create.examples")),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rawArgsHelp(cmd, args) {
				selection, err := parseProductionSiteCreateSelection(args)
				if err != nil {
					return err
				}
				return renderProductionSiteCreateHelp(ctx, cmd, selection)
			}
			return runSources(ctx, append([]string{"create"}, args...))
		},
	}
	cmd.Flags().String("type", "", ctx.loc.T("flag.production-site-type"))
	for _, name := range []string{"synch-node-id", "synch-node-ids", "auth-url", "auth-key", "auth-cert", "region-id"} {
		addFlagString(cmd, ctx, name)
	}
	addFlagBool(cmd, ctx, "preview-request")
	addNotes(cmd, ctx, "cmd.production_site.create.notes")
	addHelpLayout(cmd, helpLayoutFourSection)
	addHelpDescription(cmd, ctx, "cmd.production_site.create.help_title")
	addUsageLine(cmd, ctx, "cmd.production_site.create.usage_line")
	addUsageNotes(cmd, ctx, "cmd.production_site.create.usage_notes")
	return cmd
}

func parseProductionSiteCreateSelection(args []string) (productionSiteCreateSelection, error) {
	selection := productionSiteCreateSelection{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--help" || arg == "-h" || !strings.HasPrefix(arg, "--") {
			continue
		}
		name, value, hasInline := splitFlag(arg)
		if strings.TrimPrefix(name, "--") != "type" {
			continue
		}
		resolved, next, err := strictFlagValue(args, i, value, hasInline)
		if err != nil {
			return selection, err
		}
		selection.Type = strings.ToLower(strings.TrimSpace(resolved))
		i = next
	}
	return selection, nil
}

func renderProductionSiteCreateHelp(ctx *context, cmd *cobra.Command, selection productionSiteCreateSelection) error {
	profile := "selector"
	if selection.Type != "" {
		if selection.Type != "vmware" && selection.Type != "aws" {
			return fmt.Errorf("type must be one of vmware, aws")
		}
		profile = selection.Type
		addAnnotationValue(cmd, usageLineAnnotation, ctx.loc.T("cmd.production_site.create."+profile+".usage_line"))
		addAnnotationValue(cmd, usageNotesAnnotation, ctx.loc.T("cmd.production_site.create."+profile+".usage_notes"))
		addHelpDescription(cmd, ctx, "cmd.production_site.create."+profile+".help_title")
		applyProductionSiteCreateFlagHelp(ctx, cmd, profile)
	}
	addAnnotationValue(cmd, productionSiteCreateHelpProfileAnnotation, profile)
	return renderHelp(cmd, ctx)
}

func applyProductionSiteCreateFlagHelp(ctx *context, cmd *cobra.Command, profile string) {
	for _, name := range []string{"auth-url", "auth-key", "auth-cert"} {
		overrideFlagUsage(cmd, ctx, name, "flag.production_site.create."+profile+"."+name)
	}
}

func productionSiteCreateFlagSpecsForProfile(profile string) []flagHelpSpec {
	global := []flagHelpSpec{
		{name: "debug"},
		{name: "lang"},
		{name: "output", choices: []string{"table", "json"}, defaultValue: config.DefaultOutput},
		{name: "help"},
	}
	if profile == "selector" || profile == "" {
		return append([]flagHelpSpec{{name: "type", choices: []string{"vmware", "aws"}}}, global...)
	}
	specs := []flagHelpSpec{
		{name: "synch-node-id"},
		{name: "synch-node-ids"},
		{name: "auth-url", required: true},
		{name: "auth-key", required: true},
		{name: "auth-cert", required: true},
	}
	if profile == "aws" {
		specs = append(specs, flagHelpSpec{name: "region-id", required: true})
	}
	specs = append(specs, flagHelpSpec{name: "preview-request"})
	return append(specs, global...)
}
