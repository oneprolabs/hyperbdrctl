package commands

import (
	"fmt"
	"strings"

	appcloudaccount "hyperbdr-client/internal/app/cloudaccount"
	workflowcreate "hyperbdr-client/internal/workflow/cloudaccountcreate"
)

func parseCloudAccountFetchResourcesArgs(commandName, cloudType, storageType string, specialized bool, args []string) (parsedCloudAccountFetchResourcesCommand, error) {
	spec := appcloudaccount.FetchResourcesSpec{}
	spec.CloudType = cloudType
	spec.StorageType = fetchResourcesStorageType(storageType)
	if specialized {
		spec.CloudAuthType = "password"
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return parsedCloudAccountFetchResourcesCommand{}, fmt.Errorf("unexpected argument %q", arg)
		}
		name, value, hasInline := splitFlag(arg)
		name = strings.TrimPrefix(name, "--")
		switch name {
		case "cloud-type":
			return parsedCloudAccountFetchResourcesCommand{}, fmt.Errorf("cloud-type cannot be used with %s", commandName)
		case "storage-type":
			return parsedCloudAccountFetchResourcesCommand{}, fmt.Errorf("storage-type cannot be used with %s", commandName)
		case "cloud-account-username", "cloud-account-password":
			return parsedCloudAccountFetchResourcesCommand{}, fmt.Errorf("unknown flag: --%s", name)
		case "cloud-auth-type":
			if specialized {
				return parsedCloudAccountFetchResourcesCommand{}, fmt.Errorf("cloud-auth-type cannot be used with %s", commandName)
			}
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountFetchResourcesCommand{}, err
			}
			spec.CloudAuthType = v
			i = next
		case "boot-mode":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountFetchResourcesCommand{}, err
			}
			spec.BootMode = v
			i = next
		case "fetch-res":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountFetchResourcesCommand{}, err
			}
			spec.FetchRes = v
			i = next
		case "zone-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountFetchResourcesCommand{}, err
			}
			spec.ZoneID = v
			i = next
		case "flavor-vcpus":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountFetchResourcesCommand{}, err
			}
			spec.FlavorVCPUs = v
			i = next
		case "flavor-ram":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountFetchResourcesCommand{}, err
			}
			spec.FlavorRAM = v
			i = next
		case "compute-zone-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountFetchResourcesCommand{}, err
			}
			spec.ComputeZoneID = v
			i = next
		case "block-store-zone-id":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountFetchResourcesCommand{}, err
			}
			spec.BlockStoreZoneID = v
			i = next
		default:
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountFetchResourcesCommand{}, err
			}
			applyCreateCloudAccountSpecValue(&spec.Spec, name, v)
			i = next
		}
	}

	spec.Spec = workflowcreate.NormalizeSpec(spec.Spec)
	if specialized {
		spec.CloudAuthType = "password"
	}
	return parsedCloudAccountFetchResourcesCommand{spec: spec}, nil
}
