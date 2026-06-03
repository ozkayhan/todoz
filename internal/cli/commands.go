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

type LoadView struct {
	Lists []model.List `json:"lists"`
	Tasks []model.Task `json:"tasks"`
	Trash []model.Task `json:"trash"`
}

func cmdLoad(ctx Ctx, flags map[string]string) response.Envelope {
	st, err := ctx.Store.Load()
	if err != nil {
		return response.Error("io_error", err.Error())
	}
	filter := flags["list"]
	view := LoadView{Lists: []model.List{}, Tasks: []model.Task{}, Trash: []model.Task{}}
	for _, l := range st.ActiveLists() {
		if filter != "" && l.ID != filter {
			continue
		}
		view.Lists = append(view.Lists, l)
	}
	for _, tk := range st.ActiveTasks() {
		if filter != "" && tk.ListID != filter {
			continue
		}
		tk.IsOverdue = model.ComputeOverdue(tk, ctx.Today)
		view.Tasks = append(view.Tasks, tk)
	}
	for _, tk := range st.TrashTasks() {
		if filter != "" && tk.ListID != filter {
			continue
		}
		view.Trash = append(view.Trash, tk)
	}
	sort.Slice(view.Tasks, func(i, j int) bool {
		if view.Tasks[i].Date != view.Tasks[j].Date {
			return view.Tasks[i].Date < view.Tasks[j].Date
		}
		return view.Tasks[i].ID < view.Tasks[j].ID
	})
	sort.Slice(view.Lists, func(i, j int) bool { return view.Lists[i].CreatedAt < view.Lists[j].CreatedAt })
	return response.Success(view)
}

func cmdCompact(ctx Ctx, _ map[string]string) response.Envelope {
	if err := ctx.Store.Compact(); err != nil {
		return response.Error("io_error", err.Error())
	}
	return response.Success(map[string]string{"status": "compacted"})
}
