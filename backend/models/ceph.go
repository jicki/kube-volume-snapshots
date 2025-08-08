package models

import (
	"fmt"
	"time"
)

// CephClusterStatus Ceph集群状态信息
type CephClusterStatus struct {
	Health     string              `json:"health"`     // HEALTH_OK, HEALTH_WARN, HEALTH_ERR
	Status     string              `json:"status"`     // 集群整体状态描述
	Version    string              `json:"version"`    // Ceph版本
	Monitors   []CephMonitorInfo   `json:"monitors"`   // Monitor节点信息
	OSDs       CephOSDSummary      `json:"osds"`       // OSD汇总信息
	PGs        CephPGSummary       `json:"pgs"`        // PG汇总信息
	Capacity   CephCapacitySummary `json:"capacity"`   // 容量汇总
	UpdateTime time.Time           `json:"updateTime"` // 更新时间
}

// CephMonitorInfo Monitor节点信息
type CephMonitorInfo struct {
	Name     string `json:"name"`     // Monitor名称
	Address  string `json:"address"`  // 地址
	Status   string `json:"status"`   // 状态 (up/down)
	InQuorum bool   `json:"inQuorum"` // 是否在quorum中
}

// CephOSDSummary OSD汇总信息
type CephOSDSummary struct {
	Total int `json:"total"` // 总OSD数量
	Up    int `json:"up"`    // 在线OSD数量
	In    int `json:"in"`    // 加入集群的OSD数量
}

// CephPGSummary PG汇总信息
type CephPGSummary struct {
	Total    int `json:"total"`    // 总PG数量
	Active   int `json:"active"`   // 活跃PG数量
	Clean    int `json:"clean"`    // 干净PG数量
	Degraded int `json:"degraded"` // 降级PG数量
}

// CephCapacitySummary 容量汇总信息
type CephCapacitySummary struct {
	TotalBytes   int64   `json:"totalBytes"`   // 总容量(字节)
	UsedBytes    int64   `json:"usedBytes"`    // 已用容量(字节)
	AvailBytes   int64   `json:"availBytes"`   // 可用容量(字节)
	UsagePercent float64 `json:"usagePercent"` // 使用率百分比
}

// CephPoolInfo Pool详细信息
type CephPoolInfo struct {
	ID             int     `json:"id"`             // Pool ID
	Name           string  `json:"name"`           // Pool名称
	Type           string  `json:"type"`           // Pool类型 (replicated/erasure)
	Size           int     `json:"size"`           // 副本数或EC配置
	MinSize        int     `json:"minSize"`        // 最小副本数
	PGNum          int     `json:"pgNum"`          // PG数量
	PGPNum         int     `json:"pgpNum"`         // PGP数量
	Objects        int64   `json:"objects"`        // 对象数量
	UsedBytes      int64   `json:"usedBytes"`      // 已用字节数
	MaxAvailBytes  int64   `json:"maxAvailBytes"`  // 最大可用字节数
	UsagePercent   float64 `json:"usagePercent"`   // 使用率百分比
	ReadIOPS       int64   `json:"readIOPS"`       // 读IOPS
	WriteIOPS      int64   `json:"writeIOPS"`      // 写IOPS
	ReadBandwidth  int64   `json:"readBandwidth"`  // 读带宽(字节/秒)
	WriteBandwidth int64   `json:"writeBandwidth"` // 写带宽(字节/秒)
}

// CephClusterInfo 完整的Ceph集群信息
type CephClusterInfo struct {
	Status    CephClusterStatus `json:"status"`    // 集群状态
	Pools     []CephPoolInfo    `json:"pools"`     // Pool列表
	UpdatedAt time.Time         `json:"updatedAt"` // 更新时间
}

// CephHealthStatus Ceph健康状态枚举
const (
	CephHealthOK   = "HEALTH_OK"
	CephHealthWarn = "HEALTH_WARN"
	CephHealthErr  = "HEALTH_ERR"
)

// GetHealthStatusType 获取健康状态对应的前端显示类型
func (c *CephClusterStatus) GetHealthStatusType() string {
	switch c.Health {
	case CephHealthOK:
		return "success"
	case CephHealthWarn:
		return "warning"
	case CephHealthErr:
		return "danger"
	default:
		return "info"
	}
}

// FormatBytes 格式化字节数为人类可读格式
func FormatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	var i int
	value := float64(bytes)

	for value >= 1024 && i < len(units)-1 {
		value /= 1024
		i++
	}

	return fmt.Sprintf("%.2f %s", value, units[i])
}
