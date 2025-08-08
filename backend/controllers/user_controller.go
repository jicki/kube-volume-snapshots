package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"k8s-volume-snapshots/middleware"
	"k8s-volume-snapshots/models"
	"k8s-volume-snapshots/services"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// Register 用户注册
func (uc *UserController) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(400, err.Error()))
		return
	}

	user, err := uc.userService.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(400, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(user))
}

// Login 用户登录
func (uc *UserController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(400, err.Error()))
		return
	}

	user, err := uc.userService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(401, err.Error()))
		return
	}

	// 生成JWT token
	token, err := middleware.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(500, "生成token失败"))
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response))
}

// GetProfile 获取当前用户资料
func (uc *UserController) GetProfile(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(401, "用户未认证"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(user))
}

// ChangePassword 修改密码
func (uc *UserController) ChangePassword(c *gin.Context) {
	username, exists := middleware.GetCurrentUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(401, "用户未认证"))
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(400, err.Error()))
		return
	}

	err := uc.userService.ChangePassword(username, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]string{
		"message": "密码修改成功",
	}))
}

// GetAllUsers 获取所有用户列表（仅管理员可用）
func (uc *UserController) GetAllUsers(c *gin.Context) {
	users := uc.userService.GetAllUsers()
	c.JSON(http.StatusOK, models.NewSuccessResponse(users))
}

// DeleteUser 删除用户（仅管理员可用）
func (uc *UserController) DeleteUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(400, "用户名不能为空"))
		return
	}

	err := uc.userService.DeleteUser(username)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]string{
		"message":  "用户删除成功",
		"username": username,
	}))
}
