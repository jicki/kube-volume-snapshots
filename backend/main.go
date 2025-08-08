package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"

	"k8s-volume-snapshots/controllers"
	"k8s-volume-snapshots/middleware"
	"k8s-volume-snapshots/services"

	// 导入静态资源
	_ "k8s-volume-snapshots/static"
)

func main() {
	// 初始化多集群 Kubernetes 服务
	multiK8sService, err := services.NewMultiClusterK8sService()
	if err != nil {
		log.Fatalf("Failed to initialize Multi-Cluster Kubernetes service: %v", err)
	}

	// 初始化用户服务
	userService := services.NewUserService()

	// 初始化 Ceph 服务
	cephService, err := services.NewCephService()
	if err != nil {
		log.Printf("Warning: Failed to initialize Ceph service: %v", err)
		log.Println("Ceph功能将不可用，但应用程序将继续运行")
		// 继续运行，但 Ceph 功能将显示为不可用
	}

	// 初始化控制器
	snapshotController := controllers.NewSnapshotController(multiK8sService)
	scheduledController := controllers.NewScheduledController(multiK8sService)
	userController := controllers.NewUserController(userService)
	cephController := controllers.NewCephController(cephService)
	clusterController := controllers.NewClusterController(multiK8sService)

	// 设置 Gin 路由
	r := gin.Default()

	// 设置可信任的代理（生产环境建议设置为 nil 或具体的代理 IP）
	err = r.SetTrustedProxies(nil)
	if err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	// 配置 CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080", "http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// 设置静态文件服务
	statikFS, err := fs.New()
	if err != nil {
		log.Fatalf("Failed to initialize static file system: %v", err)
	}

	// 服务静态文件
	r.StaticFS("/static", statikFS)

	// 前端路由处理 - 对于非 API 路径，返回 index.html
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// 如果是 API 路径，返回 404
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// 否则返回 index.html（用于前端路由）
		file, err := statikFS.Open("/index.html")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load frontend"})
			return
		}
		defer file.Close()

		c.Header("Content-Type", "text/html")
		http.ServeContent(c.Writer, c.Request, "index.html", time.Time{}, file)
	})

	// 健康检查接口（公开，不需要认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "K8s Volume Snapshots Manager is running",
		})
	})

	// API 路由组
	api := r.Group("/api")
	{
		// 认证相关接口（公开路径）
		auth := api.Group("/auth")
		{
			auth.POST("/login", userController.Login)
			auth.POST("/register", userController.Register)
		}

		// 需要认证的接口
		authenticated := api.Group("")
		authenticated.Use(middleware.AuthMiddleware(userService))
		{
			// 用户相关接口（需要认证）
			user := authenticated.Group("/user")
			{
				user.GET("/profile", userController.GetProfile)
				user.POST("/change-password", userController.ChangePassword)
				// 管理员专用接口
				user.GET("/all", middleware.RequireAdmin(), userController.GetAllUsers)
				user.DELETE("/:username", middleware.RequireAdmin(), userController.DeleteUser)
			}

			// 只读接口（认证用户都可以访问）
			// VolumeSnapshotClass 相关接口
			authenticated.GET("/volumesnapshotclasses", snapshotController.GetVolumeSnapshotClasses)

			// VolumeSnapshot 查询接口
			authenticated.GET("/volumesnapshots", snapshotController.GetVolumeSnapshots)
			authenticated.GET("/volumesnapshotcontents/:name", snapshotController.GetVolumeSnapshotContent)

			// PVC 相关接口
			authenticated.GET("/pvcs", snapshotController.GetPVCs)

			// Namespace 相关接口
			authenticated.GET("/namespaces", snapshotController.GetNamespaces)

			// StorageClass 相关接口
			authenticated.GET("/storageclasses", snapshotController.GetStorageClasses)

			// 定时任务查询接口
			authenticated.GET("/scheduled-snapshots", scheduledController.GetScheduledSnapshots)

			// Ceph 集群信息接口（只读）
			ceph := authenticated.Group("/ceph")
			{
				ceph.GET("/cluster/info", cephController.GetClusterInfo)
				ceph.GET("/cluster/status", cephController.GetClusterStatus)
				ceph.GET("/pools", cephController.GetPoolsInfo)
				ceph.GET("/connection/status", cephController.GetConnectionStatus)
			}

			// K8s 集群管理接口（只读）
			clusters := authenticated.Group("/clusters")
			{
				clusters.GET("", clusterController.GetClusters)
				clusters.GET("/current", clusterController.GetCurrentCluster)
			}

			// 需要管理员权限的写操作接口
			writeOps := authenticated.Group("")
			writeOps.Use(middleware.RequireWritePermission())
			{
				// VolumeSnapshot 写操作
				writeOps.POST("/volumesnapshots", snapshotController.CreateVolumeSnapshot)
				writeOps.DELETE("/volumesnapshots/:namespace/:name", snapshotController.DeleteVolumeSnapshot)
				writeOps.POST("/volumesnapshots/:namespace/:name/force-delete", snapshotController.ForceDeleteVolumeSnapshot)

				// 定时任务写操作
				writeOps.POST("/scheduled-snapshots", scheduledController.CreateScheduledSnapshot)
				writeOps.PUT("/scheduled-snapshots/:id", scheduledController.UpdateScheduledSnapshot)
				writeOps.DELETE("/scheduled-snapshots/:id", scheduledController.DeleteScheduledSnapshot)
				writeOps.POST("/scheduled-snapshots/:id/toggle", scheduledController.ToggleScheduledSnapshot)

				// 集群切换操作（需要写权限）
				writeOps.POST("/clusters/switch", clusterController.SwitchCluster)
			}
		}
	}

	// 启动服务器
	log.Println("Server starting on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
