package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"hyperbdr-client/internal/metaoverride"
	workflowcreate "hyperbdr-client/internal/workflow/cloudaccountcreate"
)

func parseCloudAccountCreateBlockArgs(commandName, cloudType string, specialized bool, args []string) (parsedCloudAccountCreateCommand, error) {
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
		case "account-name":
			v, next, err := strictFlagValue(args, i, value, hasInline)
			if err != nil {
				return parsedCloudAccountCreateCommand{}, err
			}
			spec.AccountName = v
			assignments = append(assignments, metadataAssignment{key: "account_name", value: v})
			i = next
		case "only-verify":
			return parsedCloudAccountCreateCommand{}, fmt.Errorf("unknown flag: --%s", name)
		default:
			if handled, err := parseSpecializedBlockFlag(name, args, i, value, hasInline, cloudType, &spec, &assignments); handled {
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
			applyCreateCloudAccountSpecValue(&spec, name, v)
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
	spec.StorageType = "HyperGate"
	spec = workflowcreate.NormalizeSpec(spec)

	return parsedCloudAccountCreateCommand{
		spec:           spec,
		previewRequest: previewRequest,
	}, nil
}

func parseSpecializedBlockFlag(name string, args []string, idx int, inline string, hasInline bool, cloudType string, spec *cloudAccountCreateSpec, assignments *[]metadataAssignment) (bool, error) {
	if !createBlockExplicitFlagAllowed(cloudType, name) {
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

func createBlockExplicitFlagAllowed(cloudType, name string) bool {
	switch cloudType {
	case "aliyun_bs":
		switch name {
		case "access-key-id", "access-key-secret", "region-id", "region-name", "account-name", "auth-region-id":
			return true
		}
	case "openstack":
		switch name {
		case "auth-url", "username", "password", "user-domain-id", "project-domain-id", "project-name", "region-name", "ssh-port", "ssh-pass", "linux-hd-username", "linux-hd-password", "linux-hd-port":
			return true
		}
	}
	return false
}
