package commands

import (
	apptask "hyperbdr-client/internal/app/task"
	"hyperbdr-client/internal/output"
)

func runTasks(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("tasks", "")
	}
	service := apptask.NewService(commandAPIAdapter{ctx: ctx})
	switch args[0] {
	case "list":
		fs := newFlagSet("tasks list")
		sourceID := fs.String("source-id", "", "")
		page := fs.Int("page", 1, "")
		pageSize := fs.Int("page-size", 100, "")
		q := queryFromPairs()
		if err := parseQueryFlagsInto(fs, args[1:], q); err != nil {
			return err
		}
		resp, err := service.List(apptask.ListSpec{
			SourceID: *sourceID,
			Page:     *page,
			PageSize: *pageSize,
			Query:    q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "tasks", taskColumns())
	case "steps":
		fs := newFlagSet("tasks steps")
		taskID := fs.String("task-id", "", "")
		q := queryFromPairs()
		if err := parseQueryFlagsInto(fs, args[1:], q); err != nil {
			return err
		}
		if *taskID == "" {
			return missing(ctx, "error.missing_task_id")
		}
		resp, err := service.Steps(apptask.StepsSpec{
			TaskID: *taskID,
			Query:  q,
		})
		if err != nil {
			return err
		}
		return writeResponse(ctx, resp, "steps", stepColumns())
	default:
		return errUnknown("tasks", args[0])
	}
}

func taskColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.display_type", Field: "display_type"},
		{HeaderKey: "table.source", Field: "source"},
		{HeaderKey: "table.source_id", Field: "source_id"},
		{HeaderKey: "table.start_at", Field: "start_at"},
		{HeaderKey: "table.end_at", Field: "end_at"},
		{HeaderKey: "table.display_status", Field: "display_status"},
	}
}

func stepColumns() []output.Column {
	return []output.Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.step", Field: "step"},
		{HeaderKey: "table.status", Field: "status"},
		{HeaderKey: "table.message", Field: "message"},
	}
}
