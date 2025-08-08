//go:build ceph

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"k8s-volume-snapshots/models"

	"github.com/ceph/go-ceph/rados"
)

// CephService Ceph集群服务
type CephService struct {
	conn       *rados.Conn
	configFile string
	keyring    string

	// 缓存相关
	statusCache  *models.CephClusterInfo
	cacheMutex   sync.RWMutex
	cacheTime    time.Time
	cacheTimeout time.Duration
}

// NewCephService 创建新的Ceph服务实例
func NewCephService() (*CephService, error) {
	// 从环境变量获取配置
	configFile := os.Getenv("CEPH_CONF")
	if configFile == "" {
		configFile = "/etc/ceph/ceph.conf"
	}

	keyring := os.Getenv("CEPH_KEYRING")
	if keyring == "" {
		keyring = "/etc/ceph/ceph.client.admin.keyring"
	}

	service := &CephService{
		configFile:   configFile,
		keyring:      keyring,
		cacheTimeout: 30 * time.Second, // 缓存30秒
	}

	// 尝试连接Ceph集群
	if err := service.connect(); err != nil {
		// 如果连接失败，返回服务实例但标记为离线状态
		// 这样前端可以显示连接错误而不是完全无法访问
		return service, fmt.Errorf("failed to connect to Ceph cluster: %w", err)
	}

	return service, nil
}

// connect 连接到Ceph集群
func (c *CephService) connect() error {
	// 创建连接时指定用户名
	conn, err := rados.NewConnWithUser("test")
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}

	// 设置配置文件
	if err := conn.ReadConfigFile(c.configFile); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 设置keyring文件
	if err := conn.SetConfigOption("keyring", c.keyring); err != nil {
		return fmt.Errorf("failed to set keyring: %w", err)
	}

	// 连接集群
	if err := conn.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	return nil
}

// Close 关闭连接
func (c *CephService) Close() {
	if c.conn != nil {
		c.conn.Shutdown()
	}
}

// GetClusterInfo 获取完整的集群信息
func (c *CephService) GetClusterInfo(ctx context.Context) (*models.CephClusterInfo, error) {
	// 检查缓存
	c.cacheMutex.RLock()
	if c.statusCache != nil && time.Since(c.cacheTime) < c.cacheTimeout {
		defer c.cacheMutex.RUnlock()
		return c.statusCache, nil
	}
	c.cacheMutex.RUnlock()

	// 缓存过期，重新获取数据
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	// 双重检查
	if c.statusCache != nil && time.Since(c.cacheTime) < c.cacheTimeout {
		return c.statusCache, nil
	}

	clusterInfo := &models.CephClusterInfo{
		UpdatedAt: time.Now(),
	}

	// 获取集群状态
	status, err := c.getClusterStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster status: %w", err)
	}
	clusterInfo.Status = *status

	// 获取Pool信息
	pools, err := c.getPoolsInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pools info: %w", err)
	}
	clusterInfo.Pools = pools

	// 更新缓存
	c.statusCache = clusterInfo
	c.cacheTime = time.Now()

	return clusterInfo, nil
}

// getClusterStatus 获取集群状态
func (c *CephService) getClusterStatus(ctx context.Context) (*models.CephClusterStatus, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to Ceph cluster")
	}

	status := &models.CephClusterStatus{
		UpdateTime: time.Now(),
		Monitors:   make([]models.CephMonitorInfo, 0),
	}

	// 获取集群状态 - 修正 MonCommand 格式
	cmd := map[string]interface{}{
		"prefix": "status",
		"format": "json",
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	buf, info, err := c.conn.MonCommand(cmdBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster status: %w", err)
	}
	_ = info // 忽略info

	// 解析状态信息
	var statusData map[string]interface{}
	if err := json.Unmarshal(buf, &statusData); err != nil {
		return nil, fmt.Errorf("failed to parse status data: %w", err)
	}

	// 解析健康状态
	if health, ok := statusData["health"].(map[string]interface{}); ok {
		if healthStatus, ok := health["status"].(string); ok {
			status.Health = healthStatus
		}
	}

	// 获取Monitor信息 - 使用专门的方法
	monitors, err := c.getMonitorInfo(ctx)
	if err == nil {
		status.Monitors = monitors
	}

	// 解析OSD信息 - 修正数据路径
	if osdmap, ok := statusData["osdmap"].(map[string]interface{}); ok {
		if numOsds, ok := osdmap["num_osds"].(float64); ok {
			status.OSDs.Total = int(numOsds)
		}
		if numUpOsds, ok := osdmap["num_up_osds"].(float64); ok {
			status.OSDs.Up = int(numUpOsds)
		}
		if numInOsds, ok := osdmap["num_in_osds"].(float64); ok {
			status.OSDs.In = int(numInOsds)
		}
	}

	// 解析PG信息
	if pgmap, ok := statusData["pgmap"].(map[string]interface{}); ok {
		if numPgs, ok := pgmap["num_pgs"].(float64); ok {
			status.PGs.Total = int(numPgs)
		}
		// 简化处理，假设大部分PG都是active+clean
		status.PGs.Active = status.PGs.Total
		status.PGs.Clean = status.PGs.Total
		status.PGs.Degraded = 0
	}

	// 获取容量信息
	capacity, err := c.getCapacityInfo(ctx)
	if err == nil {
		status.Capacity = *capacity
	}

	// 获取版本信息
	status.Version = c.getClusterVersion(ctx)

	return status, nil
}

// getCapacityInfo 获取容量信息
func (c *CephService) getCapacityInfo(ctx context.Context) (*models.CephCapacitySummary, error) {
	cmd := map[string]interface{}{
		"prefix": "df",
		"format": "json",
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	buf, _, err := c.conn.MonCommand(cmdBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get df info: %w", err)
	}

	var dfData map[string]interface{}
	if err := json.Unmarshal(buf, &dfData); err != nil {
		return nil, fmt.Errorf("failed to parse df data: %w", err)
	}

	capacity := &models.CephCapacitySummary{}

	if stats, ok := dfData["stats"].(map[string]interface{}); ok {
		if totalBytes, ok := stats["total_bytes"].(float64); ok {
			capacity.TotalBytes = int64(totalBytes)
		}
		if usedBytes, ok := stats["total_used_bytes"].(float64); ok {
			capacity.UsedBytes = int64(usedBytes)
		}
		if availBytes, ok := stats["total_avail_bytes"].(float64); ok {
			capacity.AvailBytes = int64(availBytes)
		}
	}

	// 计算使用率
	if capacity.TotalBytes > 0 {
		capacity.UsagePercent = float64(capacity.UsedBytes) / float64(capacity.TotalBytes) * 100
	}

	return capacity, nil
}

// getPoolsInfo 获取Pool信息
func (c *CephService) getPoolsInfo(ctx context.Context) ([]models.CephPoolInfo, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to Ceph cluster")
	}

	// 先获取df数据，用于后续的统计信息匹配
	dfData, err := c.getDfData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get df data: %w", err)
	}

	// 获取Pool列表
	cmd := map[string]interface{}{
		"prefix": "osd pool ls",
		"detail": "detail",
		"format": "json",
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	buf, _, err := c.conn.MonCommand(cmdBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get pools list: %w", err)
	}

	var poolsData []interface{}
	if err := json.Unmarshal(buf, &poolsData); err != nil {
		return nil, fmt.Errorf("failed to parse pools data: %w", err)
	}

	pools := make([]models.CephPoolInfo, 0, len(poolsData))

	for _, poolData := range poolsData {
		if poolInfo, ok := poolData.(map[string]interface{}); ok {
			pool := models.CephPoolInfo{}

			if poolID, ok := poolInfo["pool"].(float64); ok {
				pool.ID = int(poolID)
			}
			if poolName, ok := poolInfo["pool_name"].(string); ok {
				pool.Name = poolName
			}
			if size, ok := poolInfo["size"].(float64); ok {
				pool.Size = int(size)
			}
			if minSize, ok := poolInfo["min_size"].(float64); ok {
				pool.MinSize = int(minSize)
			}
			if pgNum, ok := poolInfo["pg_num"].(float64); ok {
				pool.PGNum = int(pgNum)
			}
			if pgpNum, ok := poolInfo["pgp_num"].(float64); ok {
				pool.PGPNum = int(pgpNum)
			}

			// 判断Pool类型
			if pool.Size > 1 {
				pool.Type = "replicated"
			} else {
				pool.Type = "erasure" // 简化处理
			}

			// 从df数据中查找对应池子的统计信息
			if poolStatsData := c.findPoolStatsInDfData(dfData, pool.Name); poolStatsData != nil {
				pool.Objects = poolStatsData.Objects
				pool.UsedBytes = poolStatsData.UsedBytes
				pool.MaxAvailBytes = poolStatsData.MaxAvailBytes
				// 计算使用率：如果有可用空间数据，计算使用率百分比
				if poolStatsData.MaxAvailBytes > 0 {
					totalBytes := poolStatsData.UsedBytes + poolStatsData.MaxAvailBytes
					if totalBytes > 0 {
						pool.UsagePercent = (float64(poolStatsData.UsedBytes) / float64(totalBytes)) * 100
					} else {
						pool.UsagePercent = 0
					}
				} else {
					// 如果没有MaxAvail数据，使用Ceph提供的PercentUsed（转换为百分比）
					pool.UsagePercent = poolStatsData.PercentUsed * 100
				}
			} else {
				// 调试：如果找不到数据，至少记录一下
				fmt.Printf("DEBUG: Pool %s not found in df data\n", pool.Name)
			}

			pools = append(pools, pool)
		}
	}

	return pools, nil
}

// PoolStats Pool统计信息结构
type PoolStats struct {
	Objects       int64
	UsedBytes     int64
	MaxAvailBytes int64
	PercentUsed   float64 // Ceph提供的使用率（0-1之间的小数）
}

// getPoolStats 获取Pool统计信息
func (c *CephService) getPoolStats(ctx context.Context, poolName string) (*PoolStats, error) {
	cmd := map[string]interface{}{
		"prefix": "df",
		"format": "json",
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	buf, _, err := c.conn.MonCommand(cmdBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool stats: %w", err)
	}

	var dfData map[string]interface{}
	if err := json.Unmarshal(buf, &dfData); err != nil {
		return nil, fmt.Errorf("failed to parse df data: %w", err)
	}

	if pools, ok := dfData["pools"].([]interface{}); ok {
		for _, pool := range pools {
			if poolData, ok := pool.(map[string]interface{}); ok {
				if name, ok := poolData["name"].(string); ok && name == poolName {
					stats := &PoolStats{}
					if statsData, ok := poolData["stats"].(map[string]interface{}); ok {
						if objects, ok := statsData["objects"].(float64); ok {
							stats.Objects = int64(objects)
						}
						if bytesUsed, ok := statsData["bytes_used"].(float64); ok {
							stats.UsedBytes = int64(bytesUsed)
						}
						if maxAvail, ok := statsData["max_avail"].(float64); ok {
							stats.MaxAvailBytes = int64(maxAvail)
						}
						if percentUsed, ok := statsData["percent_used"].(float64); ok {
							stats.PercentUsed = percentUsed
						}
					}
					return stats, nil
				}
			}
		}
	}

	return &PoolStats{}, nil
}

// getDfData 获取df数据（一次性获取所有池子统计信息）
func (c *CephService) getDfData(ctx context.Context) (map[string]interface{}, error) {
	cmd := map[string]interface{}{
		"prefix": "df",
		"format": "json",
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	buf, _, err := c.conn.MonCommand(cmdBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get df data: %w", err)
	}

	var dfData map[string]interface{}
	if err := json.Unmarshal(buf, &dfData); err != nil {
		return nil, fmt.Errorf("failed to parse df data: %w", err)
	}

	return dfData, nil
}

// findPoolStatsInDfData 从df数据中查找指定池子的统计信息
func (c *CephService) findPoolStatsInDfData(dfData map[string]interface{}, poolName string) *PoolStats {
	if pools, ok := dfData["pools"].([]interface{}); ok {
		for _, pool := range pools {
			if poolData, ok := pool.(map[string]interface{}); ok {
				if name, ok := poolData["name"].(string); ok && name == poolName {
					stats := &PoolStats{}
					if statsData, ok := poolData["stats"].(map[string]interface{}); ok {
						if objects, ok := statsData["objects"].(float64); ok {
							stats.Objects = int64(objects)
						}
						if bytesUsed, ok := statsData["bytes_used"].(float64); ok {
							stats.UsedBytes = int64(bytesUsed)
						}
						if maxAvail, ok := statsData["max_avail"].(float64); ok {
							stats.MaxAvailBytes = int64(maxAvail)
						}
						if percentUsed, ok := statsData["percent_used"].(float64); ok {
							stats.PercentUsed = percentUsed
						}
					}
					return stats
				}
			}
		}
	}
	return nil
}

// getMonitorInfo 获取Monitor节点信息
func (c *CephService) getMonitorInfo(ctx context.Context) ([]models.CephMonitorInfo, error) {
	cmd := map[string]interface{}{
		"prefix": "mon dump",
		"format": "json",
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	buf, _, err := c.conn.MonCommand(cmdBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitor info: %w", err)
	}

	var monData map[string]interface{}
	if err := json.Unmarshal(buf, &monData); err != nil {
		return nil, fmt.Errorf("failed to parse monitor data: %w", err)
	}

	var monitors []models.CephMonitorInfo

	if mons, ok := monData["mons"].([]interface{}); ok {
		for _, mon := range mons {
			if monInfo, ok := mon.(map[string]interface{}); ok {
				monitor := models.CephMonitorInfo{}
				if name, ok := monInfo["name"].(string); ok {
					monitor.Name = name
				}
				if addr, ok := monInfo["public_addr"].(string); ok {
					monitor.Address = addr
				}
				monitor.Status = "up" // 简化处理，假设在mon dump中的都是up状态
				monitor.InQuorum = true
				monitors = append(monitors, monitor)
			}
		}
	}

	return monitors, nil
}

// getClusterVersion 获取集群版本
func (c *CephService) getClusterVersion(ctx context.Context) string {
	cmd := map[string]interface{}{
		"prefix": "version",
		"format": "json",
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return "Unknown"
	}

	buf, _, err := c.conn.MonCommand(cmdBytes)
	if err != nil {
		return "Unknown"
	}

	var versionData map[string]interface{}
	if err := json.Unmarshal(buf, &versionData); err != nil {
		return "Unknown"
	}

	if version, ok := versionData["version"].(string); ok {
		// 移除版本字符串中的哈希部分，只保留版本号和代号
		// 例如: "ceph version 16.2.13 (5378749ba6be3a0868b51803968ee9cde4833a3e) pacific (stable)"
		// 变为: "ceph version 16.2.13 pacific (stable)"
		if idx := strings.Index(version, " ("); idx != -1 {
			if idx2 := strings.Index(version[idx+2:], ") "); idx2 != -1 {
				// 移除括号中的哈希值部分
				cleanVersion := version[:idx] + " " + version[idx+idx2+3:]
				return cleanVersion
			}
		}
		return version
	}

	return "Unknown"
}

// IsConnected 检查是否连接到Ceph集群
func (c *CephService) IsConnected() bool {
	return c.conn != nil
}
