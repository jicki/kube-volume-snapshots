package services

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"k8s-volume-snapshots/models"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	snapshotclientset "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8sService struct {
	ClientSet         *kubernetes.Clientset
	SnapshotClientSet *snapshotclientset.Clientset
	Config            *rest.Config
	// PV信息缓存
	pvCache      map[string]*corev1.PersistentVolume
	pvCacheMutex sync.RWMutex
	pvCacheTime  time.Time
}

func NewK8sService() (*K8sService, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	// 设置客户端超时时间
	config.Timeout = 30 * time.Second
	config.QPS = 50    // 增加QPS限制
	config.Burst = 100 // 增加突发请求限制

	// 创建标准 Kubernetes 客户端
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// 创建 Snapshot 客户端
	snapshotClientSet, err := snapshotclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8sService{
		ClientSet:         clientSet,
		SnapshotClientSet: snapshotClientSet,
		Config:            config,
		pvCache:           make(map[string]*corev1.PersistentVolume),
		pvCacheMutex:      sync.RWMutex{},
	}, nil
}

func getConfig() (*rest.Config, error) {
	// 尝试使用集群内配置
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	// 使用 kubeconfig 文件
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetVolumeSnapshotClasses 获取所有 VolumeSnapshotClass
func (k *K8sService) GetVolumeSnapshotClasses(ctx context.Context) ([]snapshotv1.VolumeSnapshotClass, error) {
	vscList, err := k.SnapshotClientSet.SnapshotV1().VolumeSnapshotClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return vscList.Items, nil
}

// GetStorageClasses 获取所有 StorageClass
func (k *K8sService) GetStorageClasses(ctx context.Context) ([]storagev1.StorageClass, error) {
	scList, err := k.ClientSet.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return scList.Items, nil
}

// GetVolumeSnapshots 获取 VolumeSnapshot 列表
func (k *K8sService) GetVolumeSnapshots(ctx context.Context, namespace string) ([]snapshotv1.VolumeSnapshot, error) {
	if namespace == "" || namespace == "all" {
		vsList, err := k.SnapshotClientSet.SnapshotV1().VolumeSnapshots("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return vsList.Items, nil
	}

	vsList, err := k.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return vsList.Items, nil
}

// CreateVolumeSnapshot 创建 VolumeSnapshot
func (k *K8sService) CreateVolumeSnapshot(ctx context.Context, namespace string, vs *snapshotv1.VolumeSnapshot) (*snapshotv1.VolumeSnapshot, error) {
	return k.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).Create(ctx, vs, metav1.CreateOptions{})
}

// GetVolumeSnapshot 获取单个 VolumeSnapshot
func (k *K8sService) GetVolumeSnapshot(ctx context.Context, namespace, name string) (*snapshotv1.VolumeSnapshot, error) {
	return k.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteVolumeSnapshot 删除 VolumeSnapshot
func (k *K8sService) DeleteVolumeSnapshot(ctx context.Context, namespace, name string) error {
	return k.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ForceDeleteVolumeSnapshot 强制删除卡住的 VolumeSnapshot（移除 finalizers）
func (k *K8sService) ForceDeleteVolumeSnapshot(ctx context.Context, namespace, name string) error {
	// 获取快照
	vs, err := k.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// 移除 finalizers
	vs.Finalizers = nil
	_, err = k.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).Update(ctx, vs, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	// 如果有关联的 VolumeSnapshotContent，也要清理
	if vs.Status != nil && vs.Status.BoundVolumeSnapshotContentName != nil {
		vsc, err := k.SnapshotClientSet.SnapshotV1().VolumeSnapshotContents().Get(ctx, *vs.Status.BoundVolumeSnapshotContentName, metav1.GetOptions{})
		if err == nil {
			vsc.Finalizers = nil
			_, _ = k.SnapshotClientSet.SnapshotV1().VolumeSnapshotContents().Update(ctx, vsc, metav1.UpdateOptions{})
		}
	}

	return nil
}

// GetVolumeSnapshotContent 获取 VolumeSnapshotContent
func (k *K8sService) GetVolumeSnapshotContent(ctx context.Context, name string) (*snapshotv1.VolumeSnapshotContent, error) {
	return k.SnapshotClientSet.SnapshotV1().VolumeSnapshotContents().Get(ctx, name, metav1.GetOptions{})
}

// GetPVCs 获取指定命名空间的 PVC 列表
func (k *K8sService) GetPVCs(ctx context.Context, namespace string) ([]corev1.PersistentVolumeClaim, error) {
	if namespace == "" {
		namespace = "default"
	}

	// 支持查询所有命名空间
	if namespace == "all" {
		pvcList, err := k.ClientSet.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return pvcList.Items, nil
	}

	pvcList, err := k.ClientSet.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pvcList.Items, nil
}

// refreshPVCache 刷新PV缓存
func (k *K8sService) refreshPVCache(ctx context.Context) error {
	// 缓存有效期5分钟
	k.pvCacheMutex.RLock()
	if time.Since(k.pvCacheTime) < 5*time.Minute && len(k.pvCache) > 0 {
		k.pvCacheMutex.RUnlock()
		return nil
	}
	k.pvCacheMutex.RUnlock()

	// 批量获取所有PV
	pvList, err := k.ClientSet.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// 更新缓存
	k.pvCacheMutex.Lock()
	defer k.pvCacheMutex.Unlock()

	newCache := make(map[string]*corev1.PersistentVolume)
	for i := range pvList.Items {
		pv := &pvList.Items[i]
		newCache[pv.Name] = pv
	}

	k.pvCache = newCache
	k.pvCacheTime = time.Now()

	return nil
}

// getPVFromCache 从缓存获取PV信息
func (k *K8sService) getPVFromCache(pvName string) *corev1.PersistentVolume {
	k.pvCacheMutex.RLock()
	defer k.pvCacheMutex.RUnlock()
	return k.pvCache[pvName]
}

// GetPVCsWithPVInfo 获取包含PV详细信息的PVC列表（优化版本）
func (k *K8sService) GetPVCsWithPVInfo(ctx context.Context, namespace string) ([]models.PVCWithPVInfo, error) {
	if namespace == "" {
		namespace = "default"
	}

	// 先刷新PV缓存
	if err := k.refreshPVCache(ctx); err != nil {
		// 如果刷新缓存失败，记录错误但继续执行（使用旧缓存或空缓存）
		// 在生产环境中应该使用适当的日志记录
	}

	var pvcList *corev1.PersistentVolumeClaimList
	var err error

	// 支持查询所有命名空间
	if namespace == "all" {
		pvcList, err = k.ClientSet.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
	} else {
		pvcList, err = k.ClientSet.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
	}

	result := make([]models.PVCWithPVInfo, 0, len(pvcList.Items))

	for _, pvc := range pvcList.Items {
		pvcWithPV := models.PVCWithPVInfo{
			PVC: pvc,
		}

		// 如果PVC绑定了PV，从缓存获取PV的VolumeAttributes信息
		if pvc.Spec.VolumeName != "" {
			pv := k.getPVFromCache(pvc.Spec.VolumeName)
			if pv != nil && pv.Spec.CSI != nil && pv.Spec.CSI.VolumeAttributes != nil {
				pvcWithPV.VolumeAttributes = &models.PVVolumeAttributes{
					ImageName: pv.Spec.CSI.VolumeAttributes["imageName"],
					Pool:      pv.Spec.CSI.VolumeAttributes["pool"],
				}
			} else {
				// 设置空的VolumeAttributes来确保字段存在
				pvcWithPV.VolumeAttributes = &models.PVVolumeAttributes{}
			}
		} else {
			// 如果没有绑定PV，设置空的VolumeAttributes
			pvcWithPV.VolumeAttributes = &models.PVVolumeAttributes{}
		}

		result = append(result, pvcWithPV)
	}

	return result, nil
}

// GetNamespaces 获取所有命名空间
func (k *K8sService) GetNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	nsList, err := k.ClientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nsList.Items, nil
}
