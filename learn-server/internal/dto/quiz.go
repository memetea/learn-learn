// dto/quiz.go
package dto

import (
	"learn/internal/models"
	"time"
)

// CreateQuestionBankRequest 定义了创建题库请求的结构体
type CreateQuestionBankRequest struct {
	Name string `json:"name" validate:"required"`
}

// CreateQuestionRequest 定义了创建问题请求的结构体
type CreateQuestionRequest struct {
	Content       string                 `json:"content" validate:"required"`
	QuestionType  models.QuestionType    `json:"question_type" validate:"required"`
	Explanation   string                 `json:"explanation,omitempty"`
	AnswerOptions []AnswerOption         `json:"answer_options,omitempty"` // 仅选择题使用
	TrueFalse     *bool                  `json:"true_false,omitempty"`     // 判断题使用
	AnswerText    string                 `json:"answer_text,omitempty"`    // 问答题使用
	Blanks        []FillInTheBlankAnswer `json:"blanks,omitempty"`         // 填空题使用
	Tags          []string               `json:"tags,omitempty"`           // 标签列表
	AuthorID      uint                   `json:"author_id"`                // 问题的作者 ID
}

// UpdateQuestionRequest 用于更新问题的请求体
type UpdateQuestionRequest struct {
	QuestionBankID uint                   `json:"question_bank_id" validate:"required"`
	Content        string                 `json:"content" validate:"required"`
	QuestionType   models.QuestionType    `json:"question_type" validate:"required"`
	Explanation    string                 `json:"explanation,omitempty"`
	AnswerOptions  []AnswerOption         `json:"answer_options,omitempty"` // 仅选择题使用
	TrueFalse      *bool                  `json:"true_false,omitempty"`     // 判断题使用
	AnswerText     string                 `json:"answer_text,omitempty"`    // 问答题使用
	Blanks         []FillInTheBlankAnswer `json:"blanks,omitempty"`         // 填空题使用
	Tags           []string               `json:"tags,omitempty"`           // 标签列表
	AuthorID       uint                   `json:"author_id"`                // 问题的作者 ID
}

// QuestionBankResponse 用于返回题库的信息
type QuestionBankResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// QuestionResponse 用于返回问题的信息
type QuestionResponse struct {
	ID              uint                   `json:"id"`
	QuestionBankID  uint                   `json:"question_bank_id"`
	QuestionType    models.QuestionType    `json:"question_type"`
	Content         string                 `json:"content"`
	Explanation     string                 `json:"explanation,omitempty"`
	AnswerOptions   []AnswerOption         `json:"answer_options,omitempty"`
	TrueFalseAnswer *TrueFalseAnswer       `json:"true_false_answer,omitempty"`
	WrittenAnswer   *WrittenAnswer         `json:"written_answer,omitempty"`
	FillInTheBlanks []FillInTheBlankAnswer `json:"fill_in_the_blanks,omitempty"`
	Tags            []string               `json:"tags,omitempty"` // 返回标签
	AuthorID        uint                   `json:"author_id"`
	AuthorName      string                 `json:"author_name"` // 用户名
	CreatedAt       time.Time              `json:"created_at"`
}

// AnswerOption 表示选择题或多选题的选项
type AnswerOption struct {
	ID         uint   `json:"id,omitempty"`
	OptionText string `json:"option_text" validate:"required"`
	IsCorrect  bool   `json:"is_correct"`
}

// TrueFalseAnswer 表示判断题的答案
type TrueFalseAnswer struct {
	IsTrue bool `json:"is_true"`
}

// WrittenAnswer 表示问答题的答案
type WrittenAnswer struct {
	AnswerText string `json:"answer_text" validate:"required"`
}

// FillInTheBlankAnswer 表示填空题的填空项
type FillInTheBlankAnswer struct {
	BlankText string `json:"blank_text" validate:"required"`
}

type QuestionAttemptRequest struct {
	UserID     uint        `json:"user_id"`
	QuestionID uint        `json:"question_id"`
	Answer     interface{} `json:"answer"` // Stores the user's answer, can be string, []string, bool, etc.
}

type QuestionAttemptResponse struct {
	QuestionID         uint      `json:"question_id"`
	Attempts           uint      `json:"attempts"`
	Wrong              uint      `json:"wrong"`
	ConsecutiveCorrect uint      `json:"consecutive_correct"`
	LastAnswerAt       time.Time `json:"last_answer_at"`
}
