package commands

import (
	"fmt"
	"strings"

	"hyperbdr-client/catalog"
	workflowcreate "hyperbdr-client/internal/workflow/cloudaccountcreate"

	"github.com/spf13/cobra"
)

func newCloudAccountsCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "account", "cmd.target.account.short", "cmd.target.account.long", "cmd.target.account.examples", "cmd.target.account.notes", "target account")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.target.account.usage_line")
	addUsageNotes(cmd, ctx, "cmd.target.account.usage_notes")

	listCmd := newRawLeafCommand(ctx, "list", "cmd.cloud_accounts.list.short", "cmd.cloud_accounts.list.long", "cmd.cloud_accounts.list.examples", "cmd.cloud_accounts.list.notes", func(cmd *cobra.Command) {
		addFlagInt(cmd, ctx, "page")
		addFlagInt(cmd, ctx, "page-size")
		addFlagString(cmd, ctx, "storage-type")
	}, func(args []string) error {
		return runTargetAccounts(ctx, append([]string{"list"}, args...))
	})
	configureTargetAccountLeafHelp(listCmd, ctx, "cmd.target.account.list.usage_line", "cmd.target.account.list.usage_notes")

	detailCmd := newRawLeafCommand(ctx, "detail", "cmd.cloud_accounts.detail.short", "cmd.cloud_accounts.detail.long", "cmd.cloud_accounts.detail.examples", "cmd.cloud_accounts.detail.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
	}, func(args []string) error {
		return runTargetAccounts(ctx, append([]string{"detail"}, args...))
	})
	configureTargetAccountLeafHelp(detailCmd, ctx, "cmd.target.account.detail.usage_line", "cmd.target.account.detail.usage_notes")

	waitCmd := newRawLeafCommand(ctx, "wait", "cmd.cloud_accounts.wait.short", "cmd.cloud_accounts.wait.long", "cmd.cloud_accounts.wait.examples", "cmd.cloud_accounts.wait.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
		addFlagInt(cmd, ctx, "interval-seconds")
		addFlagInt(cmd, ctx, "timeout-seconds")
	}, func(args []string) error {
		return runTargetAccountWait(ctx, args)
	})
	configureTargetAccountLeafHelp(waitCmd, ctx, "cmd.target.account.wait.usage_line", "cmd.target.account.wait.usage_notes")

	fetchResourcesCmd := newDeprecatedCloudAccountsFetchResourcesCommand(ctx)

	fetchBlockResourcesCmd := newCloudAccountsFetchBlockResourcesCommand(ctx)
	addHelpLayout(fetchBlockResourcesCmd, helpLayoutGroup)
	addUsageLine(fetchBlockResourcesCmd, ctx, "cmd.target.account.fetch_block_resources.usage_line")
	addUsageNotes(fetchBlockResourcesCmd, ctx, "cmd.target.account.fetch_block_resources.usage_notes")

	fetchOSSResourcesCmd := newCloudAccountsFetchOSSResourcesCommand(ctx)
	addHelpLayout(fetchOSSResourcesCmd, helpLayoutGroup)
	addUsageLine(fetchOSSResourcesCmd, ctx, "cmd.target.account.fetch_oss_resources.usage_line")
	addUsageNotes(fetchOSSResourcesCmd, ctx, "cmd.target.account.fetch_oss_resources.usage_notes")

	createCmd := newCloudAccountsCreateCommand(ctx)
	configureTargetAccountLeafHelp(createCmd, ctx, "cmd.target.account.create.usage_line", "cmd.target.account.create.usage_notes")

	createBlockCmd := newCloudAccountsCreateBlockCommand(ctx)
	addHelpLayout(createBlockCmd, helpLayoutGroup)
	addUsageLine(createBlockCmd, ctx, "cmd.target.account.create_block.usage_line")
	addUsageNotes(createBlockCmd, ctx, "cmd.target.account.create_block.usage_notes")

	createOSSCmd := newCloudAccountsCreateOSSCommand(ctx)
	addHelpLayout(createOSSCmd, helpLayoutGroup)
	addUsageLine(createOSSCmd, ctx, "cmd.target.account.create_oss.usage_line")
	addUsageNotes(createOSSCmd, ctx, "cmd.target.account.create_oss.usage_notes")

	deleteCmd := newRawLeafCommand(ctx, "delete", "cmd.cloud_accounts.delete.short", "cmd.cloud_accounts.delete.long", "cmd.cloud_accounts.delete.examples", "cmd.cloud_accounts.delete.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "id")
		addFlagString(cmd, ctx, "storage-type")
		addFlagBool(cmd, ctx, "force")
	}, func(args []string) error {
		return runTargetAccounts(ctx, append([]string{"delete"}, args...))
	})
	configureTargetAccountLeafHelp(deleteCmd, ctx, "cmd.target.account.delete.usage_line", "cmd.target.account.delete.usage_notes")

	cmd.AddCommand(
		listCmd,
		detailCmd,
		waitCmd,
		fetchBlockResourcesCmd,
		fetchOSSResourcesCmd,
		createCmd,
		createBlockCmd,
		createOSSCmd,
		deleteCmd,
		fetchResourcesCmd,
	)
	return cmd
}

func newTargetAccountCommand(ctx *context) *cobra.Command {
	return newCloudAccountsCommand(ctx)
}

func configureTargetAccountLeafHelp(cmd *cobra.Command, ctx *context, usageLineKey, usageNotesKey string) {
	addHelpLayout(cmd, helpLayoutFourSection)
	addUsageLine(cmd, ctx, usageLineKey)
	addUsageNotes(cmd, ctx, usageNotesKey)
}

func newCloudAccountsCreateCommand(ctx *context) *cobra.Command {
	cmd := newRawLeafCommand(ctx, "create", "cmd.cloud_accounts.create_group.short", "cmd.cloud_accounts.create_group.long", "cmd.cloud_accounts.create_group.examples", "cmd.cloud_accounts.create_group.notes", func(cmd *cobra.Command) {
		addFlagString(cmd, ctx, "file")
		addFlagString(cmd, ctx, "body")
		addFlagBool(cmd, ctx, "preview-request")
	}, func(args []string) error {
		return runCreateCloudAccountRaw(ctx, args)
	})
	addWorkflow(cmd, ctx, "cmd.cloud_accounts.create_group.workflow")
	addRelatedCommands(cmd, ctx, "cmd.cloud_accounts.create_group.related")
	addNextSteps(cmd, ctx, "cmd.cloud_accounts.create_group.next_steps")
	return cmd
}

func newDeprecatedCloudAccountsFetchResourcesCommand(ctx *context) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "fetch-resources",
		Short:              ctx.loc.T("cmd.cloud_accounts.fetch_resources.short"),
		Long:               ctx.loc.T("cmd.cloud_accounts.fetch_resources.long"),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		Hidden:             true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := rejectLegacyGlobalFlags(cmd, args); err != nil {
				return err
			}
			return errDeprecatedFetchResourcesFlags()
		},
	}
	return cmd
}

func newCloudAccountsFetchBlockResourcesCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "fetch-block-resources", "cmd.cloud_accounts.fetch.block.short", "cmd.cloud_accounts.fetch.block.long", "cmd.cloud_accounts.fetch.block.examples", "cmd.cloud_accounts.fetch.block.notes", "target account fetch-block-resources")
	for _, entry := range catalog.EnabledBlockClouds() {
		cmd.AddCommand(newCloudAccountFetchResourcesProviderCommand(ctx, entry, "block"))
	}
	return cmd
}

func newCloudAccountsFetchOSSResourcesCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "fetch-oss-resources", "cmd.cloud_accounts.fetch.oss.short", "cmd.cloud_accounts.fetch.oss.long", "cmd.cloud_accounts.fetch.oss.examples", "cmd.cloud_accounts.fetch.oss.notes", "target account fetch-oss-resources")
	for _, entry := range catalog.EnabledObjectClouds() {
		if entry.Provider == "openstack" {
			cmd.AddCommand(newCloudAccountFetchResourcesOpenStackObjectCommand(ctx, entry))
			continue
		}
		cmd.AddCommand(newCloudAccountFetchResourcesProviderCommand(ctx, entry, "objectstorage"))
	}
	return cmd
}

func newCloudAccountFetchResourcesProviderCommand(ctx *context, entry catalog.CloudEntry, storageType string) *cobra.Command {
	shortKey, longKey, usageLineKey, usageNotesKey, exampleKey := cloudAccountFetchProviderTextKeys(storageType)
	displayName := localizedCloudEntryName(ctx, entry)

	cmd := &cobra.Command{
		Use:                entry.Provider,
		Short:              fmt.Sprintf(ctx.loc.T(shortKey), displayName),
		Long:               fmt.Sprintf(ctx.loc.T(longKey), displayName, entry.CloudType),
		Example:            strings.TrimSpace(fmt.Sprintf(ctx.loc.T(exampleKey), entry.Provider)),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rawArgsHelp(cmd, args) {
				return renderHelp(cmd, ctx)
			}
			if err := rejectLegacyGlobalFlags(cmd, args); err != nil {
				return err
			}
			return runFetchResourcesForProvider(ctx, fetchResourcesProviderCommandName(storageType, entry.Provider), entry.CloudType, fetchResourcesStorageType(storageType), args)
		},
	}

	for _, name := range []string{"access-key-id", "access-key-secret", "region-id", "boot-mode", "fetch-res"} {
		addFlagString(cmd, ctx, name)
	}

	addHelpLayout(cmd, helpLayoutFourSection)
	addAnnotationValue(cmd, usageLineAnnotation, fmt.Sprintf(ctx.loc.T(usageLineKey), entry.Provider))
	addAnnotationValue(cmd, usageNotesAnnotation, fmt.Sprintf(ctx.loc.T(usageNotesKey), entry.Provider, entry.Provider))
	return cmd
}

func newCloudAccountFetchResourcesOpenStackObjectCommand(ctx *context, entry catalog.CloudEntry) *cobra.Command {
	displayName := localizedCloudEntryName(ctx, entry)

	cmd := &cobra.Command{
		Use:                entry.Provider,
		Short:              fmt.Sprintf(ctx.loc.T("cmd.cloud_accounts.fetch.provider.oss.short"), displayName),
		Long:               fmt.Sprintf(ctx.loc.T("cmd.cloud_accounts.fetch.provider.oss.long"), displayName, entry.CloudType),
		Example:            strings.TrimSpace(ctx.loc.T("cmd.target.account.fetch_oss_resources.openstack.examples")),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rawArgsHelp(cmd, args) {
				return renderHelp(cmd, ctx)
			}
			if err := rejectLegacyGlobalFlags(cmd, args); err != nil {
				return err
			}
			return runFetchOpenStackObjectResources(ctx, "target account fetch-oss-resources openstack", args)
		},
	}

	for _, name := range []string{"auth-url", "cloud-account-username", "cloud-account-password", "user-domain-id", "fetch-res", "region-id", "project-id", "project-domain-id", "project-name", "compute-zone-id", "block-store-zone-id"} {
		addFlagString(cmd, ctx, name)
	}
	overrideFlagUsage(cmd, ctx, "auth-url", "help.target_account_fetch_object_openstack.flag.auth-url")
	overrideFlagUsage(cmd, ctx, "cloud-account-username", "help.target_account_fetch_object_openstack.flag.cloud-account-username")
	overrideFlagUsage(cmd, ctx, "cloud-account-password", "help.target_account_fetch_object_openstack.flag.cloud-account-password")
	overrideFlagUsage(cmd, ctx, "user-domain-id", "help.target_account_fetch_object_openstack.flag.user-domain-id")
	overrideFlagUsage(cmd, ctx, "fetch-res", "help.target_account_fetch_object_openstack.flag.fetch-res")
	overrideFlagUsage(cmd, ctx, "region-id", "help.target_account_fetch_object_openstack.flag.region-id")
	overrideFlagUsage(cmd, ctx, "project-id", "help.target_account_fetch_object_openstack.flag.project-id")
	overrideFlagUsage(cmd, ctx, "project-domain-id", "help.target_account_fetch_object_openstack.flag.project-domain-id")
	overrideFlagUsage(cmd, ctx, "project-name", "help.target_account_fetch_object_openstack.flag.project-name")
	overrideFlagUsage(cmd, ctx, "compute-zone-id", "help.target_account_fetch_object_openstack.flag.compute-zone-id")
	overrideFlagUsage(cmd, ctx, "block-store-zone-id", "help.target_account_fetch_object_openstack.flag.block-store-zone-id")

	addHelpLayout(cmd, helpLayoutFourSection)
	addAnnotationValue(cmd, usageLineAnnotation, ctx.loc.T("cmd.target.account.fetch_oss_resources.openstack.usage_line"))
	addAnnotationValue(cmd, usageNotesAnnotation, ctx.loc.T("cmd.target.account.fetch_oss_resources.openstack.usage_notes"))
	return cmd
}

func newCloudAccountsCreateBlockCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "create-block", "cmd.cloud_accounts.create.block.short", "cmd.cloud_accounts.create.block.long", "cmd.cloud_accounts.create.block.examples", "cmd.cloud_accounts.create.block.notes", "target account create-block")
	addWorkflow(cmd, ctx, "cmd.cloud_accounts.create.block.workflow")
	addRelatedCommands(cmd, ctx, "cmd.cloud_accounts.create.block.related")
	addNextSteps(cmd, ctx, "cmd.cloud_accounts.create.block.next_steps")
	for _, entry := range catalog.EnabledBlockClouds() {
		cmd.AddCommand(newCloudAccountsCreateProviderCommand(ctx, entry, "block"))
	}
	return cmd
}

func newCloudAccountsCreateOSSCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "create-oss", "cmd.cloud_accounts.create.oss.short", "cmd.cloud_accounts.create.oss.long", "cmd.cloud_accounts.create.oss.examples", "cmd.cloud_accounts.create.oss.notes", "target account create-oss")
	addWorkflow(cmd, ctx, "cmd.cloud_accounts.create.oss.workflow")
	addRelatedCommands(cmd, ctx, "cmd.cloud_accounts.create.oss.related")
	addNextSteps(cmd, ctx, "cmd.cloud_accounts.create.oss.next_steps")
	for _, entry := range catalog.EnabledObjectClouds() {
		cmd.AddCommand(newCloudAccountsCreateProviderCommand(ctx, entry, "objectstorage"))
	}
	return cmd
}

func newCloudAccountsCreateProviderCommand(ctx *context, entry catalog.CloudEntry, storageType string) *cobra.Command {
	specialized := workflowcreate.HasRegisteredAdapter(entry.CloudType, storageType)
	shortKey, longKey, exampleKey, notesKey, workflowKey, minimumKey, optionalKey, relatedKey, nextKey := cloudAccountCreateProviderTextKeys(entry.Provider, storageType, specialized)

	shortText, longText := providerCommandTexts(ctx, entry, storageType, specialized, shortKey, longKey)

	cmd := &cobra.Command{
		Use:                entry.Provider,
		Short:              shortText,
		Long:               longText,
		Example:            providerExample(ctx, entry, storageType, specialized, exampleKey),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rawArgsHelp(cmd, args) {
				return renderHelp(cmd, ctx)
			}
			if err := rejectLegacyGlobalFlags(cmd, args); err != nil {
				return err
			}
			return runCreateCloudAccountForProvider(ctx, providerCommandName(storageType, entry.Provider), entry.CloudType, storageType, specialized, args)
		},
	}

	addHelpLayout(cmd, helpLayoutFourSection)
	addAnnotationValue(cmd, usageLineAnnotation, fmt.Sprintf(ctx.loc.T(cloudAccountCreateProviderUsageLineKey(storageType)), entry.Provider))
	addAnnotationValue(cmd, usageNotesAnnotation, cloudAccountCreateProviderUsageNotes(ctx, entry, storageType, specialized))

	switch {
	case specialized && storageType == "block" && entry.Provider == "aliyun":
		addCloudAccountCreateAliyunBlockFlags(cmd, ctx)
	case specialized && storageType == "block" && entry.Provider == "openstack":
		addCloudAccountCreateOpenStackBlockFlags(cmd, ctx)
	case specialized && storageType == "objectstorage" && entry.Provider == "aliyun":
		addCloudAccountCreateAliyunObjectFlags(cmd, ctx)
	case specialized && storageType == "objectstorage" && entry.Provider == "openstack":
		addCloudAccountCreateOpenStackObjectFlags(cmd, ctx)
	case storageType == "block":
		addCloudAccountCreateGenericBlockFlags(cmd, ctx, false)
	default:
		addCloudAccountCreateGenericOSSFlags(cmd, ctx, false)
	}

	if notesKey != "" {
		addNotes(cmd, ctx, notesKey)
	}
	if workflowKey != "" {
		addWorkflow(cmd, ctx, workflowKey)
	}
	if minimumKey != "" {
		addMinimumFlags(cmd, ctx, minimumKey)
	}
	if optionalKey != "" {
		addCommonOptionalFlags(cmd, ctx, optionalKey)
	}
	if relatedKey != "" {
		addRelatedCommands(cmd, ctx, relatedKey)
	}
	if nextKey != "" {
		addNextSteps(cmd, ctx, nextKey)
	}
	if storageType == "objectstorage" && entry.Provider == "aliyun" {
		addQuickStart(cmd, ctx, "cmd.cloud_accounts.create.object.aliyun.quick_start")
		addAutomaticBehavior(cmd, ctx, "cmd.cloud_accounts.create.object.aliyun.automatic_behavior")
	}
	return cmd
}

func providerCommandTexts(ctx *context, entry catalog.CloudEntry, storageType string, specialized bool, shortKey, longKey string) (string, string) {
	if specialized {
		return ctx.loc.T(shortKey), ctx.loc.T(longKey)
	}

	displayName := localizedCloudEntryName(ctx, entry)
	shortText := fmt.Sprintf(ctx.loc.T(shortKey), displayName)
	longText := fmt.Sprintf(ctx.loc.T(longKey), displayName, entry.CloudType)
	return shortText, longText
}

func localizedCloudEntryName(ctx *context, entry catalog.CloudEntry) string {
	if ctx.loc.Lang() == "zh_cn" {
		return entry.NameZhCN
	}
	return entry.NameEn
}

func addCloudAccountCreateAliyunBlockFlags(cmd *cobra.Command, ctx *context) {
	for _, name := range []string{"access-key-id", "access-key-secret", "region-id", "region-name", "account-name", "auth-region-id"} {
		addFlagString(cmd, ctx, name)
	}
	addFlagString(cmd, ctx, "file")
	cmd.Flags().StringArray("set", nil, ctx.loc.T("flag.set"))
	cmd.Flags().StringArray("set-json", nil, ctx.loc.T("flag.set-json"))
	addFlagBool(cmd, ctx, "preview-request")
}

func addCloudAccountCreateOpenStackBlockFlags(cmd *cobra.Command, ctx *context) {
	for _, name := range []string{"auth-url", "username", "password", "user-domain-id", "project-domain-id", "project-name", "region-name", "ssh-port", "ssh-pass", "linux-hd-username", "linux-hd-password", "linux-hd-port"} {
		addFlagString(cmd, ctx, name)
	}
	overrideFlagUsage(cmd, ctx, "auth-url", "help.target_account_create_block_openstack.flag.auth-url")
	overrideFlagUsage(cmd, ctx, "username", "help.target_account_create_block_openstack.flag.username")
	overrideFlagUsage(cmd, ctx, "password", "help.target_account_create_block_openstack.flag.password")
	overrideFlagUsage(cmd, ctx, "user-domain-id", "help.target_account_create_block_openstack.flag.user-domain-id")
	overrideFlagUsage(cmd, ctx, "project-domain-id", "help.target_account_create_block_openstack.flag.project-domain-id")
	overrideFlagUsage(cmd, ctx, "project-name", "help.target_account_create_block_openstack.flag.project-name")
	overrideFlagUsage(cmd, ctx, "region-name", "help.target_account_create_block_openstack.flag.region-name")
	overrideFlagUsage(cmd, ctx, "ssh-port", "help.target_account_create_block_openstack.flag.ssh-port")
	overrideFlagUsage(cmd, ctx, "ssh-pass", "help.target_account_create_block_openstack.flag.ssh-pass")
	overrideFlagUsage(cmd, ctx, "linux-hd-username", "help.target_account_create_block_openstack.flag.linux-hd-username")
	overrideFlagUsage(cmd, ctx, "linux-hd-password", "help.target_account_create_block_openstack.flag.linux-hd-password")
	overrideFlagUsage(cmd, ctx, "linux-hd-port", "help.target_account_create_block_openstack.flag.linux-hd-port")
	addFlagString(cmd, ctx, "file")
	cmd.Flags().StringArray("set", nil, ctx.loc.T("flag.set"))
	cmd.Flags().StringArray("set-json", nil, ctx.loc.T("flag.set-json"))
	addFlagBool(cmd, ctx, "preview-request")
}

func addCloudAccountCreateAliyunObjectFlags(cmd *cobra.Command, ctx *context) {
	for _, name := range []string{"access-key-id", "access-key-secret", "region-id", "use-internal-ip", "boot-loader-image-id", "boot-loader-image-name", "linux-boot-image-id", "windows-boot-image-id", "linux-uefi-boot-image-id", "windows-uefi-boot-image-id"} {
		addFlagString(cmd, ctx, name)
	}
	for _, name := range []string{"region-name", "boot-loader-flavor-id", "custom-name", "file"} {
		addFlagString(cmd, ctx, name)
	}
	cmd.Flags().StringArray("set", nil, ctx.loc.T("flag.set"))
	cmd.Flags().StringArray("set-json", nil, ctx.loc.T("flag.set-json"))
	addFlagBool(cmd, ctx, "preview-request")
}

func addCloudAccountCreateOpenStackObjectFlags(cmd *cobra.Command, ctx *context) {
	for _, name := range []string{"auth-url", "username", "password", "user-domain-id", "project-domain-id", "project-id", "project-name", "region-id", "region-name", "use-internal-ip", "boot-loader-image-id", "boot-loader-image-name", "linux-boot-image-id", "windows-boot-image-id", "custom-name", "disk-bus-type-id", "disk-bus-type-name"} {
		addFlagString(cmd, ctx, name)
	}
	overrideFlagUsage(cmd, ctx, "auth-url", "help.target_account_create_object_openstack.flag.auth-url")
	overrideFlagUsage(cmd, ctx, "username", "help.target_account_create_object_openstack.flag.username")
	overrideFlagUsage(cmd, ctx, "password", "help.target_account_create_object_openstack.flag.password")
	overrideFlagUsage(cmd, ctx, "user-domain-id", "help.target_account_create_object_openstack.flag.user-domain-id")
	overrideFlagUsage(cmd, ctx, "project-domain-id", "help.target_account_create_object_openstack.flag.project-domain-id")
	overrideFlagUsage(cmd, ctx, "project-id", "help.target_account_create_object_openstack.flag.project-id")
	overrideFlagUsage(cmd, ctx, "project-name", "help.target_account_create_object_openstack.flag.project-name")
	overrideFlagUsage(cmd, ctx, "region-id", "help.target_account_create_object_openstack.flag.region-id")
	overrideFlagUsage(cmd, ctx, "region-name", "help.target_account_create_object_openstack.flag.region-name")
	overrideFlagUsage(cmd, ctx, "use-internal-ip", "help.target_account_create_object_openstack.flag.use-internal-ip")
	overrideFlagUsage(cmd, ctx, "boot-loader-image-id", "help.target_account_create_object_openstack.flag.boot-loader-image-id")
	overrideFlagUsage(cmd, ctx, "boot-loader-image-name", "help.target_account_create_object_openstack.flag.boot-loader-image-name")
	overrideFlagUsage(cmd, ctx, "linux-boot-image-id", "help.target_account_create_object_openstack.flag.linux-boot-image-id")
	overrideFlagUsage(cmd, ctx, "windows-boot-image-id", "help.target_account_create_object_openstack.flag.windows-boot-image-id")
	overrideFlagUsage(cmd, ctx, "custom-name", "help.target_account_create_object_openstack.flag.custom-name")
	overrideFlagUsage(cmd, ctx, "disk-bus-type-id", "help.target_account_create_object_openstack.flag.disk-bus-type-id")
	overrideFlagUsage(cmd, ctx, "disk-bus-type-name", "help.target_account_create_object_openstack.flag.disk-bus-type-name")
	addFlagString(cmd, ctx, "file")
	cmd.Flags().StringArray("set", nil, ctx.loc.T("flag.set"))
	cmd.Flags().StringArray("set-json", nil, ctx.loc.T("flag.set-json"))
	addFlagBool(cmd, ctx, "preview-request")
}

func addCloudAccountCreateGenericBlockFlags(cmd *cobra.Command, ctx *context, includeCloudType bool) {
	if includeCloudType {
		addFlagString(cmd, ctx, "cloud-type")
	}
	addFlagString(cmd, ctx, "cloud-auth-type")
	addFlagString(cmd, ctx, "account-name")
	addFlagString(cmd, ctx, "file")
	cmd.Flags().StringArray("set", nil, ctx.loc.T("flag.set"))
	cmd.Flags().StringArray("set-json", nil, ctx.loc.T("flag.set-json"))
	addFlagBool(cmd, ctx, "preview-request")
}

func addCloudAccountCreateGenericOSSFlags(cmd *cobra.Command, ctx *context, includeCloudType bool) {
	if includeCloudType {
		addFlagString(cmd, ctx, "cloud-type")
	}
	addFlagString(cmd, ctx, "cloud-auth-type")
	for _, name := range []string{"custom-name", "file"} {
		addFlagString(cmd, ctx, name)
	}
	cmd.Flags().StringArray("set", nil, ctx.loc.T("flag.set"))
	cmd.Flags().StringArray("set-json", nil, ctx.loc.T("flag.set-json"))
	addFlagBool(cmd, ctx, "preview-request")
}

func overrideFlagUsage(cmd *cobra.Command, ctx *context, name, key string) {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return
	}
	flag.Usage = ctx.loc.T(key)
}

func cloudAccountCreateProviderTextKeys(provider, storageType string, specialized bool) (short, longText, exampleKey, notesKey, workflowKey, minimumKey, optionalKey, relatedKey, nextKey string) {
	switch {
	case storageType == "block" && provider == "aliyun":
		return "cmd.cloud_accounts.create.block.aliyun.short",
			"cmd.cloud_accounts.create.block.aliyun.long",
			"cmd.cloud_accounts.create.block.aliyun.examples",
			"cmd.cloud_accounts.create.block.aliyun.notes",
			"cmd.cloud_accounts.create.block.aliyun.workflow",
			"cmd.cloud_accounts.create.block.aliyun.minimum_flags",
			"cmd.cloud_accounts.create.block.aliyun.common_optional_flags",
			"cmd.cloud_accounts.create.block.aliyun.related",
			"cmd.cloud_accounts.create.block.aliyun.next_steps"
	case storageType == "block" && provider == "openstack":
		return "cmd.cloud_accounts.create.block.openstack.short",
			"cmd.cloud_accounts.create.block.openstack.long",
			"cmd.cloud_accounts.create.block.openstack.examples",
			"cmd.cloud_accounts.create.block.openstack.notes",
			"cmd.cloud_accounts.create.block.openstack.workflow",
			"cmd.cloud_accounts.create.block.openstack.minimum_flags",
			"cmd.cloud_accounts.create.block.openstack.common_optional_flags",
			"cmd.cloud_accounts.create.block.openstack.related",
			"cmd.cloud_accounts.create.block.openstack.next_steps"
	case storageType == "objectstorage" && provider == "aliyun":
		return "cmd.cloud_accounts.create.object.aliyun.short",
			"cmd.cloud_accounts.create.object.aliyun.long",
			"cmd.cloud_accounts.create.object.aliyun.examples",
			"cmd.cloud_accounts.create.object.aliyun.notes",
			"cmd.cloud_accounts.create.object.aliyun.workflow",
			"cmd.cloud_accounts.create.object.aliyun.minimum_flags",
			"cmd.cloud_accounts.create.object.aliyun.common_optional_flags",
			"cmd.cloud_accounts.create.object.aliyun.related",
			"cmd.cloud_accounts.create.object.aliyun.next_steps"
	case storageType == "objectstorage" && provider == "openstack":
		return "cmd.cloud_accounts.create.object.openstack.short",
			"cmd.cloud_accounts.create.object.openstack.long",
			"cmd.cloud_accounts.create.object.openstack.examples",
			"cmd.cloud_accounts.create.object.openstack.notes",
			"cmd.cloud_accounts.create.object.openstack.workflow",
			"cmd.cloud_accounts.create.object.openstack.minimum_flags",
			"cmd.cloud_accounts.create.object.openstack.common_optional_flags",
			"cmd.cloud_accounts.create.object.openstack.related",
			"cmd.cloud_accounts.create.object.openstack.next_steps"
	default:
		if specialized {
			return "", "", "", "", "", "", "", "", ""
		}
		return "cmd.cloud_accounts.create.provider.short",
			providerLongKey(storageType),
			"",
			"cmd.cloud_accounts.create.provider.notes",
			"",
			"",
			"",
			"",
			""
	}
}

func providerExample(ctx *context, entry catalog.CloudEntry, storageType string, specialized bool, exampleKey string) string {
	if exampleKey != "" {
		return ctx.loc.T(exampleKey)
	}

	if storageType == "block" {
		return fmt.Sprintf(ctx.loc.T("cmd.cloud_accounts.create.provider.block.example"), entry.Provider)
	}
	return fmt.Sprintf(ctx.loc.T("cmd.cloud_accounts.create.provider.oss.example"), entry.Provider)
}

func providerLongKey(storageType string) string {
	if storageType == "block" {
		return "cmd.cloud_accounts.create.provider.block.long"
	}
	return "cmd.cloud_accounts.create.provider.oss.long"
}

func providerCommandName(storageType, provider string) string {
	if storageType == "block" {
		return "target account create-block " + provider
	}
	return "target account create-oss " + provider
}

func fetchResourcesStorageType(storageType string) string {
	if storageType == "block" {
		return "HyperGate"
	}
	return "objectstorage"
}

func fetchResourcesProviderCommandName(storageType, provider string) string {
	if storageType == "block" {
		return "target account fetch-block-resources " + provider
	}
	return "target account fetch-oss-resources " + provider
}

func cloudAccountFetchProviderTextKeys(storageType string) (shortKey, longKey, usageLineKey, usageNotesKey, exampleKey string) {
	if storageType == "block" {
		return "cmd.cloud_accounts.fetch.provider.block.short",
			"cmd.cloud_accounts.fetch.provider.block.long",
			"cmd.target.account.fetch_block_resources.provider.usage_line",
			"cmd.target.account.fetch_block_resources.provider.usage_notes",
			"cmd.cloud_accounts.fetch.provider.block.examples"
	}
	return "cmd.cloud_accounts.fetch.provider.oss.short",
		"cmd.cloud_accounts.fetch.provider.oss.long",
		"cmd.target.account.fetch_oss_resources.provider.usage_line",
		"cmd.target.account.fetch_oss_resources.provider.usage_notes",
		"cmd.cloud_accounts.fetch.provider.oss.examples"
}

func addAnnotationValue(cmd *cobra.Command, key, value string) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[key] = value
}

func cloudAccountCreateProviderUsageLineKey(storageType string) string {
	if storageType == "block" {
		return "cmd.target.account.create_block.provider.usage_line"
	}
	return "cmd.target.account.create_oss.provider.usage_line"
}

func cloudAccountCreateProviderUsageNotes(ctx *context, entry catalog.CloudEntry, storageType string, specialized bool) string {
	switch {
	case storageType == "block" && entry.Provider == "aliyun" && specialized:
		return ctx.loc.T("cmd.target.account.create_block.aliyun.usage_notes")
	case storageType == "block" && entry.Provider == "openstack" && specialized:
		return ctx.loc.T("cmd.target.account.create_block.openstack.usage_notes")
	case storageType == "objectstorage" && entry.Provider == "aliyun" && specialized:
		return ctx.loc.T("cmd.target.account.create_oss.aliyun.usage_notes")
	case storageType == "objectstorage" && entry.Provider == "openstack" && specialized:
		return ctx.loc.T("cmd.target.account.create_oss.openstack.usage_notes")
	case storageType == "block":
		return fmt.Sprintf(ctx.loc.T("cmd.target.account.create_block.provider.usage_notes"), localizedCloudEntryName(ctx, entry), entry.Provider, entry.CloudType, entry.Provider, entry.Provider)
	default:
		return fmt.Sprintf(ctx.loc.T("cmd.target.account.create_oss.provider.usage_notes"), localizedCloudEntryName(ctx, entry), entry.Provider, entry.CloudType, entry.Provider, entry.Provider)
	}
}
