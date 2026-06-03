#!/bin/bash
# Smoke test driver for todoz CLI
# Tests basic command flows: add-list, add-task, complete-task, load, delete

set -e

export TODO_LIB_HOME="${TODO_LIB_HOME:-/tmp/todoz-smoke-$(date +%s)}"
export TODO_APP_NAME="${TODO_APP_NAME:-smoke-test}"
mkdir -p "$TODO_LIB_HOME"

TODOZ="${TODOZ:-./todoz}"
BINARY=$(command -v $TODOZ || echo "./todoz")

if [ ! -x "$BINARY" ]; then
  echo "Error: todoz binary not found at $BINARY"
  exit 1
fi

echo "=== Todoz Smoke Test ==="
echo "Using: $BINARY"
echo "Home: $TODO_LIB_HOME"
echo ""

# Test 1: Add a list
echo "[1] Add-list"
LIST_JSON=$($BINARY add-list --name "Groceries")
LIST_ID=$(echo "$LIST_JSON" | jq -r '.data.id')
if [ "$LIST_ID" = "null" ] || [ -z "$LIST_ID" ]; then
  echo "FAIL: add-list returned no ID"
  echo "$LIST_JSON"
  exit 1
fi
echo "OK: List created: $LIST_ID"
echo ""

# Test 2: Add a task
echo "[2] Add-task"
TASK_JSON=$($BINARY add-task --title "Buy milk" --date 2026-06-10 --list "$LIST_ID")
TASK_ID=$(echo "$TASK_JSON" | jq -r '.data.id')
if [ "$TASK_ID" = "null" ] || [ -z "$TASK_ID" ]; then
  echo "FAIL: add-task returned no ID"
  echo "$TASK_JSON"
  exit 1
fi
echo "OK: Task created: $TASK_ID"
echo ""

# Test 3: Add overdue task
echo "[3] Add-task (overdue date)"
TASK2_JSON=$($BINARY add-task --title "Buy eggs" --date 2026-06-01 --list "$LIST_ID" --description "1 dozen")
TASK2_ID=$(echo "$TASK2_JSON" | jq -r '.data.id')
echo "OK: Overdue task created: $TASK2_ID"
echo ""

# Test 4: Load and verify
echo "[4] Load (verify structure)"
LOAD_JSON=$($BINARY load)
LIST_COUNT=$(echo "$LOAD_JSON" | jq '.data.lists | length')
TASK_COUNT=$(echo "$LOAD_JSON" | jq '.data.tasks | length')
TASK1_OVERDUE=$(echo "$LOAD_JSON" | jq ".data.tasks[] | select(.id==\"$TASK2_ID\") | .isOverdue")
if [ "$LIST_COUNT" != "1" ] || [ "$TASK_COUNT" != "2" ]; then
  echo "FAIL: Expected 1 list, 2 tasks; got $LIST_COUNT, $TASK_COUNT"
  exit 1
fi
if [ "$TASK1_OVERDUE" != "true" ]; then
  echo "FAIL: Overdue task not marked isOverdue=true"
  exit 1
fi
echo "OK: Load returned 1 list, 2 tasks; overdue computed"
echo ""

# Test 5: Complete task
echo "[5] Complete-task"
COMPLETE_JSON=$($BINARY complete-task "$TASK_ID")
STATUS=$(echo "$COMPLETE_JSON" | jq -r '.data.status')
if [ "$STATUS" != "completed" ]; then
  echo "FAIL: complete-task status=$STATUS"
  exit 1
fi
echo "OK: Task marked completed"
echo ""

# Test 6: Delete task
echo "[6] Delete-task (soft delete)"
DELETE_JSON=$($BINARY delete-task "$TASK2_ID")
if [ "$(echo "$DELETE_JSON" | jq -r '.ok')" != "true" ]; then
  echo "FAIL: delete-task failed"
  exit 1
fi
echo "OK: Task soft-deleted"
echo ""

# Test 7: Verify trash
echo "[7] Load (verify trash)"
TRASH_JSON=$($BINARY load)
TRASH_COUNT=$(echo "$TRASH_JSON" | jq '.data.trash | length')
if [ "$TRASH_COUNT" != "1" ]; then
  echo "FAIL: Expected 1 task in trash; got $TRASH_COUNT"
  exit 1
fi
echo "OK: Trash contains 1 task"
echo ""

# Test 8: Restore task
echo "[8] Restore-task"
RESTORE_JSON=$($BINARY restore-task "$TASK2_ID")
if [ "$(echo "$RESTORE_JSON" | jq -r '.ok')" != "true" ]; then
  echo "FAIL: restore-task failed"
  exit 1
fi
echo "OK: Task restored"
echo ""

# Test 9: Update list
echo "[9] Update-list"
UPDATE_JSON=$($BINARY update-list "$LIST_ID" --name "Shopping")
NEW_NAME=$(echo "$UPDATE_JSON" | jq -r '.data.name')
if [ "$NEW_NAME" != "Shopping" ]; then
  echo "FAIL: update-list failed"
  exit 1
fi
echo "OK: List renamed to Shopping"
echo ""

# Test 10: Compact
echo "[10] Compact"
COMPACT_JSON=$($BINARY compact)
if [ "$(echo "$COMPACT_JSON" | jq -r '.ok')" != "true" ]; then
  echo "FAIL: compact failed"
  exit 1
fi
echo "OK: Log compacted"
echo ""

echo "=== All smoke tests passed ==="
