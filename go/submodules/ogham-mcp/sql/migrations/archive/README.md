# Archived migrations

These migrations are from the OpenBrain era (pre-v0.4.0) and are already
incorporated into the main schema files. Do not run them on a current
installation -- they may alter vector dimensions or make other destructive
changes that were specific to earlier database layouts.

If you're upgrading from a pre-v0.4.0 database, start fresh with the
appropriate schema file instead:

- `sql/schema_postgres.sql` -- Neon / vanilla PostgreSQL
- `sql/schema.sql` -- Supabase Cloud
- `sql/schema_selfhost_supabase.sql` -- Self-hosted Supabase Docker
