package cli

import (
	"strings"
	"time"
	"todoz/internal/clock"
	"todoz/internal/config"
	"todoz/internal/oplog"
	"todoz/internal/response"
	"todoz/internal/store"
)

var registry = map[string]Handler{
	"load":          cmdLoad,
	"add-task":      cmdAddTask,
	"update-task":   cmdUpdateTask,
	"complete-task": cmdCompleteTask,
	"delete-task":   cmdDeleteTask,
	"restore-task":  cmdRestoreTask,
	"add-list":      cmdAddList,
	"update-list":   cmdUpdateList,
	"delete-list":   cmdDeleteList,
	"compact":       cmdCompact,
}

func Run(args []string) (string, int) {
	start := clock.Now()
	cfg, err := config.Load()
	if err != nil {
		return response.Error("config_error", err.Error()).JSON(), 1
	}
	if len(args) == 0 {
		return finish(cfg, start, "todoz", response.Error("unknown_command", "no command given"))
	}
	name := args[0]
	handler, ok := registry[name]
	if !ok {
		return finish(cfg, start, "todoz "+name, response.Error("unknown_command", "unknown command: "+name))
	}
	flags := ParseFlags(args[1:])
	ctx := Ctx{
		Store: store.New(cfg),
		Today: clock.Now().Format("2006-01-02"),
		Now:   clock.Format(clock.Now()),
	}
	res := handler(ctx, flags)
	return finish(cfg, start, "todoz "+strings.Join(args, " "), res)
}

func finish(cfg config.Config, start time.Time, request string, res response.Envelope) (string, int) {
	entry := oplog.Entry{
		Timestamp:  clock.Format(clock.Now()),
		Request:    request,
		OK:         res.OK,
		DurationMS: time.Since(start).Milliseconds(),
	}
	if res.OK {
		entry.Data = res.JSON()
	} else {
		entry.Error = res.ErrCode
		entry.Message = res.Message
	}
	_ = oplog.Record(cfg.AppLogPath(), entry)
	code := 0
	if !res.OK {
		code = 1
	}
	return res.JSON(), code
}
