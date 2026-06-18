package commands

import (
	"fmt"
	"net/url"
	"strings"

	"hyperbdr-client/catalog"
	apptargetresource "hyperbdr-client/internal/app/targetresource"

	"github.com/spf13/cobra"
)

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
	return writeTargetResourceResponse(ctx, result)
}

func runTargetResourceFetch(ctx *context, args []string) error {
	spec, err := parseTargetResourceFetchArgs(args)
	if err != nil {
		return err
	}
	service := apptargetresource.NewService(commandAPIAdapter{ctx: ctx})
	result, err := service.Fetch(spec)
	if err != nil {
		return err
	}
	return writeTargetResourceResponse(ctx, result)
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

func parseTargetResourceFetchArgs(args []string) (apptargetresource.AccountFetchSpec, error) {
	spec := apptargetresource.AccountFetchSpec{Query: url.Values{}}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return apptargetresource.AccountFetchSpec{}, fmt.Errorf("unexpected argument %q", arg)
		}
		name, value, hasInline := splitFlag(arg)
		name = strings.TrimPrefix(name, "--")
		switch name {
		case "cloud-account-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.AccountFetchSpec{}, err
			}
			spec.CloudAccountID = v
			i = next
		case "fetch-res":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.AccountFetchSpec{}, err
			}
			spec.FetchRes = v
			i = next
		case "region-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.AccountFetchSpec{}, err
			}
			spec.RegionID = v
			i = next
		case "zone-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.AccountFetchSpec{}, err
			}
			spec.ZoneID = v
			i = next
		case "flavor-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.AccountFetchSpec{}, err
			}
			spec.FlavorID = v
			i = next
		default:
			v, next, err := targetResourceOptionalFlagValue(args, i, value, hasInline)
			if err != nil {
				return apptargetresource.AccountFetchSpec{}, err
			}
			spec.Query.Set(strings.ReplaceAll(name, "-", "_"), v)
			i = next
		}
	}
	return spec, nil
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
