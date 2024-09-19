package models

type QuestionBank struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

type QuestionType struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

type Question struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	QuestionBankID uint           `json:"question_bank_id"`
	QuestionTypeID uint           `json:"question_type_id"`
	Content        string         `gorm:"not null" json:"content"`
	Explanation    string         `json:"explanation"`
	AnswerOptions  []AnswerOption `json:"answer_options" gorm:"foreignKey:QuestionID"`
}

type AnswerOption struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	QuestionID uint   `json:"question_id"`
	OptionText string `gorm:"not null" json:"option_text"`
	IsCorrect  bool   `json:"is_correct"`
}

type RelatedQuestion struct {
	QuestionID        uint `gorm:"primaryKey" json:"question_id"`
	RelatedQuestionID uint `gorm:"primaryKey" json:"related_question_id"`
}
