<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <el-icon class="login-icon"><DataBoard /></el-icon>
        <h1>K8s Volume Snapshots</h1>
        <p class="login-subtitle">卷快照管理系统</p>
      </div>

      <div class="login-form-container">
        <el-form
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          label-position="top"
          @keyup.enter="handleLogin"
        >
          <el-form-item label="用户名" prop="username">
            <el-input
              v-model="loginForm.username"
              placeholder="请输入用户名"
              size="large"
              clearable
            >
              <template #prefix><el-icon><User /></el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input
              type="password"
              v-model="loginForm.password"
              placeholder="请输入密码"
              show-password
              size="large"
              clearable
            >
              <template #prefix><el-icon><Lock /></el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              :loading="loginLoading"
              @click="handleLogin"
              class="login-button"
              size="large"
            >
              {{ loginLoading ? '登录中...' : '登录' }}
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 默认管理员账号提示 -->
      <div class="default-account-info">
        <el-divider>默认管理员账号</el-divider>
        <div class="account-item">
          <el-icon><InfoFilled /></el-icon>
          <div class="account-details">
            <p><strong>用户名:</strong> admin</p>
            <p><strong>密码:</strong> admin123</p>
            <p><strong>角色:</strong> 管理员</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  DataBoard,
  User,
  Lock,
  InfoFilled
} from '@element-plus/icons-vue'

const authStore = useAuthStore()
const router = useRouter()

// 响应式数据
const loginLoading = ref(false)

// 表单引用
const loginFormRef = ref()

// 登录表单
const loginForm = reactive({
  username: '',
  password: ''
})

// 表单验证规则
const loginRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

// 处理登录
const handleLogin = async () => {
  if (!loginFormRef.value) return

  try {
    await loginFormRef.value.validate()
    loginLoading.value = true

    const result = await authStore.login({
      username: loginForm.username,
      password: loginForm.password
    })

    if (result.success) {
      // 登录成功，跳转到仪表板
      router.push('/dashboard')
    }
  } catch (error) {
    console.error('登录失败:', error)
  } finally {
    loginLoading.value = false
  }
}

</script>

<style scoped>
.login-container {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
}

.login-card {
  width: 420px;
  padding: 40px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-icon {
  font-size: 48px;
  color: #409eff;
  margin-bottom: 16px;
}

.login-header h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.login-subtitle {
  margin: 0;
  font-size: 14px;
  color: #909399;
}

.login-form-container {
  margin-bottom: 30px;
}

.login-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 8px;
}

.default-account-info {
  margin-top: 30px;
  padding-top: 20px;
}

.default-account-info :deep(.el-divider__text) {
  background: rgba(255, 255, 255, 0.95);
  color: #909399;
  font-size: 12px;
}

.account-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
  border-left: 4px solid #409eff;
}

.account-item .el-icon {
  color: #409eff;
  font-size: 18px;
  margin-top: 2px;
  flex-shrink: 0;
}

.account-details {
  flex: 1;
}

.account-details p {
  margin: 0 0 4px 0;
  font-size: 13px;
  color: #606266;
  line-height: 1.4;
}

.account-details p:last-child {
  margin-bottom: 0;
}

.account-details strong {
  color: #303133;
}

/* 表单样式优化 */
:deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
  margin-bottom: 8px;
}

:deep(.el-input__wrapper) {
  border-radius: 8px;
  padding: 12px 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid #e4e7ed;
  transition: all 0.3s ease;
}

:deep(.el-input__wrapper:hover) {
  border-color: #409eff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.1);
}

:deep(.el-input__wrapper.is-focus) {
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

:deep(.el-form-item) {
  margin-bottom: 24px;
}
</style>
