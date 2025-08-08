package services

import (
	"context"
	
	"k8s-volume-snapshots/models"
	
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

// K8sServiceInterface 定义 Kubernetes 服务的通用接口
// 这个接口允许单集群和多集群服务使用相同的控制器代码
type K8sServiceInterface interface {
	// VolumeSnapshotClass 相关方法
	GetVolumeSnapshotClasses(ctx context.Context) ([]snapshotv1.VolumeSnapshotClass, error)
	
	// StorageClass 相关方法
	GetStorageClasses(ctx context.Context) ([]storagev1.StorageClass, error)
	
	// VolumeSnapshot 相关方法
	GetVolumeSnapshots(ctx context.Context, namespace string) ([]snapshotv1.VolumeSnapshot, error)
	CreateVolumeSnapshot(ctx context.Context, namespace string, vs *snapshotv1.VolumeSnapshot) (*snapshotv1.VolumeSnapshot, error)
	GetVolumeSnapshot(ctx context.Context, namespace, name string) (*snapshotv1.VolumeSnapshot, error)
	DeleteVolumeSnapshot(ctx context.Context, namespace, name string) error
	ForceDeleteVolumeSnapshot(ctx context.Context, namespace, name string) error
	
	// VolumeSnapshotContent 相关方法
	GetVolumeSnapshotContent(ctx context.Context, name string) (*snapshotv1.VolumeSnapshotContent, error)
	
	// PVC 相关方法
	GetPVCs(ctx context.Context, namespace string) ([]corev1.PersistentVolumeClaim, error)
	GetPVCsWithPVInfo(ctx context.Context, namespace string) ([]models.PVCWithPVInfo, error)
	
	// Namespace 相关方法
	GetNamespaces(ctx context.Context) ([]corev1.Namespace, error)
}

// MultiClusterK8sServiceInterface 扩展接口，支持多集群操作
type MultiClusterK8sServiceInterface interface {
	K8sServiceInterface
	
	// 多集群管理方法
	GetClusters() ([]*models.ClusterInfo, error)
	GetCurrentCluster() string
	SwitchCluster(clusterName string) error
	
	// 在指定集群中执行操作
	CreateVolumeSnapshotInCluster(ctx context.Context, clusterName, namespace string, vs *snapshotv1.VolumeSnapshot) (*snapshotv1.VolumeSnapshot, error)
	GetPVCsInCluster(ctx context.Context, clusterName, namespace string) ([]corev1.PersistentVolumeClaim, error)
}