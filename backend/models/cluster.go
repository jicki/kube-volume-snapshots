package models

import "time"

// ClusterConfig 集群配置结构
type ClusterConfig struct {
	Name        string `yaml:"name" json:"name"`
	DisplayName string `yaml:"display_name" json:"display_name"`
	Description string `yaml:"description" json:"description"`
	KubeConfig  string `yaml:"kubeconfig" json:"kubeconfig"`
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	
	// 可选的直接连接配置
	Server                   string `yaml:"server,omitempty" json:"server,omitempty"`
	Token                    string `yaml:"token,omitempty" json:"token,omitempty"`
	CertificateAuthorityData string `yaml:"certificate_authority_data,omitempty" json:"certificate_authority_data,omitempty"`
}

// GlobalConfig 全局配置结构
type GlobalConfig struct {
	Timeout              int `yaml:"timeout" json:"timeout"`
	QPS                  int `yaml:"qps" json:"qps"`
	Burst                int `yaml:"burst" json:"burst"`
	CacheRefreshInterval int `yaml:"cache_refresh_interval" json:"cache_refresh_interval"`
}

// MultiClusterConfig 多集群配置结构
type MultiClusterConfig struct {
	DefaultCluster string           `yaml:"default_cluster" json:"default_cluster"`
	Clusters       []*ClusterConfig `yaml:"clusters" json:"clusters"`
	Global         GlobalConfig     `yaml:"global" json:"global"`
}

// ClusterInfo 集群信息（用于前端显示）
type ClusterInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Status      string `json:"status"`      // online, offline, error
	LastCheck   time.Time `json:"last_check"`
}

// ClusterSwitchRequest 切换集群请求
type ClusterSwitchRequest struct {
	ClusterName string `json:"cluster_name" binding:"required"`
}