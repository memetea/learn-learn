// middleware/auth_middleware.go
package middleware

import (
	"context"
	"learn/internal/consts/claimkeys"
	"learn/internal/consts/contextkeys"
	"learn/internal/models"
	"learn/internal/services"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// 验证 JWT 并解析用户信息
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(authService.JwtSecret()), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// 验证 TokenVersion 是否匹配
			userID := uint(claims[claimkeys.UserId].(float64))
			tokenVersion := uint(claims[claimkeys.TokenVersion].(float64))

			user, err := authService.GetUserByID(userID)
			if err != nil {
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			if user.TokenVersion != tokenVersion {
				http.Error(w, "Token is no longer valid", http.StatusUnauthorized)
				return
			}

			if roles, ok := claims[claimkeys.Role].([]any); ok {
				for _, role := range roles {
					user.Roles = append(user.Roles, models.Role{Name: role.(string)})
				}
			}

			// 设置用户上下文，继续处理请求
			ctx := context.WithValue(r.Context(), contextkeys.User, user)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// func RequirePermission(authService *services.AuthService, permission string) func(http.Handler) http.Handler {
// 	// 确保权限存在于数据库中
// 	if err := authService.EnsurePermissionExists(permission); err != nil {
// 		// 记录错误而不是 panic，避免生产环境崩溃
// 		log.Fatalf("Failed to ensure permission exists: %v", err)
// 	}

// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			// 从上下文中获取用户
// 			user, ok := r.Context().Value(contextkeys.User).(*models.User) // 假设你在服务层定义了 User 模型
// 			if !ok {
// 				http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 				return
// 			}

// 			// 使用 AuthService 来检查用户是否具有所需的权限
// 			if !authService.HasPermission(user, permission) {
// 				http.Error(w, "Forbidden", http.StatusForbidden)
// 				return
// 			}

// 			http.Error(w, "Forbidden", http.StatusForbidden)
// 		})
// 	}
// }

// func RequirePermissionWrap(authService *services.AuthService, permission string, handler http.HandlerFunc) http.Handler {
// 	// 确保权限存在于数据库中
// 	if err := authService.EnsurePermissionExists(permission); err != nil {
// 		// 记录错误而不是 panic，避免生产环境崩溃
// 		log.Fatalf("Failed to ensure permission exists: %v", err)
// 	}

// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// 从上下文中获取用户
// 		user, ok := r.Context().Value(contextkeys.User).(*models.User) // 假设你在服务层定义了 User 模型
// 		if !ok {
// 			http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 			return
// 		}

// 		// 使用 AuthService 来检查用户是否具有所需的权限
// 		if !authService.HasPermission(user, permission) {
// 			http.Error(w, "Forbidden", http.StatusForbidden)
// 			return
// 		}

// 		// 如果用户具有权限，则调用下一个处理器
// 		handler.ServeHTTP(w, r)
// 	})
// }
