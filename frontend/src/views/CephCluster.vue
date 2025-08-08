<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div class="ceph-cluster-view" v-loading="loading">
    <!-- 连接状态提示 -->
    <el-alert
      v-if="!connectionStatus.connected"
      title="Ceph集群连接不可用"
      type="error"
      :description="connectionStatus.message"
      show-icon
      :closable="false"
      style="margin-bottom: 20px;"
    />

    <!-- 集群状态概览 -->
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-header">
              <span class="stat-title">集群健康状态</span>
              <el-tag
                :type="getHealthStatusType(clusterStatus.health)"
                size="small"
              >
                {{ getHealthStatusText(clusterStatus.health) }}
              </el-tag>
            </div>
            <div class="stat-value">{{ clusterStatus.version || 'Unknown' }}</div>
            <div class="stat-label">Ceph版本</div>
          </div>
          <el-icon class="stat-icon" :size="40" :color="getHealthColor(clusterStatus.health)">
            <Monitor />
          </el-icon>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ clusterStatus.osds?.total || 0 }}</div>
            <div class="stat-label">
              OSD总数 ({{ clusterStatus.osds?.up || 0 }} 在线)
            </div>
          </div>
          <el-icon class="stat-icon" :size="40" color="#409EFF">
                            <Box />
          </el-icon>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ clusterStatus.monitors?.length || 0 }}</div>
            <div class="stat-label">Monitor节点</div>
          </div>
          <el-icon class="stat-icon" :size="40" color="#67C23A">
            <View />
          </el-icon>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ pools.length }}</div>
            <div class="stat-label">存储池数量</div>
          </div>
          <el-icon class="stat-icon" :size="40" color="#E6A23C">
            <Files />
          </el-icon>
        </el-card>
      </el-col>
    </el-row>

    <!-- 容量信息 -->
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>集群容量使用情况</span>
              <el-button
                type="text"
                size="small"
                @click="refreshData"
                :loading="refreshing"
              >
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          <div class="capacity-info">
            <div class="capacity-item">
              <span class="capacity-label">总容量:</span>
              <span class="capacity-value">{{ formatBytes(clusterStatus.capacity?.totalBytes) }}</span>
            </div>
            <div class="capacity-item">
              <span class="capacity-label">已使用:</span>
              <span class="capacity-value used">{{ formatBytes(clusterStatus.capacity?.usedBytes) }}</span>
            </div>
            <div class="capacity-item">
              <span class="capacity-label">可用:</span>
              <span class="capacity-value available">{{ formatBytes(clusterStatus.capacity?.availBytes) }}</span>
            </div>
            <div class="capacity-item">
              <span class="capacity-label">使用率:</span>
              <span class="capacity-value">{{ (clusterStatus.capacity?.usagePercent || 0).toFixed(2) }}%</span>
            </div>

            <!-- 使用率进度条 -->
            <div class="capacity-progress" style="margin-top: 15px;">
              <el-progress
                :percentage="clusterStatus.capacity?.usagePercent || 0"
                :color="getUsageColor(clusterStatus.capacity?.usagePercent)"
                :stroke-width="8"
                :show-text="false"
              />
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>Monitor节点状态</span>
            </div>
          </template>
          <div class="monitors-list">
            <div
              v-for="monitor in clusterStatus.monitors"
              :key="monitor.name"
              class="monitor-item"
            >
              <div class="monitor-info">
                <el-icon class="monitor-icon" :color="monitor.status === 'up' ? '#67C23A' : '#F56C6C'">
                  <CircleCheck v-if="monitor.status === 'up'" />
                  <CircleClose v-else />
                </el-icon>
                <div class="monitor-details">
                  <div class="monitor-name">{{ monitor.name }}</div>
                  <div class="monitor-address">{{ monitor.address }}</div>
                </div>
              </div>
              <el-tag
                :type="monitor.inQuorum ? 'success' : 'danger'"
                size="small"
              >
                {{ monitor.inQuorum ? '在Quorum中' : '不在Quorum中' }}
              </el-tag>
            </div>

            <div v-if="!clusterStatus.monitors || clusterStatus.monitors.length === 0" class="no-data">
              暂无Monitor节点信息
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Pool信息列表 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>存储池 (Pools) 详情</span>
          <div class="header-actions">
            <el-button
              type="text"
              size="small"
              @click="refreshPools"
              :loading="refreshing"
            >
              <el-icon><Refresh /></el-icon>
              刷新Pools
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="pools"
        style="width: 100%"
        :default-sort="{prop: 'name', order: 'ascending'}"
        table-layout="auto"
      >
        <el-table-column prop="name" label="Pool名称" min-width="120" sortable>
          <template #default="scope">
            <div class="pool-name">
                              <el-icon class="pool-icon"><DataBoard /></el-icon>
              {{ scope.row.name }}
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="type" label="类型" min-width="90" sortable>
          <template #default="scope">
            <el-tag :type="scope.row.type === 'replicated' ? 'primary' : 'success'" size="small">
              {{ scope.row.type }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="size" label="副本数" min-width="70" sortable />
        <el-table-column prop="pgNum" label="PG数" min-width="70" sortable />
        <el-table-column prop="objects" label="对象数" min-width="90" sortable>
          <template #default="scope">
            {{ formatNumber(scope.row.objects) }}
          </template>
        </el-table-column>

        <el-table-column label="已使用" min-width="100" sortable prop="usedBytes">
          <template #default="scope">
            {{ formatBytes(scope.row.usedBytes) }}
          </template>
        </el-table-column>

        <el-table-column label="可用空间" min-width="100" sortable prop="maxAvailBytes">
          <template #default="scope">
            {{ formatBytes(scope.row.maxAvailBytes) }}
          </template>
        </el-table-column>

      </el-table>

      <div v-if="pools.length === 0" class="no-data-table">
        <el-empty description="暂无存储池数据" />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import {
  getCephClusterStatus,
  getCephPools,
  getCephConnectionStatus,
  formatBytes
} from '../api/ceph'
import {
  Monitor,
  Box,
  View,
  Files,
  Refresh,
  CircleCheck,
  CircleClose,
  DataBoard
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

// 响应式数据
const loading = ref(true)
const refreshing = ref(false)
const clusterStatus = reactive({
  health: '',
  version: '',
  monitors: [],
  osds: { total: 0, up: 0, in: 0 },
  pgs: { total: 0, active: 0, clean: 0, degraded: 0 },
  capacity: { totalBytes: 0, usedBytes: 0, availBytes: 0, usagePercent: 0 }
})
const pools = ref([])
const connectionStatus = reactive({
  connected: false,
  message: '',
  checkedAt: null
})

// 定时刷新
let refreshTimer = null

// 方法
const getHealthStatusType = (health) => {
  switch (health) {
    case 'HEALTH_OK': return 'success'
    case 'HEALTH_WARN': return 'warning'
    case 'HEALTH_ERR': return 'danger'
    default: return 'info'
  }
}

const getHealthStatusText = (health) => {
  switch (health) {
    case 'HEALTH_OK': return '健康'
    case 'HEALTH_WARN': return '警告'
    case 'HEALTH_ERR': return '错误'
    default: return '未知'
  }
}

const getHealthColor = (health) => {
  switch (health) {
    case 'HEALTH_OK': return '#67C23A'
    case 'HEALTH_WARN': return '#E6A23C'
    case 'HEALTH_ERR': return '#F56C6C'
    default: return '#909399'
  }
}

const getUsageColor = (percent) => {
  if (!percent) return '#409EFF'
  if (percent < 60) return '#67C23A'
  if (percent < 80) return '#E6A23C'
  return '#F56C6C'
}

const formatNumber = (num) => {
  if (!num) return '0'
  return new Intl.NumberFormat().format(num)
}

// 检查连接状态
const checkConnectionStatus = async () => {
  try {
    const data = await getCephConnectionStatus()
    Object.assign(connectionStatus, data)
  } catch (error) {
    connectionStatus.connected = false
    connectionStatus.message = 'Ceph服务不可用: ' + error.message
  }
}

// 加载集群状态
const loadClusterStatus = async () => {
  try {
    const data = await getCephClusterStatus()
    Object.assign(clusterStatus, data)
  } catch (error) {
    console.error('加载集群状态失败:', error)
    ElMessage.error('加载集群状态失败: ' + error.message)
  }
}

// 加载Pool信息
const loadPools = async () => {
  try {
    const data = await getCephPools()
    pools.value = Array.isArray(data) ? data : []
  } catch (error) {
    console.error('加载Pool信息失败:', error)
    ElMessage.error('加载Pool信息失败: ' + error.message)
    pools.value = []
  }
}

// 加载所有数据
const loadData = async () => {
  loading.value = true
  try {
    await Promise.all([
      checkConnectionStatus(),
      loadClusterStatus(),
      loadPools()
    ])
  } catch (error) {
    console.error('加载数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 刷新数据
const refreshData = async () => {
  refreshing.value = true
  try {
    await loadData()
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('数据刷新失败')
  } finally {
    refreshing.value = false
  }
}

// 刷新Pool数据
const refreshPools = async () => {
  refreshing.value = true
  try {
    await loadPools()
    ElMessage.success('Pool数据刷新成功')
  } catch (error) {
    ElMessage.error('Pool数据刷新失败')
  } finally {
    refreshing.value = false
  }
}

// 启动定时刷新
const startAutoRefresh = () => {
  // 每30秒自动刷新一次
  refreshTimer = setInterval(async () => {
    if (!refreshing.value) {
      await loadData()
    }
  }, 30000)
}

// 停止定时刷新
const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 生命周期
onMounted(async () => {
  await loadData()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.ceph-cluster-view {
  padding: 20px;
}

.stat-card {
  position: relative;
  overflow: hidden;
}

.stat-card .el-card__body {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
}

.stat-content {
  flex: 1;
}

.stat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.stat-title {
  font-size: 14px;
  color: #909399;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}

.stat-icon {
  opacity: 0.2;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

/* 容量信息样式 */
.capacity-info {
  padding: 10px 0;
}

.capacity-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.capacity-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

.capacity-value {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.capacity-value.used {
  color: #E6A23C;
}

.capacity-value.available {
  color: #67C23A;
}

/* Monitor节点样式 */
.monitors-list {
  max-height: 300px;
  overflow-y: auto;
}

.monitor-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f2f5;
}

.monitor-item:last-child {
  border-bottom: none;
}

.monitor-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.monitor-icon {
  font-size: 18px;
}

.monitor-details {
  flex: 1;
}

.monitor-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.monitor-address {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

/* Pool表格样式 */
.pool-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pool-icon {
  color: #409EFF;
  font-size: 16px;
}

.usage-cell {
  min-width: 100px;
}

.usage-text {
  font-size: 13px;
  font-weight: 500;
}

/* 无数据样式 */
.no-data {
  text-align: center;
  padding: 40px 0;
  color: #909399;
  font-size: 14px;
}

.no-data-table {
  padding: 20px 0;
}

/* 布局对齐优化 */
.el-row {
  margin-bottom: 20px;
}

.el-row:last-child {
  margin-bottom: 0;
}

.el-col {
  padding: 0 10px;
}

.el-card {
  height: 100%;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.el-card__header {
  padding: 18px 20px;
  border-bottom: 1px solid #ebeef5;
}

.el-card__body {
  padding: 20px;
}

/* 统计卡片高度统一 */
.stat-card .el-card__body {
  min-height: 120px;
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
}

/* 容量信息卡片高度调整 */
.capacity-info {
  min-height: 200px;
}

/* Monitor信息对齐 */
.monitor-list {
  min-height: 200px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .ceph-cluster-view {
    padding: 16px;
  }

  .stat-number {
    font-size: 24px;
  }

  .stat-value {
    font-size: 16px;
  }

  .capacity-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .monitor-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
}
</style>
