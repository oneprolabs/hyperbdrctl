package commands

import (
	"strings"

	appbootconfigwizard "hyperbdr-client/internal/app/bootconfigwizard"

	"github.com/spf13/cobra"
)

func newHostCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "host", "cmd.host.short", "cmd.host.long", "cmd.host.examples", "cmd.host.notes", "host")
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
		newRawLeafCommand(ctx, "cleanup-validation-host", "cmd.host.cleanup.short", "cmd.host.cleanup.long", "cmd.host.cleanup.examples", "cmd.host.cleanup.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagString(cmd, ctx, "ids")
			addFlagString(cmd, ctx, "file")
		}, func(args []string) error {
			return runHosts(ctx, append([]string{"cleanup-validation-host"}, args...))
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
		newHostBootConfigCommand(ctx),
	)
	return cmd
}

func newHostBootConfigCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "boot-config", "cmd.host.boot_config.short", "cmd.host.boot_config.long", "cmd.host.boot_config.examples", "cmd.host.boot_config.notes", "boot-config")
	cmd.AddCommand(
		newRawLeafCommand(ctx, "create", "cmd.boot_config.create.short", "cmd.boot_config.create.long", "cmd.boot_config.create.examples", "cmd.boot_config.create.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagString(cmd, ctx, "file")
		}, func(args []string) error {
			return runBootConfig(ctx, append([]string{"create"}, args...))
		}),
		newRawLeafCommand(ctx, "get", "cmd.boot_config.get.short", "cmd.boot_config.get.long", "cmd.boot_config.get.examples", "cmd.boot_config.get.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
		}, func(args []string) error {
			return runBootConfig(ctx, append([]string{"get"}, args...))
		}),
		newRawLeafCommand(ctx, "update", "cmd.boot_config.update.short", "cmd.boot_config.update.long", "cmd.boot_config.update.examples", "cmd.boot_config.update.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagString(cmd, ctx, "file")
		}, func(args []string) error {
			return runBootConfig(ctx, append([]string{"update"}, args...))
		}),
		newRawLeafCommand(ctx, "apply", "cmd.boot_config.apply.short", "cmd.boot_config.apply.long", "cmd.boot_config.apply.examples", "cmd.boot_config.apply.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
			addFlagString(cmd, ctx, "file")
		}, func(args []string) error {
			return runBootConfigCLI(ctx, append([]string{"apply"}, args...))
		}),
	)
	return cmd
}

func newBootConfigAliasCommand(ctx *context) *cobra.Command {
	cmd := newHostBootConfigCommand(ctx)
	cmd.Use = "boot-config"
	cmd.Short = ctx.loc.T("cmd.boot_config.alias.short")
	cmd.Long = ctx.loc.T("cmd.boot_config.alias.long")
	cmd.Example = strings.TrimSpace(ctx.loc.T("cmd.boot_config.alias.examples"))
	cmd.Deprecated = ctx.loc.T("cmd.boot_config.alias.deprecated")
	cmd.Hidden = true
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

func newBootConfigWizardCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "boot-config-wizard", "cmd.boot_config_wizard.short", "cmd.boot_config_wizard.long", "cmd.boot_config_wizard.examples", "cmd.boot_config_wizard.notes", "boot-config-wizard")
	cmd.AddCommand(
		newRawLeafCommand(ctx, "storages", "cmd.boot_config_wizard.storages.short", "cmd.boot_config_wizard.storages.long", "cmd.boot_config_wizard.storages.examples", "cmd.boot_config_wizard.storages.notes", func(cmd *cobra.Command) {
			addFlagInt(cmd, ctx, "page")
			addFlagInt(cmd, ctx, "page-size")
			addFlagString(cmd, ctx, "type")
			addFlagString(cmd, ctx, "status")
		}, func(args []string) error {
			return runBootConfigWizard(ctx, append([]string{"storages"}, args...))
		}),
		newRawLeafCommand(ctx, "storage-detail", "cmd.boot_config_wizard.storage_detail.short", "cmd.boot_config_wizard.storage_detail.long", "cmd.boot_config_wizard.storage_detail.examples", "cmd.boot_config_wizard.storage_detail.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "storage-id")
		}, func(args []string) error {
			return runBootConfigWizard(ctx, append([]string{"storage-detail"}, args...))
		}),
		newRawLeafCommand(ctx, "target-accounts", "cmd.boot_config_wizard.target_accounts.short", "cmd.boot_config_wizard.target_accounts.long", "cmd.boot_config_wizard.target_accounts.examples", "cmd.boot_config_wizard.target_accounts.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "status")
			addFlagString(cmd, ctx, "cloud-type")
			addFlagString(cmd, ctx, "storage-type")
		}, func(args []string) error {
			return runBootConfigWizard(ctx, append([]string{"target-accounts"}, args...))
		}),
		newRawLeafCommand(ctx, "target-auth-info", "cmd.boot_config_wizard.target_auth_info.short", "cmd.boot_config_wizard.target_auth_info.long", "cmd.boot_config_wizard.target_auth_info.examples", "cmd.boot_config_wizard.target_auth_info.notes", func(cmd *cobra.Command) {
			for _, name := range []string{"cloud-account-id", "cloud-account", "cloud-type", "storage-type", "fetch-res", "host-id", "storage-id", "network-addr-for-write-data", "network-addr-for-read-data", "region-id", "zone-id", "cloud-account-username", "cloud-account-use-public", "flavor-id", "boot-loader-flavor-id", "arch", "os-type-id", "os-type", "flavors", "flavor-vcpus", "flavor-ram", "max-nic-num", "system-volume-type-id", "volume-type-id", "default-volume-type-id", "default-pool-id", "dest-boot-mode", "network-id"} {
				addFlagString(cmd, ctx, name)
			}
		}, func(args []string) error {
			return runBootConfigWizard(ctx, append([]string{"target-auth-info"}, args...))
		}),
		newRawLeafCommand(ctx, "subnet-config", "cmd.boot_config_wizard.subnet_config.short", "cmd.boot_config_wizard.subnet_config.long", "cmd.boot_config_wizard.subnet_config.examples", "cmd.boot_config_wizard.subnet_config.notes", func(cmd *cobra.Command) {
			for _, name := range []string{"cloud-account-id", "cloud-type", "region-id", "zone-id", "network-id", "subnet-id"} {
				addFlagString(cmd, ctx, name)
			}
		}, func(args []string) error {
			return runBootConfigWizard(ctx, append([]string{"subnet-config"}, args...))
		}),
		newRawLeafCommand(ctx, "host-profile", "cmd.boot_config_wizard.host_profile.short", "cmd.boot_config_wizard.host_profile.long", "cmd.boot_config_wizard.host_profile.examples", "cmd.boot_config_wizard.host_profile.notes", func(cmd *cobra.Command) {
			addFlagString(cmd, ctx, "id")
		}, func(args []string) error {
			return runBootConfigWizard(ctx, append([]string{"host-profile"}, args...))
		}),
		newRawLeafCommand(ctx, "strategies", "cmd.boot_config_wizard.strategies.short", "cmd.boot_config_wizard.strategies.long", "cmd.boot_config_wizard.strategies.examples", "cmd.boot_config_wizard.strategies.notes", func(cmd *cobra.Command) {
			addFlagInt(cmd, ctx, "page")
			addFlagInt(cmd, ctx, "page-size")
			addFlagString(cmd, ctx, "kw")
			addFlagString(cmd, ctx, "status")
		}, func(args []string) error {
			return runBootConfigWizard(ctx, append([]string{"strategies"}, args...))
		}),
	)
	targetPlatforms := newRawLeafCommand(ctx, "target-platforms", "cmd.boot_config_wizard.target_platforms.short", "cmd.boot_config_wizard.target_platforms.long", "cmd.boot_config_wizard.target_platforms.examples", "cmd.boot_config_wizard.target_platforms.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "status")
		addFlagString(cmd, ctx, "cloud-type")
		addFlagString(cmd, ctx, "storage-type")
		addFlagString(cmd, ctx, "target-type")
	}, func(args []string) error {
		return runBootConfigWizard(ctx, append([]string{"target-platforms"}, args...))
	})
	targetPlatforms.Deprecated = ctx.loc.T("cmd.boot_config_wizard.target_platforms.deprecated")
	targetPlatforms.Hidden = true
	cmd.AddCommand(targetPlatforms)
	return cmd
}

func runBootConfigWizardTargetAccounts(ctx *context, args []string, legacyTargetPlatforms bool) error {
	commandName := "boot-config-wizard target-accounts"
	storageTypeDefault := ""
	if legacyTargetPlatforms {
		commandName = "boot-config-wizard target-platforms"
		storageTypeDefault = "objectstorage"
	}
	fs := newFlagSet(commandName)
	if legacyTargetPlatforms {
		fs.String("target-type", "recovery", "")
	}
	page := fs.Int("page", 1, "")
	pageSize := fs.Int("page-size", 100, "")
	storageType := fs.String("storage-type", storageTypeDefault, "")
	cloudType := fs.String("cloud-type", "", "")
	status := fs.String("status", "available", "")
	taskStatus := fs.String("task-status", "", "")
	rtExtra := fs.String("rt-extra", "", "")
	extra := fs.String("extra", "", "")
	rtTree := fs.String("rt-tree", "", "")
	q := queryFromPairs()
	if err := parseQueryFlagsInto(fs, args, q); err != nil {
		return err
	}
	service := appbootconfigwizard.NewService(commandAPIAdapter{ctx: ctx})
	resp, err := service.TargetAccounts(appbootconfigwizard.TargetAccountsSpec{
		StorageType:        *storageType,
		StorageTypeDefault: storageTypeDefault,
		CloudType:          *cloudType,
		Status:             *status,
		TaskStatus:         *taskStatus,
		RTExtra:            *rtExtra,
		Extra:              *extra,
		RTTree:             *rtTree,
		Page:               *page,
		PageSize:           *pageSize,
		Query:              q,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "cloud_accounts", cloudAccountColumns())
}
