#!/usr/bin/env bash
# Upgrade an existing Ogham database to the latest schema.
#
# Runs migration files in sql/migrations/ (excludes archive/).
# Each migration is idempotent -- safe to re-run.
#
# Usage:
#   ./sql/upgrade.sh postgresql://user:pass@host/dbname
#   ./sql/upgrade.sh "$DATABASE_URL"
#
# For Docker-based Postgres:
#   docker exec -i <container> psql -U <user> -d <db> < sql/migrations/012_temporal_columns.sql
#   docker exec -i <container> psql -U <user> -d <db> < sql/migrations/013_halfvec_compression.sql
#
# NOTE: Legacy migrations (002-011) are archived in sql/migrations/archive/.
# They are already incorporated into the main schema files and should NOT
# be run on current installations.

set -euo pipefail

if [ $# -eq 0 ]; then
    echo "Usage: $0 <connection-string>"
    echo ""
    echo "Examples:"
    echo "  $0 postgresql://user:pass@host:5432/dbname"
    echo "  $0 \"\$DATABASE_URL\""
    echo ""
    echo "Runs migration files in sql/migrations/ (excludes archive/)."
    echo "Each migration is idempotent -- safe to re-run."
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
MIGRATIONS_DIR="$SCRIPT_DIR/migrations"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "Error: migrations directory not found at $MIGRATIONS_DIR"
    exit 1
fi

# Only run top-level .sql files (not archive/)
migrations=$(find "$MIGRATIONS_DIR" -maxdepth 1 -name "*.sql" | sort)

if [ -z "$migrations" ]; then
    echo "No migrations to apply."
    exit 0
fi

echo "Ogham schema upgrade"
echo "===================="
echo ""

# Safety check: verify we can connect
if ! psql "$1" -c "SELECT 1" > /dev/null 2>&1; then
    echo "Error: cannot connect to database"
    echo "Check your connection string and try again."
    exit 1
fi

# Check pgvector version for halfvec support
pgvector_version=$(psql "$1" -t -c "SELECT extversion FROM pg_extension WHERE extname = 'vector'" 2>/dev/null | tr -d ' ')
if [ -n "$pgvector_version" ]; then
    echo "pgvector version: $pgvector_version"
    major=$(echo "$pgvector_version" | cut -d. -f1)
    minor=$(echo "$pgvector_version" | cut -d. -f2)
    if [ "$major" -eq 0 ] && [ "$minor" -lt 7 ]; then
        echo ""
        echo "WARNING: pgvector $pgvector_version detected. Migration 013 (halfvec)"
        echo "requires pgvector >= 0.7.0. Skipping halfvec migration."
        echo "Upgrade pgvector first, then re-run this script."
        skip_halfvec=1
    fi
fi
echo ""

count=0
for migration in $migrations; do
    name=$(basename "$migration")

    # Skip halfvec migration if pgvector is too old
    if [ "${skip_halfvec:-0}" = "1" ] && [ "$name" = "013_halfvec_compression.sql" ]; then
        echo "Skipping $name (pgvector < 0.7.0)"
        continue
    fi

    echo "Applying $name..."
    psql "$1" -f "$migration" -v ON_ERROR_STOP=1 2>&1 | grep -v "^$" | sed 's/^/  /'
    count=$((count + 1))
done

echo ""
echo "Done. Applied $count migration(s)."
