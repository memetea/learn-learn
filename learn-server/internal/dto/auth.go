package dto

import "learn/internal/models"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateUserRequest 定义了创建用户请求的结构体
type CreateUserRequest struct {
	Username string            `json:"username"`
	Password string            `json:"password"`
	Roles    []string          `json:"roles"`  // 用户可以有多个角色
	Status   models.UserStatus `json:"status"` // 用户状态
}

// CreateUserResponse 定义了创建用户响应的结构体
type CreateUserResponse struct {
	ID uint `json:"id"`
}

// UserResponse 定义返回的用户信息结构体
type UserResponse struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	TokenVersion uint   `json:"token_version"`
	Status       int    `json:"status"`
}

type UpdateUserRequest struct {
	Username string   `json:"username"`
	Password *string  `json:"password,omitempty"` // 可选的密码字段
	Roles    []string `json:"roles"`
	Status   int      `json:"status"`
}

// RoleRequest 定义了创建角色请求的结构体
type RoleRequest struct {
	Name string `json:"name"`
}

// RoleResponse 定义了创建角色响应的结构体
type RoleResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// AssignRoleRequest 定义了为用户分配角色的请求
type AssignRoleRequest struct {
	RoleName string `json:"role_name"`
}

type RolePermissionRequest struct {
	RoleName       string `json:"role_name"`
	PermissionName string `json:"permission_name"`
}

// PermissionResponse 定义了权限的响应结构体
type PermissionResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TokenPairResponse 返回 Access Token 和 Refresh Token
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenRequest 定义了刷新JWT令牌请求的结构体
type RefreshTokenRequest struct {
	Token string `json:"token"`
}

// RoleUpdateRequest represents the request payload for updating a role// RoleUpdateRequest represents the request payload for updating a role
type RoleUpdateRequest struct {
	Name        string `json:"name"`
	Permissions []int  `json:"permissions"`
}

type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type MenuItem struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Parent string `json:"parent"`
	Label  string `json:"label"`
	Icon   string `json:"icon,omitempty"`
	Order  int    `json:"order"`
}
