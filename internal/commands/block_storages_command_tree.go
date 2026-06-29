package commands

import (
	"fmt"
	"strings"

	"hyperbdr-client/catalog"
	workflowcreate "hyperbdr-client/internal/workflow/blockstoragecreate"

	"github.com/spf13/cobra"
)

const cloudSyncGatewayCreateHelpProfileAnnotation = "cloud-sync-gateway-create-help-profile"

type cloudSyncGatewayCreateProfile struct {
	Entry    catalog.CloudEntry
	Provider string
	Kind     string
}

func newCloudSyncGatewayCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "cloud-sync-gateway", "cmd.cloud_sync_gateway.short", "cmd.cloud_sync_gateway.long", "cmd.cloud_sync_gateway.examples", "cmd.cloud_sync_gateway.notes", "cloud-sync-gateway")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.cloud_sync_gateway.usage_line")
	addUsageNotes(cmd, ctx, "cmd.cloud_sync_gateway.usage_notes")

	listCmd := newRawLeafCommand(ctx, "list", "cmd.block_storages.list.short", "cmd.block_storages.list.long", "cmd.block_storages.list.examples", "cmd.block_storages.list.notes", func(cmd *cobra.Command) {
		addFlagInt(cmd, ctx, "page")
		addFlagInt(cmd, ctx, "page-size")
		addFlagString(cmd, ctx, "type")
		addFlagString(cmd, ctx, "cloud-account-id")
	}, func(args []string) error {
		return runBlockStorages(ctx, append([]string{"list"}, args...))
	})
	configureTargetCloudSyncGatewayLeafHelp(listCmd, ctx, "cmd.cloud_sync_gateway.list.usage_line", "cmd.cloud_sync_gateway.list.usage_notes")

	detailCmd := newRawLeafCommand(ctx, "detail", "cmd.block_storages.detail.short", "cmd.block_storages.detail.long", "cmd.block_storages.detail.examples", "cmd.block_storages.detail.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
	}, func(args []string) error {
		return runBlockStorages(ctx, append([]string{"detail"}, args...))
	})
	configureTargetCloudSyncGatewayLeafHelp(detailCmd, ctx, "cmd.cloud_sync_gateway.detail.usage_line", "cmd.cloud_sync_gateway.detail.usage_notes")

	deleteCmd := newRawLeafCommand(ctx, "delete", "cmd.block_storages.delete.short", "cmd.block_storages.delete.long", "cmd.block_storages.delete.examples", "cmd.block_storages.delete.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
		addFlagString(cmd, ctx, "ids")
		addFlagBool(cmd, ctx, "force")
	}, func(args []string) error {
		return runBlockStorages(ctx, append([]string{"delete"}, args...))
	})
	configureTargetCloudSyncGatewayLeafHelp(deleteCmd, ctx, "cmd.cloud_sync_gateway.delete.usage_line", "cmd.cloud_sync_gateway.delete.usage_notes")

	waitCmd := newRawLeafCommand(ctx, "wait", "cmd.block_storages.wait.short", "cmd.block_storages.wait.long", "cmd.block_storages.wait.examples", "cmd.block_storages.wait.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
		addFlagInt(cmd, ctx, "interval-seconds")
		addFlagInt(cmd, ctx, "timeout-seconds")
	}, func(args []string) error {
		return runTargetCloudSyncGatewayWait(ctx, args)
	})
	configureTargetCloudSyncGatewayLeafHelp(waitCmd, ctx, "cmd.cloud_sync_gateway.wait.usage_line", "cmd.cloud_sync_gateway.wait.usage_notes")

	cmd.AddCommand(
		listCmd,
		detailCmd,
		deleteCmd,
		waitCmd,
		newCloudSyncGatewayCreateCommand(ctx),
	)
	return cmd
}

func newCloudSyncGatewayCreateCommand(ctx *context) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "create",
		Short:              ctx.loc.T("cmd.block_storages.create.short"),
		Long:               ctx.loc.T("cmd.block_storages.create.long"),
		Example:            strings.TrimSpace(ctx.loc.T("cmd.cloud_sync_gateway.create.examples")),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider, err := parseCloudSyncGatewayCreateHelpProvider(args)
			if err != nil {
				return err
			}
			if rawArgsHelp(cmd, args) {
				return renderCloudSyncGatewayCreateHelp(ctx, cmd, provider)
			}
			return runCloudSyncGatewayCreate(ctx, args)
		},
	}
	addGenericBlockStorageCreateFlags(cmd, ctx)
	addFlagString(cmd, ctx, "cloud-type")
	addHelpLayout(cmd, helpLayoutFourSection)
	addUsageLine(cmd, ctx, "cmd.cloud_sync_gateway.create.usage_line")
	addUsageNotes(cmd, ctx, "cmd.cloud_sync_gateway.create.usage_notes")
	return cmd
}

func runCloudSyncGatewayCreate(ctx *context, args []string) error {
	parsed, err := parseBlockStorageCreateArgs("cloud-sync-gateway create", args)
	if err != nil {
		return err
	}
	if len(parsed.remainingArgs) > 0 {
		return errUnknown("cloud-sync-gateway create", parsed.remainingArgs[0])
	}
	provider := strings.TrimSpace(parsed.spec.CloudType)
	if provider == "" {
		return fmt.Errorf("cloud-type is required")
	}
	profile, err := resolveCloudSyncGatewayCreateProfile(provider)
	if err != nil {
		return err
	}
	parsed.spec.CloudType = profile.Entry.CloudType
	return executeBlockStorageCreateSpec(ctx, parsed.spec, parsed.previewRequest)
}

func parseCloudSyncGatewayCreateHelpProvider(args []string) (string, error) {
	var provider string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--help" || arg == "-h" {
			continue
		}
		if !strings.HasPrefix(arg, "--") {
			return "", errUnknown("cloud-sync-gateway create", arg)
		}
		name, value, hasInline := splitFlag(arg)
		if name != "--cloud-type" {
			if !hasInline && flagConsumesValue(name) && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				i++
			}
			continue
		}
		v, next, err := strictFlagValue(args, i, value, hasInline)
		if err != nil {
			return "", err
		}
		provider = v
		i = next
	}
	return provider, nil
}

func flagConsumesValue(name string) bool {
	switch name {
	case "--preview-request", "--debug", "-h", "--help", "-G", "--vertical":
		return false
	default:
		return true
	}
}

func renderCloudSyncGatewayCreateHelp(ctx *context, cmd *cobra.Command, provider string) error {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		addAnnotationValue(cmd, cloudSyncGatewayCreateHelpProfileAnnotation, "generic")
		addAnnotationValue(cmd, usageNotesAnnotation, cloudSyncGatewayCreateGenericUsageNotes(ctx))
		return renderHelp(cmd, ctx)
	}
	profile, err := resolveCloudSyncGatewayCreateProfile(provider)
	if err != nil {
		return err
	}
	addAnnotationValue(cmd, cloudSyncGatewayCreateHelpProfileAnnotation, profile.Kind)
	addAnnotationValue(cmd, usageLineAnnotation, fmt.Sprintf(ctx.loc.T("cmd.cloud_sync_gateway.create.provider.usage_line"), profile.Provider))
	addAnnotationValue(cmd, usageNotesAnnotation, cloudSyncGatewayCreateUsageNotes(ctx, profile))
	return renderHelp(cmd, ctx)
}

func resolveCloudSyncGatewayCreateProfile(provider string) (cloudSyncGatewayCreateProfile, error) {
	needle := strings.ToLower(strings.TrimSpace(provider))
	for _, entry := range catalog.EnabledBlockClouds() {
		if strings.ToLower(strings.TrimSpace(entry.Provider)) != needle {
			continue
		}
		kind := "provider"
		if workflowcreate.HasRegisteredAdapter(entry.CloudType) {
			switch entry.Provider {
			case "aliyun":
				kind = "aliyun"
			case "openstack":
				kind = "openstack"
			}
		}
		return cloudSyncGatewayCreateProfile{Entry: entry, Provider: entry.Provider, Kind: kind}, nil
	}
	return cloudSyncGatewayCreateProfile{}, fmt.Errorf("cloud-type %q does not support cloud-sync-gateway create", provider)
}

func cloudSyncGatewayCreateUsageNotes(ctx *context, profile cloudSyncGatewayCreateProfile) string {
	var notes string
	switch profile.Kind {
	case "aliyun":
		notes = ctx.loc.T("cmd.target.cloud_sync_gateway.create.aliyun.usage_notes")
	case "openstack":
		notes = ctx.loc.T("cmd.target.cloud_sync_gateway.create.openstack.usage_notes")
	default:
		notes = fmt.Sprintf(ctx.loc.T("cmd.target.cloud_sync_gateway.create.provider.usage_notes"), localizedCloudEntryName(ctx, profile.Entry), profile.Provider, profile.Entry.CloudType, profile.Provider, profile.Provider)
	}
	return rewriteCloudSyncGatewayCommandRefs(notes, profile.Provider)
}

func cloudSyncGatewayCreateGenericUsageNotes(ctx *context) string {
	notes := strings.TrimSpace(ctx.loc.T("cmd.cloud_sync_gateway.create.usage_notes"))
	label := "Providers:"
	if ctx.loc.Lang() == "zh_cn" {
		label = "云厂商:"
	}
	var b strings.Builder
	b.WriteString(notes)
	b.WriteString("\n\n")
	b.WriteString(label)
	for _, entry := range catalog.EnabledBlockClouds() {
		b.WriteString("\n  ")
		b.WriteString(entry.Provider)
	}
	return b.String()
}

func rewriteCloudSyncGatewayCommandRefs(notes, provider string) string {
	replacements := []struct {
		old string
		new string
	}{
		{"hyperbdrctl target cloud-sync-gateway create aliyun", "hyperbdrctl cloud-sync-gateway create --cloud-type aliyun"},
		{"hyperbdrctl target cloud-sync-gateway create openstack", "hyperbdrctl cloud-sync-gateway create --cloud-type openstack"},
		{"target cloud-sync-gateway create aliyun", "cloud-sync-gateway create --cloud-type aliyun"},
		{"target cloud-sync-gateway create openstack", "cloud-sync-gateway create --cloud-type openstack"},
		{fmt.Sprintf("hyperbdrctl target cloud-sync-gateway create %s", provider), fmt.Sprintf("hyperbdrctl cloud-sync-gateway create --cloud-type %s", provider)},
		{fmt.Sprintf("target cloud-sync-gateway create %s", provider), fmt.Sprintf("cloud-sync-gateway create --cloud-type %s", provider)},
		{"hyperbdrctl cloud-sync-gateway create aliyun", "hyperbdrctl cloud-sync-gateway create --cloud-type aliyun"},
		{"hyperbdrctl cloud-sync-gateway create openstack", "hyperbdrctl cloud-sync-gateway create --cloud-type openstack"},
		{"cloud-sync-gateway create aliyun", "cloud-sync-gateway create --cloud-type aliyun"},
		{"cloud-sync-gateway create openstack", "cloud-sync-gateway create --cloud-type openstack"},
		{fmt.Sprintf("hyperbdrctl cloud-sync-gateway create %s", provider), fmt.Sprintf("hyperbdrctl cloud-sync-gateway create --cloud-type %s", provider)},
		{fmt.Sprintf("cloud-sync-gateway create %s", provider), fmt.Sprintf("cloud-sync-gateway create --cloud-type %s", provider)},
		{"hyperbdrctl target cloud-sync-gateway resources", "hyperbdrctl cloud-resource fetch"},
		{"target cloud-sync-gateway resources", "cloud-resource fetch"},
		{"hyperbdrctl target cloud-sync-gateway subnet-config --help", "hyperbdrctl cloud-resource fetch --help"},
		{"target cloud-sync-gateway subnet-config --help", "cloud-resource fetch --help"},
		{"hyperbdrctl cloud-sync-gateway resources", "hyperbdrctl cloud-resource fetch"},
		{"cloud-sync-gateway resources", "cloud-resource fetch"},
		{"hyperbdrctl cloud-sync-gateway subnet-config --help", "hyperbdrctl cloud-resource fetch --help"},
		{"cloud-sync-gateway subnet-config --help", "cloud-resource fetch --help"},
		{"hyperbdrctl target cloud-sync-gateway", "hyperbdrctl cloud-sync-gateway"},
		{"target cloud-sync-gateway", "cloud-sync-gateway"},
	}
	for _, replacement := range replacements {
		notes = strings.ReplaceAll(notes, replacement.old, replacement.new)
	}
	return notes
}

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

	deleteCmd := newRawLeafCommand(ctx, "delete", "cmd.block_storages.delete.short", "cmd.block_storages.delete.long", "cmd.block_storages.delete.examples", "cmd.block_storages.delete.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
		addFlagString(cmd, ctx, "ids")
		addFlagBool(cmd, ctx, "force")
	}, func(args []string) error {
		return runBlockStorages(ctx, append([]string{"delete"}, args...))
	})
	configureTargetCloudSyncGatewayLeafHelp(deleteCmd, ctx, "cmd.target.cloud_sync_gateway.delete.usage_line", "cmd.target.cloud_sync_gateway.delete.usage_notes")

	waitCmd := newRawLeafCommand(ctx, "wait", "cmd.block_storages.wait.short", "cmd.block_storages.wait.long", "cmd.block_storages.wait.examples", "cmd.block_storages.wait.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
		addFlagInt(cmd, ctx, "interval-seconds")
		addFlagInt(cmd, ctx, "timeout-seconds")
	}, func(args []string) error {
		return runTargetCloudSyncGatewayWait(ctx, args)
	})
	configureTargetCloudSyncGatewayLeafHelp(waitCmd, ctx, "cmd.target.cloud_sync_gateway.wait.usage_line", "cmd.target.cloud_sync_gateway.wait.usage_notes")

	createCmd := newBlockStoragesCreateCommand(ctx)

	cmd.AddCommand(
		listCmd,
		detailCmd,
		deleteCmd,
		waitCmd,
		createCmd,
	)
	return cmd
}

func configureTargetCloudSyncGatewayLeafHelp(cmd *cobra.Command, ctx *context, usageLineKey, usageNotesKey string) {
	addHelpLayout(cmd, helpLayoutFourSection)
	addUsageLine(cmd, ctx, usageLineKey)
	addUsageNotes(cmd, ctx, usageNotesKey)
}
