#!/usr/bin/env bash
# Deploy a previously validated fork image and agent, with a coupled rollback.
set -euo pipefail
umask 077

[[ $EUID == 0 ]] || { echo "Run as root." >&2; exit 2; }
restore_compose() {
  cp -a "$backup_dir/compose.yml" "$compose_file"
  previous_image=$(cat "$backup_dir/image-id.txt")
  rollback_image="beszel-rollback:${previous_image#sha256:}"
  docker tag "$previous_image" "$rollback_image"
  sed -i -E "s|^([[:space:]]*image:).*|\1 $rollback_image|" "$compose_file"
}
if [[ $# == 2 && $1 == --rollback ]]; then
  backup_dir=$(realpath "$2")
  compose_file=$(cat "$backup_dir/compose-path.txt")
  live_dir=$(cat "$backup_dir/data-path.txt")
  [[ -f $backup_dir/beszel_data/data.db && -f $backup_dir/beszel-agent ]]
  [[ ! -e $backup_dir/failed-data ]]
  docker compose -f "$compose_file" stop -t 10 beszel
  systemctl stop beszel-agent.service
  restore_compose
  mv "$live_dir" "$backup_dir/failed-data"
  cp -a "$backup_dir/beszel_data" "$live_dir"
  cp -a "$backup_dir/beszel-agent" /usr/bin/beszel-agent
  docker compose -f "$compose_file" up -d --no-deps --pull never beszel
  systemctl start beszel-agent.service
  echo "Hub data, image configuration, and agent restored from $backup_dir"
  exit
fi
if [[ $# != 4 ]]; then
  echo "Usage: $0 COMPOSE_FILE VALIDATED_IMAGE AGENT_BINARY NEW_BACKUP_DIRECTORY" >&2
  exit 2
fi
compose_file=$(realpath "$1")
candidate_image=$2
candidate_agent=$(realpath "$3")
backup_dir=$(realpath -m "$4")
live_dir=$(dirname "$compose_file")/beszel_data
[[ $candidate_image =~ ^[a-zA-Z0-9._/:@-]+$ ]]
[[ -x $candidate_agent && -f $live_dir/data.db && ! -e $backup_dir ]]
[[ $backup_dir != "$live_dir" && $backup_dir != "$live_dir/"* ]]
[[ $(docker compose -f "$compose_file" config --services) == beszel ]]
docker image inspect "$candidate_image" >/dev/null
"$candidate_agent" -v
mkdir -m 700 "$backup_dir"
cp -a "$compose_file" "$backup_dir/compose.yml"
cp -a /usr/bin/beszel-agent "$backup_dir/beszel-agent"
docker inspect beszel -f '{{.Image}}' > "$backup_dir/image-id.txt"
printf '%s\n' "$compose_file" > "$backup_dir/compose-path.txt"
printf '%s\n' "$live_dir" > "$backup_dir/data-path.txt"
sha256sum "$backup_dir/beszel-agent" > "$backup_dir/agent-sha256.txt"

backed_up=false
finished=false
rollback() {
  result=$?
  if [[ $finished != true ]]; then
    echo "Upgrade failed; restoring the hub, data, and host agent." >&2
    docker compose -f "$compose_file" stop -t 10 beszel || true
    systemctl stop beszel-agent.service || true
    restore_compose
    if [[ $backed_up == true ]]; then
      mv "$live_dir" "$backup_dir/failed-data"
      cp -a "$backup_dir/beszel_data" "$live_dir"
    fi
    cp -a "$backup_dir/beszel-agent" /usr/bin/beszel-agent
    docker compose -f "$compose_file" up -d --no-deps --pull never beszel
    systemctl start beszel-agent.service
    echo "Rollback completed. Backup: $backup_dir" >&2
  fi
  exit "$result"
}
trap rollback EXIT

docker compose -f "$compose_file" stop -t 10 beszel
cp -a "$live_dir" "$backup_dir/beszel_data"
backed_up=true
sed -i -E "s|^([[:space:]]*image:).*|\1 $candidate_image|" "$compose_file"
docker compose -f "$compose_file" config -q
docker compose -f "$compose_file" up -d --no-deps --pull never beszel
[[ $(docker inspect beszel -f '{{.Image}}') == $(docker image inspect "$candidate_image" -f '{{.Id}}') ]]
ready=false
for ((attempt=0; attempt<45; attempt++)); do
  if curl -fsS --max-time 1 http://127.0.0.1:8090/api/health >/dev/null; then
    ready=true
    break
  fi
  sleep 1
done
[[ $ready == true ]]
systemctl stop beszel-agent.service
install -m 755 -o root -g root "$candidate_agent" /usr/bin/beszel-agent
systemctl start beszel-agent.service
systemctl is-active --quiet beszel-agent.service
docker exec beszel /beszel --version
/usr/bin/beszel-agent -v
finished=true
echo "Hub and host agent upgraded. Rollback backup: $backup_dir"
