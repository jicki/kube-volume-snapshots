<template>
  <div class="user-management">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-icon class="header-icon"><User /></el-icon>
            <span>用户管理</span>
          </div>
          <div class="header-right">
            <el-button
              type="primary"
              :icon="Plus"
              @click="showCreateDialog = true"
            >
              创建用户
            </el-button>
          </div>
        </div>
      </template>

      <!-- 用户列表 -->
      <el-table
        v-loading="loading"
        :data="users"
        stripe
        style="width: 100%"
        empty-text="暂无用户数据"
      >
        <el-table-column prop="username" label="用户名" min-width="120">
          <template #default="{ row }">
            <div class="username-cell">
              <el-icon class="user-icon"><Avatar /></el-icon>
              <span>{{ row.username }}</span>
              <el-tag v-if="row.username === 'admin'" type="warning" size="small">默认</el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'success'" size="small">
              {{ row.role === 'admin' ? '管理员' : '只读用户' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column prop="updatedAt" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updatedAt) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                type="primary"
                size="small"
                :icon="Edit"
                @click="handleResetPassword(row)"
                :disabled="row.username === currentUser?.username && !isAdmin"
              >
                重置密码
              </el-button>

              <el-button
                type="danger"
                size="small"
                :icon="Delete"
                @click="handleDeleteUser(row)"
                :disabled="row.username === 'admin' || row.username === currentUser?.username"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建用户对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建新用户"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="createForm.username"
            placeholder="请输入用户名（3-20个字符）"
            clearable
          />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="createForm.password"
            type="password"
            placeholder="请输入密码（至少6个字符）"
            show-password
            clearable
          />
        </el-form-item>

        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="createForm.confirmPassword"
            type="password"
            placeholder="请确认密码"
            show-password
            clearable
          />
        </el-form-item>

        <el-form-item label="角色" prop="role">
          <el-select v-model="createForm.role" placeholder="请选择用户角色" style="width: 100%">
            <el-option label="只读用户" value="readonly">
              <div class="role-option">
                <span>只读用户</span>
                <small>可查看所有资源，但不能创建、修改或删除</small>
              </div>
            </el-option>
            <el-option label="管理员" value="admin">
              <div class="role-option">
                <span>管理员</span>
                <small>拥有所有权限，可以管理用户和执行所有操作</small>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="handleCancelCreate">取消</el-button>
        <el-button
          type="primary"
          :loading="createLoading"
          @click="handleCreateUser"
        >
          {{ createLoading ? '创建中...' : '创建用户' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog
      v-model="showResetDialog"
      title="重置用户密码"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-alert
        title="重置密码"
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      >
        <template #default>
          确定要重置用户 <strong>{{ selectedUser?.username }}</strong> 的密码吗？
          <br>
          新密码将设置为：<strong>123456</strong>
        </template>
      </el-alert>

      <template #footer>
        <el-button @click="showResetDialog = false">取消</el-button>
        <el-button
          type="danger"
          :loading="resetLoading"
          @click="handleConfirmResetPassword"
        >
          {{ resetLoading ? '重置中...' : '确认重置' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { getAllUsers, deleteUser, register } from '@/api'
import { User, Plus, Avatar, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const authStore = useAuthStore()

// 响应式数据
const loading = ref(false)
const users = ref([])
const showCreateDialog = ref(false)
const showResetDialog = ref(false)
const createLoading = ref(false)
const resetLoading = ref(false)
const selectedUser = ref(null)

// 表单引用
const createFormRef = ref()

// 创建用户表单
const createForm = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  role: 'readonly'
})

// 计算属性
const currentUser = computed(() => authStore.currentUser)
const isAdmin = computed(() => authStore.isAdmin)

// 表单验证规则
const createRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度为3-20个字符', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (users.value.some(user => user.username === value)) {
          callback(new Error('用户名已存在'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== createForm.password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

// 格式化时间
const formatDateTime = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const data = await getAllUsers()
    users.value = data || []
  } catch (error) {
    console.error('Failed to load users:', error)
    users.value = []
  } finally {
    loading.value = false
  }
}

// 创建用户
const handleCreateUser = async () => {
  if (!createFormRef.value) return

  try {
    await createFormRef.value.validate()
    createLoading.value = true

    await register({
      username: createForm.username,
      password: createForm.password,
      role: createForm.role
    })

    ElMessage.success('用户创建成功')
    handleCancelCreate()
    loadUsers()
  } catch (error) {
    console.error('Failed to create user:', error)
  } finally {
    createLoading.value = false
  }
}

// 取消创建
const handleCancelCreate = () => {
  showCreateDialog.value = false
  createFormRef.value?.resetFields()
  Object.assign(createForm, {
    username: '',
    password: '',
    confirmPassword: '',
    role: 'readonly'
  })
}

// 重置密码
const handleResetPassword = (user) => {
  selectedUser.value = user
  showResetDialog.value = true
}

// 确认重置密码
const handleConfirmResetPassword = async () => {
  resetLoading.value = true
  try {
    // 注意：这里需要后端提供重置密码的API
    // 目前模拟重置为默认密码
    ElMessage.warning('重置密码功能需要后端API支持')
    showResetDialog.value = false
  } catch (error) {
    console.error('Failed to reset password:', error)
  } finally {
    resetLoading.value = false
  }
}

// 删除用户
const handleDeleteUser = async (user) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.username}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger'
      }
    )

    await deleteUser(user.username)
    ElMessage.success('用户删除成功')
    loadUsers()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete user:', error)
    }
  }
}

// 页面加载时获取用户列表
onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.user-management {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.header-icon {
  font-size: 20px;
  color: #409EFF;
}

.username-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-icon {
  color: #909399;
  font-size: 16px;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.role-option {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.role-option small {
  color: #909399;
  font-size: 11px;
  line-height: 1.2;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }

  .header-right {
    display: flex;
    justify-content: center;
  }

  .action-buttons {
    flex-direction: column;
    gap: 4px;
  }

  .action-buttons .el-button {
    width: 80px;
    font-size: 12px;
  }
}
</style>
