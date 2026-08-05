#!/usr/bin/env bash
set -euo pipefail

# Restores only into a disposable database inside the local Compose service.
backup_file=${1:?usage: restore-postgres-isolated.sh <backup.dump>}
restore_database=${ETHPHISH_RESTORE_DATABASE:-ethphish_restore_verify}

if [[ ! -f "$backup_file" ]]; then
  echo "backup not found: $backup_file" >&2
  exit 2
fi
if [[ "$restore_database" != "ethphish_restore_verify" ]]; then
  echo "refusing non-isolated restore database: $restore_database" >&2
  exit 2
fi

docker compose exec -T postgres psql -U ethphish -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS ${restore_database} WITH (FORCE)" -c "CREATE DATABASE ${restore_database}"
docker compose exec -T postgres pg_restore --clean --if-exists --no-owner --no-privileges -U ethphish -d "$restore_database" <"$backup_file"
docker compose exec -T postgres psql -U ethphish -d "$restore_database" -v ON_ERROR_STOP=1 -Atc "SELECT max(version_id) FROM goose_db_version WHERE is_applied = true"

echo "Isolated restore validated in database: $restore_database"
