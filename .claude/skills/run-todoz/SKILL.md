---
name: run-todoz
description: Build and test the todoz CLI — append-only todo store
---

# Run Todoz CLI

Todoz is a standalone Go CLI binary that manages todo lists via an append-only event log. JSON-over-subprocess interface. Multi-process safe.

Run via agent path: build, then run the smoke test script below. It exercises all major commands (add-list, add-task, complete, delete, restore, load with overdue computation, compact).

## Prerequisites

- Go 1.22+
- `jq` (for JSON parsing in smoke test)

```bash
apt-get update && apt-get install -y golang-go jq
```

## Build

```bash
go build -o todoz ./cmd/todoz
```

Produces `./todoz` binary in the repo root.

## Run (Agent Path)

Launch the smoke test script to exercise the full command set:

```bash
bash ./.claude/skills/run-todoz/smoke.sh
```

Script location: `./.claude/skills/run-todoz/smoke.sh` (relative to repo root).

The script:
- Builds a temporary data store (`/tmp/todoz-smoke-<timestamp>`)
- Exercises 10 command flows: add-list, add-task (normal and overdue), load, complete, delete, restore, update, compact
- Verifies JSON response structure and computed fields (overdue)
- Reports OK or FAIL for each flow

**Environment variables:**
- `TODO_LIB_HOME` — data directory (default: `/tmp/todoz-smoke-<timestamp>`)
- `TODO_APP_NAME` — app identifier for audit log (default: `smoke-test`)
- `TODOZ` — path to binary (default: `./todoz`)

Example with custom home:

```bash
TODO_LIB_HOME=~/.local/share/todoz bash ./.claude/skills/run-todoz/smoke.sh
```

## Run (Human Path)

Single command to add a list and view it:

```bash
export TODO_LIB_HOME=~/.local/share/todoz TODO_APP_NAME=cli
./todoz add-list --name "My Tasks"
./todoz load
```

Each invocation is independent — no daemon. Full state replay from the event log on every call.

## Test

Run the full Go test suite:

```bash
go test ./...
```

Expected: 53 tests pass. Test files:
- `internal/clock/clock_test.go` — Istanbul timestamps
- `internal/ids/ids_test.go` — unique ID generation
- `internal/config/config_test.go` — env var resolution
- `internal/model/status_test.go` — status enum
- `internal/model/task_test.go` — JSON marshaling
- `internal/model/overdue_test.go` — overdue computation
- `internal/events/event_test.go`, `codec_test.go`, `apply_test.go` — event log + reducer
- `internal/state/state_test.go` — in-memory state projection
- `internal/store/lock_test.go`, `log_test.go`, `store_test.go`, `compact_test.go` — persistence + locking + compaction
- `internal/oplog/oplog_test.go` — audit log
- `internal/response/response_test.go` — JSON envelope
- `internal/cli/flags_test.go`, `commands_test.go`, `router_test.go` — CLI parsing + commands
- `cmd/todoz/main_test.go` — end-to-end binary test
- `test/concurrency_test.go` — 50 concurrent appends (no data loss)

## Commands

| Command | Args | Effect |
|---------|------|--------|
| `add-list` | `--name <text>` | Create a list |
| `load` | `[--list <id>]` | Load active lists, tasks, trash; optionally filter by list |
| `add-task` | `--title <text> --date <YYYY-MM-DD> --list <id> [--description <text>]` | Create a task (pending status) |
| `update-task` | `<taskId> [--title] [--description] [--date]` | Patch task fields (at least one required) |
| `complete-task` | `<taskId>` | Mark task completed |
| `delete-task` | `<taskId> [--permanently]` | Soft-delete (trash). With `--permanently`, hide permanently. |
| `restore-task` | `<taskId>` | Restore soft-deleted task from trash |
| `update-list` | `<listId> --name <text>` | Rename a list |
| `delete-list` | `<listId>` | Soft-delete a list and cascade to all its tasks |
| `compact` | (none) | Rewrite event log to minimal form (safe operation) |

Every command returns a JSON envelope: `{"ok": true, "data": {...}}` or `{"ok": false, "error": "<code>", "message": "<text>"}`.

Exit code: 0 on success, 1 on failure.

## Gotchas

1. **Timestamp precision:** All timestamps use RFC3339 with microseconds and fixed +03:00 (Istanbul, no DST). Dates are YYYY-MM-DD bare strings. Times are never stored — only dates.

2. **Soft-delete is two-step:** First `delete-task <id>` soft-deletes to trash (visible in `load` under `trash`). Second `delete-task <id> --permanently` hides it entirely. Nothing is ever physically removed from the event log.

3. **Overdue is computed, not stored:** The `isOverdue` field in a task is computed at load time from the task's date and the current date (via `clock.Now()`). It is NOT a status. A pending task with a past date shows `"isOverdue": true` but `"status": "pending"`.

4. **List deletion cascades:** `delete-list <id>` soft-deletes the list AND all tasks in that list. They move to trash together.

5. **Full replay on every call:** `load` replays the entire event log from disk into state. For very large logs (>100k events), this is slow. Use `compact` to shrink the log.

6. **Lock timeout is 5 seconds:** If two processes contend heavily for the lock, concurrent appends may time out. The default is generous for single-machine use.

7. **No daemon:** Todoz is a stateless subprocess. Every call is independent. No server to start, no Ctrl-C needed.

## Troubleshooting

| Error | Fix |
|-------|-----|
| `config_error: cannot read random bytes` | Go crypto/rand entropy missing. Run on a system with `/dev/urandom`. |
| `invalid_date: Date format must be YYYY-MM-DD` | Date string malformed. Use exactly `YYYY-MM-DD` (e.g., `2026-06-05`). |
| `list_not_found: List <id> does not exist` | List ID wrong or list was deleted. Use `load` to see current lists. |
| `task_not_found: Task <id> does not exist` | Task ID wrong or task was deleted. |
| `io_error` | Disk read/write failed. Check `$TODO_LIB_HOME` exists and is writable. |

## Implementation Notes

Source: `/cmd/todoz/` (entry point) → `/internal/` (packages):
- `config/` — environment variable resolution
- `clock/` — Istanbul timestamp formatting
- `ids/` — random ID generation
- `model/` — Task, List, Status types + ComputeOverdue
- `events/` — event vocabulary + reducer
- `state/` — in-memory state projection
- `store/` — disk persistence + locking + compaction
- `oplog/` — audit log (request/response + timestamp + duration)
- `response/` — JSON envelope wrapper
- `cli/` — command parsing + routing

Architecture: each layer is independently testable. No third-party dependencies. Stdlib only.
