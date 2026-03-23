# Database Migrations

- Migration files are authoritative for schema changes. Runtime does not use GORM `AutoMigrate`.
- Each driver keeps its own ordered SQL files under `postgres/` and `sqlite/`.
- Filenames use zero-padded numeric versions: `0000_schema_migrations.sql`, `0001_initial.sql`, `0002_indexes.sql`.
- New migrations must be additive, deterministic, and safe to re-run through version tracking.
- Keep PostgreSQL and SQLite schemas semantically aligned for shared domain behavior.