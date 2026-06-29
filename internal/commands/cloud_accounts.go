package commands

import (
	"flag"
	"fmt"
	"strings"

	"hyperbdr-client/catalog"
	appcloudaccount "hyperbdr-client/internal/app/cloudaccount"
	"hyperbdr-client/internal/normalize/cloudinfo"
	"hyperbdr-client/internal/output"
	workflowcreate "hyperbdr-client/internal/workflow/cloudaccountcreate"
)

type cloudAccountCreateSpec = appcloudaccount.CreateSpec

func runCloudAccounts(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("cloud-account", "")
	}
	service := appcloudaccount.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "list":
		fs := newFlagSet("cloud-account list")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 100, "")
		storageType := fs.String("storage-type", "", "")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.List(appcloudaccount.ListSpec{
			Page:        *page,
			PageSize:    *pageSize,
			StorageType: *storageType,
			Query:       q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "cloud_accounts", cloudAccountColumns())
	case "detail":
		fs := newFlagSet("cloud-account detail")
		id := fs.String("id", "", "")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		if *id == "" {
			return missing(ctx, "error.missing_id")
		}
		resp, err := service.Detail(appcloudaccount.DetailSpec{
			ID:    *id,
			Query: q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	case "wait":
		return runTargetAccountWait(ctx, args[1:])
	case "create":
		return runCreateCloudAccountBySelection(ctx, cloudAccountCreateSelection{}, args[1:])
	case "delete":
		return runDeleteCloudAccount(ctx, args[1:])
	default:
		return errUnknown("cloud-account", args[0])
	}
}

func cloudAccountColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.display_username", Field: "display_username"},
		{HeaderKey: "table.cloud_type", Field: "cloud_type"},
		{HeaderKey: "table.storage_type", Field: "storage_type"},
		{HeaderKey: "table.display_status", Field: "display_status"},
		{HeaderKey: "table.display_task_status", Field: "display_task_status"},
		{HeaderKey: "table.created_at", Field: "created_at"},
	}
}

func regionColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.region_id", Field: "region_id"},
		{HeaderKey: "table.region_name", Field: "region_name"},
		{HeaderKey: "table.local_name", Field: "local_name"},
	}
}

type parsedCloudAccountFetchResourcesCommand struct {
	spec appcloudaccount.FetchResourcesSpec
}

func runFetchResourcesForProvider(ctx *context, commandName, provider, cloudType, storageType string, specialized bool, args []string) error {
	parsed, err := parseCloudAccountFetchResourcesArgs(commandName, cloudType, storageType, specialized, args)
	if err != nil {
		return err
	}
	service := appcloudaccount.NewService(commandPosterAdapter{ctx: ctx})
	resp, err := service.FetchResources(parsed.spec)
	if err != nil {
		return err
	}
	return writeAuthResourcesResponse(ctx, resp, provider, cloudType, parsed.spec.StorageType, parsed.spec.FetchRes, parsed.spec.FlavorVCPUs, parsed.spec.FlavorRAM)
}

func errDeprecatedFetchResourcesFlags() error {
	return fmt.Errorf("target account fetch-resources --cloud-type/--storage-type ... has been removed; use target account fetch-block-resources <provider> or target account fetch-oss-resources <provider>")
}

func imageColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.image_id", Field: "image_id"},
		{HeaderKey: "table.image_name", Field: "image_name"},
		{HeaderKey: "table.os_type", Field: "os_type"},
		{HeaderKey: "table.os_version", Field: "os_version"},
	}
}

func cloudAccountImageColumns(rows []map[string]interface{}) []output.Column {
	cols := imageColumns()
	if hasAnyNonEmptyField(rows, "boot_mode") {
		cols = append(cols, output.Column{HeaderKey: "table.boot_mode", Field: "boot_mode"})
	}
	return cols
}

func hasAnyNonEmptyField(rows []map[string]interface{}, field string) bool {
	for _, row := range rows {
		if value, ok := row[field].(string); ok && strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

type parsedCloudAccountCreateCommand struct {
	spec             cloudAccountCreateSpec
	previewRequest   bool
	cloudTypeSet     bool
	storageTypeSet   bool
	cloudAuthTypeSet bool
	remainingArgs    []string
}

type parsedCloudAccountCreateRawCommand struct {
	body           map[string]interface{}
	previewRequest bool
	remainingArgs  []string
}

func runCreateCloudAccountRaw(ctx *context, args []string) error {
	if err := rejectLegacyCloudAccountCreateInvocation(args); err != nil {
		return err
	}

	parsed, err := parseCloudAccountCreateRawArgs(args)
	if err != nil {
		return err
	}
	if len(parsed.remainingArgs) > 0 {
		return errUnknown("cloud-account create", parsed.remainingArgs[0])
	}
	return executeCreateCloudAccountRaw(ctx, parsed.body, parsed.previewRequest)
}

func parseCloudAccountCreateRawArgs(args []string) (parsedCloudAccountCreateRawCommand, error) {
	fs := newFlagSet("cloud-account create")
	file := fs.String("file", "", "")
	inlineBody := fs.String("body", "", "")
	previewRequest := fs.Bool("preview-request", false, "")

	if err := fs.Parse(args); err != nil {
		return parsedCloudAccountCreateRawCommand{}, err
	}

	body, err := requestBodyFromInput(*file, *inlineBody)
	if err != nil {
		return parsedCloudAccountCreateRawCommand{}, err
	}
	if body == nil {
		return parsedCloudAccountCreateRawCommand{}, fmt.Errorf("file or body is required")
	}

	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		return parsedCloudAccountCreateRawCommand{}, fmt.Errorf("cloud-account create body must be a JSON object")
	}

	return parsedCloudAccountCreateRawCommand{
		body:           bodyMap,
		previewRequest: *previewRequest,
		remainingArgs:  fs.Args(),
	}, nil
}

func parseCloudAccountCreateArgs(commandName string, args []string) (parsedCloudAccountCreateCommand, error) {
	fs := newFlagSet(commandName)
	cloudType := fs.String("cloud-type", "", "")
	cloudAuthType := fs.String("cloud-auth-type", "", "")
	accessKeyID := fs.String("access-key-id", "", "")
	accessKeySecret := fs.String("access-key-secret", "", "")
	regionID := fs.String("region-id", "", "")
	regionName := fs.String("region-name", "", "")
	accountName := fs.String("account-name", "", "")
	authRegionID := fs.String("auth-region-id", "", "")
	authURL := fs.String("auth-url", "", "")
	username := fs.String("username", "", "")
	password := fs.String("password", "", "")
	userDomainID := fs.String("user-domain-id", "", "")
	projectDomainID := fs.String("project-domain-id", "", "")
	projectID := fs.String("project-id", "", "")
	projectName := fs.String("project-name", "", "")
	storageType := fs.String("storage-type", "", "")
	useInternalIP := fs.String("use-internal-ip", "0", "")
	bootLoaderImageID := fs.String("boot-loader-image-id", "", "")
	bootLoaderImageName := fs.String("boot-loader-image-name", "", "")
	bootLoaderFlavorID := fs.String("boot-loader-flavor-id", "", "")
	linuxBootImageID := fs.String("linux-boot-image-id", "", "")
	windowsBootImageID := fs.String("windows-boot-image-id", "", "")
	linuxUEFIBootImageID := fs.String("linux-uefi-boot-image-id", "", "")
	windowsUEFIBootImageID := fs.String("windows-uefi-boot-image-id", "", "")
	customName := fs.String("custom-name", "", "")
	diskBusTypeID := fs.String("disk-bus-type-id", "", "")
	diskBusTypeName := fs.String("disk-bus-type-name", "", "")
	sshPort := fs.String("ssh-port", "", "")
	sshPass := fs.String("ssh-pass", "", "")
	linuxHDUsername := fs.String("linux-hd-username", "", "")
	linuxHDPassword := fs.String("linux-hd-password", "", "")
	linuxHDPort := fs.String("linux-hd-port", "", "")
	autoUploadImages := fs.Int("auto-upload-images", 0, "")
	uploadUEFIImage := fs.Int("upload-uefi-image", 0, "")
	onlyVerify := fs.Bool("only-verify", false, "")
	previewRequest := fs.Bool("preview-request", false, "")

	if err := fs.Parse(args); err != nil {
		return parsedCloudAccountCreateCommand{}, err
	}

	spec := cloudAccountCreateSpec{
		CloudType:              *cloudType,
		CloudAuthType:          *cloudAuthType,
		StorageType:            *storageType,
		AccessKeyID:            *accessKeyID,
		AccessKeySecret:        *accessKeySecret,
		RegionID:               *regionID,
		RegionName:             *regionName,
		AccountName:            *accountName,
		AuthRegionID:           *authRegionID,
		AuthURL:                *authURL,
		CloudAccountUsername:   *username,
		CloudAccountPassword:   *password,
		UserDomainID:           *userDomainID,
		ProjectDomainID:        *projectDomainID,
		ProjectID:              *projectID,
		ProjectName:            *projectName,
		UseInternalIP:          *useInternalIP,
		BootLoaderImageID:      *bootLoaderImageID,
		BootLoaderImageName:    *bootLoaderImageName,
		BootLoaderFlavorID:     *bootLoaderFlavorID,
		LinuxBootImageID:       *linuxBootImageID,
		WindowsBootImageID:     *windowsBootImageID,
		LinuxUEFIBootImageID:   *linuxUEFIBootImageID,
		WindowsUEFIBootImageID: *windowsUEFIBootImageID,
		CustomName:             *customName,
		DiskBusTypeID:          *diskBusTypeID,
		DiskBusTypeName:        *diskBusTypeName,
		SSHPort:                *sshPort,
		SSHPass:                *sshPass,
		LinuxHDUsername:        *linuxHDUsername,
		LinuxHDPassword:        *linuxHDPassword,
		LinuxHDPort:            *linuxHDPort,
	}
	if flagWasSet(fs, "auto-upload-images") {
		value := *autoUploadImages
		spec.AutoUploadImages = &value
	}
	if flagWasSet(fs, "upload-uefi-image") {
		value := *uploadUEFIImage
		spec.UploadUEFIImage = &value
	}
	if flagWasSet(fs, "only-verify") {
		value := *onlyVerify
		spec.OnlyVerify = &value
	}

	return parsedCloudAccountCreateCommand{
		spec:             spec,
		previewRequest:   *previewRequest,
		cloudTypeSet:     flagWasSet(fs, "cloud-type"),
		storageTypeSet:   flagWasSet(fs, "storage-type"),
		cloudAuthTypeSet: flagWasSet(fs, "cloud-auth-type"),
		remainingArgs:    fs.Args(),
	}, nil
}

func runCreateCloudAccountForProvider(ctx *context, commandName, cloudType, storageType string, specialized bool, args []string) error {
	if storageType == "objectstorage" {
		parsed, err := parseCloudAccountCreateOSSArgs(commandName, cloudType, specialized, args)
		if err != nil {
			return err
		}
		return executeCreateCloudAccountSpec(ctx, parsed.spec, parsed.previewRequest)
	}

	parsed, err := parseCloudAccountCreateBlockArgs(commandName, cloudType, specialized, args)
	if err != nil {
		return err
	}
	return executeCreateCloudAccountSpec(ctx, parsed.spec, parsed.previewRequest)
}

func rejectLegacyCloudAccountCreateInvocation(args []string) error {
	if len(args) == 0 {
		return nil
	}

	for _, arg := range args {
		if strings.HasPrefix(arg, "--cloud-type") {
			return errDeprecatedCloudAccountCreateFlags()
		}
	}

	if strings.HasPrefix(args[0], "-") {
		return nil
	}

	switch args[0] {
	case "block":
		if len(args) > 1 {
			return errLegacyCloudAccountCreateGroup("block", args[1])
		}
		return fmt.Errorf("target account create block has been removed; use target account create-block <provider>")
	case "oss", "object":
		if len(args) > 1 {
			return errLegacyCloudAccountCreateGroup("oss", args[1])
		}
		return fmt.Errorf("target account create %s has been removed; use target account create-oss <provider>", args[0])
	default:
		if commandName, ok := legacyCloudAccountCreateProviderCommand(args[0]); ok {
			return fmt.Errorf("target account create %s has been removed; use %s", args[0], commandName)
		}
	}
	return nil
}

func errDeprecatedCloudAccountCreateFlags() error {
	return fmt.Errorf("target account create --cloud-type ... has been removed; use target account create --file/--body, target account create-block <provider>, or target account create-oss <provider>")
}

func errLegacyCloudAccountCreateGroup(kind, provider string) error {
	switch normalizeLegacyProvider(provider) {
	case "aliyun":
		if kind == "block" {
			return fmt.Errorf("target account create block %s has been removed; use target account create-block aliyun", provider)
		}
		return fmt.Errorf("target account create %s %s has been removed; use target account create-oss aliyun", kind, provider)
	case "openstack":
		if kind == "block" {
			return fmt.Errorf("target account create block %s has been removed; use target account create-block openstack", provider)
		}
		return fmt.Errorf("target account create %s %s has been removed; use target account create-oss openstack", kind, provider)
	default:
		if kind == "block" {
			return fmt.Errorf("target account create block has been removed; use target account create-block <provider>")
		}
		return fmt.Errorf("target account create %s has been removed; use target account create-oss <provider>", kind)
	}
}

func legacyCloudAccountCreateProviderCommand(key string) (string, bool) {
	normalized := normalizeLegacyProvider(key)
	for _, entry := range catalog.EnabledBlockClouds() {
		if normalizeLegacyProvider(entry.Key) == normalized {
			return "target account create-block " + entry.Provider, true
		}
	}
	for _, entry := range catalog.EnabledObjectClouds() {
		if normalizeLegacyProvider(entry.Key) == normalized {
			return "target account create-oss " + entry.Provider, true
		}
	}
	return "", false
}

func normalizeLegacyProvider(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func executeCreateCloudAccountRaw(ctx *context, body map[string]interface{}, previewRequest bool) error {
	service := appcloudaccount.NewService(commandPosterAdapter{ctx: ctx})
	spec := appcloudaccount.CreateRawSpec{Body: body}
	if previewRequest {
		prepared, err := service.PrepareCreateRaw(spec)
		if err != nil {
			return err
		}
		return output.JSON(ctx.out, prepared.Body)
	}
	resp, err := service.CreateRaw(spec)
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}

func executeCreateCloudAccountSpec(ctx *context, spec cloudAccountCreateSpec, previewRequest bool) error {
	var err error
	spec = workflowcreate.NormalizeSpec(spec)
	spec, err = enrichCreateCloudAccountSpec(ctx, spec)
	if err != nil {
		return err
	}

	service := appcloudaccount.NewService(commandPosterAdapter{ctx: ctx})
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

func enrichCreateCloudAccountSpec(ctx *context, spec cloudAccountCreateSpec) (cloudAccountCreateSpec, error) {
	switch {
	case spec.CloudType == "aliyun_obs" && spec.StorageType == "objectstorage":
		return enrichAliyunObjectCloudAccountSpec(ctx, spec)
	case spec.CloudType == "openstack" && spec.StorageType == "objectstorage":
		return enrichOpenStackObjectCloudAccountSpec(ctx, spec)
	case spec.StorageType == "objectstorage":
		return enrichGenericObjectCloudAccountSpec(ctx, spec), nil
	default:
		return spec, nil
	}
}

func enrichGenericObjectCloudAccountSpec(ctx *context, spec cloudAccountCreateSpec) cloudAccountCreateSpec {
	if spec.CustomName == "" {
		spec.CustomName = defaultGenericObjectCloudAccountName(ctx, spec)
	}
	return spec
}

func enrichAliyunObjectCloudAccountSpec(ctx *context, spec cloudAccountCreateSpec) (cloudAccountCreateSpec, error) {
	var err error

	if spec.RegionName == "" {
		spec.RegionName = resolveCloudAccountRegionName(ctx, spec)
		if spec.RegionName == "" {
			spec.RegionName = spec.RegionID
		}
	}
	if spec.CustomName == "" {
		spec.CustomName = defaultAliyunObjectCloudAccountName(ctx.loc.Lang(), spec.RegionName)
	}
	if spec.BootLoaderImageID == "" {
		spec.BootLoaderImageID, spec.BootLoaderImageName, err = resolveCloudAccountBootLoaderImage(ctx, spec)
		if err != nil {
			return spec, err
		}
	}

	return spec, nil
}

func enrichOpenStackObjectCloudAccountSpec(ctx *context, spec cloudAccountCreateSpec) (cloudAccountCreateSpec, error) {
	service := appcloudaccount.NewService(commandPosterAdapter{ctx: ctx})

	regionResp, err := service.FetchOpenStackObjectResources(appcloudaccount.FetchOpenStackObjectResourcesSpec{
		AuthURL:         spec.AuthURL,
		Username:        spec.CloudAccountUsername,
		Password:        spec.CloudAccountPassword,
		UserDomainID:    spec.UserDomainID,
		FetchRes:        "region",
		RegionID:        spec.RegionID,
		ProjectID:       spec.ProjectID,
		ProjectDomainID: spec.ProjectDomainID,
		ProjectName:     spec.ProjectName,
	})
	if err != nil {
		return spec, fmt.Errorf("auto-resolve OpenStack object defaults: %w", err)
	}

	authInfo := nestedMap(regionResp.Data, "cloud_info", "auth_info")
	spec.ProjectDomainID = firstNonEmptyString(spec.ProjectDomainID, mapString(authInfo, "project_domain_id"))
	spec.ProjectID = firstNonEmptyString(spec.ProjectID, mapString(authInfo, "project_id"))
	spec.ProjectName = firstNonEmptyString(spec.ProjectName, mapString(authInfo, "project_name", "tenant_name"))
	spec.RegionID = firstNonEmptyString(spec.RegionID, mapString(authInfo, "region_id"))
	if spec.RegionName == "" {
		spec.RegionName = resolveOpenStackObjectRegionName(ctx, regionResp.Data, spec.RegionID, authInfo)
	}
	if spec.RegionName == "" {
		spec.RegionName = spec.RegionID
	}
	if spec.CustomName == "" {
		spec.CustomName = defaultOpenStackObjectCloudAccountName(ctx.loc.Lang(), spec.RegionName)
	}

	flavorResp, err := service.FetchOpenStackObjectResources(appcloudaccount.FetchOpenStackObjectResourcesSpec{
		AuthURL:         spec.AuthURL,
		Username:        spec.CloudAccountUsername,
		Password:        spec.CloudAccountPassword,
		UserDomainID:    spec.UserDomainID,
		FetchRes:        "flavor",
		RegionID:        spec.RegionID,
		ProjectID:       spec.ProjectID,
		ProjectDomainID: spec.ProjectDomainID,
		ProjectName:     spec.ProjectName,
	})
	if err != nil {
		return spec, fmt.Errorf("auto-resolve OpenStack object defaults: %w", err)
	}

	spec.BootLoaderImageID, spec.BootLoaderImageName, err = resolveOpenStackObjectBootLoaderImage(flavorResp.Data, spec.BootLoaderImageID, spec.BootLoaderImageName)
	if err != nil {
		return spec, err
	}
	spec.BootLoaderFlavorID, err = resolveOpenStackObjectBootLoaderFlavor(flavorResp.Data, spec.BootLoaderFlavorID)
	if err != nil {
		return spec, err
	}
	spec.DiskBusTypeID, spec.DiskBusTypeName, err = resolveOpenStackObjectDiskBus(flavorResp.Data, spec.DiskBusTypeID, spec.DiskBusTypeName)
	if err != nil {
		return spec, err
	}

	return spec, nil
}

func resolveCloudAccountRegionName(ctx *context, spec cloudAccountCreateSpec) string {
	if spec.CloudType == "" || spec.StorageType == "" || spec.AccessKeyID == "" || spec.AccessKeySecret == "" || spec.RegionID == "" {
		return ""
	}

	service := appcloudaccount.NewService(commandPosterAdapter{ctx: ctx})
	resp, err := service.FetchResources(appcloudaccount.FetchResourcesSpec{
		Spec: workflowcreate.Spec{
			CloudType:       spec.CloudType,
			AccessKeyID:     spec.AccessKeyID,
			AccessKeySecret: spec.AccessKeySecret,
			StorageType:     spec.StorageType,
			RegionID:        spec.RegionID,
		},
		FetchRes: "regions",
	})
	if err != nil {
		return ""
	}

	rows := cloudinfo.RegionRows(resp.Data)
	for _, row := range rows {
		if mapString(row, "region_id", "id", "value") == spec.RegionID {
			return preferredRegionLabel(ctx.loc.Lang(), row)
		}
	}
	if len(rows) == 1 {
		return preferredRegionLabel(ctx.loc.Lang(), rows[0])
	}
	return ""
}

func resolveCloudAccountBootLoaderImage(ctx *context, spec cloudAccountCreateSpec) (string, string, error) {
	if spec.CloudType == "" || spec.StorageType == "" || spec.AccessKeyID == "" || spec.AccessKeySecret == "" || spec.RegionID == "" {
		return "", "", nil
	}

	service := appcloudaccount.NewService(commandPosterAdapter{ctx: ctx})
	resp, err := service.FetchResources(appcloudaccount.FetchResourcesSpec{
		Spec: workflowcreate.Spec{
			CloudType:       spec.CloudType,
			AccessKeyID:     spec.AccessKeyID,
			AccessKeySecret: spec.AccessKeySecret,
			StorageType:     spec.StorageType,
			RegionID:        spec.RegionID,
		},
		BootMode: "bios",
		FetchRes: "boot_loader_images",
	})
	if err != nil {
		return "", "", fmt.Errorf("auto-resolve boot-loader-image-id: %w", err)
	}

	rows := cloudinfo.NormalizeImageRows(cloudinfo.ResourceRows(resp.Data, "boot_loader_images"))
	if len(rows) == 0 {
		return "", "", fmt.Errorf("auto-resolve boot-loader-image-id: no boot_loader_images candidates returned")
	}

	if spec.BootLoaderImageName != "" {
		for _, row := range rows {
			name := mapString(row, "image_name", "name", "display_name", "id", "uuid")
			if name == spec.BootLoaderImageName {
				return mapString(row, "image_id", "id", "uuid"), name, nil
			}
		}
		return "", "", fmt.Errorf("auto-resolve boot-loader-image-id: boot-loader-image-name %q not found in boot_loader_images", spec.BootLoaderImageName)
	}

	first := rows[0]
	id := mapString(first, "image_id", "id", "uuid")
	name := mapString(first, "image_name", "name", "display_name", "id", "uuid")
	if id == "" {
		return "", "", fmt.Errorf("auto-resolve boot-loader-image-id: first boot_loader_images candidate is missing an id")
	}
	return id, name, nil
}

func defaultAliyunObjectCloudAccountName(lang, regionLabel string) string {
	label := firstNonEmptyString(regionLabel, "unknown-region")
	if lang == "zh_cn" {
		return "阿里云(推荐使用，SDK v2.0)-" + label
	}
	return "Alibaba Cloud (SDK v2.0)-" + label
}

func defaultOpenStackObjectCloudAccountName(lang, regionLabel string) string {
	label := firstNonEmptyString(regionLabel, "unknown-region")
	if lang == "zh_cn" {
		return "OpenStack社区版本(Juno+)-" + label
	}
	return "OpenStackCommunity(Juno+)-" + label
}

func defaultGenericObjectCloudAccountName(ctx *context, spec cloudAccountCreateSpec) string {
	regionLabel := firstNonEmptyString(spec.RegionName, spec.RegionID, "unknown-region")
	entry, ok := catalog.FindObjectCloud(spec.CloudType)
	if !ok {
		return spec.CloudType + "-" + regionLabel
	}

	name := entry.NameEn
	if ctx.loc.Lang() == "zh_cn" {
		name = entry.NameZhCN
	}
	return name + "-" + regionLabel
}

func preferredRegionLabel(lang string, row map[string]interface{}) string {
	if lang == "zh_cn" {
		return firstNonEmptyString(
			mapString(row, "local_name"),
			mapString(row, "region_name"),
			mapString(row, "display_name"),
			mapString(row, "name"),
			mapString(row, "region_id"),
		)
	}
	return firstNonEmptyString(
		mapString(row, "region_name"),
		mapString(row, "display_name"),
		mapString(row, "name"),
		mapString(row, "local_name"),
		mapString(row, "region_id"),
	)
}

func resolveOpenStackObjectRegionName(ctx *context, data interface{}, regionID string, authInfo map[string]interface{}) string {
	if regionName := mapString(authInfo, "region_name"); regionName != "" {
		return regionName
	}

	rows := cloudinfo.RegionRows(data)
	for _, row := range rows {
		if mapString(row, "region_id", "id", "value") == regionID {
			return preferredRegionLabel(ctx.loc.Lang(), row)
		}
	}
	if len(rows) == 1 {
		return preferredRegionLabel(ctx.loc.Lang(), rows[0])
	}
	return ""
}

func resolveOpenStackObjectBootLoaderImage(data interface{}, currentID, currentName string) (string, string, error) {
	rows := cloudinfo.NormalizeImageRows(cloudinfo.ResourceRows(data, "boot_loader_images"))
	if len(rows) == 0 {
		if currentID != "" {
			return currentID, firstNonEmptyString(currentName, currentID), nil
		}
		return "", "", fmt.Errorf("auto-resolve boot-loader-image-id: no boot_loader_images candidates returned")
	}

	if currentID != "" {
		for _, row := range rows {
			if mapString(row, "image_id", "id", "uuid") == currentID {
				return currentID, firstNonEmptyString(currentName, mapString(row, "image_name", "name", "display_name"), currentID), nil
			}
		}
		return currentID, firstNonEmptyString(currentName, currentID), nil
	}

	if currentName != "" {
		for _, row := range rows {
			if mapString(row, "image_name", "name", "display_name", "id", "uuid") == currentName {
				return mapString(row, "image_id", "id", "uuid"), currentName, nil
			}
		}
		return "", "", fmt.Errorf("auto-resolve boot-loader-image-id: boot-loader-image-name %q not found in boot_loader_images", currentName)
	}

	first := rows[0]
	id := mapString(first, "image_id", "id", "uuid")
	if id == "" {
		return "", "", fmt.Errorf("auto-resolve boot-loader-image-id: first boot_loader_images candidate is missing an id")
	}
	return id, mapString(first, "image_name", "name", "display_name", "id", "uuid"), nil
}

func resolveOpenStackObjectBootLoaderFlavor(data interface{}, currentID string) (string, error) {
	rows := cloudinfo.ResourceRows(data, "boot_loader_flavors")
	if len(rows) == 0 {
		if currentID != "" {
			return currentID, nil
		}
		return "", fmt.Errorf("auto-resolve boot-loader-flavor-id: no boot_loader_flavors candidates returned")
	}

	if currentID != "" {
		return currentID, nil
	}

	for _, row := range rows {
		if mapInt(row, "is_recommend") == 1 {
			if id := mapString(row, "id", "value"); id != "" {
				return id, nil
			}
		}
	}

	id := mapString(rows[0], "id", "value")
	if id == "" {
		return "", fmt.Errorf("auto-resolve boot-loader-flavor-id: first boot_loader_flavors candidate is missing an id")
	}
	return id, nil
}

func resolveOpenStackObjectDiskBus(data interface{}, currentID, currentName string) (string, string, error) {
	rows := cloudinfo.ResourceRows(data, "disk_bus_types")
	if len(rows) == 0 {
		if currentID != "" && currentName != "" {
			return currentID, currentName, nil
		}
		return "", "", fmt.Errorf("auto-resolve disk-bus-type: no disk_bus_types candidates returned")
	}

	match := func(row map[string]interface{}) (string, string) {
		return mapString(row, "id", "value"), mapString(row, "name", "display_name", "id", "value")
	}

	if currentID != "" || currentName != "" {
		for _, row := range rows {
			id, name := match(row)
			if (currentID != "" && id == currentID) || (currentName != "" && name == currentName) {
				return firstNonEmptyString(currentID, id), firstNonEmptyString(currentName, name, id), nil
			}
		}
		if currentID != "" && currentName != "" {
			return currentID, currentName, nil
		}
		return "", "", fmt.Errorf("auto-resolve disk-bus-type: requested candidate not found")
	}

	for _, row := range rows {
		id, name := match(row)
		if id == "virtio" || name == "virtio" {
			return firstNonEmptyString(id, name), firstNonEmptyString(name, id), nil
		}
	}

	id, name := match(rows[0])
	if id == "" && name == "" {
		return "", "", fmt.Errorf("auto-resolve disk-bus-type: first disk_bus_types candidate is missing id and name")
	}
	return firstNonEmptyString(id, name), firstNonEmptyString(name, id), nil
}

func mapString(row map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := row[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func mapInt(row map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		switch value := row[key].(type) {
		case int:
			return value
		case int32:
			return int(value)
		case int64:
			return int(value)
		case float64:
			return int(value)
		}
	}
	return 0
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func nestedMap(data interface{}, path ...string) map[string]interface{} {
	value := data
	for _, key := range path {
		m, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}
		value = m[key]
	}
	if result, ok := value.(map[string]interface{}); ok {
		return result
	}
	return nil
}

func flagWasSet(fs *flag.FlagSet, name string) bool {
	wasSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			wasSet = true
		}
	})
	return wasSet
}

func runDeleteCloudAccount(ctx *context, args []string) error {
	fs := newFlagSet("cloud-account delete")
	id := fs.String("id", "", "")
	force := fs.Bool("force", false, "")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" {
		return missing(ctx, "error.missing_id")
	}

	service := appcloudaccount.NewService(commandAPIAdapter{ctx: ctx})
	resp, err := service.Delete(appcloudaccount.DeleteSpec{
		ID:    *id,
		Force: *force,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}
