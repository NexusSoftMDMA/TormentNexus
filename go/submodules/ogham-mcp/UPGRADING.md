# Upgrading Ogham

## Upgrading to v0.11.0 from v0.10.x

**Short version: run the upgrade script, restart the server.**

```bash
# 1. Pull or install the new version (pip / uv / uvx)
uv tool upgrade --refresh ogham-mcp          # if installed via uv tool
# or: pip install -U ogham-mcp
# or: uvx --refresh ogham-mcp                # if running via uvx

# 2. Apply the schema migrations to your database
./sql/upgrade.sh "$DATABASE_URL"

# 3. Restart your MCP client (Claude Desktop / Cursor / etc.)
```

That's it. Below is the detail for anyone who wants to know what changed
and why.

---

## What changed in v0.11.0

The headline feature is a **memory lifecycle state machine**
(FRESH → STABLE ↔ EDITING). Every memory now has a stage that tracks
whether it is newly written, settled, or recently retrieved. Stage
transitions happen automatically via hooks and background sweeps — no
new API surface the agent has to learn.

Under the hood, this ships three SQL migrations:

| Migration | What it does | Matters to self-hosters? |
|-----------|--------------|:-:|
| `025_memory_lifecycle.sql` | Adds lifecycle columns temporarily + per-profile decay tuning params on `profile_settings`. | Yes — idempotent; no-op if already applied. |
| `026_memory_lifecycle_split.sql` | Moves lifecycle state from `memories` into a dedicated `memory_lifecycle` table so lifecycle updates do not touch the HNSW vector index. Adds two triggers that auto-maintain lifecycle rows when memories are inserted or moved between profiles. | Yes — this is the important one. Your retrieval latency depends on this split. |
| `027_audit_log_backfill.sql` | Creates the `audit_log` table if it does not exist. Some older installs were created before `audit_log` was part of the baseline schema. | Only if your DB predates v0.9.2 roughly. No-op otherwise. |

All three migrations are **idempotent** (safe to re-run) and **additive**
for data (nothing is deleted that cannot be reconstructed from existing
data). `026` drops `memories.stage` and `memories.stage_entered_at` only
after backfilling both into `memory_lifecycle`.

## Who needs to do what

### You installed via `uv tool` / `uvx` / `pip`

```bash
# upgrade the package
uv tool upgrade --refresh ogham-mcp

# apply the migrations (copies of the SQL ship with the package)
# find the migrations dir:
OGHAM_PKG=$(python -c "import ogham, os; print(os.path.dirname(ogham.__file__))")
ls "$OGHAM_PKG/sql/migrations"

# apply:
for f in "$OGHAM_PKG/sql/migrations"/*.sql; do
    psql "$DATABASE_URL" -f "$f" -v ON_ERROR_STOP=1
done
```

Or clone the repo once and use the `upgrade.sh` helper, then discard
the clone:

```bash
git clone --depth 1 https://github.com/ogham-mcp/ogham-mcp /tmp/ogham-up
/tmp/ogham-up/sql/upgrade.sh "$DATABASE_URL"
rm -rf /tmp/ogham-up
```

### You run from a git checkout

```bash
git pull
./sql/upgrade.sh "$DATABASE_URL"
```

### You run from the Docker image (`ghcr.io/ogham-mcp/ogham-mcp`)

Pull the new image tag and recreate your container. To apply
migrations, either:

- Run `upgrade.sh` from the host against your DB (fastest), or
- Exec into the container and run `python -m ogham.cli ...` (future
  release may add a built-in `ogham migrate` command; for v0.11.0 the
  shell script is the path).

### You are a fresh install (no existing Ogham DB)

You do NOT need `upgrade.sh`. `sql/schema.sql` (or the variant for your
backend) already contains the post-v0.11.0 shape. Just run the baseline
once:

```bash
psql "$DATABASE_URL" -f sql/schema_postgres.sql      # plain Postgres / Neon
# or: sql/schema.sql                                  # Supabase Cloud
# or: sql/schema_selfhost_supabase.sql                # self-hosted Supabase
```

## Supabase users — paste-ready SQL

If you prefer to paste SQL into the Supabase Dashboard SQL Editor
instead of running `upgrade.sh`, copy the contents of each of these in
order — each is its own transaction:

1. `sql/migrations/025_memory_lifecycle.sql`
2. `sql/migrations/026_memory_lifecycle_split.sql`
3. `sql/migrations/027_audit_log_backfill.sql`

Run all three in sequence. The Supabase dashboard executes SQL as the
`postgres` role, so DDL + RLS setup both work.

## Verifying the upgrade

After running the migrations, confirm the end state:

```sql
SELECT 'memory_lifecycle exists' AS check,
       EXISTS (SELECT 1 FROM pg_tables
                 WHERE schemaname='public' AND tablename='memory_lifecycle')::text AS value
UNION ALL SELECT 'memories.stage absent',
    (NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='memories' AND column_name='stage'))::text
UNION ALL SELECT 'init_lifecycle trigger present',
    EXISTS (SELECT 1 FROM pg_trigger
              WHERE tgname='memories_init_lifecycle' AND NOT tgisinternal)::text
UNION ALL SELECT 'all memories have lifecycle rows',
    ((SELECT count(*) FROM memories) =
     (SELECT count(*) FROM memory_lifecycle))::text;
```

Expected: four `true` rows. If `memory_lifecycle` row count does not
match `memories` row count, the backfill did not complete — re-run
`026_memory_lifecycle_split.sql` (idempotent).

## If something goes wrong

Migrations are reversible, but rollbacks are **explicitly opt-in** to
prevent accidental destruction. See
[`sql/migrations/rollback/README.md`](sql/migrations/rollback/README.md)
for the ritual.

**Common issue: "relation memory_lifecycle does not exist" after upgrade**
  — means `026` did not apply. Re-run `./sql/upgrade.sh`. Forward
migrations are idempotent.

**Common issue: "role anon does not exist" on plain Postgres when
applying `027`** — fixed in the v0.11.0 release. If you pulled a
pre-release copy, update to the final cut.

**Anything else** — open an issue at
https://github.com/ogham-mcp/ogham-mcp/issues with the error text and
the output of:

```sql
SELECT extname, extversion FROM pg_extension WHERE extname = 'vector';
SELECT version();
SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY 1;
```

## What this unlocks

With the lifecycle in place, v0.11.0 also wires:

- **Automatic stage sweeps** at session start (via the `session_start`
  hook). Memories that have dwelled long enough and clear an
  importance/surprise gate are promoted to `stable` automatically.
- **30-minute editing windows** on retrieved memories, enabling
  in-place refinements of recent memories.
- **Hebbian edge strengthening** — memories retrieved together build
  stronger graph connections over time (η=0.01 per co-retrieval).
- A new **`advance_lifecycle`** MCP tool for manually triggering a
  sweep (useful after bulk imports).

For the dashboard audience, the **Go CLI dashboard** (`ogham dashboard`)
gains a Lifecycle pill row on the Overview showing FRESH / STABLE /
EDITING counts for the active profile, backed by the same
`memory_lifecycle` table. Mixed-version safe — older DB schemas fall
back to counting all memories as `fresh` without error.

See `CHANGELOG.md` for the full v0.11.0 entry.
