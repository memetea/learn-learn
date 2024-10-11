// middleware/casbin_middleware.go
package middleware

import (
	"learn/internal/consts/contextkeys"
	"learn/internal/models"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gorilla/mux"
)

// func CasbinMiddleware(enforcer *casbin.Enforcer) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			user, ok := r.Context().Value(contextkeys.User).(models.User)
// 			if !ok {
// 				http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 				return
// 			}

// 			obj := r.URL.Path
// 			act := r.Method

// 			// Casbin 权限检查
// 			for _, role := range user.Roles {
// 				ok, err := enforcer.Enforce(role.Name, obj, act)
// 				if err != nil {
// 					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 					return
// 				}
// 				if ok {
// 					next.ServeHTTP(w, r)
// 					return
// 				}
// 			}
// 			http.Error(w, "Forbidden", http.StatusForbidden)
// 		})
// 	}
// }

func CasbinMiddlewareFunc(e *casbin.Enforcer, permission string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(contextkeys.User).(models.User)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Casbin 权限检查
			for _, role := range user.Roles {
				if ok, _ := e.Enforce(role.Name, permission, ""); ok {
					next.ServeHTTP(w, r)
					return
				} else {
					http.Error(w, "Forbidden", http.StatusForbidden)
				}
			}
		})
	}
}
