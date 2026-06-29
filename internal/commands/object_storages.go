package commands

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	appobjectstorage "hyperbdr-client/internal/app/objectstorage"
	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/output"

	"github.com/spf13/cobra"
)

const objectStorageHelpProfileAnnotation = "object-storage-help-profile"

type objectStorageHelpSelection struct {
	Provider string
}

func newObjectStorageBucketsCommand(ctx *context) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "buckets",
		Short:              ctx.loc.T("cmd.oss.buckets.short"),
		Long:               ctx.loc.T("cmd.oss.buckets.long"),
		Example:            strings.TrimSpace(ctx.loc.T("cmd.oss.buckets.examples")),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			selection, err := parseObjectStorageHelpSelection(args)
			if err != nil {
				return err
			}
			if rawArgsHelp(cmd, args) {
				return renderObjectStorageBucketsHelp(ctx, cmd, selection)
			}
			return runObjectStorageBuckets(ctx, args)
		},
	}
	addObjectStorageBucketsFlags(cmd, ctx)
	return cmd
}

func newObjectStorageCreateCommand(ctx *context) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "create",
		Short:              ctx.loc.T("cmd.oss.create.short"),
		Long:               ctx.loc.T("cmd.oss.create.long"),
		Example:            strings.TrimSpace(ctx.loc.T("cmd.oss.create.examples")),
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			selection, err := parseObjectStorageHelpSelection(args)
			if err != nil {
				return err
			}
			if rawArgsHelp(cmd, args) {
				return renderObjectStorageCreateHelp(ctx, cmd, selection)
			}
			return runObjectStorageCreate(ctx, args)
		},
	}
	addObjectStorageCreateFlags(cmd, ctx)
	return cmd
}

func addObjectStorageBucketsFlags(cmd *cobra.Command, ctx *context) {
	for _, name := range []string{"provider", "auth-url", "region-id", "access-key-id", "access-key-secret", "protocol", "bucket-lookup"} {
		addFlagString(cmd, ctx, name)
	}
	addFlagBool(cmd, ctx, "use-tls")
}

func addObjectStorageCreateFlags(cmd *cobra.Command, ctx *context) {
	for _, name := range []string{"display-name", "provider", "auth-url", "region-id", "access-key-id", "access-key-secret", "protocol", "bucket-lookup", "bucket-mode", "bucket-name", "public-endpoint", "internal-endpoint", "app-id"} {
		addFlagString(cmd, ctx, name)
	}
	addFlagBool(cmd, ctx, "use-tls")
	addFlagBool(cmd, ctx, "preview-request")
}

func parseObjectStorageHelpSelection(args []string) (objectStorageHelpSelection, error) {
	selection := objectStorageHelpSelection{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--help" || arg == "-h" || !strings.HasPrefix(arg, "--") {
			continue
		}
		name, value, hasInline := splitFlag(arg)
		if strings.TrimPrefix(name, "--") != "provider" {
			continue
		}
		v, next, err := strictFlagValue(args, i, value, hasInline)
		if err != nil {
			return selection, err
		}
		selection.Provider = strings.TrimSpace(v)
		i = next
	}
	return selection, nil
}

func renderObjectStorageBucketsHelp(ctx *context, cmd *cobra.Command, selection objectStorageHelpSelection) error {
	profile := objectStorageHelpProfile(selection.Provider)
	addAnnotationValue(cmd, objectStorageHelpProfileAnnotation, profile)
	addHelpDescription(cmd, ctx, objectStorageBucketsHelpTitleKey(profile))
	addAnnotationValue(cmd, usageNotesAnnotation, objectStorageBucketsUsageNotes(ctx, profile))
	return renderHelp(cmd, ctx)
}

func renderObjectStorageCreateHelp(ctx *context, cmd *cobra.Command, selection objectStorageHelpSelection) error {
	profile := objectStorageHelpProfile(selection.Provider)
	addAnnotationValue(cmd, objectStorageHelpProfileAnnotation, profile)
	addHelpDescription(cmd, ctx, objectStorageCreateHelpTitleKey(profile))
	addAnnotationValue(cmd, usageNotesAnnotation, objectStorageCreateUsageNotes(ctx, profile))
	return renderHelp(cmd, ctx)
}

func objectStorageHelpProfile(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "custom":
		return "custom"
	default:
		return "provider"
	}
}

func objectStorageBucketsHelpTitleKey(profile string) string {
	if profile == "provider" {
		return "cmd.oss.buckets.provider.short"
	}
	return "cmd.oss.buckets.custom.short"
}

func objectStorageCreateHelpTitleKey(profile string) string {
	if profile == "provider" {
		return "cmd.oss.create.provider.short"
	}
	return "cmd.oss.create.custom.short"
}

func objectStorageBucketsUsageNotes(ctx *context, profile string) string {
	if profile == "provider" {
		return ctx.loc.T("cmd.oss.buckets.provider.usage_notes")
	}
	return ctx.loc.T("cmd.oss.buckets.custom.usage_notes")
}

func objectStorageCreateUsageNotes(ctx *context, profile string) string {
	if profile == "provider" {
		return ctx.loc.T("cmd.oss.create.provider.usage_notes")
	}
	return ctx.loc.T("cmd.oss.create.custom.usage_notes")
}

func rejectObjectStorageListTypeFlag(args []string) error {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--") {
			continue
		}
		name, _, _ := splitFlag(arg)
		if strings.TrimPrefix(name, "--") == "type" {
			return fmt.Errorf("flag provided but not defined: -type")
		}
	}
	return nil
}

func runObjectStorages(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("oss", "")
	}
	service := appobjectstorage.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "list":
		fs := newFlagSet("oss list")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 100, "")
		q := queryFromPairs()
		if err := rejectObjectStorageListTypeFlag(args[1:]); err != nil {
			return err
		}
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.List(appobjectstorage.ListSpec{
			Page:        *page,
			PageSize:    *pageSize,
			StorageType: "objectstorage",
			Query:       q,
		})
		if err != nil {
			return err
		}
		return writeObjectStorageListResponse(ctx, resp)
	case "detail":
		fs := newFlagSet("oss detail")
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
		return runObjectStorageWait(ctx, args[1:])
	case "buckets":
		return runObjectStorageBuckets(ctx, args[1:])
	case "catalog":
		return runObjectStorageCatalog(ctx, args[1:])
	case "create":
		return runObjectStorageCreate(ctx, args[1:])
	case "delete":
		return runObjectStorageDelete(ctx, args[1:])
	default:
		return errUnknown("oss", args[0])
	}
}

func runObjectStorageBuckets(ctx *context, args []string) error {
	fs := newFlagSet("oss buckets")
	provider := fs.String("provider", "", "")
	authURL := fs.String("auth-url", "", "")
	regionID := fs.String("region-id", "", "")
	accessKeyID := fs.String("access-key-id", "", "")
	accessKeySecret := fs.String("access-key-secret", "", "")
	protocol := fs.String("protocol", "", "")
	bucketLookup := fs.String("bucket-lookup", "", "")
	useTLS := fs.Bool("use-tls", true, "")

	if err := fs.Parse(args); err != nil {
		return err
	}
	spec := appobjectstorage.BucketsSpec{
		AuthURL:         *authURL,
		RegionID:        *regionID,
		AccessKeyID:     *accessKeyID,
		AccessKeySecret: *accessKeySecret,
		Protocol:        *protocol,
		BucketLookup:    *bucketLookup,
		UseTLS:          *useTLS,
	}
	if err := applyObjectStorageBucketsDefaults(ctx, fs, *provider, &spec); err != nil {
		return err
	}
	service := appobjectstorage.NewService(commandPosterAdapter{ctx: ctx})
	resp, err := service.Buckets(spec)
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "buckets", objectStorageBucketColumns())
}

func runObjectStorageCreate(ctx *context, args []string) error {
	fs := newFlagSet("oss create")
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
		AppID:            *appID,
	}
	if err := applyObjectStorageCreateDefaults(ctx, fs, *provider, &spec); err != nil {
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
	fs := newFlagSet("oss catalog")
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

func applyObjectStorageCreateDefaults(ctx *context, fs *flag.FlagSet, provider string, spec *appobjectstorage.CreateSpec) error {
	normalizedProvider := normalizeObjectStorageCreateProvider(provider)
	if normalizedProvider == "" || normalizedProvider == "custom" {
		spec.CloudType = "custom"
		if !flagWasSet(fs, "display-name") && strings.TrimSpace(spec.DisplayName) == "" {
			spec.DisplayName = ctx.loc.T("value.object_storage.display_name.custom")
		}
		return nil
	}
	if strings.TrimSpace(spec.RegionID) == "" {
		return missing(ctx, "error.missing_region_id")
	}
	matchedProvider, matchedRegion, err := resolveObjectStorageCatalogRegion(ctx, normalizedProvider, spec.RegionID)
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
	if !flagWasSet(fs, "display-name") && strings.TrimSpace(spec.DisplayName) == "" {
		spec.DisplayName = defaultObjectStorageCatalogDisplayName(matchedProvider, matchedRegion, ctx.loc.Lang())
	}
	return nil
}

func applyObjectStorageBucketsDefaults(ctx *context, fs *flag.FlagSet, provider string, spec *appobjectstorage.BucketsSpec) error {
	normalizedProvider := normalizeObjectStorageCreateProvider(provider)
	if normalizedProvider == "" || normalizedProvider == "custom" {
		return nil
	}
	if strings.TrimSpace(spec.RegionID) == "" {
		return missing(ctx, "error.missing_region_id")
	}
	_, matchedRegion, err := resolveObjectStorageCatalogRegion(ctx, normalizedProvider, spec.RegionID)
	if err != nil {
		return err
	}
	if !flagWasSet(fs, "auth-url") {
		spec.AuthURL = matchedRegion.AuthURL
	}
	if !flagWasSet(fs, "protocol") {
		spec.Protocol = matchedRegion.Protocol
	}
	if !flagWasSet(fs, "bucket-lookup") {
		spec.BucketLookup = matchedRegion.BucketLookup
	}
	return nil
}

func normalizeObjectStorageCreateProvider(provider string) string {
	return strings.TrimSpace(strings.ToLower(provider))
}

func runObjectStorageDelete(ctx *context, args []string) error {
	fs := newFlagSet("oss delete")
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
