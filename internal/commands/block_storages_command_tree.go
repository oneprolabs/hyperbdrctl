package commands

import (
	"fmt"

	"hyperbdr-client/catalog"
	workflowcreate "hyperbdr-client/internal/workflow/blockstoragecreate"

	"github.com/spf13/cobra"
)

func newBlockStoragesCreateCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "create", "cmd.block_storages.create_group.short", "cmd.block_storages.create_group.long", "cmd.block_storages.create_group.examples", "cmd.block_storages.create_group.notes", "target cloud-sync-gateway create")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.target.cloud_sync_gateway.create.usage_line")
	addUsageNotes(cmd, ctx, "cmd.target.cloud_sync_gateway.create.usage_notes")

	for _, entry := range catalog.EnabledBlockClouds() {
		cmd.AddCommand(newBlockStorageCreateProviderCommand(ctx, entry))
	}
	return cmd
}

func newBlockStorageCreateProviderCommand(ctx *context, entry catalog.CloudEntry) *cobra.Command {
	if workflowcreate.HasRegisteredAdapter(entry.CloudType) {
		switch entry.Provider {
		case "aliyun":
			return newAliyunBlockStorageCreateCommand(ctx)
		case "openstack":
			return newOpenStackBlockStorageCreateCommand(ctx)
		}
	}
	return newGenericBlockStorageCreateCommand(ctx, entry)
}

func newAliyunBlockStorageCreateCommand(ctx *context) *cobra.Command {
	cmd := newRawLeafCommand(ctx, "aliyun", "cmd.block_storages.create.aliyun.short", "cmd.block_storages.create.aliyun.long", "cmd.block_storages.create.aliyun.examples", "cmd.block_storages.create.aliyun.notes", func(cmd *cobra.Command) {
		for _, name := range []string{"cloud-account-id", "region-id", "zone-id", "image-id", "flavor-id", "network-id", "subnet-id", "fixed-ip", "system-disk-type-id", "system-disk-size", "boot-loader-image-id", "volume-proxy-type", "hg-control-network", "control-nat-ip", "hg-data-network", "data-nat-ip", "bandwidth-size", "hd-control-network"} {
			addFlagString(cmd, ctx, name)
		}
		addFlagBool(cmd, ctx, "preview-request")
	}, func(args []string) error {
		return runCreateBlockStorageForProvider(ctx, "target cloud-sync-gateway create aliyun", "aliyun_bs", args)
	})
	addHelpLayout(cmd, helpLayoutFourSection)
	addAnnotationValue(cmd, usageLineAnnotation, fmt.Sprintf(ctx.loc.T("cmd.target.cloud_sync_gateway.create.provider.usage_line"), "aliyun"))
	addUsageNotes(cmd, ctx, "cmd.target.cloud_sync_gateway.create.aliyun.usage_notes")
	return cmd
}

func newOpenStackBlockStorageCreateCommand(ctx *context) *cobra.Command {
	cmd := newRawLeafCommand(ctx, "openstack", "cmd.block_storages.create.openstack.short", "cmd.block_storages.create.openstack.long", "cmd.block_storages.create.openstack.examples", "cmd.block_storages.create.openstack.notes", func(cmd *cobra.Command) {
		for _, name := range []string{"cloud-account-id", "project-id", "region-id", "compute-zone-id", "image-id", "flavor-id", "network-id", "subnet-id", "fixed-ip", "volume-type-id", "system-disk-size", "block-store-zone-id", "boot-loader-image-id", "boot-loader-flavor-id", "project-domain-id", "boot-types-id", "volume-proxy-type", "hg-control-network", "control-nat-ip", "hg-data-network", "data-nat-ip"} {
			addFlagString(cmd, ctx, name)
		}
		addFlagBool(cmd, ctx, "preview-request")
	}, func(args []string) error {
		return runCreateBlockStorageForProvider(ctx, "target cloud-sync-gateway create openstack", "openstack", args)
	})
	addHelpLayout(cmd, helpLayoutFourSection)
	addAnnotationValue(cmd, usageLineAnnotation, fmt.Sprintf(ctx.loc.T("cmd.target.cloud_sync_gateway.create.provider.usage_line"), "openstack"))
	addUsageNotes(cmd, ctx, "cmd.target.cloud_sync_gateway.create.openstack.usage_notes")
	return cmd
}

func newGenericBlockStorageCreateCommand(ctx *context, entry catalog.CloudEntry) *cobra.Command {
	displayName := localizedCloudEntryName(ctx, entry)
	cmd := &cobra.Command{
		Use:                entry.Provider,
		Short:              fmt.Sprintf(ctx.loc.T("cmd.block_storages.create.provider.short"), displayName),
		Long:               fmt.Sprintf(ctx.loc.T("cmd.block_storages.create.provider.long"), displayName, entry.CloudType),
		Example:            fmt.Sprintf(ctx.loc.T("cmd.block_storages.create.provider.example"), entry.Provider),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rawArgsHelp(cmd, args) {
				return renderHelp(cmd, ctx)
			}
			if err := rejectLegacyGlobalFlags(cmd, args); err != nil {
				return err
			}
			return runCreateBlockStorageForProvider(ctx, genericBlockStorageCreateCommandName(entry.Provider), entry.CloudType, args)
		},
	}
	addGenericBlockStorageCreateFlags(cmd, ctx)
	addHelpLayout(cmd, helpLayoutFourSection)
	addAnnotationValue(cmd, usageLineAnnotation, fmt.Sprintf(ctx.loc.T("cmd.target.cloud_sync_gateway.create.provider.usage_line"), entry.Provider))
	addAnnotationValue(cmd, usageNotesAnnotation, fmt.Sprintf(ctx.loc.T("cmd.target.cloud_sync_gateway.create.provider.usage_notes"), displayName, entry.Provider, entry.CloudType, entry.Provider, entry.Provider))
	return cmd
}

func addGenericBlockStorageCreateFlags(cmd *cobra.Command, ctx *context) {
	for _, name := range []string{"cloud-account-id", "project-id", "region-id", "zone-id", "compute-zone-id", "image-id", "flavor-id", "network-id", "subnet-id", "fixed-ip", "system-disk-type-id", "volume-type-id", "system-disk-size", "block-store-zone-id", "boot-loader-image-id", "boot-loader-flavor-id", "project-domain-id", "boot-types-id", "volume-proxy-type", "hg-control-network", "control-nat-ip", "hg-data-network", "data-nat-ip", "bandwidth-size", "hd-control-network"} {
		addFlagString(cmd, ctx, name)
	}
	addFlagBool(cmd, ctx, "preview-request")
}

func genericBlockStorageCreateCommandName(provider string) string {
	return "target cloud-sync-gateway create " + provider
}

func newTargetCloudSyncGatewayCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "cloud-sync-gateway", "cmd.target.cloud_sync_gateway.short", "cmd.target.cloud_sync_gateway.long", "cmd.target.cloud_sync_gateway.examples", "cmd.target.cloud_sync_gateway.notes", "target cloud-sync-gateway")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.target.cloud_sync_gateway.usage_line")
	addUsageNotes(cmd, ctx, "cmd.target.cloud_sync_gateway.usage_notes")

	listCmd := newRawLeafCommand(ctx, "list", "cmd.block_storages.list.short", "cmd.block_storages.list.long", "cmd.block_storages.list.examples", "cmd.block_storages.list.notes", func(cmd *cobra.Command) {
		addFlagInt(cmd, ctx, "page")
		addFlagInt(cmd, ctx, "page-size")
		addFlagString(cmd, ctx, "type")
		addFlagString(cmd, ctx, "cloud-account-id")
	}, func(args []string) error {
		return runBlockStorages(ctx, append([]string{"list"}, args...))
	})
	configureTargetCloudSyncGatewayLeafHelp(listCmd, ctx, "cmd.target.cloud_sync_gateway.list.usage_line", "cmd.target.cloud_sync_gateway.list.usage_notes")

	detailCmd := newRawLeafCommand(ctx, "detail", "cmd.block_storages.detail.short", "cmd.block_storages.detail.long", "cmd.block_storages.detail.examples", "cmd.block_storages.detail.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
	}, func(args []string) error {
		return runBlockStorages(ctx, append([]string{"detail"}, args...))
	})
	configureTargetCloudSyncGatewayLeafHelp(detailCmd, ctx, "cmd.target.cloud_sync_gateway.detail.usage_line", "cmd.target.cloud_sync_gateway.detail.usage_notes")

	waitCmd := newRawLeafCommand(ctx, "wait", "cmd.block_storages.wait.short", "cmd.block_storages.wait.long", "cmd.block_storages.wait.examples", "cmd.block_storages.wait.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
		addFlagInt(cmd, ctx, "interval-seconds")
		addFlagInt(cmd, ctx, "timeout-seconds")
	}, func(args []string) error {
		return runTargetCloudSyncGatewayWait(ctx, args)
	})
	configureTargetCloudSyncGatewayLeafHelp(waitCmd, ctx, "cmd.target.cloud_sync_gateway.wait.usage_line", "cmd.target.cloud_sync_gateway.wait.usage_notes")

	resourcesCmd := newRawLeafCommand(ctx, "resources", "cmd.block_storages.resources.short", "cmd.block_storages.resources.long", "cmd.block_storages.resources.examples", "cmd.block_storages.resources.notes", func(cmd *cobra.Command) {
		for _, name := range []string{"cloud-account-id", "fetch-res", "region-id", "zone-id", "flavor-id", "flavor-vcpus", "flavor-ram", "purpose", "image-type"} {
			addFlagString(cmd, ctx, name)
		}
	}, func(args []string) error {
		return runBlockStorages(ctx, append([]string{"resources"}, args...))
	})
	configureTargetCloudSyncGatewayLeafHelp(resourcesCmd, ctx, "cmd.target.cloud_sync_gateway.resources.usage_line", "cmd.target.cloud_sync_gateway.resources.usage_notes")

	subnetConfigCmd := newRawLeafCommand(ctx, "subnet-config", "cmd.block_storages.subnet_config.short", "cmd.block_storages.subnet_config.long", "cmd.block_storages.subnet_config.examples", "cmd.block_storages.subnet_config.notes", func(cmd *cobra.Command) {
		for _, name := range []string{"cloud-account-id", "cloud-type", "region-id", "zone-id", "network-id", "subnet-id"} {
			addFlagString(cmd, ctx, name)
		}
	}, func(args []string) error {
		return runBlockStorages(ctx, append([]string{"subnet-config"}, args...))
	})
	configureTargetCloudSyncGatewayLeafHelp(subnetConfigCmd, ctx, "cmd.target.cloud_sync_gateway.subnet_config.usage_line", "cmd.target.cloud_sync_gateway.subnet_config.usage_notes")

	createCmd := newBlockStoragesCreateCommand(ctx)

	cmd.AddCommand(
		listCmd,
		detailCmd,
		waitCmd,
		resourcesCmd,
		subnetConfigCmd,
		createCmd,
	)
	return cmd
}

func configureTargetCloudSyncGatewayLeafHelp(cmd *cobra.Command, ctx *context, usageLineKey, usageNotesKey string) {
	addHelpLayout(cmd, helpLayoutFourSection)
	addUsageLine(cmd, ctx, usageLineKey)
	addUsageNotes(cmd, ctx, usageNotesKey)
}
