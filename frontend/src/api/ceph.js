import request from './index'

/**
 * 获取Ceph集群完整信息
 * @returns {Promise} 集群信息
 */
export const getCephClusterInfo = () => {
  return request({
    url: '/ceph/cluster/info',
    method: 'get'
  })
}

/**
 * 获取Ceph集群状态
 * @returns {Promise} 集群状态
 */
export const getCephClusterStatus = () => {
  return request({
    url: '/ceph/cluster/status',
    method: 'get'
  })
}

/**
 * 获取Ceph Pool信息
 * @returns {Promise} Pool列表
 */
export const getCephPools = () => {
  return request({
    url: '/ceph/pools',
    method: 'get'
  })
}

/**
 * 获取Ceph连接状态
 * @returns {Promise} 连接状态
 */
export const getCephConnectionStatus = () => {
  return request({
    url: '/ceph/connection/status',
    method: 'get'
  })
}

/**
 * 格式化字节数为人类可读格式
 * @param {number} bytes 字节数
 * @returns {string} 格式化后的字符串
 */
export const formatBytes = (bytes) => {
  if (bytes === 0) return '0 B'
  if (!bytes || bytes < 0) return '0 B'

  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * 格式化百分比
 * @param {number} value 数值
 * @param {number} decimals 小数位数
 * @returns {string} 格式化后的百分比
 */
export const formatPercent = (value, decimals = 2) => {
  if (!value || value < 0) return '0.00%'
  return parseFloat(value.toFixed(decimals)) + '%'
}
