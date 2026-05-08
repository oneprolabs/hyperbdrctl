package commands

import (
	"fmt"

	appblockstorage "hyperbdr-client/internal/app/blockstorage"
	"hyperbdr-client/internal/output"
)

func runBlockStorages(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("target cloud-sync-gateway", "")
	}
	service := appblockstorage.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "list":
		fs := newFlagSet("target cloud-sync-gateway list")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 100, "")
		storageType := fs.String("type", "HyperGate", "")
		cloudAccountID := fs.String("cloud-account-id", "", "")
		q := queryFromPairs()
		if err := parseQueryFlagsInto(fs, args[1:], q); err != nil {
			return err
		}
		addString(q, "cloud_account_uuid", *cloudAccountID)
		resp, err := service.List(appblockstorage.ListSpec{
			Page:        *page,
			PageSize:    *pageSize,
			StorageType: *storageType,
			Query:       q,
		})
		if err != nil {
			return err
		}
		if ctx.cfg.Output == "json" {
			return writeResponse(ctx, resp, "", nil)
		}
		return output.Table(ctx.out, ctx.loc, normalizeGatewayStorageRows(resp.Data), gatewayStorageColumns())
	case "detail":
		fs := newFlagSet("target cloud-sync-gateway detail")
		id := fs.String("id", "", "")
		q := queryFromPairs()
		if err := parseQueryFlagsInto(fs, args[1:], q); err != nil {
			return err
		}
		if *id == "" {
			return missing(ctx, "error.missing_id")
		}
		resp, err := service.Detail(appblockstorage.DetailSpec{
			ID:    *id,
			Query: q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	case "wait":
		return runTargetCloudSyncGatewayWait(ctx, args[1:])
	case "resources":
		fs := newFlagSet("target cloud-sync-gateway resources")
		cloudAccountID := fs.String("cloud-account-id", "", "")
		fetchRes := fs.String("fetch-res", "", "")
		regionID := fs.String("region-id", "", "")
		zoneID := fs.String("zone-id", "", "")
		flavorID := fs.String("flavor-id", "", "")
		flavorVCPUs := fs.String("flavor-vcpus", "", "")
		flavorRAM := fs.String("flavor-ram", "", "")
		purpose := fs.String("purpose", "make_hg", "")
		imageType := fs.String("image-type", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return executeGatewayResources(ctx, *cloudAccountID, *fetchRes, *regionID, *zoneID, *flavorID, *flavorVCPUs, *flavorRAM, *purpose, *imageType)
	case "subnet-config":
		fs := newFlagSet("target cloud-sync-gateway subnet-config")
		cloudAccountID := fs.String("cloud-account-id", "", "")
		cloudType := fs.String("cloud-type", "", "")
		regionID := fs.String("region-id", "", "")
		zoneID := fs.String("zone-id", "", "")
		networkID := fs.String("network-id", "", "")
		subnetID := fs.String("subnet-id", "", "")
		q := queryFromPairs()
		if err := parseQueryFlagsInto(fs, args[1:], q); err != nil {
			return err
		}
		return executeGatewaySubnetConfig(ctx, *cloudAccountID, *cloudType, *regionID, *zoneID, *networkID, *subnetID, q)
	case "create":
		return runCreateBlockStorage(ctx, args[1:])
	default:
		return errUnknown("target cloud-sync-gateway", args[0])
	}
}

func runCreateBlockStorage(ctx *context, args []string) error {
	parsed, err := parseBlockStorageCreateArgs("target cloud-sync-gateway create", args)
	if err != nil {
		return err
	}
	if len(parsed.remainingArgs) > 0 {
		return errUnknown("target cloud-sync-gateway create", parsed.remainingArgs[0])
	}
	return executeBlockStorageCreateSpec(ctx, parsed.spec, parsed.previewRequest)
}

type parsedBlockStorageCreateCommand struct {
	spec           blockStorageCreateSpec
	previewRequest bool
	cloudTypeSet   bool
	remainingArgs  []string
}

func parseBlockStorageCreateArgs(commandName string, args []string) (parsedBlockStorageCreateCommand, error) {
	fs := newFlagSet("target cloud-sync-gateway create")
	cloudAccountID := fs.String("cloud-account-id", "", "")
	cloudType := fs.String("cloud-type", "", "")
	projectID := fs.String("project-id", "", "")
	regionID := fs.String("region-id", "", "")
	zoneID := fs.String("zone-id", "", "")
	computeZoneID := fs.String("compute-zone-id", "", "")
	imageID := fs.String("image-id", "", "")
	flavorID := fs.String("flavor-id", "", "")
	networkID := fs.String("network-id", "", "")
	subnetID := fs.String("subnet-id", "", "")
	fixedIP := fs.String("fixed-ip", "", "")
	systemDiskTypeID := fs.String("system-disk-type-id", "", "")
	volumeTypeID := fs.String("volume-type-id", "", "")
	systemDiskSize := fs.String("system-disk-size", "", "")
	blockStoreZoneID := fs.String("block-store-zone-id", "", "")
	bootLoaderImageID := fs.String("boot-loader-image-id", "", "")
	bootLoaderFlavorID := fs.String("boot-loader-flavor-id", "", "")
	projectDomainID := fs.String("project-domain-id", "", "")
	bootTypesID := fs.String("boot-types-id", "boot_from_volume", "")
	volumeProxyType := fs.String("volume-proxy-type", "s3", "")
	hgControlNetwork := fs.String("hg-control-network", "floating_ip_without_proxy", "")
	controlNATIP := fs.String("control-nat-ip", "", "")
	hgDataNetwork := fs.String("hg-data-network", "floating_ip_without_proxy", "")
	dataNATIP := fs.String("data-nat-ip", "", "")
	bandwidthSize := fs.String("bandwidth-size", "", "")
	hdControlNetwork := fs.String("hd-control-network", "floating_ip_with_hg_proxy", "")
	previewRequest := fs.Bool("preview-request", false, "")
	if err := fs.Parse(args); err != nil {
		return parsedBlockStorageCreateCommand{}, err
	}

	spec := blockStorageCreateSpec{
		CloudAccountID:     *cloudAccountID,
		CloudType:          *cloudType,
		ProjectID:          *projectID,
		RegionID:           *regionID,
		ZoneID:             *zoneID,
		ComputeZoneID:      *computeZoneID,
		ImageID:            *imageID,
		FlavorID:           *flavorID,
		NetworkID:          *networkID,
		SubnetID:           *subnetID,
		FixedIP:            *fixedIP,
		SystemDiskTypeID:   *systemDiskTypeID,
		VolumeTypeID:       *volumeTypeID,
		SystemDiskSize:     *systemDiskSize,
		BlockStoreZoneID:   *blockStoreZoneID,
		BootLoaderImageID:  *bootLoaderImageID,
		BootLoaderFlavorID: *bootLoaderFlavorID,
		ProjectDomainID:    *projectDomainID,
		BootTypesID:        *bootTypesID,
		VolumeProxyType:    *volumeProxyType,
		HGControlNetwork:   *hgControlNetwork,
		ControlNATIP:       *controlNATIP,
		HGDataNetwork:      *hgDataNetwork,
		DataNATIP:          *dataNATIP,
		BandwidthSize:      *bandwidthSize,
		HDControlNetwork:   *hdControlNetwork,
	}
	return parsedBlockStorageCreateCommand{
		spec:           spec,
		previewRequest: *previewRequest,
		cloudTypeSet:   flagWasSet(fs, "cloud-type"),
		remainingArgs:  fs.Args(),
	}, nil
}

func runCreateBlockStorageForProvider(ctx *context, commandName, cloudType string, args []string) error {
	parsed, err := parseBlockStorageCreateArgs(commandName, args)
	if err != nil {
		return err
	}
	if len(parsed.remainingArgs) > 0 {
		return errUnknown(commandName, parsed.remainingArgs[0])
	}
	if parsed.cloudTypeSet {
		return fmt.Errorf("cloud-type cannot be used with %s", commandName)
	}
	parsed.spec.CloudType = cloudType
	return executeBlockStorageCreateSpec(ctx, parsed.spec, parsed.previewRequest)
}
