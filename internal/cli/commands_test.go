package cli

import (
	"testing"
	"todoz/internal/config"
	"todoz/internal/store"
)

func testCtx(t *testing.T) Ctx {
	t.Helper()
	t.Setenv("TODO_LIB_HOME", t.TempDir())
	t.Setenv("TODO_APP_NAME", "test")
	c, err := config.Load()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	return Ctx{Store: store.New(c), Today: "2026-06-03", Now: "2026-06-03T10:00:00.000000+03:00"}
}

func TestAddListSuccess(t *testing.T) {
	ctx := testCtx(t)
	res := cmdAddList(ctx, ParseFlags([]string{"--name", "Groceries"}))
	if !res.OK {
		t.Fatalf("expected ok, got %+v", res)
	}
	st, _ := ctx.Store.Load()
	if len(st.ActiveLists()) != 1 {
		t.Fatalf("list not persisted: %+v", st.Lists)
	}
}

func TestAddListRequiresName(t *testing.T) {
	ctx := testCtx(t)
	res := cmdAddList(ctx, ParseFlags([]string{}))
	if res.OK || res.ErrCode != "invalid_operation" {
		t.Fatalf("expected invalid_operation, got %+v", res)
	}
}

func seedList(t *testing.T, ctx Ctx) string {
	t.Helper()
	res := cmdAddList(ctx, ParseFlags([]string{"--name", "L"}))
	return res.Data.(map[string]string)["id"]
}

func TestAddTaskSuccess(t *testing.T) {
	ctx := testCtx(t)
	listID := seedList(t, ctx)
	res := cmdAddTask(ctx, ParseFlags([]string{"--title", "Buy milk", "--date", "2026-06-05", "--list", listID}))
	if !res.OK {
		t.Fatalf("expected ok, got %+v", res)
	}
}

func TestAddTaskRejectsBadDate(t *testing.T) {
	ctx := testCtx(t)
	listID := seedList(t, ctx)
	res := cmdAddTask(ctx, ParseFlags([]string{"--title", "x", "--date", "06-05-2026", "--list", listID}))
	if res.OK || res.ErrCode != "invalid_date" {
		t.Fatalf("expected invalid_date, got %+v", res)
	}
}

func TestAddTaskRejectsMissingList(t *testing.T) {
	ctx := testCtx(t)
	res := cmdAddTask(ctx, ParseFlags([]string{"--title", "x", "--date", "2026-06-05", "--list", "nope"}))
	if res.OK || res.ErrCode != "list_not_found" {
		t.Fatalf("expected list_not_found, got %+v", res)
	}
}

func TestCompleteTask(t *testing.T) {
	ctx := testCtx(t)
	listID := seedList(t, ctx)
	add := cmdAddTask(ctx, ParseFlags([]string{"--title", "x", "--date", "2026-06-05", "--list", listID}))
	taskID := add.Data.(map[string]string)["id"]
	res := cmdCompleteTask(ctx, ParseFlags([]string{taskID}))
	if !res.OK {
		t.Fatalf("complete failed: %+v", res)
	}
	st, _ := ctx.Store.Load()
	if st.Tasks[taskID].Status != "completed" {
		t.Fatalf("task not completed: %+v", st.Tasks[taskID])
	}
}

func TestUpdateTask(t *testing.T) {
	ctx := testCtx(t)
	listID := seedList(t, ctx)
	add := cmdAddTask(ctx, ParseFlags([]string{"--title", "Old", "--date", "2026-06-05", "--list", listID}))
	taskID := add.Data.(map[string]string)["id"]
	res := cmdUpdateTask(ctx, ParseFlags([]string{taskID, "--title", "New"}))
	if !res.OK {
		t.Fatalf("update failed: %+v", res)
	}
	st, _ := ctx.Store.Load()
	if st.Tasks[taskID].Title != "New" {
		t.Fatalf("title not updated: %+v", st.Tasks[taskID])
	}
}

func seedTask(t *testing.T, ctx Ctx, listID string) string {
	t.Helper()
	add := cmdAddTask(ctx, ParseFlags([]string{"--title", "x", "--date", "2026-06-05", "--list", listID}))
	return add.Data.(map[string]string)["id"]
}

func TestDeleteTaskSoftThenPermanent(t *testing.T) {
	ctx := testCtx(t)
	listID := seedList(t, ctx)
	taskID := seedTask(t, ctx, listID)
	if res := cmdDeleteTask(ctx, ParseFlags([]string{taskID})); !res.OK {
		t.Fatalf("soft delete failed: %+v", res)
	}
	st, _ := ctx.Store.Load()
	if !st.Tasks[taskID].IsDeleted || st.Tasks[taskID].IsHiddenTrash {
		t.Fatalf("want soft-deleted, got %+v", st.Tasks[taskID])
	}
	if res := cmdDeleteTask(ctx, ParseFlags([]string{taskID, "--permanently"})); !res.OK {
		t.Fatalf("permanent delete failed: %+v", res)
	}
	st, _ = ctx.Store.Load()
	if !st.Tasks[taskID].IsHiddenTrash {
		t.Fatalf("want hidden, got %+v", st.Tasks[taskID])
	}
}

func TestRestoreTask(t *testing.T) {
	ctx := testCtx(t)
	listID := seedList(t, ctx)
	taskID := seedTask(t, ctx, listID)
	_ = cmdDeleteTask(ctx, ParseFlags([]string{taskID}))
	if res := cmdRestoreTask(ctx, ParseFlags([]string{taskID})); !res.OK {
		t.Fatalf("restore failed: %+v", res)
	}
	st, _ := ctx.Store.Load()
	if st.Tasks[taskID].IsDeleted {
		t.Fatalf("want restored, got %+v", st.Tasks[taskID])
	}
}

func TestDeleteListCascades(t *testing.T) {
	ctx := testCtx(t)
	listID := seedList(t, ctx)
	taskID := seedTask(t, ctx, listID)
	if res := cmdDeleteList(ctx, ParseFlags([]string{listID})); !res.OK {
		t.Fatalf("delete-list failed: %+v", res)
	}
	st, _ := ctx.Store.Load()
	if !st.Lists[listID].IsDeleted || !st.Tasks[taskID].IsDeleted {
		t.Fatalf("cascade failed: list=%+v task=%+v", st.Lists[listID], st.Tasks[taskID])
	}
}

func TestUpdateList(t *testing.T) {
	ctx := testCtx(t)
	listID := seedList(t, ctx)
	if res := cmdUpdateList(ctx, ParseFlags([]string{listID, "--name", "Renamed"})); !res.OK {
		t.Fatalf("update-list failed: %+v", res)
	}
	st, _ := ctx.Store.Load()
	if st.Lists[listID].Name != "Renamed" {
		t.Fatalf("rename failed: %+v", st.Lists[listID])
	}
}

func TestLoadReturnsViewWithOverdue(t *testing.T) {
	ctx := testCtx(t)
	listID := seedList(t, ctx)
	_ = cmdAddTask(ctx, ParseFlags([]string{"--title", "late", "--date", "2026-06-01", "--list", listID}))
	_ = cmdAddTask(ctx, ParseFlags([]string{"--title", "soon", "--date", "2026-06-10", "--list", listID}))
	res := cmdLoad(ctx, ParseFlags([]string{}))
	if !res.OK {
		t.Fatalf("load failed: %+v", res)
	}
	view := res.Data.(LoadView)
	if len(view.Lists) != 1 || len(view.Tasks) != 2 {
		t.Fatalf("unexpected view counts: %+v", view)
	}
	var overdueCount int
	for _, tk := range view.Tasks {
		if tk.IsOverdue {
			overdueCount++
		}
	}
	if overdueCount != 1 {
		t.Fatalf("want 1 overdue, got %d", overdueCount)
	}
}

func TestLoadFilterByList(t *testing.T) {
	ctx := testCtx(t)
	a := seedList(t, ctx)
	b := seedList(t, ctx)
	_ = seedTask(t, ctx, a)
	_ = seedTask(t, ctx, b)
	res := cmdLoad(ctx, ParseFlags([]string{"--list", a}))
	view := res.Data.(LoadView)
	if len(view.Tasks) != 1 || view.Tasks[0].ListID != a {
		t.Fatalf("filter failed: %+v", view.Tasks)
	}
}

func TestCompactCommand(t *testing.T) {
	ctx := testCtx(t)
	listID := seedList(t, ctx)
	_ = seedTask(t, ctx, listID)
	res := cmdCompact(ctx, ParseFlags([]string{}))
	if !res.OK {
		t.Fatalf("compact failed: %+v", res)
	}
	st, _ := ctx.Store.Load()
	if len(st.ActiveLists()) != 1 || len(st.ActiveTasks()) != 1 {
		t.Fatalf("state lost after compact: %+v", st)
	}
}

func seedTaskWithDate(t *testing.T, ctx Ctx, listID, title, date string) string {
	t.Helper()
	res := cmdAddTask(ctx, ParseFlags([]string{"--title", title, "--date", date, "--list", listID}))
	if !res.OK {
		t.Fatalf("seedTask failed: %+v", res)
	}
	return res.Data.(map[string]string)["id"]
}

func TestCmdLoad_DaysBack(t *testing.T) {
	ctx := testCtx(t)
	ctx.Today = "2026-06-03"
	listID := seedList(t, ctx)
	seedTaskWithDate(t, ctx, listID, "Old", "2026-05-20")
	seedTaskWithDate(t, ctx, listID, "Recent", "2026-06-01")
	res := cmdLoad(ctx, ParseFlags([]string{"--days-back", "7"}))
	if !res.OK {
		t.Fatalf("load failed: %+v", res)
	}
	view := res.Data.(LoadView)
	if len(view.Tasks) != 1 || view.Tasks[0].Title != "Recent" {
		t.Fatalf("want [Recent], got %+v", view.Tasks)
	}
}

func TestCmdLoad_StatusFilter(t *testing.T) {
	ctx := testCtx(t)
	ctx.Today = "2026-06-03"
	listID := seedList(t, ctx)
	taskID := seedTaskWithDate(t, ctx, listID, "T1", "2026-06-05")
	seedTaskWithDate(t, ctx, listID, "T2", "2026-06-06")
	cmdCompleteTask(ctx, ParseFlags([]string{taskID}))
	res := cmdLoad(ctx, ParseFlags([]string{"--status", "pending"}))
	view := res.Data.(LoadView)
	if len(view.Tasks) != 1 || view.Tasks[0].Title != "T2" {
		t.Fatalf("want [T2], got %+v", view.Tasks)
	}
}

func TestCmdLoad_SortByTitleReverse(t *testing.T) {
	ctx := testCtx(t)
	ctx.Today = "2026-06-03"
	listID := seedList(t, ctx)
	seedTaskWithDate(t, ctx, listID, "Bravo", "2026-06-01")
	seedTaskWithDate(t, ctx, listID, "Alpha", "2026-06-02")
	seedTaskWithDate(t, ctx, listID, "Charlie", "2026-06-03")
	res := cmdLoad(ctx, ParseFlags([]string{"--sort-by", "title", "--sort-reverse"}))
	view := res.Data.(LoadView)
	if view.Tasks[0].Title != "Charlie" {
		t.Fatalf("want Charlie first, got %s", view.Tasks[0].Title)
	}
}

func TestCmdLoad_ConflictError(t *testing.T) {
	ctx := testCtx(t)
	res := cmdLoad(ctx, ParseFlags([]string{"--days-back", "7", "--after-date", "2026-01-01"}))
	if res.OK || res.ErrCode != "invalid_operation" {
		t.Fatalf("want invalid_operation, got %+v", res)
	}
}

func TestCmdLoad_GroupBy(t *testing.T) {
	ctx := testCtx(t)
	ctx.Today = "2026-06-03"
	listID := seedList(t, ctx)
	seedTaskWithDate(t, ctx, listID, "T1", "2026-06-01")
	seedTaskWithDate(t, ctx, listID, "T2", "2026-06-02")
	res := cmdLoad(ctx, ParseFlags([]string{"--group-by", "date"}))
	if !res.OK {
		t.Fatalf("load failed: %+v", res)
	}
	view := res.Data.(GroupedLoadView)
	if len(view.Groups) != 2 {
		t.Fatalf("want 2 date groups, got %d: %+v", len(view.Groups), view.Groups)
	}
}
