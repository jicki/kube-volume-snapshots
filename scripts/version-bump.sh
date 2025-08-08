#!/bin/bash

set -e

VERSION_FILE="VERSION"

# 检查 VERSION 文件是否存在
if [ ! -f "$VERSION_FILE" ]; then
    echo "❌ VERSION 文件不存在"
    exit 1
fi

# 读取当前版本
CURRENT_VERSION=$(cat $VERSION_FILE)
echo "📋 当前版本: $CURRENT_VERSION"

# 解析版本号
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# 检查版本号格式
if [[ ! "$MAJOR" =~ ^[0-9]+$ ]] || [[ ! "$MINOR" =~ ^[0-9]+$ ]] || [[ ! "$PATCH" =~ ^[0-9]+$ ]]; then
    echo "❌ 版本号格式错误: $CURRENT_VERSION （应为 x.y.z 格式）"
    exit 1
fi

# 获取升级类型
BUMP_TYPE=${1:-patch}

case $BUMP_TYPE in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "❌ 无效的升级类型: $BUMP_TYPE （支持: major, minor, patch）"
        exit 1
        ;;
esac

# 生成新版本号
NEW_VERSION="$MAJOR.$MINOR.$PATCH"

# 写入新版本号
echo "$NEW_VERSION" > $VERSION_FILE

echo "✅ 版本已升级: $CURRENT_VERSION → $NEW_VERSION"
echo "🏷️  升级类型: $BUMP_TYPE"