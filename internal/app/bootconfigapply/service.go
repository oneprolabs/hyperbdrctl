package bootconfigapply

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/metaoverride"
	"hyperbdr-client/internal/normalize/cloudinfo"
)

type API interface {
	Get(path string, q url.Values) (client.APIResponse, error)
	Post(path string, body interface{}) (client.APIResponse, error)
}

type Service struct {
	api API
}

type ApplyInput struct {
	HostID         string
	File           string
	Dynamic        map[string]string
	Sets           []string
	SetJSONs       []string
	PreviewRequest bool
}

type MutationResult struct {
	Response     client.APIResponse
	Operation    string
	Metadata     map[string]interface{}
	HostID       string
	BootConfigID string
	NoOp         bool
}

type PreparedRequest struct {
	Operation    string
	Path         string
	Body         map[string]interface{}
	Metadata     map[string]interface{}
	HostID       string
	BootConfigID string
	NoOp         bool
}

type hostDisk struct {
	Index      int
	DiskID     string
	IsBootDisk bool
}

type hostDetail struct {
	Data         map[string]interface{}
	BootConfigID string
	Disks        []hostDisk
}

type storageDefaults struct {
	IsObjectStorage        bool
	ZoneID                 string
	StorageName            string
	StorageDisplayName     string
	WriteNetwork           string
	WriteNetworkName       string
	ReadNetwork            string
	ReadNetworkName        string
	PoolID                 string
	PoolName               string
	DefaultPoolDisplayName string
	PoolByVolumeType       map[string]poolRef
	StorageID              string
	IsBlockStorage         bool
}

type poolRef struct {
	ID          string
	Name        string
	DisplayName string
}

func NewService(api API) Service {
	return Service{api: api}
}

func (s Service) Apply(input ApplyInput) (MutationResult, error) {
	prepared, err := s.PrepareRequest(input)
	if err != nil {
		return MutationResult{}, err
	}
	if prepared.NoOp {
		return MutationResult{
			Operation:    prepared.Operation,
			Metadata:     prepared.Metadata,
			HostID:       prepared.HostID,
			BootConfigID: prepared.BootConfigID,
			NoOp:         true,
		}, nil
	}
	resp, err := s.api.Post(prepared.Path, prepared.Body)
	return MutationResult{
		Response:     resp,
		Operation:    prepared.Operation,
		Metadata:     prepared.Metadata,
		HostID:       prepared.HostID,
		BootConfigID: prepared.BootConfigID,
		NoOp:         prepared.NoOp,
	}, err
}

func (s Service) PrepareRequest(input ApplyInput) (PreparedRequest, error) {
	metadata, err := metadataFile(input)
	if err != nil {
		return PreparedRequest{}, err
	}
	for key, value := range input.Dynamic {
		metadata[key] = value
	}
	for _, raw := range input.Sets {
		path, value, err := metaoverride.SplitAssignment(raw)
		if err != nil {
			return PreparedRequest{}, err
		}
		if err := metaoverride.ApplyPathValue(metadata, path, metaoverride.InferValue(value)); err != nil {
			return PreparedRequest{}, err
		}
	}
	for _, raw := range input.SetJSONs {
		path, value, err := metaoverride.SplitAssignment(raw)
		if err != nil {
			return PreparedRequest{}, err
		}
		var decoded interface{}
		if err := json.Unmarshal([]byte(value), &decoded); err != nil {
			return PreparedRequest{}, fmt.Errorf("invalid JSON for %s: %w", path, err)
		}
		if err := metaoverride.ApplyPathValue(metadata, path, decoded); err != nil {
			return PreparedRequest{}, err
		}
	}

	if err := s.enrichCloudAccountMetadata(metadata); err != nil {
		return PreparedRequest{}, err
	}
	storageInfo, err := s.enrichStorageMetadata(metadata)
	if err != nil {
		return PreparedRequest{}, err
	}
	hostInfo, err := s.hostDetail(input.HostID)
	if err != nil {
		return PreparedRequest{}, err
	}
	s.enrichRepairHostMapperDefaults(metadata)
	s.enrichHostDerivedMetadata(metadata, hostInfo)
	driver := resolveApplyDriver(metadata, storageInfo)
	if err := driver.Enrich(s, &applyDriverSpec{
		HostID:      input.HostID,
		Metadata:    metadata,
		HostInfo:    hostInfo,
		StorageInfo: storageInfo,
		Dynamic:     input.Dynamic,
	}); err != nil {
		return PreparedRequest{}, err
	}
	s.enrichDefaultNICs(metadata)
	if err := s.enrichDiskVolumeMapper(metadata, hostInfo.Disks, storageInfo); err != nil {
		return PreparedRequest{}, err
	}

	if hostInfo.BootConfigID != "" {
		currentMetadata, err := s.bootConfigMetadata(hostInfo.BootConfigID)
		if err != nil {
			return PreparedRequest{}, err
		}
		diffMetadata := diffTopLevelMetadata(metadata, currentMetadata)
		return PreparedRequest{
			Operation:    "update",
			Path:         "/api/v2/batchUpdateBootConfigs",
			Metadata:     metadata,
			HostID:       input.HostID,
			BootConfigID: hostInfo.BootConfigID,
			NoOp:         len(diffMetadata) == 0,
			Body: map[string]interface{}{
				"batch_update": []map[string]interface{}{
					{
						"id":           hostInfo.BootConfigID,
						"migration_id": input.HostID,
						"metadata":     diffMetadata,
					},
				},
			},
		}, nil
	}
	return PreparedRequest{
		Operation: "create",
		Path:      "/api/v2/batchBootConfigs",
		Metadata:  metadata,
		HostID:    input.HostID,
		Body: map[string]interface{}{
			"batch_create": []map[string]interface{}{
				{
					"migration_id": input.HostID,
					"metadata":     metadata,
				},
			},
		},
	}, nil
}

func MutationView(result MutationResult) map[string]interface{} {
	if result.NoOp {
		view := map[string]interface{}{
			"operation": result.Operation,
			"no_op":     true,
		}
		if result.HostID != "" {
			view["migration_id"] = result.HostID
		}
		if result.BootConfigID != "" {
			view["boot_config_id"] = result.BootConfigID
		}
		return view
	}
	if data := mapFromData(result.Response.Data); data != nil {
		view := cloneMap(data)
		view["operation"] = result.Operation
		return view
	}
	if result.Response.Data == nil && result.Response.Raw != nil {
		view := cloneMap(result.Response.Raw)
		view["operation"] = result.Operation
		return view
	}
	return map[string]interface{}{"operation": result.Operation}
}

func metadataFile(input ApplyInput) (map[string]interface{}, error) {
	if input.File == "" {
		if len(input.Dynamic) == 0 && len(input.Sets) == 0 && len(input.SetJSONs) == 0 {
			return nil, fmt.Errorf("file is required when no metadata override flags are provided")
		}
		return map[string]interface{}{}, nil
	}
	return metaoverride.ReadObjectFile(input.File, "metadata object", map[string]string{
		"batch_create": "batch_create wrapper",
		"batch_update": "batch_update wrapper",
	})
}

func (s Service) hostDetail(hostID string) (hostDetail, error) {
	resp, err := s.api.Get("/api/v2/getHostDetail", queryFromPairs("host_id", hostID))
	if err != nil {
		return hostDetail{}, err
	}
	data := responseMap(resp)
	if data == nil {
		return hostDetail{}, fmt.Errorf("host %s did not return detail object", hostID)
	}
	return hostDetail{
		Data:         data,
		BootConfigID: findBootConfigID(data),
		Disks:        hostDisksFromDetail(data),
	}, nil
}

func (s Service) bootConfigMetadata(bootConfigID string) (map[string]interface{}, error) {
	resp, err := s.api.Post("/api/v2/batchGetBootConfigs", map[string]interface{}{
		"batch_get": []map[string]interface{}{
			{"id": bootConfigID},
		},
	})
	if err != nil {
		return nil, err
	}
	data := mapFromData(resp.Data)
	if data == nil {
		return nil, fmt.Errorf("boot config %s did not return detail object", bootConfigID)
	}
	rows := listMaps(data["boot_configs"])
	if len(rows) == 0 {
		return nil, fmt.Errorf("boot config %s did not return metadata", bootConfigID)
	}
	metadata, ok := rows[0]["metadata"].(map[string]interface{})
	if !ok || metadata == nil {
		return nil, fmt.Errorf("boot config %s did not return metadata object", bootConfigID)
	}
	return metadata, nil
}

func (s Service) enrichCloudAccountMetadata(metadata map[string]interface{}) error {
	accountID := mapStringValue(metadata["cloud_account_id"])
	if accountID == "" {
		return nil
	}
	resp, err := s.api.Get("/hypermotion/v1/cloud_accounts/"+url.PathEscape(accountID), url.Values{})
	if err != nil {
		return err
	}
	data := responseMap(resp)
	if data == nil {
		return nil
	}
	cloudTypeDisplayName := gatewayCloudTypeDisplayName(data)
	setMissingString(metadata, "cloud_type", gatewayAccountField(data, "cloud_type"))
	setMissingString(metadata, "cloud_type_name",
		firstNonEmptyString(
			gatewayAccountField(data, "cloud_type_name"),
			gatewayAccountField(data, "cloud_type_display_name"),
			cloudTypeDisplayName,
		),
	)
	setMissingString(metadata, "cloud_type_display_name",
		firstNonEmptyString(
			gatewayAccountField(data, "cloud_type_display_name"),
			cloudTypeDisplayName,
		),
	)
	setMissingString(metadata, "cloud_account_name",
		firstNonEmptyString(
			gatewayAccountField(data, "cloud_account_name"),
			gatewayAccountField(data, "account_name"),
			gatewayAccountField(data, "display_name"),
			gatewayAccountField(data, "custom_name"),
			gatewayAccountField(data, "name"),
		),
	)
	setMissingString(metadata, "cloud_account_username",
		firstNonEmptyString(
			gatewayAccountField(data, "cloud_account_username"),
			gatewayAccountField(data, "username"),
			gatewayAccountField(data, "auth_key"),
		),
	)
	setMissingString(metadata, "region_id", gatewayAccountRegionID(data))
	return nil
}

func (s Service) enrichStorageMetadata(metadata map[string]interface{}) (storageDefaults, error) {
	storageID := mapStringValue(metadata["storage_id"])
	if storageID == "" {
		return storageDefaults{}, nil
	}
	resp, err := s.api.Get("/api/v2/getStorageDetailInfo", queryFromPairs("storage_id", storageID))
	if err != nil {
		return storageDefaults{}, err
	}
	data := responseMap(resp)
	if data == nil {
		return storageDefaults{}, nil
	}
	detail := storageDetailRoot(data)
	flat := flattenMaps(data)
	detailFlat := flattenMaps(detail)
	poolID, poolName := objectStoragePool(detail, flat)
	storageName := sanitizeObjectStorageName(firstNonEmptyString(
		mapStringValue(detail["storage_name"]),
		mapStringValue(detail["name"]),
		mapStringValue(data["storage_name"]),
		mapStringValue(data["name"]),
	), poolID, poolName)
	info := storageDefaults{
		IsObjectStorage: isObjectStorage(mergeMaps(detailFlat, flat)),
		ZoneID: firstNonEmptyString(
			mapStringValue(detail["zone_id"]),
			mapStringValue(detail["auth_zone_id"]),
			mapStringValue(detailFlat["zone_id"]),
			mapStringValue(detailFlat["auth_zone_id"]),
			mapStringValue(flat["zone_id"]),
			mapStringValue(flat["auth_zone_id"]),
		),
		StorageName:            storageName,
		StorageDisplayName:     objectStorageDisplayName(detail, data, storageName, poolName),
		PoolID:                 poolID,
		PoolName:               poolName,
		DefaultPoolDisplayName: firstNonEmptyString(mapStringValue(detailFlat["default_pool_display_name"]), mapStringValue(flat["default_pool_display_name"])),
		PoolByVolumeType:       blockStoragePoolsByVolumeType(detail),
		StorageID:              storageID,
		IsBlockStorage:         isBlockStorage(detailFlat, flat),
	}
	info.WriteNetwork, info.WriteNetworkName = resolveStorageNetwork(detail, "network_addr_for_write_data", "network_addrs_for_write_data")
	if info.WriteNetwork == "" {
		info.WriteNetwork, info.WriteNetworkName = resolveStorageNetwork(flat, "network_addr_for_write_data", "network_addrs_for_write_data")
	}
	info.ReadNetwork, info.ReadNetworkName = resolveStorageNetwork(detail, "network_addr_for_read_data", "network_addrs_for_read_data")
	if info.ReadNetwork == "" {
		info.ReadNetwork, info.ReadNetworkName = resolveStorageNetwork(flat, "network_addr_for_read_data", "network_addrs_for_read_data")
	}
	if info.WriteNetwork == "" {
		info.WriteNetwork, info.WriteNetworkName = resolveStorageConfigNetwork(detail, "public_endpoint")
	}
	if info.ReadNetwork == "" {
		info.ReadNetwork, info.ReadNetworkName = resolveStorageConfigNetwork(detail, "internal_endpoint")
	}

	if info.IsObjectStorage {
		setMissingString(metadata, "storage_type", "objectstorage")
	}
	return info, nil
}

func (s Service) enrichObjectStorageCloudInfoMetadata(hostID string, metadata map[string]interface{}, hostData map[string]interface{}) error {
	accountID := mapStringValue(metadata["cloud_account_id"])
	cloudType := mapStringValue(metadata["cloud_type"])
	if accountID == "" || cloudType == "" {
		return nil
	}

	if mapStringValue(metadata["region_id"]) != "" || mapStringValue(metadata["zone_id"]) != "" {
		data, err := s.cloudInfo(hostID, metadata, "regions,zones")
		if err != nil {
			return err
		}
		regionID := mapStringValue(metadata["region_id"])
		if regionID != "" {
			row := findRegionRow(data, regionID)
			if row != nil {
				setMissingString(metadata, "region_name", resourceDisplayName(row, regionID))
			} else {
				setMissingString(metadata, "region_name", regionID)
			}
		}
		zoneID := mapStringValue(metadata["zone_id"])
		if zoneID != "" {
			row := findResourceRow(data, "zones", zoneID)
			if row != nil {
				setMissingString(metadata, "zone_name", resourceDisplayName(row, zoneID))
			} else {
				setMissingString(metadata, "zone_name", zoneID)
			}
		}
	}

	if mapStringValue(metadata["flavor_id"]) != "" {
		data, err := s.cloudInfo(hostID, metadata, "flavors")
		if err != nil {
			return err
		}
		flavorID := mapStringValue(metadata["flavor_id"])
		row := findResourceRow(data, "flavors", flavorID)
		if row != nil {
			setMissingString(metadata, "flavor_name", resourceDisplayName(row, flavorID))
			setMissingString(metadata, "flavors", firstNonEmptyString(mapStringValue(row["flavors"]), mapStringValue(row["value"]), "id-"+flavorID))
			setMissingValue(metadata, "flavor_vcpus", row["vcpus"])
			setMissingValue(metadata, "flavor_ram", firstNonNil(row["ram_GB"], row["ram"]))
			setMissingValue(metadata, "max_nic_num", row["max_nic_num"])
			setMissingString(metadata, "boot_loader_flavor_id", flavorID)
			setMissingString(metadata, "boot_loader_flavor_name", resourceDisplayName(row, flavorID))
		} else {
			setMissingString(metadata, "flavor_name", flavorID)
			setMissingString(metadata, "flavors", "id-"+flavorID)
			setMissingString(metadata, "boot_loader_flavor_id", flavorID)
			setMissingString(metadata, "boot_loader_flavor_name", flavorID)
		}
	}

	if mapStringValue(metadata["os_type_id"]) != "" || inferOSFamily(metadata, hostData) != "" {
		data, err := s.cloudInfo(hostID, metadata, "os_types")
		if err != nil {
			return err
		}
		osTypeID := mapStringValue(metadata["os_type_id"])
		if osTypeID == "" {
			switch inferOSFamily(metadata, hostData) {
			case "linux":
				osTypeID = "id-Linux"
			case "windows":
				osTypeID = "id-Windows"
			}
			setMissingString(metadata, "os_type_id", osTypeID)
		}
		if osTypeID != "" {
			row := findResourceRow(data, "os_types", osTypeID)
			if row != nil {
				setMissingString(metadata, "os_type_name", resourceOSTypeName(row, osTypeID))
				setMissingString(metadata, "os_type", firstNonEmptyString(mapStringValue(row["os_type"]), inferOSFamilyValue(osTypeID)))
			} else {
				setMissingString(metadata, "os_type_name", defaultOSTypeName(osTypeID))
				setMissingString(metadata, "os_type", inferOSFamilyValue(osTypeID))
			}
		}
	}

	if mapStringValue(metadata["volume_type_id"]) != "" ||
		mapStringValue(metadata["system_volume_type_id"]) != "" ||
		mapStringValue(metadata["default_volume_type_id"]) != "" ||
		missingObjectStorageVolumeTypeDefaults(metadata) {
		data, err := s.cloudInfo(hostID, metadata, "system_volume_types,volume_types")
		if err != nil {
			return err
		}
		volumeTypeID := firstNonEmptyString(
			mapStringValue(metadata["volume_type_id"]),
			preferredResourceID(data, "volume_types"),
			preferredResourceID(data, "system_volume_types"),
		)
		systemVolumeTypeID := firstNonEmptyString(
			mapStringValue(metadata["system_volume_type_id"]),
			preferredResourceID(data, "system_volume_types"),
			volumeTypeID,
		)
		defaultVolumeTypeID := firstNonEmptyString(
			mapStringValue(metadata["default_volume_type_id"]),
			volumeTypeID,
			systemVolumeTypeID,
		)
		setMissingString(metadata, "volume_type_id", volumeTypeID)
		setMissingString(metadata, "system_volume_type_id", systemVolumeTypeID)
		setMissingString(metadata, "default_volume_type_id", defaultVolumeTypeID)
		if volumeTypeID != "" {
			row := findResourceRow(data, "volume_types", volumeTypeID)
			if row == nil {
				row = findResourceRow(data, "system_volume_types", volumeTypeID)
			}
			if row != nil {
				setMissingString(metadata, "volume_type_name", resourceStableName(row, volumeTypeID))
				setMissingString(metadata, "volume_type_display_name", resourceDisplayName(row, volumeTypeID))
			} else {
				setMissingString(metadata, "volume_type_name", volumeTypeID)
				setMissingString(metadata, "volume_type_display_name", volumeTypeID)
			}
		}
		if systemVolumeTypeID != "" {
			row := findResourceRow(data, "system_volume_types", systemVolumeTypeID)
			if row == nil {
				row = findResourceRow(data, "volume_types", systemVolumeTypeID)
			}
			if row != nil {
				setMissingString(metadata, "system_volume_type_name", resourceStableName(row, systemVolumeTypeID))
				setMissingString(metadata, "system_volume_type_display_name", resourceDisplayName(row, systemVolumeTypeID))
			} else {
				setMissingString(metadata, "system_volume_type_name", systemVolumeTypeID)
				setMissingString(metadata, "system_volume_type_display_name", systemVolumeTypeID)
			}
		}
		if defaultVolumeTypeID != "" {
			row := findResourceRow(data, "volume_types", defaultVolumeTypeID)
			if row == nil {
				row = findResourceRow(data, "system_volume_types", defaultVolumeTypeID)
			}
			if row != nil {
				setMissingString(metadata, "default_volume_type_name", resourceStableName(row, defaultVolumeTypeID))
				setMissingString(metadata, "default_volume_type_display_name", resourceDisplayName(row, defaultVolumeTypeID))
			} else {
				setMissingString(metadata, "default_volume_type_name", defaultVolumeTypeID)
				setMissingString(metadata, "default_volume_type_display_name", defaultVolumeTypeID)
			}
		}
	}

	if mapStringValue(metadata["network_id"]) != "" || mapStringValue(metadata["security_group_id"]) != "" {
		data, err := s.cloudInfo(hostID, metadata, "networks,security_groups")
		if err != nil {
			return err
		}
		networkID := mapStringValue(metadata["network_id"])
		if networkID != "" {
			row := findResourceRow(data, "networks", networkID)
			if row != nil {
				setMissingString(metadata, "network_name", firstNonEmptyString(mapStringValue(row["name"]), networkID))
				setMissingString(metadata, "network_display_name", firstNonEmptyString(mapStringValue(row["display_name"]), mapStringValue(row["name"]), networkID))
			} else {
				setMissingString(metadata, "network_name", networkID)
				setMissingString(metadata, "network_display_name", networkID)
			}
		}
		securityGroupID := mapStringValue(metadata["security_group_id"])
		if securityGroupID != "" {
			row := findResourceRow(data, "security_groups", securityGroupID)
			if row != nil {
				setMissingString(metadata, "security_group_name", firstNonEmptyString(mapStringValue(row["name"]), securityGroupID))
				setMissingString(metadata, "security_group_display_name", firstNonEmptyString(mapStringValue(row["display_name"]), mapStringValue(row["name"]), securityGroupID))
			} else {
				setMissingString(metadata, "security_group_name", securityGroupID)
				setMissingString(metadata, "security_group_display_name", securityGroupID)
			}
		}
	}

	return nil
}

func missingObjectStorageVolumeTypeDefaults(metadata map[string]interface{}) bool {
	return mapStringValue(metadata["volume_type_id"]) == "" &&
		mapStringValue(metadata["system_volume_type_id"]) == "" &&
		mapStringValue(metadata["default_volume_type_id"]) == ""
}

func (s Service) enrichObjectStorageSubnetMetadata(metadata map[string]interface{}) error {
	accountID := mapStringValue(metadata["cloud_account_id"])
	cloudType := mapStringValue(metadata["cloud_type"])
	regionID := mapStringValue(metadata["region_id"])
	zoneID := mapStringValue(metadata["zone_id"])
	networkID := mapStringValue(metadata["network_id"])
	subnetID := mapStringValue(metadata["subnet_id"])
	if accountID == "" || cloudType == "" || regionID == "" || zoneID == "" || networkID == "" || subnetID == "" {
		return nil
	}
	resp, err := s.api.Get("/api/v3/getSubnetConfig", queryFromPairs(
		"cloud_account_id", accountID,
		"cloud_type", cloudType,
		"region_id", regionID,
		"zone_id", zoneID,
		"network_id", networkID,
		"subnet_id", subnetID,
	))
	if err != nil {
		return err
	}
	data := responseMap(resp)
	if data == nil {
		return nil
	}
	row := findSubnetRow(data, subnetID)
	if row == nil {
		return nil
	}
	setMissingString(metadata, "subnet_name", firstNonEmptyString(mapStringValue(row["name"]), subnetID))
	setMissingString(metadata, "subnet_display_name", firstNonEmptyString(mapStringValue(row["display_name"]), mapStringValue(row["name"]), subnetID))
	return nil
}

func (s Service) enrichBlockStorageNetworkMetadata(hostID string, metadata map[string]interface{}, dynamic map[string]string) error {
	accountID := mapStringValue(metadata["cloud_account_id"])
	cloudType := mapStringValue(metadata["cloud_type"])
	regionID := mapStringValue(metadata["region_id"])
	zoneID := mapStringValue(metadata["zone_id"])
	subnetID := mapStringValue(metadata["subnet_id"])
	_, explicitNetwork := dynamic["network_id"]
	_, explicitSubnet := dynamic["subnet_id"]
	_, explicitSecurityGroup := dynamic["security_group_id"]
	if accountID == "" || cloudType == "" || regionID == "" || zoneID == "" {
		return nil
	}

	if subnetID != "" || (!explicitSubnet && !explicitNetwork && !explicitSecurityGroup) {
		resp, err := s.api.Get("/api/v3/getSubnetConfig", queryFromPairs(
			"cloud_account_id", accountID,
			"cloud_type", cloudType,
			"region_id", regionID,
			"zone_id", zoneID,
			"subnet_id", subnetID,
		))
		if err != nil {
			return err
		}
		data := responseMap(resp)
		if data == nil {
			if subnetID != "" && !explicitSubnet {
				return fmt.Errorf("subnet %s is not available in zone %s", subnetID, zoneID)
			}
			return nil
		}
		row := findSubnetRow(data, subnetID)
		if row == nil && subnetID == "" {
			subnets := listMaps(data["subnets"])
			if len(subnets) > 0 {
				row = subnets[0]
			}
		}
		if row == nil {
			if subnetID != "" && !explicitSubnet {
				return fmt.Errorf("subnet %s is not available in zone %s", subnetID, zoneID)
			}
			return nil
		}
		subnetNetworkID := firstNonEmptyString(mapStringValue(row["network_id"]), mapStringValue(row["vpc_id"]))
		if subnetNetworkID == "" {
			if !explicitSubnet && subnetID != "" {
				return fmt.Errorf("subnet %s did not return network_id", subnetID)
			}
			return nil
		}
		if currentSubnetID := mapStringValue(metadata["subnet_id"]); currentSubnetID == "" {
			metadata["subnet_id"] = firstNonEmptyString(mapStringValue(row["id"]), mapStringValue(row["subnet_id"]))
		}
		if currentNetworkID := mapStringValue(metadata["network_id"]); currentNetworkID != "" && currentNetworkID != subnetNetworkID && !explicitNetwork {
			return fmt.Errorf("subnet %s belongs to network %s, not %s", subnetID, subnetNetworkID, currentNetworkID)
		}
		if !explicitNetwork || mapStringValue(metadata["network_id"]) == "" {
			metadata["network_id"] = subnetNetworkID
		}
		setMissingString(metadata, "subnet_name", firstNonEmptyString(mapStringValue(row["name"]), subnetID))
		setMissingString(metadata, "subnet_display_name", firstNonEmptyString(mapStringValue(row["display_name"]), mapStringValue(row["name"]), subnetID))
	}

	if mapStringValue(metadata["network_id"]) == "" && mapStringValue(metadata["security_group_id"]) == "" {
		return nil
	}
	data, err := s.blockStorageCloudInfo(hostID, metadata, "networks,security_groups")
	if err != nil {
		return err
	}
	networkID := mapStringValue(metadata["network_id"])
	if networkID != "" {
		row := findResourceRow(data, "networks", networkID)
		if row == nil && !explicitNetwork {
			return fmt.Errorf("network %s is not available in zone %s", networkID, zoneID)
		}
		if row != nil {
			setMissingString(metadata, "network_name", firstNonEmptyString(mapStringValue(row["name"]), networkID))
			setMissingString(metadata, "network_display_name", firstNonEmptyString(mapStringValue(row["display_name"]), mapStringValue(row["name"]), networkID))
		}
	}
	securityGroupID := mapStringValue(metadata["security_group_id"])
	if securityGroupID == "" && !explicitSecurityGroup {
		for _, row := range cloudinfo.ResourceRows(data, "security_groups") {
			sgNetworkID := firstNonEmptyString(
				mapStringValue(row["network_id"]),
				mapStringValue(row["vpc_id"]),
				mapStringValue(row["network_uuid"]),
				mapStringValue(row["vpc_uuid"]),
			)
			if networkID != "" && sgNetworkID != "" && sgNetworkID != networkID {
				continue
			}
			metadata["security_group_id"] = firstNonEmptyString(mapStringValue(row["id"]), mapStringValue(row["security_group_id"]))
			setMissingString(metadata, "security_group_name", firstNonEmptyString(mapStringValue(row["name"]), mapStringValue(row["id"])))
			setMissingString(metadata, "security_group_display_name", firstNonEmptyString(mapStringValue(row["display_name"]), mapStringValue(row["name"]), mapStringValue(row["id"])))
			return nil
		}
		return nil
	}
	if securityGroupID == "" {
		return nil
	}
	row := findResourceRow(data, "security_groups", securityGroupID)
	if row == nil && !explicitSecurityGroup {
		return fmt.Errorf("security group %s is not available for network %s", securityGroupID, networkID)
	}
	if row == nil {
		return nil
	}
	sgNetworkID := firstNonEmptyString(
		mapStringValue(row["network_id"]),
		mapStringValue(row["vpc_id"]),
		mapStringValue(row["network_uuid"]),
		mapStringValue(row["vpc_uuid"]),
	)
	if networkID != "" && sgNetworkID != "" && sgNetworkID != networkID && !explicitSecurityGroup && !explicitNetwork {
		return fmt.Errorf("security group %s belongs to network %s, not %s", securityGroupID, sgNetworkID, networkID)
	}
	setMissingString(metadata, "security_group_name", firstNonEmptyString(mapStringValue(row["name"]), securityGroupID))
	setMissingString(metadata, "security_group_display_name", firstNonEmptyString(mapStringValue(row["display_name"]), mapStringValue(row["name"]), securityGroupID))
	return nil
}

func (s Service) enrichHostDerivedMetadata(metadata map[string]interface{}, hostInfo hostDetail) {
	osFamily := inferOSFamily(metadata, hostInfo.Data)
	if mapStringValue(metadata["os_type_id"]) == "" {
		switch osFamily {
		case "linux":
			metadata["os_type_id"] = "id-Linux"
		case "windows":
			metadata["os_type_id"] = "id-Windows"
		}
	}
	if mapStringValue(metadata["os_type"]) == "" {
		if osFamily != "" {
			metadata["os_type"] = inferOSFamilyTitle(osFamily)
		} else {
			metadata["os_type"] = inferOSFamilyValue(mapStringValue(metadata["os_type_id"]))
		}
	}
	if mapStringValue(metadata["os_type_name"]) == "" && mapStringValue(metadata["os_type_id"]) != "" {
		metadata["os_type_name"] = defaultOSTypeName(mapStringValue(metadata["os_type_id"]))
	}
	if mapStringValue(metadata["arch"]) == "" {
		if arch := inferHostArch(hostInfo.Data); arch != "" {
			metadata["arch"] = arch
		}
	}
	if mapStringValue(metadata["dest_boot_mode"]) == "" {
		if bootMode := inferHostBootMode(hostInfo.Data); bootMode != "" {
			metadata["dest_boot_mode"] = bootMode
		}
	}
	if bootMode := mapStringValue(metadata["dest_boot_mode"]); bootMode != "" {
		setMissingString(metadata, "dest_boot_mode_name", bootModeDisplayName(bootMode))
	}
	if flavorID := mapStringValue(metadata["flavor_id"]); flavorID != "" {
		setMissingString(metadata, "boot_loader_flavor_id", flavorID)
	}
}

func (s Service) enrichRepairHostMapperDefaults(metadata map[string]interface{}) {
	defaults := map[string]interface{}{
		"enable_dhcp_mode":     "1",
		"enable_inject_driver": "1",
		"enable_repair_fs":     "1",
		"os_version":           "auto_check",
		"os_display_name":      "auto_check",
		"pre_script":           "",
		"post_script":          "",
	}
	current, _ := metadata["repair_host_mapper"].(map[string]interface{})
	if current == nil {
		current = map[string]interface{}{}
		metadata["repair_host_mapper"] = current
	}
	for key, value := range defaults {
		if _, exists := current[key]; !exists {
			current[key] = value
		}
	}
}

func (s Service) enrichDefaultNICs(metadata map[string]interface{}) {
	if !shouldGenerateDefaultNIC(metadata) {
		return
	}
	rows := ensureMapperRowsForKey(metadata, "nics", 1)
	row := ensureRowMap(rows[0])
	rows[0] = row

	setMissingInt(row, "index", 0)
	setMissingString(row, "network_id", mapStringValue(metadata["network_id"]))
	setMissingString(row, "network_name", mapStringValue(metadata["network_name"]))
	setMissingString(row, "network_display_name", mapStringValue(metadata["network_display_name"]))
	setMissingString(row, "subnet_id", mapStringValue(metadata["subnet_id"]))
	setMissingString(row, "subnet_name", mapStringValue(metadata["subnet_name"]))
	setMissingString(row, "subnet_display_name", mapStringValue(metadata["subnet_display_name"]))
	setDefaultString(row, "fixed_ip", mapStringValue(metadata["fixed_ip"]))
	setDefaultString(row, "floating_ip", mapStringValue(metadata["floating_ip"]))
	setDefaultString(row, "is_fixed_ip_input", firstNonEmptyString(mapStringValue(metadata["is_fixed_ip_input"]), "false"))
	setDefaultString(row, "is_fixed_ip_select", firstNonEmptyString(mapStringValue(metadata["is_fixed_ip_select"]), "false"))
	setDefaultString(row, "fixed_ip_id", mapStringValue(metadata["fixed_ip_id"]))
	setDefaultString(row, "floating_ip_id", mapStringValue(metadata["floating_ip_id"]))
	setDefaultString(row, "floating_ip_address", mapStringValue(metadata["floating_ip_address"]))
	setDefaultString(row, "bandwidth_id", mapStringValue(metadata["bandwidth_id"]))
	setDefaultString(row, "bandwidth_name", mapStringValue(metadata["bandwidth_name"]))
	setDefaultString(row, "bandwidth_size", mapStringValue(metadata["bandwidth_size"]))

	securityGroups, _ := row["security_groups"].([]interface{})
	if len(securityGroups) == 0 && (mapStringValue(metadata["security_group_id"]) != "" || mapStringValue(metadata["security_group_name"]) != "") {
		securityGroups = []interface{}{map[string]interface{}{}}
	}
	if len(securityGroups) > 0 {
		sgRow := ensureRowMap(securityGroups[0])
		securityGroups[0] = sgRow
		setMissingString(sgRow, "id", mapStringValue(metadata["security_group_id"]))
		setMissingString(sgRow, "name", firstNonEmptyString(mapStringValue(metadata["security_group_name"]), mapStringValue(metadata["security_group_display_name"]), mapStringValue(metadata["security_group_id"])))
		row["security_groups"] = securityGroups
	}

	metadata["nics"] = rows
}

func (s Service) enrichDiskVolumeMapper(metadata map[string]interface{}, disks []hostDisk, storageInfo storageDefaults) error {
	systemVolumeTypeID := mapStringValue(metadata["system_volume_type_id"])
	volumeTypeID := mapStringValue(metadata["volume_type_id"])
	if systemVolumeTypeID == "" && volumeTypeID == "" {
		return nil
	}
	if len(disks) == 0 {
		return nil
	}

	rows := ensureMapperRows(metadata, len(disks))
	defaultVolumeTypeID := firstNonEmptyString(mapStringValue(metadata["default_volume_type_id"]), volumeTypeID, systemVolumeTypeID)
	defaultVolumeTypeName := firstNonEmptyString(
		mapStringValue(metadata["default_volume_type_name"]),
		defaultVolumeTypeID,
	)
	defaultVolumeTypeDisplayName := mapStringValue(metadata["default_volume_type_display_name"])
	poolID := firstNonEmptyString(mapStringValue(metadata["default_pool_id"]), storageInfo.PoolID)
	poolName := firstNonEmptyString(mapStringValue(metadata["default_pool_name"]), storageInfo.PoolName, poolID)
	for idx, disk := range disks {
		row := ensureRowMap(rows[idx])
		rows[idx] = row
		selectedTypeID := mapStringValue(row["volume_type_id"])
		if selectedTypeID == "" {
			selectedTypeID = volumeTypeID
			if disk.IsBootDisk && systemVolumeTypeID != "" {
				selectedTypeID = systemVolumeTypeID
			}
		}
		selectedTypeName := selectedTypeID
		selectedTypeDisplayName := ""
		if selectedTypeID == firstNonEmptyString(mapStringValue(metadata["system_volume_type_id"]), volumeTypeID) && disk.IsBootDisk && systemVolumeTypeID != "" {
			selectedTypeName = firstNonEmptyString(mapStringValue(metadata["system_volume_type_name"]), selectedTypeID)
			selectedTypeDisplayName = mapStringValue(metadata["system_volume_type_display_name"])
		} else {
			selectedTypeName = firstNonEmptyString(mapStringValue(metadata["volume_type_name"]), selectedTypeID)
			selectedTypeDisplayName = mapStringValue(metadata["volume_type_display_name"])
		}
		setMissingInt(row, "index", disk.Index)
		setMissingString(row, "disk_id", disk.DiskID)
		setMissingBool(row, "is_boot_disk", disk.IsBootDisk)
		if storageInfo.IsBlockStorage {
			if mapStringValue(row["pool_id"]) == "" && mapStringValue(row["pool_name"]) == "" {
				pool, err := blockStoragePoolForVolumeType(storageInfo, selectedTypeID)
				if err != nil {
					return fmt.Errorf("disk_volume_mapper[%d]: %w", idx, err)
				}
				setMissingString(row, "pool_id", pool.ID)
				setMissingString(row, "pool_name", pool.Name)
			}
		} else {
			setMissingString(row, "pool_id", poolID)
			setMissingString(row, "pool_name", poolName)
		}
		setMissingString(row, "default_volume_type_id", defaultVolumeTypeID)
		setMissingString(row, "default_volume_type_name", defaultVolumeTypeName)
		setMissingString(row, "default_volume_type_display_name", defaultVolumeTypeDisplayName)
		setMissingString(row, "volume_type_id", selectedTypeID)
		setMissingString(row, "volume_type_name", selectedTypeName)
		setMissingString(row, "volume_type_display_name", selectedTypeDisplayName)
	}
	metadata["disk_volume_mapper"] = rows
	return nil
}

func blockStoragePoolForVolumeType(storageInfo storageDefaults, volumeTypeID string) (poolRef, error) {
	if volumeTypeID == "" {
		return poolRef{}, fmt.Errorf("block storage %s missing volume type for disk_volume_mapper row", storageInfo.StorageID)
	}
	if pool, ok := storageInfo.PoolByVolumeType[volumeTypeID]; ok {
		return pool, nil
	}
	return poolRef{}, fmt.Errorf("block storage %s has no storage pool matching volume type %s", storageInfo.StorageID, volumeTypeID)
}

func ensureMapperRows(metadata map[string]interface{}, required int) []interface{} {
	return ensureMapperRowsForKey(metadata, "disk_volume_mapper", required)
}

func ensureMapperRowsForKey(metadata map[string]interface{}, key string, required int) []interface{} {
	existing, _ := metadata[key].([]interface{})
	if len(existing) >= required {
		return existing
	}
	rows := make([]interface{}, 0, required)
	rows = append(rows, existing...)
	for len(rows) < required {
		rows = append(rows, map[string]interface{}{})
	}
	return rows
}

func shouldGenerateDefaultNIC(metadata map[string]interface{}) bool {
	return mapStringValue(metadata["network_id"]) != "" ||
		mapStringValue(metadata["subnet_id"]) != "" ||
		mapStringValue(metadata["security_group_id"]) != "" ||
		mapStringValue(metadata["bandwidth_size"]) != ""
}

func ensureRowMap(value interface{}) map[string]interface{} {
	if row, ok := value.(map[string]interface{}); ok && row != nil {
		return row
	}
	return map[string]interface{}{}
}

func hostDisksFromDetail(data map[string]interface{}) []hostDisk {
	var rawRows []map[string]interface{}
	for _, key := range []string{"disks", "disk_infos", "disk_info"} {
		if rows := listMaps(data[key]); len(rows) > 0 {
			rawRows = rows
			break
		}
	}
	disks := make([]hostDisk, 0, len(rawRows))
	hasExplicitBoot := false
	for idx, row := range rawRows {
		diskID := firstNonEmptyString(
			mapStringValue(row["disk_id"]),
			mapStringValue(row["id"]),
			mapStringValue(row["uuid"]),
		)
		if diskID == "" {
			continue
		}
		isBoot := mapBoolValue(row["is_boot_disk"]) || mapBoolValue(row["is_boot"]) || mapBoolValue(row["boot"]) || mapBoolValue(row["system_disk"])
		if isBoot {
			hasExplicitBoot = true
		}
		disks = append(disks, hostDisk{
			Index:      mapIntValue(row["index"], idx),
			DiskID:     diskID,
			IsBootDisk: isBoot,
		})
	}
	if len(disks) > 0 && !hasExplicitBoot {
		disks[0].IsBootDisk = true
	}
	return disks
}

func findBootConfigID(data map[string]interface{}) string {
	if bootConfigID, _ := data["boot_config_id"].(string); bootConfigID != "" {
		return bootConfigID
	}
	bootConfig, ok := data["boot_config"].(map[string]interface{})
	if !ok || bootConfig == nil {
		return ""
	}
	bootConfigID, _ := bootConfig["id"].(string)
	return bootConfigID
}

func resolveStorageNetwork(flat map[string]interface{}, primaryKey, listKey string) (string, string) {
	value := mapStringValue(flat[primaryKey])
	name := mapStringValue(flat[primaryKey+"_name"])
	if value != "" && name != "" {
		return value, name
	}
	if value == "" {
		value, name = networkCandidate(flat[listKey])
	} else if name == "" {
		_, name = networkCandidate(flat[listKey])
		if name == "" {
			name = value
		}
	}
	return value, name
}

func networkCandidate(value interface{}) (string, string) {
	switch typed := value.(type) {
	case string:
		if typed == "" {
			return "", ""
		}
		return typed, typed
	case []interface{}:
		for _, item := range typed {
			if key, name := networkCandidate(item); key != "" {
				return key, name
			}
		}
	case map[string]interface{}:
		key := firstNonEmptyString(
			mapStringValue(typed["value"]),
			mapStringValue(typed["name"]),
			mapStringValue(typed["network_addr"]),
			mapStringValue(typed["id"]),
		)
		name := firstNonEmptyString(
			mapStringValue(typed["display_name"]),
			mapStringValue(typed["label"]),
			mapStringValue(typed["name"]),
			mapStringValue(typed["value"]),
			key,
		)
		return key, name
	}
	return "", ""
}

func gatewayAccountField(data map[string]interface{}, key string) string {
	if value := mapStringValue(data[key]); value != "" {
		return value
	}
	if account, ok := data["cloud_account"].(map[string]interface{}); ok {
		if value := mapStringValue(account[key]); value != "" {
			return value
		}
	}
	return ""
}

func gatewayAccountRegionID(data map[string]interface{}) string {
	return firstNonEmptyString(
		mapStringValue(data["region_id"]),
		mapStringValue(data["auth_region_id"]),
		mapStringValue(nestedAccountValue(data, "region_id")),
		mapStringValue(nestedAccountValue(data, "region_type_list")),
		mapStringValue(nestedAccountValue(data, "auth_region_id")),
		mapStringValue(nestedMetadataValue(data, "region_type_list")),
		mapStringValue(nestedMetadataValue(data, "auth_region_id")),
	)
}

func gatewayCloudTypeDisplayName(data map[string]interface{}) string {
	if value := gatewayAccountField(data, "cloud_type_display_name"); value != "" {
		return value
	}
	customName := gatewayAccountField(data, "custom_name")
	regionName := firstNonEmptyString(
		gatewayAccountField(data, "region_name"),
		gatewayAccountField(data, "display_region_name"),
	)
	if customName != "" && regionName != "" {
		suffix := "-" + regionName
		if strings.HasSuffix(customName, suffix) {
			return strings.TrimSuffix(customName, suffix)
		}
	}
	return ""
}

func nestedAccountValue(data map[string]interface{}, key string) interface{} {
	if account, ok := data["cloud_account"].(map[string]interface{}); ok {
		return account[key]
	}
	return nil
}

func nestedMetadataValue(data map[string]interface{}, key string) interface{} {
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		if value := metadata[key]; value != nil {
			return value
		}
	}
	if account, ok := data["cloud_account"].(map[string]interface{}); ok {
		if metadata, ok := account["metadata"].(map[string]interface{}); ok {
			return metadata[key]
		}
	}
	return nil
}

func queryFromPairs(pairs ...string) url.Values {
	q := url.Values{}
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i+1] != "" {
			q.Set(pairs[i], pairs[i+1])
		}
	}
	return q
}

func mapFromData(data interface{}) map[string]interface{} {
	if mapped, ok := data.(map[string]interface{}); ok {
		return mapped
	}
	return nil
}

func diffTopLevelMetadata(next, current map[string]interface{}) map[string]interface{} {
	diff := map[string]interface{}{}
	for key, nextValue := range next {
		if !jsonValueEqual(nextValue, current[key]) {
			diff[key] = nextValue
		}
	}
	return diff
}

func jsonValueEqual(left, right interface{}) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return reflect.DeepEqual(left, right)
	}
	return bytes.Equal(leftJSON, rightJSON)
}

func cloneMap(src map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(src)+1)
	for key, value := range src {
		out[key] = value
	}
	return out
}

func responseMap(resp client.APIResponse) map[string]interface{} {
	if data := mapFromData(resp.Data); data != nil {
		return data
	}
	if resp.Raw != nil {
		return resp.Raw
	}
	return nil
}

func flattenMaps(root map[string]interface{}) map[string]interface{} {
	if root == nil {
		return map[string]interface{}{}
	}
	out := map[string]interface{}{}
	var walk func(map[string]interface{})
	walk = func(current map[string]interface{}) {
		for key, value := range current {
			if _, exists := out[key]; !exists {
				out[key] = value
			}
			if nested, ok := value.(map[string]interface{}); ok {
				walk(nested)
			}
		}
	}
	walk(root)
	return out
}

func mergeMaps(primary, secondary map[string]interface{}) map[string]interface{} {
	out := cloneMap(secondary)
	for key, value := range primary {
		out[key] = value
	}
	return out
}

func listMaps(value interface{}) []map[string]interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}
	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]interface{}); ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func storageDetailRoot(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return map[string]interface{}{}
	}
	if storage, ok := data["storage"].(map[string]interface{}); ok && storage != nil {
		return storage
	}
	if storages := listMaps(data["storages"]); len(storages) > 0 {
		return storages[0]
	}
	return data
}

func blockStoragePoolsByVolumeType(detail map[string]interface{}) map[string]poolRef {
	pools := listMaps(detail["storage_pools"])
	if len(pools) == 0 {
		return nil
	}
	byType := make(map[string]poolRef, len(pools))
	for _, pool := range pools {
		name := mapStringValue(pool["name"])
		if name == "" {
			continue
		}
		byType[name] = poolRef{
			ID:          firstNonEmptyString(mapStringValue(pool["uuid"]), mapStringValue(pool["id"])),
			Name:        name,
			DisplayName: firstNonEmptyString(mapStringValue(pool["display_name"]), name),
		}
	}
	return byType
}

func objectStoragePool(detail, flat map[string]interface{}) (string, string) {
	poolID := firstNonEmptyString(
		mapStringValue(detail["default_pool_id"]),
		mapStringValue(detail["pool_id"]),
		mapStringValue(flat["default_pool_id"]),
		mapStringValue(flat["pool_id"]),
	)
	poolName := firstNonEmptyString(
		mapStringValue(detail["default_pool_name"]),
		mapStringValue(detail["pool_name"]),
		mapStringValue(flat["default_pool_name"]),
		mapStringValue(flat["pool_name"]),
	)
	if poolID != "" || poolName != "" {
		return poolID, firstNonEmptyString(poolName, poolID)
	}
	pools := listMaps(detail["storage_pools"])
	if len(pools) == 0 {
		return "", ""
	}
	firstPool := pools[0]
	poolID = firstNonEmptyString(
		mapStringValue(firstPool["uuid"]),
		mapStringValue(firstPool["id"]),
	)
	poolName = firstNonEmptyString(
		mapStringValue(firstPool["name"]),
		mapStringValue(firstPool["display_name"]),
		poolID,
	)
	return poolID, poolName
}

func sanitizeObjectStorageName(value, poolID, poolName string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	lowered := strings.ToLower(trimmed)
	if lowered == "pool uuid" || trimmed == poolID || trimmed == poolName {
		return ""
	}
	return trimmed
}

func objectStorageDisplayName(detail, root map[string]interface{}, storageName, poolName string) string {
	base := firstNonEmptyString(
		mapStringValue(detail["storage_display_name"]),
		mapStringValue(detail["display_name"]),
		mapStringValue(root["storage_display_name"]),
		mapStringValue(root["display_name"]),
	)
	if base == "" {
		return ""
	}
	if poolName == "" {
		return base
	}
	if storageName != "" {
		return base
	}
	if strings.Contains(base, "("+poolName+")") {
		return base
	}
	return base + "(" + poolName + ")"
}

func resolveStorageConfigNetwork(detail map[string]interface{}, endpointKey string) (string, string) {
	config, _ := detail["config"].(map[string]interface{})
	if config == nil {
		return "", ""
	}
	endpoint := mapStringValue(config[endpointKey])
	if endpoint == "" {
		return "", ""
	}
	return endpointKey, endpoint
}

func isObjectStorage(flat map[string]interface{}) bool {
	for _, key := range []string{"type", "storage_type"} {
		switch normalizeStorageType(mapStringValue(flat[key])) {
		case "objectstorage":
			return true
		case "hypergate", "blockstorage":
			return false
		}
	}
	return false
}

func isBlockStorage(detailFlat, flat map[string]interface{}) bool {
	for _, key := range []string{"type", "storage_type"} {
		switch normalizeStorageType(firstNonEmptyString(mapStringValue(detailFlat[key]), mapStringValue(flat[key]))) {
		case "hypergate", "blockstorage":
			return true
		case "objectstorage":
			return false
		}
	}
	return len(listMaps(detailFlat["storage_pools"])) > 0 || len(listMaps(flat["storage_pools"])) > 0
}

func isObjectStorageWorkflow(metadata map[string]interface{}, storageInfo storageDefaults) bool {
	if storageInfo.IsObjectStorage {
		return true
	}
	if normalizeStorageType(mapStringValue(metadata["storage_type"])) == "objectstorage" {
		return true
	}
	return strings.HasSuffix(strings.ToLower(mapStringValue(metadata["cloud_type"])), "_obs")
}

func isAliyunBlockStorageWorkflow(metadata map[string]interface{}, flat map[string]interface{}) bool {
	if strings.EqualFold(mapStringValue(metadata["cloud_type"]), "aliyun_bs") {
		return true
	}
	return strings.EqualFold(mapStringValue(flat["cloud_type"]), "aliyun_bs")
}

func normalizeStorageType(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "_", "")
	return normalized
}

func setMissingString(target map[string]interface{}, key, value string) {
	if value == "" {
		return
	}
	if existing := mapStringValue(target[key]); existing != "" {
		return
	}
	target[key] = value
}

func setDefaultString(target map[string]interface{}, key, value string) {
	if _, exists := target[key]; exists {
		return
	}
	target[key] = value
}

func setMissingValue(target map[string]interface{}, key string, value interface{}) {
	if value == nil {
		return
	}
	if _, exists := target[key]; exists {
		return
	}
	target[key] = value
}

func setMissingInt(target map[string]interface{}, key string, value int) {
	if _, exists := target[key]; exists {
		return
	}
	target[key] = value
}

func setMissingBool(target map[string]interface{}, key string, value bool) {
	if _, exists := target[key]; exists {
		return
	}
	target[key] = value
}

func mapStringValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func mapBoolValue(value interface{}) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return typed == "1" || strings.EqualFold(typed, "true")
	case float64:
		return typed != 0
	case int:
		return typed != 0
	default:
		return false
	}
}

func mapIntValue(value interface{}, fallback int) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(typed)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstNonNil(values ...interface{}) interface{} {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func (s Service) cloudInfo(hostID string, metadata map[string]interface{}, fetchRes string) (map[string]interface{}, error) {
	accountID := mapStringValue(metadata["cloud_account_id"])
	cloudType := mapStringValue(metadata["cloud_type"])
	if accountID == "" || cloudType == "" {
		return nil, nil
	}
	q := queryFromPairs(
		"cloud_account_id", accountID,
		"cloud_type", cloudType,
		"storage_type", "objectstorage",
		"fetch_res", fetchRes,
		"host_id", hostID,
		"storage_id", mapStringValue(metadata["storage_id"]),
		"network_addr_for_write_data", mapStringValue(metadata["network_addr_for_write_data"]),
		"network_addr_for_read_data", mapStringValue(metadata["network_addr_for_read_data"]),
		"region_id", mapStringValue(metadata["region_id"]),
		"zone_id", mapStringValue(metadata["zone_id"]),
		"cloud_account_username", mapStringValue(metadata["cloud_account_username"]),
		"cloud_account_use_public", mapStringValue(metadata["cloud_account_use_public"]),
		"flavor_id", mapStringValue(metadata["flavor_id"]),
		"boot_loader_flavor_id", mapStringValue(metadata["boot_loader_flavor_id"]),
		"arch", mapStringValue(metadata["arch"]),
		"os_type_id", mapStringValue(metadata["os_type_id"]),
		"os_type", mapStringValue(metadata["os_type"]),
		"flavors", mapStringValue(metadata["flavors"]),
		"flavor_vcpus", stringifyValue(metadata["flavor_vcpus"]),
		"flavor_ram", stringifyValue(metadata["flavor_ram"]),
		"max_nic_num", stringifyValue(metadata["max_nic_num"]),
		"system_volume_type_id", mapStringValue(metadata["system_volume_type_id"]),
		"volume_type_id", mapStringValue(metadata["volume_type_id"]),
		"default_volume_type_id", mapStringValue(metadata["default_volume_type_id"]),
		"default_pool_id", mapStringValue(metadata["default_pool_id"]),
		"dest_boot_mode", mapStringValue(metadata["dest_boot_mode"]),
		"network_id", mapStringValue(metadata["network_id"]),
	)
	resp, err := s.api.Get("/api/v3/getCloudInfo", q)
	if err != nil {
		return nil, err
	}
	return responseMap(resp), nil
}

func (s Service) blockStorageCloudInfo(hostID string, metadata map[string]interface{}, fetchRes string) (map[string]interface{}, error) {
	accountID := mapStringValue(metadata["cloud_account_id"])
	cloudType := mapStringValue(metadata["cloud_type"])
	if accountID == "" || cloudType == "" {
		return nil, nil
	}
	q := queryFromPairs(
		"cloud_account_id", accountID,
		"cloud_type", cloudType,
		"storage_type", "HyperGate",
		"fetch_res", fetchRes,
		"host_id", hostID,
		"storage_id", mapStringValue(metadata["storage_id"]),
		"region_id", mapStringValue(metadata["region_id"]),
		"zone_id", mapStringValue(metadata["zone_id"]),
		"network_id", mapStringValue(metadata["network_id"]),
		"subnet_id", mapStringValue(metadata["subnet_id"]),
		"security_group_id", mapStringValue(metadata["security_group_id"]),
		"flavor_id", mapStringValue(metadata["flavor_id"]),
		"volume_type_id", mapStringValue(metadata["volume_type_id"]),
		"system_volume_type_id", mapStringValue(metadata["system_volume_type_id"]),
	)
	resp, err := s.api.Get("/api/v3/getCloudInfo", q)
	if err != nil {
		return nil, err
	}
	return responseMap(resp), nil
}

func stringifyValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case int:
		return strconv.Itoa(typed)
	case int32:
		return strconv.Itoa(int(typed))
	case int64:
		return strconv.Itoa(int(typed))
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.Itoa(int(typed))
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

func findRegionRow(data map[string]interface{}, regionID string) map[string]interface{} {
	if data == nil || regionID == "" {
		return nil
	}
	for _, row := range cloudinfo.RegionRows(data) {
		if row["region_id"] == regionID || row["id"] == regionID || row["value"] == regionID {
			return row
		}
	}
	return nil
}

func findResourceRow(data map[string]interface{}, key, id string) map[string]interface{} {
	if data == nil || id == "" {
		return nil
	}
	for _, row := range cloudinfo.ResourceRows(data, key) {
		if found := findResourceRowRecursive(row, key, id); found != nil {
			return found
		}
	}
	return nil
}

func findResourceRowRecursive(row map[string]interface{}, key, id string) map[string]interface{} {
	if row == nil {
		return nil
	}
	for _, candidate := range resourceRowCandidates(row, key) {
		if candidate == id && candidate != "" {
			return row
		}
	}
	for _, child := range listMaps(row["children"]) {
		if found := findResourceRowRecursive(child, key, id); found != nil {
			return found
		}
	}
	return nil
}

func preferredResourceID(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	for _, row := range cloudinfo.ResourceRows(data, key) {
		if preferred := preferredResourceIDRecursive(row, key); preferred != "" {
			return preferred
		}
	}
	return ""
}

func preferredResourceIDRecursive(row map[string]interface{}, key string) string {
	if row == nil {
		return ""
	}
	if preferredResourceRow(row) {
		if id := firstNonEmptyString(resourceRowCandidates(row, key)...); id != "" {
			return id
		}
	}
	for _, child := range listMaps(row["children"]) {
		if id := preferredResourceIDRecursive(child, key); id != "" {
			return id
		}
	}
	if id := firstNonEmptyString(resourceRowCandidates(row, key)...); id != "" {
		return id
	}
	return ""
}

func preferredResourceRow(row map[string]interface{}) bool {
	switch row["is_recommend"] {
	case true, 1, "1", float64(1):
		return true
	}
	switch row["default"] {
	case true, 1, "1", float64(1):
		return true
	}
	return false
}

func resourceRowCandidates(row map[string]interface{}, key string) []string {
	candidates := []string{
		mapStringValue(row["id"]),
		mapStringValue(row["value"]),
		mapStringValue(row["uuid"]),
	}
	switch key {
	case "zones":
		candidates = append(candidates, mapStringValue(row["zone_id"]))
	case "flavors":
		candidates = append(candidates, mapStringValue(row["flavor_id"]))
	case "os_types":
		candidates = append(candidates,
			mapStringValue(row["os_type_id"]),
			mapStringValue(row["name"]),
			"id-"+mapStringValue(row["id"]),
			"id-"+mapStringValue(row["name"]),
		)
	case "volume_types":
		candidates = append(candidates, mapStringValue(row["volume_type_id"]))
	case "system_volume_types":
		candidates = append(candidates, mapStringValue(row["system_volume_type_id"]), mapStringValue(row["volume_type_id"]))
	case "networks":
		candidates = append(candidates, mapStringValue(row["network_id"]))
	case "security_groups":
		candidates = append(candidates, mapStringValue(row["security_group_id"]))
	}
	return candidates
}

func findSubnetRow(data map[string]interface{}, subnetID string) map[string]interface{} {
	if data == nil || subnetID == "" {
		return nil
	}
	for _, row := range listMaps(data["subnets"]) {
		if firstNonEmptyString(mapStringValue(row["id"]), mapStringValue(row["subnet_id"])) == subnetID {
			return row
		}
	}
	if row, ok := data["subnet"].(map[string]interface{}); ok {
		if firstNonEmptyString(mapStringValue(row["id"]), mapStringValue(row["subnet_id"])) == subnetID {
			return row
		}
	}
	for _, row := range flattenSubnets(data) {
		if firstNonEmptyString(mapStringValue(row["id"]), mapStringValue(row["subnet_id"])) == subnetID {
			return row
		}
	}
	return nil
}

func flattenSubnets(data map[string]interface{}) []map[string]interface{} {
	cloudInfoMap, _ := data["cloud_info"].(map[string]interface{})
	if cloudInfoMap == nil {
		return nil
	}
	return listMaps(cloudInfoMap["subnets"])
}

func resourceDisplayName(row map[string]interface{}, fallback string) string {
	return firstNonEmptyString(
		mapStringValue(row["display_name"]),
		mapStringValue(row["name"]),
		mapStringValue(row["local_name"]),
		mapStringValue(row["region_name"]),
		fallback,
	)
}

func resourceStableName(row map[string]interface{}, fallback string) string {
	return firstNonEmptyString(
		mapStringValue(row["name"]),
		mapStringValue(row["id"]),
		mapStringValue(row["value"]),
		fallback,
	)
}

func resourceOSTypeName(row map[string]interface{}, osTypeID string) string {
	return firstNonEmptyString(
		mapStringValue(row["display_name"]),
		mapStringValue(row["local_name"]),
		defaultOSTypeName(osTypeID),
		resourceDisplayName(row, osTypeID),
	)
}

func inferOSFamily(metadata map[string]interface{}, hostData map[string]interface{}) string {
	if value := strings.ToLower(mapStringValue(metadata["os_type"])); value != "" {
		if strings.Contains(value, "windows") {
			return "windows"
		}
		if strings.Contains(value, "linux") {
			return "linux"
		}
	}
	if value := strings.ToLower(mapStringValue(metadata["os_type_id"])); value != "" {
		if strings.Contains(value, "windows") {
			return "windows"
		}
		if strings.Contains(value, "linux") {
			return "linux"
		}
	}
	if hostData == nil {
		return ""
	}
	flat := flattenMaps(hostData)
	for _, key := range []string{"os_type", "os_name", "os_caption", "os_full_name", "platform"} {
		value := strings.ToLower(mapStringValue(flat[key]))
		if strings.Contains(value, "windows") {
			return "windows"
		}
		if strings.Contains(value, "linux") {
			return "linux"
		}
	}
	return ""
}

func inferOSFamilyValue(osTypeID string) string {
	switch strings.ToLower(osTypeID) {
	case "id-linux":
		return "Linux"
	case "id-windows":
		return "Windows"
	default:
		return ""
	}
}

func inferOSFamilyTitle(osFamily string) string {
	switch strings.ToLower(osFamily) {
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return ""
	}
}

func defaultOSTypeName(osTypeID string) string {
	switch strings.ToLower(osTypeID) {
	case "id-linux":
		return "Auto Match (Linux)"
	case "id-windows":
		return "Auto Match (Windows)"
	default:
		return osTypeID
	}
}

func inferHostArch(hostData map[string]interface{}) string {
	if hostData == nil {
		return ""
	}
	flat := flattenMaps(hostData)
	return firstNonEmptyString(
		mapStringValue(flat["arch"]),
		mapStringValue(flat["architecture"]),
	)
}

func inferHostBootMode(hostData map[string]interface{}) string {
	if hostData == nil {
		return ""
	}
	flat := flattenMaps(hostData)
	return firstNonEmptyString(
		mapStringValue(flat["dest_boot_mode"]),
		mapStringValue(flat["fixed_boot_mode"]),
		mapStringValue(flat["boot_mode"]),
	)
}

func bootModeDisplayName(mode string) string {
	switch strings.ToLower(mode) {
	case "uefi":
		return "UEFI"
	case "bios":
		return "BIOS"
	default:
		return mode
	}
}
