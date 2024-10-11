package middleware_test

import (
	"learn/internal/models"
	"learn/internal/services"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestEnv 创建内存数据库并返回 AuthService
func SetupTestEnv(t *testing.T) (*services.AuthService, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	authService := services.NewAuthService(db, "jwt_secret", time.Hour, 7*24*time.Hour, 24*time.Hour)
	return authService, db
}

// func TestRequirePermission_EnsurePermissionExists(t *testing.T) {
// 	authService, db := SetupTestEnv(t)

// 	permission := "test_permission"
// 	router := mux.NewRouter()

// 	// 创建并保存权限
// 	perm := models.Permission{Name: permission}
// 	if err := db.Create(&perm).Error; err != nil {
// 		t.Fatalf("Failed to create permission: %v", err)
// 	}

// 	// 创建并保存角色，并将权限分配给角色
// 	role := models.Role{Name: "admin"}
// 	if err := db.Create(&role).Error; err != nil {
// 		t.Fatalf("Failed to create role: %v", err)
// 	}
// 	if err := db.Model(&role).Association("Permissions").Append(&perm); err != nil {
// 		t.Fatalf("Failed to associate permission with role: %v", err)
// 	}

// 	// 创建并保存用户，并将角色分配给用户
// 	user := models.User{Username: "testuser"}
// 	if err := db.Create(&user).Error; err != nil {
// 		t.Fatalf("Failed to create user: %v", err)
// 	}
// 	if err := db.Model(&user).Association("Roles").Append(&role); err != nil {
// 		t.Fatalf("Failed to associate role with user: %v", err)
// 	}

// 	// 注册路由并使用中间件
// 	router.Handle("/test", middleware.RequirePermission(authService, permission, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	})))

// 	req, err := http.NewRequest("GET", "/test", nil)
// 	if err != nil {
// 		t.Fatalf("Failed to create request: %v", err)
// 	}

// 	// 模拟用户登录，将用户添加到请求的上下文中
// 	ctx := context.WithValue(req.Context(), contextkeys.User, &user)
// 	req = req.WithContext(ctx)

// 	rr := httptest.NewRecorder()

// 	router.ServeHTTP(rr, req)

// 	if rr.Code != http.StatusOK {
// 		t.Errorf("Expected status code %v, got %v", http.StatusOK, rr.Code)
// 	}

// 	// 检查权限是否已添加到数据库中
// 	var permissionModel models.Permission
// 	if err := db.Where("name = ?", permission).First(&permissionModel).Error; err != nil {
// 		t.Errorf("Expected permission to be added to the database, but it was not found")
// 	}
// }

// func TestRequirePermission_UserWithPermission(t *testing.T) {
// 	authService, db := SetupTestEnv(t)

// 	// 创建并保存权限
// 	permission := models.Permission{Name: "test_permission"}
// 	if err := db.Create(&permission).Error; err != nil {
// 		t.Fatalf("Failed to create permission: %v", err)
// 	}

// 	// 创建并保存角色，并关联权限
// 	role := models.Role{Name: "admin"}
// 	if err := db.Create(&role).Error; err != nil {
// 		t.Fatalf("Failed to create role: %v", err)
// 	}
// 	if err := db.Model(&role).Association("Permissions").Append(&permission); err != nil {
// 		t.Fatalf("Failed to associate permission with role: %v", err)
// 	}

// 	// 创建并保存用户，并关联角色
// 	user := models.User{Username: "testuser"}
// 	if err := db.Create(&user).Error; err != nil {
// 		t.Fatalf("Failed to create user: %v", err)
// 	}
// 	if err := db.Model(&user).Association("Roles").Append(&role); err != nil {
// 		t.Fatalf("Failed to associate role with user: %v", err)
// 	}

// 	router := mux.NewRouter()
// 	router.Handle("/test", middleware.RequirePermission(authService, "test_permission", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	})))

// 	req, err := http.NewRequest("GET", "/test", nil)
// 	if err != nil {
// 		t.Fatalf("Failed to create request: %v", err)
// 	}

// 	// 模拟用户登录，将用户添加到请求的上下文中
// 	ctx := context.WithValue(req.Context(), contextkeys.User, &user)
// 	req = req.WithContext(ctx)

// 	rr := httptest.NewRecorder()
// 	router.ServeHTTP(rr, req)

// 	if rr.Code != http.StatusOK {
// 		t.Errorf("Expected status code %v, got %v", http.StatusOK, rr.Code)
// 	}
// }

// func TestRequirePermission_UserWithoutPermission(t *testing.T) {
// 	authService, db := SetupTestEnv(t)

// 	// 创建权限，但不将其分配给角色
// 	permission := models.Permission{Name: "test_permission"}
// 	if err := db.Create(&permission).Error; err != nil {
// 		t.Fatalf("Failed to create permission: %v", err)
// 	}

// 	// 创建并保存角色，但不关联权限
// 	role := models.Role{Name: "user"}
// 	if err := db.Create(&role).Error; err != nil {
// 		t.Fatalf("Failed to create role: %v", err)
// 	}

// 	// 创建并保存用户，并关联角色
// 	user := models.User{Username: "testuser"}
// 	if err := db.Create(&user).Error; err != nil {
// 		t.Fatalf("Failed to create user: %v", err)
// 	}
// 	if err := db.Model(&user).Association("Roles").Append(&role); err != nil {
// 		t.Fatalf("Failed to associate role with user: %v", err)
// 	}

// 	router := mux.NewRouter()
// 	router.Handle("/test", middleware.RequirePermission(authService, "test_permission", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	})))

// 	req, err := http.NewRequest("GET", "/test", nil)
// 	if err != nil {
// 		t.Fatalf("Failed to create request: %v", err)
// 	}

// 	// 模拟用户登录，将用户添加到请求的上下文中
// 	ctx := context.WithValue(req.Context(), contextkeys.User, &user)
// 	req = req.WithContext(ctx)

// 	rr := httptest.NewRecorder()
// 	router.ServeHTTP(rr, req)

// 	if rr.Code != http.StatusForbidden {
// 		t.Errorf("Expected status code %v, got %v", http.StatusForbidden, rr.Code)
// 	}
// }

// func TestRequirePermission_Unauthorized(t *testing.T) {
// 	authService, _ := SetupTestEnv(t)

// 	router := mux.NewRouter()
// 	router.Handle("/test", middleware.RequirePermission(authService, "test_permission", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	})))

// 	req, err := http.NewRequest("GET", "/test", nil)
// 	if err != nil {
// 		t.Fatalf("Failed to create request: %v", err)
// 	}

// 	rr := httptest.NewRecorder()
// 	router.ServeHTTP(rr, req)

// 	if rr.Code != http.StatusUnauthorized {
// 		t.Errorf("Expected status code %v, got %v", http.StatusUnauthorized, rr.Code)
// 	}
// }
