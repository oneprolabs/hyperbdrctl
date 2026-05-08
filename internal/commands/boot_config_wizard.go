package commands

import (
	appbootconfigwizard "hyperbdr-client/internal/app/bootconfigwizard"
	normalizecloudinfo "hyperbdr-client/internal/normalize/cloudinfo"
	"hyperbdr-client/internal/output"
)

func runBootConfigWizard(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("boot-config-wizard", "")
	}

	service := appbootconfigwizard.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "storages":
		fs := newFlagSet("boot-config-wizard storages")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 100, "")
		storageType := fs.String("type", "", "")
		status := fs.String("status", "available", "")
		q := queryFromPairs()
		if err := parseQueryFlagsInto(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.Storages(appbootconfigwizard.StoragesSpec{
			Page:        *page,
			PageSize:    *pageSize,
			StorageType: *storageType,
			Status:      *status,
			Query:       q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "storages", objectStorageColumns())
	case "storage-detail":
		fs := newFlagSet("boot-config-wizard storage-detail")
		storageID := fs.String("storage-id", "", "")
		q := queryFromPairs()
		if err := parseQueryFlagsInto(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.StorageDetail(appbootconfigwizard.StorageDetailSpec{
			StorageID: *storageID,
			Query:     q,
		})
		if err != nil {
			return err
		}
		if rows := normalizecloudinfo.StorageNetworkRows(resp.Data); len(rows) > 0 && ctx.cfg.Output != "json" {
			return output.Table(ctx.out, ctx.loc, rows, wizardStorageNetworkColumns())
		}
		return writeResponse(ctx, resp, "", nil)
	case "target-platforms":
		return runBootConfigWizardTargetAccounts(ctx, args[1:], true)
	case "target-accounts":
		return runBootConfigWizardTargetAccounts(ctx, args[1:], false)
	case "target-auth-info":
		fs := newFlagSet("boot-config-wizard target-auth-info")
		cloudAccountID := fs.String("cloud-account-id", "", "")
		cloudAccountCompat := fs.String("cloud-account", "", "")
		cloudType := fs.String("cloud-type", "", "")
		storageType := fs.String("storage-type", "", "")
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
		if err := parseQueryFlagsInto(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.TargetAuthInfo(appbootconfigwizard.TargetAuthInfoSpec{
			CloudAccountID:        *cloudAccountID,
			CloudAccountCompat:    *cloudAccountCompat,
			CloudType:             *cloudType,
			StorageType:           *storageType,
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
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	case "subnet-config":
		fs := newFlagSet("boot-config-wizard subnet-config")
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
		resp, err := service.SubnetConfig(appbootconfigwizard.SubnetConfigSpec{
			CloudAccountID: *cloudAccountID,
			CloudType:      *cloudType,
			RegionID:       *regionID,
			ZoneID:         *zoneID,
			NetworkID:      *networkID,
			SubnetID:       *subnetID,
			Query:          q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	case "host-profile":
		fs := newFlagSet("boot-config-wizard host-profile")
		id := fs.String("id", "", "")
		q := queryFromPairs()
		if err := parseQueryFlagsInto(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.HostProfile(appbootconfigwizard.HostProfileSpec{
			ID:    *id,
			Query: q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	case "strategies":
		fs := newFlagSet("boot-config-wizard strategies")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 100, "")
		kw := fs.String("kw", "", "")
		status := fs.String("status", "", "")
		q := queryFromPairs()
		if err := parseQueryFlagsInto(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.Strategies(appbootconfigwizard.StrategiesSpec{
			Page:     *page,
			PageSize: *pageSize,
			KW:       *kw,
			Status:   *status,
			Query:    q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "policies", wizardStrategyColumns())
	default:
		return errUnknown("boot-config-wizard", args[0])
	}
}

func wizardStorageNetworkColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.type", Field: "type"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.value", Field: "value"},
	}
}

func wizardStrategyColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.status", Field: "status"},
		{HeaderKey: "table.created_at", Field: "created_at"},
	}
}
