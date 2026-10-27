#!/usr/bin/env bash
# Run migrations against a private data copy with all networking disabled.
set -euo pipefail
umask 077

if [[ $# != 3 || $EUID != 0 ]]; then
  echo "Usage (as root): $0 SOURCE_DATA IMAGE NEW_OUTPUT_DIRECTORY" >&2
  exit 2
fi
source_dir=$(realpath "$1")
candidate_image=$2
output_dir=$(realpath -m "$3")
if [[ $output_dir == "$source_dir" || $output_dir == "$source_dir/"* || $output_dir == *"'"* ]]; then
  echo "Output must be a new directory outside the source; single quotes are unsupported." >&2
  exit 2
fi
[[ -f $source_dir/data.db ]]
mkdir -m 700 "$output_dir"
mkdir "$output_dir/data"
cp -a "$source_dir/." "$output_dir/data/"

# Online SQLite backups include committed WAL data. Discard only sidecars in
# the new copy, never in the live directory.
for database in data.db auxiliary.db; do
  if [[ -f $source_dir/$database ]]; then
    rm -f "$output_dir/data/$database" "$output_dir/data/$database-wal" "$output_dir/data/$database-shm"
    sqlite3 -readonly "$source_dir/$database" ".backup '$output_dir/data/$database'"
  fi
done

counts() {
  sqlite3 -readonly "$output_dir/data/data.db" \
    "SELECT 'systems',count(*) FROM systems UNION ALL SELECT 'alerts',count(*) FROM alerts UNION ALL SELECT 'alerts_history',count(*) FROM alerts_history UNION ALL SELECT 'system_stats',count(*) FROM system_stats UNION ALL SELECT 'container_stats',count(*) FROM container_stats UNION ALL SELECT 'smart_devices',count(*) FROM smart_devices UNION ALL SELECT 'zfs_pools',count(*) FROM zfs_pools;"
}
counts > "$output_dir/counts-before.txt"
container_id=''
cleanup() {
  if [[ -n $container_id ]]; then
    docker logs "$container_id" > "$output_dir/hub.log" 2>&1 || true
    docker rm -f "$container_id" >/dev/null
  fi
}
trap cleanup EXIT
container_id=$(docker run -d --network none --memory 1g --cpus 2 \
  -e DISABLE_SSH=true -e APP_URL=http://127.0.0.1:8090 \
  -v "$output_dir/data:/beszel_data" "$candidate_image")
container_pid=$(docker inspect -f '{{.State.Pid}}' "$container_id")
ready=false
for ((attempt=0; attempt<45; attempt++)); do
  if nsenter -t "$container_pid" -n curl -fsS --max-time 1 http://127.0.0.1:8090/api/health > "$output_dir/health.json" 2>/dev/null; then
    ready=true
    break
  fi
  [[ $(docker inspect -f '{{.State.Running}}' "$container_id") == true ]] || break
  sleep 1
done
if [[ $ready != true ]]; then
  echo "Candidate failed to become healthy; private log: $output_dir/hub.log" >&2
  exit 1
fi
docker stop -t 10 "$container_id" >/dev/null
counts > "$output_dir/counts-after.txt"
diff -u "$output_dir/counts-before.txt" "$output_dir/counts-after.txt"
[[ $(sqlite3 -readonly "$output_dir/data/data.db" 'PRAGMA integrity_check;') == ok ]]
fork_options=$(sqlite3 -readonly "$output_dir/data/data.db" \
  "SELECT count(*) FROM _collections c,json_each(c.fields) f,json_each(f.value,'$.values') v WHERE c.name='alerts' AND json_extract(f.value,'$.name')='name' AND v.value IN ('Reboot','Filesystem');")
[[ $fork_options == 2 ]]
echo "PASS: candidate healthy; retained system, alert, history, and ZFS row counts; database integrity OK; both fork alert options available."
echo "Private validation copy: $output_dir"
