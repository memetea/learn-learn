// models/quiz.go
package models

import (
	"encoding/json"
	"time"
)

type QuestionType int

const (
	QuestionTypeSingleChoice QuestionType = iota
	QuestionTypeMultipleChoice
	QuestionTypeTrueFalse
	QuestionTypeWrittenAnswer
	QuestionTypeFillInTheBlank // 新增填空题类型
)

func (q QuestionType) String() string {
	switch q {
	case QuestionTypeSingleChoice:
		return "单选题"
	case QuestionTypeMultipleChoice:
		return "多选题"
	case QuestionTypeTrueFalse:
		return "判断题"
	case QuestionTypeWrittenAnswer:
		return "问答题"
	case QuestionTypeFillInTheBlank:
		return "填空题"
	}
	return ""
}

type QuestionBank struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

type Question struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	QuestionBankID uint         `json:"question_bank_id"`
	QuestionType   QuestionType `json:"question_type"` // 题目类型：选择题、判断题、问答题、填空题等
	Content        string       `gorm:"not null" json:"content"`
	Explanation    string       `json:"explanation"`
	AuthorID       uint         `json:"author_id"` // 用户ID，关联到用户表
	CreatedAt      time.Time    `json:"created_at" gorm:"autoCreateTime"`
	AutoGenerated  bool         `gorm:"default:false" json:"auto_generated"`

	// 定义关联
	Author          User                   `gorm:"foreignKey:AuthorID" json:"author"` // 使用外键关联用户表
	AnswerOptions   []AnswerOption         `json:"answer_options,omitempty" gorm:"foreignKey:QuestionID"`
	TrueFalseAnswer *TrueFalseAnswer       `json:"true_false_answer,omitempty" gorm:"foreignKey:QuestionID"`
	WrittenAnswer   *WrittenAnswer         `json:"written_answer,omitempty" gorm:"foreignKey:QuestionID"`
	FillInTheBlanks []FillInTheBlankAnswer `json:"fill_in_the_blanks,omitempty" gorm:"foreignKey:QuestionID"` // 新增填空题关联
	Tags            []Tag                  `json:"tags,omitempty" gorm:"many2many:question_tags;"`            // 关联标签
}

// 选择题和多选题的选项存储
type AnswerOption struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	QuestionID uint   `json:"question_id"`
	OptionText string `gorm:"not null" json:"option_text"` // 选项内容
	IsCorrect  bool   `json:"is_correct"`                  // 是否是正确答案
}

// 判断题的答案，只有 true 或 false
type TrueFalseAnswer struct {
	QuestionID uint `gorm:"primaryKey" json:"question_id"`
	IsTrue     bool `json:"is_true"` // 正确答案是 true 或 false
}

// 问答题的答案
type WrittenAnswer struct {
	QuestionID uint   `gorm:"primaryKey" json:"question_id" `
	AnswerText string `gorm:"not null" json:"answer_text"` // 问答题答案
}

// 填空题的答案
type FillInTheBlankAnswer struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	QuestionID uint   `json:"question_id"`
	BlankText  string `gorm:"not null" json:"blank_text"` // 填空的正确答案
}

type Tag struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

type RelatedQuestion struct {
	QuestionID        uint `gorm:"primaryKey" json:"question_id"`
	RelatedQuestionID uint `gorm:"primaryKey" json:"related_question_id"`
}

type QuestionAttempt struct {
	ID                 uint            `gorm:"primaryKey" json:"id"`
	UserID             uint            `json:"user_id"`
	QuestionID         uint            `json:"question_id"`
	Attempts           uint            `json:"attempts"`
	Wrong              uint            `json:"wrong"`
	ConsecutiveCorrect uint            `json:"consecutive_correct"`
	LastAnswer         json.RawMessage `json:"last_answer"` // Used to store the last answer
	LastAnswerAt       time.Time       `json:"last_answer_at"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// Update logic example
func (qa *QuestionAttempt) UpdateAnswer(answer []byte, isCorrect bool) {
	qa.Attempts++
	qa.LastAnswer = answer
	if isCorrect {
		qa.ConsecutiveCorrect++
	} else {
		qa.Wrong++
		qa.ConsecutiveCorrect = 0
	}
	qa.LastAnswerAt = time.Now()
}
