<template>
  <div class="snapshot-classes">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>VolumeSnapshotClass 管理</span>
          <el-button type="primary" @click="loadData">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-table
        :data="snapshotClasses"
        style="width: 100%"
        v-loading="loading"
        stripe
      >
        <el-table-column prop="volumeSnapshotClass.metadata.name" label="名称" width="200">
          <template #default="scope">
            <el-tag type="info">{{ scope.row.volumeSnapshotClass?.metadata?.name || '-' }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="CSI 驱动" width="300">
          <template #default="scope">
            {{ scope.row.volumeSnapshotClass?.driver || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="关联存储类" width="200">
          <template #default="scope">
            <el-tag
              v-if="scope.row.relatedStorageClass"
              type="success"
            >
              {{ scope.row.relatedStorageClass.metadata.name }}
            </el-tag>
            <span v-else class="no-data">未找到</span>
          </template>
        </el-table-column>

        <el-table-column label="删除策略" width="120">
          <template #default="scope">
            <el-tag
              :type="scope.row.volumeSnapshotClass?.deletionPolicy === 'Delete' ? 'danger' : 'warning'"
              size="small"
            >
              {{ scope.row.volumeSnapshotClass?.deletionPolicy || '-' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="参数" min-width="300">
          <template #default="scope">
            <div v-if="scope.row.volumeSnapshotClass?.parameters">
              <el-tag
                v-for="(value, key) in scope.row.volumeSnapshotClass.parameters"
                :key="key"
                size="small"
                style="margin: 2px;"
              >
                {{ key }}: {{ value }}
              </el-tag>
            </div>
            <span v-else class="no-data">无参数</span>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.volumeSnapshotClass?.metadata?.creationTimestamp) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="150" fixed="right">
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

    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="VolumeSnapshotClass 详情"
      width="70%"
    >
      <div v-if="selectedItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">
            {{ selectedItem?.volumeSnapshotClass?.metadata?.name || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="CSI 驱动">
            {{ selectedItem?.volumeSnapshotClass?.driver || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="删除策略">
            <el-tag
              :type="selectedItem?.volumeSnapshotClass?.deletionPolicy === 'Delete' ? 'danger' : 'warning'"
            >
              {{ selectedItem?.volumeSnapshotClass?.deletionPolicy || '-' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(selectedItem?.volumeSnapshotClass?.metadata?.creationTimestamp) }}
          </el-descriptions-item>
        </el-descriptions>

        <div style="margin-top: 20px;">
          <h4>参数配置</h4>
          <el-table
            :data="getParametersArray(selectedItem?.volumeSnapshotClass?.parameters)"
            style="width: 100%; margin-top: 10px;"
            size="small"
          >
            <el-table-column prop="key" label="参数名" />
            <el-table-column prop="value" label="参数值" />
          </el-table>
        </div>

        <div v-if="selectedItem?.relatedStorageClass" style="margin-top: 20px;">
          <h4>关联存储类信息</h4>
          <el-descriptions :column="2" border size="small" style="margin-top: 10px;">
            <el-descriptions-item label="存储类名称">
              {{ selectedItem?.relatedStorageClass?.metadata?.name || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="供应商">
              {{ selectedItem?.relatedStorageClass?.provisioner || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="回收策略">
              {{ selectedItem?.relatedStorageClass?.reclaimPolicy || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="卷绑定模式">
              {{ selectedItem?.relatedStorageClass?.volumeBindingMode || '-' }}
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
import { ref, onMounted } from 'vue'
import { getVolumeSnapshotClasses } from '../api'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const snapshotClasses = ref([])
const loading = ref(false)
const detailDialogVisible = ref(false)
const selectedItem = ref(null)

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN')
}

const getParametersArray = (parameters) => {
  if (!parameters) return []
  return Object.entries(parameters).map(([key, value]) => ({ key, value }))
}

const viewDetails = (item) => {
  selectedItem.value = item
  detailDialogVisible.value = true
}

const loadData = async () => {
  loading.value = true
  try {
    const data = await getVolumeSnapshotClasses()
    snapshotClasses.value = Array.isArray(data) ? data : []
  } catch (error) {
    ElMessage.error('加载 VolumeSnapshotClass 失败: ' + error.message)
    snapshotClasses.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.snapshot-classes {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
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

.el-descriptions {
  margin-top: 10px;
}

h4 {
  color: #303133;
  margin: 0;
}
</style>
