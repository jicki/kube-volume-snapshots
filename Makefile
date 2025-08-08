.PHONY: init start-backend start-frontend start build build-all build-statik build-image build-version version-bump clean help

# 默认目标
help:
	@echo "可用的命令："
	@echo "  init          - 初始化项目依赖"
	@echo "  start-backend - 启动后端服务"
	@echo "  start-frontend- 启动前端服务"
	@echo "  start         - 同时启动前后端服务"
	@echo "  build         - 构建项目（分离模式）"
	@echo "  build-all     - 构建单一可执行文件（包含前端）"
	@echo "  build-statik  - 构建前端并生成 statik 文件"
	@echo "  build-image   - 构建 Docker 镜像（使用 VERSION 文件版本）"
	@echo "  build-version - 构建指定版本镜像（make build-version VERSION=x.x.x）"
	@echo "  version-bump  - 升级版本号（make version-bump TYPE=patch|minor|major）"
	@echo "  clean         - 清理构建文件"
	@echo ""
	@echo "版本管理："
	@if [ -f "VERSION" ]; then \
		echo "  当前版本: $$(cat VERSION)"; \
	else \
		echo "  当前版本: 未设置"; \
	fi

# 初始化项目
init:
	@chmod +x scripts/*.sh
	@./scripts/init.sh

# 启动后端
start-backend:
	@chmod +x scripts/start-backend.sh
	@./scripts/start-backend.sh

# 启动前端
start-frontend:
	@chmod +x scripts/start-frontend.sh
	@./scripts/start-frontend.sh

# 同时启动前后端（需要在不同终端中运行）
start:
	@echo "请在不同的终端窗口中运行："
	@echo "终端1: make start-backend"
	@echo "终端2: make start-frontend"

# 构建项目（分离模式）
build:
	@echo "构建后端（不包含 Ceph 支持）..."
	@cd backend && go build -o ../bin/k8s-volume-snapshots main.go
	@echo "构建前端..."
	@cd frontend && npm run build
	@echo "构建完成！"

# 构建项目（包含 Ceph 支持）- 需要 librados
build-with-ceph:
	@echo "构建后端（包含 Ceph 支持）..."
	@cd backend && go build -tags ceph -o ../bin/k8s-volume-snapshots main.go
	@echo "构建前端..."
	@cd frontend && npm run build
	@echo "构建完成！"

# 构建单一可执行文件（包含前端）
build-all:
	@chmod +x scripts/build-all.sh
	@./scripts/build-all.sh

# 构建前端并生成 statik 文件
build-statik:
	@chmod +x scripts/build-statik.sh
	@./scripts/build-statik.sh

# 构建 Docker 镜像
build-image:
	@chmod +x scripts/build-docker.sh
	@./scripts/build-docker.sh

# 构建指定版本的 Docker 镜像
build-version:
	@chmod +x scripts/build-docker.sh
	@if [ -z "$(VERSION)" ]; then \
		echo "请指定版本号: make build-version VERSION=x.x.x"; \
		exit 1; \
	fi
	@./scripts/build-docker.sh $(VERSION)

# 版本升级
version-bump:
	@chmod +x scripts/version-bump.sh
	@./scripts/version-bump.sh $(TYPE)

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf bin/
	@rm -rf frontend/dist/
	@rm -rf backend/static/
	@echo "清理完成！"