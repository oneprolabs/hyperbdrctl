package commands

import (
	"strings"

	"github.com/spf13/cobra"
)

func newHostCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "host", "cmd.host.short", "cmd.host.long", "cmd.host.examples", "cmd.host.notes", "host")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.host.usage_line")
	addUsageNotes(cmd, ctx, "cmd.host.usage_notes")
	cmd.AddCommand(
		newRawLeafCommand(ctx, "list", "cmd.host.list.short", "cmd.host.list.long", "cmd.host.list.examples", "cmd.host.list.notes", func(cmd *cobra.Command) {
			addFlagInt(cmd, ctx, "page")
			addFlagInt(cmd, ctx, "page-size")
			addFlagString(cmd, ctx, "status")
			addFlagString(cmd, ctx, "boot-status")
			addFlagString(cmd, ctx, "kw")
			addFlagString(cmd, ctx, "cloud-type")
			addFlagString(cmd, ctx, "ids")
			addFlagString(cmd, ctx, "macs")
		}, func(args []string) error {
			return runHosts(ctx, append([]string{"list"}, args...))
		}),
		newRawLeafCommand(ctx, "detail", "cmd.host.detail.short", "cmd.host.detail.long", "cmd.host.detail.examples", "cmd.host.detail.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
		}, func(args []string) error {
			return runHosts(ctx, append([]string{"detail"}, args...))
		}),
		newRawLeafCommand(ctx, "snapshots", "cmd.host.snapshots.short", "cmd.host.snapshots.long", "cmd.host.snapshots.examples", "cmd.host.snapshots.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagString(cmd, ctx, "status")
			addFlagBool(cmd, ctx, "sync-detail")
		}, func(args []string) error {
			return runHosts(ctx, append([]string{"snapshots"}, args...))
		}),
		newRawLeafCommand(ctx, "sync", "cmd.host.sync.short", "cmd.host.sync.long", "cmd.host.sync.examples", "cmd.host.sync.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagString(cmd, ctx, "ids")
			addFlagString(cmd, ctx, "mode")
			addFlagInt(cmd, ctx, "transfer-speed")
			addFlagString(cmd, ctx, "file")
		}, func(args []string) error {
			return runHosts(ctx, append([]string{"sync"}, args...))
		}),
		newRawLeafCommand(ctx, "register", "cmd.host.register.short", "cmd.host.register.long", "cmd.host.register.examples", "cmd.host.register.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "vm-id")
			addFlagString(cmd, ctx, "vm-ids")
		}, func(args []string) error {
			return runHosts(ctx, append([]string{"register"}, args...))
		}),
		newRawLeafCommand(ctx, "boot", "cmd.host.boot.short", "cmd.host.boot.long", "cmd.host.boot.examples", "cmd.host.boot.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagString(cmd, ctx, "file")
			addFlagString(cmd, ctx, "snapshot-id")
			addFlagString(cmd, ctx, "boot-instance-purpose")
			addFlagString(cmd, ctx, "cloud-type")
		}, func(args []string) error {
			return runHosts(ctx, append([]string{"boot"}, args...))
		}),
		newRawLeafCommand(ctx, "clean", "cmd.host.clean.short", "cmd.host.clean.long", "cmd.host.clean.examples", "cmd.host.clean.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagString(cmd, ctx, "ids")
			addFlagString(cmd, ctx, "file")
		}, func(args []string) error {
			return runHosts(ctx, append([]string{"clean"}, args...))
		}),
		newRawLeafCommand(ctx, "deregister", "cmd.host.deregister.short", "cmd.host.deregister.long", "cmd.host.deregister.examples", "cmd.host.deregister.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagString(cmd, ctx, "ids")
			addFlagBool(cmd, ctx, "force")
			addFlagString(cmd, ctx, "file")
		}, func(args []string) error {
			return runHosts(ctx, append([]string{"deregister"}, args...))
		}),
		newRawLeafCommand(ctx, "wait", "cmd.host.wait.short", "cmd.host.wait.long", "cmd.host.wait.examples", "cmd.host.wait.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagString(cmd, ctx, "ids")
			addFlagString(cmd, ctx, "operation")
			addFlagInt(cmd, ctx, "interval-seconds")
			addFlagInt(cmd, ctx, "timeout-seconds")
			addFlagBool(cmd, ctx, "include-steps")
		}, func(args []string) error {
			return runHosts(ctx, append([]string{"wait"}, args...))
		}),
	)
	for _, child := range cmd.Commands() {
		addHelpLayout(child, helpLayoutFourSection)
		addUsageLine(child, ctx, "cmd.host."+strings.ReplaceAll(child.Name(), "-", "_")+".usage_line")
		addUsageNotes(child, ctx, "cmd.host."+strings.ReplaceAll(child.Name(), "-", "_")+".usage_notes")
	}
	return cmd
}

func newBootConfigCLIAliasCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "boot-config-cli", "cmd.boot_config_cli.short", "cmd.boot_config_cli.long", "cmd.boot_config_cli.examples", "cmd.boot_config_cli.notes", "boot-config-cli")
	cmd.Deprecated = ctx.loc.T("cmd.boot_config_cli.deprecated")
	cmd.Hidden = true
	cmd.AddCommand(newRawLeafCommand(ctx, "apply", "cmd.boot_config.apply.short", "cmd.boot_config.apply.long", "cmd.boot_config.apply.examples", "cmd.boot_config.apply.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
		addFlagString(cmd, ctx, "file")
	}, func(args []string) error {
		return runBootConfigCLI(ctx, append([]string{"apply"}, args...))
	}))
	return cmd
}

func newBatchBootConfigCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "batch-boot-config", "cmd.batch_boot_config.short", "cmd.batch_boot_config.long", "cmd.batch_boot_config.examples", "cmd.batch_boot_config.notes", "batch-boot-config")
	cmd.Hidden = true
	cmd.AddCommand(
		newRawLeafCommand(ctx, "create", "cmd.batch_boot_config.create.short", "cmd.batch_boot_config.create.long", "cmd.batch_boot_config.create.examples", "cmd.batch_boot_config.create.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "file")
		}, func(args []string) error {
			return runBatchBootConfig(ctx, append([]string{"create"}, args...))
		}),
		newRawLeafCommand(ctx, "get", "cmd.batch_boot_config.get.short", "cmd.batch_boot_config.get.long", "cmd.batch_boot_config.get.examples", "cmd.batch_boot_config.get.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "ids")
		}, func(args []string) error {
			return runBatchBootConfig(ctx, append([]string{"get"}, args...))
		}),
		newRawLeafCommand(ctx, "update", "cmd.batch_boot_config.update.short", "cmd.batch_boot_config.update.long", "cmd.batch_boot_config.update.examples", "cmd.batch_boot_config.update.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "file")
		}, func(args []string) error {
			return runBatchBootConfig(ctx, append([]string{"update"}, args...))
		}),
	)
	return cmd
}
