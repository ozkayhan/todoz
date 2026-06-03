# Todoz Design Specification

**Project:** Todoz  
**Date:** 2026-06-03  
**Version:** 1.0  
**Status:** Approved

---

## Executive Summary

Todoz is a standalone, high-performance todo library written in Go, designed for integration into any application or stack. It provides offline-first task management with a CLI interface and JSON I/O, enabling seamless multi-app synchronization without conflicts.

**Key characteristics:**
- Append-only event log (zero conflicts)
- CLI binary + JSON interface (language-agnostic)
- Portable configuration (environment-driven)
- Comprehensive logging (audit trail)
- Soft-delete with hidden trash (nothing truly deleted)
- ~8ms load time for 5000 tasks

---

## 1. Data Model

### Task

```json
{
  "id": "uuid-or-number",
  "title": "string (required)",
  "description": "string (optional)",
  "listId": "string (required, belongs to exactly one list)",
  "date": "YYYY-MM-DD (required, no time component)",
  "status": "pending | completed | overdue",
  "createdAt": "ISO 8601 with Istanbul TZ (+03:00)",
  "updatedAt": "ISO 8601 timestamp"
}
```

**Status Logic:**
- `pending`: task not yet completed
- `completed`: task marked done by user

**Overdue Flag (computed, not stored):**
- `isOverdue: true` if status = `pending` AND date < today
- `isOverdue: false` otherwise
- Purely visual indicator for UI; task stays `pending` until explicitly marked `completed`

### List

```json
{
  "id": "string",
  "name": "string",
  "createdAt": "ISO 8601 with Istanbul TZ (+03:00)"
}
```

**Constraints:**
- Each task belongs to exactly one list
- Deleting a list deletes all its tasks (moved to trash, never truly deleted)

### Trash

Tasks in trash have additional fields:
```json
{
  "...task fields...",
  "isDeleted": true,
  "deletedAt": "ISO 8601 timestamp",
  "isHiddenTrash": false
}
```

`isHiddenTrash: true` means permanently deleted (hidden from all apps, but still in event log).

---

## 2. Interface: CLI Binary

Todoz is distributed as a single Go binary: `todoz`

### CLI Commands

```bash
# Load all data
todoz load [--list <listId>]

# Task operations
todoz add-task --title "..." --date 2026-06-05 --list <listId> [--description "..."]
todoz update-task <taskId> --title "..." [--date ...] [--description ...]
todoz complete-task <taskId>
todoz delete-task <taskId>                    # Soft delete (to trash)
todoz delete-task <taskId> --permanently      # Hard delete (hidden trash)
todoz restore-task <taskId>                   # Restore from trash

# List operations
todoz add-list --name "Groceries"
todoz update-list <listId> --name "..."
todoz delete-list <listId>                    # Deletes all tasks in list

# Utility
todoz compact                                  # Compact event log
todoz logs [--app <appName>] [--since <time>] # View operation logs
```

### Output Format

All commands return JSON:

**Success:**
```json
{
  "ok": true,
  "data": { "id": "...", "title": "..." }
}
```

**Error:**
```json
{
  "ok": false,
  "error": "task_not_found",
  "message": "Task abc123 does not exist"
}
```

### Integration Example (Node.js)

```javascript
const { execSync } = require('child_process');

function todoLoad() {
  const result = JSON.parse(execSync('todoz load').toString());
  if (!result.ok) throw new Error(result.message);
  return result.data;
}

function todoAdd(title, date, listId) {
  const cmd = `todoz add-task --title "${title}" --date ${date} --list ${listId}`;
  const result = JSON.parse(execSync(cmd).toString());
  return result.ok ? result.data : null;
}
```

---

## 3. Storage Architecture

### Directory Structure

```
$TODO_LIB_HOME/               # Environment variable, defaults to ~/.config/todoz/
├── events.jsonl              # Append-only event log (each line = 1 event)
├── events.jsonl.backup       # Backup during compaction
├── .lock                      # Temporary lock file for atomic appends
└── logs/
    ├── app-1.log             # Per-app operation logs
    └── app-2.log
```

### Event Log (events.jsonl)

Each line is a complete JSON object representing one mutation:

```json
{"type":"list_created","listId":"list-001","name":"Groceries","createdAt":"2026-06-03T10:00:00.000000+03:00"}
{"type":"task_created","taskId":"task-001","listId":"list-001","title":"Buy milk","date":"2026-06-05","description":"2L","createdAt":"2026-06-03T10:01:00.000000+03:00"}
{"type":"task_updated","taskId":"task-001","updates":{"description":"2L whole milk"},"updatedAt":"2026-06-03T10:02:00.000000+03:00"}
{"type":"task_completed","taskId":"task-001","completedAt":"2026-06-03T11:00:00.000000+03:00"}
{"type":"task_deleted","taskId":"task-001","deletedAt":"2026-06-03T12:00:00.000000+03:00"}
{"type":"task_permanently_deleted","taskId":"task-001","permanentlyDeletedAt":"2026-06-03T12:01:00.000000+03:00"}
{"type":"list_deleted","listId":"list-001","deletedAt":"2026-06-03T12:02:00.000000+03:00"}
```

**Event types:**
- `list_created`, `list_updated`, `list_deleted`
- `task_created`, `task_updated`, `task_completed`, `task_deleted`, `task_permanently_deleted`, `task_restored`

### State Computation

On every `todoz load` command:
1. Read entire events.jsonl
2. Replay all events in order
3. Build in-memory state (lists + tasks + trash)
4. Return JSON to caller

**Why full load every time?** Simplicity and safety. Eliminates any possibility of stale state or partial loads causing conflicts.

### Compaction

When events.jsonl grows large (e.g., >10,000 events):
```bash
todoz compact
```

Process:
1. Load full state
2. Write state.json (computed state snapshot)
3. Write new events.jsonl (only current + future events)
4. Keep events.jsonl.backup

Subsequent loads can use state.json as starting point, then replay only recent events (faster for large logs).

---

## 4. Concurrency & Safety

### Multi-Process Access

Multiple apps calling `todoz` simultaneously is safe:

```
App 1: todoz add-task --title "Task 1" --list L1
App 2: todoz add-task --title "Task 2" --list L1
App 3: todoz complete-task task-001
```

**Guarantees:**
1. Each invocation loads full log (no stale state)
2. Mutation computed from current state
3. New event appended atomically (OS-level file locking or atomic write)
4. Binary exits
5. Next app sees new event in log

**No conflicts possible.** Append-only + atomic writes = sequential consistency.

### Atomicity

Event append uses:
- **macOS/Linux:** `fcntl` locks or `atomic.WriteFile`
- **Windows:** `LockFile` API

Ensures partial writes never occur.

---

## 5. Trash System (Soft Delete)

Tasks are never truly deleted. Instead:

**First delete (soft):**
```bash
todoz delete-task task-001
```
Event: `task_deleted` with `deletedAt` timestamp. Task status: `isDeleted: true`.

When app loads: `trash[]` array includes this task.

**Second delete (permanent/hidden):**
```bash
todoz delete-task task-001 --permanently
```
Event: `task_permanently_deleted`. Task status: `isHiddenTrash: true`.

When app loads: trash array does NOT include this task (hidden).

**Restore:**
```bash
todoz restore-task task-001
```
Event: `task_restored`. Status back to pending/completed.

**Reality:** Event log retains all history forever. Users never see permanently deleted tasks, but audit trail exists.

---

## 6. Comprehensive Logging

Every request logged with full details: parameters, response, result, timestamp (microsecond precision), duration. Success and failure both logged.

### Log Location

`$TODO_LIB_HOME/logs/<app-name>.log`

### Log Format

Chronological, one request per block. Full parameters, full response, microsecond timestamps.

```
[2026-06-03T10:30:45.123456+03:00]
REQUEST: todoz add-task --title "Buy milk" --date 2026-06-05 --list list-001
RESPONSE: ok=true
DATA: {"id":"task-001","title":"Buy milk","date":"2026-06-05","listId":"list-001","status":"pending","createdAt":"2026-06-03T10:30:45.123456+03:00"}
DURATION: 1ms

[2026-06-03T10:30:46.234567+03:00]
REQUEST: todoz add-task --title "Invalid" --date invalid-date --list list-001
RESPONSE: ok=false
ERROR: invalid_date
MESSAGE: Date format must be YYYY-MM-DD, got: invalid-date
DURATION: 0ms

[2026-06-03T10:30:47.345678+03:00]
REQUEST: todoz complete-task task-001
RESPONSE: ok=true
DATA: {"id":"task-001","status":"completed","completedAt":"2026-06-03T10:30:47.345678+03:00"}
DURATION: 1ms

[2026-06-03T10:30:48.456789+03:00]
REQUEST: todoz load --list list-001
RESPONSE: ok=true
DATA: {"tasks":1247,"lists":12,"trash":3}
DURATION: 8ms

[2026-06-03T10:30:49.567890+03:00]
REQUEST: todoz delete-task nonexistent-id
RESPONSE: ok=false
ERROR: task_not_found
MESSAGE: Task nonexistent-id does not exist
DURATION: 0ms
```

App name determined by `TODO_APP_NAME` environment variable (defaults to binary name or `unknown`).

### Querying Logs

```bash
todoz logs                                # All logs, all apps
todoz logs --app my-app                   # Specific app
todoz logs --since "2026-06-03T10:00:00+03:00" # Since ISO timestamp with Istanbul TZ
todoz logs --app my-app --failed          # Only failed requests
todoz logs --app my-app --success         # Only successful requests
```

---

## 7. Timestamp Standard

All timestamps across events, logs, API responses: **ISO 8601 format with Istanbul timezone offset (+03:00).**

**Format:** `YYYY-MM-DDTHH:MM:SS.ssssss+03:00`

Examples:
```json
{
  "createdAt": "2026-06-03T10:30:45.123456+03:00",
  "updatedAt": "2026-06-03T11:15:22.654321+03:00"
}
```

Log entry:
```
[2026-06-03T10:30:45.123456+03:00]
REQUEST: todoz add-task ...
```

Query:
```bash
todoz logs --since "2026-06-03T10:00:00+03:00"
```

**Requirement:** System clock must be set to Istanbul timezone. No internet calls for time sync. Binary reads system local time directly.

**Why?** Unambiguous timestamps across all data and logs, even if data moved between machines. No dependency on internet or NTP.

---

## 8. Configuration

### Environment Variables

```bash
TODO_LIB_HOME     # Path to todo library directory (default: ~/.config/todoz/)
TODO_APP_NAME     # App identifier for logging (default: caller process name)
TODO_LOCK_TIMEOUT # File lock timeout in seconds (default: 5)
```

### Setup Example

```bash
# One-time setup
export TODO_LIB_HOME=~/my-projects/shared-todo

# All apps use this
my-app-1
my-app-2
my-app-3
```

---

## 8. Performance

### Benchmarks (Go implementation)

Scenario: 5000 tasks in log, 3 apps each add 50 + complete 30 tasks.

| Metric | Value |
|--------|-------|
| Load 5000 events | ~8ms |
| Per-operation (add/complete) | ~0.1ms |
| Write 80 events | ~1-2ms |
| Total (startup + 3 apps) | ~45ms |

**Memory:** ~50MB for 5000 tasks in memory.

**Scalability:** Tested up to 50,000 events. Load time ~40ms (still acceptable).

---

## 9. Testing Strategy

### Unit Tests
- Event application logic (does event X produce state Y?)
- State computation (replay produces correct state)
- Trash logic (soft delete, permanent delete, restore)

### Integration Tests
- CLI invocation (test each command)
- Load → mutate → load cycles
- Full workflow (create list, add tasks, complete, delete, restore)

### Concurrency Tests
- Simulate 3+ concurrent apps writing to same log
- Verify no corruption, no data loss
- Verify atomicity of appends

### Performance Tests
- Load time vs task count
- Write performance vs log size
- Memory usage under load

---

## 10. Error Handling

All errors returned as JSON:

```json
{
  "ok": false,
  "error": "task_not_found",
  "message": "Task abc123 does not exist in any list"
}
```

**Error codes:**
- `task_not_found`: Task ID doesn't exist
- `list_not_found`: List ID doesn't exist
- `invalid_date`: Date not in YYYY-MM-DD format
- `invalid_operation`: e.g., complete already-completed task
- `permission_denied`: File I/O permission error
- `corrupted_log`: Event log corrupted (recovery attempted)
- `lock_timeout`: Could not acquire file lock in time

---

## 11. Implementation Principles (Non-Negotiable)

These constraints govern HOW the project is built, independent of features:

### Code Quality
- **English only:** No Turkish anywhere — code, comments, docs, identifiers, log messages. 100% English.
- **KISS:** Simplest implementation that works. No clever tricks.
- **YAGNI:** Build only what the spec requires. No speculative abstractions.
- **TDD:** Tests written first, then implementation. Red → Green → Refactor.

### AI-Native Documentation
- **Hyper-documented:** Every system, architecture decision, and code unit explained. No unexplained architecture or code.
- **Readable for AI agents:** Structured so AI coding agents read and understand effortlessly. Clear naming, explicit intent, doc comments on every exported symbol.

### Extreme Modularity
- **No long files:** Every file small, focused, single-purpose.
- **Meaningful decomposition:** Split into small, meaningful modules. No grab-bag files.
- **Meaningful folder structure:** Organize folders by clear responsibility/domain. Most modular layout possible.

### Required Documentation Artifacts

In addition to the source, the project must ship two top-level Markdown documents:

1. **`system_explain.md`** — Full technical architecture document.
   - Covers the entire project architecture to the deepest technical detail.
   - Includes diagrams (ASCII/Mermaid), tables, data flows, module map, event lifecycle, concurrency model, storage layout.
   - Audience: developers and AI agents who need to understand the system internals.

2. **`USAGE.md`** — Integration guide for consuming apps.
   - How any external app communicates with todoz (subprocess + JSON contract).
   - Full command directory (every command, every flag).
   - Many concrete examples across languages.
   - Edge-case scenarios with example request/response pairs (errors, trash, concurrent access, overdue computation, empty state, etc).
   - Audience: developers integrating todoz into their apps.

---

## 12. Future Considerations

(Out of scope for v1, documented for reference)

- **Search/filtering:** Query API (find tasks by title, date, status)
- **Undo/redo:** Transaction support
- **Sync:** Multi-device sync (via cloud or P2P)
- **Language bindings:** Native libraries for Node.js, Python, etc.
- **Web UI:** Simple web interface for task management

---

## Conclusion

Todoz is a minimal, robust, offline-first todo library optimized for integration. Its append-only architecture eliminates conflicts, its CLI interface enables language-agnostic use, and its portable configuration allows deployment anywhere.

**Ready for implementation.**
