package controllers

import (
	"net/http"

	"k8s-volume-snapshots/models"
	"k8s-volume-snapshots/services"

	"github.com/gin-gonic/gin"
)

type ClusterController struct {
	multiClusterService *services.MultiClusterK8sService
}

func NewClusterController(multiClusterService *services.MultiClusterK8sService) *ClusterController {
	return &ClusterController{
		multiClusterService: multiClusterService,
	}
}

// GetClusters 获取所有集群信息
func (cc *ClusterController) GetClusters(c *gin.Context) {
	clusters, err := cc.multiClusterService.GetClusters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
		return
	}

	response := gin.H{
		"clusters": clusters,
		"current":  cc.multiClusterService.GetCurrentCluster(),
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(response))
}

// GetCurrentCluster 获取当前集群信息
func (cc *ClusterController) GetCurrentCluster(c *gin.Context) {
	currentCluster := cc.multiClusterService.GetCurrentCluster()
	
	clusters, err := cc.multiClusterService.GetClusters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
		return
	}

	// 查找当前集群的详细信息
	var current *models.ClusterInfo
	for _, cluster := range clusters {
		if cluster.Name == currentCluster {
			current = cluster
			break
		}
	}

	if current == nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(404, "Current cluster not found"))
		return
	}

	response := gin.H{
		"cluster": current,
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(response))
}

// SwitchCluster 切换集群
func (cc *ClusterController) SwitchCluster(c *gin.Context) {
	var req models.ClusterSwitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(400, err.Error()))
		return
	}

	if err := cc.multiClusterService.SwitchCluster(req.ClusterName); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(400, err.Error()))
		return
	}

	response := gin.H{
		"message": "Cluster switched successfully",
		"current": req.ClusterName,
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(response))
}