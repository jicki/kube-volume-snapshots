#!/bin/bash

set -e

echo "🏗️ 构建完整项目（包含前端静态资源）..."

# 检查必要工具
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

# 构建前端
echo "📦 构建前端项目..."
cd frontend
npm run build

# 检查构建结果
if [ ! -d "dist" ]; then
    echo "❌ 前端构建失败，未找到 dist 目录"
    exit 1
fi

cd ..

# 安装 statik（如果未安装）
if ! command -v statik &> /dev/null; then
    echo "📥 安装 statik..."
    go install github.com/rakyll/statik@latest
fi

# 生成 statik 文件
echo "🔄 生成 statik 文件..."
statik -src=frontend/dist -dest=backend -p=static -f

# 构建后端
echo "🔧 构建后端项目..."
cd backend
go build -o ../bin/k8s-volume-snapshots main.go

cd ..

echo "✅ 项目构建完成！"
echo "📁 可执行文件: bin/k8s-volume-snapshots"
echo "📁 静态资源文件: backend/static/statik.go"