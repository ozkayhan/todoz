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
