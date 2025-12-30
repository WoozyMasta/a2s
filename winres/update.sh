#!/bin/sh
set -eu

: "${VERSION=${1:?}}"
: "${INPUT_FILE:=winres/winres.json}"
: "${OUTPUT_FILE=winres/winres.build.json}"

[ ! -f "$INPUT_FILE" ] && { >&2 echo "Error: $INPUT_FILE not found"; exit 1; }

if command -v jq >/dev/null 2>&1; then
  jq --arg v "$VERSION" '
    .RT_VERSION["#1"]["0000"].fixed.file_version = $v |
    .RT_VERSION["#1"]["0000"].fixed.product_version = $v |
    .RT_VERSION["#1"]["0000"].info["0409"].FileVersion = $v |
    .RT_VERSION["#1"]["0000"].info["0409"].ProductVersion = $v |
    .RT_MANIFEST["#1"]["0409"].identity.version = $v
  ' "$INPUT_FILE" > "$OUTPUT_FILE"

  if [ ! -s "$OUTPUT_FILE" ]; then
    >&2 echo "Warning: jq produced empty file, using original"
    rm -f "$OUTPUT_FILE"
    cp "$INPUT_FILE" "$OUTPUT_FILE"
  fi
else
  >&2 echo "Warning: jq not found, using original"
  cp "$INPUT_FILE" "$OUTPUT_FILE"
fi
