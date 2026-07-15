package commands

import (
	"fmt"
	"net/url"
	"strings"

	"hyperbdr-client/catalog"
	apptargetresource "hyperbdr-client/internal/app/targetresource"
	"hyperbdr-client/internal/config"

	"github.com/spf13/cobra"
)

const cloudResourceFetchHelpProfileAnnotation = "cloud-resource-fetch-help-profile"

type cloudResourceFetchSelection struct {
	CloudAccountID    string
	Provider          string
	PublicStorageType string
	BackendStorage    string
	DirectFlagSet     bool
	directArgs        []string
	positionalArgs    []string
}

type cloudResourceFetchProfile struct {
	Entry       catalog.CloudEntry
	Provider    string
	StorageType string
	Kind        string
}

func newCloudResourceCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "cloud-resource", "cmd.cloud_resource.short", "cmd.cloud_resource.long", "cmd.cloud_resource.examples", "cmd.cloud_resource.notes", "cloud-resource")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.cloud_resource.usage_line")
	addUsageNotes(cmd, ctx, "cmd.cloud_resource.usage_notes")
	cmd.AddCommand(newCloudResourceFetchCommand(ctx))
	return cmd
}

func newCloudResourceFetchCommand(ctx *context) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "fetch",
		Short:              ctx.loc.T("cmd.cloud_resource.fetch.short"),
		Long:               ctx.loc.T("cmd.cloud_resource.fetch.long"),
		Example:            strings.TrimSpace(ctx.loc.T("cmd.cloud_resource.fetch.examples")),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			selection, err := parseCloudResourceFetchSelection(args)
			if err != nil {
				return err
			}
			if rawArgsHelp(cmd, args) {
				return renderCloudResourceFetchHelp(ctx, cmd, selection)
			}
			return runCloudResourceFetchBySelection(ctx, selection, args)
		},
	}
	addCloudResourceFetchAllFlags(cmd, ctx)
	addHelpLayout(cmd, helpLayoutFourSection)
	addUsageLine(cmd, ctx, "cmd.cloud_resource.fetch.usage_line")
	addUsageNotes(cmd, ctx, "cmd.cloud_resource.fetch.usage_notes")
	return cmd
}

func addCloudResourceFetchAllFlags(cmd *cobra.Command, ctx *context) {
	for _, name := range []string{
		"cloud-account-id", "cloud-type", "storage-type", "cloud-auth-type",
		"fetch-res", "region-id", "zone-id", "flavor-id", "flavor-vcpus", "flavor-ram", "boot-mode",
		"access-key-id", "access-key-secret", "access-id", "access-secret",
		"auth-url", "username", "password", "user-domain-id",
		"project-id", "project-domain-id", "project-name", "compute-zone-id", "block-store-zone-id",
	} {
		addFlagString(cmd, ctx, name)
	}
}

func newTargetResourceCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "resource", "cmd.target.resource.short", "cmd.target.resource.long", "cmd.target.resource.examples", "cmd.target.resource.notes", "target resource")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.target.resource.usage_line")
	addUsageNotes(cmd, ctx, "cmd.target.resource.usage_notes")

	cmd.AddCommand(
		newTargetResourceProviderGroupCommand(ctx, "block"),
		newTargetResourceProviderGroupCommand(ctx, "oss"),
		newTargetResourceFetchCommand(ctx),
	)
	return cmd
}

func newTargetResourceProviderGroupCommand(ctx *context, kind string) *cobra.Command {
	shortKey, longKey, exampleKey, notesKey, usageLineKey, usageNotesKey := targetResourceProviderGroupTextKeys(kind)
	groupName := "target resource " + kind
	cmd := newGroupCommand(ctx, kind, shortKey, longKey, exampleKey, notesKey, groupName)
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, usageLineKey)
	addUsageNotes(cmd, ctx, usageNotesKey)

	var clouds []catalog.CloudEntry
	if kind == "block" {
		clouds = catalog.EnabledBlockClouds()
	} else {
		clouds = catalog.EnabledObjectClouds()
	}
	for _, entry := range clouds {
		cmd.AddCommand(newTargetResourceProviderCommand(ctx, kind, entry))
	}
	return cmd
}

func newTargetResourceProviderCommand(ctx *context, kind string, entry catalog.CloudEntry) *cobra.Command {
	shortKey, longKey, usageLineKey, usageNotesKey, exampleKey := targetResourceProviderTextKeys(kind, entry.Provider)
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
			return runTargetResourceDirectAuth(ctx, targetResourceProviderCommandName(kind, entry.Provider), entry.CloudType, targetResourceStorageType(kind), args)
		},
	}

	if entry.Provider == "openstack" {
		for _, name := range []string{"auth-url", "username", "password", "user-domain-id", "fetch-res", "region-id", "zone-id", "project-id", "project-domain-id", "project-name", "compute-zone-id", "block-store-zone-id"} {
			addFlagString(cmd, ctx, name)
		}
	} else {
		for _, name := range []string{"cloud-auth-type", "fetch-res", "region-id", "zone-id", "boot-mode"} {
			addFlagString(cmd, ctx, name)
		}
	}

	addHelpLayout(cmd, helpLayoutFourSection)
	addAnnotationValue(cmd, usageLineAnnotation, fmt.Sprintf(ctx.loc.T(usageLineKey), entry.Provider))
	if entry.Provider == "openstack" {
		addAnnotationValue(cmd, usageNotesAnnotation, ctx.loc.T(usageNotesKey))
	} else {
		addAnnotationValue(cmd, usageNotesAnnotation, fmt.Sprintf(ctx.loc.T(usageNotesKey), displayName, entry.Provider, entry.Provider, entry.Provider))
	}
	return cmd
}

func newTargetResourceFetchCommand(ctx *context) *cobra.Command {
	cmd := newRawLeafCommand(ctx, "fetch", "cmd.target.resource.fetch.short", "cmd.target.resource.fetch.long", "cmd.target.resource.fetch.examples", "cmd.target.resource.fetch.notes", func(cmd *cobra.Command) {
		for _, name := range []string{"cloud-account-id", "fetch-res", "region-id", "zone-id", "flavor-id"} {
			addFlagString(cmd, ctx, name)
		}
	}, func(args []string) error {
		return runTargetResourceFetch(ctx, args)
	})
	addHelpLayout(cmd, helpLayoutFourSection)
	addUsageLine(cmd, ctx, "cmd.target.resource.fetch.usage_line")
	addUsageNotes(cmd, ctx, "cmd.target.resource.fetch.usage_notes")
	return cmd
}

func runTargetResourceDirectAuth(ctx *context, commandName, cloudType, storageType string, args []string) error {
	spec, err := parseTargetResourceDirectAuthArgs(commandName, cloudType, storageType, args)
	if err != nil {
		return err
	}
	service := apptargetresource.NewService(commandAPIAdapter{ctx: ctx})
	result, err := service.DirectAuth(spec)
	if err != nil {
		return err
	}
	return writeTargetResourceResponse(ctx, result, nil)
}

func runTargetResourceFetch(ctx *context, args []string) error {
	parsed, err := parseTargetResourceFetchArgs(args)
	if err != nil {
		return err
	}
	service := apptargetresource.NewService(commandAPIAdapter{ctx: ctx})
	result, err := service.Fetch(parsed.spec)
	if err != nil {
		return err
	}
	return writeTargetResourceResponse(ctx, result, parsed.meta)
}

func runCloudResourceFetchBySelection(ctx *context, selection cloudResourceFetchSelection, args []string) error {
	if len(selection.positionalArgs) > 0 {
		return fmt.Errorf("unexpected argument %q", selection.positionalArgs[0])
	}
	if selection.CloudAccountID != "" {
		if selection.DirectFlagSet {
			return fmt.Errorf("cloud-account-id cannot be used with cloud-type, storage-type, or credential flags")
		}
		return runTargetResourceFetch(ctx, args)
	}

	profile, err := resolveCloudResourceDirectProfile(selection)
	if err != nil {
		return err
	}
	return runTargetResourceDirectAuth(ctx, "cloud-resource fetch", profile.Entry.CloudType, profile.StorageType, selection.directArgs)
}

func parseCloudResourceFetchSelection(args []string) (cloudResourceFetchSelection, error) {
	selection := cloudResourceFetchSelection{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--help" || arg == "-h" {
			continue
		}
		if !strings.HasPrefix(arg, "--") {
			selection.positionalArgs = append(selection.positionalArgs, arg)
			continue
		}
		name, value, hasInline := splitFlag(arg)
		name = strings.TrimPrefix(name, "--")
		switch name {
		case "cloud-account-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return selection, err
			}
			selection.CloudAccountID = v
			i = next
		case "cloud-type":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return selection, err
			}
			selection.Provider = v
			selection.DirectFlagSet = true
			i = next
		case "storage-type":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return selection, err
			}
			public, backend, err := normalizeCloudResourcePublicStorageType(v)
			if err != nil {
				return selection, err
			}
			selection.PublicStorageType = public
			selection.BackendStorage = backend
			selection.DirectFlagSet = true
			i = next
		default:
			if isCloudResourceCredentialFlag(name) {
				selection.DirectFlagSet = true
			}
			selection.directArgs = append(selection.directArgs, arg)
			if !hasInline && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				selection.directArgs = append(selection.directArgs, args[i+1])
				i++
			}
		}
	}
	return selection, nil
}

func isCloudResourceCredentialFlag(name string) bool {
	switch name {
	case "cloud-auth-type", "access-key-id", "access-key-secret", "access-id", "access-secret", "auth-url", "username", "password", "user-domain-id":
		return true
	default:
		return false
	}
}

func normalizeCloudResourcePublicStorageType(value string) (public, backend string, err error) {
	switch strings.TrimSpace(value) {
	case "block":
		return "block", "HyperGate", nil
	case "object":
		return "object", "objectstorage", nil
	case "":
		return "", "", nil
	default:
		return "", "", fmt.Errorf("storage-type must be block or object")
	}
}

func resolveCloudResourceDirectProfile(selection cloudResourceFetchSelection) (cloudResourceFetchProfile, error) {
	if strings.TrimSpace(selection.PublicStorageType) == "" {
		return cloudResourceFetchProfile{}, fmt.Errorf("storage-type is required")
	}
	if strings.TrimSpace(selection.Provider) == "" {
		return cloudResourceFetchProfile{}, fmt.Errorf("cloud-type is required")
	}

	entry, ok := findCloudResourceProvider(selection.Provider, selection.BackendStorage)
	if !ok {
		return cloudResourceFetchProfile{}, fmt.Errorf("cloud-type %q does not support storage-type %s", selection.Provider, selection.PublicStorageType)
	}
	return cloudResourceFetchProfile{
		Entry:       entry,
		Provider:    entry.Provider,
		StorageType: selection.BackendStorage,
		Kind:        cloudResourceStorageKind(selection.BackendStorage),
	}, nil
}

func findCloudResourceProvider(provider, storageType string) (catalog.CloudEntry, bool) {
	var clouds []catalog.CloudEntry
	if cloudResourceStorageKind(storageType) == "block" {
		clouds = catalog.EnabledBlockClouds()
	} else {
		clouds = catalog.EnabledObjectClouds()
	}
	needle := strings.ToLower(strings.TrimSpace(provider))
	for _, entry := range clouds {
		if strings.ToLower(strings.TrimSpace(entry.Provider)) == needle {
			return entry, true
		}
	}
	return catalog.CloudEntry{}, false
}

func findCloudResourceBackendProfile(cloudType, storageType string) (cloudResourceFetchProfile, bool) {
	var entry catalog.CloudEntry
	var ok bool
	if cloudResourceStorageKind(storageType) == "block" {
		entry, ok = catalog.FindBlockCloud(cloudType)
	} else {
		entry, ok = catalog.FindObjectCloud(cloudType)
	}
	if !ok {
		return cloudResourceFetchProfile{}, false
	}
	return cloudResourceFetchProfile{
		Entry:       entry,
		Provider:    entry.Provider,
		StorageType: storageType,
		Kind:        cloudResourceStorageKind(storageType),
	}, true
}

func cloudResourceStorageKind(storageType string) string {
	switch strings.ToLower(strings.TrimSpace(storageType)) {
	case "hypergate", "blockstorage", "block", "block_storage":
		return "block"
	default:
		return "objectstorage"
	}
}

func renderCloudResourceFetchHelp(ctx *context, cmd *cobra.Command, selection cloudResourceFetchSelection) error {
	profile := "generic"
	switch {
	case selection.CloudAccountID != "":
		if selection.DirectFlagSet {
			return fmt.Errorf("cloud-account-id cannot be used with cloud-type, storage-type, or credential flags")
		}
		cfg, err := config.Resolve(ctx.flags)
		if err != nil {
			return err
		}
		ctx.cfg = cfg
		accountCtx, err := apptargetresource.NewService(commandAPIAdapter{ctx: ctx}).CloudAccountContext(selection.CloudAccountID)
		if err != nil {
			return err
		}
		if strings.TrimSpace(accountCtx.CloudType) == "" {
			return fmt.Errorf("cloud-type cannot be inferred from cloud-account-id")
		}
		if strings.TrimSpace(accountCtx.StorageType) == "" {
			return fmt.Errorf("storage-type cannot be inferred from cloud-account-id")
		}
		backendProfile, _ := findCloudResourceBackendProfile(accountCtx.CloudType, accountCtx.StorageType)
		profile = "account|" + cloudResourceStorageKind(accountCtx.StorageType)
		if backendProfile.Provider == "openstack" {
			profile = "account|openstack"
		}
		addUsageLine(cmd, ctx, "cmd.cloud_resource.fetch.account.usage_line")
		addAnnotationValue(cmd, usageNotesAnnotation, cloudResourceAccountUsageNotes(ctx, accountCtx, backendProfile))
		addHelpDescription(cmd, ctx, "cmd.cloud_resource.fetch.account.short")
	case selection.PublicStorageType != "" && selection.Provider == "":
		profile = selection.PublicStorageType
		addUsageLine(cmd, ctx, cloudResourceStorageUsageLineKey(selection.PublicStorageType))
		addAnnotationValue(cmd, usageNotesAnnotation, cloudResourceStorageUsageNotes(ctx, selection.BackendStorage))
		addHelpDescription(cmd, ctx, cloudResourceStorageShortKey(selection.PublicStorageType))
	case selection.PublicStorageType != "" && selection.Provider != "":
		resolved, err := resolveCloudResourceDirectProfile(selection)
		if err != nil {
			return err
		}
		profile = "direct|" + resolved.Kind
		if resolved.Provider == "openstack" {
			profile = "direct|openstack"
		}
		applyCloudResourceDirectHelp(ctx, cmd, resolved)
	}
	addAnnotationValue(cmd, cloudResourceFetchHelpProfileAnnotation, profile)
	return renderHelp(cmd, ctx)
}

func cloudResourceStorageUsageLineKey(publicStorageType string) string {
	if publicStorageType == "block" {
		return "cmd.cloud_resource.fetch.block_storage.usage_line"
	}
	return "cmd.cloud_resource.fetch.object_storage.usage_line"
}

func cloudResourceStorageShortKey(publicStorageType string) string {
	if publicStorageType == "block" {
		return "cmd.cloud_resource.fetch.block_storage.short"
	}
	return "cmd.cloud_resource.fetch.object_storage.short"
}

func cloudResourceStorageUsageNotes(ctx *context, storageType string) string {
	var notes string
	var providers []catalog.CloudEntry
	if cloudResourceStorageKind(storageType) == "block" {
		notes = ctx.loc.T("cmd.cloud_resource.fetch.block_storage.usage_notes")
		providers = catalog.EnabledBlockClouds()
	} else {
		notes = ctx.loc.T("cmd.cloud_resource.fetch.object_storage.usage_notes")
		providers = catalog.EnabledObjectClouds()
	}
	if len(providers) == 0 {
		return notes
	}
	label := "Providers:"
	if ctx.loc.Lang() == "zh_cn" {
		label = "云厂商:"
	}
	var b strings.Builder
	b.WriteString(strings.TrimSpace(notes))
	b.WriteString("\n\n")
	b.WriteString(label)
	for _, entry := range providers {
		b.WriteString("\n  ")
		b.WriteString(entry.Provider)
	}
	return b.String()
}

func applyCloudResourceDirectHelp(ctx *context, cmd *cobra.Command, profile cloudResourceFetchProfile) {
	addAnnotationValue(cmd, usageLineAnnotation, fmt.Sprintf(ctx.loc.T(cloudResourceDirectUsageLineKey(profile.Kind)), profile.Provider))
	addAnnotationValue(cmd, usageNotesAnnotation, cloudResourceDirectUsageNotes(ctx, profile))
	addHelpDescription(cmd, ctx, cloudResourceDirectShortKey(profile.Kind))
}

func cloudResourceDirectUsageNotes(ctx *context, profile cloudResourceFetchProfile) string {
	if profile.Provider == "aliyun" || profile.Provider == "huawei" {
		notes := ctx.loc.T(fmt.Sprintf("help.cloud_resource.fetch.%s.%s", profile.Provider, cloudResourcePublicStorageType(profile.Kind)))
		return fmt.Sprintf(notes, ctx.loc.T("help.cloud_resource.provider."+profile.Provider), profile.Provider)
	}
	if profile.Provider == "openstack" {
		return ctx.loc.T(fmt.Sprintf("help.cloud_resource.fetch.openstack.%s", cloudResourcePublicStorageType(profile.Kind)))
	}
	return fmt.Sprintf(ctx.loc.T("cmd.cloud_resource.fetch.provider.usage_notes"), localizedCloudEntryName(ctx, profile.Entry), profile.Provider, cloudResourcePublicStorageType(profile.Kind), profile.Provider, cloudResourcePublicStorageType(profile.Kind))
}

func cloudResourceDirectUsageLineKey(kind string) string {
	if kind == "block" {
		return "cmd.cloud_resource.fetch.provider.block.usage_line"
	}
	return "cmd.cloud_resource.fetch.provider.object.usage_line"
}

func cloudResourceDirectShortKey(kind string) string {
	if kind == "block" {
		return "cmd.cloud_resource.fetch.provider.block.short"
	}
	return "cmd.cloud_resource.fetch.provider.object.short"
}

func cloudResourcePublicStorageType(kind string) string {
	if kind == "block" {
		return "block"
	}
	return "object"
}

func cloudResourceAccountUsageNotes(ctx *context, accountCtx apptargetresource.CloudAccountContext, profile cloudResourceFetchProfile) string {
	provider := profile.Provider
	if provider == "" {
		provider = accountCtx.CloudType
	}
	return fmt.Sprintf(ctx.loc.T("cmd.cloud_resource.fetch.account.usage_notes"), accountCtx.CloudAccountID, accountCtx.CloudType, accountCtx.StorageType, provider)
}

func parseTargetResourceDirectAuthArgs(commandName, cloudType, storageType string, args []string) (apptargetresource.DirectAuthSpec, error) {
	spec := apptargetresource.DirectAuthSpec{
		CloudType:     cloudType,
		StorageType:   storageType,
		DynamicFields: map[string]string{},
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return apptargetresource.DirectAuthSpec{}, fmt.Errorf("unexpected argument %q", arg)
		}
		name, value, hasInline := splitFlag(arg)
		name = strings.TrimPrefix(name, "--")
		switch name {
		case "cloud-account-id":
			return apptargetresource.DirectAuthSpec{}, fmt.Errorf("cloud-account-id cannot be used with %s", commandName)
		case "cloud-type":
			return apptargetresource.DirectAuthSpec{}, fmt.Errorf("cloud-type cannot be used with %s", commandName)
		case "storage-type":
			return apptargetresource.DirectAuthSpec{}, fmt.Errorf("storage-type cannot be used with %s", commandName)
		case "cloud-auth-type":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.DirectAuthSpec{}, err
			}
			spec.CloudAuthType = v
			i = next
		case "fetch-res":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.DirectAuthSpec{}, err
			}
			spec.FetchRes = v
			i = next
		case "region-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.DirectAuthSpec{}, err
			}
			spec.RegionID = v
			i = next
		case "zone-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.DirectAuthSpec{}, err
			}
			spec.ZoneID = v
			i = next
		case "boot-mode":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.DirectAuthSpec{}, err
			}
			spec.BootMode = v
			i = next
		default:
			v, next, err := targetResourceOptionalFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.DirectAuthSpec{}, err
			}
			spec.DynamicFields[strings.ReplaceAll(name, "-", "_")] = v
			i = next
		}
	}
	return spec, nil
}

type parsedTargetResourceFetchCommand struct {
	spec apptargetresource.AccountFetchSpec
	meta map[string]interface{}
}

func parseTargetResourceFetchArgs(args []string) (parsedTargetResourceFetchCommand, error) {
	spec := apptargetresource.AccountFetchSpec{Query: url.Values{}}
	meta := map[string]interface{}{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return parsedTargetResourceFetchCommand{}, fmt.Errorf("unexpected argument %q", arg)
		}
		name, value, hasInline := splitFlag(arg)
		name = strings.TrimPrefix(name, "--")
		switch name {
		case "cloud-account-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedTargetResourceFetchCommand{}, err
			}
			spec.CloudAccountID = v
			i = next
		case "fetch-res":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedTargetResourceFetchCommand{}, err
			}
			spec.FetchRes = v
			i = next
		case "region-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedTargetResourceFetchCommand{}, err
			}
			spec.RegionID = v
			i = next
		case "zone-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedTargetResourceFetchCommand{}, err
			}
			spec.ZoneID = v
			i = next
		case "flavor-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedTargetResourceFetchCommand{}, err
			}
			spec.FlavorID = v
			i = next
		case "flavor-vcpus", "flavor-ram":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedTargetResourceFetchCommand{}, err
			}
			meta[strings.ReplaceAll(name, "-", "_")] = v
			i = next
		default:
			v, next, err := targetResourceOptionalFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedTargetResourceFetchCommand{}, err
			}
			spec.Query.Set(strings.ReplaceAll(name, "-", "_"), v)
			i = next
		}
	}
	return parsedTargetResourceFetchCommand{spec: spec, meta: meta}, nil
}

func targetResourceOptionalFlagValue(args []string, idx int, inline string, hasInline bool) (string, int, error) {
	if hasInline {
		return inline, idx, nil
	}
	if idx+1 < len(args) && !strings.HasPrefix(args[idx+1], "--") {
		return args[idx+1], idx + 1, nil
	}
	return "true", idx, nil
}

func targetResourceStorageType(kind string) string {
	if kind == "block" {
		return "HyperGate"
	}
	return "objectstorage"
}

func targetResourceProviderCommandName(kind, provider string) string {
	return "target resource " + kind + " " + provider
}

func targetResourceProviderGroupTextKeys(kind string) (shortKey, longKey, exampleKey, notesKey, usageLineKey, usageNotesKey string) {
	if kind == "block" {
		return "cmd.target.resource.block.short",
			"cmd.target.resource.block.long",
			"cmd.target.resource.block.examples",
			"cmd.target.resource.block.notes",
			"cmd.target.resource.block.usage_line",
			"cmd.target.resource.block.usage_notes"
	}
	return "cmd.target.resource.oss.short",
		"cmd.target.resource.oss.long",
		"cmd.target.resource.oss.examples",
		"cmd.target.resource.oss.notes",
		"cmd.target.resource.oss.usage_line",
		"cmd.target.resource.oss.usage_notes"
}

func targetResourceProviderTextKeys(kind, provider string) (shortKey, longKey, usageLineKey, usageNotesKey, exampleKey string) {
	if kind == "block" && provider == "openstack" {
		return "cmd.target.resource.provider.block.short",
			"cmd.target.resource.provider.block.long",
			"cmd.target.resource.block.openstack.usage_line",
			"cmd.target.resource.block.openstack.usage_notes",
			"cmd.target.resource.block.openstack.examples"
	}
	if kind == "oss" && provider == "openstack" {
		return "cmd.target.resource.provider.oss.short",
			"cmd.target.resource.provider.oss.long",
			"cmd.target.resource.oss.openstack.usage_line",
			"cmd.target.resource.oss.openstack.usage_notes",
			"cmd.target.resource.oss.openstack.examples"
	}
	if kind == "block" {
		return "cmd.target.resource.provider.block.short",
			"cmd.target.resource.provider.block.long",
			"cmd.target.resource.block.provider.usage_line",
			"cmd.target.resource.block.provider.usage_notes",
			"cmd.target.resource.provider.block.examples"
	}
	return "cmd.target.resource.provider.oss.short",
		"cmd.target.resource.provider.oss.long",
		"cmd.target.resource.oss.provider.usage_line",
		"cmd.target.resource.oss.provider.usage_notes",
		"cmd.target.resource.provider.oss.examples"
}
