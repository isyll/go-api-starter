#!/usr/bin/env bash

# Renumber migration files (increment or decrement) starting from a given number.
# Usage: ./scripts/renumber_migrations.sh <start_number> [migrations_path] [--decrement]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if [[ $# -eq 0 ]]; then
    echo "Usage: $0 <start_number> [migrations_path] [--decrement]"
    echo "Example: $0 18                        (increment from 18)"
    echo "Example: $0 18 migrations             (increment from 18 in migrations/)"
    echo "Example: $0 21 --decrement            (decrement from 21)"
    echo "Example: $0 21 migrations --decrement"
    exit 1
fi

START_NUM=$1
MIGRATION_DIR="$BACKEND_DIR/migrations"
DECREMENT=false

shift
while [[ $# -gt 0 ]]; do
    case "$1" in
        --decrement|-d)
            DECREMENT=true
            ;;
        *)
            # Allow absolute or relative paths
            if [[ "$1" = /* ]]; then
                MIGRATION_DIR="$1"
            else
                MIGRATION_DIR="$BACKEND_DIR/$1"
            fi
            ;;
    esac
    shift
done

cd "$MIGRATION_DIR" || {
    echo "Error: Cannot access directory $MIGRATION_DIR"
    exit 1
}

if ! [[ "$START_NUM" =~ ^[0-9]+$ ]]; then
    echo "Error: The parameter must be a number"
    exit 1
fi

# Collect six-digit-prefixed migration files via globbing.
# nullglob makes the glob expand to nothing (rather than the
# literal pattern) when no files match.
shopt -s nullglob
mapfile -t FILES < <(printf '%s\n' [0-9][0-9][0-9][0-9][0-9][0-9]_*.sql)
shopt -u nullglob

if [[ ${#FILES[@]} -eq 0 ]]; then
    echo "No migration files found in: $(pwd)"
    echo "Expected format: 000001_filename.sql"
    exit 1
fi

if [[ "$DECREMENT" = true ]]; then
    IFS=$'\n' mapfile -t FILES < <(printf '%s\n' "${FILES[@]}" | sort)
    OPERATION="decrement"
    CHANGE=-1
else
    IFS=$'\n' mapfile -t FILES < <(printf '%s\n' "${FILES[@]}" | sort -r)
    OPERATION="increment"
    CHANGE=1
fi

echo "=== Directory: $(pwd) ==="
echo "=== Operation: ${OPERATION} ==="
echo "=== Files to renumber ==="
for file in "${FILES[@]}"; do
    NUM="${file:0:6}"
    NUM_DEC=$((10#$NUM))
    if [[ "$NUM_DEC" -ge "$START_NUM" ]]; then
        echo "$file"
    fi
done

echo ""
read -r -p "Do you want to continue? (y/n) " -n 1 REPLY
echo ""

if [[ ! $REPLY =~ ^[OoYy]$ ]]; then
    echo "Operation canceled"
    exit 0
fi

echo ""
echo "=== Starting renumbering ==="

for file in "${FILES[@]}"; do
    CURRENT_NUM="${file:0:6}"
    CURRENT_NUM=$((10#$CURRENT_NUM))

    if [[ "$CURRENT_NUM" -ge "$START_NUM" ]]; then
        NEW_NUM=$((CURRENT_NUM + CHANGE))

        if [[ "$NEW_NUM" -lt 1 ]]; then
            echo "⚠ Warning: Cannot decrement $file below 000001, skipping"
            continue
        fi

        NEW_NUM_FORMATTED=$(printf "%06d" "$NEW_NUM")
        REST_OF_NAME="${file:6}"
        NEW_FILE="${NEW_NUM_FORMATTED}${REST_OF_NAME}"

        mv "$file" "$NEW_FILE"
        echo "✓ $file -> $NEW_FILE"
    fi
done

echo ""
echo "=== Renumbering completed ==="
if [[ "$DECREMENT" = true ]]; then
    echo "You can now use number $START_NUM for a new migration"
else
    echo "Files from $START_NUM onward shifted up by 1"
fi
