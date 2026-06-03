package cli

import (
	"todoz/internal/events"
	"todoz/internal/ids"
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
