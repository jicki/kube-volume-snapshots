package controllers

import (
	"context"
	"net/http"
	"time"

	"k8s-volume-snapshots/models"
	"k8s-volume-snapshots/services"

	"github.com/gin-gonic/gin"
)

// CephController Ceph集群控制器
type CephController struct {
	cephService *services.CephService
}

// NewCephController 创建新的Ceph控制器
func NewCephController(cephService *services.CephService) *CephController {
	return &CephController{
		cephService: cephService,
	}
}

// GetClusterInfo 获取完整的集群信息
// @Summary 获取Ceph集群信息
// @Description 获取Ceph集群的状态、Pool信息等
// @Tags Ceph
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=models.CephClusterInfo} "成功"
// @Failure 500 {object} models.APIResponse "服务器错误"
// @Router /api/ceph/cluster/info [get]
func (c *CephController) GetClusterInfo(ctx *gin.Context) {
	// 设置超时上下文
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 检查服务是否可用
	if !c.cephService.IsConnected() {
		ctx.JSON(http.StatusServiceUnavailable, models.NewErrorResponse(
			http.StatusServiceUnavailable,
			"Ceph集群连接不可用，请检查配置和网络连接",
		))
		return
	}

	clusterInfo, err := c.cephService.GetClusterInfo(timeoutCtx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"获取Ceph集群信息失败: "+err.Error(),
		))
		return
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(clusterInfo))
}

// GetClusterStatus 获取集群状态信息
// @Summary 获取Ceph集群状态
// @Description 获取Ceph集群的健康状态、Monitor、OSD等基本信息
// @Tags Ceph
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=models.CephClusterStatus} "成功"
// @Failure 500 {object} models.APIResponse "服务器错误"
// @Router /api/ceph/cluster/status [get]
func (c *CephController) GetClusterStatus(ctx *gin.Context) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !c.cephService.IsConnected() {
		// 返回模拟的离线状态
		offlineStatus := &models.CephClusterStatus{
			Health:     "HEALTH_ERR",
			Status:     "集群连接不可用",
			Version:    "Unknown",
			Monitors:   make([]models.CephMonitorInfo, 0),
			OSDs:       models.CephOSDSummary{Total: 0, Up: 0, In: 0},
			PGs:        models.CephPGSummary{Total: 0, Active: 0, Clean: 0, Degraded: 0},
			Capacity:   models.CephCapacitySummary{},
			UpdateTime: time.Now(),
		}

		ctx.JSON(http.StatusOK, models.NewSuccessResponse(offlineStatus))
		return
	}

	clusterInfo, err := c.cephService.GetClusterInfo(timeoutCtx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"获取集群状态失败: "+err.Error(),
		))
		return
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(clusterInfo.Status))
}

// GetPoolsInfo 获取Pool信息
// @Summary 获取Ceph Pool信息
// @Description 获取所有Pool的详细信息，包括容量、使用率等
// @Tags Ceph
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]models.CephPoolInfo} "成功"
// @Failure 500 {object} models.APIResponse "服务器错误"
// @Router /api/ceph/pools [get]
func (c *CephController) GetPoolsInfo(ctx *gin.Context) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !c.cephService.IsConnected() {
		ctx.JSON(http.StatusServiceUnavailable, models.NewErrorResponse(
			http.StatusServiceUnavailable,
			"Ceph集群连接不可用",
		))
		return
	}

	clusterInfo, err := c.cephService.GetClusterInfo(timeoutCtx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			http.StatusInternalServerError,
			"获取Pool信息失败: "+err.Error(),
		))
		return
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(clusterInfo.Pools))
}

// GetConnectionStatus 获取连接状态
// @Summary 获取Ceph连接状态
// @Description 检查与Ceph集群的连接状态
// @Tags Ceph
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=map[string]interface{}} "成功"
// @Router /api/ceph/connection/status [get]
func (c *CephController) GetConnectionStatus(ctx *gin.Context) {
	status := map[string]interface{}{
		"connected":   c.cephService.IsConnected(),
		"checkedAt":   time.Now(),
		"serviceName": "Ceph集群管理服务",
	}

	if c.cephService.IsConnected() {
		status["message"] = "已连接到Ceph集群"
	} else {
		status["message"] = "未连接到Ceph集群，请检查配置"
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(status))
}
