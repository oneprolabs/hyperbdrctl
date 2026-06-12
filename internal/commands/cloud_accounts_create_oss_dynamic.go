package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"hyperbdr-client/internal/metaoverride"
	workflowcreate "hyperbdr-client/internal/workflow/cloudaccountcreate"
)

type metadataAssignment struct {
	key   string
	value interface{}
}

func parseCloudAccountCreateOSSArgs(commandName, cloudType string, specialized bool, args []string) (parsedCloudAccountCreateCommand, error) {
	spec := cloudAccountCreateSpec{
		MetadataOverrides: map[string]interface{}{},
		RequestOverrides:  map[string]interface{}{},
	}
	filePath := ""
	sets := []string{}
	setJSONs := []string{}
	assignments := []metadataAssignment{}
	previewRequest := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return parsedCloudAccountCreateCommand{}, fmt.Errorf("unexpected argument %q", arg)
		}
		name, value, hasInline := splitFlag(arg)
		name = strings.TrimPrefix(name, "--")
		switch name {
		case "preview-request":
			if hasInline {
				enabled, err := strconv.ParseBool(value)
				if err != nil {
					return parsedCloudAccountCreateCommand{}, fmt.Errorf("%s requires boolean value", arg)
				}
				previewRequest = enabled
				continue
			}
			previewRequest = true
		case "file":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountCreateCommand{}, err
			}
			filePath = v
			i = next
		case "set":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountCreateCommand{}, err
			}
			sets = append(sets, v)
			i = next
		case "set-json":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountCreateCommand{}, err
			}
			setJSONs = append(setJSONs, v)
			i = next
		case "cloud-type":
			return parsedCloudAccountCreateCommand{}, fmt.Errorf("cloud-type cannot be used with %s", commandName)
		case "storage-type":
			return parsedCloudAccountCreateCommand{}, fmt.Errorf("storage-type cannot be used with %s", commandName)
		case "cloud-account-username", "cloud-account-password":
			return parsedCloudAccountCreateCommand{}, fmt.Errorf("unknown flag: --%s", name)
		case "cloud-auth-type":
			if specialized {
				return parsedCloudAccountCreateCommand{}, fmt.Errorf("cloud-auth-type cannot be used with %s", commandName)
			}
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountCreateCommand{}, err
			}
			spec.CloudAuthType = v
			i = next
		case "custom-name":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountCreateCommand{}, err
			}
			spec.CustomName = v
			assignments = append(assignments, metadataAssignment{key: "custom_name", value: v})
			i = next
		case "auto-upload-images", "upload-uefi-image", "only-verify":
			return parsedCloudAccountCreateCommand{}, fmt.Errorf("unknown flag: --%s", name)
		default:
			if handled, err := parseSpecializedOSSFlag(name, args, i, value, hasInline, cloudType, &spec, &assignments); handled {
				if err != nil {
					return parsedCloudAccountCreateCommand{}, err
				}
				if hasInline {
					continue
				}
				i++
				continue
			}

			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountCreateCommand{}, err
			}
			key := createCloudAccountDynamicMetadataKey(name)
			if err := validateCreateCloudAccountDynamicMetadataKey(name, key); err != nil {
				return parsedCloudAccountCreateCommand{}, err
			}
			assignments = append(assignments, metadataAssignment{key: key, value: v})
			i = next
		}
	}

	metadata, err := loadCreateCloudAccountMetadataFile(filePath)
	if err != nil {
		return parsedCloudAccountCreateCommand{}, err
	}
	for _, raw := range setJSONs {
		path, value, err := metaoverride.SplitAssignment(raw)
		if err != nil {
			return parsedCloudAccountCreateCommand{}, err
		}
		var decoded interface{}
		if err := json.Unmarshal([]byte(value), &decoded); err != nil {
			return parsedCloudAccountCreateCommand{}, fmt.Errorf("invalid JSON for %s: %w", path, err)
		}
		if err := metaoverride.ApplyPathValue(metadata, path, decoded); err != nil {
			return parsedCloudAccountCreateCommand{}, err
		}
	}
	for _, raw := range sets {
		path, value, err := metaoverride.SplitAssignment(raw)
		if err != nil {
			return parsedCloudAccountCreateCommand{}, err
		}
		if err := metaoverride.ApplyPathValue(metadata, path, metaoverride.InferValue(value)); err != nil {
			return parsedCloudAccountCreateCommand{}, err
		}
	}
	for _, assignment := range assignments {
		metadata[assignment.key] = assignment.value
	}

	spec.MetadataOverrides = metadata
	spec.CloudType = cloudType
	spec.StorageType = "objectstorage"
	spec = workflowcreate.NormalizeSpec(spec)

	return parsedCloudAccountCreateCommand{
		spec:           spec,
		previewRequest: previewRequest,
	}, nil
}

func parseSpecializedOSSFlag(name string, args []string, idx int, inline string, hasInline bool, cloudType string, spec *cloudAccountCreateSpec, assignments *[]metadataAssignment) (bool, error) {
	if !createOSSExplicitFlagAllowed(cloudType, name) {
		return false, nil
	}
	value, _, err := strictFlagValue(args, idx, inline, hasInline)
	if err != nil {
		return true, err
	}
	applyCreateCloudAccountSpecValue(spec, name, value)
	if metadataKey, ok := createCloudAccountMetadataKey(name); ok {
		*assignments = append(*assignments, metadataAssignment{key: metadataKey, value: value})
	}
	return true, nil
}

func createOSSExplicitFlagAllowed(cloudType, name string) bool {
	switch cloudType {
	case "aliyun_obs":
		switch name {
		case "access-key-id", "access-key-secret", "region-id", "region-name", "use-internal-ip", "boot-loader-image-id", "boot-loader-image-name", "boot-loader-flavor-id", "linux-boot-image-id", "windows-boot-image-id", "linux-uefi-boot-image-id", "windows-uefi-boot-image-id":
			return true
		}
	case "openstack":
		switch name {
		case "auth-url", "username", "password", "user-domain-id", "project-domain-id", "project-id", "project-name", "region-id", "region-name", "use-internal-ip", "boot-loader-image-id", "boot-loader-image-name", "linux-boot-image-id", "windows-boot-image-id", "disk-bus-type-id", "disk-bus-type-name":
			return true
		}
	}
	return false
}

func applyCreateCloudAccountSpecValue(spec *cloudAccountCreateSpec, name, value string) {
	switch name {
	case "access-key-id":
		spec.AccessKeyID = value
	case "access-key-secret":
		spec.AccessKeySecret = value
	case "account-name":
		spec.AccountName = value
	case "auth-region-id":
		spec.AuthRegionID = value
	case "region-id":
		spec.RegionID = value
	case "region-name":
		spec.RegionName = value
	case "auth-url":
		spec.AuthURL = value
	case "username":
		spec.CloudAccountUsername = value
	case "password":
		spec.CloudAccountPassword = value
	case "user-domain-id":
		spec.UserDomainID = value
	case "project-domain-id":
		spec.ProjectDomainID = value
	case "project-id":
		spec.ProjectID = value
	case "project-name":
		spec.ProjectName = value
	case "ssh-port":
		spec.SSHPort = value
	case "ssh-pass":
		spec.SSHPass = value
	case "linux-hd-username":
		spec.LinuxHDUsername = value
	case "linux-hd-password":
		spec.LinuxHDPassword = value
	case "linux-hd-port":
		spec.LinuxHDPort = value
	case "use-internal-ip":
		spec.UseInternalIP = value
	case "boot-loader-image-id":
		spec.BootLoaderImageID = value
	case "boot-loader-image-name":
		spec.BootLoaderImageName = value
	case "boot-loader-flavor-id":
		spec.BootLoaderFlavorID = value
	case "linux-boot-image-id":
		spec.LinuxBootImageID = value
	case "windows-boot-image-id":
		spec.WindowsBootImageID = value
	case "linux-uefi-boot-image-id":
		spec.LinuxUEFIBootImageID = value
	case "windows-uefi-boot-image-id":
		spec.WindowsUEFIBootImageID = value
	case "disk-bus-type-id":
		spec.DiskBusTypeID = value
	case "disk-bus-type-name":
		spec.DiskBusTypeName = value
	}
}

func createCloudAccountMetadataKey(name string) (string, bool) {
	switch name {
	case "username":
		return "username", true
	case "password":
		return "password", true
	case "auth-url":
		return "auth_url", true
	case "use-internal-ip":
		return "use_internal_ip_for_control", true
	}
	if name == "" {
		return "", false
	}
	return strings.ReplaceAll(name, "-", "_"), true
}

func createCloudAccountDynamicMetadataKey(name string) string {
	if key, ok := createCloudAccountMetadataKey(name); ok {
		return key
	}
	return strings.ReplaceAll(name, "-", "_")
}

func validateCreateCloudAccountDynamicMetadataKey(flagName, metadataKey string) error {
	switch metadataKey {
	case "cloud_type", "storage_type":
		return fmt.Errorf("%s cannot be used as dynamic metadata override", flagName)
	}
	return nil
}

func loadCreateCloudAccountMetadataFile(path string) (map[string]interface{}, error) {
	if path == "" {
		return map[string]interface{}{}, nil
	}
	return metaoverride.ReadObjectFile(path, "metadata object", map[string]string{
		"cloud_account": "cloud_account wrapper",
		"batch_create":  "batch_create wrapper",
		"batch_update":  "batch_update wrapper",
	})
}
