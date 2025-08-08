<template>
  <div class="cluster-management">
    <div class="page-header">
      <h2>集群管理</h2>
      <p class="page-description">管理和切换 Kubernetes 集群</p>
    </div>

    <!-- 当前集群信息卡片 -->
    <el-card class="current-cluster-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <h3>当前集群</h3>
          <el-button @click="refreshClusters" :loading="loading" circle>
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>
      </template>
      <div v-if="currentCluster" class="current-cluster-info">
        <div class="cluster-info-row">
          <div class="info-item">
            <span class="label">集群名称:</span>
            <span class="value">{{ currentCluster.name }}</span>
          </div>
          <div class="info-item">
            <span class="label">显示名称:</span>
            <span class="value">{{ currentCluster.display_name }}</span>
          </div>
          <div class="info-item">
            <span class="label">状态:</span>
            <el-tag :type="getStatusType(currentCluster.status)">
              {{ getStatusText(currentCluster.status) }}
            </el-tag>
          </div>
        </div>
        <div class="cluster-description">
          <span class="label">描述:</span>
          <span class="value">{{ currentCluster.description }}</span>
        </div>
      </div>
      <div v-else class="no-cluster">
        <el-empty description="未找到当前集群信息" />
      </div>
    </el-card>

    <!-- 所有集群列表 -->
    <el-card class="clusters-list-card" shadow="hover">
      <template #header>
        <h3>所有集群</h3>
      </template>

      <el-table v-loading="loading" :data="clusters" style="width: 100%">
        <el-table-column prop="name" label="集群名称" width="150" />
        <el-table-column prop="display_name" label="显示名称" width="180" />
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="启用状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">
              {{ row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最后检查" width="180">
          <template #default="{ row }">
            {{ formatLastCheck(row.last_check) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.name !== currentClusterName && row.enabled && row.status === 'online'"
              type="primary"
              size="small"
              @click="handleSwitchCluster(row.name)"
              :loading="switchingCluster === row.name"
            >
              切换
            </el-button>
            <el-tag v-else-if="row.name === currentClusterName" type="success" size="small">
              当前
            </el-tag>
            <el-tag v-else type="info" size="small">
              不可用
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 集群连接统计 -->
    <el-row :gutter="20" class="cluster-stats">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ totalClusters }}</div>
            <div class="stat-label">总集群数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number online">{{ onlineClusters }}</div>
            <div class="stat-label">在线集群</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number offline">{{ offlineClusters }}</div>
            <div class="stat-label">离线集群</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number enabled">{{ enabledClusters }}</div>
            <div class="stat-label">启用集群</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { getClusters, getCurrentCluster, switchCluster } from '@/api'
import { Refresh } from '@element-plus/icons-vue'

const clusters = ref([])
const currentCluster = ref(null)
const currentClusterName = ref('')
const loading = ref(false)
const switchingCluster = ref('')

// 计算属性
const totalClusters = computed(() => clusters.value.length)
const onlineClusters = computed(() => clusters.value.filter(c => c.status === 'online').length)
const offlineClusters = computed(() => clusters.value.filter(c => c.status === 'offline' || c.status === 'error').length)
const enabledClusters = computed(() => clusters.value.filter(c => c.enabled).length)

// 获取状态类型
const getStatusType = (status) => {
  switch (status) {
    case 'online': return 'success'
    case 'offline': return 'warning'
    case 'error': return 'danger'
    case 'disabled': return 'info'
    default: return 'info'
  }
}

// 获取状态文本
const getStatusText = (status) => {
  switch (status) {
    case 'online': return '在线'
    case 'offline': return '离线'
    case 'error': return '错误'
    case 'disabled': return '禁用'
    default: return '未知'
  }
}

// 格式化最后检查时间
const formatLastCheck = (lastCheck) => {
  if (!lastCheck) return '-'
  return new Date(lastCheck).toLocaleString()
}

// 加载集群数据
const loadClusters = async () => {
  try {
    loading.value = true
    const clustersData = await getClusters()
    clusters.value = clustersData.clusters || []
    currentClusterName.value = clustersData.current || ''

    // 加载当前集群详情
    if (currentClusterName.value) {
      const currentData = await getCurrentCluster()
      currentCluster.value = currentData.cluster
    }
  } catch (error) {
    console.error('加载集群数据失败:', error)
    ElMessage.error('加载集群数据失败')
  } finally {
    loading.value = false
  }
}

// 刷新集群数据
const refreshClusters = async () => {
  await loadClusters()
  ElMessage.success('集群数据已刷新')
}

// 处理切换集群
const handleSwitchCluster = async (clusterName) => {
  try {
    switchingCluster.value = clusterName
    await switchCluster(clusterName)
    ElMessage.success(`已切换到集群: ${clusterName}`)
    await loadClusters() // 重新加载数据
  } catch (error) {
    console.error('切换集群失败:', error)
    ElMessage.error('切换集群失败')
  } finally {
    switchingCluster.value = ''
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadClusters()
})
</script>

<style scoped>
.cluster-management {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.page-description {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.current-cluster-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
  color: #303133;
}

.current-cluster-info {
  padding: 10px 0;
}

.cluster-info-row {
  display: flex;
  gap: 40px;
  margin-bottom: 15px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cluster-description {
  display: flex;
  align-items: center;
  gap: 8px;
}

.label {
  color: #909399;
  font-weight: 500;
}

.value {
  color: #303133;
  font-weight: 400;
}

.no-cluster {
  text-align: center;
  padding: 40px 0;
}

.clusters-list-card {
  margin-bottom: 20px;
}

.cluster-stats {
  margin-top: 20px;
}

.stat-card {
  text-align: center;
  cursor: default;
}

.stat-content {
  padding: 10px 0;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 5px;
}

.stat-number.online {
  color: #67c23a;
}

.stat-number.offline {
  color: #f56c6c;
}

.stat-number.enabled {
  color: #409eff;
}

.stat-label {
  color: #909399;
  font-size: 14px;
}

/* 响应式布局 */
@media (max-width: 768px) {
  .cluster-info-row {
    flex-direction: column;
    gap: 15px;
  }

  .cluster-stats .el-col {
    margin-bottom: 15px;
  }
}
</style>
