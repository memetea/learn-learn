package routes

import (
	"learn/internal/api"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *mux.Router {
	router := mux.NewRouter()

	// 创建 App 实例并注入数据库
	app := &api.App{DB: db}

	// 定义路由
	router.HandleFunc("/question_banks", app.GetQuestionBanks).Methods("GET")
	router.HandleFunc("/question_banks", app.CreateQuestionBank).Methods("POST")     // 新增题库
	router.HandleFunc("/question_banks/{id}", app.UpdateQuestionBank).Methods("PUT") // 修改题库
	router.HandleFunc("/question_banks/{bank_id}/questions", app.GetQuestionsByBank).Methods("GET")
	router.HandleFunc("/questions/{question_id}", app.GetQuestionDetails).Methods("GET")

	return router
}
