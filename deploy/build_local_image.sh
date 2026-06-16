#!/usr/bin/env bash
# 本地构建带 frontend-local 定制前端的镜像。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
IMAGE_NAME="${IMAGE_NAME:-sub2api-local:latest}"
PUSH="${PUSH:-false}"
PLATFORM="${PLATFORM:-linux/amd64}"
CLEAN_AFTER_PUSH="${CLEAN_AFTER_PUSH:-true}"
REMOVE_BUILDER_AFTER_PUSH="${REMOVE_BUILDER_AFTER_PUSH:-false}"

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

cleanup_after_push() {
    if [[ "${CLEAN_AFTER_PUSH}" != "true" ]]; then
        return
    fi

    echo "Push 成功，开始清理本地构建产物..."

    if docker image inspect "${IMAGE_NAME}" >/dev/null 2>&1; then
        docker image rm "${IMAGE_NAME}" >/dev/null
        echo "已删除本地镜像：${IMAGE_NAME}"
    else
        echo "未发现本地镜像标签：${IMAGE_NAME}"
    fi

    docker buildx prune --builder "${BUILDER_NAME}" --all --force >/dev/null
    echo "已清理 buildx 缓存：${BUILDER_NAME}"

    if [[ "${REMOVE_BUILDER_AFTER_PUSH}" == "true" ]]; then
        docker buildx rm "${BUILDER_NAME}" >/dev/null
        echo "已删除 buildx builder：${BUILDER_NAME}"
    fi
}

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

    cleanup_after_push
else
    docker build "${BUILD_ARGS[@]}" .
fi
