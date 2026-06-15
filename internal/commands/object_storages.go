package commands

import (
	appobjectstorage "hyperbdr-client/internal/app/objectstorage"
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
		return writeResponse(ctx, resp, "storages", objectStorageColumns())
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
	case "create":
		return runObjectStorageCreate(ctx, args[1:])
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
	file := fs.String("file", "", "")
	displayName := fs.String("display-name", "", "")
	authURL := fs.String("auth-url", "", "")
	regionID := fs.String("region-id", "", "")
	accessKeyID := fs.String("access-key-id", "", "")
	accessKeySecret := fs.String("access-key-secret", "", "")
	protocol := fs.String("protocol", "s3", "")
	bucketLookup := fs.String("bucket-lookup", "dns", "")
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

	var rawBody interface{}
	if *file != "" {
		body, err := apiRequestBody(*file, "")
		if err != nil {
			return err
		}
		rawBody = body
	}

	service := appobjectstorage.NewService(commandPosterAdapter{ctx: ctx})
	if *previewRequest {
		prepared, err := service.PrepareCreate(appobjectstorage.CreateSpec{
			RawBody:          rawBody,
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
		})
		if err != nil {
			return err
		}
		return output.JSON(ctx.out, prepared.Body)
	}
	resp, err := service.Create(appobjectstorage.CreateSpec{
		RawBody:          rawBody,
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
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}

func runTargetOSS(ctx *context, args []string) error {
	return runObjectStorages(ctx, args)
}

func objectStorageColumns() []output.Column {
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
