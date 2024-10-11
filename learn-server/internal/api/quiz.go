// api/quiz.go
package api

import (
	"encoding/json"
	"learn/internal/dto"
	"learn/internal/models"
	"learn/internal/services"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type QuizHandler struct {
	QuizService *services.QuizService
}

func (h *QuizHandler) GetApiEndpoints() []APIEndpoint {
	return []APIEndpoint{
		{"/quiz/question_banks", "GET", h.GetQuestionBanks, "quiz:read", "查看题库"},
		{"/quiz/question_banks", "POST", h.CreateQuestionBank, "quiz:edit", "创建题库"},
		{"/quiz/question_banks/{id}/questions", "GET", h.GetQuestions, "quiz:read", "查看题目"},
		{"/quiz/question_banks/{id}/questions", "POST", h.CreateQuestion, "quiz:edit", "创建题目"},
		{"/quiz/questions/{id}", "GET", h.GetQuestionDetail, "quiz:read", "获取问题详细信息"},
		{"/quiz/questions/{id}", "PUT", h.UpdateQuestion, "quiz:edit", "编辑问题"},
		{"/quiz/questions/{id}", "DELETE", h.DeleteQuestion, "quiz:edit", "删除问题"},
		{"/quiz/question_banks/{id}/random_questions", "GET", h.GetRandomQuestions, "", "随机获取题目"},

		{"/quiz/question_attempts", "POST", h.RecordQuestionAttempt, "quiz:edit", "记录答题尝试"},
		{"/quiz/question_attempts/{user_id}/{question_bank_id}", "GET", h.GetQuestionAttempts, "quiz:read", "获取用户的答题尝试情况"},
	}
}

// GetQuestionBanks 获取题库列表
// @Summary 获取题库列表
// @Description 获取所有的题库
// @Tags QuestionBank
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Success 200 {object} Response[[]dto.QuestionBankResponse] "题库列表"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /quiz/question_banks [get]
func (h *QuizHandler) GetQuestionBanks(w http.ResponseWriter, r *http.Request) {
	questionBanks, err := h.QuizService.GetQuestionBanks()
	if err != nil {
		Error(w, "Failed to retrieve question banks", http.StatusInternalServerError)
		return
	}

	response := make([]dto.QuestionBankResponse, len(questionBanks))
	for i, bank := range questionBanks {
		response[i] = dto.QuestionBankResponse{ID: bank.ID, Name: bank.Name}
	}

	Success(w, response, nil, http.StatusOK)
}

// CreateQuestionBank 创建新的题库
// @Summary 创建题库
// @Description 创建一个新的题库
// @Tags QuestionBank
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param questionBank body dto.CreateQuestionBankRequest true "创建题库请求"
// @Success 201 {object} Response[dto.QuestionBankResponse] "创建成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /quiz/question_banks [post]
func (h *QuizHandler) CreateQuestionBank(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateQuestionBankRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	questionBank, err := h.QuizService.CreateQuestionBank(req.Name)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, dto.QuestionBankResponse{ID: questionBank.ID, Name: questionBank.Name}, nil, http.StatusCreated)
}

// GetQuestions 获取题库中的问题基本信息（支持分页和标签过滤）
// @Summary 获取题库问题
// @Description 获取指定题库的所有问题（分页查询）
// @Tags Question
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param id path int true "题库 ID"
// @Param tag query string false "标签过滤"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} Response[[]dto.QuestionResponse] "问题列表"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /quiz/question_banks/{id}/questions [get]
func (h *QuizHandler) GetQuestions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bankIDStr := vars["id"]

	bankID, err := strconv.ParseUint(bankIDStr, 10, 32)
	if err != nil {
		Error(w, "Invalid bank ID", http.StatusBadRequest)
		return
	}

	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")
	tag := r.URL.Query().Get("tag") // 获取 tag 参数
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	questions, total, err := h.QuizService.GetQuestionsWithPagination(uint(bankID), tag, page, pageSize)
	if err != nil {
		Error(w, "Failed to retrieve questions", http.StatusInternalServerError)
		return
	}

	var questionResponses []dto.QuestionResponse
	for _, q := range questions {
		questionResponse := dto.QuestionResponse{
			ID:             q.ID,
			QuestionBankID: q.QuestionBankID,
			QuestionType:   q.QuestionType,
			Content:        q.Content,
			CreatedAt:      q.CreatedAt,
			AuthorID:       q.AuthorID,
		}
		questionResponses = append(questionResponses, questionResponse)
	}

	Success(w, questionResponses, &PaginationMeta{
		TotalRecords: total,
		PageSize:     pageSize,
		CurrentPage:  page,
	}, http.StatusOK)
}

// GetRandomQuestions 获取随机问题
// @Summary 获取随机问题
// @Description 随机获取题库中的问题
// @Tags Question
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param id path int true "题库 ID"
// @Param limit query int false "随机题目数量"
// @Success 200 {object} Response[[]dto.QuestionResponse] "随机题目列表"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /quiz/question_banks/{id}/random_questions [get]
func (h *QuizHandler) GetRandomQuestions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bankIDStr := vars["id"]

	bankID, err := strconv.ParseUint(bankIDStr, 10, 32)
	if err != nil {
		Error(w, "Invalid bank ID", http.StatusBadRequest)
		return
	}

	// 获取 limit 参数
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 5 // 默认获取5个随机题目
	}

	questions, err := h.QuizService.GetRandomQuestions(uint(bankID), limit)
	if err != nil {
		Error(w, "Failed to retrieve random questions", http.StatusInternalServerError)
		return
	}

	var questionResponses []dto.QuestionResponse
	for _, q := range questions {
		questionResponse := dto.QuestionResponse{
			ID:             q.ID,
			QuestionBankID: q.QuestionBankID,
			QuestionType:   q.QuestionType,
			Content:        q.Content,
			CreatedAt:      q.CreatedAt,
			AuthorID:       q.AuthorID,
		}
		switch q.QuestionType {
		case models.QuestionTypeSingleChoice, models.QuestionTypeMultipleChoice:
			for _, option := range q.AnswerOptions {
				questionResponse.AnswerOptions = append(questionResponse.AnswerOptions, dto.AnswerOption{
					ID:         option.ID,
					OptionText: option.OptionText,
				})
			}
		case models.QuestionTypeFillInTheBlank:
			for range q.FillInTheBlanks {
				questionResponse.FillInTheBlanks = append(questionResponse.FillInTheBlanks, dto.FillInTheBlankAnswer{
					BlankText: "",
				})
			}
		}
		questionResponses = append(questionResponses, questionResponse)
	}

	Success(w, questionResponses, nil, http.StatusOK)
}

// GetQuestionDetail 获取问题详情
// @Summary 获取问题详情
// @Description 获取指定问题的详细信息，包括答案和标签
// @Tags Question
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param id path int true "问题 ID"
// @Success 200 {object} Response[dto.QuestionResponse] "问题详情"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /quiz/questions/{id} [get]
func (h *QuizHandler) GetQuestionDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	questionIDStr := vars["id"]

	questionID, err := strconv.ParseUint(questionIDStr, 10, 32)
	if err != nil {
		Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	// 从 QuizService 获取问题详细信息（包括答案和标签）
	question, err := h.QuizService.GetQuestionDetail(uint(questionID))
	if err != nil {
		Error(w, "Failed to retrieve question details", http.StatusInternalServerError)
		return
	}

	// 构建 dto.QuestionResponse（包括答案和标签）
	questionResponse := dto.QuestionResponse{
		ID:             question.ID,
		QuestionBankID: question.QuestionBankID,
		QuestionType:   question.QuestionType,
		Content:        question.Content,
		Explanation:    question.Explanation,
		CreatedAt:      question.CreatedAt,
		AuthorID:       question.AuthorID,
		AuthorName:     question.Author.Username,
	}

	// 填充标签
	for _, tag := range question.Tags {
		questionResponse.Tags = append(questionResponse.Tags, tag.Name)
	}

	// 根据题目类型填充不同的答案选项
	switch question.QuestionType {
	case models.QuestionTypeSingleChoice, models.QuestionTypeMultipleChoice:
		for _, option := range question.AnswerOptions {
			questionResponse.AnswerOptions = append(questionResponse.AnswerOptions, dto.AnswerOption{
				ID:         option.ID,
				OptionText: option.OptionText,
				IsCorrect:  option.IsCorrect,
			})
		}
	case models.QuestionTypeTrueFalse:
		if question.TrueFalseAnswer != nil {
			questionResponse.TrueFalseAnswer = &dto.TrueFalseAnswer{
				IsTrue: question.TrueFalseAnswer.IsTrue,
			}
		}
	case models.QuestionTypeWrittenAnswer:
		if question.WrittenAnswer != nil {
			questionResponse.WrittenAnswer = &dto.WrittenAnswer{
				AnswerText: question.WrittenAnswer.AnswerText,
			}
		}
	case models.QuestionTypeFillInTheBlank:
		for _, blank := range question.FillInTheBlanks {
			questionResponse.FillInTheBlanks = append(questionResponse.FillInTheBlanks, dto.FillInTheBlankAnswer{
				BlankText: blank.BlankText,
			})
		}
	}

	Success(w, questionResponse, nil, http.StatusOK)
}

// CreateQuestion 创建新的问题
// @Summary 创建问题
// @Description 创建一个新的问题
// @Tags Question
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param id path int true "题库 ID"
// @Param question body dto.CreateQuestionRequest true "创建问题请求"
// @Success 201 {object} Response[dto.QuestionResponse] "创建成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /quiz/question_banks/{id}/questions [post]
func (h *QuizHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bankID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		Error(w, "Invalid bank ID", http.StatusBadRequest)
		return
	}

	var req dto.CreateQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	question := models.Question{
		QuestionBankID: uint(bankID),
		Content:        req.Content,
		QuestionType:   req.QuestionType,
		Explanation:    req.Explanation,
		AuthorID:       req.AuthorID,
	}

	// 根据题目类型处理答案
	switch req.QuestionType {
	case models.QuestionTypeMultipleChoice, models.QuestionTypeSingleChoice:
		var options []models.AnswerOption
		for _, option := range req.AnswerOptions {
			options = append(options, models.AnswerOption{
				OptionText: option.OptionText,
				IsCorrect:  option.IsCorrect,
			})
		}
		question.AnswerOptions = options

	case models.QuestionTypeTrueFalse:
		question.TrueFalseAnswer = &models.TrueFalseAnswer{
			IsTrue: *req.TrueFalse,
		}

	case models.QuestionTypeWrittenAnswer:
		question.WrittenAnswer = &models.WrittenAnswer{
			AnswerText: req.AnswerText,
		}

	case models.QuestionTypeFillInTheBlank:
		var blanks []models.FillInTheBlankAnswer
		for _, blank := range req.Blanks {
			blanks = append(blanks, models.FillInTheBlankAnswer{
				BlankText: blank.BlankText,
			})
		}
		question.FillInTheBlanks = blanks
	}

	// 处理标签
	for _, tagName := range req.Tags {
		tag := models.Tag{Name: tagName}
		question.Tags = append(question.Tags, tag)
	}

	createdQuestion, err := h.QuizService.CreateQuestion(question)
	if err != nil {
		Error(w, "Failed to create question", http.StatusInternalServerError)
		return
	}

	response := dto.QuestionResponse{
		ID:             createdQuestion.ID,
		QuestionBankID: createdQuestion.QuestionBankID,
		QuestionType:   createdQuestion.QuestionType,
		Content:        createdQuestion.Content,
		Explanation:    createdQuestion.Explanation,
		AuthorID:       createdQuestion.AuthorID,
		AuthorName:     createdQuestion.Author.Username,
		CreatedAt:      createdQuestion.CreatedAt,
	}
	switch createdQuestion.QuestionType {
	case models.QuestionTypeMultipleChoice, models.QuestionTypeSingleChoice:
		for _, option := range createdQuestion.AnswerOptions {
			response.AnswerOptions = append(response.AnswerOptions, dto.AnswerOption{
				ID:         option.ID,
				OptionText: option.OptionText,
				IsCorrect:  option.IsCorrect,
			})
		}
	case models.QuestionTypeTrueFalse:
		response.TrueFalseAnswer = &dto.TrueFalseAnswer{
			IsTrue: createdQuestion.TrueFalseAnswer.IsTrue,
		}
	case models.QuestionTypeWrittenAnswer:
		response.WrittenAnswer = &dto.WrittenAnswer{
			AnswerText: createdQuestion.WrittenAnswer.AnswerText,
		}
	case models.QuestionTypeFillInTheBlank:
		for _, blank := range createdQuestion.FillInTheBlanks {
			response.FillInTheBlanks = append(response.FillInTheBlanks, dto.FillInTheBlankAnswer{
				BlankText: blank.BlankText,
			})
		}
	}
	for _, tag := range createdQuestion.Tags {
		response.Tags = append(response.Tags, tag.Name)
	}

	Success(w, response, nil, http.StatusCreated)
}

// UpdateQuestion 更新问题
// @Summary 更新问题
// @Description 编辑指定问题
// @Tags Question
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param id path int true "问题 ID"
// @Param question body dto.UpdateQuestionRequest true "更新问题请求"
// @Success 200 {object} Response[dto.QuestionResponse] "更新成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /quiz/questions/{id} [put]
func (h *QuizHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	questionID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	var req dto.UpdateQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	question := models.Question{
		ID:             uint(questionID),
		QuestionBankID: req.QuestionBankID,
		Content:        req.Content,
		QuestionType:   req.QuestionType,
		Explanation:    req.Explanation,
		AuthorID:       req.AuthorID,
	}

	// 根据题目类型处理答案更新
	switch req.QuestionType {
	case models.QuestionTypeMultipleChoice, models.QuestionTypeSingleChoice:
		var options []models.AnswerOption
		for _, option := range req.AnswerOptions {
			options = append(options, models.AnswerOption{
				OptionText: option.OptionText,
				IsCorrect:  option.IsCorrect,
			})
		}
		question.AnswerOptions = options

	case models.QuestionTypeTrueFalse:
		question.TrueFalseAnswer = &models.TrueFalseAnswer{
			IsTrue: *req.TrueFalse,
		}

	case models.QuestionTypeWrittenAnswer:
		question.WrittenAnswer = &models.WrittenAnswer{
			AnswerText: req.AnswerText,
		}

	case models.QuestionTypeFillInTheBlank:
		var blanks []models.FillInTheBlankAnswer
		for _, blank := range req.Blanks {
			blanks = append(blanks, models.FillInTheBlankAnswer{
				BlankText: blank.BlankText,
			})
		}
		question.FillInTheBlanks = blanks
	}

	// 处理标签
	var tags []models.Tag
	for _, tagName := range req.Tags {
		tag := models.Tag{Name: tagName}
		tags = append(tags, tag)
	}
	question.Tags = tags

	updatedQuestion, err := h.QuizService.UpdateQuestion(question)
	if err != nil {
		Error(w, "Failed to update question", http.StatusInternalServerError)
		return
	}

	response := dto.QuestionResponse{
		ID:             updatedQuestion.ID,
		QuestionBankID: updatedQuestion.QuestionBankID,
		QuestionType:   updatedQuestion.QuestionType,
		Content:        updatedQuestion.Content,
		Explanation:    updatedQuestion.Explanation,
		AuthorID:       updatedQuestion.AuthorID,
		AuthorName:     updatedQuestion.Author.Username,
		CreatedAt:      updatedQuestion.CreatedAt,
	}

	Success(w, response, nil, http.StatusOK)
}

// DeleteQuestion 删除问题
// @Summary 删除问题
// @Description 删除指定问题
// @Tags Question
// @Security ApiKeyAuth
// @Produce  json
// @Param id path int true "问题 ID"
// @Success 204 "删除成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /quiz/questions/{id} [delete]
func (h *QuizHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	questionIDStr := vars["id"]

	questionID, err := strconv.ParseUint(questionIDStr, 10, 32)
	if err != nil {
		Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	if err := h.QuizService.DeleteQuestion(uint(questionID)); err != nil {
		Error(w, "Failed to delete question", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RecordQuestionAttempt 记录用户的答题尝试
// @Summary 记录用户的答题尝试
// @Description 记录用户对特定问题的答题情况
// @Tags QuestionAttempt
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param input body dto.QuestionAttemptRequest true "答题尝试信息"
// @Success 200 {object} Response[dto.QuestionAttemptResponse] "记录成功"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /quiz/question_attempts [post]
func (h *QuizHandler) RecordQuestionAttempt(w http.ResponseWriter, r *http.Request) {
	var req dto.QuestionAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	attempt, err := h.QuizService.RecordQuestionAttempt(req.UserID, req.QuestionID, req.Answer)
	if err != nil {
		Error(w, "Failed to record question attempt", http.StatusInternalServerError)
		return
	}

	response := dto.QuestionAttemptResponse{
		QuestionID:         attempt.QuestionID,
		Attempts:           attempt.Attempts,
		Wrong:              attempt.Wrong,
		ConsecutiveCorrect: attempt.ConsecutiveCorrect,
		LastAnswerAt:       attempt.LastAnswerAt,
	}

	Success(w, response, nil, http.StatusOK)
}

// GetQuestionAttempts 获取用户的答题尝试情况
// @Summary 获取用户的答题尝试情况
// @Description 获取用户在特定题库中的答题情况
// @Tags QuestionAttempt
// @Security ApiKeyAuth
// @Produce  json
// @Param user_id path int true "用户 ID"
// @Param question_bank_id path int true "题库 ID"
// @Success 200 {object} Response[[]dto.QuestionAttemptResponse] "答题尝试列表"
// @Failure 400 {object} ErrorResponse "无效请求"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /quiz/question_attempts/{user_id}/{question_bank_id} [get]
func (h *QuizHandler) GetQuestionAttempts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["user_id"]
	questionBankIDStr := vars["question_bank_id"]

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	questionBankID, err := strconv.ParseUint(questionBankIDStr, 10, 32)
	if err != nil {
		Error(w, "Invalid question bank ID", http.StatusBadRequest)
		return
	}

	attempts, err := h.QuizService.GetQuestionAttempts(uint(userID), uint(questionBankID), 3) // 假设3为连续正确的阈值
	if err != nil {
		Error(w, "Failed to get question attempts", http.StatusInternalServerError)
		return
	}

	Success(w, attempts, nil, http.StatusOK)
}
