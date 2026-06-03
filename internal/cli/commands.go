package cli

import (
	"sort"
	"todoz/internal/events"
	"todoz/internal/ids"
	"todoz/internal/model"
	"todoz/internal/response"
	"todoz/internal/store"
)

type Ctx struct {
	Store *store.Store
	Today string
	Now   string
}

type Handler func(ctx Ctx, flags map[string]string) response.Envelope

// cmdAddList creates a new list. Required flag: --name.
func cmdAddList(ctx Ctx, flags map[string]string) response.Envelope {
	name := flags["name"]
	if name == "" {
		return response.Error("invalid_operation", "add-list requires --name")
	}
	id := ids.New("list")
	err := ctx.Store.Append(events.Event{
		Type: events.TypeListCreated, At: ctx.Now, ListID: id, ListName: name,
	})
	if err != nil {
		return response.Error("io_error", err.Error())
	}
	return response.Success(map[string]string{"id": id, "name": name})
}

// cmdAddTask creates a task. Required: --title, --date (YYYY-MM-DD), --list. Optional: --description.
func cmdAddTask(ctx Ctx, flags map[string]string) response.Envelope {
	title := flags["title"]
	date := flags["date"]
	listID := flags["list"]
	if title == "" || date == "" || listID == "" {
		return response.Error("invalid_operation", "add-task requires --title, --date, and --list")
	}
	if !ValidDate(date) {
		return response.Error("invalid_date", "Date format must be YYYY-MM-DD, got: "+date)
	}
	st, err := ctx.Store.Load()
	if err != nil {
		return response.Error("io_error", err.Error())
	}
	if l, ok := st.Lists[listID]; !ok || l.IsDeleted {
		return response.Error("list_not_found", "List "+listID+" does not exist")
	}
	id := ids.New("task")
	err = ctx.Store.Append(events.Event{
		Type: events.TypeTaskCreated, At: ctx.Now,
		TaskID: id, ListID: listID, Title: title, Description: flags["description"], Date: date,
	})
	if err != nil {
		return response.Error("io_error", err.Error())
	}
	return response.Success(map[string]string{"id": id, "title": title, "listId": listID, "date": date})
}

// cmdUpdateTask updates a task's title, description, and/or date. Positional: task ID. At least one --flag required.
func cmdUpdateTask(ctx Ctx, flags map[string]string) response.Envelope {
	id := flags["_"]
	if id == "" {
		return response.Error("invalid_operation", "update-task requires a task id")
	}
	st, err := ctx.Store.Load()
	if err != nil {
		return response.Error("io_error", err.Error())
	}
	if _, ok := st.Tasks[id]; !ok {
		return response.Error("task_not_found", "Task "+id+" does not exist")
	}
	updates := map[string]string{}
	if v, has := flags["title"]; has {
		updates["title"] = v
	}
	if v, has := flags["description"]; has {
		updates["description"] = v
	}
	if v, has := flags["date"]; has {
		if !ValidDate(v) {
			return response.Error("invalid_date", "Date format must be YYYY-MM-DD, got: "+v)
		}
		updates["date"] = v
	}
	if len(updates) == 0 {
		return response.Error("invalid_operation", "update-task requires at least one of --title, --description, --date")
	}
	if err := ctx.Store.Append(events.Event{Type: events.TypeTaskUpdated, At: ctx.Now, TaskID: id, Updates: updates}); err != nil {
		return response.Error("io_error", err.Error())
	}
	return response.Success(map[string]string{"id": id})
}

// cmdCompleteTask marks a task completed. Positional: task ID.
func cmdCompleteTask(ctx Ctx, flags map[string]string) response.Envelope {
	id := flags["_"]
	if id == "" {
		return response.Error("invalid_operation", "complete-task requires a task id")
	}
	st, err := ctx.Store.Load()
	if err != nil {
		return response.Error("io_error", err.Error())
	}
	if _, ok := st.Tasks[id]; !ok {
		return response.Error("task_not_found", "Task "+id+" does not exist")
	}
	if err := ctx.Store.Append(events.Event{Type: events.TypeTaskCompleted, At: ctx.Now, TaskID: id}); err != nil {
		return response.Error("io_error", err.Error())
	}
	return response.Success(map[string]string{"id": id, "status": "completed"})
}

// cmdDeleteTask soft-deletes a task (to trash). With --permanently it hides the task from apps forever.
func cmdDeleteTask(ctx Ctx, flags map[string]string) response.Envelope {
	id := flags["_"]
	if id == "" {
		return response.Error("invalid_operation", "delete-task requires a task id")
	}
	st, err := ctx.Store.Load()
	if err != nil {
		return response.Error("io_error", err.Error())
	}
	if _, ok := st.Tasks[id]; !ok {
		return response.Error("task_not_found", "Task "+id+" does not exist")
	}
	evType := events.TypeTaskDeleted
	if flags["permanently"] == "true" {
		evType = events.TypeTaskPermanentlyDeleted
	}
	if err := ctx.Store.Append(events.Event{Type: evType, At: ctx.Now, TaskID: id}); err != nil {
		return response.Error("io_error", err.Error())
	}
	return response.Success(map[string]string{"id": id})
}

// cmdRestoreTask brings a soft-deleted task back out of the trash.
func cmdRestoreTask(ctx Ctx, flags map[string]string) response.Envelope {
	id := flags["_"]
	if id == "" {
		return response.Error("invalid_operation", "restore-task requires a task id")
	}
	st, err := ctx.Store.Load()
	if err != nil {
		return response.Error("io_error", err.Error())
	}
	if _, ok := st.Tasks[id]; !ok {
		return response.Error("task_not_found", "Task "+id+" does not exist")
	}
	if err := ctx.Store.Append(events.Event{Type: events.TypeTaskRestored, At: ctx.Now, TaskID: id}); err != nil {
		return response.Error("io_error", err.Error())
	}
	return response.Success(map[string]string{"id": id})
}

// cmdUpdateList renames a list. Required: positional list id, --name.
func cmdUpdateList(ctx Ctx, flags map[string]string) response.Envelope {
	id := flags["_"]
	name := flags["name"]
	if id == "" || name == "" {
		return response.Error("invalid_operation", "update-list requires a list id and --name")
	}
	st, err := ctx.Store.Load()
	if err != nil {
		return response.Error("io_error", err.Error())
	}
	if _, ok := st.Lists[id]; !ok {
		return response.Error("list_not_found", "List "+id+" does not exist")
	}
	if err := ctx.Store.Append(events.Event{Type: events.TypeListUpdated, At: ctx.Now, ListID: id, Updates: map[string]string{"name": name}}); err != nil {
		return response.Error("io_error", err.Error())
	}
	return response.Success(map[string]string{"id": id, "name": name})
}

// cmdDeleteList soft-deletes a list and cascades to all its tasks.
func cmdDeleteList(ctx Ctx, flags map[string]string) response.Envelope {
	id := flags["_"]
	if id == "" {
		return response.Error("invalid_operation", "delete-list requires a list id")
	}
	st, err := ctx.Store.Load()
	if err != nil {
		return response.Error("io_error", err.Error())
	}
	if _, ok := st.Lists[id]; !ok {
		return response.Error("list_not_found", "List "+id+" does not exist")
	}
	if err := ctx.Store.Append(events.Event{Type: events.TypeListDeleted, At: ctx.Now, ListID: id}); err != nil {
		return response.Error("io_error", err.Error())
	}
	return response.Success(map[string]string{"id": id})
}

// LoadView is the read model returned by the load command: active lists, active tasks (with IsOverdue computed), and the current trash.
type LoadView struct {
	Lists []model.List `json:"lists"`
	Tasks []model.Task `json:"tasks"`
	Trash []model.Task `json:"trash"`
}

// GroupedLoadView is returned by cmdLoad when --group-by is set.
type GroupedLoadView struct {
	Groups  map[string][]model.Task `json:"groups"`
	Lists   []model.List            `json:"lists"`
	Summary *QuerySummary           `json:"summary,omitempty"`
}

// cmdLoad returns filtered, sorted, optionally grouped tasks.
func cmdLoad(ctx Ctx, flags map[string]string) response.Envelope {
	if err := ValidateQueryFlags(flags); err != nil {
		return response.Error("invalid_operation", err.Error())
	}
	opts, err := ParseQueryOptions(flags)
	if err != nil {
		return response.Error("invalid_operation", err.Error())
	}
	st, err := ctx.Store.Load()
	if err != nil {
		return response.Error("io_error", err.Error())
	}

	var sourceTasks []model.Task
	if opts.TrashOnly {
		sourceTasks = st.TrashTasks()
	} else {
		sourceTasks = st.ActiveTasks()
	}

	for i := range sourceTasks {
		sourceTasks[i].IsOverdue = model.ComputeOverdue(sourceTasks[i], ctx.Today)
	}

	activeLists := st.ActiveLists()
	sort.Slice(activeLists, func(i, j int) bool { return activeLists[i].CreatedAt < activeLists[j].CreatedAt })

	result := ApplyQuery(sourceTasks, activeLists, opts, ctx.Today)

	if opts.GroupBy != "" {
		view := GroupedLoadView{
			Groups:  result.Groups,
			Lists:   activeLists,
			Summary: result.Summary,
		}
		return response.Success(view)
	}

	trash := []model.Task{}
	if !opts.TrashOnly && !opts.NoTrash {
		trash = st.TrashTasks()
	}
	view := LoadView{
		Lists: activeLists,
		Tasks: result.Tasks,
		Trash: trash,
	}
	if result.Summary != nil {
		type loadWithSummary struct {
			LoadView
			Summary *QuerySummary `json:"summary"`
		}
		return response.Success(loadWithSummary{view, result.Summary})
	}
	return response.Success(view)
}

// cmdCompact rewrites the event log to its minimal form (see store.Compact).
func cmdCompact(ctx Ctx, _ map[string]string) response.Envelope {
	if err := ctx.Store.Compact(); err != nil {
		return response.Error("io_error", err.Error())
	}
	return response.Success(map[string]string{"status": "compacted"})
}
