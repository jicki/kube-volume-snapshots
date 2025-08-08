<template>
  <div class="scheduled-tasks">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>定时快照任务管理</span>
          <div class="header-actions">
            <el-select
              v-model="filterCluster"
              placeholder="过滤集群"
              style="width: 200px; margin-right: 10px;"
              clearable
              @change="handleClusterFilter"
            >
              <el-option label="所有集群" value="" />
              <el-option
                v-for="cluster in availableClusters"
                :key="cluster"
                :label="cluster"
                :value="cluster"
              />
            </el-select>
            <el-button type="success" @click="showCreateDialog">
              <el-icon><Plus /></el-icon>
              创建定时任务
            </el-button>
            <el-button type="primary" @click="loadTasks" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="filteredTasks"
        style="width: 100%"
        v-loading="loading"
        stripe
        table-layout="auto"
        :scroll-x="true"
      >
        <el-table-column prop="name" label="任务名称" min-width="120">
          <template #default="scope">
            <el-tag type="info">{{ scope.row.name }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="目标集群" min-width="140">
          <template #default="scope">
            <div class="cluster-tags" v-if="scope.row.targetClusters && scope.row.targetClusters.length > 0">
              <el-tag
                v-for="cluster in scope.row.targetClusters"
                :key="cluster"
                type="primary"
                size="small"
                :title="cluster"
              >
                {{ cluster }}
              </el-tag>
            </div>
            <el-tag v-else type="warning" size="small">未配置</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="namespace" label="命名空间" min-width="100" />

        <el-table-column prop="pvcName" label="源 PVC" min-width="120" />

        <el-table-column prop="volumeSnapshotClassName" label="快照类" min-width="140">
          <template #default="scope">
            <el-tag type="success" size="small">{{ scope.row.volumeSnapshotClassName }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="cronExpression" label="Cron 表达式" min-width="140">
          <template #default="scope">
            <el-tag type="warning" size="small">{{ scope.row.cronExpression }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="状态" min-width="80">
          <template #default="scope">
            <el-switch
              v-model="scope.row.enabled"
              @change="toggleTask(scope.row)"
              :loading="scope.row.toggling"
            />
          </template>
        </el-table-column>

        <el-table-column label="上次执行" min-width="140">
          <template #default="scope">
            <span v-if="scope.row.lastExecuted">
              {{ formatTime(scope.row.lastExecuted) }}
            </span>
            <span v-else class="no-data">未执行</span>
          </template>
        </el-table-column>

        <el-table-column label="下次执行" min-width="140">
          <template #default="scope">
            <span v-if="scope.row.nextExecution && scope.row.enabled">
              {{ formatTime(scope.row.nextExecution) }}
            </span>
            <span v-else class="no-data">-</span>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" min-width="140">
          <template #default="scope">
            {{ formatTime(scope.row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="160" fixed="right">
          <template #default="scope">
            <el-button
              size="small"
              @click="editTask(scope.row)"
            >
              编辑
            </el-button>
            <el-button
              size="small"
              type="danger"
              @click="confirmDelete(scope.row)"
              :loading="scope.row.deleting"
              :disabled="scope.row.deleting"
            >
              {{ scope.row.deleting ? '删除中...' : '删除' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑定时任务' : '创建定时任务'"
      width="600px"
    >
      <el-form
        :model="form"
        :rules="formRules"
        ref="formRef"
        label-width="120px"
      >
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称" />
        </el-form-item>

        <el-form-item label="目标集群" prop="targetClusters" required>
          <el-select
            v-model="form.targetClusters"
            placeholder="请选择目标集群（可多选）"
            style="width: 100%"
            multiple
            filterable
          >
            <el-option
              v-for="cluster in clusters"
              :key="cluster.name"
              :label="`${cluster.display_name} (${cluster.name})`"
              :value="cluster.name"
              :disabled="!cluster.enabled || cluster.status !== 'online'"
            >
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>{{ cluster.display_name || cluster.name }}</span>
                <el-tag
                  :type="cluster.status === 'online' ? 'success' : cluster.status === 'offline' ? 'danger' : 'warning'"
                  size="small"
                >
                  {{ cluster.status }}
                </el-tag>
              </div>
            </el-option>
          </el-select>
          <div style="margin-top: 5px; font-size: 12px; color: #909399;">
            必须选择至少一个目标集群。可以选择多个集群同时执行快照任务。
          </div>
        </el-form-item>

        <el-form-item label="命名空间" prop="namespace">
          <el-select
            v-model="form.namespace"
            placeholder="选择命名空间"
            style="width: 100%"
            filterable
            clearable
            @change="onNamespaceChange"
          >
            <el-option
              v-for="ns in namespaces"
              :key="ns.value"
              :label="ns.label"
              :value="ns.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="快照类" prop="volumeSnapshotClassName">
          <el-select
            v-model="form.volumeSnapshotClassName"
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
            v-model="form.pvcName"
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

        <el-form-item label="执行频率" prop="scheduleType">
          <el-select
            v-model="form.scheduleType"
            placeholder="选择执行频率"
            style="width: 100%"
            @change="form.advancedMode = form.scheduleType === 'custom'"
          >
            <el-option
              v-for="option in scheduleTypeOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>

        <!-- 执行时间选择 -->
        <el-form-item
          v-if="form.scheduleType !== 'custom'"
          label="执行时间"
          prop="scheduleTime"
        >
          <el-time-picker
            v-model="form.scheduleTime"
            format="HH:mm"
            value-format="HH:mm"
            placeholder="选择执行时间"
            style="width: 100%"
          />
        </el-form-item>

        <!-- 星期选择 -->
        <el-form-item
          v-if="form.scheduleType === 'weekly'"
          label="执行星期"
        >
          <el-select
            v-model="form.scheduleWeekday"
            placeholder="选择星期"
            style="width: 100%"
          >
            <el-option
              v-for="day in weekdayOptions"
              :key="day.value"
              :label="day.label"
              :value="day.value"
            />
          </el-select>
        </el-form-item>

        <!-- 日期选择 -->
        <el-form-item
          v-if="form.scheduleType === 'monthly'"
          label="执行日期"
        >
          <el-select
            v-model="form.scheduleDate"
            placeholder="选择日期"
            style="width: 100%"
          >
            <el-option
              v-for="date in 31"
              :key="date"
              :label="`每月${date}号`"
              :value="date"
            />
          </el-select>
        </el-form-item>

        <!-- 自定义 Cron 表达式 -->
        <el-form-item
          v-if="form.scheduleType === 'custom'"
          label="Cron 表达式"
          prop="cronExpression"
        >
          <el-input
            v-model="form.cronExpression"
            placeholder="例如: 0 0 2 * * * (每天凌晨2点)"
          />
          <div style="margin-top: 5px; font-size: 12px; color: #909399;">
            <p>格式：秒 分 时 日 月 周 (6个字段)</p>
            <p>• 每天 2:00: 0 0 2 * * *</p>
            <p>• 每小时: 0 0 * * * *</p>
            <p>• 每周日 3:00: 0 0 3 * * 0</p>
          </div>
        </el-form-item>

        <!-- 预览生成的 Cron 表达式 -->
        <el-form-item v-if="form.scheduleType !== 'custom'" label="预览">
          <el-input
            :value="generateCronExpression()"
            readonly
            placeholder="自动生成的 Cron 表达式"
          />
          <div style="margin-top: 5px; font-size: 12px; color: #909399;">
            根据您的选择自动生成的 Cron 表达式
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">
          {{ isEditing ? '更新' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import {
  getScheduledSnapshots,
  createScheduledSnapshot,
  updateScheduledSnapshot,
  deleteScheduledSnapshot,
  toggleScheduledSnapshot,
  getVolumeSnapshotClasses,
  getPVCs,
  getNamespaces,
  getClusters
} from '../api'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const tasks = ref([])
const snapshotClasses = ref([])
const pvcs = ref([])
const namespaces = ref([])
const clusters = ref([])
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const pvcLoading = ref(false)
const isEditing = ref(false)
const editingId = ref('')
const filterCluster = ref('')

const form = reactive({
  name: '',
  namespace: 'default',
  pvcName: '',
  volumeSnapshotClassName: '',
  cronExpression: '',
  targetClusters: [], // 新增目标集群数组
  // 新增定时选择相关字段
  scheduleType: 'daily', // daily, weekly, monthly, custom
  scheduleTime: '02:00', // HH:mm 格式
  scheduleWeekday: 1, // 1-7 (周一到周日)
  scheduleDate: 1, // 1-31 (每月第几天)
  advancedMode: false // 是否显示高级模式
})

const formRules = computed(() => ({
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  targetClusters: [{ required: true, message: '请选择目标集群', trigger: 'change' }],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  pvcName: [{ required: true, message: '请选择源 PVC', trigger: 'change' }],
  volumeSnapshotClassName: [{ required: true, message: '请选择快照类', trigger: 'change' }],
  scheduleTime: form.scheduleType !== 'custom' ? [{ required: true, message: '请选择执行时间', trigger: 'change' }] : [],
  cronExpression: form.scheduleType === 'custom' ? [{ required: true, message: '请输入 Cron 表达式', trigger: 'blur' }] : []
}))

const formRef = ref()

// 计算属性：过滤后的任务列表
const filteredTasks = computed(() => {
  if (!filterCluster.value) {
    return tasks.value
  }

  return tasks.value.filter(task => {
    // 由于目标集群现在是必填的，所有任务都应该有targetClusters
    if (task.targetClusters && task.targetClusters.length > 0) {
      return task.targetClusters.includes(filterCluster.value)
    }

    // 对于旧的没有目标集群的任务，不在任何过滤器中显示
    return false
  })
})

// 计算属性：可用的集群过滤选项
const availableClusters = computed(() => {
  const clusterSet = new Set()

  // 从集群列表中获取所有可用的集群
  clusters.value.forEach(cluster => {
    if (cluster.name) {
      clusterSet.add(cluster.name)
    }
  })

  // 也从现有任务中获取集群（以防某些集群已被删除但任务仍存在）
  tasks.value.forEach(task => {
    if (task.targetClusters && task.targetClusters.length > 0) {
      task.targetClusters.forEach(cluster => clusterSet.add(cluster))
    }
  })

  return Array.from(clusterSet)
})

// 定时类型选项
const scheduleTypeOptions = [
  { label: '每天', value: 'daily' },
  { label: '每周', value: 'weekly' },
  { label: '每月', value: 'monthly' },
  { label: '自定义', value: 'custom' }
]

// 星期选项
const weekdayOptions = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
  { label: '周日', value: 0 }
]

// 将定时选择转换为 cron 表达式
const generateCronExpression = () => {
  const [hour, minute] = form.scheduleTime.split(':').map(Number)

  switch (form.scheduleType) {
    case 'daily':
      return `0 ${minute} ${hour} * * *`
    case 'weekly':
      return `0 ${minute} ${hour} * * ${form.scheduleWeekday}`
    case 'monthly':
      return `0 ${minute} ${hour} ${form.scheduleDate} * *`
    case 'custom':
    default:
      return form.cronExpression
  }
}

// 从 cron 表达式解析出定时选择
const parseCronExpression = (cronExpr) => {
  if (!cronExpr) return

  const parts = cronExpr.split(' ')
  if (parts.length !== 6) return

  const [, min, hour, date, month, weekday] = parts

  // 设置时间
  form.scheduleTime = `${hour.padStart(2, '0')}:${min.padStart(2, '0')}`

  // 判断类型
  if (date !== '*' && month === '*' && weekday === '*') {
    form.scheduleType = 'monthly'
    form.scheduleDate = parseInt(date)
  } else if (date === '*' && month === '*' && weekday !== '*') {
    form.scheduleType = 'weekly'
    form.scheduleWeekday = parseInt(weekday)
  } else if (date === '*' && month === '*' && weekday === '*') {
    form.scheduleType = 'daily'
  } else {
    form.scheduleType = 'custom'
    form.advancedMode = true
  }
}

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN')
}

const loadTasks = async () => {
  loading.value = true
  try {
    tasks.value = await getScheduledSnapshots()
    // 为每个任务添加状态标记
    tasks.value.forEach(task => {
      task.toggling = false
      task.deleting = false
    })
  } catch (error) {
    ElMessage.error('加载定时任务失败: ' + error.message)
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
  if (!form.namespace) {
    console.log('ScheduledTasks - loadPVCs: namespace is empty')
    pvcs.value = []
    return
  }

  console.log('ScheduledTasks - loadPVCs: loading PVCs for namespace:', form.namespace)
  pvcLoading.value = true
  try {
    const data = await getPVCs(form.namespace)
    console.log('ScheduledTasks - loadPVCs: received data:', data)
    // 适配新的数据结构: PVCWithPVInfo[] -> 扁平化处理
    pvcs.value = (data || []).map(item => ({
      ...item.pvc,
      volumeAttributes: item.volumeAttributes
    }))
    console.log('ScheduledTasks - loadPVCs: processed pvcs:', pvcs.value)
  } catch (error) {
    console.error('ScheduledTasks - loadPVCs: error loading PVCs:', error)
    ElMessage.error('加载 PVC 失赅: ' + error.message)
    pvcs.value = []
  } finally {
    pvcLoading.value = false
  }
}

const onNamespaceChange = () => {
  // 命名空间变化时清空已选择的PVC，并重新加载PVC列表
  form.pvcName = ''
  pvcs.value = []
  if (form.namespace) {
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
  } catch (error) {
    ElMessage.error('加载命名空间失败: ' + error.message)
  }
}

const loadClusters = async () => {
  try {
    const response = await getClusters()
    // 后端返回的数据结构: { clusters: [...], current: "..." }
    clusters.value = response.clusters || []
    console.log('ScheduledTasks - loadClusters: loaded clusters:', clusters.value)
  } catch (error) {
    console.error('ScheduledTasks - loadClusters: error:', error)
    ElMessage.error('加载集群信息失败: ' + error.message)
  }
}

const handleClusterFilter = (value) => {
  filterCluster.value = value
  console.log('ScheduledTasks - handleClusterFilter: selected filter:', value)
}

const showCreateDialog = () => {
  console.log('ScheduledTasks - showCreateDialog: opening create dialog')
  // 检查认证状态
  const authStore = useAuthStore()
  console.log('ScheduledTasks - showCreateDialog: auth token exists:', !!authStore.token)
  console.log('ScheduledTasks - showCreateDialog: user:', authStore.user)

  isEditing.value = false
  Object.assign(form, {
    name: '',
    namespace: 'default',
    pvcName: '',
    volumeSnapshotClassName: '',
    cronExpression: '',
    targetClusters: [],
    scheduleType: 'daily',
    scheduleTime: '02:00',
    scheduleWeekday: 1,
    scheduleDate: 1,
    advancedMode: false
  })
  // 重置PVC列表
  pvcs.value = []
  dialogVisible.value = true
  // 预加载默认命名空间的PVC
  loadPVCs()
}

const editTask = (task) => {
  isEditing.value = true
  editingId.value = task.id
  Object.assign(form, {
    name: task.name,
    namespace: task.namespace,
    pvcName: task.pvcName,
    volumeSnapshotClassName: task.volumeSnapshotClassName,
    cronExpression: task.cronExpression,
    targetClusters: task.targetClusters || [],
    scheduleType: 'daily',
    scheduleTime: '02:00',
    scheduleWeekday: 1,
    scheduleDate: 1,
    advancedMode: false
  })

  // 解析现有的 cron 表达式
  parseCronExpression(task.cronExpression)

  // 重置PVC列表并加载对应命名空间的PVC
  pvcs.value = []
  if (form.namespace) {
    loadPVCs()
  }

  dialogVisible.value = true
}

const submitForm = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        // 根据用户选择生成 cron 表达式
        const finalCronExpression = generateCronExpression()

        const submitData = {
          name: form.name,
          namespace: form.namespace,
          pvcName: form.pvcName,
          volumeSnapshotClassName: form.volumeSnapshotClassName,
          cronExpression: finalCronExpression,
          targetClusters: form.targetClusters // 现在目标集群是必填的，不再需要判断
        }

        if (isEditing.value) {
          await updateScheduledSnapshot(editingId.value, submitData)
          ElMessage.success('定时任务更新成功，正在刷新列表...')
        } else {
          await createScheduledSnapshot(submitData)
          ElMessage.success('定时任务创建成功，正在刷新列表...')
        }
        dialogVisible.value = false
        // 延迟刷新确保任务状态同步完成
        setTimeout(() => {
          loadTasks()
        }, 1000)
      } catch (error) {
        ElMessage.error(`${isEditing.value ? '更新' : '创建'}定时任务失败: ` + error.message)
      } finally {
        submitting.value = false
      }
    }
  })
}

const toggleTask = async (task) => {
  task.toggling = true
  try {
    await toggleScheduledSnapshot(task.id)
    ElMessage.success(`定时任务已${task.enabled ? '启用' : '禁用'}，正在刷新状态...`)
    // 延迟刷新确保任务状态同步完成
    setTimeout(() => {
      loadTasks()
    }, 1000)
  } catch (error) {
    // 恢复原状态
    task.enabled = !task.enabled
    ElMessage.error('切换任务状态失败: ' + error.message)
  } finally {
    task.toggling = false
  }
}

const confirmDelete = async (task) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除定时任务 "${task.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 添加删除状态标记
    task.deleting = true

    await deleteScheduledSnapshot(task.id)
    ElMessage.success('定时任务删除成功，正在刷新列表...')

    // 延迟刷新确保任务状态同步完成
    setTimeout(() => {
      loadTasks()
    }, 1000)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除定时任务失败: ' + error.message)
    }
    // 移除删除状态标记
    task.deleting = false
  }
}

onMounted(() => {
  console.log('ScheduledTasks - onMounted: initializing component')
  loadTasks()
  loadSnapshotClasses()
  loadNamespaces()
  loadClusters()
})
</script>

<style scoped>
.scheduled-tasks {
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

/* 集群标签样式优化 */
.cluster-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
  align-items: center;
}

.cluster-tags .el-tag {
  margin: 1px;
  max-width: 100px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 响应式布局优化 */
@media (max-width: 1200px) {
  .el-table {
    font-size: 13px;
  }

  .el-table .cell {
    padding: 0 8px;
  }

  .el-tag {
    font-size: 12px;
    padding: 2px 6px;
    height: auto;
    line-height: 1.2;
  }
}

@media (max-width: 992px) {
  .el-table {
    font-size: 12px;
  }

  .el-table .cell {
    padding: 0 6px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .el-tag {
    font-size: 11px;
    padding: 1px 4px;
  }

  .el-button {
    font-size: 12px;
    padding: 6px 10px;
  }
}

@media (max-width: 768px) {
  .scheduled-tasks {
    padding: 10px;
  }

  .card-header {
    flex-direction: column;
    gap: 10px;
  }

  .header-actions {
    width: 100%;
    justify-content: center;
    flex-wrap: wrap;
    gap: 8px;
  }

  .header-actions .el-select {
    width: 150px !important;
    margin-right: 0 !important;
  }

  .el-table {
    font-size: 11px;
  }

  .el-table .cell {
    padding: 0 4px;
  }

  .el-tag {
    font-size: 10px;
    padding: 1px 3px;
  }

  .el-button {
    font-size: 11px;
    padding: 4px 8px;
  }

  .el-switch {
    transform: scale(0.8);
  }
}
</style>
