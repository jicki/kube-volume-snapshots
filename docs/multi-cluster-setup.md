# 多集群配置示例

本文档展示如何配置和使用多集群功能。

## 1. 环境准备

假设您有三个 Kubernetes 集群：

- **生产环境集群**: `prod-cluster` (API: https://k8s-prod.example.com:6443)
- **测试环境集群**: `test-cluster` (API: https://k8s-test.example.com:6443)  
- **开发环境集群**: `dev-cluster` (API: https://k8s-dev.example.com:6443)

## 2. 准备 kubeconfig 文件

为每个集群准备独立的 kubeconfig 文件：

```bash
# 创建配置目录
mkdir -p ~/.kube/clusters

# 复制各集群的 kubeconfig
cp /path/to/prod-kubeconfig ~/.kube/clusters/config-prod
cp /path/to/test-kubeconfig ~/.kube/clusters/config-test
cp /path/to/dev-kubeconfig ~/.kube/clusters/config-dev
```

## 3. 创建多集群配置文件

创建 `config/clusters.yaml` 文件：

```yaml
---
# 多集群配置文件
default_cluster: "production"

clusters:
  - name: "production"
    display_name: "生产环境"
    description: "生产环境 Kubernetes 集群"
    kubeconfig: "/root/.kube/clusters/config-prod"
    enabled: true
    
  - name: "testing"
    display_name: "测试环境"
    description: "测试环境 Kubernetes 集群"
    kubeconfig: "/root/.kube/clusters/config-test"
    enabled: true
    
  - name: "development"
    display_name: "开发环境"
    description: "开发环境 Kubernetes 集群"
    kubeconfig: "/root/.kube/clusters/config-dev"
    enabled: true

global:
  timeout: 30
  qps: 50
  burst: 100
  cache_refresh_interval: 5
```

## 4. Docker Compose 部署

更新 `docker/docker-compose.yml`：

```yaml
version: '3.8'

services:
  k8s-volume-snapshots:
    build:
      context: ..
      dockerfile: Dockerfile
    container_name: k8s-volume-snapshots
    ports:
      - "8081:8081"
    environment:
      - GIN_MODE=release
      - MULTI_CLUSTER_CONFIG=/etc/k8s-volume-snapshots/clusters.yaml
    volumes:
      # 挂载多集群配置文件
      - ../config/clusters.yaml:/etc/k8s-volume-snapshots/clusters.yaml:ro
      # 挂载各集群的 kubeconfig 文件
      - ~/.kube/clusters/config-prod:/root/.kube/clusters/config-prod:ro
      - ~/.kube/clusters/config-test:/root/.kube/clusters/config-test:ro
      - ~/.kube/clusters/config-dev:/root/.kube/clusters/config-dev:ro
    restart: unless-stopped
```

## 5. Kubernetes 部署

### 5.1 创建多集群 ConfigMap

```bash
kubectl create configmap multi-cluster-config \
  --from-file=clusters.yaml=config/clusters.yaml \
  -n default
```

### 5.2 创建 kubeconfig Secret（如果需要外部集群访问）

```bash
# 创建包含多个 kubeconfig 的 Secret
kubectl create secret generic multi-cluster-configs \
  --from-file=config-prod=~/.kube/clusters/config-prod \
  --from-file=config-test=~/.kube/clusters/config-test \
  --from-file=config-dev=~/.kube/clusters/config-dev \
  -n default
```

### 5.3 更新 StatefulSet

参考项目中的 `k8s/statefulset.yaml` 文件，其中已包含多集群配置的挂载。

## 6. 使用方式

### 6.1 启动应用

```bash
# 使用 Docker Compose
cd docker && docker-compose up -d

# 或者直接运行二进制文件（需要配置环境变量）
export MULTI_CLUSTER_CONFIG=/path/to/clusters.yaml
./bin/k8s-volume-snapshots
```

### 6.2 通过界面管理

1. 访问 http://localhost:8081
2. 登录系统
3. 进入 "集群管理" 页面
4. 查看所有配置的集群状态
5. 点击 "切换" 按钮切换到不同的集群
6. 在其他页面（快照管理、PVC管理等）操作当前选中的集群

### 6.3 通过 API 管理

```bash
# 获取所有集群信息
curl -H "Authorization: Bearer <token>" \
  http://localhost:8081/api/clusters

# 获取当前集群
curl -H "Authorization: Bearer <token>" \
  http://localhost:8081/api/clusters/current

# 切换集群
curl -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"cluster_name": "testing"}' \
  http://localhost:8081/api/clusters/switch
```

## 7. 故障排除

### 7.1 集群连接失败

如果某个集群显示为 "离线" 或 "错误" 状态：

1. 检查 kubeconfig 文件是否正确
2. 验证网络连接性
3. 确认集群 API Server 可访问
4. 检查认证凭据是否有效

### 7.2 权限问题

确保每个集群中都有正确的 RBAC 配置：

```bash
# 在每个集群中应用 RBAC 配置
kubectl apply -f config/rbac.yaml --context=prod-context
kubectl apply -f config/rbac.yaml --context=test-context
kubectl apply -f config/rbac.yaml --context=dev-context
```

### 7.3 日志查看

```bash
# Docker Compose 部署
docker-compose logs -f k8s-volume-snapshots

# Kubernetes 部署
kubectl logs -f statefulset/k8s-volume-snapshots -n default
```

## 8. 高级配置

### 8.1 直接 API 连接

对于某些集群，可以不使用 kubeconfig，直接配置 API 连接：

```yaml
clusters:
  - name: "external-cluster"
    display_name: "外部集群"
    description: "通过 API 直接连接的集群"
    server: "https://external-k8s.example.com:6443"
    token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
    certificate_authority_data: "LS0tLS1CRUdJTi..."
    enabled: true
```

### 8.2 集群分组

可以通过 description 字段对集群进行分组管理：

```yaml
clusters:
  - name: "prod-asia"
    display_name: "生产环境-亚洲"
    description: "[生产] 亚洲地区生产集群"
    enabled: true
    
  - name: "prod-europe"  
    display_name: "生产环境-欧洲"
    description: "[生产] 欧洲地区生产集群"
    enabled: true
```

### 8.3 动态配置热重载

系统支持配置文件的热重载。修改配置文件后，可以通过重启应用来应用新配置，或者在未来版本中实现配置热重载功能。

## 9. 安全考虑

1. **kubeconfig 安全**: 确保 kubeconfig 文件的权限设置正确（600）
2. **网络安全**: 使用 TLS 和证书验证
3. **访问控制**: 为应用创建专用的 Service Account，避免使用管理员权限
4. **密钥管理**: 在生产环境中使用 Kubernetes Secret 管理敏感信息

## 10. 监控和告警

建议监控以下指标：

1. 集群连接状态
2. 集群响应时间
3. API 请求成功率
4. 集群切换频率

可以通过应用的健康检查接口获取这些信息。