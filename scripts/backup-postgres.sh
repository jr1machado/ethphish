#!/usr/bin/env sh
set -eu

# Creates a compressed logical backup from the development Compose database.
# Backups live outside the repository so database contents cannot be committed.
backup_dir="${ETHPHISH_BACKUP_DIR:-/tmp/ethphish-backups}"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_file="${backup_dir}/ethphish-${timestamp}.dump"

mkdir -p "$backup_dir"
umask 077
docker compose exec -T postgres pg_dump --format=custom --no-owner --no-privileges -U ethphish -d ethphish >"$backup_file"
printf 'Backup created: %s\n' "$backup_file"
