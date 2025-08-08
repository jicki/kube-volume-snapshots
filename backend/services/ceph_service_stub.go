//go:build !ceph

package services

import (
	"context"
	"fmt"
	"time"

	"k8s-volume-snapshots/models"
)

// CephService Ceph集群服务 (Stub 实现)
type CephService struct {
	enabled bool
}

// NewCephService 创建新的Ceph服务实例 (Stub 实现)
func NewCephService() (*CephService, error) {
	return &CephService{enabled: false}, fmt.Errorf("Ceph support not compiled in (use build tag 'ceph' to enable)")
}

// Close 关闭连接 (Stub)
func (c *CephService) Close() {
	// No-op
}

// GetClusterInfo 获取完整的集群信息 (Stub)
func (c *CephService) GetClusterInfo(ctx context.Context) (*models.CephClusterInfo, error) {
	return &models.CephClusterInfo{
		Status: models.CephClusterStatus{
			Health:     "HEALTH_ERR",
			Status:     "Ceph 功能未启用",
			Version:    "N/A",
			Monitors:   make([]models.CephMonitorInfo, 0),
			OSDs:       models.CephOSDSummary{Total: 0, Up: 0, In: 0},
			PGs:        models.CephPGSummary{Total: 0, Active: 0, Clean: 0, Degraded: 0},
			Capacity:   models.CephCapacitySummary{},
			UpdateTime: time.Now(),
		},
		Pools:     make([]models.CephPoolInfo, 0),
		UpdatedAt: time.Now(),
	}, nil
}

// IsConnected 检查是否连接到Ceph集群 (Stub)
func (c *CephService) IsConnected() bool {
	return false
}
