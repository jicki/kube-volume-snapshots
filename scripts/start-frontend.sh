#!/bin/bash

echo "🚀 启动前端服务..."

cd frontend

# 检查依赖
if [ ! -d "node_modules" ]; then
    echo "❌ 未找到 node_modules，请先运行 ./scripts/init.sh"
    exit 1
fi

echo "🔧 启动前端开发服务器..."
npm run serve