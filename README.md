# todoz

A standalone, offline, append-only todo store you call as a subprocess from any app in any language. No daemon. No conflicts. Nothing ever truly deleted.

## Build

```bash
go build -o todoz ./cmd/todoz
```

## Quick Start

```bash
export TODO_LIB_HOME=~/todoz-data
export TODO_APP_NAME=my-app
./todoz add-list --name "Groceries"
./todoz load
```

## Documentation

- **[USAGE.md](USAGE.md)** — how to integrate todoz into your app (commands, examples, edge cases).
- **[system_explain.md](system_explain.md)** — full technical architecture.

## Principles

English-only, KISS, YAGNI, TDD, stdlib-only, extreme modularity.
