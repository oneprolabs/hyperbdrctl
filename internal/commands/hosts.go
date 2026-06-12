package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	apphost "hyperbdr-client/internal/app/host"
	"hyperbdr-client/internal/output"
)

func runHosts(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("host", "")
	}
	service := apphost.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "list":
		fs := newFlagSet("host list")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 10, "")
		status := fs.String("status", "", "")
		bootStatus := fs.String("boot-status", "", "")
		kw := fs.String("kw", "", "")
		cloudType := fs.String("cloud-type", "", "")
		ids := fs.String("ids", "", "")
		macs := fs.String("macs", "", "")
		extra, err := parseQueryFlagSetPassthrough(fs, args[1:])
		if err != nil {
			return err
		}
		resp, err := service.List(apphost.ListSpec{
			Status:     *status,
			BootStatus: *bootStatus,
			KW:         *kw,
			CloudType:  *cloudType,
			IDs:        *ids,
			MACs:       *macs,
			Page:       *page,
			PageSize:   *pageSize,
			Query:      extra,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "hosts", hostColumns())
	case "detail":
		fs := newFlagSet("host detail")
		id := fs.String("id", "", "")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		if *id == "" {
			return missing(ctx, "error.missing_id")
		}
		resp, err := service.Detail(apphost.DetailSpec{
			ID:    *id,
			Query: q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	case "snapshots":
		fs := newFlagSet("host snapshots")
		id := fs.String("id", "", "")
		status := fs.String("status", "", "")
		syncDetail := fs.Bool("sync-detail", false, "")
		q := queryFromPairs("sheet", "snapshot")
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		if *id == "" {
			return missing(ctx, "error.missing_id")
		}
		resp, err := service.Snapshots(apphost.SnapshotsSpec{
			ID:         *id,
			Status:     *status,
			SyncDetail: *syncDetail,
			Query:      q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "snapshots", snapshotColumns())
	case "sync":
		return runHostsSync(ctx, args[1:])
	case "register":
		return runHostsRegister(ctx, args[1:])
	case "boot":
		return runHostsBoot(ctx, args[1:])
	case "cleanup-validation-host":
		return runHostsCleanupValidationHost(ctx, args[1:])
	case "deregister":
		return runHostsDeregister(ctx, args[1:])
	case "wait":
		return runHostsWait(ctx, args[1:])
	default:
		return errUnknown("host", args[0])
	}
}

func runHostsSync(ctx *context, args []string) error {
	fs := newFlagSet("host sync")
	id := fs.String("id", "", "")
	ids := fs.String("ids", "", "")
	mode := &stringFlag{}
	transferSpeed := &intFlag{}
	fs.Var(mode, "mode", "")
	fs.Var(transferSpeed, "transfer-speed", "")
	file := fs.String("file", "", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var rawBody map[string]interface{}
	if *file != "" {
		body, err := readObjectFile(*file)
		if err != nil {
			return err
		}
		rawBody = body
	}
	var modeValue *string
	if mode.set {
		modeValue = &mode.value
	}
	var transferSpeedValue *int
	if transferSpeed.set {
		transferSpeedValue = &transferSpeed.value
	}
	service := apphost.NewService(commandAPIAdapter{ctx: ctx})
	resp, err := service.Sync(apphost.SyncSpec{
		RawBody:       rawBody,
		ID:            *id,
		IDs:           *ids,
		Mode:          modeValue,
		TransferSpeed: transferSpeedValue,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}

func runHostsRegister(ctx *context, args []string) error {
	fs := newFlagSet("host register")
	vmID := fs.String("vm-id", "", "")
	vmIDs := fs.String("vm-ids", "", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	service := apphost.NewService(commandAPIAdapter{ctx: ctx})
	resp, err := service.Register(apphost.RegisterSpec{
		VMID:  *vmID,
		VMIDs: *vmIDs,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}

func runHostsBoot(ctx *context, args []string) error {
	fs := newFlagSet("host boot")
	file := fs.String("file", "", "")
	id := fs.String("id", "", "")
	known := map[string]*string{}
	for _, name := range bootBodyFields() {
		known[name] = fs.String(strings.ReplaceAll(name, "_", "-"), "", "")
	}
	unknown, err := parseBodyFlagSetPassthrough(fs, args)
	if err != nil {
		return err
	}
	var rawBody map[string]interface{}
	if *file != "" {
		body, err := readObjectFile(*file)
		if err != nil {
			return err
		}
		rawBody = body
	}
	knownValues := map[string]string{}
	for key, ptr := range known {
		if ptr != nil {
			knownValues[key] = *ptr
		}
	}
	service := apphost.NewService(commandAPIAdapter{ctx: ctx})
	resp, err := service.Boot(apphost.BootSpec{
		RawBody: rawBody,
		ID:      *id,
		Known:   knownValues,
		Unknown: unknown,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}

func runHostsCleanupValidationHost(ctx *context, args []string) error {
	fs := newFlagSet("host cleanup-validation-host")
	id := fs.String("id", "", "")
	ids := fs.String("ids", "", "")
	file := fs.String("file", "", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var rawBody map[string]interface{}
	if *file != "" {
		body, err := readObjectFile(*file)
		if err != nil {
			return err
		}
		rawBody = body
	}
	service := apphost.NewService(commandAPIAdapter{ctx: ctx})
	resp, err := service.CleanupValidationHost(apphost.CleanupValidationHostSpec{
		RawBody: rawBody,
		ID:      *id,
		IDs:     *ids,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}

func runHostsDeregister(ctx *context, args []string) error {
	fs := newFlagSet("host deregister")
	id := fs.String("id", "", "")
	ids := fs.String("ids", "", "")
	force := fs.Bool("force", false, "")
	file := fs.String("file", "", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var rawBody map[string]interface{}
	if *file != "" {
		body, err := readObjectFile(*file)
		if err != nil {
			return err
		}
		rawBody = body
	}
	service := apphost.NewService(commandAPIAdapter{ctx: ctx})
	resp, err := service.Deregister(apphost.DeregisterSpec{
		RawBody: rawBody,
		ID:      *id,
		IDs:     *ids,
		Force:   *force,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}

func runHostsWait(ctx *context, args []string) error {
	fs := newFlagSet("host wait")
	id := fs.String("id", "", "")
	ids := fs.String("ids", "", "")
	operation := fs.String("operation", "", "")
	intervalSeconds := fs.Int("interval-seconds", 60, "")
	timeoutSeconds := fs.Int("timeout-seconds", 3600, "")
	includeSteps := fs.Bool("include-steps", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	service := apphost.NewService(commandAPIAdapter{ctx: ctx})
	result, err := service.Wait(apphost.WaitSpec{
		ID:           *id,
		IDs:          *ids,
		Operation:    *operation,
		Interval:     time.Duration(*intervalSeconds) * time.Second,
		Timeout:      time.Duration(*timeoutSeconds) * time.Second,
		IncludeSteps: *includeSteps,
	})
	if err != nil {
		return err
	}
	if ctx.cfg.Output == "json" {
		if err := output.JSON(ctx.out, result.Rows); err != nil {
			return err
		}
	} else if err := output.Table(ctx.out, ctx.loc, result.Rows, waitColumns()); err != nil {
		return err
	}
	if result.Failed {
		return fmt.Errorf("one or more host wait operations failed")
	}
	return nil
}

func waitColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.operation", Field: "operation"},
		{HeaderKey: "table.result", Field: "result"},
		{HeaderKey: "table.status", Field: "status"},
		{HeaderKey: "table.display_status", Field: "display_status"},
		{HeaderKey: "table.task_id", Field: "task_id"},
		{HeaderKey: "table.elapsed_seconds", Field: "elapsed_seconds"},
		{HeaderKey: "table.error", Field: "error"},
	}
}

func bootBodyFields() []string {
	return []string{
		"migration_id",
		"snapshot_id",
		"boot_instance_purpose",
		"cloud_type",
	}
}

func readObjectFile(path string) (map[string]interface{}, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var body map[string]interface{}
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, err
	}
	return body, nil
}

func parseBodyFlagSet(fs *flag.FlagSet, args []string) (map[string]interface{}, error) {
	return parseBodyFlagSetWithPassthrough(fs, args, false)
}

func parseBodyFlagSetPassthrough(fs *flag.FlagSet, args []string) (map[string]interface{}, error) {
	return parseBodyFlagSetWithPassthrough(fs, args, true)
}

func parseBodyFlagSetWithPassthrough(fs *flag.FlagSet, args []string, allowUnknown bool) (map[string]interface{}, error) {
	known := map[string]bool{}
	fs.VisitAll(func(f *flag.Flag) {
		known[f.Name] = true
	})
	filtered := make([]string, 0, len(args))
	extra := map[string]interface{}{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") || arg == "--" {
			filtered = append(filtered, arg)
			continue
		}
		nameValue := strings.TrimPrefix(arg, "--")
		name := nameValue
		value := ""
		hasValue := false
		if idx := strings.Index(nameValue, "="); idx >= 0 {
			name = nameValue[:idx]
			value = nameValue[idx+1:]
			hasValue = true
		}
		if known[name] {
			filtered = append(filtered, arg)
			if !hasValue && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				filtered = append(filtered, args[i+1])
				i++
			}
			continue
		}
		if !allowUnknown {
			return nil, errUnknownLongFlag(name)
		}
		if !hasValue {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				value = args[i+1]
				i++
			} else {
				value = "true"
			}
		}
		extra[strings.ReplaceAll(name, "-", "_")] = apphost.InferValue(value)
	}
	if err := fs.Parse(filtered); err != nil {
		return nil, err
	}
	return extra, nil
}

type stringFlag struct {
	value string
	set   bool
}

func (f *stringFlag) Set(value string) error {
	f.value = value
	f.set = true
	return nil
}

func (f *stringFlag) String() string {
	return f.value
}

type intFlag struct {
	value int
	set   bool
}

func (f *intFlag) Set(value string) error {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	f.value = parsed
	f.set = true
	return nil
}

func (f *intFlag) String() string {
	return strconv.Itoa(f.value)
}

func hostColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.status", Field: "status"},
		{HeaderKey: "table.boot_status", Field: "boot_status"},
		{HeaderKey: "table.health_status", Field: "health_status"},
		{HeaderKey: "table.host_type", Field: "host_type"},
		{HeaderKey: "table.os_type", Field: "os_type"},
	}
}

func snapshotColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.status", Field: "status"},
		{HeaderKey: "table.created_at", Field: "created_at"},
	}
}
