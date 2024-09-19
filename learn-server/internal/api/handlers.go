package api

import (
	"encoding/json"
	"learn/internal/models"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type App struct {
	DB *gorm.DB
}

// GetQuestionBanks 获取题库列表
// @Summary 获取题库列表
// @Description 获取所有的题库
// @Tags QuestionBank
// @Accept  json
// @Produce  json
// @Success 200 {array} models.QuestionBank
// @Router /question_banks [get]
func (app *App) GetQuestionBanks(w http.ResponseWriter, r *http.Request) {
	var questionBanks []models.QuestionBank
	app.DB.Find(&questionBanks)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questionBanks)
}

// GetQuestionsByBank 获取题库中的问题
// @Summary 获取题库中的问题
// @Description 获取某个题库中的所有问题
// @Tags Question
// @Accept  json
// @Produce  json
// @Param bank_id path int true "题库ID"
// @Success 200 {array} models.Question
// @Failure 404 {string} string "Question bank not found"
// @Router /question_banks/{bank_id}/questions [get]
func (app *App) GetQuestionsByBank(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var questions []models.Question
	if err := app.DB.Where("question_bank_id = ?", params["bank_id"]).Preload("AnswerOptions").Find(&questions).Error; err != nil {
		http.Error(w, "Question bank not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
}

// GetQuestionDetails 获取问题详情
// @Summary 获取问题详情
// @Description 获取某个问题的详细信息，包括答案选项
// @Tags Question
// @Accept  json
// @Produce  json
// @Param question_id path int true "问题ID"
// @Success 200 {object} models.Question
// @Failure 404 {string} string "Question not found"
// @Router /questions/{question_id} [get]
func (app *App) GetQuestionDetails(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var question models.Question
	if err := app.DB.Preload("AnswerOptions").First(&question, params["question_id"]).Error; err != nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(question)
}

// CreateQuestionBank 新增题库
// @Summary 新增题库
// @Description 新增题库
// @Tags QuestionBank
// @Accept  json
// @Produce  json
// @Param name body models.QuestionBank true "题库名称"
// @Success 201 {object} models.QuestionBank
// @Failure 400 {string} string "Invalid input"
// @Failure 500 {string} string "Failed to create question bank"
// @Router /question_banks [post]
func (app *App) CreateQuestionBank(w http.ResponseWriter, r *http.Request) {
	var newQuestionBank models.QuestionBank

	// 从请求体中解析 JSON 数据
	if err := json.NewDecoder(r.Body).Decode(&newQuestionBank); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// 插入数据到数据库
	if err := app.DB.Create(&newQuestionBank).Error; err != nil {
		http.Error(w, "Failed to create question bank", http.StatusInternalServerError)
		return
	}

	// 返回新创建的题库信息
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newQuestionBank)
}

// UpdateQuestionBank 修改题库
// @Summary 修改题库
// @Description 根据ID修改题库
// @Tags QuestionBank
// @Accept  json
// @Produce  json
// @Param id path int true "题库ID"
// @Param name body models.QuestionBank true "题库名称"
// @Success 200 {object} models.QuestionBank
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "Question bank not found"
// @Failure 500 {string} string "Failed to update question bank"
// @Router /question_banks/{id} [put]
func (app *App) UpdateQuestionBank(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// 查找指定ID的题库
	var questionBank models.QuestionBank
	if err := app.DB.First(&questionBank, id).Error; err != nil {
		http.Error(w, "Question bank not found", http.StatusNotFound)
		return
	}

	// 从请求体中解析新的数据
	if err := json.NewDecoder(r.Body).Decode(&questionBank); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// 更新数据到数据库
	if err := app.DB.Save(&questionBank).Error; err != nil {
		http.Error(w, "Failed to update question bank", http.StatusInternalServerError)
		return
	}

	// 返回更新后的题库信息
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questionBank)
}
