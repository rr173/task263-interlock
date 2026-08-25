#!/usr/bin/env bash
# 构建评测镜像：usage: bash build_benzhi_docker.sh <镜像名> <平台>
# 例：bash build_benzhi_docker.sh my-project linux/arm64
set -euo pipefail

IMAGE_NAME="${1:-task263-interlock}"
PLATFORM="${2:-linux/amd64}"

docker buildx build \
  --platform "${PLATFORM}" \
  -f benzhi.Dockerfile \
  -t "${IMAGE_NAME}" \
  .
