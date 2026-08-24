package commands

import (
	"fmt"
	"strings"

	"hyperbdr-client/catalog"
	workflowcreate "hyperbdr-client/internal/workflow/cloudaccountcreate"

	"github.com/spf13/cobra"
)

const cloudAccountCreateHelpProfileAnnotation = "cloud-account-create-help-profile"

type cloudAccountCreateSelection struct {
	Provider          string
	PublicStorageType string
	BackendStorage    string
}

type cloudAccountCreateProfile struct {
	Entry       catalog.CloudEntry
	Provider    string
	StorageType string
	Specialized bool
}

func addCloudAccountCreateAllFlags(cmd *cobra.Command, ctx *context) {
	for _, name := range []string{
		"cloud-type", "storage-type", "file", "body", "cloud-auth-type",
		"access-key-id", "access-key-secret", "region-id", "region-name", "account-name", "auth-region-id",
		"auth-url", "username", "password", "user-domain-id", "project-domain-id", "project-id", "project-name",
		"use-internal-ip", "boot-loader-image-id", "boot-loader-image-name", "boot-loader-flavor-id",
		"linux-boot-image-id", "windows-boot-image-id", "linux-uefi-boot-image-id", "windows-uefi-boot-image-id",
		"custom-name", "disk-bus-type-id", "disk-bus-type-name",
		"ssh-port", "ssh-pass", "linux-hd-username", "linux-hd-password", "linux-hd-port",
	} {
		addFlagString(cmd, ctx, name)
	}
	cmd.Flags().StringArray("set", nil, ctx.loc.T("flag.set"))
	cmd.Flags().StringArray("set-json", nil, ctx.loc.T("flag.set-json"))
	addFlagBool(cmd, ctx, "preview-request")
}

func parseCloudAccountCreateSelection(args []string) (cloudAccountCreateSelection, []string, error) {
	selection := cloudAccountCreateSelection{}
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--help" || arg == "-h" {
			remaining = append(remaining, arg)
			continue
		}
		if !strings.HasPrefix(arg, "--") {
			remaining = append(remaining, arg)
			continue
		}
		name, value, hasInline := splitFlag(arg)
		name = strings.TrimPrefix(name, "--")
		switch name {
		case "cloud-type":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return selection, nil, err
			}
			selection.Provider = v
			i = next
		case "storage-type":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return selection, nil, err
			}
			public, backend, err := normalizeCloudAccountCreatePublicStorageType(v)
			if err != nil {
				return selection, nil, err
			}
			selection.PublicStorageType = public
			selection.BackendStorage = backend
			i = next
		default:
			remaining = append(remaining, arg)
		}
	}
	return selection, remaining, nil
}

func normalizeCloudAccountCreatePublicStorageType(value string) (public, backend string, err error) {
	switch strings.TrimSpace(value) {
	case "block":
		return "block", "block", nil
	case "object":
		return "object", "objectstorage", nil
	case "":
		return "", "", nil
	default:
		return "", "", fmt.Errorf("storage-type must be block or object")
	}
}

func resolveCloudAccountCreateProfile(selection cloudAccountCreateSelection) (cloudAccountCreateProfile, error) {
	if strings.TrimSpace(selection.PublicStorageType) == "" {
		return cloudAccountCreateProfile{}, fmt.Errorf("storage-type is required")
	}
	if strings.TrimSpace(selection.Provider) == "" {
		return cloudAccountCreateProfile{}, fmt.Errorf("cloud-type is required")
	}
	var entry catalog.CloudEntry
	var ok bool
	switch selection.BackendStorage {
	case "block":
		entry, ok = catalog.FindBlockCloud(selection.Provider)
	case "objectstorage":
		entry, ok = catalog.FindObjectCloud(selection.Provider)
	}
	if !ok || !entry.Enabled {
		return cloudAccountCreateProfile{}, fmt.Errorf("cloud-type %q does not support storage-type %s", selection.Provider, selection.PublicStorageType)
	}
	return cloudAccountCreateProfile{
		Entry:       entry,
		Provider:    entry.Provider,
		StorageType: selection.BackendStorage,
		Specialized: workflowcreate.HasRegisteredAdapter(entry.CloudType, selection.BackendStorage),
	}, nil
}

func renderCloudAccountCreateHelp(ctx *context, cmd *cobra.Command, selection cloudAccountCreateSelection) error {
	profile := "generic"
	switch {
	case selection.PublicStorageType == "block" && selection.Provider == "":
		profile = "block_storage"
		addUsageLine(cmd, ctx, "cmd.cloud_account.create.block_storage.usage_line")
		addAnnotationValue(cmd, usageNotesAnnotation, cloudAccountCreateStorageUsageNotes(ctx, "block"))
		addHelpDescription(cmd, ctx, "cmd.cloud_account.create.block_storage.short")
	case selection.PublicStorageType == "object" && selection.Provider == "":
		profile = "object_storage"
		addUsageLine(cmd, ctx, "cmd.cloud_account.create.object_storage.usage_line")
		addAnnotationValue(cmd, usageNotesAnnotation, cloudAccountCreateStorageUsageNotes(ctx, "objectstorage"))
		addHelpDescription(cmd, ctx, "cmd.cloud_account.create.object_storage.short")
	case selection.PublicStorageType != "" && selection.Provider != "":
		resolved, err := resolveCloudAccountCreateProfile(selection)
		if err != nil {
			return err
		}
		profile = resolved.StorageType + "|" + resolved.Provider
		applyCloudAccountCreateProviderHelp(ctx, cmd, resolved)
	}
	addAnnotationValue(cmd, cloudAccountCreateHelpProfileAnnotation, profile)
	return renderHelp(cmd, ctx)
}

func applyCloudAccountCreateProviderHelp(ctx *context, cmd *cobra.Command, profile cloudAccountCreateProfile) {
	if profile.StorageType == "objectstorage" && (profile.Provider == "aliyun" || profile.Provider == "openstack") {
		overrideFlagUsage(cmd, ctx, "use-internal-ip", "flag.cloud-account.use-internal-ip")
	}
	setDynamicParameterHelpContext(cmd, dynamicParameterHelpContext{
		Command:      dynamicParameterHelpCloudAccountCreate,
		Provider:     profile.Provider,
		CloudType:    profile.Entry.CloudType,
		Architecture: profile.Entry.Architecture,
		StorageType:  profile.StorageType,
	})
	addAnnotationValue(cmd, usageLineAnnotation, fmt.Sprintf(ctx.loc.T(cloudAccountCreateProviderDirectUsageLineKey(profile.StorageType)), profile.Provider))
	addAnnotationValue(cmd, usageNotesAnnotation, cloudAccountCreateProviderDirectUsageNotes(ctx, profile))
	shortKey, longKey, _, _, _, _, _, _, _ := cloudAccountCreateProviderTextKeys(profile.Provider, profile.StorageType, profile.Specialized)
	shortText, longText := providerCommandTexts(ctx, profile.Entry, profile.StorageType, profile.Specialized, shortKey, longKey)
	if profile.Provider == "huawei" {
		shortKey = "cmd.cloud_accounts.create.block.huawei.short"
		if profile.StorageType == "objectstorage" {
			shortKey = "cmd.cloud_accounts.create.object.huawei.short"
		}
		shortText = ctx.loc.T(shortKey)
	}
	cmd.Short = shortText
	cmd.Long = longText
	addAnnotationValue(cmd, helpDescriptionAnnotation, shortText)
}

func runCreateCloudAccountBySelection(ctx *context, selection cloudAccountCreateSelection, args []string) error {
	profile, err := resolveCloudAccountCreateProfile(selection)
	if err != nil {
		return err
	}
	_, remaining, err := parseCloudAccountCreateSelection(args)
	if err != nil {
		return err
	}
	return runCreateCloudAccountForProvider(ctx, providerCommandName(profile.StorageType, profile.Provider), profile.Entry.CloudType, profile.StorageType, profile.Specialized, remaining)
}

func cloudAccountCreateProviderDirectUsageLineKey(storageType string) string {
	if storageType == "block" {
		return "cmd.cloud_account.create.provider.block.usage_line"
	}
	return "cmd.cloud_account.create.provider.object.usage_line"
}

func cloudAccountCreateProviderDirectUsageNotes(ctx *context, profile cloudAccountCreateProfile) string {
	directKey := ""
	switch {
	case profile.StorageType == "block" && profile.Provider == "aliyun":
		directKey = "help.cloud_account.block.aliyun"
	case profile.StorageType == "block" && profile.Provider == "huawei":
		directKey = "help.cloud_account.block.huawei"
	case profile.StorageType == "block" && profile.Provider == "openstack":
		directKey = "help.cloud_account.block.openstack"
	case profile.StorageType == "objectstorage" && profile.Provider == "aliyun":
		directKey = "help.cloud_account.object.aliyun"
	case profile.StorageType == "objectstorage" && profile.Provider == "huawei":
		directKey = "help.cloud_account.object.huawei"
	case profile.StorageType == "objectstorage" && profile.Provider == "openstack":
		directKey = "help.cloud_account.object.openstack"
	}
	if directKey != "" {
		return appendCloudAccountCreateSupplementalHelp(ctx, ctx.loc.T(directKey), profile)
	}
	notes := cloudAccountCreateProviderUsageNotes(ctx, profile.Entry, profile.StorageType, profile.Specialized)
	notes = rewriteCloudAccountCreateDirectCommandRefs(notes, profile.Provider, profile.Entry.CloudType, profile.StorageType)
	notes = strings.ReplaceAll(notes, "target account", "cloud-account")
	return appendCloudAccountCreateSupplementalHelp(ctx, notes, profile)
}

func cloudAccountCreateStorageUsageNotes(ctx *context, storageType string) string {
	var notes string
	var providers []catalog.CloudEntry
	if storageType == "block" {
		notes = ctx.loc.T("cmd.cloud_account.create.block_storage.usage_notes")
		providers = catalog.EnabledBlockClouds()
	} else {
		notes = ctx.loc.T("cmd.cloud_account.create.object_storage.usage_notes")
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

func rewriteCloudAccountCreateDirectCommandRefs(notes, provider, cloudType, storageType string) string {
	publicStorageType := "object"
	if storageType == "block" {
		publicStorageType = "block"
	}
	createPath := fmt.Sprintf("hyperbdrctl cloud-account create --cloud-type %s --storage-type %s", provider, publicStorageType)
	createFragment := fmt.Sprintf("cloud-account create --cloud-type %s --storage-type %s", provider, publicStorageType)
	resourcePath := fmt.Sprintf("hyperbdrctl cloud-resource fetch --cloud-type %s --storage-type %s", provider, publicStorageType)
	resourceFragment := fmt.Sprintf("cloud-resource fetch --cloud-type %s --storage-type %s", provider, publicStorageType)
	replacements := []struct {
		old string
		new string
	}{
		{fmt.Sprintf("hyperbdrctl target account create-block %s", cloudType), createPath},
		{fmt.Sprintf("hyperbdrctl target account create-oss %s", cloudType), createPath},
		{fmt.Sprintf("target account create-block %s", cloudType), createFragment},
		{fmt.Sprintf("target account create-oss %s", cloudType), createFragment},
		{fmt.Sprintf("hyperbdrctl target account create-block %s", provider), createPath},
		{fmt.Sprintf("hyperbdrctl target account create-oss %s", provider), createPath},
		{fmt.Sprintf("target account create-block %s", provider), createFragment},
		{fmt.Sprintf("target account create-oss %s", provider), createFragment},
		{fmt.Sprintf("hyperbdrctl target account fetch-block-resources %s", provider), resourcePath},
		{fmt.Sprintf("hyperbdrctl target account fetch-oss-resources %s", provider), resourcePath},
		{fmt.Sprintf("target account fetch-block-resources %s", provider), resourceFragment},
		{fmt.Sprintf("target account fetch-oss-resources %s", provider), resourceFragment},
	}
	for _, replacement := range replacements {
		notes = strings.ReplaceAll(notes, replacement.old, replacement.new)
	}
	notes = strings.ReplaceAll(notes, fmt.Sprintf("%%!(EXTRA string=%s)", provider), "")
	return notes
}
