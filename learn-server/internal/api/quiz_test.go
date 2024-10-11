// api/quiz_test.go
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

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 设置测试数据库和必要服务
func setupTestQuizHandler() (*api.QuizHandler, *services.AuthService, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	err = db.AutoMigrate(&models.QuestionBank{}, &models.Question{},
		&models.AnswerOption{}, &models.TrueFalseAnswer{}, &models.WrittenAnswer{},
		&models.FillInTheBlankAnswer{}, &models.Tag{},
		&models.User{}, &models.QuestionAttempt{})
	if err != nil {
		return nil, nil, err
	}

	quizService := services.NewQuizService(db)
	authService := services.NewAuthService(db, "jwt_secret", 60, 3600, 86400)
	quizHandler := &api.QuizHandler{QuizService: quizService}

	return quizHandler, authService, nil
}

// Helper function to create a user using AuthService
func createTestUser(authService *services.AuthService, username string) (*models.User, error) {
	user, err := authService.CreateUser(username, "password", []string{}, models.StatusActive)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func TestCreateMultipleChoiceQuestion(t *testing.T) {
	handler, authService, err := setupTestQuizHandler()
	if err != nil {
		t.Fatalf("Failed to setup quiz handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/quiz/question_banks/{id}/questions", handler.CreateQuestion).Methods("POST")

	// 使用 QuizService 创建题库
	questionBank, err := handler.QuizService.CreateQuestionBank("Sample Bank")
	if err != nil {
		t.Fatalf("Failed to create question bank: %v", err)
	}

	// 使用 AuthService 创建用户
	user, err := createTestUser(authService, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 创建一个选择题
	requestBody, _ := json.Marshal(dto.CreateQuestionRequest{
		Content:      "What is 2+2?",
		QuestionType: models.QuestionTypeMultipleChoice, // 多选题类型
		AnswerOptions: []dto.AnswerOption{
			{OptionText: "3", IsCorrect: false},
			{OptionText: "4", IsCorrect: true},
		},
		Tags:     []string{"math", "basic"},
		AuthorID: user.ID, // 设置问题的作者
	})
	req, err := http.NewRequest("POST", "/quiz/question_banks/"+strconv.Itoa(int(questionBank.ID))+"/questions", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	t.Logf("Response status code: %v", rr.Code)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %v, got %v", http.StatusCreated, rr.Code)
	}

	// 使用创建的 ID 再次获取问题详细信息
	var response api.Response[dto.QuestionResponse]
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// 通过 GetQuestionDetail 获取完整问题数据
	detailedQuestion, err := handler.QuizService.GetQuestionDetail(response.Data.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve created question details: %v", err)
	}

	// 验证标签是否正确保存
	if len(detailedQuestion.Tags) != 2 || detailedQuestion.Tags[0].Name != "math" || detailedQuestion.Tags[1].Name != "basic" {
		t.Errorf("Expected tags ['math', 'basic'], got %v", detailedQuestion.Tags)
	}
}

func TestCreateTrueFalseQuestion(t *testing.T) {
	handler, authService, err := setupTestQuizHandler()
	if err != nil {
		t.Fatalf("Failed to setup quiz handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/quiz/question_banks/{id}/questions", handler.CreateQuestion).Methods("POST")

	// 使用 QuizService 创建题库
	questionBank, err := handler.QuizService.CreateQuestionBank("Sample Bank")
	if err != nil {
		t.Fatalf("Failed to create question bank: %v", err)
	}

	// 使用 AuthService 创建用户
	user, err := createTestUser(authService, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 创建一个判断题
	trueFalseValue := true
	requestBody, _ := json.Marshal(dto.CreateQuestionRequest{
		Content:      "Is 5 > 3?",
		QuestionType: models.QuestionTypeTrueFalse, // 判断题类型
		TrueFalse:    &trueFalseValue,
		Tags:         []string{"logic", "comparison"}, // 添加标签
		AuthorID:     user.ID,                         // 设置问题的作者
	})
	req, err := http.NewRequest("POST", "/quiz/question_banks/"+strconv.Itoa(int(questionBank.ID))+"/questions", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %v, got %v", http.StatusCreated, rr.Code)
	}

	var response api.Response[dto.QuestionResponse]
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Data.Content != "Is 5 > 3?" {
		t.Errorf("Expected question content 'Is 5 > 3?', got '%v'", response.Data.Content)
	}

	if len(response.Data.Tags) != 2 || response.Data.Tags[0] != "logic" || response.Data.Tags[1] != "comparison" {
		t.Errorf("Expected tags ['logic', 'comparison'], got %v", response.Data.Tags)
	}
}

func TestCreateWrittenQuestion(t *testing.T) {
	handler, authService, err := setupTestQuizHandler()
	if err != nil {
		t.Fatalf("Failed to setup quiz handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/quiz/question_banks/{id}/questions", handler.CreateQuestion).Methods("POST")

	// 使用 QuizService 创建题库
	questionBank, err := handler.QuizService.CreateQuestionBank("Sample Bank")
	if err != nil {
		t.Fatalf("Failed to create question bank: %v", err)
	}

	// 使用 AuthService 创建用户
	user, err := createTestUser(authService, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 创建一个问答题
	requestBody, _ := json.Marshal(dto.CreateQuestionRequest{
		Content:      "Explain what RESTful APIs are.",
		QuestionType: models.QuestionTypeWrittenAnswer, // 简答题类型
		AnswerText:   "RESTful APIs are based on representational state transfer...",
		Tags:         []string{"technology", "API"}, // 添加标签
		AuthorID:     user.ID,                       // 设置问题的作者
	})
	req, err := http.NewRequest("POST", "/quiz/question_banks/"+strconv.Itoa(int(questionBank.ID))+"/questions", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %v, got %v", http.StatusCreated, rr.Code)
	}

	var response api.Response[dto.QuestionResponse]
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Data.Content != "Explain what RESTful APIs are." {
		t.Errorf("Expected question content 'Explain what RESTful APIs are.', got '%v'", response.Data.Content)
	}

	if len(response.Data.Tags) != 2 || response.Data.Tags[0] != "technology" || response.Data.Tags[1] != "API" {
		t.Errorf("Expected tags ['technology', 'API'], got %v", response.Data.Tags)
	}
}

func TestCreateFillInTheBlankQuestion(t *testing.T) {
	handler, authService, err := setupTestQuizHandler()
	if err != nil {
		t.Fatalf("Failed to setup quiz handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/quiz/question_banks/{id}/questions", handler.CreateQuestion).Methods("POST")

	// 使用 QuizService 创建题库
	questionBank, err := handler.QuizService.CreateQuestionBank("Sample Bank")
	if err != nil {
		t.Fatalf("Failed to create question bank: %v", err)
	}

	// 使用 AuthService 创建用户
	user, err := createTestUser(authService, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 创建一个填空题
	requestBody, _ := json.Marshal(dto.CreateQuestionRequest{
		Content:      "The capital of France is __.",
		QuestionType: models.QuestionTypeFillInTheBlank, // 填空题类型
		Blanks: []dto.FillInTheBlankAnswer{
			{BlankText: "Paris"},
		},
		Tags:     []string{"geography", "capital"}, // 添加标签
		AuthorID: user.ID,                          // 设置问题的作者
	})
	req, err := http.NewRequest("POST", "/quiz/question_banks/"+strconv.Itoa(int(questionBank.ID))+"/questions", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %v, got %v", http.StatusCreated, rr.Code)
	}

	var response api.Response[dto.QuestionResponse]
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Data.Content != "The capital of France is __." {
		t.Errorf("Expected question content 'The capital of France is __.', got '%v'", response.Data.Content)
	}

	if len(response.Data.FillInTheBlanks) != 1 || response.Data.FillInTheBlanks[0].BlankText != "Paris" {
		t.Errorf("Expected blank answer 'Paris', got '%v'", response.Data.FillInTheBlanks)
	}

	if len(response.Data.Tags) != 2 || response.Data.Tags[0] != "geography" || response.Data.Tags[1] != "capital" {
		t.Errorf("Expected tags ['geography', 'capital'], got %v", response.Data.Tags)
	}
}

func TestRecordQuestionAttempt(t *testing.T) {
	handler, authService, err := setupTestQuizHandler()
	if err != nil {
		t.Fatalf("Failed to setup quiz handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/quiz/question_attempts", handler.RecordQuestionAttempt).Methods("POST")
	router.HandleFunc("/quiz/question_banks/{id}/questions", handler.CreateQuestion).Methods("POST")

	// Using AuthService to create a user
	user, err := createTestUser(authService, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Using QuizService to create a question bank
	questionBank, err := handler.QuizService.CreateQuestionBank("Sample Bank")
	if err != nil {
		t.Fatalf("Failed to create question bank: %v", err)
	}

	// Create a single choice question via HTTP API with answers included
	questionRequestBody, _ := json.Marshal(dto.CreateQuestionRequest{
		Content:      "What is 2+2?",
		QuestionType: models.QuestionTypeSingleChoice,
		AuthorID:     user.ID,
		AnswerOptions: []dto.AnswerOption{
			{OptionText: "3", IsCorrect: false},
			{OptionText: "4", IsCorrect: true},
		},
	})
	createQuestionReq := httptest.NewRequest(http.MethodPost, "/quiz/question_banks/"+strconv.Itoa(int(questionBank.ID))+"/questions", bytes.NewBuffer(questionRequestBody))
	createQuestionReq.Header.Set("Content-Type", "application/json")
	createQuestionW := httptest.NewRecorder()

	router.ServeHTTP(createQuestionW, createQuestionReq)

	if createQuestionW.Result().StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create question via HTTP API, status code: %v", createQuestionW.Result().StatusCode)
	}

	var createQuestionResponse api.Response[dto.QuestionResponse]
	err = json.NewDecoder(createQuestionW.Body).Decode(&createQuestionResponse)
	if err != nil {
		t.Fatalf("Failed to decode create question response: %v", err)
	}

	questionID := createQuestionResponse.Data.ID

	// Record an attempt for the single choice question with the correct answer
	reqBody := dto.QuestionAttemptRequest{
		UserID:     user.ID,
		QuestionID: questionID,
		Answer:     []uint{createQuestionResponse.Data.AnswerOptions[1].ID}, // Selecting the correct option
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/quiz/question_attempts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %v, got %v", http.StatusOK, res.StatusCode)
	}

	// Check response body for success status
	var responseBody map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&responseBody)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if responseBody["status"] != "success" {
		t.Fatalf("Expected success status, got %v", responseBody["status"])
	}

	// Verify the attempt record in the database
	attempt, err := handler.QuizService.GetQuestionAttempt(user.ID, questionID)
	if err != nil {
		t.Fatalf("Failed to retrieve question attempt: %v", err)
	}
	if attempt == nil {
		t.Fatalf("Expected an attempt record, but got nil")
	}

	if attempt.Attempts != 1 {
		t.Errorf("Expected 1 attempt, got %v", attempt.Attempts)
	}

	if attempt.ConsecutiveCorrect != 1 {
		t.Errorf("Expected 1 consecutive correct, got %v", attempt.ConsecutiveCorrect)
	}

	// Verify last answer is recorded
	if attempt.LastAnswer == nil {
		t.Fatalf("Expected last answer to be recorded, but got nil")
	}

	// Create a true/false question via HTTP API with answer included
	trueFalseValue := true
	questionRequestBody, _ = json.Marshal(dto.CreateQuestionRequest{
		Content:      "Is 5 greater than 3?",
		QuestionType: models.QuestionTypeTrueFalse,
		AuthorID:     user.ID,
		TrueFalse:    &trueFalseValue, // Setting the correct answer to true
	})
	createQuestionReq = httptest.NewRequest(http.MethodPost, "/quiz/question_banks/"+strconv.Itoa(int(questionBank.ID))+"/questions", bytes.NewBuffer(questionRequestBody))
	createQuestionReq.Header.Set("Content-Type", "application/json")
	createQuestionW = httptest.NewRecorder()

	router.ServeHTTP(createQuestionW, createQuestionReq)

	if createQuestionW.Result().StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create question via HTTP API, status code: %v", createQuestionW.Result().StatusCode)
	}

	err = json.NewDecoder(createQuestionW.Body).Decode(&createQuestionResponse)
	if err != nil {
		t.Fatalf("Failed to decode create question response: %v", err)
	}

	trueFalseQuestionID := createQuestionResponse.Data.ID

	// Record an attempt for the true/false question with the correct answer
	reqBody = dto.QuestionAttemptRequest{
		UserID:     user.ID,
		QuestionID: trueFalseQuestionID,
		Answer:     true,
	}
	jsonBody, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/quiz/question_attempts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	res = w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %v, got %v", http.StatusOK, res.StatusCode)
	}

	// Check response body for success status
	err = json.NewDecoder(res.Body).Decode(&responseBody)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if responseBody["status"] != "success" {
		t.Fatalf("Expected success status, got %v", responseBody["status"])
	}

	// Verify the attempt record in the database
	attempt, err = handler.QuizService.GetQuestionAttempt(user.ID, trueFalseQuestionID)
	if err != nil {
		t.Fatalf("Failed to retrieve question attempt: %v", err)
	}
	if attempt == nil {
		t.Fatalf("Expected an attempt record, but got nil")
	}

	if attempt.Attempts != 1 {
		t.Errorf("Expected 1 attempt, got %v", attempt.Attempts)
	}

	if attempt.ConsecutiveCorrect != 1 {
		t.Errorf("Expected 1 consecutive correct, got %v", attempt.ConsecutiveCorrect)
	}

	// Verify last answer is recorded
	if attempt.LastAnswer == nil {
		t.Fatalf("Expected last answer to be recorded, but got nil")
	}

	// Create a written question via HTTP API
	questionRequestBody, _ = json.Marshal(dto.CreateQuestionRequest{
		Content:      "Explain what RESTful APIs are.",
		QuestionType: models.QuestionTypeWrittenAnswer,
		AuthorID:     user.ID,
	})
	createQuestionReq = httptest.NewRequest(http.MethodPost, "/quiz/question_banks/"+strconv.Itoa(int(questionBank.ID))+"/questions", bytes.NewBuffer(questionRequestBody))
	createQuestionReq.Header.Set("Content-Type", "application/json")
	createQuestionW = httptest.NewRecorder()

	router.ServeHTTP(createQuestionW, createQuestionReq)

	if createQuestionW.Result().StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create question via HTTP API, status code: %v", createQuestionW.Result().StatusCode)
	}

	err = json.NewDecoder(createQuestionW.Body).Decode(&createQuestionResponse)
	if err != nil {
		t.Fatalf("Failed to decode create question response: %v", err)
	}

	writtenQuestionID := createQuestionResponse.Data.ID

	// Record an attempt for the written question without an answer
	reqBody = dto.QuestionAttemptRequest{
		UserID:     user.ID,
		QuestionID: writtenQuestionID,
		Answer:     "", // No answer provided
	}
	jsonBody, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/quiz/question_attempts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	res = w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %v, got %v", http.StatusOK, res.StatusCode)
	}

	// Check response body for success status
	err = json.NewDecoder(res.Body).Decode(&responseBody)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if responseBody["status"] != "success" {
		t.Fatalf("Expected success status, got %v", responseBody["status"])
	}

	// Verify the attempt record in the database
	attempt, err = handler.QuizService.GetQuestionAttempt(user.ID, writtenQuestionID)
	if err != nil {
		t.Fatalf("Failed to retrieve question attempt: %v", err)
	}
	if attempt == nil {
		t.Fatalf("Expected an attempt record, but got nil")
	}

	if attempt.Attempts != 1 {
		t.Errorf("Expected 1 attempt, got %v", attempt.Attempts)
	}

	if attempt.ConsecutiveCorrect != 0 {
		t.Errorf("Expected 0 consecutive correct, got %v", attempt.ConsecutiveCorrect)
	}

	// Verify last answer is recorded
	if attempt.LastAnswer == nil {
		t.Fatalf("Expected last answer to be recorded, but got nil")
	}
}

func TestGetQuestionAttempts(t *testing.T) {
	handler, authService, err := setupTestQuizHandler()
	if err != nil {
		t.Fatalf("Failed to setup quiz handler: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/quiz/question_attempts/{user_id}/{question_bank_id}", handler.GetQuestionAttempts).Methods("GET")
	router.HandleFunc("/quiz/question_banks/{id}/questions", handler.CreateQuestion).Methods("POST")

	// Using AuthService to create a user
	user, err := createTestUser(authService, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Using QuizService to create a question bank
	questionBank, err := handler.QuizService.CreateQuestionBank("Sample Bank")
	if err != nil {
		t.Fatalf("Failed to create question bank: %v", err)
	}

	// Create a question via HTTP API with answer included
	trueFalseValue := true
	questionRequestBody, _ := json.Marshal(dto.CreateQuestionRequest{
		Content:      "What is 2+2?",
		QuestionType: models.QuestionTypeTrueFalse,
		AuthorID:     user.ID,
		TrueFalse:    &trueFalseValue, // Setting the correct answer to true
	})
	createQuestionReq := httptest.NewRequest(http.MethodPost, "/quiz/question_banks/"+strconv.Itoa(int(questionBank.ID))+"/questions", bytes.NewBuffer(questionRequestBody))
	createQuestionReq.Header.Set("Content-Type", "application/json")
	createQuestionW := httptest.NewRecorder()

	router.ServeHTTP(createQuestionW, createQuestionReq)

	if createQuestionW.Result().StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create question via HTTP API, status code: %v", createQuestionW.Result().StatusCode)
	}

	var createQuestionResponse api.Response[dto.QuestionResponse]
	err = json.NewDecoder(createQuestionW.Body).Decode(&createQuestionResponse)
	if err != nil {
		t.Fatalf("Failed to decode create question response: %v", err)
	}

	questionID := createQuestionResponse.Data.ID

	// Record an attempt for the question
	_, err = handler.QuizService.RecordQuestionAttempt(user.ID, questionID, true)
	if err != nil {
		t.Fatalf("Failed to record question attempt: %v", err)
	}

	// Request to get user's question attempts
	req := httptest.NewRequest(http.MethodGet, "/quiz/question_attempts/"+strconv.Itoa(int(user.ID))+"/"+strconv.Itoa(int(questionBank.ID)), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, res.StatusCode)
	}

	var response api.Response[[]dto.QuestionAttemptResponse]
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Data) != 1 {
		t.Errorf("Expected 1 attempt record, got %v", len(response.Data))
	}

	if response.Data[0].QuestionID != questionID {
		t.Errorf("Expected QuestionID %v, got %v", questionID, response.Data[0].QuestionID)
	}
}
