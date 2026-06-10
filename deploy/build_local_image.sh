#!/usr/bin/env bash
# 本地构建带 frontend-local 定制前端的镜像。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
IMAGE_NAME="${IMAGE_NAME:-sub2api-local:latest}"
PUSH="${PUSH:-false}"
PLATFORM="${PLATFORM:-linux/amd64}"

# 国内基础镜像源
NODE_IMAGE="${NODE_IMAGE:-docker.m.daocloud.io/library/node:24-alpine}"
GOLANG_IMAGE="${GOLANG_IMAGE:-docker.m.daocloud.io/library/golang:1.26.4-alpine}"
ALPINE_IMAGE="${ALPINE_IMAGE:-docker.m.daocloud.io/library/alpine:3.21}"
POSTGRES_IMAGE="${POSTGRES_IMAGE:-docker.m.daocloud.io/library/postgres:18-alpine}"

cd "${REPO_ROOT}"

BUILD_ARGS=(
    -t "${IMAGE_NAME}"
    --build-arg "GOPROXY=https://goproxy.cn,direct"
    --build-arg "GOSUMDB=sum.golang.google.cn"
    --build-arg "NODE_IMAGE=${NODE_IMAGE}"
    --build-arg "GOLANG_IMAGE=${GOLANG_IMAGE}"
    --build-arg "ALPINE_IMAGE=${ALPINE_IMAGE}"
    --build-arg "POSTGRES_IMAGE=${POSTGRES_IMAGE}"
    -f "${REPO_ROOT}/Dockerfile.frontend-local"
)

if [[ "${PUSH}" == "true" ]]; then
    if [[ "${IMAGE_NAME}" != ghcr.io/* ]]; then
        echo "PUSH=true 时 IMAGE_NAME 必须以 ghcr.io/ 开头，当前为：${IMAGE_NAME}" >&2
        exit 1
    fi

    BUILDER_NAME="${BUILDER_NAME:-sub2api-builder}"
    if ! docker buildx inspect "${BUILDER_NAME}" >/dev/null 2>&1; then
        docker buildx create --name "${BUILDER_NAME}" \
            --driver docker-container \
            --driver-opt image=docker.m.daocloud.io/moby/buildkit:buildx-stable-1 \
            --use
    else
        docker buildx use "${BUILDER_NAME}"
    fi

    docker buildx build \
        "${BUILD_ARGS[@]}" \
        --platform "${PLATFORM}" \
        --push \
        .
else
    docker build "${BUILD_ARGS[@]}" .
fi
