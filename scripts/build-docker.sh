#!/bin/bash

set -e

# 清理旧的statik文件
echo "🗑️ 清理旧的statik文件..."
rm -f backend/static/statik.go

# 获取版本号
VERSION_FILE="VERSION"
if [ -f "$VERSION_FILE" ]; then
    VERSION=$(cat $VERSION_FILE)
else
    VERSION="latest"
fi

# 支持命令行参数覆盖版本
if [ ! -z "$1" ]; then
    VERSION="$1"
fi

echo "🐳 构建 Docker 镜像 (版本: $VERSION)..."

# 检查 Docker 是否可用
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装或不在 PATH 中"
    exit 1
fi

# 检查 Docker 守护进程是否运行
if ! docker ps &> /dev/null; then
    echo "❌ Docker 守护进程未运行，请启动 Docker"
    exit 1
fi

# 显示 Docker 版本信息
echo "📋 Docker 版本信息："
docker --version

# 构建镜像 (明确指定 amd64 架构)
echo "🏗️ 开始构建镜像 k8s-volume-snapshots:$VERSION (amd64 架构)..."
docker build --platform linux/amd64 -t k8s-volume-snapshots:$VERSION .

# 同时打上 latest 标签
docker tag k8s-volume-snapshots:$VERSION k8s-volume-snapshots:latest

# 打标签并推送到镜像仓库 (添加架构后缀)
REGISTRY_IMAGE="reg.deeproute.ai/deeproute-public/k8s-volume-snapshots:$VERSION-amd64"
docker tag k8s-volume-snapshots:$VERSION $REGISTRY_IMAGE

echo "📤 推送镜像到仓库..."
docker push $REGISTRY_IMAGE

# 如果版本不是 latest，也推送 latest 标签 (amd64架构)
if [ "$VERSION" != "latest" ]; then
    REGISTRY_LATEST="reg.deeproute.ai/deeproute-public/k8s-volume-snapshots:latest-amd64"
    docker tag k8s-volume-snapshots:$VERSION $REGISTRY_LATEST
    docker push $REGISTRY_LATEST
fi

# 显示构建结果
echo "📊 镜像构建完成！"
docker images | grep k8s-volume-snapshots

echo "✅ Docker 镜像构建和推送成功！"
echo "🏷️  本地镜像: k8s-volume-snapshots:$VERSION"
echo "🌐 远程镜像: $REGISTRY_IMAGE"