package api_test

import (
	"bytes"
	"encoding/json"
	"learn/internal/api"
	"learn/internal/dto"
	"learn/internal/models"
	"learn/internal/services"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setupTestAuthHandler() (*api.AuthHandler, error) {
	db, err := setupTestDB()
	if err != nil {
		return nil, err
	}

	authService := services.NewAuthService(db, "jwt_secret",
		time.Hour, 7*24*time.Hour, 24*time.Hour)
	return &api.AuthHandler{AuthService: authService}, nil
}

func TestCreateUser(t *testing.T) {
	handler, err := setupTestAuthHandler()
	if err != nil {
		t.Fatalf("Failed to setup auth handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/users", handler.CreateUser).Methods("POST")
	router.HandleFunc("/roles", handler.CreateRole).Methods("POST")

	// 先创建一个 "admin" 角色
	roleRequestBody, _ := json.Marshal(dto.RoleRequest{
		Name: "admin",
	})
	roleReq, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(roleRequestBody))
	roleRR := httptest.NewRecorder()
	router.ServeHTTP(roleRR, roleReq)

	if roleRR.Code != http.StatusCreated {
		t.Fatalf("Failed to create role 'admin': %v", roleRR.Code)
	}

	// 然后创建一个用户并分配 "admin" 角色
	requestBody, _ := json.Marshal(dto.CreateUserRequest{
		Username: "testuser",
		Password: "password",
		Roles:    []string{"admin"},
	})
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %v, got %v", http.StatusCreated, rr.Code)
	}

	var response api.Response[dto.CreateUserResponse]
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Data.ID == 0 {
		t.Errorf("Expected non-zero user ID, got %v", response.Data.ID)
	}
}

func TestCreateRole(t *testing.T) {
	handler, err := setupTestAuthHandler()
	if err != nil {
		t.Fatalf("Failed to setup auth handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/roles", handler.CreateRole).Methods("POST")

	requestBody, _ := json.Marshal(dto.RoleRequest{
		Name: "admin",
	})
	req, err := http.NewRequest("POST", "/roles", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %v, got %v", http.StatusCreated, rr.Code)
	}

	var response api.Response[dto.RoleResponse]
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Data.ID == 0 {
		t.Errorf("Expected non-zero role ID, got %v", response.Data.ID)
	}

	if response.Data.Name != "admin" {
		t.Errorf("Expected role name 'admin', got '%v'", response.Data.Name)
	}
}

func TestLogin(t *testing.T) {
	handler, err := setupTestAuthHandler()
	if err != nil {
		t.Fatalf("Failed to setup auth handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/auth/login", handler.Login).Methods("POST")
	router.HandleFunc("/users", handler.CreateUser).Methods("POST")

	// Helper function to attempt login and verify response
	performLogin := func(username, password string, expectedStatusCode int, shouldHaveToken bool) {
		loginRequestBody, _ := json.Marshal(dto.LoginRequest{
			Username: username,
			Password: password,
		})
		loginReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginRequestBody))
		loginRR := httptest.NewRecorder()
		router.ServeHTTP(loginRR, loginReq)

		if loginRR.Code != expectedStatusCode {
			t.Errorf("Expected status code %v, got %v", expectedStatusCode, loginRR.Code)
		}

		var loginResponse api.Response[dto.TokenPairResponse]
		_ = json.NewDecoder(loginRR.Body).Decode(&loginResponse)
		if shouldHaveToken && (loginResponse.Data.AccessToken == "" || loginResponse.Data.RefreshToken == "") {
			t.Errorf("Expected non-empty tokens, got access_token: '%v', refresh_token: '%v'", loginResponse.Data.AccessToken, loginResponse.Data.RefreshToken)
		} else if !shouldHaveToken && (loginResponse.Data.AccessToken != "" || loginResponse.Data.RefreshToken != "") {
			t.Errorf("Expected empty tokens, got access_token: '%v', refresh_token: '%v'", loginResponse.Data.AccessToken, loginResponse.Data.RefreshToken)
		}
	}

	// Create users with different statuses
	userStatuses := []models.UserStatus{
		models.StatusInactive,
		models.StatusActive,
		models.StatusPending,
		models.StatusSuspended,
	}
	for _, status := range userStatuses {
		userRequestBody, _ := json.Marshal(dto.CreateUserRequest{
			Username: "testuser_" + strconv.Itoa(int(status)),
			Password: "password",
			Status:   status,
		})
		userReq, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(userRequestBody))
		userRR := httptest.NewRecorder()
		router.ServeHTTP(userRR, userReq)

		if userRR.Code != http.StatusCreated {
			t.Fatalf("Failed to create user with status %v: %v", status, userRR.Code)
		}
	}

	// Test login for inactive user (should fail)
	performLogin("testuser_0", "password", http.StatusUnauthorized, false)

	// Test login for active user (should succeed)
	performLogin("testuser_1", "password", http.StatusOK, true)

	// Test login for pending user (should fail)
	performLogin("testuser_2", "password", http.StatusUnauthorized, false)

	// Test login for suspended user (should fail)
	performLogin("testuser_3", "password", http.StatusUnauthorized, false)
}

func TestRefreshToken(t *testing.T) {
	handler, err := setupTestAuthHandler()
	if err != nil {
		t.Fatalf("Failed to setup auth handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/auth/login", handler.Login).Methods("POST")
	router.HandleFunc("/auth/refresh", handler.RefreshToken).Methods("POST")
	router.HandleFunc("/roles", handler.CreateRole).Methods("POST")
	router.HandleFunc("/users", handler.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", handler.UpdateUser).Methods("PUT")

	// 创建一个状态为 Active 的用户并登录
	userRequestBody, _ := json.Marshal(dto.CreateUserRequest{
		Username: "activeuser",
		Password: "password",
		Status:   models.StatusActive, // Active
	})
	userReq, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(userRequestBody))
	userRR := httptest.NewRecorder()
	router.ServeHTTP(userRR, userReq)

	if userRR.Code != http.StatusCreated {
		t.Fatalf("Failed to create active user: %v", userRR.Code)
	}

	var createUserResponse api.Response[dto.CreateUserResponse]
	_ = json.NewDecoder(userRR.Body).Decode(&createUserResponse)
	userID := createUserResponse.Data.ID

	var loginResponse api.Response[dto.TokenPairResponse]
	loginRequestBody, _ := json.Marshal(dto.LoginRequest{
		Username: "activeuser",
		Password: "password",
	})
	loginReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginRequestBody))
	loginRR := httptest.NewRecorder()
	router.ServeHTTP(loginRR, loginReq)
	_ = json.NewDecoder(loginRR.Body).Decode(&loginResponse)
	accessToken := loginResponse.Data.AccessToken
	refreshToken := loginResponse.Data.RefreshToken

	// 确认登录成功并返回了有效的 tokens
	if loginRR.Code != http.StatusOK || accessToken == "" || refreshToken == "" {
		t.Fatalf("Login failed for active user with status code %v", loginRR.Code)
	}

	// 尝试刷新 Active 用户的令牌
	refreshRequestBody, _ := json.Marshal(dto.RefreshTokenRequest{
		Token: refreshToken,
	})
	refreshReq, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(refreshRequestBody))
	refreshRR := httptest.NewRecorder()
	router.ServeHTTP(refreshRR, refreshReq)

	if refreshRR.Code != http.StatusOK {
		t.Errorf("Expected status code %v, got %v for active user", http.StatusOK, refreshRR.Code)
	}

	// 验证返回的令牌不为空
	var refreshResponse api.Response[dto.TokenPairResponse]
	_ = json.NewDecoder(refreshRR.Body).Decode(&refreshResponse)
	if refreshResponse.Data.AccessToken == "" || refreshResponse.Data.RefreshToken == "" {
		t.Errorf("Expected non-empty tokens for active user, got access_token: '%v', refresh_token: '%v'", refreshResponse.Data.AccessToken, refreshResponse.Data.RefreshToken)
	}

	// 更新用户状态为 Inactive
	updateRequestBody, _ := json.Marshal(dto.UpdateUserRequest{
		Status: int(models.StatusInactive), // Inactive
	})
	updateReq, _ := http.NewRequest("PUT", "/users/"+strconv.Itoa(int(userID)), bytes.NewBuffer(updateRequestBody))
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusOK {
		t.Fatalf("Failed to update user status to inactive: %v", updateRR.Code)
	}

	// 再次使用旧的 Refresh Token 尝试进行刷新
	refreshReq, _ = http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(refreshRequestBody))
	refreshRR = httptest.NewRecorder()
	router.ServeHTTP(refreshRR, refreshReq)

	// 验证 Inactive 用户不能刷新令牌
	if refreshRR.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %v, got %v for inactive user", http.StatusUnauthorized, refreshRR.Code)
	}
}

func TestGetUsers(t *testing.T) {
	handler, err := setupTestAuthHandler()
	if err != nil {
		t.Fatalf("Failed to setup auth handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/users", handler.GetUsers).Methods("GET")
	router.HandleFunc("/roles", handler.CreateRole).Methods("POST")
	router.HandleFunc("/users", handler.CreateUser).Methods("POST")

	// 先创建一个 "admin" 角色
	roleRequestBody, _ := json.Marshal(dto.RoleRequest{
		Name: "admin",
	})
	roleReq, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(roleRequestBody))
	roleRR := httptest.NewRecorder()
	router.ServeHTTP(roleRR, roleReq)

	if roleRR.Code != http.StatusCreated {
		t.Fatalf("Failed to create role 'admin': %v", roleRR.Code)
	}

	// 然后创建一个用户并分配 "admin" 角色
	userRequestBody, _ := json.Marshal(dto.CreateUserRequest{
		Username: "testuser",
		Password: "password",
		Roles:    []string{"admin"},
	})
	userReq, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(userRequestBody))
	userRR := httptest.NewRecorder()
	router.ServeHTTP(userRR, userReq)

	if userRR.Code != http.StatusCreated {
		t.Fatalf("Failed to create user: %v", userRR.Code)
	}

	// 测试获取用户列表
	req, _ := http.NewRequest("GET", "/users", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, got %v", http.StatusOK, rr.Code)
	}

	var response api.Response[[]dto.UserResponse]
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Data) == 0 {
		t.Errorf("Expected at least one user, got %v", len(response.Data))
	}
}

func TestGetUserRoles(t *testing.T) {
	handler, err := setupTestAuthHandler()
	if err != nil {
		t.Fatalf("Failed to setup auth handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/users/{id}/roles", handler.GetUserRoles).Methods("GET")
	router.HandleFunc("/roles", handler.CreateRole).Methods("POST")
	router.HandleFunc("/users", handler.CreateUser).Methods("POST")

	// 先创建一个 "admin" 角色
	roleRequestBody, _ := json.Marshal(dto.RoleRequest{
		Name: "admin",
	})
	roleReq, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(roleRequestBody))
	roleRR := httptest.NewRecorder()
	router.ServeHTTP(roleRR, roleReq)

	if roleRR.Code != http.StatusCreated {
		t.Fatalf("Failed to create role 'admin': %v", roleRR.Code)
	}

	// 然后创建一个用户并分配 "admin" 角色
	userRequestBody, _ := json.Marshal(dto.CreateUserRequest{
		Username: "testuser",
		Password: "password",
		Roles:    []string{"admin"},
	})
	userReq, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(userRequestBody))
	userRR := httptest.NewRecorder()
	router.ServeHTTP(userRR, userReq)

	if userRR.Code != http.StatusCreated {
		t.Fatalf("Failed to create user: %v", userRR.Code)
	}

	var createUserResponse api.Response[dto.CreateUserResponse]
	_ = json.NewDecoder(userRR.Body).Decode(&createUserResponse)
	userID := createUserResponse.Data.ID

	// 获取用户角色
	req, _ := http.NewRequest("GET", "/users/"+strconv.Itoa(int(userID))+"/roles", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, rr.Code)
	}

	var response api.Response[[]dto.RoleResponse]
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Data) == 0 {
		t.Errorf("Expected at least one role, got %v", len(response.Data))
	}
}

func TestUpdateUser(t *testing.T) {
	handler, err := setupTestAuthHandler()
	if err != nil {
		t.Fatalf("Failed to setup auth handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/users/{id}", handler.UpdateUser).Methods("PUT")
	router.HandleFunc("/roles", handler.CreateRole).Methods("POST")
	router.HandleFunc("/users", handler.CreateUser).Methods("POST")

	// 先创建一个 "admin" 角色
	roleRequestBody, _ := json.Marshal(dto.RoleRequest{
		Name: "admin",
	})
	roleReq, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(roleRequestBody))
	roleRR := httptest.NewRecorder()
	router.ServeHTTP(roleRR, roleReq)

	if roleRR.Code != http.StatusCreated {
		t.Fatalf("Failed to create role 'admin': %v", roleRR.Code)
	}

	// 然后创建一个用户并分配 "admin" 角色
	userRequestBody, _ := json.Marshal(dto.CreateUserRequest{
		Username: "testuser",
		Password: "password",
		Roles:    []string{"admin"},
	})
	userReq, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(userRequestBody))
	userRR := httptest.NewRecorder()
	router.ServeHTTP(userRR, userReq)

	if userRR.Code != http.StatusCreated {
		t.Fatalf("Failed to create user: %v", userRR.Code)
	}

	var createUserResponse api.Response[dto.CreateUserResponse]
	_ = json.NewDecoder(userRR.Body).Decode(&createUserResponse)
	userID := createUserResponse.Data.ID

	// 更新用户信息（不修改密码）
	updateRequestBody, _ := json.Marshal(dto.UpdateUserRequest{
		Username: "updateduser",
		Status:   1,
	})
	updateReq, _ := http.NewRequest("PUT", "/users/"+strconv.Itoa(int(userID)), bytes.NewBuffer(updateRequestBody))
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, updateRR.Code)
	}

	// 验证用户是否更新，且密码未变
	user, _ := handler.AuthService.GetUserByID(userID)
	if user.Username != "updateduser" {
		t.Errorf("Expected username 'updateduser', got '%v'", user.Username)
	}

	err = services.VerifyPassword(user.Password, "password")
	if err != nil {
		t.Errorf("Expected password to remain unchanged, but it was modified")
	}

	// 更新用户信息并修改密码
	newPassword := "newpassword"
	updateRequestBodyWithPassword, _ := json.Marshal(dto.UpdateUserRequest{
		Username: "updateduser",
		Status:   1,
		Password: &newPassword,
	})
	updateReqWithPassword, _ := http.NewRequest("PUT", "/users/"+strconv.Itoa(int(userID)), bytes.NewBuffer(updateRequestBodyWithPassword))
	updateRRWithPassword := httptest.NewRecorder()
	router.ServeHTTP(updateRRWithPassword, updateReqWithPassword)

	if updateRRWithPassword.Code != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, updateRRWithPassword.Code)
	}

	// 验证用户密码是否更新
	user, _ = handler.AuthService.GetUserByID(userID)
	err = services.VerifyPassword(user.Password, newPassword)
	if err != nil {
		t.Errorf("Expected password to be updated, but it was not")
	}
}

func TestDeleteRole(t *testing.T) {
	handler, err := setupTestAuthHandler()
	if err != nil {
		t.Fatalf("Failed to setup auth handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/roles", handler.CreateRole).Methods("POST")
	router.HandleFunc("/roles/{id}", handler.DeleteRole).Methods("DELETE")

	// 先创建一个角色
	roleRequestBody, _ := json.Marshal(dto.RoleRequest{
		Name: "testrole",
	})
	roleReq, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(roleRequestBody))
	roleRR := httptest.NewRecorder()
	router.ServeHTTP(roleRR, roleReq)

	var createRoleResponse api.Response[dto.RoleResponse]
	_ = json.NewDecoder(roleRR.Body).Decode(&createRoleResponse)
	roleID := createRoleResponse.Data.ID

	// 删除角色
	deleteReq, _ := http.NewRequest("DELETE", "/roles/"+strconv.Itoa(int(roleID)), nil)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNoContent {
		t.Errorf("Expected status code %v, got %v", http.StatusNoContent, deleteRR.Code)
	}

	// 验证角色是否已删除
	role, err := handler.AuthService.GetRoleByID(roleID)
	if err == nil || role.Name == "testrole" {
		t.Errorf("Expected role to be deleted, but it still exists")
	}
}

func TestGetPermissions(t *testing.T) {
	handler, err := setupTestAuthHandler()
	if err != nil {
		t.Fatalf("Failed to setup auth handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/permissions", handler.GetPermissions).Methods("GET")
	router.HandleFunc("/roles", handler.CreateRole).Methods("POST")

	// 先创建一些权限
	permissions := []string{"read", "write"}
	for _, perm := range permissions {
		err := handler.AuthService.EnsurePermissionExists(perm, "")
		if err != nil {
			t.Fatalf("Failed to ensure permission exists: %v", err)
		}
	}

	req, _ := http.NewRequest("GET", "/permissions", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, rr.Code)
	}

	var response api.Response[[]dto.PermissionResponse]
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Data) == 0 {
		t.Errorf("Expected at least one permission, got %v", len(response.Data))
	}
}
