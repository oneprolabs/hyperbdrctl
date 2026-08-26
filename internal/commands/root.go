package commands

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"

	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/config"
	"hyperbdr-client/internal/i18n"
	"hyperbdr-client/internal/output"
	"hyperbdr-client/internal/version"
)

type context struct {
	out    io.Writer
	errOut io.Writer
	flags  config.Flags
	cfg    config.Resolved
	loc    i18n.Localizer
	client *client.Client
}

func Execute(args []string, out, errOut io.Writer) error {
	args, flags, err := extractGlobalFlags(args)
	if err != nil {
		return err
	}
	if wantsVersion(args) {
		cfg, err := config.Resolve(flags)
		if err != nil {
			return err
		}
		ctx := &context{out: out, errOut: errOut, flags: flags, cfg: cfg, loc: i18n.New(cfg.Lang)}
		return renderVersion(ctx)
	}
	if wantsHelp(args) {
		if objectStorageCreateHelpRequiresConfig(args) {
			cfg, err := config.Resolve(flags)
			if err != nil {
				return err
			}
			ctx := &context{out: out, errOut: errOut, flags: flags, cfg: cfg, loc: i18n.New(cfg.Lang)}
			return executeRootCommand(args, ctx)
		}
		lang, err := config.ResolveLang(flags)
		if err != nil {
			return err
		}
		ctx := &context{out: out, errOut: errOut, flags: flags, loc: i18n.New(lang)}
		return executeRootCommand(args, ctx)
	}
	cfg, err := config.Resolve(flags)
	if err != nil {
		return err
	}
	ctx := &context{out: out, errOut: errOut, flags: flags, cfg: cfg, loc: i18n.New(cfg.Lang)}
	return executeRootCommand(args, ctx)
}

func objectStorageCreateHelpRequiresConfig(args []string) bool {
	commandArgs := args
	if len(commandArgs) > 0 && commandArgs[0] == "help" {
		commandArgs = commandArgs[1:]
	}
	if len(commandArgs) < 2 || commandArgs[0] != "oss" || commandArgs[1] != "create" {
		return false
	}
	provider := ""
	for i := 2; i < len(commandArgs); i++ {
		name, value, inline := splitFlag(commandArgs[i])
		if name != "--provider" {
			continue
		}
		if !inline && i+1 < len(commandArgs) {
			value = commandArgs[i+1]
			i++
		}
		provider = strings.ToLower(strings.TrimSpace(value))
	}
	return provider != "custom"
}

func executeRootCommand(args []string, ctx *context) error {
	root := newRootCommand(ctx)
	root.SetOut(ctx.out)
	root.SetErr(ctx.errOut)
	root.SetArgs(args)
	root.SilenceUsage = true
	root.SilenceErrors = true
	return root.Execute()
}

func extractGlobalFlags(args []string) ([]string, config.Flags, error) {
	var flags config.Flags
	filtered := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		key, value, hasValue := splitFlag(arg)
		switch key {
		case "--lang":
			v, next, err := flagValue(args, i, value, hasValue)
			if err != nil {
				return nil, flags, err
			}
			flags.Lang = v
			i = next
		case "--output", "-o":
			v, next, err := flagValue(args, i, value, hasValue)
			if err != nil {
				return nil, flags, err
			}
			flags.Output = v
			i = next
		case "--vertical", "-G":
			flags.Vertical = true
			if hasValue {
				flags.Vertical = value == "true" || value == "1"
			}
		case "--debug":
			flags.Debug = true
			flags.DebugSet = true
			if hasValue {
				flags.Debug = value == "true" || value == "1"
			}
		default:
			filtered = append(filtered, arg)
		}
	}
	return filtered, flags, nil
}

func splitFlag(arg string) (key, value string, hasValue bool) {
	if strings.HasPrefix(arg, "--") || arg == "-o" || arg == "-G" {
		if idx := strings.Index(arg, "="); idx >= 0 {
			return arg[:idx], arg[idx+1:], true
		}
	}
	return arg, "", false
}

func hasFlagToken(args []string, name string) bool {
	for _, arg := range args {
		if arg == name || strings.HasPrefix(arg, name+"=") {
			return true
		}
	}
	return false
}

func errUnknownLongFlag(name string) error {
	return fmt.Errorf("unknown flag: --%s", name)
}

func flagValue(args []string, idx int, inline string, hasInline bool) (string, int, error) {
	if hasInline {
		return inline, idx, nil
	}
	if idx+1 >= len(args) {
		return "", idx, fmt.Errorf("%s requires value", args[idx])
	}
	return args[idx+1], idx + 1, nil
}

func wantsHelp(args []string) bool {
	if len(args) == 0 {
		return true
	}
	return hasHelpToken(args)
}

func wantsVersion(args []string) bool {
	return len(args) == 1 && args[0] == "--version"
}

func hasHelpToken(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "help", "--help", "-h":
			return true
		}
	}
	return false
}

func renderVersion(ctx *context) error {
	info := version.Current()
	if ctx.cfg.Output == "json" {
		return output.JSON(ctx.out, info)
	}
	_, err := fmt.Fprintf(ctx.out, "hyperbdrctl version %s\n", info.Version)
	return err
}

func runConfig(ctx *context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("config requires get or set")
	}
	switch args[0] {
	case "get":
		fs := newFlagSet("config get")
		showSecret := fs.Bool("show-secret", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		cfg := ctx.cfg.Config
		password := cfg.Password
		if password != "" && !*showSecret {
			password = "********"
		}
		return writeValue(ctx, map[string]interface{}{
			"host":        cfg.Host,
			"username":    cfg.Username,
			"password":    password,
			"scene":       cfg.Scene,
			"lang":        cfg.Lang,
			"insecure":    cfg.Insecure,
			"output":      cfg.Output,
			"debug":       cfg.Debug,
			"config_path": ctx.cfg.ConfigPath,
			"token_path":  ctx.cfg.CachePath,
		})
	case "set":
		fs := newFlagSet("config set")
		host := fs.String("host", "", "")
		username := fs.String("username", "", "")
		password := fs.String("password", "", "")
		scene := fs.String("scene", "", "")
		insecure := fs.Bool("insecure", false, "")
		insecureSet := hasFlagToken(args[1:], "--insecure")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		current := ctx.cfg.Config
		if *host != "" {
			current.Host = *host
		}
		if *username != "" {
			current.Username = *username
		}
		if *password != "" {
			current.Password = *password
		}
		if *scene != "" {
			current.Scene = *scene
		}
		if insecureSet {
			current.Insecure = *insecure
		}
		candidateConfig, err := config.Normalize(current)
		if err != nil {
			return err
		}
		candidate := config.Resolved{Config: candidateConfig, ConfigPath: ctx.cfg.ConfigPath, CachePath: ctx.cfg.CachePath}
		loginClient, err := client.NewWithDebug(candidate, ctx.errOut)
		if err != nil {
			return err
		}
		if _, err := loginClient.Login(); err != nil {
			return err
		}
		if err := config.Save(ctx.cfg.ConfigPath, candidateConfig); err != nil {
			return err
		}
		fmt.Fprintln(ctx.out, ctx.loc.T("config.saved"))
		return nil
	default:
		return fmt.Errorf("unknown config command %q", args[0])
	}
}

func ensureClient(ctx *context) (*client.Client, error) {
	if ctx.client != nil {
		return ctx.client, nil
	}
	c, err := client.NewWithDebug(ctx.cfg, ctx.errOut)
	if err != nil {
		return nil, err
	}
	ctx.client = c
	return c, nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func parseQueryFlagSet(fs *flag.FlagSet, args []string) (url.Values, error) {
	return parseQueryFlagSetWithPassthrough(fs, args, false)
}

func parseQueryFlagSetPassthrough(fs *flag.FlagSet, args []string) (url.Values, error) {
	return parseQueryFlagSetWithPassthrough(fs, args, true)
}

func parseQueryFlagSetWithPassthrough(fs *flag.FlagSet, args []string, allowUnknown bool) (url.Values, error) {
	known := map[string]bool{}
	fs.VisitAll(func(f *flag.Flag) {
		known[f.Name] = true
	})
	filtered := make([]string, 0, len(args))
	extra := url.Values{}
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
		extra.Set(strings.ReplaceAll(name, "-", "_"), value)
	}
	if err := fs.Parse(filtered); err != nil {
		return nil, err
	}
	return extra, nil
}

func parseQueryFlagsInto(fs *flag.FlagSet, args []string, q url.Values) error {
	extra, err := parseQueryFlagSet(fs, args)
	if err != nil {
		return err
	}
	mergeQuery(q, extra)
	return nil
}

func parseQueryFlagsIntoPassthrough(fs *flag.FlagSet, args []string, q url.Values) error {
	extra, err := parseQueryFlagSetPassthrough(fs, args)
	if err != nil {
		return err
	}
	mergeQuery(q, extra)
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

func mergeQuery(dst url.Values, src url.Values) {
	for key, values := range src {
		for _, value := range values {
			dst.Set(key, value)
		}
	}
}

func addInt(q url.Values, key string, value int) {
	if value > 0 {
		q.Set(key, fmt.Sprintf("%d", value))
	}
}

func addString(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

func addBool(q url.Values, key string, value bool) {
	if value {
		q.Set(key, "true")
	}
}

func fetch(ctx *context, path string, q url.Values) (client.APIResponse, error) {
	c, err := ensureClient(ctx)
	if err != nil {
		return client.APIResponse{}, err
	}
	return c.Get(path, q)
}

func post(ctx *context, path string, body interface{}) (client.APIResponse, error) {
	c, err := ensureClient(ctx)
	if err != nil {
		return client.APIResponse{}, err
	}
	return c.Post(path, body)
}

func deleteWithBody(ctx *context, path string, body interface{}) (client.APIResponse, error) {
	c, err := ensureClient(ctx)
	if err != nil {
		return client.APIResponse{}, err
	}
	return c.Delete(path, body)
}

func writeResponse(ctx *context, resp client.APIResponse, listKey string, columns []output.Column) error {
	if ctx.cfg.Output == "json" {
		if resp.Raw != nil {
			return output.JSON(ctx.out, resp.Raw)
		}
		return output.JSON(ctx.out, resp.Data)
	}
	if listKey != "" {
		rows := listFromData(resp.Data, listKey)
		return writeRows(ctx, rows, columns)
	}
	if m := mapFromData(resp.Data); m != nil {
		return writeHumanValue(ctx, m)
	}
	if resp.Data == nil && resp.Raw != nil {
		return writeHumanValue(ctx, resp.Raw)
	}
	return writeHumanValue(ctx, resp.Data)
}

func writeValue(ctx *context, value interface{}) error {
	if ctx.cfg.Output == "json" {
		return output.JSON(ctx.out, value)
	}
	return writeHumanValue(ctx, value)
}

func writeRows(ctx *context, rows []map[string]interface{}, columns []output.Column) error {
	if ctx.flags.Vertical {
		return output.Vertical(ctx.out, ctx.loc, rows, columns)
	}
	return output.Table(ctx.out, ctx.loc, rows, columns)
}

func writeHumanValue(ctx *context, value interface{}) error {
	if unwrapped, ok := unwrapSingleNestedValue(value); ok {
		return output.JSON(ctx.out, unwrapped)
	}
	if m, ok := value.(map[string]interface{}); ok {
		if mapHasOnlyScalarValues(m) {
			return output.KeyValue(ctx.out, ctx.loc, m)
		}
		return output.JSON(ctx.out, m)
	}
	return output.JSON(ctx.out, value)
}

func unwrapSingleNestedValue(value interface{}) (interface{}, bool) {
	m, ok := value.(map[string]interface{})
	if !ok || len(m) != 1 {
		return nil, false
	}
	for _, nested := range m {
		if !isScalarValue(nested) {
			return nested, true
		}
	}
	return nil, false
}

func mapHasOnlyScalarValues(m map[string]interface{}) bool {
	for _, value := range m {
		if !isScalarValue(value) {
			return false
		}
	}
	return true
}

func isScalarValue(value interface{}) bool {
	if value == nil {
		return true
	}
	kind := reflect.TypeOf(value).Kind()
	switch kind {
	case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
		return false
	default:
		return true
	}
}

func listFromData(data interface{}, key string) []map[string]interface{} {
	if items, ok := data.([]interface{}); ok {
		rows := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			if row, ok := item.(map[string]interface{}); ok {
				rows = append(rows, row)
			}
		}
		return rows
	}
	m := mapFromData(data)
	if m == nil {
		return nil
	}
	items, ok := findListItems(m, key)
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

func mapFromData(data interface{}) map[string]interface{} {
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func findListItems(m map[string]interface{}, preferred string) ([]interface{}, bool) {
	keys := []string{
		preferred,
		"hosts",
		"targets",
		"tasks",
		"sources",
		"vms",
		"cloud_accounts",
		"cloudAccounts",
		"storages",
		"snapshots",
		"policies",
		"strategies",
		"networks",
		"regions",
		"zones",
		"steps",
		"pages",
		"items",
		"list",
	}
	seen := map[string]bool{}
	for _, key := range keys {
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		if items, ok := m[key].([]interface{}); ok {
			return items, true
		}
	}
	for _, value := range m {
		if items, ok := value.([]interface{}); ok {
			return items, true
		}
	}
	return nil, false
}
