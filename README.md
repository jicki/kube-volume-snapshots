# K8s Volume Snapshots Manager

[![Version](https://img.shields.io/badge/version-2.1.26-blue.svg)](VERSION)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js&logoColor=white)](https://vuejs.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

基于 Ceph CSI RBD 的 Kubernetes Volume Snapshot 管理系统，支持多集群管理，提供完整的快照创建、管理、定时任务和集群监控功能。

## 🚀 功能特性

### 核心功能
- 🗂️ **VolumeSnapshotClass 管理** - 展示集群中的快照类及关联的存储类
- 📸 **VolumeSnapshot 管理** - 创建、查看、删除卷快照，支持多命名空间
- 📋 **VolumeSnapshotContent 查看** - 展示快照内容详细信息
- 💾 **PVC 管理** - 持久卷声明的查看和管理
- ⏰ **定时快照任务** - 基于 Cron 表达式的自动快照创建和管理
- 🔀 **多集群管理** - 支持同时管理多个 Kubernetes 集群，动态切换集群

### 集群管理
- 🔗 **多集群支持** - 支持配置和管理多个 Kubernetes 集群
- 🔄 **集群切换** - 动态切换不同集群进行管理操作
- 📊 **集群状态监控** - 实时监控所有集群的连接状态和健康信息
- 🏗️ **Ceph 集群监控** - 实时监控 Ceph 集群状态和健康信息
- 📊 **仪表板** - 集群快照资源概览和统计信息
- 👥 **用户管理** - 用户账户创建、权限管理和认证
- 🔐 **登录认证** - 安全的用户身份验证系统

### 技术架构
- **后端**: Golang + Gin + Kubernetes Go Client + Ceph Go SDK
- **前端**: Vue 3 + Element Plus + Composition API  
- **打包**: 使用 Statik 将前端打包为 Go 静态资源
- **部署**: 单一可执行文件，支持容器化部署
- **存储**: 支持 Ceph CSI RBD 和其他 CSI 驱动
- **认证**: JWT Token 认证机制

## 📁 项目结构

```
k8s-volume-snapshots/
├── VERSION                          # 版本信息
├── Makefile                         # 构建和部署脚本
├── Dockerfile                       # 容器镜像构建文件
├── README.md                        # 项目文档
├── backend/                         # Golang 后端
│   ├── main.go                     # 主程序入口
│   ├── go.mod                      # Go 模块定义
│   ├── go.sum                      # 依赖版本锁定
│   ├── controllers/                # 控制器层 - API 接口
│   │   ├── snapshot_controller.go  # 快照管理 API
│   │   ├── scheduled_controller.go # 定时任务 API
│   │   ├── user_controller.go      # 用户管理 API
│   │   └── ceph_controller.go      # Ceph 集群 API
│   ├── services/                   # 服务层 - 业务逻辑
│   │   ├── k8s_service.go         # Kubernetes 操作服务
│   │   ├── user_service.go        # 用户管理服务
│   │   ├── ceph_service.go        # Ceph 集群服务
│   │   └── ceph_service_stub.go   # Ceph 服务存根
│   ├── models/                     # 数据模型
│   │   ├── snapshot.go            # 快照相关数据结构
│   │   └── ceph.go                # Ceph 相关数据结构
│   ├── middleware/                 # 中间件
│   │   └── auth.go                # 认证中间件
│   └── static/                     # 静态资源（自动生成）
├── frontend/                       # Vue 前端
│   ├── package.json               # Node.js 依赖配置
│   ├── package-lock.json          # 依赖版本锁定
│   ├── vue.config.js              # Vue 构建配置
│   ├── public/
│   │   └── index.html             # HTML 模板
│   └── src/
│       ├── main.js                # 应用入口
│       ├── App.vue                # 根组件
│       ├── router/                # 路由配置
│       │   └── index.js
│       ├── stores/                # 状态管理
│       │   └── auth.js            # 认证状态管理
│       ├── api/                   # API 接口封装
│       │   ├── index.js           # API 基础配置
│       │   └── ceph.js            # Ceph API 接口
│       └── views/                 # 页面组件
│           ├── Login.vue          # 登录页面
│           ├── Dashboard.vue      # 仪表板
│           ├── SnapshotClasses.vue # 快照类管理
│           ├── Snapshots.vue      # 快照管理
│           ├── PVCs.vue           # PVC 管理
│           ├── ScheduledTasks.vue # 定时任务管理
│           ├── CephCluster.vue    # Ceph 集群监控
│           └── UserManagement.vue # 用户管理
├── scripts/                        # 构建和部署脚本
│   ├── init.sh                    # 项目初始化
│   ├── build-all.sh               # 构建单一可执行文件
│   ├── build-statik.sh            # 构建静态资源
│   ├── build-docker.sh            # 构建 Docker 镜像
│   ├── build-frontend-docker.sh   # 构建前端镜像（已弃用）
│   ├── start-backend.sh           # 启动后端服务
│   ├── start-frontend.sh          # 启动前端服务
│   ├── test-build.sh              # 测试构建
│   └── version-bump.sh            # 版本升级
├── docker/                         # Docker 相关配置
│   ├── docker-compose.yml         # Docker Compose 配置
│   ├── Dockerfile.backend         # 后端镜像（已弃用）
│   ├── Dockerfile.frontend        # 前端镜像（已弃用）
│   └── nginx.conf                 # Nginx 配置（已弃用）
├── k8s/                           # Kubernetes 部署配置
│   └── statefulset.yaml          # StatefulSet 部署文件
├── config/                        # 配置文件
│   └── rbac.yaml                 # RBAC 权限配置
├── docs/                          # 项目文档
└── bin/                           # 构建输出目录
```

## 🚀 快速开始

### 环境要求
- Go 1.21+
- Node.js 14+
- Kubernetes 集群（支持 VolumeSnapshot）
- 配置好的 kubeconfig
- （可选）Ceph 集群和相关客户端库

### ⚡ 一键启动

项目采用单一镜像架构，支持多种启动方式：

#### 方式 1：单一可执行文件（推荐）
```bash
# 初始化并构建包含前端的单一可执行文件
make init
make build-all

# 直接运行
./bin/k8s-volume-snapshots
```

#### 方式 2：开发模式
```bash
# 初始化项目依赖
make init

# 启动后端服务（在终端1中运行）
make start-backend

# 启动前端服务（在终端2中运行）
make start-frontend
```

#### 方式 3：Docker 镜像
```bash
# 构建 Docker 镜像
make build-image

# 使用 Docker Compose 运行
cd docker && docker-compose up -d
```

### 🔧 手动初始化

**后端依赖**
```bash
cd backend
go mod tidy
go mod download
```

**前端依赖**
```bash
cd frontend
npm install
```

### 🌐 访问应用

**单一端口访问**：
- 应用界面：http://localhost:8081
- API 接口：http://localhost:8081/api/*
- 静态资源：http://localhost:8081/static/*

**开发模式**：
- 前端：http://localhost:8080
- 后端：http://localhost:8081

## 🔐 权限配置

确保运行程序的账户具有以下 Kubernetes 权限：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: volume-snapshot-manager
rules:
# 访问 PVC、PV 和命名空间
- apiGroups: [""]
  resources: ["persistentvolumeclaims", "persistentvolumes", "namespaces"]
  verbs: ["get", "list"]
# 访问存储类
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
  verbs: ["get", "list"]
# 访问快照相关资源
- apiGroups: ["snapshot.storage.k8s.io"]
  resources: ["volumesnapshotclasses", "volumesnapshots", "volumesnapshotcontents"]
  verbs: ["get", "list", "create", "delete", "update"]
```

应用 RBAC 配置：
```bash
kubectl apply -f config/rbac.yaml
```

## 🔀 多集群配置

### 配置文件

系统支持通过配置文件管理多个 Kubernetes 集群。创建配置文件 `config/clusters.yaml`：

```yaml
---
# 多集群配置文件
# Multi-cluster configuration for k8s-volume-snapshots

# 默认集群（如果未指定集群，将使用此集群）
default_cluster: "production"

# 集群配置列表
clusters:
  # 生产环境集群
  - name: "production"
    display_name: "生产环境"
    description: "主生产环境 Kubernetes 集群"
    kubeconfig: "/root/.kube/config-prod"
    enabled: true
    
  # 测试环境集群  
  - name: "testing"
    display_name: "测试环境"
    description: "测试环境 Kubernetes 集群"
    kubeconfig: "/root/.kube/config-test"
    enabled: true
    
  # 开发环境集群
  - name: "development"
    display_name: "开发环境" 
    description: "开发环境 Kubernetes 集群"
    kubeconfig: "/root/.kube/config-dev"
    enabled: true

# 全局配置
global:
  # 连接超时时间（秒）
  timeout: 30
  # 请求QPS限制
  qps: 50
  # 突发请求限制
  burst: 100
  # 缓存刷新间隔（分钟）
  cache_refresh_interval: 5
```

### 配置文件位置

系统按以下优先级查找配置文件：
1. 环境变量 `MULTI_CLUSTER_CONFIG` 指定的路径
2. 当前目录下的 `config/clusters.yaml`
3. 默认路径 `/etc/k8s-volume-snapshots/clusters.yaml`

### 集群配置选项

每个集群可以配置以下选项：

- **name**: 集群唯一标识符
- **display_name**: 在界面上显示的集群名称
- **description**: 集群描述信息
- **kubeconfig**: kubeconfig 文件路径（可选）
- **enabled**: 是否启用该集群
- **server**: API Server 地址（可选，直接连接）
- **token**: 访问令牌（可选，直接连接）
- **certificate_authority_data**: CA 证书数据（可选）

### 单集群兼容模式

如果没有提供多集群配置文件，系统将自动回退到单集群模式，使用以下配置：
- 集群内部署：使用 Service Account
- 集群外部署：使用默认的 kubeconfig 文件（`~/.kube/config`）

## 📚 API 接口文档

### 认证接口
- `POST /api/login` - 用户登录
- `POST /api/logout` - 用户登出
- `GET /api/profile` - 获取用户信息

### 用户管理
- `GET /api/users` - 获取用户列表
- `POST /api/users` - 创建新用户
- `PUT /api/users/:id` - 更新用户信息
- `DELETE /api/users/:id` - 删除用户

### VolumeSnapshotClass
- `GET /api/volumesnapshotclasses` - 获取快照类列表

### VolumeSnapshot
- `GET /api/volumesnapshots?namespace=<ns>` - 获取快照列表
- `POST /api/volumesnapshots` - 创建快照
- `DELETE /api/volumesnapshots/<namespace>/<name>` - 删除快照

### VolumeSnapshotContent
- `GET /api/volumesnapshotcontents/<name>` - 获取快照内容

### PVC 管理
- `GET /api/pvcs?namespace=<ns>` - 获取 PVC 列表

### 定时任务
- `GET /api/scheduled-snapshots` - 获取定时任务列表
- `POST /api/scheduled-snapshots` - 创建定时任务
- `PUT /api/scheduled-snapshots/<id>` - 更新定时任务
- `DELETE /api/scheduled-snapshots/<id>` - 删除定时任务
- `POST /api/scheduled-snapshots/<id>/toggle` - 启用/禁用定时任务

### Ceph 集群
- `GET /api/ceph/status` - 获取 Ceph 集群状态
- `GET /api/ceph/health` - 获取 Ceph 集群健康信息
- `GET /api/ceph/pools` - 获取存储池列表
- `GET /api/ceph/osds` - 获取 OSD 信息

### 仪表板
- `GET /api/dashboard` - 获取仪表板统计信息

### 多集群管理
- `GET /api/clusters` - 获取所有集群信息和状态
- `GET /api/clusters/current` - 获取当前集群信息
- `POST /api/clusters/switch` - 切换到指定集群

## 📖 使用说明

### 1. 用户登录
1. 访问应用首页，系统将重定向到登录页面
2. 输入用户名和密码进行登录
3. 登录成功后跳转到仪表板

### 2. 快照类管理
访问 "快照类" 页面查看集群中可用的 VolumeSnapshotClass 和相关的 StorageClass 信息。

### 3. 创建快照
1. 进入 "快照管理" 页面
2. 点击 "创建快照" 按钮
3. 填写快照名称、选择命名空间、快照类和源 PVC
4. 提交创建

### 4. PVC 管理
在 "PVC 管理" 页面可以查看所有命名空间的持久卷声明，了解存储使用情况。

### 5. 定时任务管理
1. 进入 "定时任务" 页面
2. 点击 "创建定时任务"
3. 配置任务名称、命名空间、PVC、快照类和 Cron 表达式
4. 任务创建后自动按计划执行

### 6. Ceph 集群监控
在 "Ceph 集群" 页面可以：
- 查看集群整体状态和健康信息
- 监控存储池使用情况
- 查看 OSD 运行状态

### 7. 多集群管理
在 "集群管理" 页面可以：
- 查看所有配置的 Kubernetes 集群
- 监控各集群的连接状态和健康信息
- 切换到不同的集群进行管理操作
- 查看当前正在使用的集群信息

### 8. 用户管理
管理员可以在 "用户管理" 页面：
- 创建新用户账户
- 管理用户权限
- 删除不需要的用户

### Cron 表达式示例
- `0 2 * * *` - 每天凌晨 2 点
- `0 */6 * * *` - 每 6 小时
- `0 0 * * 0` - 每周日午夜
- `0 30 1 * * *` - 每天凌晨 1:30

## 🔨 构建选项

### 构建命令说明
```bash
make help                 # 查看所有可用命令
make init                 # 初始化项目依赖
make build                # 构建项目（分离模式）
make build-with-ceph      # 构建项目（包含 Ceph 支持）
make build-all            # 构建单一可执行文件（包含前端）
make build-statik         # 仅生成前端静态资源
make build-image          # 构建 Docker 镜像
make build-version        # 构建指定版本镜像
make version-bump         # 升级版本号
make clean                # 清理构建文件
```

### 版本管理
```bash
# 查看当前版本
cat VERSION

# 升级版本
make version-bump TYPE=patch    # 补丁版本 (2.1.26 -> 2.1.27)
make version-bump TYPE=minor    # 次要版本 (2.1.26 -> 2.2.0)
make version-bump TYPE=major    # 主要版本 (2.1.26 -> 3.0.0)

# 构建指定版本镜像
make build-version VERSION=2.1.27
```

## 🚀 部署选项

### 1. 单一可执行文件部署（推荐）

构建包含前端资源的单一可执行文件：

```bash
# 构建
make build-all

# 直接运行（确保有 kubeconfig）
./bin/k8s-volume-snapshots

# 后台运行
nohup ./bin/k8s-volume-snapshots > app.log 2>&1 &
```

### 2. Docker 部署

使用 Docker Compose 快速部署：

```bash
# 准备多集群配置文件（可选）
cp config/clusters.yaml /path/to/your/clusters-config.yaml
# 编辑配置文件，添加你的集群信息

# 构建并启动服务
cd docker
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 3. Kubernetes 部署

部署到 Kubernetes 集群：

```bash
# 1. 创建 RBAC 权限
kubectl apply -f config/rbac.yaml

# 2. 构建镜像（如需要）
make build-image
docker tag k8s-volume-snapshots:$(cat VERSION) your-registry/k8s-volume-snapshots:$(cat VERSION)
docker push your-registry/k8s-volume-snapshots:$(cat VERSION)

# 3. 部署应用
kubectl apply -f k8s/statefulset.yaml

# 4. 检查状态
kubectl get pods -l app=k8s-volume-snapshots

# 5. 访问应用（获取服务地址）
kubectl get svc k8s-volume-snapshots-service
```

### 4. 创建 VolumeSnapshotClass

如果集群中没有 VolumeSnapshotClass，可以使用示例配置：

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: csi-rbdplugin-snapclass
driver: rbd.csi.ceph.com
deletionPolicy: Delete
parameters:
  clusterID: <your-cluster-id>
  csi.storage.k8s.io/snapshotter-secret-name: csi-rbd-secret
  csi.storage.k8s.io/snapshotter-secret-namespace: ceph-csi
```

## 📋 配置文件说明

### Kubernetes 配置
- `config/rbac.yaml` - RBAC 权限配置
- `k8s/statefulset.yaml` - Kubernetes StatefulSet 部署配置

### Docker 配置
- `Dockerfile` - 单一镜像构建文件（多阶段构建）
- `docker/docker-compose.yml` - Docker Compose 配置
- `docker/Dockerfile.backend` - 后端单独镜像（已弃用）
- `docker/Dockerfile.frontend` - 前端单独镜像（已弃用）
- `docker/nginx.conf` - Nginx 配置（已弃用）

### 构建配置
- `Makefile` - 主要构建和部署命令
- `VERSION` - 项目版本配置
- `scripts/` - 各种构建和部署脚本

## 🎯 架构优势

### 单一镜像架构
- ✅ **简化部署**：从 2 个镜像减少到 1 个镜像
- ✅ **减少资源占用**：无需额外的 Nginx 容器
- ✅ **降低复杂度**：消除前后端通信配置
- ✅ **提高性能**：静态资源直接由 Go 服务提供
- ✅ **便于维护**：单一入口点，统一日志管理

### 功能完备性
- 🔐 **完整的用户认证系统**：支持用户登录、权限管理
- 📊 **全面的监控功能**：Kubernetes 和 Ceph 集群监控
- ⏰ **灵活的定时任务**：支持复杂的快照计划
- 🔧 **易于扩展**：模块化设计，便于添加新功能

## 🛠️ 开发指南

### 添加新的功能模块
1. **后端**：
   - 在 `models/` 中定义数据结构
   - 在 `services/` 中实现业务逻辑
   - 在 `controllers/` 中创建 API 接口
   - 在 `main.go` 中注册路由

2. **前端**：
   - 在 `views/` 中创建页面组件
   - 在 `api/` 中封装 API 调用
   - 在 `router/` 中配置路由
   - 更新导航菜单

### 添加新的 CSI 驱动支持
1. 修改 `backend/services/k8s_service.go` 中的驱动匹配逻辑
2. 更新前端显示逻辑以支持新驱动的特殊参数
3. 添加相应的测试用例

### 扩展 Ceph 集群功能
在 `backend/services/ceph_service.go` 中可以添加：
- 更多集群监控指标
- 存储池管理功能
- OSD 管理操作

### 扩展定时任务功能
在 `backend/controllers/scheduled_controller.go` 中可以添加：
- 任务执行历史记录
- 失败重试机制
- 邮件通知功能

## 🔍 故障排除

### 常见问题

**1. Kubernetes 连接失败**
- 检查 kubeconfig 文件路径和权限
- 确认集群可访问性
- 验证 RBAC 权限配置

**2. 创建快照失败**
- 验证 VolumeSnapshotClass 配置
- 检查 PVC 状态是否为 Bound
- 确认 CSI 驱动支持快照功能

**3. 定时任务不执行**
- 检查 Cron 表达式格式
- 查看后端日志确认错误信息
- 确认快照类和 PVC 是否存在

**4. Ceph 集群连接失败**
- 检查 Ceph 配置文件
- 确认 Ceph 客户端库安装
- 验证网络连接和认证信息

**5. 用户登录失败**
- 检查用户凭据
- 确认用户账户状态
- 查看认证服务日志

**6. 前端页面无法访问**
- 确认静态资源构建成功
- 检查 statik 文件生成
- 验证路由配置

### 日志查看
```bash
# 查看应用日志（单一可执行文件模式）
tail -f app.log

# 查看 Docker 容器日志
docker logs k8s-volume-snapshots

# 查看 Kubernetes Pod 日志
kubectl logs -f deployment/k8s-volume-snapshots
```

### 调试模式
```bash
# 启用详细日志
export GIN_MODE=debug
./bin/k8s-volume-snapshots

# 或在环境变量中设置
GIN_MODE=debug make start-backend
```

---

**版本**: v2.2.3
