#!/bin/bash

set -e

echo "🧪 测试构建流程..."

# 清理之前的构建
echo "🧹 清理之前的构建文件..."
make clean

# 初始化项目
echo "📋 初始化项目..."
make init

# 构建单一可执行文件
echo "🏗️ 构建单一可执行文件..."
make build-all

# 检查构建结果
if [ -f "bin/k8s-volume-snapshots" ]; then
    echo "✅ 单一可执行文件构建成功"
    ls -lh bin/k8s-volume-snapshots
else
    echo "❌ 单一可执行文件构建失败"
    exit 1
fi

# 检查静态资源文件
if [ -f "backend/static/statik.go" ]; then
    echo "✅ 静态资源文件生成成功"
    ls -lh backend/static/statik.go
else
    echo "❌ 静态资源文件生成失败"
    exit 1
fi

echo "🎉 构建测试完成！"
echo ""
echo "运行命令："
echo "./bin/k8s-volume-snapshots"