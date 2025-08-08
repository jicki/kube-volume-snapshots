package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"k8s-volume-snapshots/models"
	"k8s-volume-snapshots/services"
)

var (
	// JWT密钥，从环境变量读取，如果没有则使用随机生成的密钥
	jwtSecret []byte
)

func init() {
	// 从环境变量读取JWT密钥
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// 如果没有设置环境变量，生成一个随机密钥（注意：重启后会变化）
		bytes := make([]byte, 32)
		rand.Read(bytes)
		secret = hex.EncodeToString(bytes)
		println("警告: 使用随机生成的JWT密钥，重启后将失效。建议设置JWT_SECRET环境变量")
	}
	jwtSecret = []byte(secret)
}

// Claims JWT声明结构
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(user *models.User) (string, error) {
	// 设置token过期时间为24小时
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 公开路径，不需要认证
		publicPaths := []string{
			"/api/auth/login",
			"/api/auth/register",
		}

		path := c.Request.URL.Path
		for _, publicPath := range publicPaths {
			if path == publicPath {
				c.Next()
				return
			}
		}

		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(401, "缺少Authorization头"))
			c.Abort()
			return
		}

		// 检查Bearer格式
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(401, "Authorization头格式错误"))
			c.Abort()
			return
		}

		// 解析token
		claims, err := ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(401, "无效的token"))
			c.Abort()
			return
		}

		// 验证用户是否仍然存在
		user, err := userService.GetUser(claims.Username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(401, "用户不存在"))
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", user)
		c.Set("username", user.Username)
		c.Set("role", user.Role)

		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(403, "权限不足"))
			c.Abort()
			return
		}

		userRole := role.(string)

		// admin角色可以访问所有资源
		if userRole == "admin" {
			c.Next()
			return
		}

		// 检查是否有所需角色
		if userRole != requiredRole {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(403, "权限不足"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 需要管理员权限
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// GetCurrentUser 从上下文获取当前用户
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	return user.(*models.User), true
}

// GetCurrentUsername 从上下文获取当前用户名
func GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return username.(string), true
}

// RequireWritePermission 需要写权限（创建、删除、修改操作）
func RequireWritePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(403, "权限不足"))
			c.Abort()
			return
		}

		userRole := role.(string)

		// 只有admin用户可以进行写操作
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(403, "只有管理员可以执行此操作"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// IsAdmin 检查当前用户是否为管理员
func IsAdmin(c *gin.Context) bool {
	role, exists := c.Get("role")
	if !exists {
		return false
	}
	return role.(string) == "admin"
}

// IsReadonly 检查当前用户是否为只读用户
func IsReadonly(c *gin.Context) bool {
	role, exists := c.Get("role")
	if !exists {
		return false
	}
	return role.(string) == "readonly"
}
