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
