package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"learn/internal/models"
	"learn/internal/routes"

	"github.com/stretchr/testify/assert"
)

var testDB *gorm.DB

// setupTestDB 初始化内存中的 SQLite 数据库，并将其注入到 App 结构体中
func setupTestDB(t *testing.T) {
	// 初始化内存数据库
	var err error
	testDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err, "failed to connect to test database")

	// 自动迁移表结构
	err = testDB.AutoMigrate(&models.QuestionBank{}, &models.QuestionType{}, &models.Question{}, &models.AnswerOption{}, &models.RelatedQuestion{})
	assert.NoError(t, err, "failed to migrate database")
}

func TestGetQuestionBanks(t *testing.T) {
	setupTestDB(t) // 初始化并清理数据库

	// 插入测试数据
	err := testDB.Create(&models.QuestionBank{Name: "三年级英语"}).Error
	assert.NoError(t, err, "failed to create question bank")

	err = testDB.Create(&models.QuestionBank{Name: "四年级数学"}).Error
	assert.NoError(t, err, "failed to create question bank")

	// 发起请求
	req, err := http.NewRequest("GET", "/question_banks", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// 使用路由器处理请求
	router := routes.SetupRouter(testDB)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// 解析响应体并检查
	var questionBanks []models.QuestionBank
	err = json.Unmarshal(rr.Body.Bytes(), &questionBanks)
	assert.NoError(t, err)

	// 检查题库的数量和内容
	assert.Len(t, questionBanks, 2)
	assert.Equal(t, "三年级英语", questionBanks[0].Name)
	assert.Equal(t, "四年级数学", questionBanks[1].Name)
}

func TestGetQuestionsByBank(t *testing.T) {
	setupTestDB(t) // 初始化并清理数据库

	// 插入测试数据
	qb := models.QuestionBank{Name: "三年级英语"}
	err := testDB.Create(&qb).Error
	assert.NoError(t, err, "failed to create question bank")

	qt := models.QuestionType{Name: "选择题"}
	err = testDB.Create(&qt).Error
	assert.NoError(t, err, "failed to create question type")

	question := models.Question{
		QuestionBankID: qb.ID,
		QuestionTypeID: qt.ID,
		Content:        "What is the capital of France?",
		Explanation:    "Paris is the capital of France.",
	}
	err = testDB.Create(&question).Error
	assert.NoError(t, err, "failed to create question")

	// 发起请求
	req, err := http.NewRequest("GET", "/question_banks/1/questions", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// 使用路由器处理请求
	router := routes.SetupRouter(testDB)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// 解析响应体并检查
	var questions []models.Question
	err = json.Unmarshal(rr.Body.Bytes(), &questions)
	assert.NoError(t, err)

	// 检查问题是否返回正确
	assert.Len(t, questions, 1)
	assert.Equal(t, "What is the capital of France?", questions[0].Content)
}

func TestGetQuestionDetails(t *testing.T) {
	setupTestDB(t) // 初始化并清理数据库

	// 插入测试数据
	qb := models.QuestionBank{Name: "三年级英语"}
	err := testDB.Create(&qb).Error
	assert.NoError(t, err, "failed to create question bank")

	qt := models.QuestionType{Name: "选择题"}
	err = testDB.Create(&qt).Error
	assert.NoError(t, err, "failed to create question type")

	question := models.Question{
		QuestionBankID: qb.ID,
		QuestionTypeID: qt.ID,
		Content:        "What is the capital of France?",
		Explanation:    "Paris is the capital of France.",
	}
	err = testDB.Create(&question).Error
	assert.NoError(t, err, "failed to create question")

	// 插入答案选项
	option1 := models.AnswerOption{QuestionID: question.ID, OptionText: "London", IsCorrect: false}
	option2 := models.AnswerOption{QuestionID: question.ID, OptionText: "Paris", IsCorrect: true}
	err = testDB.Create(&option1).Error
	assert.NoError(t, err, "failed to create answer option")
	err = testDB.Create(&option2).Error
	assert.NoError(t, err, "failed to create answer option")

	// 发起请求
	req, err := http.NewRequest("GET", "/questions/1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// 使用路由器处理请求
	router := routes.SetupRouter(testDB)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// 解析响应体并检查
	var questionDetails models.Question
	err = json.Unmarshal(rr.Body.Bytes(), &questionDetails)
	assert.NoError(t, err)

	// 检查问题内容和选项
	assert.Equal(t, "What is the capital of France?", questionDetails.Content)
	assert.Len(t, questionDetails.AnswerOptions, 2)
	assert.Equal(t, "Paris", questionDetails.AnswerOptions[1].OptionText)
	assert.True(t, questionDetails.AnswerOptions[1].IsCorrect)
}

func TestCreateQuestionBank(t *testing.T) {
	setupTestDB(t)

	// 创建一个新的题库
	newQuestionBank := map[string]string{
		"name": "五年级科学",
	}

	reqBody, _ := json.Marshal(newQuestionBank)
	req, err := http.NewRequest("POST", "/question_banks", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	router := routes.SetupRouter(testDB)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var createdBank models.QuestionBank
	err = json.Unmarshal(rr.Body.Bytes(), &createdBank)
	assert.NoError(t, err)

	assert.Equal(t, "五年级科学", createdBank.Name)
}

func TestUpdateQuestionBank(t *testing.T) {
	setupTestDB(t)

	// 创建一个题库
	qb := models.QuestionBank{Name: "五年级科学"}
	err := testDB.Create(&qb).Error
	assert.NoError(t, err)

	// 更新题库的名称
	updatedQuestionBank := map[string]string{
		"name": "六年级科学",
	}

	reqBody, _ := json.Marshal(updatedQuestionBank)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/question_banks/%d", qb.ID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	router := routes.SetupRouter(testDB)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var updatedBank models.QuestionBank
	err = json.Unmarshal(rr.Body.Bytes(), &updatedBank)
	assert.NoError(t, err)

	assert.Equal(t, "六年级科学", updatedBank.Name)
}
