#!/bin/bash

set -e

echo "🏗️ 构建前端并生成 statik 文件..."

# 检查必要工具
if ! command -v npm &> /dev/null; then
    echo "❌ npm 未安装"
    exit 1
fi

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

# 清理旧的statik文件
echo "🗑️ 清理旧的statik文件..."
rm -f backend/static/statik.go

# 安装 statik（如果未安装）
if ! command -v statik &> /dev/null; then
    echo "📥 安装 statik..."
    go install github.com/rakyll/statik@latest
fi

# 生成 statik 文件
echo "🔄 生成 statik 文件..."
statik -src=frontend/dist -dest=backend -p=static -f

echo "✅ statik 文件生成完成！"
echo "📁 生成的文件: backend/static/statik.go"