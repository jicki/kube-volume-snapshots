package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"k8s-volume-snapshots/models"
)

const (
	// 用户数据存储文件路径
	UserDataFile = "/data/users.json"
	// bcrypt 成本参数
	BcryptCost = 12
)

type UserService struct {
	users    map[string]*models.User
	mutex    sync.RWMutex
	dataFile string
}

func NewUserService() *UserService {
	service := &UserService{
		users:    make(map[string]*models.User),
		dataFile: UserDataFile,
	}

	// 加载用户数据
	service.loadUsers()

	// 如果没有用户，创建默认管理员账户
	if len(service.users) == 0 {
		service.createDefaultAdmin()
	}

	return service
}

// loadUsers 从文件加载用户数据
func (s *UserService) loadUsers() {
	if _, err := os.Stat(s.dataFile); os.IsNotExist(err) {
		fmt.Printf("用户数据文件不存在，使用空的用户列表\n")
		return
	}

	data, err := ioutil.ReadFile(s.dataFile)
	if err != nil {
		fmt.Printf("读取用户数据文件失败: %v\n", err)
		return
	}

	var users []models.User
	if err := json.Unmarshal(data, &users); err != nil {
		fmt.Printf("解析用户数据失败: %v\n", err)
		return
	}

	// 将用户加载到内存中
	for _, user := range users {
		userCopy := user // 避免闭包引用问题
		s.users[user.Username] = &userCopy
	}

	fmt.Printf("成功加载 %d 个用户\n", len(users))
}

// saveUsers 保存用户数据到文件
func (s *UserService) saveUsers() error {
	// 确保目录存在
	dir := filepath.Dir(s.dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %v", err)
	}

	// 将map转换为slice进行序列化
	var users []models.User
	for _, user := range s.users {
		// 不保存密码到文件中，但在内存中保留加密后的密码
		userCopy := *user
		users = append(users, userCopy)
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化用户数据失败: %v", err)
	}

	if err := ioutil.WriteFile(s.dataFile, data, 0600); err != nil {
		return fmt.Errorf("写入用户数据文件失败: %v", err)
	}

	return nil
}

// createDefaultAdmin 创建默认管理员账户
func (s *UserService) createDefaultAdmin() {
	defaultAdmin := &models.User{
		ID:        s.generateID(),
		Username:  "admin",
		Role:      "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置默认密码为 "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), BcryptCost)
	if err != nil {
		fmt.Printf("创建默认管理员账户失败: %v\n", err)
		return
	}
	defaultAdmin.Password = string(hashedPassword)

	s.users[defaultAdmin.Username] = defaultAdmin

	// 保存到文件
	if err := s.saveUsers(); err != nil {
		fmt.Printf("保存默认管理员账户失败: %v\n", err)
		return
	}

	fmt.Printf("创建默认管理员账户: admin / admin123\n")
}

// generateID 生成随机ID
func (s *UserService) generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Register 用户注册
func (s *UserService) Register(req models.RegisterRequest) (*models.User, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查用户名是否已存在
	if _, exists := s.users[req.Username]; exists {
		return nil, errors.New("用户名已存在")
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 创建新用户
	user := &models.User{
		ID:        s.generateID(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		Role:      req.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存到内存
	s.users[user.Username] = user

	// 保存到文件
	if err := s.saveUsers(); err != nil {
		// 如果保存失败，从内存中移除
		delete(s.users, user.Username)
		return nil, fmt.Errorf("保存用户数据失败: %v", err)
	}

	// 返回用户信息时不包含密码
	userResult := *user
	userResult.Password = ""
	return &userResult, nil
}

// Login 用户登录验证
func (s *UserService) Login(req models.LoginRequest) (*models.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 查找用户
	user, exists := s.users[req.Username]
	if !exists {
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 返回用户信息时不包含密码
	userResult := *user
	userResult.Password = ""
	return &userResult, nil
}

// GetUser 根据用户名获取用户信息
func (s *UserService) GetUser(username string) (*models.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, errors.New("用户不存在")
	}

	// 返回用户信息时不包含密码
	userResult := *user
	userResult.Password = ""
	return &userResult, nil
}

// GetAllUsers 获取所有用户列表（仅管理员可用）
func (s *UserService) GetAllUsers() []models.User {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var users []models.User
	for _, user := range s.users {
		userCopy := *user
		userCopy.Password = "" // 不返回密码
		users = append(users, userCopy)
	}

	return users
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(username string, req models.ChangePasswordRequest) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	user, exists := s.users[username]
	if !exists {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), BcryptCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	// 更新密码
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	// 保存到文件
	if err := s.saveUsers(); err != nil {
		return fmt.Errorf("保存用户数据失败: %v", err)
	}

	return nil
}

// DeleteUser 删除用户（仅管理员可用）
func (s *UserService) DeleteUser(username string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 不能删除管理员账户
	if username == "admin" {
		return errors.New("不能删除默认管理员账户")
	}

	if _, exists := s.users[username]; !exists {
		return errors.New("用户不存在")
	}

	// 从内存中删除
	delete(s.users, username)

	// 保存到文件
	if err := s.saveUsers(); err != nil {
		return fmt.Errorf("保存用户数据失败: %v", err)
	}

	return nil
}

// ValidateUser 验证用户是否存在且有效
func (s *UserService) ValidateUser(username string) (*models.User, error) {
	return s.GetUser(username)
}
