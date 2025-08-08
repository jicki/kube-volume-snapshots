#!/bin/bash

echo "🚀 启动后端服务..."

cd backend

# 检查依赖
if [ ! -f "go.mod" ]; then
    echo "❌ 未找到 go.mod，请先运行 ./scripts/init.sh"
    exit 1
fi

# 检查 Kubernetes 配置
if [ ! -f "$HOME/.kube/config" ]; then
    echo "⚠️  警告：未找到 ~/.kube/config，请确保 Kubernetes 配置正确"
fi

echo "🔧 构建并启动后端服务..."
go run main.go