import { defineStore } from 'pinia'
import { ElMessage, ElMessageBox } from 'element-plus'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    token: localStorage.getItem('token') || null,
    isAuthenticated: false
  }),

  getters: {
    isAdmin: (state) => state.user?.role === 'admin',
    isReadonly: (state) => state.user?.role === 'readonly',
    currentUser: (state) => state.user,
    hasToken: (state) => !!state.token
  },

  actions: {
    // 设置认证信息
    setAuth (token, user) {
      this.token = token
      this.user = user
      this.isAuthenticated = true
      localStorage.setItem('token', token)
      localStorage.setItem('user', JSON.stringify(user))
    },

    // 清除认证信息
    clearAuth () {
      this.token = null
      this.user = null
      this.isAuthenticated = false
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    },

    // 从本地存储恢复认证状态
    restoreAuth () {
      const token = localStorage.getItem('token')
      const userStr = localStorage.getItem('user')

      if (token && userStr) {
        try {
          const user = JSON.parse(userStr)
          this.token = token
          this.user = user
          this.isAuthenticated = true
          return true
        } catch (error) {
          console.error('Failed to parse user data:', error)
          this.clearAuth()
          return false
        }
      }
      return false
    },

    // 检查token是否有效
    async checkToken () {
      if (!this.token) {
        return false
      }

      try {
        // 通过调用需要认证的接口来验证token
        const baseURL = process.env.NODE_ENV === 'production' ? '/api' : 'http://localhost:8081/api'
        const response = await fetch(`${baseURL}/user/profile`, {
          headers: {
            Authorization: `Bearer ${this.token}`
          }
        })

        if (response.ok) {
          const data = await response.json()
          if (data.code === 200) {
            this.user = data.data
            this.isAuthenticated = true
            localStorage.setItem('user', JSON.stringify(data.data))
            return true
          }
        }

        // Token 无效，清除认证信息
        this.clearAuth()
        return false
      } catch (error) {
        console.error('Token verification failed:', error)
        this.clearAuth()
        return false
      }
    },

    // 登录
    async login (credentials) {
      try {
        const baseURL = process.env.NODE_ENV === 'production' ? '/api' : 'http://localhost:8081/api'
        const response = await fetch(`${baseURL}/auth/login`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(credentials)
        })

        const data = await response.json()

        if (data.code === 200) {
          this.setAuth(data.data.token, data.data.user)
          ElMessage.success('登录成功')
          return { success: true, data: data.data }
        } else {
          ElMessage.error(data.message || '登录失败')
          return { success: false, message: data.message }
        }
      } catch (error) {
        const message = '网络错误，请稍后重试'
        ElMessage.error(message)
        return { success: false, message }
      }
    },

    // 注册
    async register (userData) {
      try {
        const baseURL = process.env.NODE_ENV === 'production' ? '/api' : 'http://localhost:8081/api'
        const response = await fetch(`${baseURL}/auth/register`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(userData)
        })

        const data = await response.json()

        if (data.code === 200) {
          ElMessage.success('注册成功')
          return { success: true, data: data.data }
        } else {
          ElMessage.error(data.message || '注册失败')
          return { success: false, message: data.message }
        }
      } catch (error) {
        const message = '网络错误，请稍后重试'
        ElMessage.error(message)
        return { success: false, message }
      }
    },

    // 退出登录
    async logout (showConfirm = true) {
      const doLogout = () => {
        this.clearAuth()
        ElMessage.success('已退出登录')
        // 跳转到登录页
        window.location.href = '/login'
      }

      if (showConfirm) {
        try {
          await ElMessageBox.confirm('确定要退出登录吗？', '确认', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          })
          doLogout()
        } catch {
          // 用户取消
        }
      } else {
        doLogout()
      }
    },

    // 修改密码
    async changePassword (passwordData) {
      try {
        const baseURL = process.env.NODE_ENV === 'production' ? '/api' : 'http://localhost:8081/api'
        const response = await fetch(`${baseURL}/user/change-password`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${this.token}`
          },
          body: JSON.stringify(passwordData)
        })

        const data = await response.json()

        if (data.code === 200) {
          ElMessage.success('密码修改成功')
          return { success: true }
        } else {
          ElMessage.error(data.message || '密码修改失败')
          return { success: false, message: data.message }
        }
      } catch (error) {
        const message = '网络错误，请稍后重试'
        ElMessage.error(message)
        return { success: false, message }
      }
    }
  }
})
