package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s-volume-snapshots/middleware"
	"k8s-volume-snapshots/models"
	"k8s-volume-snapshots/services"
)

type SnapshotController struct {
	k8sService services.K8sServiceInterface
}

func NewSnapshotController(k8sService services.K8sServiceInterface) *SnapshotController {
	return &SnapshotController{
		k8sService: k8sService,
	}
}

// GetVolumeSnapshotClasses 获取 VolumeSnapshotClass 列表
func (c *SnapshotController) GetVolumeSnapshotClasses(ctx *gin.Context) {
	vscList, err := c.k8sService.GetVolumeSnapshotClasses(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
		return
	}

	// 获取相关的 StorageClass 信息
	scList, err := c.k8sService.GetStorageClasses(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
		return
	}

	// 创建 StorageClass 映射
	scMap := make(map[string]*storagev1.StorageClass)
	for i := range scList {
		scMap[scList[i].Name] = &scList[i]
	}

	var result []models.VolumeSnapshotClassInfo
	for _, vsc := range vscList {
		info := models.VolumeSnapshotClassInfo{
			VolumeSnapshotClass: vsc,
		}

		// 尝试找到相关的 StorageClass（基于 driver 匹配）
		for _, sc := range scList {
			if sc.Provisioner == vsc.Driver {
				info.RelatedStorageClass = &sc
				break
			}
		}

		result = append(result, info)
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// GetVolumeSnapshots 获取 VolumeSnapshot 列表
func (c *SnapshotController) GetVolumeSnapshots(ctx *gin.Context) {
	namespace := ctx.Query("namespace")

	vsList, err := c.k8sService.GetVolumeSnapshots(context.Background(), namespace)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
		return
	}

	var result []models.VolumeSnapshotInfo
	for _, vs := range vsList {
		info := models.VolumeSnapshotInfo{
			VolumeSnapshot: vs,
		}

		// 获取对应的 VolumeSnapshotContent
		if vs.Status != nil && vs.Status.BoundVolumeSnapshotContentName != nil {
			vsc, err := c.k8sService.GetVolumeSnapshotContent(context.Background(), *vs.Status.BoundVolumeSnapshotContentName)
			if err == nil {
				info.VolumeSnapshotContent = vsc
			}
		}

		// 获取对应的 PVC 信息
		if vs.Spec.Source.PersistentVolumeClaimName != nil {
			pvcs, err := c.k8sService.GetPVCs(context.Background(), vs.Namespace)
			if err == nil {
				for _, pvc := range pvcs {
					if pvc.Name == *vs.Spec.Source.PersistentVolumeClaimName {
						info.PVC = &pvc
						break
					}
				}
			}
		}

		result = append(result, info)
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(result))
}

// CreateVolumeSnapshot 创建 VolumeSnapshot
func (c *SnapshotController) CreateVolumeSnapshot(ctx *gin.Context) {
	var req models.CreateVolumeSnapshotRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.NewErrorResponse(400, err.Error()))
		return
	}

	// 获取当前用户信息
	username, exists := middleware.GetCurrentUsername(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, models.NewErrorResponse(401, "用户未认证"))
		return
	}

	// 设置创建者信息
	req.CreatedBy = username

	// 创建 VolumeSnapshot 对象
	vs := &snapshotv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels: map[string]string{
				"created-by": username,
				"app":        "k8s-volume-snapshots",
			},
			Annotations: map[string]string{
				"k8s-volume-snapshots/created-by": username,
			},
		},
		Spec: snapshotv1.VolumeSnapshotSpec{
			Source: snapshotv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &req.PVCName,
			},
			VolumeSnapshotClassName: &req.VolumeSnapshotClassName,
		},
	}

	createdVS, err := c.k8sService.CreateVolumeSnapshot(context.Background(), req.Namespace, vs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, models.NewSuccessResponse(createdVS))
}

// DeleteVolumeSnapshot 删除 VolumeSnapshot
func (c *SnapshotController) DeleteVolumeSnapshot(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	// 首先检查快照是否存在
	_, err := c.k8sService.GetVolumeSnapshot(context.Background(), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.NewErrorResponse(404, "快照不存在"))
		return
	}

	// 执行删除操作
	err = c.k8sService.DeleteVolumeSnapshot(context.Background(), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, "删除快照失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(map[string]string{
		"message": "删除请求已提交，快照删除可能需要一些时间完成",
		"name":    name,
	}))
}

// ForceDeleteVolumeSnapshot 强制删除卡住的 VolumeSnapshot
func (c *SnapshotController) ForceDeleteVolumeSnapshot(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	// 获取快照详情
	vs, err := c.k8sService.GetVolumeSnapshot(context.Background(), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.NewErrorResponse(404, "快照不存在"))
		return
	}

	// 检查是否有 deletionTimestamp，表示快照卡在删除状态
	if vs.DeletionTimestamp == nil {
		ctx.JSON(http.StatusBadRequest, models.NewErrorResponse(400, "快照不在删除状态，请使用普通删除"))
		return
	}

	// 尝试强制清理 finalizers
	err = c.k8sService.ForceDeleteVolumeSnapshot(context.Background(), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, "强制删除失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(map[string]string{
		"message": "卡住的快照已强制清理",
		"name":    name,
	}))
}

// GetVolumeSnapshotContent 获取 VolumeSnapshotContent 详情
func (c *SnapshotController) GetVolumeSnapshotContent(ctx *gin.Context) {
	name := ctx.Param("name")

	vsc, err := c.k8sService.GetVolumeSnapshotContent(context.Background(), name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(vsc))
}

// GetPVCs 获取 PVC 列表
func (c *SnapshotController) GetPVCs(ctx *gin.Context) {
	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	pvcs, err := c.k8sService.GetPVCsWithPVInfo(context.Background(), namespace)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(pvcs))
}

// GetNamespaces 获取所有命名空间
func (c *SnapshotController) GetNamespaces(ctx *gin.Context) {
	namespaces, err := c.k8sService.GetNamespaces(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(namespaces))
}

// GetStorageClasses 获取所有存储类
func (c *SnapshotController) GetStorageClasses(ctx *gin.Context) {
	storageClasses, err := c.k8sService.GetStorageClasses(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(storageClasses))
}
