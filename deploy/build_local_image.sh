#!/usr/bin/env bash
# 本地构建带 frontend-local 定制前端的镜像。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
IMAGE_NAME="${IMAGE_NAME:-sub2api-local:latest}"
NODE_IMAGE="${NODE_IMAGE:-docker.1ms.run/library/node:24-alpine}"
GOLANG_IMAGE="${GOLANG_IMAGE:-docker.1ms.run/library/golang:1.26.3-alpine}"
ALPINE_IMAGE="${ALPINE_IMAGE:-docker.1ms.run/library/alpine:3.21}"
POSTGRES_IMAGE="${POSTGRES_IMAGE:-docker.1ms.run/library/postgres:18-alpine}"

docker build -t "${IMAGE_NAME}" \
    --build-arg GOPROXY=https://goproxy.cn,direct \
    --build-arg GOSUMDB=sum.golang.google.cn \
    --build-arg NODE_IMAGE="${NODE_IMAGE}" \
    --build-arg GOLANG_IMAGE="${GOLANG_IMAGE}" \
    --build-arg ALPINE_IMAGE="${ALPINE_IMAGE}" \
    --build-arg POSTGRES_IMAGE="${POSTGRES_IMAGE}" \
    -f "${REPO_ROOT}/Dockerfile.frontend-local" \
    "${REPO_ROOT}"
