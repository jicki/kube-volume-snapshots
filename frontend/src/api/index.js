import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

// 创建 axios 实例
const api = axios.create({
  baseURL: process.env.NODE_ENV === 'production' ? '/api' : 'http://localhost:8081/api',
  timeout: 30000 // 30秒超时，适应大量PVC的情况
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    // 添加认证头
    const authStore = useAuthStore()
    console.log('API - request interceptor: token exists:', !!authStore.token)
    console.log('API - request interceptor: making request to:', config.url)
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    } else {
      console.warn('API - request interceptor: no token available')
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    const { data } = response
    if (data.code !== 200) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    return data.data
  },
  error => {
    const message = error.response?.data?.message || error.message || '网络错误'

    // 处理401认证错误
    if (error.response?.status === 401) {
      console.error('API - response interceptor: 401 Unauthorized received')
      console.error('API - response interceptor: error details:', error.response?.data)
      const authStore = useAuthStore()
      authStore.clearAuth()
      ElMessage.error('认证已过期，请重新登录')
      // 跳转到登录页
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
      return Promise.reject(error)
    }

    ElMessage.error(message)
    return Promise.reject(error)
  }
)

// VolumeSnapshotClass 相关 API
export const getVolumeSnapshotClasses = () => {
  return api.get('/volumesnapshotclasses')
}

// VolumeSnapshot 相关 API
export const getVolumeSnapshots = (namespace = '') => {
  return api.get('/volumesnapshots', { params: { namespace } })
}

export const createVolumeSnapshot = (data) => {
  return api.post('/volumesnapshots', data)
}

export const deleteVolumeSnapshot = (namespace, name) => {
  return api.delete(`/volumesnapshots/${namespace}/${name}`)
}

export const forceDeleteVolumeSnapshot = (namespace, name) => {
  return api.post(`/volumesnapshots/${namespace}/${name}/force-delete`)
}

// VolumeSnapshotContent 相关 API
export const getVolumeSnapshotContent = (name) => {
  return api.get(`/volumesnapshotcontents/${name}`)
}

// PVC 相关 API
export const getPVCs = (namespace = 'default') => {
  console.log('API - getPVCs: requesting PVCs for namespace:', namespace)
  return api.get('/pvcs', { params: { namespace } })
    .then(response => {
      console.log('API - getPVCs: response received:', response)
      return response
    })
    .catch(error => {
      console.error('API - getPVCs: request failed:', error)
      throw error
    })
}

// Namespace 相关 API
export const getNamespaces = () => {
  return api.get('/namespaces')
}

// StorageClass 相关 API
export const getStorageClasses = () => {
  return api.get('/storageclasses')
}

// 定时任务相关 API
export const getScheduledSnapshots = () => {
  return api.get('/scheduled-snapshots')
}

export const createScheduledSnapshot = (data) => {
  return api.post('/scheduled-snapshots', data)
}

export const updateScheduledSnapshot = (id, data) => {
  return api.put(`/scheduled-snapshots/${id}`, data)
}

export const deleteScheduledSnapshot = (id) => {
  return api.delete(`/scheduled-snapshots/${id}`)
}

export const toggleScheduledSnapshot = (id) => {
  return api.post(`/scheduled-snapshots/${id}/toggle`)
}

// 用户认证相关 API
export const login = (credentials) => {
  return api.post('/auth/login', credentials)
}

export const register = (userData) => {
  return api.post('/auth/register', userData)
}

// 用户管理相关 API
export const getUserProfile = () => {
  return api.get('/user/profile')
}

export const changePassword = (passwordData) => {
  return api.post('/user/change-password', passwordData)
}

export const getAllUsers = () => {
  return api.get('/user/all')
}

export const deleteUser = (username) => {
  return api.delete(`/user/${username}`)
}

// 集群管理相关 API
export const getClusters = () => {
  return api.get('/clusters')
}

export const getCurrentCluster = () => {
  return api.get('/clusters/current')
}

export const switchCluster = (clusterName) => {
  return api.post('/clusters/switch', { cluster_name: clusterName })
}

export default api
