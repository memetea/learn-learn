package main

import (
	"context"
	"flag"
	"learn/config"
	"learn/internal/api"
	"learn/internal/database"
	"learn/internal/routes"
	"learn/internal/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// @title Question Bank API
// @version 1.0
// @description This is an API for managing question banks.
// @host localhost:8080
// @BasePath /

func main() {
	// 定义命令行参数
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// 读取配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	// 初始化数据库
	db := initDB(cfg)
	db = db.Debug()

	// 初始化服务
	authService, quizService := initServices(db, cfg)

	//初始化admin
	initAdmin(authService, cfg)

	// 初始化处理器
	authHandler, questionHandler := initHandlers(authService, quizService)

	// 初始化路由
	// enforcer, err := loadCasbinEnforcer(authService)
	// if err != nil {
	// 	log.Fatalf("Failed to load casbin enforcer: %v", err)
	// }
	router := initRouter(authHandler, questionHandler)
	enableSwagger(router, cfg.Server.Address)
	// printRoutes(router)

	// 创建并启动服务器
	srv := startServer(cfg, router)

	// 监听关闭信号并优雅关闭
	gracefulShutdown(srv)
}

// 初始化数据库
func initDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	err = database.Migrate(db)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	return db
}

// 初始化服务层
func initServices(db *gorm.DB, cfg *config.Config) (*services.AuthService, *services.QuizService) {
	authService := services.NewAuthService(db, cfg.JWT.Secret, cfg.JWT.AccessTokenDuration, cfg.JWT.RefreshTokenDuration, 24*time.Hour)
	quizService := services.NewQuizService(db)
	return authService, quizService
}

// 初始化处理器
func initHandlers(authService *services.AuthService, quizService *services.QuizService) (*api.AuthHandler, *api.QuizHandler) {
	authHandler := &api.AuthHandler{AuthService: authService}
	questionHandler := &api.QuizHandler{QuizService: quizService}
	return authHandler, questionHandler
}

// 初始化路由
func initRouter(authHandler *api.AuthHandler, quizHandler *api.QuizHandler) *mux.Router {
	router := mux.NewRouter()

	register := routes.NewRoutesRegister(router, authHandler.AuthService)
	err := register.RegisterRoutes(authHandler, quizHandler)
	if err != nil {
		log.Fatalf("Failed to register routes: %v", err)
	}

	// 增加健康检查端点
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()

		log.Printf("Route %s %s %v", route.GetName(), path, methods)
		return nil
	})
	return router
}

func initAdmin(authService *services.AuthService, cfg *config.Config) {
	//insert admin role first
	_, err := authService.CreateRole("admin")
	if err != nil {
		log.Panicf("Failed to create admin role: %v", err)
	}

	err = authService.InitializeAdminUser(cfg.DefaultAdmin.Username,
		cfg.DefaultAdmin.Password)
	if err != nil {
		log.Fatalf("Admin initialize error: %v", err)
	}
}

// func loadCasbinEnforcer(authService *services.AuthService) (*casbin.Enforcer, error) {
// 	// 加载 Casbin 模型
// 	m, err := model.NewModelFromString(`
// 		[request_definition]
// 		r = sub, obj, act

// 		[policy_definition]
// 		p = sub, obj, act

// 		[role_definition]
// 		g = _, _

// 		[policy_effect]
// 		e = some(where (p.eft == allow))

// 		[matchers]
// 		m = (r.obj == p.obj || p.obj == "*") && (r.act == p.act || p.act == "*") && g(r.sub, p.sub)
// 	`)
// 	if err != nil {
// 		log.Fatalf("failed to load model: %v", err)
// 	}

// 	// 初始化 Casbin enforcer
// 	enforcer, err := casbin.NewEnforcer(m)
// 	if err != nil {
// 		log.Fatalf("failed to create enforcer: %v", err)
// 	}

// 	enforcer.AddPolicy("admin", "*", "")

// 	roles, err := authService.GetRoles()
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, role := range roles {
// 		roleName := strings.ToLower(role.Name)
// 		permissions, err := authService.GetRolePermissions(role.ID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for _, permission := range permissions {
// 			enforcer.AddPolicy(roleName, permission.Name, "")
// 		}
// 	}
// 	return enforcer, nil
// }

// 启动服务器
func startServer(cfg *config.Config, router *mux.Router) *http.Server {
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins(cfg.Server.AllowedOrigins), // 指定允许的来源
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      corsMiddleware(router),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Printf("Starting server on %s", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", cfg.Server.Address, err)
		}
	}()

	return srv
}

// 优雅关闭服务器
func gracefulShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
