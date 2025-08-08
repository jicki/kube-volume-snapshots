<template>
  <div id="app">
    <!-- 登录页面布局 -->
    <div v-if="isLoginPage" class="login-layout">
      <router-view />
    </div>

    <!-- 主应用布局 -->
    <el-container v-else class="app-container">
      <el-aside class="app-sidebar" width="260px">
        <div class="sidebar-header">
          <h1 class="app-title">
            <el-icon class="title-icon"><DataBoard /></el-icon>
            <div class="title-text">
              <div class="title-main">Volume Snapshots</div>
              <div class="title-sub">K8s 管理器</div>
            </div>
          </h1>
        </div>
        <div class="sidebar-nav">
          <el-menu
            :default-active="$route.path"
            router
            background-color="#2c3e50"
            text-color="#ecf0f1"
            active-text-color="#3498db"
            class="sidebar-menu"
          >
            <el-menu-item index="/dashboard">
              <el-icon><Monitor /></el-icon>
              <span>仪表板</span>
            </el-menu-item>
            <el-menu-item index="/pvcs">
              <el-icon><Files /></el-icon>
              <span>PVC 管理</span>
            </el-menu-item>
            <el-menu-item index="/snapshot-classes">
              <el-icon><Folder /></el-icon>
              <span>快照类</span>
            </el-menu-item>
            <el-menu-item index="/snapshots">
              <el-icon><CameraFilled /></el-icon>
              <span>快照管理</span>
            </el-menu-item>
            <el-menu-item index="/scheduled">
              <el-icon><Clock /></el-icon>
              <span>定时任务</span>
            </el-menu-item>
            <el-menu-item index="/ceph">
              <el-icon><Monitor /></el-icon>
              <span>Ceph集群</span>
            </el-menu-item>
            <el-menu-item index="/clusters">
              <el-icon><Connection /></el-icon>
              <span>集群管理</span>
            </el-menu-item>
            <el-menu-item v-if="isAdmin" index="/users">
              <el-icon><User /></el-icon>
              <span>用户管理</span>
            </el-menu-item>
          </el-menu>
        </div>
      </el-aside>

      <el-container class="main-container">
        <el-header class="top-header" height="60px">
          <div class="header-breadcrumb">
            <span class="current-page">{{ getCurrentPageTitle() }}</span>
          </div>
          <div class="header-right">
            <div class="user-info">
              <el-dropdown @command="handleUserCommand">
                <div class="user-avatar">
                  <el-icon class="avatar-icon"><User /></el-icon>
                  <span class="username">{{ currentUser?.username }}</span>
                  <el-tag
                    :type="currentUser?.role === 'admin' ? 'danger' : 'success'"
                    size="small"
                    class="role-tag"
                  >
                    {{ currentUser?.role === 'admin' ? '管理员' : '只读' }}
                  </el-tag>
                  <el-icon class="dropdown-icon"><ArrowDown /></el-icon>
                </div>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="profile">
                      <el-icon><User /></el-icon>
                      个人资料
                    </el-dropdown-item>
                    <el-dropdown-item command="changePassword">
                      <el-icon><Lock /></el-icon>
                      修改密码
                    </el-dropdown-item>
                    <el-dropdown-item divided command="logout">
                      <el-icon><SwitchButton /></el-icon>
                      退出登录
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </el-header>

        <el-main class="app-main">
          <router-view />
        </el-main>
      </el-container>
    </el-container>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="showChangePasswordDialog"
      title="修改密码"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form label-width="100px">
        <el-form-item label="原密码">
          <el-input
            v-model="changePasswordForm.oldPassword"
            type="password"
            placeholder="请输入原密码"
            show-password
            clearable
          />
        </el-form-item>

        <el-form-item label="新密码">
          <el-input
            v-model="changePasswordForm.newPassword"
            type="password"
            placeholder="请输入新密码（至少6个字符）"
            show-password
            clearable
          />
        </el-form-item>

        <el-form-item label="确认密码">
          <el-input
            v-model="changePasswordForm.confirmPassword"
            type="password"
            placeholder="请确认新密码"
            show-password
            clearable
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="handleCancelChangePassword">取消</el-button>
        <el-button
          type="primary"
          :loading="changePasswordLoading"
          @click="handleChangePassword"
        >
          {{ changePasswordLoading ? '修改中...' : '确认修改' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 个人资料对话框 -->
    <el-dialog
      v-model="showProfileDialog"
      title="个人资料"
      width="400px"
    >
      <div class="profile-info">
        <div class="profile-item">
          <label>用户名：</label>
          <span>{{ currentUser?.username }}</span>
        </div>
        <div class="profile-item">
          <label>角色：</label>
          <el-tag :type="currentUser?.role === 'admin' ? 'danger' : 'success'" size="small">
            {{ currentUser?.role === 'admin' ? '管理员' : '只读用户' }}
          </el-tag>
        </div>
        <div class="profile-item">
          <label>创建时间：</label>
          <span>{{ formatDateTime(currentUser?.createdAt) }}</span>
        </div>
        <div class="profile-item">
          <label>更新时间：</label>
          <span>{{ formatDateTime(currentUser?.updatedAt) }}</span>
        </div>
      </div>

      <template #footer>
        <el-button type="primary" @click="showProfileDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElDialog, ElForm, ElFormItem, ElInput, ElButton, ElMessage } from 'element-plus'
import {
  DataBoard,
  Monitor,
  Folder,
  CameraFilled,
  Clock,
  Files,
  User,
  ArrowDown,
  Lock,
  SwitchButton,
  Connection
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 响应式数据
const showChangePasswordDialog = ref(false)
const showProfileDialog = ref(false)
const changePasswordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})
const changePasswordLoading = ref(false)

// 计算属性
const currentUser = computed(() => authStore.currentUser)
const isAdmin = computed(() => authStore.isAdmin)
const isLoginPage = computed(() => route.path === '/login')

// 页面标题映射
const getCurrentPageTitle = () => {
  const titleMap = {
    '/dashboard': '仪表板',
    '/pvcs': 'PVC 管理',
    '/snapshot-classes': '快照类',
    '/snapshots': '快照管理',
    '/scheduled': '定时任务',
    '/ceph': 'Ceph集群',
    '/clusters': '集群管理',
    '/users': '用户管理',
    '/login': '登录'
  }
  return titleMap[route.path] || '仪表板'
}

// 处理用户命令
const handleUserCommand = (command) => {
  switch (command) {
    case 'profile':
      showProfileDialog.value = true
      break
    case 'changePassword':
      showChangePasswordDialog.value = true
      break
    case 'logout':
      authStore.logout()
      break
  }
}

// 处理修改密码
const handleChangePassword = async () => {
  // 简单验证
  if (!changePasswordForm.value.oldPassword || !changePasswordForm.value.newPassword) {
    ElMessage.error('请填写完整信息')
    return
  }

  if (changePasswordForm.value.newPassword !== changePasswordForm.value.confirmPassword) {
    ElMessage.error('两次输入的密码不一致')
    return
  }

  if (changePasswordForm.value.newPassword.length < 6) {
    ElMessage.error('新密码长度至少6个字符')
    return
  }

  changePasswordLoading.value = true
  try {
    const result = await authStore.changePassword({
      oldPassword: changePasswordForm.value.oldPassword,
      newPassword: changePasswordForm.value.newPassword
    })

    if (result.success) {
      showChangePasswordDialog.value = false
      // 重置表单
      changePasswordForm.value = {
        oldPassword: '',
        newPassword: '',
        confirmPassword: ''
      }
    }
  } catch (error) {
    console.error('Change password failed:', error)
  } finally {
    changePasswordLoading.value = false
  }
}

// 取消修改密码
const handleCancelChangePassword = () => {
  showChangePasswordDialog.value = false
  changePasswordForm.value = {
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
}

// 格式化时间
const formatDateTime = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

// 页面加载时恢复认证状态
onMounted(async () => {
  console.log('🚀 App.vue onMounted - 开始初始化认证')

  // 恢复认证状态
  const restored = authStore.restoreAuth()
  console.log('📱 认证状态恢复结果:', restored)

  if (restored) {
    console.log('🔍 验证token有效性...')
    const isValid = await authStore.checkToken()
    console.log('✅ Token验证结果:', isValid)

    if (!isValid) {
      console.log('❌ Token无效，跳转到登录页')
      router.push('/login')
      return
    }
  } else {
    console.log('❌ 没有认证信息，跳转到登录页')
    router.push('/login')
    return
  }

  console.log('✅ 认证检查完成，用户已登录')
})
</script>

<style scoped>
/* 登录页面布局 */
.login-layout {
  height: 100vh;
  width: 100vw;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.app-container {
  height: 100vh;
}

.app-sidebar {
  background: #2c3e50;
  box-shadow: 2px 0 8px rgba(0,0,0,0.1);
  overflow: hidden;
}

.sidebar-header {
  padding: 20px 16px;
  border-bottom: 1px solid #34495e;
  background: #34495e;
}

.app-title {
  color: #ecf0f1;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-icon {
  font-size: 24px;
  color: #3498db;
  flex-shrink: 0;
}

.title-text {
  flex: 1;
  min-width: 0;
}

.title-main {
  font-size: 16px;
  font-weight: 600;
  line-height: 1.2;
  color: #ecf0f1;
  margin-bottom: 2px;
}

.title-sub {
  font-size: 12px;
  font-weight: 400;
  line-height: 1.1;
  color: #bdc3c7;
  opacity: 0.8;
}

.sidebar-nav {
  padding: 10px 0;
}

.sidebar-menu {
  border: none;
}

.sidebar-menu .el-menu-item {
  height: 50px;
  line-height: 50px;
  padding: 0 20px;
  display: flex;
  align-items: center;
  transition: all 0.3s ease;
}

.sidebar-menu .el-menu-item:hover {
  background-color: #34495e;
  color: #3498db;
}

.sidebar-menu .el-menu-item.is-active {
  background-color: #3498db;
  color: #ffffff;
  border-right: 3px solid #2980b9;
}

.sidebar-menu .el-menu-item .el-icon {
  margin-right: 12px;
  font-size: 18px;
}

.main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.top-header {
  background: #ffffff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0,0,0,0.1);
}

.header-breadcrumb {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info .user-avatar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #f8f9fa;
}

.user-info .user-avatar:hover {
  background: #e9ecef;
}

.user-info .avatar-icon {
  font-size: 18px;
  color: #409EFF;
}

.user-info .username {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.user-info .role-tag {
  margin-left: 4px;
}

.user-info .dropdown-icon {
  font-size: 12px;
  color: #909399;
  transition: transform 0.3s ease;
}

.user-info .dropdown-icon:hover {
  transform: rotate(180deg);
}

.current-page {
  font-size: 18px;
  font-weight: 500;
  color: #303133;
}

.app-main {
  background-color: #f5f7fa;
  padding: 20px;
  flex: 1;
  overflow-y: auto;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .app-sidebar {
    width: 200px !important;
  }

  .app-title {
    font-size: 16px;
  }

  .sidebar-menu .el-menu-item {
    height: 45px;
    line-height: 45px;
    padding: 0 16px;
  }

  .current-page {
    font-size: 16px;
  }

  .app-main {
    padding: 16px;
  }
}

@media (max-width: 480px) {
  .app-sidebar {
    width: 180px !important;
  }

  .sidebar-header {
    padding: 16px 12px;
  }

  .title-main {
    font-size: 14px;
  }

  .title-sub {
    font-size: 10px;
  }

  .sidebar-menu .el-menu-item {
    padding: 0 12px;
  }

  .sidebar-menu .el-menu-item span {
    font-size: 14px;
  }
}

@media (max-width: 380px) {
  .app-sidebar {
    width: 160px !important;
  }

  .title-text {
    display: none;
  }

  .title-icon {
    font-size: 20px;
    margin: 0 auto;
  }
}
</style>

<style>
/* 更精确的CSS重置，避免影响组件库样式 */
html, body {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

*, *:before, *:after {
  box-sizing: inherit;
}

body {
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
}

#app {
  height: 100vh;
}

/* 确保表格组件正常显示 */
.el-table {
  --el-table-border-color: #ebeef5;
  --el-table-border: 1px solid var(--el-table-border-color);
  --el-table-text-color: #606266;
  --el-table-header-text-color: #909399;
  --el-table-header-bg-color: #fafafa;
  --el-table-bg-color: #ffffff;
  --el-table-current-row-bg-color: #ecf5ff;
  --el-table-row-hover-bg-color: #f5f7fa;
  width: 100% !important;
}

/* 确保表格容器能够自适应 */
.el-card .el-card__body {
  padding: 20px;
  overflow-x: auto;
}

/* 表格自适应容器宽度 */
.el-table, .el-table__expanded-cell {
  width: 100%;
}

.el-table .el-table__body-wrapper {
  overflow-x: auto;
}

/* 个人资料对话框样式 */
.profile-info {
  padding: 16px 0;
}

.profile-item {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
  padding: 8px 0;
  border-bottom: 1px solid #f0f2f5;
}

.profile-item:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.profile-item label {
  font-weight: 500;
  color: #606266;
  width: 80px;
  flex-shrink: 0;
}

.profile-item span {
  color: #303133;
  flex: 1;
}
</style>
