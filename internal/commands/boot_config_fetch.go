package commands

import (
	"fmt"
	"strings"

	"hyperbdr-client/catalog"
	appbootconfigquery "hyperbdr-client/internal/app/bootconfigquery"

	"github.com/spf13/cobra"
)

func newBootConfigFetchBlockResourcesCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "fetch-block-resources", "cmd.boot_config_top.fetch.block.short", "cmd.boot_config_top.fetch.block.long", "cmd.boot_config_top.fetch.block.examples", "cmd.boot_config_top.fetch.block.notes", "boot-config fetch-block-resources")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.boot_config_top.fetch_block_resources.usage_line")
	addUsageNotes(cmd, ctx, "cmd.boot_config_top.fetch_block_resources.usage_notes")
	for _, entry := range catalog.EnabledBlockClouds() {
		cmd.AddCommand(newBootConfigFetchResourcesProviderCommand(ctx, entry, "block"))
	}
	return cmd
}

func newBootConfigFetchOSSResourcesCommand(ctx *context) *cobra.Command {
	cmd := newGroupCommand(ctx, "fetch-oss-resources", "cmd.boot_config_top.fetch.oss.short", "cmd.boot_config_top.fetch.oss.long", "cmd.boot_config_top.fetch.oss.examples", "cmd.boot_config_top.fetch.oss.notes", "boot-config fetch-oss-resources")
	addHelpLayout(cmd, helpLayoutGroup)
	addUsageLine(cmd, ctx, "cmd.boot_config_top.fetch_oss_resources.usage_line")
	addUsageNotes(cmd, ctx, "cmd.boot_config_top.fetch_oss_resources.usage_notes")
	for _, entry := range catalog.EnabledObjectClouds() {
		cmd.AddCommand(newBootConfigFetchResourcesProviderCommand(ctx, entry, "objectstorage"))
	}
	return cmd
}

func newBootConfigFetchResourcesProviderCommand(ctx *context, entry catalog.CloudEntry, storageType string) *cobra.Command {
	shortKey, longKey, usageLineKey, exampleKey := bootConfigFetchProviderTextKeys(storageType)
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
			return runBootConfigFetchResourcesForProvider(ctx, bootConfigFetchResourcesProviderCommandName(storageType, entry.Provider), entry.Provider, entry.CloudType, storageType, args)
		},
	}

	addBootConfigFetchResourceFlags(cmd, ctx)
	addHelpLayout(cmd, helpLayoutFourSection)
	addAnnotationValue(cmd, usageLineAnnotation, fmt.Sprintf(ctx.loc.T(usageLineKey), entry.Provider))
	addAnnotationValue(cmd, usageNotesAnnotation, bootConfigFetchProviderUsageNotes(ctx, entry, storageType))
	return cmd
}

func addBootConfigFetchResourceFlags(cmd *cobra.Command, ctx *context) {
	for _, name := range []string{
		"cloud-account-id",
		"fetch-res",
		"host-id",
		"storage-id",
		"network-addr-for-write-data",
		"network-addr-for-read-data",
		"region-id",
		"zone-id",
		"cloud-account-username",
		"cloud-account-use-public",
		"flavor-id",
		"boot-loader-flavor-id",
		"arch",
		"os-type-id",
		"os-type",
		"flavors",
		"flavor-vcpus",
		"flavor-ram",
		"max-nic-num",
		"system-volume-type-id",
		"volume-type-id",
		"default-volume-type-id",
		"default-pool-id",
		"dest-boot-mode",
		"network-id",
	} {
		addFlagString(cmd, ctx, name)
	}
}

func runBootConfigFetchResourcesForProvider(ctx *context, commandName, provider, cloudType, storageType string, args []string) error {
	fs := newFlagSet(commandName)
	cloudAccountID := fs.String("cloud-account-id", "", "")
	fetchRes := fs.String("fetch-res", "", "")
	hostID := fs.String("host-id", "", "")
	storageID := fs.String("storage-id", "", "")
	writeNetwork := fs.String("network-addr-for-write-data", "", "")
	readNetwork := fs.String("network-addr-for-read-data", "", "")
	regionID := fs.String("region-id", "", "")
	zoneID := fs.String("zone-id", "", "")
	cloudAccountUsername := fs.String("cloud-account-username", "", "")
	cloudAccountUsePublic := fs.String("cloud-account-use-public", "", "")
	flavorID := fs.String("flavor-id", "", "")
	bootLoaderFlavorID := fs.String("boot-loader-flavor-id", "", "")
	arch := fs.String("arch", "", "")
	osTypeID := fs.String("os-type-id", "", "")
	osType := fs.String("os-type", "", "")
	flavors := fs.String("flavors", "", "")
	flavorVCPUs := fs.String("flavor-vcpus", "", "")
	flavorRAM := fs.String("flavor-ram", "", "")
	maxNICNum := fs.String("max-nic-num", "", "")
	systemVolumeTypeID := fs.String("system-volume-type-id", "", "")
	volumeTypeID := fs.String("volume-type-id", "", "")
	defaultVolumeTypeID := fs.String("default-volume-type-id", "", "")
	defaultPoolID := fs.String("default-pool-id", "", "")
	destBootMode := fs.String("dest-boot-mode", "", "")
	networkID := fs.String("network-id", "", "")
	q := queryFromPairs()
	if err := parseQueryFlagsIntoPassthrough(fs, args, q); err != nil {
		return err
	}

	spec := appbootconfigquery.ResourcesSpec{
		CloudAccountID:        *cloudAccountID,
		CloudType:             cloudType,
		StorageType:           bootConfigFetchResourcesStorageType(storageType),
		FetchRes:              *fetchRes,
		HostID:                *hostID,
		StorageID:             *storageID,
		WriteNetwork:          *writeNetwork,
		ReadNetwork:           *readNetwork,
		RegionID:              *regionID,
		ZoneID:                *zoneID,
		CloudAccountUsername:  *cloudAccountUsername,
		CloudAccountUsePublic: *cloudAccountUsePublic,
		FlavorID:              *flavorID,
		BootLoaderFlavorID:    *bootLoaderFlavorID,
		Arch:                  *arch,
		OSTypeID:              *osTypeID,
		OSType:                *osType,
		Flavors:               *flavors,
		FlavorVCPUs:           *flavorVCPUs,
		FlavorRAM:             *flavorRAM,
		MaxNICNum:             *maxNICNum,
		SystemVolumeTypeID:    *systemVolumeTypeID,
		VolumeTypeID:          *volumeTypeID,
		DefaultVolumeTypeID:   *defaultVolumeTypeID,
		DefaultPoolID:         *defaultPoolID,
		DestBootMode:          *destBootMode,
		NetworkID:             *networkID,
		Query:                 q,
	}
	if err := validateBootConfigFetchResources(provider, storageType, spec); err != nil {
		return err
	}

	service := appbootconfigquery.NewService(commandAPIAdapter{ctx: ctx})
	resp, err := service.Resources(spec)
	if err != nil {
		return err
	}
	return writeAuthResourcesResponse(ctx, resp, provider, cloudType, spec.StorageType, spec.FetchRes, spec.FlavorVCPUs, spec.FlavorRAM)
}

func validateBootConfigFetchResources(provider, storageType string, spec appbootconfigquery.ResourcesSpec) error {
	if strings.TrimSpace(spec.CloudAccountID) == "" {
		return fmt.Errorf("cloud-account-id is required")
	}
	if provider == "aliyun" && storageType == "objectstorage" {
		return validateAliyunObjectBootConfigFetchResources(spec)
	}
	if provider == "huawei" && storageType == "objectstorage" {
		return validateHuaweiObjectBootConfigFetchResources(spec)
	}
	return nil
}

func validateAliyunObjectBootConfigFetchResources(spec appbootconfigquery.ResourcesSpec) error {
	if strings.TrimSpace(spec.FetchRes) == "" {
		return nil
	}
	requested := normalizeAuthRequestedResources(spec.FetchRes)
	needZone := false
	needFlavor := false
	for _, resource := range requested {
		switch resource {
		case "flavors", "os_types", "subnets", "security_groups":
			needZone = true
		case "system_volume_types", "volume_types":
			needZone = true
			needFlavor = true
		}
	}
	if needZone && strings.TrimSpace(spec.ZoneID) == "" {
		return fmt.Errorf("zone-id is required")
	}
	if needFlavor && strings.TrimSpace(spec.FlavorID) == "" {
		return fmt.Errorf("flavor-id is required")
	}
	return nil
}

func validateHuaweiObjectBootConfigFetchResources(spec appbootconfigquery.ResourcesSpec) error {
	if strings.TrimSpace(spec.FetchRes) == "" {
		return nil
	}
	requested := normalizeAuthRequestedResources(spec.FetchRes)
	needZone := false
	for _, resource := range requested {
		switch resource {
		case "flavors", "os_types", "system_volume_types", "volume_types":
			needZone = true
		}
	}
	if needZone && strings.TrimSpace(spec.ZoneID) == "" {
		return fmt.Errorf("zone-id is required")
	}
	return nil
}

func bootConfigFetchResourcesStorageType(storageType string) string {
	if storageType == "block" {
		return "HyperGate"
	}
	return "objectstorage"
}

func bootConfigFetchResourcesProviderCommandName(storageType, provider string) string {
	if storageType == "block" {
		return "boot-config fetch-block-resources " + provider
	}
	return "boot-config fetch-oss-resources " + provider
}

func bootConfigFetchProviderTextKeys(storageType string) (shortKey, longKey, usageLineKey, exampleKey string) {
	if storageType == "block" {
		return "cmd.boot_config_top.fetch.provider.block.short",
			"cmd.boot_config_top.fetch.provider.block.long",
			"cmd.boot_config_top.fetch_block_resources.provider.usage_line",
			"cmd.boot_config_top.fetch.provider.block.examples"
	}
	return "cmd.boot_config_top.fetch.provider.oss.short",
		"cmd.boot_config_top.fetch.provider.oss.long",
		"cmd.boot_config_top.fetch_oss_resources.provider.usage_line",
		"cmd.boot_config_top.fetch.provider.oss.examples"
}

func bootConfigFetchProviderUsageNotes(ctx *context, entry catalog.CloudEntry, storageType string) string {
	switch {
	case storageType == "objectstorage" && entry.Provider == "aliyun":
		return ctx.loc.T("cmd.boot_config_top.fetch_oss_resources.aliyun.usage_notes")
	case storageType == "objectstorage" && entry.Provider == "huawei":
		return ctx.loc.T("cmd.boot_config_top.fetch_oss_resources.huawei.usage_notes")
	case storageType == "block":
		return fmt.Sprintf(ctx.loc.T("cmd.boot_config_top.fetch_block_resources.provider.usage_notes"), localizedCloudEntryName(ctx, entry), entry.Provider)
	default:
		return fmt.Sprintf(ctx.loc.T("cmd.boot_config_top.fetch_oss_resources.provider.usage_notes"), localizedCloudEntryName(ctx, entry), entry.Provider)
	}
}
