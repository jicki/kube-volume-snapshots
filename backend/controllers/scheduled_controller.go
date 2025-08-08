package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	"github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s-volume-snapshots/middleware"
	"k8s-volume-snapshots/models"
	"k8s-volume-snapshots/services"
)

const (
	// 定时任务数据存储文件路径
	TaskDataFile = "/data/scheduled_tasks.json"
)

type ScheduledController struct {
	k8sService     services.K8sServiceInterface
	cron           *cron.Cron
	scheduledTasks map[string]*models.ScheduledSnapshot
	cronEntries    map[string]cron.EntryID
	mutex          sync.RWMutex
	dataFile       string
}

func NewScheduledController(k8sService services.K8sServiceInterface) *ScheduledController {
	c := cron.New(cron.WithSeconds())
	c.Start()

	controller := &ScheduledController{
		k8sService:     k8sService,
		cron:           c,
		scheduledTasks: make(map[string]*models.ScheduledSnapshot),
		cronEntries:    make(map[string]cron.EntryID),
		dataFile:       TaskDataFile,
	}

	// 加载持久化的任务数据
	controller.loadTasks()

	// 重新启动已启用的定时任务
	controller.restartEnabledTasks()

	return controller
}

// loadTasks 从文件加载任务数据
func (c *ScheduledController) loadTasks() {
	if _, err := os.Stat(c.dataFile); os.IsNotExist(err) {
		fmt.Printf("定时任务数据文件不存在，使用空的任务列表\n")
		return
	}

	data, err := ioutil.ReadFile(c.dataFile)
	if err != nil {
		fmt.Printf("读取定时任务数据文件失败: %v\n", err)
		return
	}

	var tasks []models.ScheduledSnapshot
	if err := json.Unmarshal(data, &tasks); err != nil {
		fmt.Printf("解析定时任务数据失败: %v\n", err)
		return
	}

	// 将任务加载到内存中
	for _, task := range tasks {
		taskCopy := task // 避免闭包引用问题
		c.scheduledTasks[task.ID] = &taskCopy
	}

	fmt.Printf("成功加载 %d 个定时任务\n", len(tasks))
}

// saveTasks 保存任务数据到文件
func (c *ScheduledController) saveTasks() error {
	// 确保目录存在
	dir := filepath.Dir(c.dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %v", err)
	}

	// 将map转换为slice进行序列化
	var tasks []models.ScheduledSnapshot
	for _, task := range c.scheduledTasks {
		tasks = append(tasks, *task)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化任务数据失败: %v", err)
	}

	if err := ioutil.WriteFile(c.dataFile, data, 0644); err != nil {
		return fmt.Errorf("写入任务数据文件失败: %v", err)
	}

	return nil
}

// restartEnabledTasks 重新启动已启用的定时任务
func (c *ScheduledController) restartEnabledTasks() {
	for _, task := range c.scheduledTasks {
		if task.Enabled {
			taskPtr := task // 避免闭包引用问题
			entryID, err := c.cron.AddFunc(task.CronExpression, func() {
				c.executeSnapshot(taskPtr)
			})
			if err != nil {
				fmt.Printf("重新启动定时任务失败 %s: %v\n", task.Name, err)
			} else {
				c.cronEntries[task.ID] = entryID
				fmt.Printf("重新启动定时任务: %s\n", task.Name)
			}
		}
	}
}

// GetScheduledSnapshots 获取所有定时任务
func (c *ScheduledController) GetScheduledSnapshots(ctx *gin.Context) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// 初始化为空切片而不是nil，确保JSON序列化为[]而不是null
	tasks := []models.ScheduledSnapshot{}
	for _, task := range c.scheduledTasks {
		// 更新下次执行时间
		if entryID, exists := c.cronEntries[task.ID]; exists && task.Enabled {
			if entry := c.cron.Entry(entryID); entry.Valid() {
				nextTime := entry.Next
				task.NextExecution = &nextTime
			}
		}
		tasks = append(tasks, *task)
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(tasks))
}

// CreateScheduledSnapshot 创建定时任务
func (c *ScheduledController) CreateScheduledSnapshot(ctx *gin.Context) {
	var req models.ScheduledSnapshot
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

	// 验证 cron 表达式（6字段，包含秒）
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(req.CronExpression)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.NewErrorResponse(400, fmt.Sprintf("Invalid cron expression (6 fields required): %v", err)))
		return
	}

	// 生成唯一 ID
	req.ID = fmt.Sprintf("%s-%s-%d", req.Namespace, req.Name, time.Now().Unix())
	req.CreatedBy = username // 设置创建者
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	req.Enabled = true

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 保存任务到内存
	c.scheduledTasks[req.ID] = &req

	// 添加定时任务
	if req.Enabled {
		taskPtr := &req // 避免闭包引用问题
		entryID, err := c.cron.AddFunc(req.CronExpression, func() {
			c.executeSnapshot(taskPtr)
		})
		if err != nil {
			// 如果添加定时任务失败，从内存中移除
			delete(c.scheduledTasks, req.ID)
			ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
			return
		}
		c.cronEntries[req.ID] = entryID
	}

	// 保存数据到文件
	if err := c.saveTasks(); err != nil {
		fmt.Printf("保存定时任务数据失败: %v\n", err)
		// 不影响API响应，只记录日志
	}

	ctx.JSON(http.StatusCreated, models.NewSuccessResponse(req))
}

// UpdateScheduledSnapshot 更新定时任务
func (c *ScheduledController) UpdateScheduledSnapshot(ctx *gin.Context) {
	id := ctx.Param("id")
	var req models.ScheduledSnapshot
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.NewErrorResponse(400, err.Error()))
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	existingTask, exists := c.scheduledTasks[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, models.NewErrorResponse(404, "Scheduled task not found"))
		return
	}

	// 验证 cron 表达式（6字段，包含秒）
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(req.CronExpression)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.NewErrorResponse(400, fmt.Sprintf("Invalid cron expression (6 fields required): %v", err)))
		return
	}

	// 移除旧的定时任务
	if entryID, exists := c.cronEntries[id]; exists {
		c.cron.Remove(entryID)
		delete(c.cronEntries, id)
	}

	// 更新任务信息
	req.ID = id
	req.CreatedAt = existingTask.CreatedAt
	req.UpdatedAt = time.Now()

	// 更新内存中的任务
	c.scheduledTasks[id] = &req

	// 添加新的定时任务
	if req.Enabled {
		taskPtr := &req // 避免闭包引用问题
		entryID, err := c.cron.AddFunc(req.CronExpression, func() {
			c.executeSnapshot(taskPtr)
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
			return
		}
		c.cronEntries[req.ID] = entryID
	}

	// 保存数据到文件
	if err := c.saveTasks(); err != nil {
		fmt.Printf("保存定时任务数据失败: %v\n", err)
		// 不影响API响应，只记录日志
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(req))
}

// DeleteScheduledSnapshot 删除定时任务
func (c *ScheduledController) DeleteScheduledSnapshot(ctx *gin.Context) {
	id := ctx.Param("id")

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查任务是否存在
	if _, exists := c.scheduledTasks[id]; !exists {
		ctx.JSON(http.StatusNotFound, models.NewErrorResponse(404, "Scheduled task not found"))
		return
	}

	// 移除定时任务
	if entryID, exists := c.cronEntries[id]; exists {
		c.cron.Remove(entryID)
		delete(c.cronEntries, id)
	}

	// 删除任务记录
	delete(c.scheduledTasks, id)

	// 保存数据到文件
	if err := c.saveTasks(); err != nil {
		fmt.Printf("保存定时任务数据失败: %v\n", err)
		// 不影响API响应，只记录日志
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

// ToggleScheduledSnapshot 启用/禁用定时任务
func (c *ScheduledController) ToggleScheduledSnapshot(ctx *gin.Context) {
	id := ctx.Param("id")

	c.mutex.Lock()
	defer c.mutex.Unlock()

	task, exists := c.scheduledTasks[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, models.NewErrorResponse(404, "Scheduled task not found"))
		return
	}

	// 切换状态
	task.Enabled = !task.Enabled
	task.UpdatedAt = time.Now()

	// 移除旧的定时任务
	if entryID, exists := c.cronEntries[id]; exists {
		c.cron.Remove(entryID)
		delete(c.cronEntries, id)
	}

	// 如果启用，添加新的定时任务
	if task.Enabled {
		taskPtr := task // 避免闭包引用问题
		entryID, err := c.cron.AddFunc(task.CronExpression, func() {
			c.executeSnapshot(taskPtr)
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, err.Error()))
			return
		}
		c.cronEntries[id] = entryID
	}

	// 保存数据到文件
	if err := c.saveTasks(); err != nil {
		fmt.Printf("保存定时任务数据失败: %v\n", err)
		// 不影响API响应，只记录日志
	}

	ctx.JSON(http.StatusOK, models.NewSuccessResponse(task))
}

// executeSnapshot 执行快照创建
func (c *ScheduledController) executeSnapshot(task *models.ScheduledSnapshot) {
	now := time.Now()

	c.mutex.Lock()
	task.LastExecuted = &now
	c.mutex.Unlock()

	// 生成快照名称（添加时间戳）
	snapshotName := fmt.Sprintf("%s-%d", task.Name, now.Unix())

	// 确定目标集群列表
	targetClusters := task.TargetClusters
	if len(targetClusters) == 0 {
		// 如果没有指定目标集群，只在当前集群执行
		c.executeSnapshotInCurrentCluster(task, snapshotName, now)
		return
	}

	// 检查是否支持多集群操作
	multiClusterService, ok := c.k8sService.(services.MultiClusterK8sServiceInterface)
	if !ok {
		// 如果不是多集群服务，只在当前集群执行
		fmt.Printf("Multi-cluster operation not supported, executing in current cluster only\n")
		c.executeSnapshotInCurrentCluster(task, snapshotName, now)
		return
	}

	// 并发在多个集群中创建快照
	var wg sync.WaitGroup
	errorChan := make(chan error, len(targetClusters))

	for _, clusterName := range targetClusters {
		wg.Add(1)
		go func(cluster string) {
			defer wg.Done()
			err := c.executeSnapshotInCluster(multiClusterService, task, snapshotName, cluster, now)
			if err != nil {
				errorChan <- fmt.Errorf("cluster %s: %v", cluster, err)
			}
		}(clusterName)
	}

	wg.Wait()
	close(errorChan)

	// 收集错误信息
	var errors []string
	for err := range errorChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		fmt.Printf("Failed to create scheduled snapshot %s in some clusters: %v\n", snapshotName, errors)
	} else {
		fmt.Printf("Successfully created scheduled snapshot %s in all target clusters: %v\n", snapshotName, targetClusters)
	}
}

// executeSnapshotInCurrentCluster 在当前集群中执行快照创建
func (c *ScheduledController) executeSnapshotInCurrentCluster(task *models.ScheduledSnapshot, snapshotName string, now time.Time) {
	// 验证PVC是否存在
	pvcs, err := c.k8sService.GetPVCs(context.Background(), task.Namespace)
	if err != nil {
		fmt.Printf("Failed to get PVCs for scheduled snapshot %s: %v\n", snapshotName, err)
		return
	}

	pvcExists := false
	for _, pvc := range pvcs {
		if pvc.Name == task.PVCName {
			pvcExists = true
			break
		}
	}

	if !pvcExists {
		fmt.Printf("Failed to create scheduled snapshot %s: PVC '%s' not found in namespace '%s'\n", snapshotName, task.PVCName, task.Namespace)
		return
	}

	// 创建 VolumeSnapshot
	vs := c.createVolumeSnapshotSpec(task, snapshotName, now)

	_, err = c.k8sService.CreateVolumeSnapshot(context.Background(), task.Namespace, vs)
	if err != nil {
		fmt.Printf("Failed to create scheduled snapshot %s: %v\n", snapshotName, err)
	} else {
		fmt.Printf("Successfully created scheduled snapshot: %s (created by: %s)\n", snapshotName, task.CreatedBy)
	}
}

// executeSnapshotInCluster 在指定集群中执行快照创建
func (c *ScheduledController) executeSnapshotInCluster(multiClusterService services.MultiClusterK8sServiceInterface, task *models.ScheduledSnapshot, snapshotName, clusterName string, now time.Time) error {
	// 验证指定集群中的PVC是否存在
	pvcs, err := multiClusterService.GetPVCsInCluster(context.Background(), clusterName, task.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get PVCs: %v", err)
	}

	pvcExists := false
	for _, pvc := range pvcs {
		if pvc.Name == task.PVCName {
			pvcExists = true
			break
		}
	}

	if !pvcExists {
		return fmt.Errorf("PVC '%s' not found in namespace '%s'", task.PVCName, task.Namespace)
	}

	// 创建 VolumeSnapshot
	vs := c.createVolumeSnapshotSpec(task, snapshotName, now)

	// 在指定集群中创建快照
	_, err = multiClusterService.CreateVolumeSnapshotInCluster(context.Background(), clusterName, task.Namespace, vs)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %v", err)
	}

	fmt.Printf("Successfully created scheduled snapshot %s in cluster %s (created by: %s)\n", snapshotName, clusterName, task.CreatedBy)
	return nil
}

// createVolumeSnapshotSpec 创建VolumeSnapshot规格
func (c *ScheduledController) createVolumeSnapshotSpec(task *models.ScheduledSnapshot, snapshotName string, now time.Time) *snapshotv1.VolumeSnapshot {
	return &snapshotv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      snapshotName,
			Namespace: task.Namespace,
			Labels: map[string]string{
				"scheduled-task-id":   task.ID,
				"scheduled-task-name": task.Name,
				"created-by":          task.CreatedBy,
				"app":                 "k8s-volume-snapshots",
			},
			Annotations: map[string]string{
				"k8s-volume-snapshots/created-by":     task.CreatedBy,
				"k8s-volume-snapshots/scheduled-task": task.ID,
				"k8s-volume-snapshots/created-at":     now.Format(time.RFC3339),
			},
		},
		Spec: snapshotv1.VolumeSnapshotSpec{
			Source: snapshotv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &task.PVCName,
			},
			VolumeSnapshotClassName: &task.VolumeSnapshotClassName,
		},
	}
}
