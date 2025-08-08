<template>
  <div class="pvcs-container">
    <!-- 页面标题 -->
    <el-card class="header-card">
      <div class="header-content">
        <div class="title-section">
          <h2>PVC 管理</h2>
          <p class="subtitle">持久卷声明 (PersistentVolumeClaims) 管理</p>
        </div>
                <div class="action-section">
          <div class="filter-row">
            <el-select
              v-model="selectedNamespace"
              @change="loadPVCs"
              placeholder="选择命名空间"
              filterable
              allow-create
              class="namespace-select"
            >
              <el-option label="所有命名空间" value="all" />
              <el-option
                v-for="ns in namespaces"
                :key="ns.metadata.name"
                :label="ns.metadata.name"
                :value="ns.metadata.name"
              />
            </el-select>
            <el-select
              v-model="selectedStorageClass"
              @change="applyFilters"
              placeholder="选择存储类"
              filterable
              clearable
              class="storage-class-select"
            >
              <el-option label="所有存储类" value="" />
              <el-option
                v-for="sc in storageClasses"
                :key="sc.metadata.name"
                :label="sc.metadata.name"
                :value="sc.metadata.name"
              />
            </el-select>
          </div>
                  <div class="action-row">
            <el-button
              type="primary"
              :icon="Refresh"
              @click="loadPVCs"
              :loading="refreshing"
            >
              {{ refreshing ? '刷新中...' : '刷新' }}
            </el-button>
            <div class="auto-refresh-container">
              <el-switch
                v-model="autoRefreshEnabled"
                @change="toggleAutoRefresh"
                class="auto-refresh-switch"
                :active-icon="Timer"
              />
              <span class="auto-refresh-label">自动刷新</span>
            </div>
          </div>
        </div>
      </div>
    </el-Card>

    <!-- PVC 列表 -->
    <el-card class="table-card">
            <template #header>
        <div class="card-header">
          <span>PVC 列表 ({{ filteredPVCs.length }} / {{ pvcs.length }})</span>
        </div>
      </template>

            <el-table
        :data="filteredPVCs"
        v-loading="loading"
        stripe
        :default-sort="{ prop: 'metadata.creationTimestamp', order: 'descending' }"
        table-layout="auto"
        class="responsive-table"
        :scroll-x="true"
      >
                <el-table-column label="名称" min-width="150" sortable show-overflow-tooltip>
          <template #default="scope">
            <div class="name-cell">
              <el-text type="primary" style="font-weight: 500;">
                {{ scope.row.metadata.name }}
              </el-text>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="命名空间" min-width="100" sortable>
          <template #default="scope">
            <el-tag size="small" type="info">
              {{ scope.row.metadata.namespace }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="80" sortable>
          <template #default="scope">
            <el-tag
              :type="getPVCStatusType(scope.row.status.phase)"
              size="small"
            >
              {{ scope.row.status.phase || '未知' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="存储类" min-width="140" sortable show-overflow-tooltip>
          <template #default="scope">
            <span v-if="scope.row.spec.storageClassName">
              {{ scope.row.spec.storageClassName }}
            </span>
            <span v-else class="no-data">-</span>
          </template>
        </el-table-column>

        <el-table-column label="请求大小" width="90" sortable>
          <template #default="scope">
            {{ scope.row.spec.resources?.requests?.storage || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="实际大小" width="90" sortable>
          <template #default="scope">
            {{ scope.row.status?.capacity?.storage || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="访问模式" min-width="100">
          <template #default="scope">
            <div class="access-modes">
              <el-tag
                v-for="mode in scope.row.spec.accessModes"
                :key="mode"
                size="small"
                type="success"
                style="margin-right: 2px;"
              >
                {{ formatAccessMode(mode) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="绑定的PV" min-width="200" show-overflow-tooltip>
          <template #default="scope">
            <span v-if="scope.row.spec.volumeName" class="pv-name">
              {{ scope.row.spec.volumeName }}
            </span>
            <span v-else class="no-data">未绑定</span>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" min-width="120" sortable>
          <template #default="scope">
            {{ formatTime(scope.row.metadata.creationTimestamp) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="80" fixed="right">
          <template #default="scope">
            <el-button
              size="small"
              @click="viewDetails(scope.row)"
            >
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- PVC 详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="`PVC 详情 - ${selectedPVC?.metadata?.name}`"
      width="70%"
      :before-close="closeDetailDialog"
    >
      <div v-if="selectedPVC" class="detail-content">
        <!-- 基本信息 -->
        <el-descriptions title="基本信息" :column="2" border>
          <el-descriptions-item label="名称">
            {{ selectedPVC.metadata.name }}
          </el-descriptions-item>
          <el-descriptions-item label="命名空间">
            <el-tag size="small" type="info">
              {{ selectedPVC.metadata.namespace }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag
              :type="getPVCStatusType(selectedPVC.status?.phase)"
              size="small"
            >
              {{ selectedPVC.status?.phase || '未知' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="存储类">
            {{ selectedPVC.spec.storageClassName || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="请求大小">
            {{ selectedPVC.spec.resources?.requests?.storage || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="实际大小">
            {{ selectedPVC.status?.capacity?.storage || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="绑定的PV">
            {{ selectedPVC.spec.volumeName || '未绑定' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(selectedPVC.metadata.creationTimestamp) }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- PV详细信息 -->
        <el-descriptions
          v-if="selectedPVC.volumeAttributes"
          title="PV 存储信息"
          :column="2"
          border
          class="mt-4"
        >
          <el-descriptions-item label="镜像名称">
            <span class="code-text">{{ selectedPVC.volumeAttributes.imageName || '-' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="存储池">
            <span class="code-text">{{ selectedPVC.volumeAttributes.pool || '-' }}</span>
          </el-descriptions-item>
        </el-descriptions>

        <!-- 访问模式 -->
        <el-divider content-position="left">访问模式</el-divider>
        <div class="access-modes-detail">
          <el-tag
            v-for="mode in selectedPVC.spec.accessModes"
            :key="mode"
            type="success"
            style="margin-right: 8px;"
          >
            {{ formatAccessMode(mode) }} ({{ mode }})
          </el-tag>
        </div>

        <!-- 标签 -->
        <el-divider content-position="left">标签</el-divider>
        <div class="labels-section">
          <el-tag
            v-for="(value, key) in selectedPVC.metadata.labels"
            :key="key"
            style="margin-right: 8px; margin-bottom: 4px;"
          >
            {{ key }}: {{ value }}
          </el-tag>
          <span v-if="!selectedPVC.metadata.labels || Object.keys(selectedPVC.metadata.labels).length === 0" class="no-data">
            无标签
          </span>
        </div>

        <!-- 注解 -->
        <el-divider content-position="left">注解</el-divider>
        <div class="annotations-section">
          <div
            v-for="(value, key) in selectedPVC.metadata.annotations"
            :key="key"
            class="annotation-item"
          >
            <el-text type="info" size="small">{{ key }}:</el-text>
            <el-text size="small" style="margin-left: 8px;">{{ value }}</el-text>
          </div>
          <span v-if="!selectedPVC.metadata.annotations || Object.keys(selectedPVC.metadata.annotations).length === 0" class="no-data">
            无注解
          </span>
        </div>

        <!-- 条件 -->
        <el-divider content-position="left">条件</el-divider>
        <div v-if="selectedPVC.status?.conditions?.length > 0">
          <el-table :data="selectedPVC.status.conditions" size="small">
            <el-table-column prop="type" label="类型" width="150" />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="scope">
                <el-tag
                  :type="scope.row.status === 'True' ? 'success' : 'warning'"
                  size="small"
                >
                  {{ scope.row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="reason" label="原因" width="120" />
            <el-table-column prop="message" label="消息" />
            <el-table-column prop="lastTransitionTime" label="时间" width="160">
              <template #default="scope">
                {{ formatTime(scope.row.lastTransitionTime) }}
              </template>
            </el-table-column>
          </el-table>
        </div>
        <span v-else class="no-data">无条件信息</span>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getPVCs, getNamespaces, getStorageClasses } from '../api'
import { Refresh, Timer } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const pvcs = ref([])
const namespaces = ref([])
const storageClasses = ref([])
const selectedNamespace = ref('all')
const selectedStorageClass = ref('')
const loading = ref(false)
const refreshing = ref(false)
const detailDialogVisible = ref(false)
const selectedPVC = ref(null)

// 自动刷新相关
const autoRefreshEnabled = ref(false)
const autoRefreshTimer = ref(null)

// 过滤后的PVC列表
const filteredPVCs = computed(() => {
  let filtered = pvcs.value

  // 按存储类过滤
  if (selectedStorageClass.value) {
    filtered = filtered.filter(pvc =>
      pvc.spec.storageClassName === selectedStorageClass.value
    )
  }

  return filtered
})

// 加载数据
const loadPVCs = async () => {
  if (refreshing.value) return

  refreshing.value = true
  loading.value = true

  try {
    const data = await getPVCs(selectedNamespace.value)
    // 适配新的数据结构: PVCWithPVInfo[]
    pvcs.value = (data || []).map(item => ({
      ...item.pvc,
      volumeAttributes: item.volumeAttributes
    }))
  } catch (error) {
    ElMessage.error('获取PVC列表失败: ' + error.message)
    pvcs.value = []
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

const loadNamespaces = async () => {
  try {
    const data = await getNamespaces()
    namespaces.value = data || []
  } catch (error) {
    ElMessage.error('获取命名空间列表失败: ' + error.message)
  }
}

const loadStorageClasses = async () => {
  try {
    const data = await getStorageClasses()
    storageClasses.value = data || []
  } catch (error) {
    ElMessage.error('获取存储类列表失败: ' + error.message)
  }
}

// 应用过滤器
const applyFilters = () => {
  // 过滤逻辑已在 computed 中实现，这里可以添加额外的处理
}

// 格式化函数
const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN')
}

const formatAccessMode = (mode) => {
  const modeMap = {
    ReadWriteOnce: 'RWO',
    ReadOnlyMany: 'ROX',
    ReadWriteMany: 'RWX',
    ReadWriteOncePod: 'RWOP'
  }
  return modeMap[mode] || mode
}

const getPVCStatusType = (phase) => {
  switch (phase) {
    case 'Bound':
      return 'success'
    case 'Pending':
      return 'warning'
    case 'Lost':
      return 'danger'
    default:
      return 'info'
  }
}

// 详情相关
const viewDetails = (pvc) => {
  selectedPVC.value = pvc
  detailDialogVisible.value = true
}

const closeDetailDialog = () => {
  detailDialogVisible.value = false
  selectedPVC.value = null
}

// 自动刷新功能
const startAutoRefresh = () => {
  if (autoRefreshTimer.value) {
    clearInterval(autoRefreshTimer.value)
  }
  autoRefreshTimer.value = setInterval(() => {
    loadPVCs()
  }, 30000) // 30秒刷新一次
}

const stopAutoRefresh = () => {
  if (autoRefreshTimer.value) {
    clearInterval(autoRefreshTimer.value)
    autoRefreshTimer.value = null
  }
}

const toggleAutoRefresh = () => {
  if (autoRefreshEnabled.value) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

// 生命周期
onMounted(async () => {
  await Promise.all([
    loadNamespaces(),
    loadStorageClasses(),
    loadPVCs()
  ])
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.pvcs-container {
  padding: 20px;
  width: 100%;
  box-sizing: border-box;
}

.header-card {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 20px;
}

.title-section h2 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.subtitle {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.action-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex-shrink: 0;
  min-width: 360px;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.action-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: nowrap;
  min-width: 0;
}

.auto-refresh-container {
  display: flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
}

.namespace-select {
  min-width: 180px;
}

.storage-class-select {
  min-width: 180px;
}

.auto-refresh-switch {
  margin-left: 8px;
}

.auto-refresh-label {
  font-size: 14px;
  color: #606266;
}

.table-card {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  overflow-x: auto;
}

.responsive-table {
  width: 100%;
  min-width: 800px;

  /* 确保表格内容完全显示 */
  .el-table__header-wrapper,
  .el-table__body-wrapper {
    overflow-x: auto;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.name-cell {
  display: flex;
  align-items: center;
}

.access-modes {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.no-data {
  color: #c0c4cc;
  font-style: italic;
}

.code-text {
  font-family: 'Monaco', 'Consolas', 'Courier New', monospace;
  font-size: 12px;
  background-color: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  color: #e83e8c;
  border: 1px solid #e8e8e8;
}

.mt-4 {
  margin-top: 16px;
}

.detail-content {
  max-height: 70vh;
  overflow-y: auto;
}

.access-modes-detail {
  margin: 16px 0;
}

.labels-section, .annotations-section {
  margin: 16px 0;
  min-height: 32px;
}

.annotation-item {
  margin-bottom: 8px;
  padding: 8px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

/* PV名称样式 */
.pv-name {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  word-break: break-all;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .header-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .action-section {
    width: 100%;
    min-width: unset;
  }

  .filter-row,
  .action-row {
    justify-content: flex-start;
  }
}

@media (max-width: 992px) {
  .namespace-select,
  .storage-class-select {
    min-width: 160px;
  }

  .responsive-table {
    font-size: 14px;
  }

  .el-tag {
    font-size: 12px;
  }
}

@media (max-width: 768px) {
  .pvcs-container {
    padding: 12px;
  }

  .header-content {
    gap: 12px;
  }

  .filter-row,
  .action-row {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    flex-wrap: wrap;
  }

  .namespace-select,
  .storage-class-select {
    width: 100%;
    min-width: unset;
  }

  .auto-refresh-container {
    align-self: flex-start;
  }

  .responsive-table {
    font-size: 12px;
  }

  .el-table .el-table__cell {
    padding: 8px 4px;
  }

  .access-modes .el-tag {
    margin-right: 2px;
    margin-bottom: 2px;
  }
}

@media (max-width: 576px) {
  .title-section h2 {
    font-size: 20px;
  }

  .subtitle {
    font-size: 12px;
  }

  .card-header {
    font-size: 14px;
  }

  .el-table .el-table__cell {
    padding: 6px 2px;
  }

  .el-button--small {
    padding: 4px 8px;
    font-size: 12px;
  }
}
</style>
