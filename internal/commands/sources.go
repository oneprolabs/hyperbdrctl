package commands

import (
	"fmt"
	"reflect"
	"strings"

	appsource "hyperbdr-client/internal/app/source"
	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/output"
)

func runSources(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("source", "")
	}
	service := appsource.NewService(commandAPIAdapter{ctx: ctx})
	createService := appsource.NewCreateService(commandPosterAdapter{ctx: ctx})
	switch args[0] {
	case "list":
		fs := newFlagSet("source list")
		sourceType := fs.String("type", "", "")
		kw := fs.String("kw", "", "")
		bindingStatus := fs.String("binding-status", "", "")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 10, "")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		if *sourceType == "" {
			return missing(ctx, "error.missing_source_type")
		}
		if *sourceType == "agent" {
			return fmt.Errorf(ctx.loc.T("error.sources_list_agent_install"))
		}
		resp, err := service.List(appsource.ListSpec{
			SourceType:    *sourceType,
			KW:            *kw,
			BindingStatus: *bindingStatus,
			Page:          *page,
			PageSize:      *pageSize,
			Query:         q,
		})
		if err != nil {
			return err
		}
		return writeSourceListResponse(ctx, resp)
	case "detail":
		fs := newFlagSet("source detail")
		id := fs.String("id", "", "")
		sourceType := fs.String("type", "", "")
		bindingStatus := fs.String("binding-status", "", "")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 10, "")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		if *id == "" {
			return missing(ctx, "error.missing_id")
		}
		resp, err := service.Detail(appsource.DetailSpec{
			ID:            *id,
			SourceType:    *sourceType,
			BindingStatus: *bindingStatus,
			Page:          *page,
			PageSize:      *pageSize,
			Query:         q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	case "vms":
		fs := newFlagSet("source vms")
		connectionType := fs.String("connection-type", "", "")
		connectionUUID := fs.String("connection-uuid", "", "")
		registered := fs.String("registered", "", "")
		kw := fs.String("kw", "", "")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 10, "")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		if *connectionType == "" {
			return missing(ctx, "error.missing_connection_type")
		}
		resp, err := service.VMs(appsource.VMsSpec{
			ConnectionType: *connectionType,
			ConnectionUUID: *connectionUUID,
			Registered:     *registered,
			KW:             *kw,
			Page:           *page,
			PageSize:       *pageSize,
			Query:          q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "vms", vmColumns())
	case "agent-install":
		fs := newFlagSet("source agent-install")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.AgentInstall(appsource.AgentInstallSpec{Query: q})
		if err != nil {
			return err
		}
		return writeAgentInstallResponse(ctx, resp)
	case "agentless-install":
		fs := newFlagSet("source agentless-install")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.AgentlessInstall(appsource.AgentlessInstallSpec{Query: q})
		if err != nil {
			return err
		}
		return writeAgentlessInstallResponse(ctx, resp)
	case "sync-nodes":
		fs := newFlagSet("source sync-nodes")
		nodeType := fs.String("type", "proxy", "")
		status := fs.String("status", "online", "")
		q := queryFromPairs()
		if err := parseQueryFlagsIntoPassthrough(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.SynchNodes(appsource.SynchNodesSpec{
			Type:   *nodeType,
			Status: *status,
			Query:  q,
		})
		if err != nil {
			return err
		}
		return writeSyncNodesResponse(ctx, resp)
	case "create":
		fs := newFlagSet("source create")
		sourceType := fs.String("type", "", "")
		synchNodeID := fs.String("synch-node-id", "", "")
		synchNodeIDs := fs.String("synch-node-ids", "", "")
		authURL := fs.String("auth-url", "", "")
		authKey := fs.String("auth-key", "", "")
		authCert := fs.String("auth-cert", "", "")
		regionID := fs.String("region-id", "", "")
		previewRequest := fs.Bool("preview-request", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		spec := appsource.CreateSpec{
			Type:         *sourceType,
			SynchNodeID:  *synchNodeID,
			SynchNodeIDs: *synchNodeIDs,
			AuthURL:      *authURL,
			AuthKey:      *authKey,
			AuthCert:     *authCert,
			RegionID:     *regionID,
		}
		if *previewRequest {
			prepared, err := createService.PrepareCreate(spec)
			if err != nil {
				return err
			}
			return output.JSON(ctx.out, prepared.Body)
		}
		resp, err := createService.Create(spec)
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "", nil)
	default:
		return errUnknown("source", args[0])
	}
}

func sourceColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.uuid", Field: "uuid"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.type", Field: "type"},
		{HeaderKey: "table.status", Field: "status"},
	}
}

func writeSourceListResponse(ctx *context, resp client.APIResponse) error {
	if ctx.cfg.Output == "json" {
		return writeResponse(ctx, resp, "", nil)
	}
	return writeRows(ctx, sourceListRows(resp), sourceColumns())
}

func sourceListRows(resp client.APIResponse) []map[string]interface{} {
	rows := listFromData(resp.Data, "sources")
	if len(rows) == 0 {
		rows = sourceListRowsFromRaw(resp.Raw)
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		normalized := map[string]interface{}{}
		for key, value := range row {
			normalized[key] = value
		}
		normalized["id"] = mapString(row, "id", "uuid")
		normalized["uuid"] = mapString(row, "uuid", "id")
		normalized["name"] = mapString(row, "name", "displayname", "display_auth_url", "auth_url")
		normalized["type"] = normalizeDisplayedSourceType(mapString(row, "type", "display_type"))
		normalized["status"] = firstNonEmptyString(mapString(row, "display_source_status"), mapString(row, "status"))
		out = append(out, normalized)
	}
	return out
}

func sourceListRowsFromRaw(raw map[string]interface{}) []map[string]interface{} {
	if raw == nil {
		return nil
	}
	sources, ok := mapValue(raw["sources"])
	if !ok {
		return nil
	}
	items := listValue(sources["data"])
	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		row, ok := mapValue(item)
		if !ok {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

func vmColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.status", Field: "status"},
		{HeaderKey: "table.os_type", Field: "os_type"},
	}
}

func writeSyncNodesResponse(ctx *context, resp client.APIResponse) error {
	if ctx.cfg.Output == "json" {
		return writeResponse(ctx, resp, "", nil)
	}
	return renderSyncNodeCards(ctx, syncNodeViews(resp.Data))
}

type syncNodeView struct {
	ID          string
	Name        string
	HostIP      string
	Version     string
	Status      string
	Connections []syncNodeConnectionView
}

type syncNodeConnectionView struct {
	Type    string
	Status  string
	Address string
	UUID    string
}

func syncNodeViews(data interface{}) []syncNodeView {
	rows := listFromData(data, "synch_nodes")
	out := make([]syncNodeView, 0, len(rows))
	for _, row := range rows {
		out = append(out, syncNodeView{
			ID:          mapString(row, "id", "uuid"),
			Name:        mapString(row, "name", "node_name", "host_ip"),
			HostIP:      mapString(row, "host_ip"),
			Version:     mapString(row, "version"),
			Status:      firstNonEmptyString(mapString(row, "display_status"), mapString(row, "status")),
			Connections: syncNodeConnections(row["connections"]),
		})
	}
	return out
}

func syncNodeConnections(value interface{}) []syncNodeConnectionView {
	items := listValue(value)
	out := make([]syncNodeConnectionView, 0, len(items))
	for _, item := range items {
		row, ok := mapValue(item)
		if !ok {
			continue
		}
		out = append(out, syncNodeConnectionView{
			Type:    normalizeDisplayedSourceType(mapString(row, "type")),
			Status:  mapString(row, "status"),
			Address: mapString(row, "auth_url"),
			UUID:    mapString(row, "uuid"),
		})
	}
	return out
}

func normalizeDisplayedSourceType(sourceType string) string {
	switch strings.ToLower(strings.TrimSpace(sourceType)) {
	case "vsphere":
		return "vmware"
	default:
		return sourceType
	}
}

func renderSyncNodeCards(ctx *context, nodes []syncNodeView) error {
	for idx, node := range nodes {
		if err := writeAgentInstallSectionTitle(ctx, fmt.Sprintf("%s %d", ctx.loc.T("label.sync_nodes.node"), idx+1)); err != nil {
			return err
		}
		if err := output.OrderedKeyValue(ctx.out, ctx.loc, []output.KeyValueRow{
			{Key: ctx.loc.T("table.id"), Value: node.ID},
			{Key: ctx.loc.T("table.name"), Value: node.Name},
			{Key: ctx.loc.T("table.node_ip"), Value: node.HostIP},
			{Key: ctx.loc.T("table.version"), Value: node.Version},
			{Key: ctx.loc.T("table.status"), Value: node.Status},
		}); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(ctx.out, "\n%s:\n", ctx.loc.T("label.sync_nodes.connections")); err != nil {
			return err
		}
		if len(node.Connections) == 0 {
			if _, err := fmt.Fprintf(ctx.out, "  - %s\n", ctx.loc.T("label.sync_nodes.none")); err != nil {
				return err
			}
		} else {
			for i, conn := range node.Connections {
				title := conn.Type
				if title == "" {
					title = fmt.Sprintf("%s %d", ctx.loc.T("label.sync_nodes.connection"), i+1)
				}
				if _, err := fmt.Fprintf(ctx.out, "%d. %s\n", i+1, title); err != nil {
					return err
				}
				if conn.Status != "" {
					if _, err := fmt.Fprintf(ctx.out, "   %s  %s\n", ctx.loc.T("table.status"), conn.Status); err != nil {
						return err
					}
				}
				if conn.Address != "" {
					if _, err := fmt.Fprintf(ctx.out, "   %s  %s\n", ctx.loc.T("label.sync_nodes.address"), conn.Address); err != nil {
						return err
					}
				}
				if conn.UUID != "" {
					if _, err := fmt.Fprintf(ctx.out, "   %s    %s\n", ctx.loc.T("table.uuid"), conn.UUID); err != nil {
						return err
					}
				}
			}
		}
		if idx < len(nodes)-1 {
			if _, err := fmt.Fprintln(ctx.out); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeAgentlessInstallResponse(ctx *context, resp client.APIResponse) error {
	if ctx.cfg.Output == "json" {
		return writeResponse(ctx, resp, "", nil)
	}
	view, ok := buildAgentlessInstallView(agentlessInstallResponseData(resp))
	if !ok {
		return renderAgentlessInstallEmpty(ctx)
	}
	return renderAgentlessInstallView(ctx, view)
}

func writeAgentInstallResponse(ctx *context, resp client.APIResponse) error {
	if ctx.cfg.Output == "json" {
		return writeResponse(ctx, resp, "", nil)
	}
	view, ok := buildAgentInstallView(ctx, agentInstallResponseData(resp))
	if !ok {
		return renderAgentInstallEmpty(ctx)
	}
	return renderAgentInstallView(ctx, view)
}

func agentInstallResponseData(resp client.APIResponse) interface{} {
	if resp.Data != nil {
		return resp.Data
	}
	if resp.Raw != nil {
		return resp.Raw
	}
	return nil
}

func agentlessInstallResponseData(resp client.APIResponse) interface{} {
	if resp.Raw != nil {
		return resp.Raw
	}
	if resp.Data != nil {
		return resp.Data
	}
	return nil
}

func buildAgentInstallView(ctx *context, data interface{}) (agentInstallView, bool) {
	agent, ok := agentInstallPayload(data)
	if !ok {
		return agentInstallView{}, false
	}

	view := agentInstallView{
		Title:       stringMapValue(agent, "title"),
		Description: stringMapValue(agent, "description"),
	}

	if linux, ok := mapValue(agent["linux"]); ok {
		supportedSystems := collectSystemNames(ctx, nestedList(linux, "system_version", "url_list"), false)
		view.Linux = agentInstallPlatformView{
			Title:            firstNonEmpty(stringMapValue(linux, "title"), ctx.loc.T("value.agent_install.platform.linux")),
			Description:      stringMapValue(linux, "description"),
			SupportedSystems: supportedSystems,
			Recommendations:  buildLinuxAgentInstallRecommendations(ctx, supportedSystems),
			Commands: []agentInstallCommandRow{
				{
					Variant: ctx.loc.T("value.agent_install.variant.default"),
					Command: stringMapValue(linux, "install_link"),
				},
				{
					Variant: ctx.loc.T("value.agent_install.variant.dkms"),
					Command: stringMapValue(linux, "install_link_dkms"),
				},
			},
		}
	}

	if windows, ok := mapValue(agent["windows"]); ok {
		items := nestedList(windows, "system_version", "url_list")
		supportedSystems := collectSystemNames(ctx, items, true)
		view.Windows = agentInstallPlatformView{
			Title:            firstNonEmpty(stringMapValue(windows, "title"), ctx.loc.T("value.agent_install.platform.windows")),
			Description:      stringMapValue(windows, "description"),
			SupportedSystems: supportedSystems,
			Recommendations:  buildWindowsAgentInstallRecommendations(ctx, supportedSystems),
			Commands: append([]agentInstallCommandRow{
				{
					Variant: ctx.loc.T("value.agent_install.variant.install_cmd"),
					Command: windowsInstallCommand(items),
				},
			}, windowsDownloadCommands(ctx, items)...),
		}
	}

	if !view.HasContent() {
		return agentInstallView{}, false
	}
	return view, true
}

type agentlessInstallView struct {
	Title       string
	Description string
	Command     string
}

func buildAgentlessInstallView(data interface{}) (agentlessInstallView, bool) {
	root, ok := mapValue(data)
	if !ok {
		return agentlessInstallView{}, false
	}
	if payload, ok := mapValue(root["agentless"]); ok {
		root = payload
	}
	title := stringMapValue(root, "title")
	view := agentlessInstallView{
		Title:       firstNonEmpty(title, "Agentless Install"),
		Description: stringMapValue(root, "description"),
		Command:     firstNonEmpty(stringMapValue(root, "install_link"), stringMapValue(root, "link"), stringMapValue(root, "command")),
	}
	if view.Command == "" {
		if step, ok := agentlessInstallStep(root["data"]); ok {
			view.Command = firstNonEmpty(stringMapValue(step, "install_link"), stringMapValue(step, "link"), stringMapValue(step, "command"))
			if view.Description == "" {
				view.Description = firstNonEmpty(stringMapValue(step, "description"), stringMapValue(step, "install_title"))
			}
			if title == "" {
				view.Title = stringMapValue(step, "title")
			}
		}
	}
	if view.Command == "" && root["linux"] != nil {
		if linux, ok := mapValue(root["linux"]); ok {
			view.Command = firstNonEmpty(stringMapValue(linux, "install_link"), stringMapValue(linux, "link"), stringMapValue(linux, "command"))
			if view.Description == "" {
				view.Description = stringMapValue(linux, "description")
			}
			if title := stringMapValue(linux, "title"); title != "" {
				view.Title = title
			}
		}
	}
	if view.Command == "" {
		return agentlessInstallView{}, false
	}
	return view, true
}

func agentlessInstallStep(value interface{}) (map[string]interface{}, bool) {
	items := listValue(value)
	var fallback map[string]interface{}
	for _, item := range items {
		row, ok := mapValue(item)
		if !ok {
			continue
		}
		command := firstNonEmpty(stringMapValue(row, "install_link"), stringMapValue(row, "link"), stringMapValue(row, "command"))
		if command == "" {
			continue
		}
		if fallback == nil {
			fallback = row
		}
		lowerCommand := strings.ToLower(command)
		if strings.Contains(lowerCommand, "bash") || strings.Contains(lowerCommand, "powershell") {
			return row, true
		}
		if stringMapValue(row, "install_title") != "" {
			return row, true
		}
		if _, ok := mapValue(row["system_version"]); ok {
			return row, true
		}
	}
	if fallback != nil {
		return fallback, true
	}
	return nil, false
}

func agentInstallPayload(data interface{}) (map[string]interface{}, bool) {
	root, ok := mapValue(data)
	if !ok {
		return nil, false
	}
	if agent, ok := mapValue(root["agent"]); ok {
		return agent, true
	}
	if _, hasLinux := root["linux"]; hasLinux {
		return root, true
	}
	if _, hasWindows := root["windows"]; hasWindows {
		return root, true
	}
	return nil, false
}

type agentInstallView struct {
	Title       string
	Description string
	Linux       agentInstallPlatformView
	Windows     agentInstallPlatformView
}

func (v agentInstallView) HasContent() bool {
	return v.Title != "" || v.Description != "" || v.Linux.HasContent() || v.Windows.HasContent()
}

type agentInstallPlatformView struct {
	Title            string
	Description      string
	SupportedSystems []string
	Recommendations  []string
	Commands         []agentInstallCommandRow
}

type agentInstallCommandRow struct {
	Variant string
	Command string
}

func (v agentInstallPlatformView) HasContent() bool {
	return v.Title != "" || v.Description != "" || len(v.SupportedSystems) > 0 || len(v.Recommendations) > 0 || len(v.CommandRows()) > 0
}

func (v agentInstallPlatformView) CommandRows() []agentInstallCommandRow {
	rows := make([]agentInstallCommandRow, 0, len(v.Commands))
	for _, row := range v.Commands {
		if row.Command == "" {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

func (v agentInstallPlatformView) BestChoice(empty string) string {
	rows := v.CommandRows()
	if len(rows) == 0 {
		return empty
	}
	return rows[0].Variant
}

func (v agentInstallPlatformView) VariantSummary(empty string) string {
	rows := v.CommandRows()
	if len(rows) == 0 {
		return empty
	}
	variants := make([]string, 0, len(rows))
	for _, row := range rows {
		variants = append(variants, row.Variant)
	}
	return strings.Join(variants, ", ")
}

func renderAgentInstallView(ctx *context, view agentInstallView) error {
	if err := writeAgentInstallSectionTitle(ctx, firstNonEmpty(view.Title, ctx.loc.T("label.agent_install.overview"))); err != nil {
		return err
	}
	if view.Description != "" {
		if _, err := fmt.Fprintln(ctx.out, view.Description); err != nil {
			return err
		}
	}
	wroteCard := false
	if err := writeAgentInstallPlatformCard(ctx, view.Linux, wroteCard); err != nil {
		return err
	}
	if view.Linux.HasContent() {
		wroteCard = true
	}
	return writeAgentInstallPlatformCard(ctx, view.Windows, wroteCard)
}

func writeAgentInstallPlatformCard(ctx *context, view agentInstallPlatformView, needsLeadingSpace bool) error {
	if !view.HasContent() {
		return nil
	}
	if needsLeadingSpace || view.Description != "" {
		if _, err := fmt.Fprintln(ctx.out); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(ctx.out, "+ %s\n", view.Title); err != nil {
		return err
	}
	if view.Description != "" {
		if _, err := fmt.Fprintf(ctx.out, "  %s: %s\n", ctx.loc.T("label.agent_install.description"), view.Description); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(ctx.out, "  %s: %s\n", ctx.loc.T("label.agent_install.best_choice"), view.BestChoice(ctx.loc.T("label.agent_install.none"))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(ctx.out, "  %s: %s\n", ctx.loc.T("label.agent_install.variants"), view.VariantSummary(ctx.loc.T("label.agent_install.none"))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(ctx.out, "  %s:\n", ctx.loc.T("label.agent_install.supported_systems")); err != nil {
		return err
	}
	if len(view.SupportedSystems) == 0 {
		if _, err := fmt.Fprintf(ctx.out, "    - %s\n", ctx.loc.T("label.agent_install.none")); err != nil {
			return err
		}
	} else {
		for _, system := range view.SupportedSystems {
			if _, err := fmt.Fprintf(ctx.out, "    - %s\n", system); err != nil {
				return err
			}
		}
	}
	if len(view.Recommendations) > 0 {
		if _, err := fmt.Fprintf(ctx.out, "  %s:\n", ctx.loc.T("label.agent_install.recommendations")); err != nil {
			return err
		}
		for _, recommendation := range view.Recommendations {
			if _, err := fmt.Fprintf(ctx.out, "    - %s\n", recommendation); err != nil {
				return err
			}
		}
	}
	for _, row := range view.CommandRows() {
		if _, err := fmt.Fprintln(ctx.out); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(ctx.out, "  [%s]\n", row.Variant); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(ctx.out, "  %s\n", row.Command); err != nil {
			return err
		}
	}
	return nil
}

func renderAgentInstallEmpty(ctx *context) error {
	if err := writeAgentInstallSectionTitle(ctx, ctx.loc.T("label.agent_install.overview")); err != nil {
		return err
	}
	return output.OrderedKeyValue(ctx.out, ctx.loc, []output.KeyValueRow{
		{Key: ctx.loc.T("label.agent_install.description"), Value: ctx.loc.T("label.agent_install.empty")},
	})
}

func renderAgentlessInstallView(ctx *context, view agentlessInstallView) error {
	if err := writeAgentInstallSectionTitle(ctx, firstNonEmpty(view.Title, ctx.loc.T("label.agentless_install.overview"))); err != nil {
		return err
	}
	rows := []output.KeyValueRow{}
	if view.Description != "" {
		rows = append(rows, output.KeyValueRow{
			Key:   ctx.loc.T("label.agentless_install.description"),
			Value: view.Description,
		})
	}
	rows = append(rows, []output.KeyValueRow{
		{Key: ctx.loc.T("label.agentless_install.os"), Value: ctx.loc.T("value.agentless_install.os")},
		{Key: ctx.loc.T("label.agentless_install.minimum"), Value: ctx.loc.T("value.agentless_install.minimum")},
		{Key: ctx.loc.T("label.agentless_install.filesystem"), Value: ctx.loc.T("value.agentless_install.filesystem")},
		{Key: ctx.loc.T("label.agentless_install.network"), Value: ctx.loc.T("value.agentless_install.network")},
		{Key: ctx.loc.T("label.agentless_install.user"), Value: ctx.loc.T("value.agentless_install.user")},
	}...)
	rows = append(rows, output.KeyValueRow{
		Key:   ctx.loc.T("label.agentless_install.command"),
		Value: view.Command,
	})
	return output.OrderedKeyValue(ctx.out, ctx.loc, rows)
}

func renderAgentlessInstallEmpty(ctx *context) error {
	if err := writeAgentInstallSectionTitle(ctx, ctx.loc.T("label.agentless_install.overview")); err != nil {
		return err
	}
	return output.OrderedKeyValue(ctx.out, ctx.loc, []output.KeyValueRow{
		{Key: ctx.loc.T("label.agentless_install.description"), Value: ctx.loc.T("label.agentless_install.empty")},
	})
}

func writeAgentInstallSectionTitle(ctx *context, title string) error {
	if title == "" {
		title = ctx.loc.T("label.agent_install.overview")
	}
	return writeSectionTitle(ctx, title)
}

func nestedList(root map[string]interface{}, keys ...string) []interface{} {
	current := interface{}(root)
	for _, key := range keys {
		nextMap, ok := mapValue(current)
		if !ok {
			return nil
		}
		current = nextMap[key]
	}
	return listValue(current)
}

func collectSystemNames(ctx *context, items []interface{}, skipInstallCmd bool) []string {
	names := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		row, ok := mapValue(item)
		if !ok {
			continue
		}
		system := stringMapValue(row, "system")
		if system == "" || system == "稳定版本" || system == "测试版本" {
			continue
		}
		if invalidAgentInstallSystem(system, skipInstallCmd) || seen[system] {
			continue
		}
		seen[system] = true
		names = append(names, normalizeAgentInstallSystemName(ctx, system))
	}
	return names
}

func normalizeAgentInstallSystemName(ctx *context, system string) string {
	switch system {
	case "Windows_server_32bit":
		return ctx.loc.T("value.agent_install.system.windows_x86")
	case "Windows_server_64bit":
		return ctx.loc.T("value.agent_install.system.windows_x64")
	default:
		return system
	}
}

func buildLinuxAgentInstallRecommendations(ctx *context, systems []string) []string {
	if len(systems) == 0 {
		return nil
	}
	return []string{
		ctx.loc.T("value.agent_install.recommendation.linux_ubuntu_dkms"),
		ctx.loc.T("value.agent_install.recommendation.linux_centos_dkms"),
	}
}

func buildWindowsAgentInstallRecommendations(ctx *context, systems []string) []string {
	if len(systems) == 0 {
		return nil
	}
	return []string{ctx.loc.T("value.agent_install.recommendation.windows_install_cmd")}
}

func hasUbuntuDKMSRecommendedSystem(systems []string) bool {
	for _, system := range systems {
		if strings.Contains(system, "Ubuntu 20.04") || strings.Contains(system, "Ubuntu 22.04") || strings.Contains(system, "Ubuntu 24.04") {
			return true
		}
	}
	return false
}

func hasCentOSDKMSRecommendedSystem(systems []string) bool {
	for _, system := range systems {
		if strings.Contains(system, "CentOS 8") || strings.Contains(system, "CentOS 9") || strings.Contains(system, "CentOS Linux 8") || strings.Contains(system, "CentOS Linux 9") {
			return true
		}
		if strings.Contains(system, "CentOS Linux 6/7/8") || strings.Contains(system, "CentOS Linux 7/8/9") || strings.Contains(system, "CentOS Linux 8/9") {
			return true
		}
	}
	return false
}

func hasWindowsInstallCommandRecommendedSystem(systems []string) bool {
	for _, system := range systems {
		if strings.Contains(system, "Windows Server 2019") || strings.Contains(system, "Windows Server 2022") || strings.Contains(system, "Windows Server 2025") {
			return true
		}
		if strings.Contains(system, "Windows_server_2019") || strings.Contains(system, "Windows_server_2022") || strings.Contains(system, "Windows_server_2025") {
			return true
		}
	}
	return false
}

func invalidAgentInstallSystem(system string, skipInstallCmd bool) bool {
	if system == "" || system == "绋冲畾鐗堟湰" || system == "娴嬭瘯鐗堟湰" {
		return true
	}
	if skipInstallCmd && system == "Windows_install_cmd" {
		return true
	}
	return false
}

func windowsInstallCommand(items []interface{}) string {
	for _, item := range items {
		row, ok := mapValue(item)
		if !ok {
			continue
		}
		if stringMapValue(row, "system") == "Windows_install_cmd" {
			return stringMapValue(row, "link")
		}
	}
	return ""
}

func windowsDownloadCommands(ctx *context, items []interface{}) []agentInstallCommandRow {
	rows := make([]agentInstallCommandRow, 0, 2)
	if link := windowsSystemLink(items, "Windows_server_32bit"); link != "" {
		rows = append(rows, agentInstallCommandRow{
			Variant: ctx.loc.T("value.agent_install.system.windows_x86"),
			Command: link,
		})
	}
	if link := windowsSystemLink(items, "Windows_server_64bit"); link != "" {
		rows = append(rows, agentInstallCommandRow{
			Variant: ctx.loc.T("value.agent_install.system.windows_x64"),
			Command: link,
		})
	}
	return rows
}

func windowsSystemLink(items []interface{}, systemName string) string {
	for _, item := range items {
		row, ok := mapValue(item)
		if !ok {
			continue
		}
		if stringMapValue(row, "system") == systemName {
			return stringMapValue(row, "link")
		}
	}
	return ""
}

func stringMapValue(m map[string]interface{}, key string) string {
	value, _ := m[key].(string)
	return value
}

func mapValue(value interface{}) (map[string]interface{}, bool) {
	if value == nil {
		return nil, false
	}
	if m, ok := value.(map[string]interface{}); ok {
		return m, true
	}
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || rv.Kind() != reflect.Map || rv.Type().Key().Kind() != reflect.String {
		return nil, false
	}
	out := make(map[string]interface{}, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		out[iter.Key().String()] = iter.Value().Interface()
	}
	return out, true
}

func listValue(value interface{}) []interface{} {
	if value == nil {
		return nil
	}
	if items, ok := value.([]interface{}); ok {
		return items
	}
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		items := make([]interface{}, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			items = append(items, rv.Index(i).Interface())
		}
		return items
	default:
		return nil
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
