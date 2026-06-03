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
