package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const (
	notesAnnotation               = "notes"
	workflowAnnotation            = "workflow"
	quickStartAnnotation          = "quick-start"
	automaticBehaviorAnnotation   = "automatic-behavior"
	minimumFlagsAnnotation        = "minimum-flags"
	commonOptionalFlagsAnnotation = "common-optional-flags"
	relatedCommandsAnnotation     = "related-commands"
	nextStepsAnnotation           = "next-steps"
	flagsTitleAnnotation          = "flags-title"
	usageNotesAnnotation          = "usage-notes"
	usageLineAnnotation           = "usage-line"
	helpDescriptionAnnotation     = "help-description"
	helpLayoutAnnotation          = "help-layout"

	helpLayoutLegacy      = "legacy"
	helpLayoutFourSection = "four-section"
	helpLayoutGroup       = "group"
	helpLayoutRoot        = "root"
)

func renderHelp(cmd *cobra.Command, ctx *context) error {
	switch helpLayout(cmd) {
	case helpLayoutRoot:
		return renderRootHelp(cmd, ctx)
	case helpLayoutGroup:
		return renderGroupHelp(cmd, ctx)
	case helpLayoutFourSection:
		return renderFourSectionHelp(cmd, ctx)
	default:
		return renderLegacyHelp(cmd, ctx)
	}
}

func renderLegacyHelp(cmd *cobra.Command, ctx *context) error {
	fmt.Fprintf(ctx.out, "%s: %s\n", ctx.loc.T("help.section_usage"), cmd.UseLine())
	if cmd.Deprecated != "" {
		fmt.Fprintf(ctx.out, "\n%s: %s\n", ctx.loc.T("help.section_deprecated"), cmd.Deprecated)
	}
	if strings.TrimSpace(cmd.Short) != "" {
		fmt.Fprintf(ctx.out, "\n%s\n", strings.TrimSpace(cmd.Short))
	}
	if long := strings.TrimSpace(cmd.Long); long != "" && long != strings.TrimSpace(cmd.Short) {
		fmt.Fprintf(ctx.out, "\n%s\n", long)
	}
	if aliases := cmd.Aliases; len(aliases) > 0 {
		fmt.Fprintf(ctx.out, "\n%s: %s\n", ctx.loc.T("help.section_aliases"), strings.Join(aliases, ", "))
	}
	if quickStart := strings.TrimSpace(cmd.Annotations[quickStartAnnotation]); quickStart != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_quick_start"), quickStart)
	}
	if minimumFlags := strings.TrimSpace(cmd.Annotations[minimumFlagsAnnotation]); minimumFlags != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_minimum_flags"), minimumFlags)
	}
	if automaticBehavior := strings.TrimSpace(cmd.Annotations[automaticBehaviorAnnotation]); automaticBehavior != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_automatic_behavior"), automaticBehavior)
	}
	if examples := strings.TrimSpace(cmd.Example); examples != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_examples"), examples)
	}
	if notes := strings.TrimSpace(cmd.Annotations[notesAnnotation]); notes != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_notes"), notes)
	}
	if workflow := strings.TrimSpace(cmd.Annotations[workflowAnnotation]); workflow != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_workflow"), workflow)
	}
	if commonOptionalFlags := strings.TrimSpace(cmd.Annotations[commonOptionalFlagsAnnotation]); commonOptionalFlags != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_common_optional_flags"), commonOptionalFlags)
	}
	if subcommands := visibleSubcommands(cmd); len(subcommands) > 0 {
		fmt.Fprintf(ctx.out, "\n%s:\n", ctx.loc.T("help.section_commands"))
		for _, sub := range subcommands {
			fmt.Fprintf(ctx.out, "  %-24s %s\n", sub.Name(), sub.Short)
		}
	}
	if relatedCommands := strings.TrimSpace(cmd.Annotations[relatedCommandsAnnotation]); relatedCommands != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_related_commands"), relatedCommands)
	}
	if nextSteps := strings.TrimSpace(cmd.Annotations[nextStepsAnnotation]); nextSteps != "" {
		renderHelpSection(ctx, ctx.loc.T("help.section_next_steps"), nextSteps)
	}
	if flags := strings.TrimRight(cmd.NonInheritedFlags().FlagUsagesWrapped(120), "\n"); strings.TrimSpace(flags) != "" {
		fmt.Fprintf(ctx.out, "\n%s:\n%s", helpSectionTitle(cmd, ctx, flagsTitleAnnotation, "help.section_flags"), flags)
	}
	globalFlags := strings.TrimRight(cmd.InheritedFlags().FlagUsagesWrapped(120), "\n")
	if cmd == cmd.Root() {
		globalFlags = strings.TrimRight(cmd.PersistentFlags().FlagUsagesWrapped(120), "\n")
	}
	if strings.TrimSpace(globalFlags) != "" {
		fmt.Fprintf(ctx.out, "\n%s:\n%s", ctx.loc.T("help.section_global_flags"), globalFlags)
	}
	fmt.Fprintln(ctx.out)
	return nil
}

func helpLayout(cmd *cobra.Command) string {
	if layout := strings.TrimSpace(cmd.Annotations[helpLayoutAnnotation]); layout != "" {
		return layout
	}
	return helpLayoutLegacy
}

func helpUseLine(cmd *cobra.Command) string {
	if usageLine := strings.TrimSpace(cmd.Annotations[usageLineAnnotation]); usageLine != "" {
		return usageLine
	}
	return cmd.UseLine()
}

func helpUsageNotes(cmd *cobra.Command) string {
	return strings.TrimSpace(cmd.Annotations[usageNotesAnnotation])
}

func renderHelpSection(ctx *context, title, body string) {
	fmt.Fprintf(ctx.out, "\n%s:\n%s\n", title, indentHelpBody(body))
}

func indentHelpBody(body string) string {
	lines := strings.Split(strings.Trim(body, "\n"), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			lines[i] = ""
			continue
		}
		lines[i] = "  " + line
	}
	return strings.Join(lines, "\n")
}

func helpSectionTitle(cmd *cobra.Command, ctx *context, annotation, defaultKey string) string {
	if title := strings.TrimSpace(cmd.Annotations[annotation]); title != "" {
		return title
	}
	return ctx.loc.T(defaultKey)
}

func visibleSubcommands(cmd *cobra.Command) []*cobra.Command {
	children := make([]*cobra.Command, 0)
	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" {
			continue
		}
		children = append(children, child)
	}
	return children
}

func rawArgsHelp(_ *cobra.Command, args []string) bool {
	if len(args) == 1 && args[0] == "--" {
		return false
	}
	return hasHelpToken(args)
}

func addLocalizedAnnotation(cmd *cobra.Command, ctx *context, annotation, key string) {
	text := strings.TrimSpace(ctx.loc.T(key))
	if text == "" || text == key {
		return
	}
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[annotation] = text
}

func addWorkflow(cmd *cobra.Command, ctx *context, key string) {
	addLocalizedAnnotation(cmd, ctx, workflowAnnotation, key)
}

func addQuickStart(cmd *cobra.Command, ctx *context, key string) {
	addLocalizedAnnotation(cmd, ctx, quickStartAnnotation, key)
}

func addAutomaticBehavior(cmd *cobra.Command, ctx *context, key string) {
	addLocalizedAnnotation(cmd, ctx, automaticBehaviorAnnotation, key)
}

func addMinimumFlags(cmd *cobra.Command, ctx *context, key string) {
	addLocalizedAnnotation(cmd, ctx, minimumFlagsAnnotation, key)
}

func addCommonOptionalFlags(cmd *cobra.Command, ctx *context, key string) {
	addLocalizedAnnotation(cmd, ctx, commonOptionalFlagsAnnotation, key)
}

func addRelatedCommands(cmd *cobra.Command, ctx *context, key string) {
	addLocalizedAnnotation(cmd, ctx, relatedCommandsAnnotation, key)
}

func addNextSteps(cmd *cobra.Command, ctx *context, key string) {
	addLocalizedAnnotation(cmd, ctx, nextStepsAnnotation, key)
}

func addFlagsTitle(cmd *cobra.Command, ctx *context, key string) {
	addLocalizedAnnotation(cmd, ctx, flagsTitleAnnotation, key)
}

func addUsageNotes(cmd *cobra.Command, ctx *context, key string) {
	addLocalizedAnnotation(cmd, ctx, usageNotesAnnotation, key)
}

func addUsageLine(cmd *cobra.Command, ctx *context, key string) {
	addLocalizedAnnotation(cmd, ctx, usageLineAnnotation, key)
}

func addHelpDescription(cmd *cobra.Command, ctx *context, key string) {
	addLocalizedAnnotation(cmd, ctx, helpDescriptionAnnotation, key)
}

func addHelpLayout(cmd *cobra.Command, layout string) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[helpLayoutAnnotation] = layout
}
