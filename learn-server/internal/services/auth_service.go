package services

import (
	"errors"
	"learn/internal/consts/claimkeys"
	"learn/internal/dto"
	"learn/internal/models"
	"log"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// VerifyPassword compares a bcrypt hashed password with a plain text password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

type AuthService struct {
	db                        *gorm.DB
	casbinEnforcer            *casbin.Enforcer
	jwtSecret                 string
	accessTokenDuration       time.Duration
	refreshTokenDuration      time.Duration
	refreshTokenSlidingWindow time.Duration
}

func NewAuthService(db *gorm.DB, jwtSecret string,
	accessTokenDuration, refreshTokenDuration, refreshTokenSlidingWindow time.Duration) *AuthService {
	// 加载 Casbin 模型
	m, err := model.NewModelFromString(`
		[request_definition]
		r = sub, obj, act

		[policy_definition]
		p = sub, obj, act

		[role_definition]
		g = _, _

		[policy_effect]
		e = some(where (p.eft == allow))

		[matchers]
		m = (r.obj == p.obj || p.obj == "*") && (r.act == p.act || p.act == "*") && g(r.sub, p.sub)
	`)
	if err != nil {
		log.Fatalf("failed to load model: %v", err)
	}

	// 初始化 Casbin enforcer
	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		log.Fatalf("failed to create enforcer: %v", err)
	}

	s := &AuthService{
		db:                        db,
		casbinEnforcer:            enforcer,
		jwtSecret:                 jwtSecret,
		accessTokenDuration:       accessTokenDuration,
		refreshTokenDuration:      refreshTokenDuration,
		refreshTokenSlidingWindow: refreshTokenSlidingWindow,
	}

	s.loadCasbinEnforcer()

	return s
}

func (s *AuthService) loadCasbinEnforcer() error {
	s.casbinEnforcer.ClearPolicy()
	s.casbinEnforcer.AddPolicy("admin", "*", "")
	roles, err := s.GetRoles()
	if err != nil {
		return err
	}
	for _, role := range roles {
		roleName := strings.ToLower(role.Name)
		permissions, err := s.GetRolePermissions(role.ID)
		if err != nil {
			return err
		}
		for _, permission := range permissions {
			s.casbinEnforcer.AddPolicy(roleName, permission.Name, "")
		}
	}
	return nil
}

func (s *AuthService) CasbinEnforcer() *casbin.Enforcer {
	return s.casbinEnforcer
}

func (s *AuthService) JwtSecret() string {
	return s.jwtSecret
}

func (s *AuthService) CreateUser(username, password string, roles []string, status models.UserStatus) (models.User, error) {
	var user models.User

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	// 创建用户
	user = models.User{
		Username:     username,
		Password:     string(hashedPassword),
		TokenVersion: 1, // 初始化 TokenVersion
		Status:       status,
	}

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// 关联角色
		for _, roleName := range roles {
			var role models.Role
			if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
				return err
			}
			if err := tx.Model(&user).Association("Roles").Append(&role); err != nil {
				return err
			}
		}

		return nil
	})

	// 返回结果
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (s *AuthService) GetUsers() ([]models.User, error) {
	var users []models.User

	// 只选择需要的字段进行查询
	err := s.db.Select("id, username, created_at, updated_at, status").Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *AuthService) UpdateUser(userID uint, req dto.UpdateUserRequest) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	// 更新用户名和状态
	user.Username = req.Username
	user.Status = models.UserStatus(req.Status)

	// 如果密码不为空，则更新密码
	if req.Password != nil && *req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 更新角色
		if err := tx.Model(&user).Association("Roles").Clear(); err != nil {
			return err
		}
		for _, roleName := range req.Roles {
			var role models.Role
			if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
				return err
			}
			if err := tx.Model(&user).Association("Roles").Append(&role); err != nil {
				return err
			}
		}
		return tx.Save(&user).Error
	})
	return err
}

func (s *AuthService) ActivateUser(userID uint) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("status", models.StatusActive).Error
}

func (s *AuthService) DeactivateUser(userID uint) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("status", models.StatusInactive).Error
}

func (s *AuthService) InitializeAdminUser(username, password string) error {
	var count int64

	// 检查用户表中是否存在任何用户
	err := s.db.Model(&models.User{}).Count(&count).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if count > 0 {
		return nil // 系统中已经有用户存在
	}

	// 如果没有用户，则创建一个带有 admin 角色的用户
	_, err = s.CreateUser(username, password, []string{"admin"}, models.StatusActive)
	return err
}

func (s *AuthService) GetUserByID(userID uint) (models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *AuthService) GetUserByIDWithRoles(userID uint) (models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles").First(&user, userID).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// 检查用户是否具有指定权限
// func (s *AuthService) HasPermission(user *models.User, requiredPermission string) bool {
// 	for _, role := range user.Roles {
// 		var permissions []models.Permission
// 		s.db.Model(&role).Association("Permissions").Find(&permissions)

// 		for _, permission := range permissions {
// 			if permission.Name == requiredPermission {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// EnsurePermissionExists 检查权限是否存在，如果不存在则添加
func (s *AuthService) EnsurePermissionExists(permissionName, description string) error {
	var permission models.Permission
	if err := s.db.Where("name = ?", permissionName).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果权限不存在，则创建它
			permission = models.Permission{Name: permissionName, Description: description}
			if err := s.db.Create(&permission).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if permission.Description != description {
		permission.Description = description
		if err := s.db.Save(&permission).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *AuthService) CreateRole(roleName string) (models.Role, error) {
	var role models.Role

	// 首先检查角色是否已存在
	if err := s.db.Where("name = ?", roleName).First(&role).Error; err == nil {
		// 如果找到已有的角色，直接返回
		return role, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果查询过程中发生其他错误，返回错误
		return models.Role{}, err
	}

	// 如果角色不存在，创建新角色
	role = models.Role{Name: roleName}
	if err := s.db.Create(&role).Error; err != nil {
		return models.Role{}, err
	}

	return role, nil
}

func (s *AuthService) GetRoleByID(roleID uint) (models.Role, error) {
	var role models.Role
	if err := s.db.Preload("Permissions").First(&role, roleID).Error; err != nil {
		return models.Role{}, err
	}
	return role, nil
}

func (s *AuthService) AssignPermissionToRole(roleName, permissionName string) error {
	var role models.Role
	var permission models.Permission

	// 查找角色
	if err := s.db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return err
	}

	// 查找权限
	if err := s.db.Where("name = ?", permissionName).First(&permission).Error; err != nil {
		return err
	}

	// 将权限分配给角色
	return s.db.Model(&role).Association("Permissions").Append(&permission)
}

func (s *AuthService) GetRoles() ([]models.Role, error) {
	var roles []models.Role
	if err := s.db.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *AuthService) DeleteRole(roleID uint) error {
	return s.db.Delete(&models.Role{}, roleID).Error
}

// UpdateRole updates the role with the given ID, including its name and permissions
func (s *AuthService) UpdateRole(roleID uint, req dto.RoleUpdateRequest) error {
	var role models.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return err
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		role.Name = req.Name
		if err := tx.Save(&role).Error; err != nil {
			return err
		}
		// Update role's permissions
		var permissions []models.Permission
		if err := tx.Where("id IN ?", req.Permissions).Find(&permissions).Error; err != nil {
			return err
		}
		if err := tx.Model(&role).Association("Permissions").Replace(&permissions); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return s.loadCasbinEnforcer()
}

// GetRolePermissions retrieves the permissions associated with a specific role
func (s *AuthService) GetRolePermissions(roleID uint) ([]models.Permission, error) {
	var role models.Role
	if err := s.db.Preload("Permissions").First(&role, roleID).Error; err != nil {
		return nil, err
	}

	return role.Permissions, nil
}

func (s *AuthService) GetPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	if err := s.db.Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (s *AuthService) InvalidateUserTokens(userID uint) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("token_version", gorm.Expr("token_version + 1")).Error
}

func (s *AuthService) Authenticate(username, password string) (models.User, error) {
	var user models.User

	// 查询用户信息
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return models.User{}, errors.New("user not found")
	}

	// 检查用户状态是否为 Active
	if user.Status != models.StatusActive {
		return models.User{}, errors.New("user is not active")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return models.User{}, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *AuthService) LoadUserRoles(user *models.User) error {
	if err := s.db.Model(user).Association("Roles").Find(&user.Roles); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) GenerateAccessToken(user models.User) (string, error) {
	if user.Roles == nil {
		s.LoadUserRoles(&user)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		claimkeys.UserId:       user.ID,
		claimkeys.UserName:     user.Username,
		claimkeys.Role:         user.GetRoles(),
		claimkeys.TokenVersion: user.TokenVersion,
		claimkeys.Exp:          time.Now().Add(s.accessTokenDuration).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) GenerateTokens(user models.User) (string, string, error) {
	// 生成 Access Token
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// 生成 Refresh Token，使用较长的有效期
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		claimkeys.UserId:       user.ID,
		claimkeys.TokenVersion: user.TokenVersion,
		claimkeys.Exp:          time.Now().Add(s.refreshTokenDuration).Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenString, nil
}

func (s *AuthService) RefreshTokens(refreshTokenString string) (string, string, error) {
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint(claims[claimkeys.UserId].(float64))
		tokenVersion := uint(claims[claimkeys.TokenVersion].(float64))
		expirationTime := int64(claims[claimkeys.Exp].(float64))
		currentTime := time.Now().Unix()

		// 从数据库获取用户
		var user models.User
		if err := s.db.First(&user, userID).Error; err != nil {
			return "", "", err
		}

		// 检查 TokenVersion 是否匹配
		if user.TokenVersion != tokenVersion {
			return "", "", errors.New("token is no longer valid")
		}

		// 检查用户状态是否为 Active
		if user.Status != models.StatusActive {
			return "", "", errors.New("user is not active")
		}

		// 检查是否需要更新 Refresh Token
		if expirationTime-currentTime < int64(s.refreshTokenSlidingWindow.Seconds()) {
			// 生成新的 Access Token 和 Refresh Token
			newAccessToken, newRefreshToken, err := s.GenerateTokens(user)
			if err != nil {
				return "", "", err
			}
			return newAccessToken, newRefreshToken, nil
		}

		// 只生成新的 Access Token
		newAccessToken, err := s.GenerateAccessToken(user)
		if err != nil {
			return "", "", err
		}

		return newAccessToken, refreshTokenString, nil
	}

	return "", "", errors.New("invalid token")
}

func (s *AuthService) InvalidateUserToken(userID uint) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("token_version", gorm.Expr("token_version + ?", 1)).Error
}
