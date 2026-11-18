#!/bin/sh
# Create .env from .env.example for this checkout (mise run env).
# Each git worktree gets a slot number. The slot gives its own compose project
# name and its own host ports, so parallel worktrees do not conflict.
# The main checkout has slot 0 and keeps the default ports.
# The slot list is in the shared git folder, so all worktrees see it.
set -eu

test -f .env && exit 0

git_dir=$(git rev-parse --absolute-git-dir)
common_dir=$(cd "$(git rev-parse --git-common-dir)" && pwd -P)
here=$(pwd -P)

if [ "$git_dir" = "$common_dir" ]; then
  slot=0
else
  slots="$common_dir/app-worktree-slots"
  lock="$common_dir/app-worktree-slots.lock"
  # mkdir is atomic, so two worktrees cannot take the same slot.
  tries=0
  until mkdir "$lock" 2>/dev/null; do
    tries=$((tries + 1))
    if [ "$tries" -gt 50 ]; then
      echo "Cannot lock $lock. Remove it if no other 'mise run env' runs." >&2
      exit 1
    fi
    sleep 0.1
  done
  trap 'rmdir "$lock"' EXIT
  touch "$slots"

  slot=$(awk -F '\t' -v p="$here" '$2 == p { print $1; exit }' "$slots")
  if [ -z "$slot" ]; then
    # Take the first free slot. A slot is free when its folder does not exist.
    n=1
    while [ "$n" -le 99 ]; do
      path=$(awk -F '\t' -v s="$n" '$1 == s { print $2; exit }' "$slots")
      if [ -z "$path" ] || [ ! -d "$path" ]; then
        slot=$n
        break
      fi
      n=$((n + 1))
    done
    if [ -z "$slot" ]; then
      echo "All 99 worktree slots are in use. Clean $slots." >&2
      exit 1
    fi
    awk -F '\t' -v s="$slot" '$1 != s' "$slots" > "$slots.tmp"
    printf '%s\t%s\n' "$slot" "$here" >> "$slots.tmp"
    mv "$slots.tmp" "$slots"
  fi
fi

name=$(basename "$here" | tr '[:upper:]' '[:lower:]' | tr -c 'a-z0-9_-\n' '-' | sed 's/-*$//')
http=$((8088 + slot))
pg=$((54320 + slot))

sed \
  -e "s|^COMPOSE_PROJECT_NAME=.*|COMPOSE_PROJECT_NAME=$name|" \
  -e "s|^WORKTREE_SLOT=.*|WORKTREE_SLOT=$slot|" \
  -e "s|^HTTP_PORT=.*|HTTP_PORT=$http|" \
  -e "s|^POSTGRES_PORT=.*|POSTGRES_PORT=$pg|" \
  -e "s|^S3_PORT=.*|S3_PORT=$((8333 + slot))|" \
  -e "s|^SMTP_PORT=.*|SMTP_PORT=$((1025 + slot))|" \
  -e "s|^MOCK_PORT=.*|MOCK_PORT=$((4010 + slot))|" \
  -e "s|^DASHBOARD_PORT=.*|DASHBOARD_PORT=$((5173 + slot))|" \
  -e "s|^SITE_PORT=.*|SITE_PORT=$((4330 + slot))|" \
  -e "s|^CORE_URL=.*|CORE_URL=http://api.localhost:$http|" \
  -e "s|@localhost:54320/|@localhost:$pg/|" \
  .env.example > .env

echo "Created .env: slot $slot, project $name, core on http://api.localhost:$http"
