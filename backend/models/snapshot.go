package models

import (
	"time"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

// VolumeSnapshotClassInfo VolumeSnapshotClass 信息
type VolumeSnapshotClassInfo struct {
	VolumeSnapshotClass snapshotv1.VolumeSnapshotClass `json:"volumeSnapshotClass"`
	RelatedStorageClass *storagev1.StorageClass        `json:"relatedStorageClass,omitempty"`
}

// VolumeSnapshotInfo VolumeSnapshot 信息
type VolumeSnapshotInfo struct {
	VolumeSnapshot        snapshotv1.VolumeSnapshot         `json:"volumeSnapshot"`
	VolumeSnapshotContent *snapshotv1.VolumeSnapshotContent `json:"volumeSnapshotContent,omitempty"`
	PVC                   *corev1.PersistentVolumeClaim     `json:"pvc,omitempty"`
}

// CreateVolumeSnapshotRequest 创建 VolumeSnapshot 请求
type CreateVolumeSnapshotRequest struct {
	Name                    string `json:"name" binding:"required"`
	Namespace               string `json:"namespace" binding:"required"`
	PVCName                 string `json:"pvcName" binding:"required"`
	VolumeSnapshotClassName string `json:"volumeSnapshotClassName" binding:"required"`
	CreatedBy               string `json:"createdBy,omitempty"` // 创建者用户名
}

// ScheduledSnapshot 定时快照任务
type ScheduledSnapshot struct {
	ID                      string     `json:"id"`
	Name                    string     `json:"name" binding:"required"`
	Namespace               string     `json:"namespace" binding:"required"`
	PVCName                 string     `json:"pvcName" binding:"required"`
	VolumeSnapshotClassName string     `json:"volumeSnapshotClassName" binding:"required"`
	CronExpression          string     `json:"cronExpression" binding:"required"`
	Enabled                 bool       `json:"enabled"`
	CreatedBy               string     `json:"createdBy,omitempty"` // 创建者用户名
	CreatedAt               time.Time  `json:"createdAt"`
	UpdatedAt               time.Time  `json:"updatedAt"`
	LastExecuted            *time.Time `json:"lastExecuted,omitempty"`
	NextExecution           *time.Time `json:"nextExecution,omitempty"`
	TargetClusters          []string   `json:"targetClusters,omitempty"` // 目标集群列表，为空时仅在当前集群执行
}

// ScheduledSnapshotStatus 定时快照任务状态
type ScheduledSnapshotStatus struct {
	ID            string     `json:"id"`
	Enabled       bool       `json:"enabled"`
	LastExecuted  *time.Time `json:"lastExecuted,omitempty"`
	NextExecution *time.Time `json:"nextExecution,omitempty"`
	ErrorMessage  string     `json:"errorMessage,omitempty"`
}

// APIResponse 统一 API 响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) APIResponse {
	return APIResponse{
		Code:    code,
		Message: message,
	}
}

// PVVolumeAttributes PV的VolumeAttributes信息
type PVVolumeAttributes struct {
	ImageName string `json:"imageName,omitempty"`
	Pool      string `json:"pool,omitempty"`
}

// PVCWithPVInfo 包含PV信息的PVC
type PVCWithPVInfo struct {
	PVC              corev1.PersistentVolumeClaim `json:"pvc"`
	VolumeAttributes *PVVolumeAttributes          `json:"volumeAttributes,omitempty"`
}

// User 用户模型
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username" binding:"required"`
	Password  string    `json:"password,omitempty"` // 返回时不包含密码
	Role      string    `json:"role" binding:"required,oneof=readonly admin"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=readonly admin"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}
