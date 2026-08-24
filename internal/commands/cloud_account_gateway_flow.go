package commands

import (
	"errors"
	"net/url"

	appblockstorage "hyperbdr-client/internal/app/blockstorage"
	"hyperbdr-client/internal/output"
)

func executeGatewaySubnetConfig(ctx *context, accountID, cloudType, regionID, zoneID, networkID, subnetID string, q url.Values) error {
	service := appblockstorage.NewService(commandAPIAdapter{ctx: ctx})
	resp, err := service.SubnetConfig(appblockstorage.SubnetConfigSpec{
		CloudAccountID: accountID,
		CloudType:      cloudType,
		RegionID:       regionID,
		ZoneID:         zoneID,
		NetworkID:      networkID,
		SubnetID:       subnetID,
		Query:          q,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}

func executeBlockStorageCreateSpec(ctx *context, spec blockStorageCreateSpec, previewRequest bool) error {
	if spec.CloudType == "huawei_bs" && createMetadataKeyPresent(spec.ExplicitMetadataKeys, "region_id") {
		return errors.New(ctx.loc.T("error.cloud_sync_gateway.create.huawei.region_id"))
	}
	service := appblockstorage.NewService(commandAPIAdapter{ctx: ctx})
	if previewRequest {
		prepared, err := service.PrepareCreate(spec)
		if err != nil {
			return err
		}
		return output.JSON(ctx.out, prepared.Body)
	}
	resp, err := service.Create(spec)
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}

func createMetadataKeyPresent(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
