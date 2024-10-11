// api/auth.go
package api

import (
	"encoding/json"
	"learn/internal/dto"
	"learn/internal/models"
	"learn/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type AuthHandler struct {
	AuthService *services.AuthService
}

func (h *AuthHandler) GetApiEndpoints() []APIEndpoint {
	return []APIEndpoint{
		//anonomous
		{"/auth/login", http.MethodPost, h.Login, "", "用户登录"},
		{"/auth/register", "POST", h.RegisterUser, "", "用户注册"},
		{"/auth/refresh", http.MethodPost, h.RefreshToken, "", ""},

		//admin:users
		{"/users", "GET", h.GetUsers, "users:read", "获取用户列表"},
		{"/users", "POST", h.CreateUser, "users:edit", "创建用户"},
		{"/users/{id}", "PUT", h.UpdateUser, "users:edit", "编辑用户"},
		{"/users/{id}/roles", "GET", h.GetUserRoles, "users:read", ""},
		{"/users/{id}/invalidate_session", "POST", h.InvalidateUserSession, "users:logout", "踢出登录"},

		//admin:roles
		{"/roles", "GET", h.GetRoles, "roles:read", ""},
		{"/roles", "POST", h.CreateRole, "roles:create", ""},
		{"/roles/{id}", http.MethodDelete, h.DeleteRole, "roles:delete", ""},
		{"/roles/{id}", http.MethodPut, h.UpdateRoleHandler, "roles:edit", ""},
		{"/roles/{id}/permissions", http.MethodGet, h.GetRolePermissionsHandler, "roles:read", ""},

		//admin: permissions
		{"/permissions", "GET", h.GetPermissions, "permissions:read", ""},
	}
}

// CreateUser 处理创建用户的请求
// @Summary 创建用户
// @Description 创建一个新用户
// @Tags User
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param user body dto.CreateUserRequest true "创建用户请求"
// @Success 201 {object} Response[dto.CreateUserResponse] "创建成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /users [post]
func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.AuthService.CreateUser(req.Username, req.Password, req.Roles, req.Status)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, dto.CreateUserResponse{ID: user.ID}, nil, http.StatusCreated)
}

// GetUsers 处理获取用户列表的请求
// @Summary 获取用户列表
// @Description 获取所有用户的列表
// @Tags User
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Success 200 {object} Response[[]dto.UserResponse] "请求成功"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /users [get]
func (h *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.AuthService.GetUsers()
	if err != nil {
		Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	// 转换为响应结构体
	var userResponses []dto.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, dto.UserResponse{
			ID:           user.ID,
			Username:     user.Username,
			CreatedAt:    user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
			TokenVersion: user.TokenVersion,
			Status:       int(user.Status),
		})
	}

	Success(w, userResponses, nil, http.StatusOK)
}

// GetUserRoles 获取指定用户的角色
// @Summary 获取用户角色
// @Description 根据用户ID获取该用户的角色列表
// @Tags User
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param id path int true "用户 ID"
// @Success 200 {object} Response[[]dto.RoleResponse] "获取成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /users/{id}/roles [get]
func (h *AuthHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.AuthService.GetUserByIDWithRoles(uint(userID))
	if err != nil {
		Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 构建 dto.RoleResponse 数组
	var roleResponses []dto.RoleResponse
	for _, role := range user.Roles {
		roleResponses = append(roleResponses, dto.RoleResponse{
			ID:   role.ID,
			Name: role.Name,
		})
	}

	Success(w, roleResponses, nil, http.StatusOK)
}

// UpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 更新指定用户的信息，包括用户名、角色和状态
// @Tags User
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param id path int true "用户 ID"
// @Param user body dto.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} Response[string] "更新成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /users/{id} [put]
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = h.AuthService.UpdateUser(uint(userID), req)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, "Update user success", nil, http.StatusOK)
}

// CreateRole 处理创建角色的请求
// @Summary 创建角色
// @Description 创建一个新角色
// @Tags Role
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param role body dto.RoleRequest true "创建角色请求"
// @Success 201 {object} Response[dto.RoleResponse] "创建成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /roles [post]
func (h *AuthHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req dto.RoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	role, err := h.AuthService.CreateRole(req.Name)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, dto.RoleResponse{ID: role.ID, Name: role.Name}, nil, http.StatusCreated)
}

// UpdateRole 更新角色信息
// @Summary 更新角色信息
// @Description 更新指定角色的信息，包括角色名称和权限
// @Tags Role
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param id path int true "角色 ID"
// @Param role body dto.RoleUpdateRequest true "更新角色请求"
// @Success 200 {object} Response[string] "更新成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /roles/{id} [put]
func (h *AuthHandler) UpdateRoleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	var req dto.RoleUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = h.AuthService.UpdateRole(uint(roleID), req)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, "Update role success", nil, http.StatusOK)
}

// GetRolePermissions 获取角色权限
// @Summary 获取角色权限
// @Description 获取指定角色的权限列表
// @Tags Role
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param id path int true "角色 ID"
// @Success 200 {object} Response[[]dto.PermissionResponse] "获取成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /roles/{id}/permissions [get]
func (h *AuthHandler) GetRolePermissionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	permissions, err := h.AuthService.GetRolePermissions(uint(roleID))
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var permissionResponses []dto.PermissionResponse
	for _, permission := range permissions {
		permissionResponses = append(permissionResponses, dto.PermissionResponse{ID: permission.ID, Name: permission.Name, Description: permission.Description})
	}

	Success(w, permissionResponses, nil, http.StatusOK)
}

// GetRoles 处理获取角色列表的请求
// @Summary 获取角色列表
// @Description 获取所有角色的列表
// @Tags Role
// @Security ApiKeyAuth
// @Produce  json
// @Success 200 {object} Response[[]dto.RoleResponse] "获取成功"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /roles [get]
func (h *AuthHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.AuthService.GetRoles()
	if err != nil {
		Error(w, "Failed to fetch roles", http.StatusInternalServerError)
		return
	}

	var roleResponses []dto.RoleResponse
	for _, role := range roles {
		roleResponses = append(roleResponses, dto.RoleResponse{ID: role.ID, Name: role.Name})
	}

	Success(w, roleResponses, nil, http.StatusOK)
}

// DeleteRole 处理删除角色的请求
// @Summary 删除角色
// @Description 根据角色ID删除指定的角色
// @Tags Role
// @Security ApiKeyAuth
// @Param id path int true "角色 ID"
// @Success 204 "删除成功，无内容返回"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /roles/{id} [delete]
func (h *AuthHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleIDStr := vars["id"]

	// 将 roleID 从 string 转换为 uint
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	err = h.AuthService.DeleteRole(uint(roleID))
	if err != nil {
		Error(w, "Failed to delete role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPermissions 处理获取权限列表的请求
// @Summary 获取权限列表
// @Description 获取所有权限的列表
// @Tags Permission
// @Security ApiKeyAuth
// @Produce  json
// @Success 200 {object} Response[[]dto.PermissionResponse] "获取成功"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /permissions [get]
func (h *AuthHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.AuthService.GetPermissions()
	if err != nil {
		Error(w, "Failed to fetch permissions", http.StatusInternalServerError)
		return
	}

	var permissionResponses []dto.PermissionResponse
	for _, permission := range permissions {
		permissionResponses = append(permissionResponses, dto.PermissionResponse{
			ID: permission.ID, Name: permission.Name, Description: permission.Description})
	}

	Success(w, permissionResponses, nil, http.StatusOK)
}

// Login 处理用户登录请求
// @Summary 用户登录
// @Description 处理用户登录并生成JWT令牌和刷新令牌
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param login body dto.LoginRequest true "登录请求"
// @Success 200 {object} Response[dto.TokenPairResponse] "登录成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 401 {object} ErrorResponse "认证失败"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.AuthService.Authenticate(req.Username, req.Password)
	if err != nil {
		Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 加载用户的角色信息
	if err := h.AuthService.LoadUserRoles(&user); err != nil {
		Error(w, "Failed to load user roles", http.StatusInternalServerError)
		return
	}

	// 生成 Access Token 和 Refresh Token
	accessToken, refreshToken, err := h.AuthService.GenerateTokens(user)
	if err != nil {
		Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	Success(w, dto.TokenPairResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil, http.StatusOK)
}

// RefreshToken 处理JWT令牌刷新请求
// @Summary 刷新JWT令牌
// @Description 刷新JWT令牌，支持滑动过期的双Token机制
// @Tags Auth
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param refreshToken body dto.RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} Response[dto.TokenPairResponse] "刷新成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 401 {object} ErrorResponse "认证失败"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 调用 AuthService 的 RefreshTokens 方法
	newAccessToken, newRefreshToken, err := h.AuthService.RefreshTokens(req.Token)
	if err != nil {
		Error(w, "Failed to refresh token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// 返回新的令牌
	Success(w, dto.TokenPairResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil, http.StatusOK)
}

// InvalidateUserSession invalidates a user's active session by incrementing their token version, rendering all active tokens invalid.
// @Summary Invalidate User Session
// @Description Invalidates the user's current session, effectively logging them out by incrementing their token version.
// @Tags User
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} Response[string] "User session invalidated successfully"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/{id}/invalidate_session [post]
func (h *AuthHandler) InvalidateUserSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.AuthService.InvalidateUserToken(uint(userID))
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, "User session invalidated successfully", nil, http.StatusOK)
}

// RegisterUser 用户注册
// @Summary 用户注册
// @Description 注册新用户
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param input body dto.RegisterUserRequest true "用户注册信息"
// @Success 201 {object} Response[string] "注册成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /auth/register [post]
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err := h.AuthService.CreateUser(req.Username, req.Password, []string{}, models.StatusPending)
	if err != nil {
		Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	Success(w, "User registered successfully", nil, http.StatusCreated)
}
