#!/bin/bash

# K8s Volume Snapshots - 前端构建和Docker打包脚本
# 解决前端缓存和statik生成问题

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 获取项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# 读取版本号
if [ -f "VERSION" ]; then
    VERSION=$(cat VERSION)
else
    VERSION="dev"
fi

log_info "🚀 开始构建 K8s Volume Snapshots v$VERSION"

# 1. 清理前端缓存
log_info "🧹 清理前端缓存..."
cd frontend
rm -rf dist/
rm -rf node_modules/.cache/
rm -rf .cache/

# 2. 重新构建前端
log_info "📦 构建前端应用..."
npm run build

# 检查构建结果
if [ ! -d "dist" ]; then
    log_error "❌ 前端构建失败，dist目录不存在"
    exit 1
fi

cd ..

# 3. 清理旧的statik文件
log_info "🗑️ 清理旧的statik文件..."
rm -f backend/static/statik.go

# 4. 生成新的statik文件
log_info "🔧 生成新的statik文件..."
statik -src=frontend/dist -dest=backend -p=static -f

# 检查statik生成结果
if [ ! -f "backend/static/statik.go" ]; then
    log_error "❌ statik文件生成失败"
    exit 1
fi

STATIK_SIZE=$(wc -c < backend/static/statik.go)
log_success "✅ statik文件生成成功 (大小: ${STATIK_SIZE} bytes)"

# 5. 更新版本号
if [ "$1" = "--bump-version" ]; then
    # 自动递增版本号
    if [[ $VERSION =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        MAJOR=${BASH_REMATCH[1]}
        MINOR=${BASH_REMATCH[2]}
        PATCH=${BASH_REMATCH[3]}
        NEW_PATCH=$((PATCH + 1))
        NEW_VERSION="$MAJOR.$MINOR.$NEW_PATCH"
        echo "$NEW_VERSION" > VERSION
        log_info "📈 版本号更新: $VERSION -> $NEW_VERSION"
        VERSION=$NEW_VERSION
    fi
fi

# 6. 清理Docker环境
log_info "🐳 清理Docker环境..."
docker system prune -f > /dev/null 2>&1 || true

# 7. 构建Docker镜像
log_info "📦 构建Docker镜像 v$VERSION..."
docker build \
    --platform linux/amd64 \
    --no-cache \
    -t k8s-volume-snapshots:$VERSION \
    -t reg.deeproute.ai/deeproute-public/k8s-volume-snapshots:$VERSION-amd64 \
    .

if [ $? -eq 0 ]; then
    log_success "✅ Docker镜像构建成功"
else
    log_error "❌ Docker镜像构建失败"
    exit 1
fi

# 8. 推送镜像（可选）
if [ "$2" = "--push" ]; then
    log_info "📤 推送Docker镜像..."
    docker push reg.deeproute.ai/deeproute-public/k8s-volume-snapshots:$VERSION-amd64
    
    if [ $? -eq 0 ]; then
        log_success "✅ Docker镜像推送成功"
    else
        log_error "❌ Docker镜像推送失败"
        exit 1
    fi
fi

# 9. 更新Kubernetes部署文件
log_info "⚙️ 更新Kubernetes部署文件..."
sed -i.bak "s|image: reg.deeproute.ai/deeproute-public/k8s-volume-snapshots:.*-amd64|image: reg.deeproute.ai/deeproute-public/k8s-volume-snapshots:$VERSION-amd64|g" k8s/deployment.yaml
rm -f k8s/deployment.yaml.bak

# 10. 部署到Kubernetes（可选）
if [ "$3" = "--deploy" ]; then
    log_info "🚀 部署到Kubernetes..."
    kubectl apply -f k8s/deployment.yaml
    
    if [ $? -eq 0 ]; then
        log_success "✅ 部署到Kubernetes成功"
        
        # 等待Pod就绪
        log_info "⏳ 等待Pod就绪..."
        kubectl wait --for=condition=ready pod -l app=k8s-volume-snapshots --timeout=300s
        
        # 显示部署状态
        kubectl get pods -l app=k8s-volume-snapshots
        kubectl describe pod -l app=k8s-volume-snapshots | grep Image | head -1
    else
        log_error "❌ 部署到Kubernetes失败"
        exit 1
    fi
fi

# 11. 验证应用健康状态（可选）
if [ "$4" = "--verify" ]; then
    log_info "🔍 验证应用健康状态..."
    sleep 10
    
    HEALTH_CHECK=$(curl -s http://10.9.9.110:8081/health | jq -r '.status' 2>/dev/null || echo "error")
    if [ "$HEALTH_CHECK" = "ok" ]; then
        log_success "✅ 应用健康检查通过"
    else
        log_warning "⚠️ 应用健康检查失败，请手动验证"
    fi
fi

log_success "🎉 构建流程完成！"
log_info "📋 版本信息: v$VERSION"
log_info "🔗 访问地址: http://10.9.9.110:8081"

echo
echo "📚 使用方法:"
echo "  ./scripts/build-frontend-docker.sh                     # 基础构建"
echo "  ./scripts/build-frontend-docker.sh --bump-version      # 自动递增版本"
echo "  ./scripts/build-frontend-docker.sh --bump-version --push --deploy --verify  # 完整流程"