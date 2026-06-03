# Todoz — Integration Guide

## 1. What Todoz Is

A standalone binary your app calls as a subprocess. Send a command, get one JSON object back. No server, no library linking, no shared database parsing.

## 2. Setup

```bash
export TODO_LIB_HOME=/path/to/shared-todo   # where data lives (portable)
export TODO_APP_NAME=my-app                 # your app's name (for the audit log)
```

## 3. The Response Contract

Every command prints ONE JSON object:

**Success:**
```json
{"ok": true, "data": { ... }}
```

**Failure:**
```json
{"ok": false, "error": "task_not_found", "message": "Task x does not exist"}
```

Exit code is 0 on success, 1 on failure. Always check `ok`.

## 4. Command Directory

### load
`todoz load [flags]`

Returns: `{"lists": [...], "tasks": [...], "trash": [...]}`
When `--group-by` is set, returns: `{"groups": {"2026-06-01": [...]}, "lists": [...], "summary": {...}}`

#### Filtering Flags
| Flag | Description |
|------|-------------|
| `--days-back N` | Tasks with date ≥ today minus N days |
| `--after-date YYYY-MM-DD` | Tasks with date ≥ this date (mutually exclusive with `--days-back`) |
| `--before-date YYYY-MM-DD` | Tasks with date ≤ this date |
| `--status pending\|completed` | Only tasks with this status |
| `--overdue` | Only pending tasks where date < today |
| `--lists "id1,id2"` | Only tasks in these list IDs (comma-separated) |
| `--list id` | Filter to a single list ID (legacy, still supported) |
| `--search TEXT` | Substring match on title or description (case-insensitive) |
| `--no-trash` | Exclude trash from response |
| `--trash-only` | Return only trash tasks (mutually exclusive with `--no-trash`) |

#### Sorting Flags
| Flag | Description |
|------|-------------|
| `--sort-by date\|title\|created\|status` | Sort field (default: date) |
| `--sort-reverse` | Reverse sort order |

#### Grouping Flags
| Flag | Description |
|------|-------------|
| `--group-by list\|date\|status` | Group tasks; changes response shape to `groups` map |

#### Output / Aggregation Flags
| Flag | Description |
|------|-------------|
| `--summary` | Append `{"total":N,"pending":N,"completed":N,"overdue":N}` to response |
| `--count` | Same as `--summary` (only counts) |

#### Conflict Rules
- `--days-back` and `--after-date` cannot be used together
- `--no-trash` and `--trash-only` cannot be used together

#### Examples
```bash
# Last 7 days, pending only
todoz load --days-back 7 --status pending

# Overdue tasks grouped by list
todoz load --overdue --group-by list

# Search in a date range, sorted descending
todoz load --after-date 2026-05-01 --before-date 2026-06-01 --search "urgent" --sort-by date --sort-reverse

# Summary stats for all tasks
todoz load --summary

# Tasks in two lists, grouped by date
todoz load --lists "list-abc,list-def" --group-by date
```

### add-task
`todoz add-task --title <text> --date <YYYY-MM-DD> --list <listId> [--description <text>]`

Creates a new task in pending status.

Example:
```bash
todoz add-task --title "Buy milk" --date 2026-06-05 --list list-abc
```

### update-task
`todoz update-task <taskId> [--title <text>] [--description <text>] [--date <YYYY-MM-DD>]`

Updates one or more fields. At least one field required.

### complete-task
`todoz complete-task <taskId>`

Marks task as completed.

### delete-task
`todoz delete-task <taskId> [--permanently]`

First call soft-deletes to trash. With `--permanently`, hides from apps forever.

### restore-task
`todoz restore-task <taskId>`

Brings a soft-deleted task back out of trash.

### add-list
`todoz add-list --name <text>`

Creates a new list.

### update-list
`todoz update-list <listId> --name <text>`

Renames a list.

### delete-list
`todoz delete-list <listId>`

Soft-deletes a list and cascades to all its tasks.

### compact
`todoz compact`

Rewrites the event log to its minimal form (safe operation).

## 5. Integration Examples

**Node.js:**
```javascript
const { execSync } = require('child_process');

function todoz(cmd) {
  const output = execSync(`todoz ${cmd}`, {
    env: {
      ...process.env,
      TODO_LIB_HOME: '/path/to/todoz-data',
      TODO_APP_NAME: 'my-app'
    }
  }).toString();
  return JSON.parse(output);
}

const result = todoz('load');
if (result.ok) {
  console.log(result.data.tasks);
}
```

**Python:**
```python
import subprocess
import json

def todoz(cmd):
    env = {'TODO_LIB_HOME': '/path/to/todoz-data', 'TODO_APP_NAME': 'my-app'}
    result = subprocess.run(
        ['todoz'] + cmd.split(),
        capture_output=True, text=True, env=env
    )
    return json.loads(result.stdout)

result = todoz('load')
if result['ok']:
    print(result['data']['tasks'])
```

**Bash:**
```bash
export TODO_LIB_HOME=/path/to/todoz-data
export TODO_APP_NAME=my-app

todoz load | jq '.data.tasks[]'
```

**Go:**
```go
import "os/exec"
import "encoding/json"

func TodozLoad() (Data, error) {
    cmd := exec.Command("todoz", "load")
    cmd.Env = os.Environ()
    cmd.Env = append(cmd.Env, "TODO_LIB_HOME=/path/to/todoz-data")
    out, _ := cmd.Output()
    var result Response
    json.Unmarshal(out, &result)
    return result.Data, nil
}
```

## 6. Edge-Case Scenarios

| Scenario | Request | Response |
|----------|---------|----------|
| Empty store | `load` | `lists:[], tasks:[], trash:[]` |
| Invalid date | `add-task --date 06-05-2026` | `ok:false, error:"invalid_date"` |
| Soft-delete twice | First: `delete-task X`, Second: `delete-task X --permanently` | First → trash, Second → hidden |
| Restore hidden task | `restore-task` on hidden task | `ok:false, error:"task_not_found"` (must be soft-deleted first) |
| Concurrent writers | 50 processes append simultaneously | Lock serializes; all writes succeed, count matches 50 |
| Overdue | pending task, date < today | `isOverdue: true` (status stays pending) |
| Unknown ID | `complete-task bad-id` | `ok:false, error:"task_not_found"` |
| Cascade delete | `delete-list L` with 5 tasks | List and all 5 tasks soft-deleted |

## 7. Operation Log

Written to `$TODO_LIB_HOME/logs/<app>.log`

Format:
```
[2026-06-03T10:30:45.123456+03:00]
REQUEST: todoz add-task --title "Buy milk"
RESPONSE: ok=true
DATA: {"id":"task-abc","title":"Buy milk","listId":"list-xyz","date":"2026-06-05"}
DURATION: 2ms

[2026-06-03T10:30:47.654321+03:00]
REQUEST: todoz delete-task bad-id
RESPONSE: ok=false
ERROR: task_not_found
MESSAGE: Task bad-id does not exist
DURATION: 1ms
```

Read the oplog to debug failures or audit who did what.

## 8. Performance Notes

Full replay per call. For very large logs (>100k events), run `todoz compact` periodically.
