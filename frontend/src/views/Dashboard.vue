<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div class="dashboard-view" v-loading="loading">
    <!-- 当前集群信息栏 -->
    <el-card class="current-cluster-info" shadow="hover" v-if="currentCluster">
      <div class="cluster-info-container">
        <div class="cluster-info-left">
          <el-icon :size="24" color="#409EFF" class="cluster-icon">
            <Monitor />
          </el-icon>
          <div class="cluster-text">
            <div class="cluster-name">{{ currentCluster.display_name }}</div>
            <div class="cluster-description">{{ currentCluster.description }}</div>
          </div>
        </div>
        <div class="cluster-info-right">
          <el-tag :type="getClusterStatusType(currentCluster.status)" size="large">
            <el-icon><Connection /></el-icon>
            {{ getClusterStatusText(currentCluster.status) }}
          </el-tag>
          <el-button
            type="primary"
            size="small"
            @click="$router.push('/clusters')"
            style="margin-left: 12px;"
          >
            集群管理
          </el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="20" :style="{ marginTop: currentCluster ? '20px' : '0' }">
      <!-- 统计卡片 -->
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ stats.snapshotClasses }}</div>
            <div class="stat-label">快照类</div>
          </div>
          <el-icon class="stat-icon" :size="40" color="#409EFF">
            <Folder />
          </el-icon>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ stats.totalSnapshots }}</div>
            <div class="stat-label">总快照数</div>
          </div>
          <el-icon class="stat-icon" :size="40" color="#67C23A">
            <CameraFilled />
          </el-icon>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ stats.readySnapshots }}</div>
            <div class="stat-label">就绪快照</div>
          </div>
          <el-icon class="stat-icon" :size="40" color="#E6A23C">
            <SuccessFilled />
          </el-icon>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ stats.scheduledTasks }}</div>
            <div class="stat-label">定时任务</div>
          </div>
          <el-icon class="stat-icon" :size="40" color="#F56C6C">
            <Clock />
          </el-icon>
        </el-card>
      </el-col>
    </el-row>

    <!-- Ceph集群状态 -->
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="6">
        <el-card class="stat-card ceph-card">
          <div class="stat-content">
            <div class="stat-number" :style="{color: getCephHealthColor()}">
              {{ getCephStatusText() }}
            </div>
            <div class="stat-label">Ceph集群状态</div>
          </div>
                       <el-icon class="stat-icon" :size="40" :color="getCephHealthColor()">
               <Monitor />
             </el-icon>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ cephStats.totalPools }}</div>
            <div class="stat-label">存储池数量</div>
          </div>
                       <el-icon class="stat-icon" :size="40" color="#E6A23C">
               <DataBoard />
             </el-icon>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ cephStats.totalOSDs }}</div>
            <div class="stat-label">OSD节点数</div>
          </div>
                       <el-icon class="stat-icon" :size="40" color="#67C23A">
               <Box />
             </el-icon>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
                            <div class="stat-number">{{ cephStats.usagePercent }}%</div>
            <div class="stat-label">存储使用率</div>
          </div>
          <el-icon class="stat-icon" :size="40" color="#409EFF">
            <PieChart />
          </el-icon>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快照类列表 -->
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>VolumeSnapshotClass 列表</span>
              <el-button type="primary" size="small" @click="$router.push('/snapshot-classes')">
                查看全部
              </el-button>
            </div>
          </template>
          <el-table :data="snapshotClasses.slice(0, 5)" style="width: 100%">
            <el-table-column prop="volumeSnapshotClass.metadata.name" label="名称" />
            <el-table-column prop="volumeSnapshotClass.driver" label="驱动" />
            <el-table-column prop="relatedStorageClass.metadata.name" label="关联存储类" />
          </el-table>
        </el-card>
      </el-col>

      <!-- 最近快照 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近快照</span>
              <el-button type="primary" size="small" @click="$router.push('/snapshots')">
                查看全部
              </el-button>
            </div>
          </template>
          <el-table :data="recentSnapshots.slice(0, 5)" style="width: 100%" table-layout="auto">
            <el-table-column prop="volumeSnapshot.metadata.name" label="名称" min-width="120" />
            <el-table-column prop="volumeSnapshot.metadata.namespace" label="命名空间" min-width="100" />
            <el-table-column label="状态" min-width="80">
              <template #default="scope">
                <el-tag
                  :type="getSnapshotStatusType(scope.row.volumeSnapshot)"
                  size="small"
                >
                  {{ getSnapshotStatus(scope.row.volumeSnapshot) }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getVolumeSnapshotClasses, getVolumeSnapshots, getScheduledSnapshots, getCurrentCluster } from '../api'
import { getCephClusterStatus, getCephPools } from '../api/ceph'
import { Folder, CameraFilled, SuccessFilled, Clock, Monitor, DataBoard, Box, PieChart, Connection } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const loading = ref(true)

const currentCluster = ref(null)

const stats = reactive({
  snapshotClasses: 0,
  totalSnapshots: 0,
  readySnapshots: 0,
  scheduledTasks: 0
})

const cephStats = reactive({
  status: 'N/A',
  totalPools: 0,
  totalOSDs: 0,
  usagePercent: 0,
  connected: false
})

const snapshotClasses = ref([])
const recentSnapshots = ref([])

const getSnapshotStatus = (snapshot) => {
  if (snapshot.status?.readyToUse) {
    return '就绪'
  } else if (snapshot.status?.error) {
    return '错误'
  } else {
    return '创建中'
  }
}

const getSnapshotStatusType = (snapshot) => {
  if (snapshot.status?.readyToUse) {
    return 'success'
  } else if (snapshot.status?.error) {
    return 'danger'
  } else {
    return 'warning'
  }
}

const getCephHealthColor = () => {
  if (!cephStats.connected) return '#909399'
  switch (cephStats.status) {
    case 'HEALTH_OK': return '#67C23A'
    case 'HEALTH_WARN': return '#E6A23C'
    case 'HEALTH_ERR': return '#F56C6C'
    default: return '#909399'
  }
}

const getCephStatusText = () => {
  if (!cephStats.connected) return '未连接'
  switch (cephStats.status) {
    case 'HEALTH_OK': return '健康'
    case 'HEALTH_WARN': return '警告'
    case 'HEALTH_ERR': return '错误'
    default: return '未知'
  }
}

const getClusterStatusType = (status) => {
  switch (status) {
    case 'online': return 'success'
    case 'offline': return 'warning'
    case 'error': return 'danger'
    default: return 'info'
  }
}

const getClusterStatusText = (status) => {
  switch (status) {
    case 'online': return '在线'
    case 'offline': return '离线'
    case 'error': return '错误'
    default: return '未知'
  }
}

const loadData = async () => {
  if (!authStore.isAuthenticated) {
    console.log('用户未认证，跳过数据加载')
    loading.value = false
    return
  }

  try {
    loading.value = true
    console.log('开始加载仪表板数据...')

    // 加载当前集群信息
    try {
      const clusterData = await getCurrentCluster()
      currentCluster.value = clusterData.cluster
      console.log('当前集群信息:', currentCluster.value)
    } catch (error) {
      console.warn('加载当前集群信息失败:', error)
      currentCluster.value = null
    }

    // 加载快照类
    const snapshotClassesData = await getVolumeSnapshotClasses()
    snapshotClasses.value = Array.isArray(snapshotClassesData) ? snapshotClassesData : []
    stats.snapshotClasses = snapshotClasses.value.length

    // 加载快照
    const snapshotsData = await getVolumeSnapshots()
    if (Array.isArray(snapshotsData)) {
      recentSnapshots.value = snapshotsData.sort((a, b) => {
        const timeA = new Date(a.volumeSnapshot.metadata.creationTimestamp)
        const timeB = new Date(b.volumeSnapshot.metadata.creationTimestamp)
        return timeB - timeA
      })
      stats.totalSnapshots = snapshotsData.length
      stats.readySnapshots = snapshotsData.filter(s => s.volumeSnapshot.status?.readyToUse).length
    } else {
      recentSnapshots.value = []
      stats.totalSnapshots = 0
      stats.readySnapshots = 0
    }

    // 加载定时任务
    const scheduledData = await getScheduledSnapshots()
    stats.scheduledTasks = Array.isArray(scheduledData) ? scheduledData.length : 0

    // 加载Ceph集群状态（不阻塞主要功能）
    try {
      const cephData = await getCephClusterStatus()
      if (cephData) {
        cephStats.status = cephData.health || 'N/A'
        cephStats.totalOSDs = cephData.osds?.total || 0
        cephStats.usagePercent = Math.round(cephData.capacity?.usagePercent || 0)
        cephStats.connected = true

        // 获取Pool数量 - 直接从集群状态获取
        cephStats.totalPools = cephData.pools?.length || 0
        // 如果集群状态中没有pools信息，则单独请求
        if (cephStats.totalPools === 0) {
          try {
            const poolsData = await getCephPools()
            cephStats.totalPools = Array.isArray(poolsData) ? poolsData.length : 0
          } catch (poolError) {
            console.warn('获取Ceph存储池数据失败:', poolError)
            cephStats.totalPools = 0
          }
        }
      }
    } catch (error) {
      console.warn('加载Ceph集群状态失败:', error)
      cephStats.connected = false
      cephStats.status = '连接失败'
    }

    console.log('仪表板数据加载完成')
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

// 初始化仪表板
const initDashboard = async () => {
  console.log('初始化仪表板，当前认证状态:', authStore.isAuthenticated)

  // 如果已经认证，直接加载数据
  if (authStore.isAuthenticated) {
    await loadData()
    return
  }

  // 尝试恢复认证状态
  const restored = authStore.restoreAuth()
  console.log('尝试恢复认证状态:', restored)

  if (restored) {
    // 验证token有效性
    console.log('验证token有效性...')
    const isValid = await authStore.checkToken()
    console.log('Token有效性验证结果:', isValid)

    if (isValid) {
      await loadData()
    } else {
      console.log('Token无效，等待用户登录')
      loading.value = false
    }
  } else {
    console.log('无认证信息，等待用户登录')
    loading.value = false
  }
}

onMounted(() => {
  initDashboard()
})
</script>

<style scoped>
.dashboard-view {
  padding: 20px;
}

.current-cluster-info {
  margin-bottom: 0;
  border-left: 4px solid #409EFF;
}

.cluster-info-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.cluster-info-left {
  display: flex;
  align-items: center;
}

.cluster-icon {
  margin-right: 16px;
  flex-shrink: 0;
}

.cluster-text {
  display: flex;
  flex-direction: column;
}

.cluster-name {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.cluster-description {
  font-size: 14px;
  color: #909399;
}

.cluster-info-right {
  display: flex;
  align-items: center;
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

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
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

.el-table {
  font-size: 14px;
}

/* 确保表格内容正确对齐 */
.el-table .el-table__row td {
  padding: 12px 0;
}

.el-table .el-table__header th {
  padding: 12px 0;
  font-weight: 500;
}

/* 修复表格单元格内容的显示 */
.el-table .cell {
  padding: 0 10px;
}

/* 响应式布局 */
@media (max-width: 768px) {
  .cluster-info-container {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }

  .cluster-info-right {
    width: 100%;
    justify-content: flex-start;
  }

  .cluster-stats .el-col {
    margin-bottom: 15px;
  }
}
</style>
