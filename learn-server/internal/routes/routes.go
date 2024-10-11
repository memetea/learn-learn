package routes

import (
	"encoding/json"
	"learn/internal/api"
	"learn/internal/consts/contextkeys"
	"learn/internal/dto"
	"learn/internal/middleware"
	"learn/internal/models"
	"learn/internal/services"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type RoutesRegister struct {
	r           *mux.Router
	sr          *mux.Router //login required
	authService *services.AuthService
	providers   []api.APIEndpointProvider
}

func NewRoutesRegister(r *mux.Router, authSerivce *services.AuthService) *RoutesRegister {
	mgr := &RoutesRegister{
		r:           r,
		sr:          r.PathPrefix("/").Subrouter(),
		authService: authSerivce,
	}
	mgr.sr.Use(middleware.AuthMiddleware(authSerivce))
	//mgr.sr.Use(middleware.CasbinMiddleware(enforcer))
	mgr.sr.HandleFunc("/menu", mgr.getMenu).Methods("GET")
	return mgr
}

func (r *RoutesRegister) registerApiEndpoint(endpoint api.APIEndpoint) {
	if len(endpoint.Permission) > 0 {
		handler := middleware.CasbinMiddlewareFunc(r.authService.CasbinEnforcer(), endpoint.Permission)(endpoint.Handler)
		r.sr.Handle(endpoint.Path, handler).Methods(endpoint.Method)
	} else {
		r.r.HandleFunc(endpoint.Path, endpoint.Handler).Methods(endpoint.Method)
	}
}

func (r *RoutesRegister) RegisterRoutes(providers ...api.APIEndpointProvider) error {
	for _, provider := range providers {
		r.providers = append(r.providers, provider)
		for _, endpoint := range provider.GetApiEndpoints() {
			r.registerApiEndpoint(endpoint)
			if len(endpoint.Permission) > 0 {
				err := r.authService.EnsurePermissionExists(endpoint.Permission, endpoint.PermissionDescription)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type MenuItem struct {
	dto.MenuItem
	Permission string `json:"permission"`
}

func readMenuFromFile(path string) ([]MenuItem, error) {
	menu := struct {
		Menus []MenuItem `json:"menus"`
	}{}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(content, &menu); err != nil {
		return nil, err
	}
	return menu.Menus, nil
}

var menus []MenuItem

func init() {
	var err error
	menus, err = readMenuFromFile("menu.json")
	if err != nil {
		log.Println(err)
	}
}

func (rr *RoutesRegister) getMenu(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(contextkeys.User).(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	menu := make([]MenuItem, 0)
	for _, m := range menus {
		if m.Permission == "" {
			menu = append(menu, m)
		} else {
			for _, role := range user.Roles {
				if ok, _ := rr.authService.CasbinEnforcer().Enforce(role.Name, m.Permission, ""); ok {
					menu = append(menu, m)
				}
			}
		}
	}

	api.Success(w, menu, nil, http.StatusOK)
}
