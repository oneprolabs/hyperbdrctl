package commands

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	appobjectstorage "hyperbdr-client/internal/app/objectstorage"
	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/output"
)

func runObjectStorages(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("target oss", "")
	}
	service := appobjectstorage.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "list":
		fs := newFlagSet("target oss list")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 100, "")
		storageType := fs.String("type", "objectstorage", "")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.List(appobjectstorage.ListSpec{
			Page:        *page,
			PageSize:    *pageSize,
			StorageType: *storageType,
			Query:       q,
		})
		if err != nil {
			return err
		}
		return writeObjectStorageListResponse(ctx, resp)
	case "detail":
		fs := newFlagSet("target oss detail")
		id := fs.String("id", "", "")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		if *id == "" {
			return missing(ctx, "error.missing_id")
		}
		resp, err := service.Detail(appobjectstorage.DetailSpec{
			ID:    *id,
			Query: q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	case "wait":
		return runTargetOSSWait(ctx, args[1:])
	case "buckets":
		return runObjectStorageBuckets(ctx, args[1:])
	case "catalog":
		return runObjectStorageCatalog(ctx, args[1:])
	case "create":
		return runObjectStorageCreate(ctx, args[1:])
	case "delete":
		return runObjectStorageDelete(ctx, args[1:])
	default:
		return errUnknown("target oss", args[0])
	}
}

func runObjectStorageBuckets(ctx *context, args []string) error {
	fs := newFlagSet("target oss buckets")
	authURL := fs.String("auth-url", "", "")
	regionID := fs.String("region-id", "", "")
	accessKeyID := fs.String("access-key-id", "", "")
	accessKeySecret := fs.String("access-key-secret", "", "")
	protocol := fs.String("protocol", "s3", "")
	bucketLookup := fs.String("bucket-lookup", "dns", "")
	useTLS := fs.Bool("use-tls", true, "")

	if err := fs.Parse(args); err != nil {
		return err
	}
	service := appobjectstorage.NewService(commandPosterAdapter{ctx: ctx})
	resp, err := service.Buckets(appobjectstorage.BucketsSpec{
		AuthURL:         *authURL,
		RegionID:        *regionID,
		AccessKeyID:     *accessKeyID,
		AccessKeySecret: *accessKeySecret,
		Protocol:        *protocol,
		BucketLookup:    *bucketLookup,
		UseTLS:          *useTLS,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "buckets", objectStorageBucketColumns())
}

func runObjectStorageCreate(ctx *context, args []string) error {
	fs := newFlagSet("target oss create")
	displayName := fs.String("display-name", "", "")
	provider := fs.String("provider", "", "")
	authURL := fs.String("auth-url", "", "")
	regionID := fs.String("region-id", "", "")
	accessKeyID := fs.String("access-key-id", "", "")
	accessKeySecret := fs.String("access-key-secret", "", "")
	protocol := fs.String("protocol", "", "")
	bucketLookup := fs.String("bucket-lookup", "", "")
	useTLS := fs.Bool("use-tls", true, "")
	bucketMode := fs.String("bucket-mode", "existing", "")
	bucketName := fs.String("bucket-name", "", "")
	publicEndpoint := fs.String("public-endpoint", "", "")
	internalEndpoint := fs.String("internal-endpoint", "", "")
	cloudTypeSelect := fs.String("cloud-type-select", "", "")
	appID := fs.String("app-id", "", "")
	previewRequest := fs.Bool("preview-request", false, "")

	if err := fs.Parse(args); err != nil {
		return err
	}

	spec := appobjectstorage.CreateSpec{
		DisplayName:      *displayName,
		AuthURL:          *authURL,
		RegionID:         *regionID,
		AccessKeyID:      *accessKeyID,
		AccessKeySecret:  *accessKeySecret,
		Protocol:         *protocol,
		BucketLookup:     *bucketLookup,
		UseTLS:           *useTLS,
		BucketMode:       *bucketMode,
		BucketName:       *bucketName,
		PublicEndpoint:   *publicEndpoint,
		InternalEndpoint: *internalEndpoint,
		CloudTypeSelect:  *cloudTypeSelect,
		AppID:            *appID,
	}
	if err := applyObjectStorageCatalogDefaults(ctx, fs, *provider, &spec); err != nil {
		return err
	}

	service := appobjectstorage.NewService(commandPosterAdapter{ctx: ctx})
	if *previewRequest {
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

func runObjectStorageCatalog(ctx *context, args []string) error {
	fs := newFlagSet("target oss catalog")
	provider := fs.String("provider", "", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if ctx.cfg.Output == "json" {
		rawProviders, err := loadObjectStorageCatalogRaw(ctx)
		if err != nil {
			return err
		}
		if *provider == "" {
			return output.JSON(ctx.out, rawProviders)
		}
		matchedProvider, ok := findObjectStorageCatalogProviderRaw(rawProviders, *provider)
		if !ok {
			return fmt.Errorf(ctx.loc.T("error.object_storage_catalog_provider_not_found"), *provider)
		}
		return output.JSON(ctx.out, matchedProvider)
	}

	providers, err := loadObjectStorageCatalog(ctx)
	if err != nil {
		return err
	}
	if *provider == "" {
		return writeRows(ctx, objectStorageCatalogProviderRows(providers, ctx.loc.Lang()), objectStorageCatalogProviderColumns())
	}
	matchedProvider, ok := findObjectStorageCatalogProvider(providers, *provider)
	if !ok {
		return fmt.Errorf(ctx.loc.T("error.object_storage_catalog_provider_not_found"), *provider)
	}
	return writeRows(ctx, objectStorageCatalogRegionRows(*matchedProvider, ctx.loc.Lang()), objectStorageCatalogRegionColumns())
}

func applyObjectStorageCatalogDefaults(ctx *context, fs *flag.FlagSet, provider string, spec *appobjectstorage.CreateSpec) error {
	if strings.TrimSpace(provider) == "" {
		return nil
	}
	if strings.TrimSpace(spec.RegionID) == "" {
		return missing(ctx, "error.missing_region_id")
	}
	matchedProvider, matchedRegion, err := resolveObjectStorageCatalogRegion(ctx, provider, spec.RegionID)
	if err != nil {
		return err
	}
	spec.CloudType = matchedProvider.ID
	if !flagWasSet(fs, "auth-url") {
		spec.AuthURL = matchedRegion.AuthURL
	}
	if !flagWasSet(fs, "public-endpoint") {
		spec.PublicEndpoint = matchedRegion.ExternalEndpoint
	}
	if !flagWasSet(fs, "internal-endpoint") {
		spec.InternalEndpoint = matchedRegion.InternalEndpoint
	}
	if !flagWasSet(fs, "protocol") {
		spec.Protocol = matchedRegion.Protocol
	}
	if !flagWasSet(fs, "bucket-lookup") {
		spec.BucketLookup = matchedRegion.BucketLookup
	}
	if !flagWasSet(fs, "cloud-type-select") {
		spec.CloudTypeSelect = matchedProvider.ID + "," + spec.RegionID
	}
	return nil
}

func runObjectStorageDelete(ctx *context, args []string) error {
	fs := newFlagSet("target oss delete")
	id := fs.String("id", "", "")
	force := fs.Bool("force", false, "")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return missing(ctx, "error.missing_id")
	}

	service := appobjectstorage.NewService(commandPosterAdapter{ctx: ctx})
	if !*force {
		associatedResp, err := service.AssociatedResources(appobjectstorage.AssociatedResourcesSpec{ID: *id})
		if err != nil {
			return err
		}
		if msg := objectStorageDeleteForceMessage(ctx, *id, associatedResp.Data); msg != "" {
			return fmt.Errorf("%s", msg)
		}
	}

	resp, err := service.Delete(appobjectstorage.DeleteSpec{
		ID:    *id,
		Force: *force,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}

func runTargetOSS(ctx *context, args []string) error {
	return runObjectStorages(ctx, args)
}

func objectStorageDeleteForceMessage(ctx *context, storageID string, data interface{}) string {
	rows := listFromData(data, "resources")
	if len(rows) == 0 {
		return ""
	}
	hosts := associatedObjectStorageHostNames(rows)
	if len(hosts) == 0 {
		return fmt.Sprintf(ctx.loc.T("error.object_storage_delete_has_associated_resources"), storageID)
	}
	return fmt.Sprintf(ctx.loc.T("error.object_storage_delete_has_associated_hosts"), storageID, strings.Join(hosts, ", "))
}

func associatedObjectStorageHostNames(rows []map[string]interface{}) []string {
	seen := map[string]struct{}{}
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		name := firstNonEmptyString(
			mapString(row, "host_name"),
			mapString(row, "host_id"),
		)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func objectStorageColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.bucket_name", Field: "bucket_name"},
		{HeaderKey: "table.region", Field: "region"},
		{HeaderKey: "table.auth_url", Field: "auth_url"},
		{HeaderKey: "table.status", Field: "status"},
	}
}

func wizardObjectStorageColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.uuid", Field: "uuid"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.storage_type", Field: "type"},
		{HeaderKey: "table.status", Field: "status"},
		{HeaderKey: "table.region", Field: "region_name"},
	}
}

func objectStorageBucketColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.bucket_name", Field: "name"},
		{HeaderKey: "table.location", Field: "location"},
		{HeaderKey: "table.created_at", Field: "created_at"},
	}
}

func writeObjectStorageListResponse(ctx *context, resp client.APIResponse) error {
	if ctx.cfg.Output == "json" {
		return writeResponse(ctx, resp, "storages", nil)
	}
	return writeRows(ctx, normalizeObjectStorageListRows(resp.Data), objectStorageColumns())
}

func normalizeObjectStorageListRows(data interface{}) []map[string]interface{} {
	rows := listFromData(data, "storages")
	normalized := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		config := nestedMap(row, "config")
		normalized = append(normalized, map[string]interface{}{
			"id":          firstNonEmptyString(mapString(row, "id"), mapString(row, "uuid")),
			"name":        firstNonEmptyString(mapString(row, "display_name", "name"), mapString(config, "display_name", "name")),
			"bucket_name": firstNonEmptyString(mapString(row, "bucket_name"), mapString(config, "bucket_name")),
			"region": firstNonEmptyString(
				mapString(row, "region_name", "display_region_name", "region_display_name", "region_id"),
				mapString(config, "region_name", "region_id"),
			),
			"auth_url": firstNonEmptyString(
				mapString(row, "auth_url"),
				mapString(config, "auth_url", "public_endpoint", "internal_endpoint"),
			),
			"status": firstNonEmptyString(mapString(row, "display_status", "status"), mapString(config, "status")),
		})
	}
	return normalized
}
