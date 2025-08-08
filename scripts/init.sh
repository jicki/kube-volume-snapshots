#!/bin/bash

set -e

echo "🚀 初始化 K8s Volume Snapshots Manager 项目..."

# 检查必要的工具
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ $1 未安装，请先安装 $1"
        exit 1
    fi
}

echo "📋 检查必要工具..."
check_command go
check_command node
check_command npm

# 初始化后端
echo "🔧 初始化后端项目..."
cd backend
echo "📦 下载 Go 依赖..."
go mod tidy
go mod download

echo "✅ 后端依赖安装完成"

# 初始化前端
echo "🔧 初始化前端项目..."
cd ../frontend
echo "📦 安装 npm 依赖..."
npm install

echo "✅ 前端依赖安装完成"

cd ..

echo "🎉 项目初始化完成！"
echo ""
echo "启动方式："
echo "1. 启动后端：cd backend && go run main.go"
echo "2. 启动前端：cd frontend && npm run serve"
echo ""
echo "访问地址："
echo "- 前端：http://localhost:8080"
echo "- 后端 API：http://localhost:8081"