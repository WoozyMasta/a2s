#!/usr/bin/env bash

# Gets all response pages from IStoreService/GetAppList/v1
# with a list of all game IDs (excluding DLC, apps, etc.)
# and generates a summary result as an apps.json file
#
# Require curl and jq tools
#
# Usage:
#   ./get.sh [api key] [out file]
# or
#   STEAM_KEY=abc OUT_FILE=apps.json ./get.sh

set -euo pipefail

: "${STEAM_KEY:=${1:?}}"
: "${PER_PAGE:=10000}" # 50k limit, 10k default
: "${OUT_FILE:=${2:-apps.json}}"

get() {
  local last=0

  while :; do
    response="$(
      curl -fG# "https://api.steampowered.com/IStoreService/GetAppList/v1/" \
        -d key="$STEAM_KEY" \
        -d include_games=true \
        -d include_dlc=false \
        -d include_software=false \
        -d include_videos=false \
        -d include_hardware=false \
        -d max_results="$PER_PAGE" \
        -d last_appid="$last" | jq -cer '.'
    )"

    ! jq -er '.response.have_more_results' &>/dev/null <<<"$response" && break
    last="$(jq -er '.response.last_appid' <<<"$response")"

    jq -cer '.response.apps' <<<"$response"
  done
}

get | jq -ers 'add | {apps: .}' > "$OUT_FILE"
