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
