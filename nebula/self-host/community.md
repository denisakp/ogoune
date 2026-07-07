# Community (SQLite)

The Community Edition runs as a single binary with **no external dependencies** — SQLite for storage, an in-process TimingWheel scheduler.

## Run

```bash
DB_DRIVER=sqlite \
  SQLITE_PATH=./ogoune.db \
  SCHEDULER_MODE=timingwheel \
  APP_SECRET_KEY=$(openssl rand -hex 32) \
  ./ogoune
```

## Build from source

```bash
make build        # frontend + backend → dist/ogoune
```

Best for: a single node, small-to-medium fleets, homelab, or evaluation.
