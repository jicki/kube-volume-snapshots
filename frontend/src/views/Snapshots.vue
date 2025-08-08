<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div class="snapshots-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>VolumeSnapshot 管理</span>
          <div class="header-actions">
            <el-select
              v-model="selectedNamespace"
              placeholder="选择命名空间"
              style="width: 200px; margin-right: 10px;"
              @change="loadSnapshots"
            >
              <el-option
                v-for="ns in namespaces"
                :key="ns.value"
                :label="ns.label"
                :value="ns.value"
              />
            </el-select>
            <el-button type="success" @click="showCreateDialog">
              <el-icon><Plus /></el-icon>
              创建快照
            </el-button>
            <el-button type="primary" @click="loadSnapshots" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
            <el-button
              :type="autoRefreshEnabled ? 'success' : 'info'"
              @click="autoRefreshEnabled = !autoRefreshEnabled; toggleAutoRefresh()"
              size="small"
            >
              <el-icon><Timer /></el-icon>
              {{ autoRefreshEnabled ? '自动刷新:开' : '自动刷新:关' }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="snapshots"
        style="width: 100%"
        v-loading="loading"
        stripe
        table-layout="auto"
      >
        <el-table-column prop="volumeSnapshot.metadata.name" label="名称" min-width="160">
          <template #default="scope">
            <el-tag type="info">{{ scope.row.volumeSnapshot.metadata.name }}</el-tag>
          </template>
        </el-table-column>

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

        <el-table-column label="源 PVC" min-width="120">
          <template #default="scope">
            <span v-if="scope.row.volumeSnapshot.spec.source.persistentVolumeClaimName">
              {{ scope.row.volumeSnapshot.spec.source.persistentVolumeClaimName }}
            </span>
            <span v-else class="no-data">-</span>
          </template>
        </el-table-column>

        <el-table-column label="快照类" min-width="140">
          <template #default="scope">
            <el-tag
              v-if="scope.row.volumeSnapshot.spec.volumeSnapshotClassName"
              type="success"
              size="small"
            >
              {{ scope.row.volumeSnapshot.spec.volumeSnapshotClassName }}
            </el-tag>
            <span v-else class="no-data">-</span>
          </template>
        </el-table-column>

        <el-table-column label="大小" min-width="80">
          <template #default="scope">
            <span v-if="scope.row.volumeSnapshot.status?.restoreSize">
              {{ formatSize(scope.row.volumeSnapshot.status.restoreSize) }}
            </span>
            <span v-else class="no-data">-</span>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" min-width="140">
          <template #default="scope">
            {{ formatTime(scope.row.volumeSnapshot.metadata.creationTimestamp) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button
              size="small"
              @click="viewDetails(scope.row)"
            >
              详情
            </el-button>
            <el-button
              v-if="!isSnapshotStuck(scope.row)"
              size="small"
              type="danger"
              @click="confirmDelete(scope.row)"
              :loading="scope.row.deleting"
              :disabled="scope.row.deleting"
            >
              {{ scope.row.deleting ? '删除中...' : '删除' }}
            </el-button>
            <el-button
              v-if="isSnapshotStuck(scope.row)"
              size="small"
              type="warning"
              @click="confirmForceDelete(scope.row)"
              :loading="scope.row.forceDeleting"
              :disabled="scope.row.forceDeleting"
            >
              {{ scope.row.forceDeleting ? '强制删除中...' : '强制删除' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建快照对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="创建 VolumeSnapshot"
      width="500px"
    >
      <el-form
        :model="createForm"
        :rules="createRules"
        ref="createFormRef"
        label-width="120px"
      >
        <el-form-item label="快照名称" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入快照名称" />
        </el-form-item>

        <el-form-item label="命名空间" prop="namespace">
          <el-select
            v-model="createForm.namespace"
            placeholder="选择命名空间"
            style="width: 100%"
            filterable
            clearable
            @change="onNamespaceChange"
          >
            <el-option
              v-for="ns in namespaces.filter(ns => ns.value !== 'all')"
              :key="ns.value"
              :label="ns.label"
              :value="ns.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="快照类" prop="volumeSnapshotClassName">
          <el-select
            v-model="createForm.volumeSnapshotClassName"
            placeholder="选择快照类"
            style="width: 100%"
            filterable
            clearable
          >
            <el-option
              v-for="vsc in snapshotClasses"
              :key="vsc.volumeSnapshotClass.metadata.name"
              :label="vsc.volumeSnapshotClass.metadata.name"
              :value="vsc.volumeSnapshotClass.metadata.name"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="源 PVC" prop="pvcName">
          <el-select
            v-model="createForm.pvcName"
            placeholder="选择 PVC"
            style="width: 100%"
            filterable
            clearable
            :loading="pvcLoading"
            @focus="loadPVCs"
          >
            <el-option
              v-for="pvc in pvcs"
              :key="pvc.metadata?.name || 'unknown'"
              :label="`${pvc.metadata?.name || 'Unknown PVC'} (${pvc.spec?.resources?.requests?.storage || 'Unknown Size'})`"
              :value="pvc.metadata?.name || ''"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="createSnapshot">确定</el-button>
      </template>
    </el-dialog>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="VolumeSnapshot 详情"
      width="70%"
    >
      <div v-if="selectedSnapshot">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">
            {{ selectedSnapshot.volumeSnapshot.metadata.name }}
          </el-descriptions-item>
          <el-descriptions-item label="命名空间">
            {{ selectedSnapshot.volumeSnapshot.metadata.namespace }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getSnapshotStatusType(selectedSnapshot.volumeSnapshot)">
              {{ getSnapshotStatus(selectedSnapshot.volumeSnapshot) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="源 PVC">
            {{ selectedSnapshot.volumeSnapshot.spec.source.persistentVolumeClaimName || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="快照类">
            {{ selectedSnapshot.volumeSnapshot.spec.volumeSnapshotClassName || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="大小">
            {{ selectedSnapshot.volumeSnapshot.status?.restoreSize ? formatSize(selectedSnapshot.volumeSnapshot.status.restoreSize) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(selectedSnapshot.volumeSnapshot.metadata.creationTimestamp) }}
          </el-descriptions-item>
          <el-descriptions-item label="快照内容">
            {{ selectedSnapshot.volumeSnapshot.status?.boundVolumeSnapshotContentName || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="selectedSnapshot.volumeSnapshotContent" style="margin-top: 20px;">
          <h4>VolumeSnapshotContent 信息</h4>
          <el-descriptions :column="2" border size="small" style="margin-top: 10px;">
            <el-descriptions-item label="内容名称">
              {{ selectedSnapshot.volumeSnapshotContent.metadata.name }}
            </el-descriptions-item>
            <el-descriptions-item label="删除策略">
              {{ selectedSnapshot.volumeSnapshotContent.spec.deletionPolicy }}
            </el-descriptions-item>
            <el-descriptions-item label="快照句柄">
              {{ selectedSnapshot.volumeSnapshotContent.status?.snapshotHandle || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="准备时间">
              {{ selectedSnapshot.volumeSnapshotContent.status?.creationTime ? formatTime(selectedSnapshot.volumeSnapshotContent.status.creationTime) : '-' }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>

      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { getVolumeSnapshots, createVolumeSnapshot, deleteVolumeSnapshot, forceDeleteVolumeSnapshot, getVolumeSnapshotClasses, getPVCs, getNamespaces } from '../api'
import { Plus, Refresh, Timer } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const snapshots = ref([])
const snapshotClasses = ref([])
const pvcs = ref([])
const namespaces = ref([])
const loading = ref(false)
const creating = ref(false)
const pvcLoading = ref(false)
const selectedNamespace = ref('all')
const autoRefreshTimer = ref(null)
const autoRefreshEnabled = ref(false)

const createDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const selectedSnapshot = ref(null)

const createForm = reactive({
  name: '',
  namespace: 'default',
  pvcName: '',
  volumeSnapshotClassName: ''
})

const createRules = {
  name: [{ required: true, message: '请输入快照名称', trigger: 'blur' }],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  pvcName: [{ required: true, message: '请选择源 PVC', trigger: 'change' }],
  volumeSnapshotClassName: [{ required: true, message: '请选择快照类', trigger: 'change' }]
}

const createFormRef = ref()

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

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN')
}

const formatSize = (size) => {
  if (!size) return '-'
  const units = ['B', 'Ki', 'Mi', 'Gi', 'Ti']
  let value = parseInt(size)
  let unitIndex = 0

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }

  return `${value.toFixed(1)}${units[unitIndex]}`
}

const loadSnapshots = async () => {
  loading.value = true
  try {
    snapshots.value = await getVolumeSnapshots(selectedNamespace.value)
  } catch (error) {
    ElMessage.error('加载快照失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const loadSnapshotClasses = async () => {
  try {
    snapshotClasses.value = await getVolumeSnapshotClasses()
  } catch (error) {
    ElMessage.error('加载快照类失败: ' + error.message)
  }
}

const loadPVCs = async () => {
  if (!createForm.namespace) {
    console.log('Snapshots - loadPVCs: namespace is empty')
    pvcs.value = []
    return
  }

  console.log('Snapshots - loadPVCs: loading PVCs for namespace:', createForm.namespace)
  pvcLoading.value = true
  try {
    const data = await getPVCs(createForm.namespace)
    console.log('Snapshots - loadPVCs: received data:', data)
    // 适配新的数据结构: PVCWithPVInfo[] -> 扁平化处理
    pvcs.value = (data || []).map(item => ({
      ...item.pvc,
      volumeAttributes: item.volumeAttributes
    }))
    console.log('Snapshots - loadPVCs: processed pvcs:', pvcs.value)
  } catch (error) {
    console.error('Snapshots - loadPVCs: error loading PVCs:', error)
    ElMessage.error('加载 PVC 失败: ' + error.message)
    pvcs.value = []
  } finally {
    pvcLoading.value = false
  }
}

const onNamespaceChange = () => {
  // 命名空间变化时清空已选择的PVC，并重新加载PVC列表
  createForm.pvcName = ''
  pvcs.value = []
  if (createForm.namespace) {
    loadPVCs()
  }
}

const loadNamespaces = async () => {
  try {
    const namespacesData = await getNamespaces()
    namespaces.value = namespacesData.map(ns => ({
      label: ns.metadata.name,
      value: ns.metadata.name
    }))
    // 添加"所有命名空间"选项到头部
    namespaces.value.unshift({ label: '所有命名空间', value: 'all' })
  } catch (error) {
    ElMessage.error('加载命名空间失败: ' + error.message)
  }
}

const showCreateDialog = () => {
  console.log('Snapshots - showCreateDialog: opening create dialog')
  // 检查认证状态
  const authStore = useAuthStore()
  console.log('Snapshots - showCreateDialog: auth token exists:', !!authStore.token)
  console.log('Snapshots - showCreateDialog: user:', authStore.user)

  Object.assign(createForm, {
    name: '',
    namespace: 'default',
    pvcName: '',
    volumeSnapshotClassName: ''
  })
  // 重置PVC列表
  pvcs.value = []
  createDialogVisible.value = true
  // 预加载默认命名空间的PVC
  loadPVCs()
}

const createSnapshot = async () => {
  if (!createFormRef.value) return

  await createFormRef.value.validate(async (valid) => {
    if (valid) {
      creating.value = true
      try {
        await createVolumeSnapshot(createForm)
        ElMessage.success('快照创建成功，正在刷新列表...')
        createDialogVisible.value = false
        // 延迟刷新确保K8s资源状态同步完成
        setTimeout(() => {
          loadSnapshots()
        }, 2000)
      } catch (error) {
        ElMessage.error('创建快照失败: ' + error.message)
      } finally {
        creating.value = false
      }
    }
  })
}

const viewDetails = (snapshot) => {
  selectedSnapshot.value = snapshot
  detailDialogVisible.value = true
}

// 检查快照是否卡住（有删除时间戳但仍然存在，或状态异常）
const isSnapshotStuck = (snapshot) => {
  const vs = snapshot.volumeSnapshot
  // 检查是否有 deletionTimestamp 但仍然存在
  if (vs.metadata.deletionTimestamp) {
    return true
  }
  // 检查是否 readyToUse 为 false 并且有错误状态
  if (vs.status && vs.status.readyToUse === false && vs.status.error) {
    return true
  }
  return false
}

const confirmDelete = async (snapshot) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除快照 "${snapshot.volumeSnapshot.metadata.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 添加删除状态标记
    snapshot.deleting = true

    await deleteVolumeSnapshot(
      snapshot.volumeSnapshot.metadata.namespace,
      snapshot.volumeSnapshot.metadata.name
    )
    ElMessage.success('快照删除成功，正在刷新列表...')

    // 延迟刷新确保K8s资源状态同步完成
    setTimeout(() => {
      loadSnapshots()
    }, 2000)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除快照失败: ' + error.message)
    }
    // 移除删除状态标记
    snapshot.deleting = false
  }
}

const confirmForceDelete = async (snapshot) => {
  try {
    await ElMessageBox.confirm(
      `快照 "${snapshot.volumeSnapshot.metadata.name}" 似乎卡在删除状态。确定要强制删除吗？这将清理所有保护性标记。`,
      '确认强制删除',
      {
        confirmButtonText: '强制删除',
        cancelButtonText: '取消',
        type: 'warning',
        dangerouslyUseHTMLString: true
      }
    )

    // 添加强制删除状态标记
    snapshot.forceDeleting = true

    await forceDeleteVolumeSnapshot(
      snapshot.volumeSnapshot.metadata.namespace,
      snapshot.volumeSnapshot.metadata.name
    )
    ElMessage.success('快照强制删除成功，正在刷新列表...')

    // 延迟刷新确保K8s资源状态同步完成
    setTimeout(() => {
      loadSnapshots()
    }, 2000)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('强制删除快照失败: ' + error.message)
    }
    // 移除强制删除状态标记
    snapshot.forceDeleting = false
  }
}

// 启用/禁用自动刷新
const toggleAutoRefresh = () => {
  if (autoRefreshEnabled.value) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

// 开始自动刷新
const startAutoRefresh = () => {
  if (autoRefreshTimer.value) {
    clearInterval(autoRefreshTimer.value)
  }
  autoRefreshTimer.value = setInterval(() => {
    if (!loading.value) {
      loadSnapshots()
    }
  }, 30000) // 每30秒刷新一次
}

// 停止自动刷新
const stopAutoRefresh = () => {
  if (autoRefreshTimer.value) {
    clearInterval(autoRefreshTimer.value)
    autoRefreshTimer.value = null
  }
}

onMounted(() => {
  loadSnapshots()
  loadSnapshotClasses()
  loadNamespaces()
  // 默认启用自动刷新
  autoRefreshEnabled.value = true
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.snapshots-view {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
}

.no-data {
  color: #c0c4cc;
  font-style: italic;
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

h4 {
  color: #303133;
  margin: 0;
}
</style>
