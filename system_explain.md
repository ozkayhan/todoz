# Todoz — System Architecture (Technical)

## 1. Purpose & Constraints

Standalone Go binary, no daemon, offline-only, language-agnostic JSON-over-subprocess interface.
Single user, thousands of tasks, dozens of lists, multi-process safe.

## 2. High-Level Component Map

```
                     +--------------------+
 consuming app  ─────► |  cmd/todoz (main)  |
 (any language)        +---------+----------+
                               │ os.Args
                               ▼
                     +--------------------+
                     |   internal/cli     |  router → flags → command handler
                     +---------+----------+
                      │        │         │
          ┌───────────┘        │         └────────────┐
          ▼                    ▼                      ▼
  +-------------+       +-------------+        +-------------+
  |  response   |       |   store     |        |   oplog     |
  | (envelope)  |       | (log+lock)  |        | (audit log) |
  +-------------+       +------+------+        +-------------+
                               │
                 ┌─────────────┼──────────────┐
                 ▼             ▼              ▼
           +---------+   +---------+    +---------+
           | events  |   |  state  |    |  model  |
           | (reduce)|   |(project)|    | (data)  |
           +---------+   +---------+    +---------+
```

## 3. Module Responsibility Table

| Package | Responsibility | I/O? | Depends on |
|---------|---------------|------|-----------|
| `config` | env vars → paths | reads env | – |
| `clock` | Istanbul timestamps | reads wall clock | – |
| `ids` | random IDs | crypto/rand | – |
| `model` | Task/List/Status, overdue | none | – |
| `events` | event vocabulary + reducer | none | model, state |
| `state` | in-memory projection | none | model |
| `store` | log read/append, lock, compact | disk | config, events, state, model |
| `oplog` | audit log blocks | disk | – |
| `response` | JSON envelope | none | – |
| `cli` | parse, dispatch, query pipeline, record | via store/oplog | all of the above |
| `cmd/todoz` | process entry | stdout/exit | cli |

## 4. Data Model

**Status:** Only `pending` or `completed`. Overdue is NOT a status—it's computed at read time.

**Task:**
- ID: "task-<32 hex>"
- Title: required text
- Description: optional text
- ListID: owning list (required)
- Date: YYYY-MM-DD (required)
- Status: pending or completed
- IsOverdue: computed flag (pending AND Date < today)
- IsDeleted: soft-deleted (in trash, visible)
- IsHiddenTrash: permanently hidden
- CreatedAt, UpdatedAt: ISO8601 timestamps
- CompletedAt, DeletedAt: set when status changes

**List:**
- ID: "list-<32 hex>"
- Name: required human-readable name
- CreatedAt: ISO8601 timestamp
- IsDeleted: soft-deleted

## 5. Event Log Design

Append-only `events.jsonl` (one event per line, JSON).

**Event Types:**
- `list_created`: Creates a list
- `list_updated`: Updates list name
- `list_deleted`: Soft-deletes list and cascades to all its tasks
- `task_created`: Creates a task (sets status=pending)
- `task_updated`: Patches task title/description/date (uses Updates map)
- `task_completed`: Transitions status to completed
- `task_deleted`: Soft-deletes task
- `task_permanently_deleted`: Hides task (IsHiddenTrash=true)
- `task_restored`: Undoes soft-delete

Each event carries a timestamp (`At`) and the minimal fields needed to represent the mutation.

## 6. Event Lifecycle (Sequence)

Example: `add-task --title "Buy milk" --date 2026-06-05 --list L`:
1. `cli.Run` parses args, builds Ctx (Today, Now timestamps)
2. `cmdAddTask` validates list exists, date format is valid
3. `Store.Append(task_created event)` acquires lock → AppendEvent → Release
4. `oplog.Record(request/response)` writes audit entry
5. Print JSON envelope, exit 0

## 7. Concurrency Model

Lock file via `O_CREATE|O_EXCL` (atomic exclusive creation). 5-second timeout. Each append is serialized. No daemon—full replay on every Load ensures no stale reads.

Two processes contending for the lock:
```
Process A:                          Process B:
Acquire(.lock) ✓ at T0
                                    Acquire(.lock) ✗ EEXIST, retry...
AppendEvent(...) at T0+1ms
Release(.lock) at T0+2ms
                                    Acquire(.lock) ✓ at T0+5ms
                                    AppendEvent(...) at T0+6ms
                                    Release(.lock) at T0+7ms
```

## 8. Storage Layout

```
$TODO_LIB_HOME/
├── events.jsonl           # Append-only log
├── events.jsonl.backup    # Pre-compaction backup
├── .lock                  # Exclusive lock file (created while appending)
└── logs/
    ├── app1.log           # Per-app operation audit log
    ├── app2.log
    └── ...
```

## 9. Timestamp Standard

RFC3339 with 6-digit fractional seconds (microseconds) and fixed +03:00 offset.
Format: `2026-06-03T10:30:45.123456+03:00`
Produced only by `clock` package, never constructed elsewhere.

## 10. Trash Semantics

State machine:
- **Active:** IsDeleted=false, IsHiddenTrash=false
- **Trash:** IsDeleted=true, IsHiddenTrash=false
- **Hidden:** IsDeleted=true, IsHiddenTrash=true (unseen by apps)

Transitions:
- Active → Trash: `delete-task` (first soft-delete)
- Trash → Active: `restore-task` (undo soft-delete)
- Trash → Hidden: `delete-task --permanently` (permanent hide)
- List deletion cascades: all tasks in the deleted list move to Trash

## 11. Compaction

**When/why:** To shrink the log after many mutations. Safe operation under lock.

**Algorithm:** Load full state (replay all events), synthesize minimal event set that recreates it (one created event per live item + status/deletion transitions), write to temp file, atomically rename.

**Atomicity:** Temp rename ensures crash mid-way never corrupts the live log. Backup retained.

## 12. Error Code Catalogue

| Code | Meaning |
|------|---------|
| `invalid_operation` | Missing or conflicting flags |
| `invalid_date` | Date not YYYY-MM-DD format |
| `task_not_found` | Unknown task ID |
| `list_not_found` | Unknown or deleted list ID |
| `unknown_command` | Unrecognized command |
| `io_error` | Disk read/write/lock failure |
| `config_error` | Bad environment/config |

## 13. Extension Guide

**To add a new command:**
1. Implement a handler `func cmdXyz(ctx Ctx, flags map[string]string) response.Envelope` in `internal/cli/commands.go`
2. Add `"xyz": cmdXyz` to the registry in `internal/cli/router.go`

**To add a new event type:**
1. Define `const TypeXyz = "xyz"` in `internal/events/event.go`
2. Add a case in `Apply()` in `internal/events/apply.go`
3. Add a case in `snapshotEvents()` in `internal/store/compact.go` if the event represents state

This architecture keeps concerns separated (model, events, state, store, oplog, cli) and makes each layer independently testable.

## 14. Query Pipeline (Sprint2)

The `load` command supports rich filtering, sorting, grouping, and aggregation via the query layer in `internal/cli`:

**QueryOptions:** Parsed flag representation of all query parameters (dates, status, lists, search text, sort order, grouping strategy, summary request).

**ValidateQueryFlags:** Pre-parse validation in `query_validate.go` detects conflicting flags early (e.g., `--days-back` + `--after-date`, or `--no-trash` + `--trash-only`) and validates enum values (status, sort-by, group-by, output-format).

**ApplyQuery:** Pure stateless pipeline in `query.go` that:
1. Filters tasks by date range (AfterDate/BeforeDate/DaysBack), status, overdue, list membership, and search text
2. Sorts by the requested field (date, title, created, status) with optional reverse
3. Groups by list name, date, or status (if `--group-by` specified), else returns flat array
4. Computes QuerySummary (total, pending, completed, overdue counts) if requested

**Response shape:** Normally `{"lists": [...], "tasks": [...], "trash": [...]}`. With `--group-by`, becomes `{"groups": {"key1": [...], ...}, "lists": [...], "summary": {...}}` (summary only if `--summary` or `--count` flag).

All filtering is case-insensitive for search. Dates are YYYY-MM-DD strings, compared lexicographically (safe for dates). Overdue computation uses the Today timestamp from context.
