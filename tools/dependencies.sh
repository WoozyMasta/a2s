#!/usr/bin/env bash
set -eu

: "${UPX_VERSION:=4.2.4}"

if ! command -v curl &>/dev/null || ! command -v xz &>/dev/null ; then
  sudo apt update
  sudo apt install -y xz-utils curl
fi

if ! command -v upx &>/dev/null; then
  curl -#Lo upx.tar.xz \
    "https://github.com/upx/upx/releases/download/v$UPX_VERSION/upx-$UPX_VERSION-amd64_linux.tar.xz"
  tar -xvf upx.tar.xz --strip-components=1 "upx-$UPX_VERSION-amd64_linux/upx"
  chmod +x upx
  sudo mv upx /usr/local/bin/
fi

command -v go-winres &>/dev/null ||
  go install github.com/tc-hib/go-winres@latest

command -v cyclonedx-gomod &>/dev/null ||
  go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest
