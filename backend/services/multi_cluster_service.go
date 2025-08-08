package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	"gopkg.in/yaml.v2"
)

type MultiClusterK8sService struct {
	config           *models.MultiClusterConfig
	clusters         map[string]*ClusterClient
	currentCluster   string
	mutex            sync.RWMutex
	configPath       string
	clusterInfoCache map[string]*models.ClusterInfo
	cacheMutex       sync.RWMutex
	lastCacheUpdate  time.Time
}

type ClusterClient struct {
	Config            *rest.Config
	ClientSet         *kubernetes.Clientset
	SnapshotClientSet *snapshotclientset.Clientset
	ClusterInfo       *models.ClusterConfig
	Status            string
	LastCheck         time.Time
	// PV缓存
	pvCache      map[string]*corev1.PersistentVolume
	pvCacheMutex sync.RWMutex
	pvCacheTime  time.Time
}

func NewMultiClusterK8sService() (*MultiClusterK8sService, error) {
	configPath := getMultiClusterConfigPath()
	
	service := &MultiClusterK8sService{
		clusters:         make(map[string]*ClusterClient),
		configPath:       configPath,
		clusterInfoCache: make(map[string]*models.ClusterInfo),
	}

	// 尝试加载配置文件
	if err := service.loadConfig(); err != nil {
		log.Printf("Failed to load multi-cluster config, falling back to single cluster mode: %v", err)
		// 如果加载失败，使用默认单集群配置
		return service.createFallbackService()
	}

	// 初始化所有启用的集群
	if err := service.initializeClusters(); err != nil {
		return nil, fmt.Errorf("failed to initialize clusters: %v", err)
	}

	// 设置默认当前集群
	service.currentCluster = service.config.DefaultCluster

	return service, nil
}

func getMultiClusterConfigPath() string {
	// 优先级：环境变量 > 相对路径 > 默认路径
	if configPath := os.Getenv("MULTI_CLUSTER_CONFIG"); configPath != "" {
		return configPath
	}
	
	// 检查相对路径
	if _, err := os.Stat("config/clusters.yaml"); err == nil {
		return "config/clusters.yaml"
	}
	
	// 默认路径
	return "/etc/k8s-volume-snapshots/clusters.yaml"
}

func (m *MultiClusterK8sService) loadConfig() error {
	data, err := ioutil.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	config := &models.MultiClusterConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	// 设置默认值
	if config.Global.Timeout == 0 {
		config.Global.Timeout = 30
	}
	if config.Global.QPS == 0 {
		config.Global.QPS = 50
	}
	if config.Global.Burst == 0 {
		config.Global.Burst = 100
	}
	if config.Global.CacheRefreshInterval == 0 {
		config.Global.CacheRefreshInterval = 5
	}

	m.config = config
	return nil
}

func (m *MultiClusterK8sService) createFallbackService() (*MultiClusterK8sService, error) {
	// 创建默认单集群配置
	defaultConfig := &models.MultiClusterConfig{
		DefaultCluster: "default",
		Clusters: []*models.ClusterConfig{
			{
				Name:        "default",
				DisplayName: "默认集群",
				Description: "默认 Kubernetes 集群",
				Enabled:     true,
			},
		},
		Global: models.GlobalConfig{
			Timeout:              30,
			QPS:                  50,
			Burst:                100,
			CacheRefreshInterval: 5,
		},
	}

	m.config = defaultConfig
	m.currentCluster = "default"

	// 初始化默认集群
	if err := m.initializeClusters(); err != nil {
		return nil, fmt.Errorf("failed to initialize default cluster: %v", err)
	}

	return m, nil
}

func (m *MultiClusterK8sService) initializeClusters() error {
	for _, clusterConfig := range m.config.Clusters {
		if !clusterConfig.Enabled {
			continue
		}

		client, err := m.createClusterClient(clusterConfig)
		if err != nil {
			log.Printf("Failed to initialize cluster %s: %v", clusterConfig.Name, err)
			// 创建一个标记为错误状态的客户端
			client = &ClusterClient{
				ClusterInfo: clusterConfig,
				Status:      "error",
				LastCheck:   time.Now(),
				pvCache:     make(map[string]*corev1.PersistentVolume),
			}
		} else {
			client.Status = "online"
			client.LastCheck = time.Now()
		}

		m.clusters[clusterConfig.Name] = client
	}

	return nil
}

func (m *MultiClusterK8sService) createClusterClient(clusterConfig *models.ClusterConfig) (*ClusterClient, error) {
	var config *rest.Config
	var err error

	// 如果提供了直接连接配置，使用这些配置
	if clusterConfig.Server != "" {
		config = &rest.Config{
			Host:        clusterConfig.Server,
			BearerToken: clusterConfig.Token,
		}
		if clusterConfig.CertificateAuthorityData != "" {
			config.CAData = []byte(clusterConfig.CertificateAuthorityData)
		} else {
			config.Insecure = true
		}
	} else if clusterConfig.KubeConfig != "" {
		// 使用 kubeconfig 文件
		config, err = clientcmd.BuildConfigFromFlags("", clusterConfig.KubeConfig)
		if err != nil {
			return nil, err
		}
	} else {
		// 尝试使用集群内配置或默认配置
		if config, err = rest.InClusterConfig(); err != nil {
			// 使用默认 kubeconfig
			var kubeconfig string
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return nil, err
			}
		}
	}

	// 设置客户端配置
	config.Timeout = time.Duration(m.config.Global.Timeout) * time.Second
	config.QPS = float32(m.config.Global.QPS)
	config.Burst = m.config.Global.Burst

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

	return &ClusterClient{
		Config:            config,
		ClientSet:         clientSet,
		SnapshotClientSet: snapshotClientSet,
		ClusterInfo:       clusterConfig,
		pvCache:           make(map[string]*corev1.PersistentVolume),
	}, nil
}

// GetCurrentCluster 获取当前集群名称
func (m *MultiClusterK8sService) GetCurrentCluster() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.currentCluster
}

// SwitchCluster 切换到指定集群
func (m *MultiClusterK8sService) SwitchCluster(clusterName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.clusters[clusterName]; !exists {
		return fmt.Errorf("cluster %s not found", clusterName)
	}

	if !m.clusters[clusterName].ClusterInfo.Enabled {
		return fmt.Errorf("cluster %s is disabled", clusterName)
	}

	if m.clusters[clusterName].Status == "error" {
		return fmt.Errorf("cluster %s is in error state", clusterName)
	}

	m.currentCluster = clusterName
	return nil
}

// GetClusters 获取所有集群信息
func (m *MultiClusterK8sService) GetClusters() ([]*models.ClusterInfo, error) {
	m.cacheMutex.RLock()
	if time.Since(m.lastCacheUpdate) < time.Duration(m.config.Global.CacheRefreshInterval)*time.Minute &&
		len(m.clusterInfoCache) > 0 {
		result := make([]*models.ClusterInfo, 0, len(m.clusterInfoCache))
		for _, info := range m.clusterInfoCache {
			result = append(result, info)
		}
		m.cacheMutex.RUnlock()
		return result, nil
	}
	m.cacheMutex.RUnlock()

	// 刷新集群状态
	m.mutex.RLock()
	clusters := make([]*models.ClusterInfo, 0, len(m.clusters))
	for _, client := range m.clusters {
		status := m.checkClusterStatus(client)
		info := &models.ClusterInfo{
			Name:        client.ClusterInfo.Name,
			DisplayName: client.ClusterInfo.DisplayName,
			Description: client.ClusterInfo.Description,
			Enabled:     client.ClusterInfo.Enabled,
			Status:      status,
			LastCheck:   client.LastCheck,
		}
		clusters = append(clusters, info)
	}
	m.mutex.RUnlock()

	// 更新缓存
	m.cacheMutex.Lock()
	m.clusterInfoCache = make(map[string]*models.ClusterInfo)
	for _, info := range clusters {
		m.clusterInfoCache[info.Name] = info
	}
	m.lastCacheUpdate = time.Now()
	m.cacheMutex.Unlock()

	return clusters, nil
}

func (m *MultiClusterK8sService) checkClusterStatus(client *ClusterClient) string {
	if !client.ClusterInfo.Enabled {
		return "disabled"
	}

	if client.ClientSet == nil {
		return "error"
	}

	// 检查集群连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.ClientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		client.Status = "offline"
		client.LastCheck = time.Now()
		return "offline"
	}

	client.Status = "online"
	client.LastCheck = time.Now()
	return "online"
}

// GetCurrentClient 获取当前集群的客户端
func (m *MultiClusterK8sService) GetCurrentClient() (*ClusterClient, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	client, exists := m.clusters[m.currentCluster]
	if !exists {
		return nil, fmt.Errorf("current cluster %s not found", m.currentCluster)
	}

	if client.Status == "error" || client.ClientSet == nil {
		return nil, fmt.Errorf("current cluster %s is not available", m.currentCluster)
	}

	return client, nil
}

// 以下方法将当前集群的操作转发到具体的客户端
// 这样可以保持与原有 K8sService 接口的兼容性

func (m *MultiClusterK8sService) GetVolumeSnapshotClasses(ctx context.Context) ([]snapshotv1.VolumeSnapshotClass, error) {
	client, err := m.GetCurrentClient()
	if err != nil {
		return nil, err
	}

	vscList, err := client.SnapshotClientSet.SnapshotV1().VolumeSnapshotClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return vscList.Items, nil
}

func (m *MultiClusterK8sService) GetStorageClasses(ctx context.Context) ([]storagev1.StorageClass, error) {
	client, err := m.GetCurrentClient()
	if err != nil {
		return nil, err
	}

	scList, err := client.ClientSet.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return scList.Items, nil
}

func (m *MultiClusterK8sService) GetVolumeSnapshots(ctx context.Context, namespace string) ([]snapshotv1.VolumeSnapshot, error) {
	client, err := m.GetCurrentClient()
	if err != nil {
		return nil, err
	}

	if namespace == "" || namespace == "all" {
		vsList, err := client.SnapshotClientSet.SnapshotV1().VolumeSnapshots("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return vsList.Items, nil
	}

	vsList, err := client.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return vsList.Items, nil
}

func (m *MultiClusterK8sService) CreateVolumeSnapshot(ctx context.Context, namespace string, vs *snapshotv1.VolumeSnapshot) (*snapshotv1.VolumeSnapshot, error) {
	client, err := m.GetCurrentClient()
	if err != nil {
		return nil, err
	}

	return client.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).Create(ctx, vs, metav1.CreateOptions{})
}

// CreateVolumeSnapshotInCluster 在指定集群中创建VolumeSnapshot
func (m *MultiClusterK8sService) CreateVolumeSnapshotInCluster(ctx context.Context, clusterName, namespace string, vs *snapshotv1.VolumeSnapshot) (*snapshotv1.VolumeSnapshot, error) {
	m.mutex.RLock()
	client, exists := m.clusters[clusterName]
	m.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}
	
	if !client.ClusterInfo.Enabled {
		return nil, fmt.Errorf("cluster %s is disabled", clusterName)
	}
	
	if client.Status == "error" || client.ClientSet == nil {
		return nil, fmt.Errorf("cluster %s is not available", clusterName)
	}
	
	return client.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).Create(ctx, vs, metav1.CreateOptions{})
}

// GetPVCsInCluster 获取指定集群中的PVC列表
func (m *MultiClusterK8sService) GetPVCsInCluster(ctx context.Context, clusterName, namespace string) ([]corev1.PersistentVolumeClaim, error) {
	m.mutex.RLock()
	client, exists := m.clusters[clusterName]
	m.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}
	
	if !client.ClusterInfo.Enabled {
		return nil, fmt.Errorf("cluster %s is disabled", clusterName)
	}
	
	if client.Status == "error" || client.ClientSet == nil {
		return nil, fmt.Errorf("cluster %s is not available", clusterName)
	}
	
	if namespace == "" {
		namespace = "default"
	}
	
	// 支持查询所有命名空间
	if namespace == "all" {
		pvcList, err := client.ClientSet.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return pvcList.Items, nil
	}
	
	pvcList, err := client.ClientSet.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pvcList.Items, nil
}

func (m *MultiClusterK8sService) GetVolumeSnapshot(ctx context.Context, namespace, name string) (*snapshotv1.VolumeSnapshot, error) {
	client, err := m.GetCurrentClient()
	if err != nil {
		return nil, err
	}

	return client.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (m *MultiClusterK8sService) DeleteVolumeSnapshot(ctx context.Context, namespace, name string) error {
	client, err := m.GetCurrentClient()
	if err != nil {
		return err
	}

	return client.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (m *MultiClusterK8sService) ForceDeleteVolumeSnapshot(ctx context.Context, namespace, name string) error {
	client, err := m.GetCurrentClient()
	if err != nil {
		return err
	}

	// 获取快照
	vs, err := client.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// 移除 finalizers
	vs.Finalizers = nil
	_, err = client.SnapshotClientSet.SnapshotV1().VolumeSnapshots(namespace).Update(ctx, vs, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	// 如果有关联的 VolumeSnapshotContent，也要清理
	if vs.Status != nil && vs.Status.BoundVolumeSnapshotContentName != nil {
		vsc, err := client.SnapshotClientSet.SnapshotV1().VolumeSnapshotContents().Get(ctx, *vs.Status.BoundVolumeSnapshotContentName, metav1.GetOptions{})
		if err == nil {
			vsc.Finalizers = nil
			_, _ = client.SnapshotClientSet.SnapshotV1().VolumeSnapshotContents().Update(ctx, vsc, metav1.UpdateOptions{})
		}
	}

	return nil
}

func (m *MultiClusterK8sService) GetVolumeSnapshotContent(ctx context.Context, name string) (*snapshotv1.VolumeSnapshotContent, error) {
	client, err := m.GetCurrentClient()
	if err != nil {
		return nil, err
	}

	return client.SnapshotClientSet.SnapshotV1().VolumeSnapshotContents().Get(ctx, name, metav1.GetOptions{})
}

func (m *MultiClusterK8sService) GetPVCs(ctx context.Context, namespace string) ([]corev1.PersistentVolumeClaim, error) {
	client, err := m.GetCurrentClient()
	if err != nil {
		return nil, err
	}

	if namespace == "" {
		namespace = "default"
	}

	// 支持查询所有命名空间
	if namespace == "all" {
		pvcList, err := client.ClientSet.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return pvcList.Items, nil
	}

	pvcList, err := client.ClientSet.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pvcList.Items, nil
}

func (m *MultiClusterK8sService) GetPVCsWithPVInfo(ctx context.Context, namespace string) ([]models.PVCWithPVInfo, error) {
	client, err := m.GetCurrentClient()
	if err != nil {
		return nil, err
	}

	if namespace == "" {
		namespace = "default"
	}

	// 先刷新PV缓存
	if err := m.refreshPVCache(ctx, client); err != nil {
		// 如果刷新缓存失败，记录错误但继续执行
	}

	var pvcList *corev1.PersistentVolumeClaimList

	// 支持查询所有命名空间
	if namespace == "all" {
		pvcList, err = client.ClientSet.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
	} else {
		pvcList, err = client.ClientSet.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
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
			pv := m.getPVFromCache(client, pvc.Spec.VolumeName)
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

func (m *MultiClusterK8sService) GetNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	client, err := m.GetCurrentClient()
	if err != nil {
		return nil, err
	}

	nsList, err := client.ClientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nsList.Items, nil
}

// refreshPVCache 刷新指定客户端的PV缓存
func (m *MultiClusterK8sService) refreshPVCache(ctx context.Context, client *ClusterClient) error {
	// 缓存有效期5分钟
	client.pvCacheMutex.RLock()
	if time.Since(client.pvCacheTime) < 5*time.Minute && len(client.pvCache) > 0 {
		client.pvCacheMutex.RUnlock()
		return nil
	}
	client.pvCacheMutex.RUnlock()

	// 批量获取所有PV
	pvList, err := client.ClientSet.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// 更新缓存
	client.pvCacheMutex.Lock()
	defer client.pvCacheMutex.Unlock()

	newCache := make(map[string]*corev1.PersistentVolume)
	for i := range pvList.Items {
		pv := &pvList.Items[i]
		newCache[pv.Name] = pv
	}

	client.pvCache = newCache
	client.pvCacheTime = time.Now()

	return nil
}

// getPVFromCache 从指定客户端的缓存获取PV信息
func (m *MultiClusterK8sService) getPVFromCache(client *ClusterClient, pvName string) *corev1.PersistentVolume {
	client.pvCacheMutex.RLock()
	defer client.pvCacheMutex.RUnlock()
	return client.pvCache[pvName]
}

