// services/quiz_service.go
package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"learn/internal/dto"
	"learn/internal/models"
	"time"

	"gorm.io/gorm"
)

type QuizService struct {
	db *gorm.DB
}

func NewQuizService(db *gorm.DB) *QuizService {
	return &QuizService{db: db}
}

// GetQuestionBanks returns all question banks
func (s *QuizService) GetQuestionBanks() ([]models.QuestionBank, error) {
	var questionBanks []models.QuestionBank
	if err := s.db.Find(&questionBanks).Error; err != nil {
		return nil, err
	}
	return questionBanks, nil
}

// CreateQuestionBank creates a new question bank
func (s *QuizService) CreateQuestionBank(name string) (*models.QuestionBank, error) {
	questionBank := models.QuestionBank{Name: name}
	if err := s.db.Create(&questionBank).Error; err != nil {
		return nil, err
	}
	return &questionBank, nil
}

// CreateQuestion creates a new question with associated tags and answers based on question type
// services/quiz_service.go
func (s *QuizService) CreateQuestion(question models.Question) (*models.Question, error) {
	tx := s.db.Begin()

	// 处理标签的创建或关联
	for i, tag := range question.Tags {
		var existingTag models.Tag
		// 查找是否存在相同名称的标签
		if err := tx.Where("name = ?", tag.Name).First(&existingTag).Error; err != nil {
			// 如果不存在，则创建新标签
			if err := tx.Create(&question.Tags[i]).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create new tag: %w", err)
			}
		} else {
			question.Tags[i] = existingTag
		}
	}

	// 尝试创建问题
	if err := tx.Create(&question).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create question: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &question, nil
}

// UpdateQuestion updates an existing question and its related answers and tags based on question type
func (s *QuizService) UpdateQuestion(question models.Question) (*models.Question, error) {
	tx := s.db.Begin()

	// 处理标签的创建或关联
	for i, tag := range question.Tags {
		var existingTag models.Tag
		// 查找是否存在相同名称的标签
		if err := tx.Where("name = ?", tag.Name).First(&existingTag).Error; err != nil {
			// 如果不存在，则创建新标签
			if err := tx.Create(&question.Tags[i]).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create new tag: %w", err)
			}
		} else {
			question.Tags[i] = existingTag
		}
	}

	// 更新问题内容和解释
	if err := tx.Omit("CreatedAt").Save(&question).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 根据问题类型处理答案更新
	switch question.QuestionType {
	case models.QuestionTypeMultipleChoice, models.QuestionTypeSingleChoice:
		// 删除旧的选项并插入新的选项
		if err := tx.Where("question_id = ?", question.ID).Delete(&models.AnswerOption{}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		for i := range question.AnswerOptions {
			question.AnswerOptions[i].QuestionID = question.ID
			if err := tx.Create(&question.AnswerOptions[i]).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}

	case models.QuestionTypeTrueFalse:
		// 更新或创建判断题答案
		if question.TrueFalseAnswer != nil {
			if err := tx.Save(question.TrueFalseAnswer).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}

	case models.QuestionTypeWrittenAnswer:
		// 更新或创建问答题答案
		if question.WrittenAnswer != nil {
			if err := tx.Save(question.WrittenAnswer).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}

	case models.QuestionTypeFillInTheBlank:
		// 删除旧的填空答案并插入新的填空答案
		if err := tx.Where("question_id = ?", question.ID).Delete(&models.FillInTheBlankAnswer{}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		for i := range question.FillInTheBlanks {
			question.FillInTheBlanks[i].QuestionID = question.ID
			if err := tx.Create(&question.FillInTheBlanks[i]).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	// 更新标签
	if err := tx.Model(&question).Association("Tags").Replace(question.Tags); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update tags: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &question, nil
}

// DeleteQuestion deletes an existing question and its related answers based on the question type
func (s *QuizService) DeleteQuestion(questionID uint) error {
	tx := s.db.Begin()

	var question models.Question
	if err := tx.First(&question, questionID).Error; err != nil {
		tx.Rollback()
		return errors.New("question not found")
	}

	// 根据类型删除相关答案
	switch question.QuestionType {
	case models.QuestionTypeMultipleChoice, models.QuestionTypeSingleChoice:
		if err := tx.Where("question_id = ?", questionID).Delete(&models.AnswerOption{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	case models.QuestionTypeTrueFalse:
		if err := tx.Where("question_id = ?", questionID).Delete(&models.TrueFalseAnswer{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	case models.QuestionTypeWrittenAnswer:
		if err := tx.Where("question_id = ?", questionID).Delete(&models.WrittenAnswer{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	case models.QuestionTypeFillInTheBlank:
		if err := tx.Where("question_id = ?", questionID).Delete(&models.FillInTheBlankAnswer{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 删除问题本身
	if err := tx.Delete(&models.Question{}, questionID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetQuestionDetail 获取问题详细信息（包括答案）
func (s *QuizService) GetQuestionDetail(questionID uint) (*models.Question, error) {
	var question models.Question

	// 查询问题基本信息
	if err := s.db.First(&question, questionID).Error; err != nil {
		return nil, err
	}

	// 根据问题类型预加载相关的答案表
	switch question.QuestionType {
	case models.QuestionTypeMultipleChoice, models.QuestionTypeSingleChoice:
		if err := s.db.Preload("AnswerOptions").First(&question, questionID).Error; err != nil {
			return nil, err
		}
	case models.QuestionTypeTrueFalse:
		if err := s.db.Preload("TrueFalseAnswer").First(&question, questionID).Error; err != nil {
			return nil, err
		}
	case models.QuestionTypeWrittenAnswer:
		if err := s.db.Preload("WrittenAnswer").First(&question, questionID).Error; err != nil {
			return nil, err
		}
	case models.QuestionTypeFillInTheBlank:
		if err := s.db.Preload("FillInTheBlanks").First(&question, questionID).Error; err != nil {
			return nil, err
		}
	}

	// 预加载标签
	if err := s.db.Model(&question).Association("Tags").Find(&question.Tags); err != nil {
		return nil, err
	}

	return &question, nil
}

// GetQuestions retrieves all questions from a specific question bank
func (s *QuizService) GetQuestions(questionBankID uint, tag string) ([]models.Question, error) {
	var questions []models.Question

	// 通过标签进行问题过滤
	if tag != "" {
		err := s.db.Joins("JOIN question_tags qt ON qt.question_id = questions.id").
			Joins("JOIN tags t ON qt.tag_id = t.id").
			Where("questions.question_bank_id = ? AND t.name = ?", questionBankID, tag).
			Find(&questions).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := s.db.Where("question_bank_id = ?", questionBankID).Find(&questions).Error
		if err != nil {
			return nil, err
		}
	}

	return questions, nil
}

// GetQuestionsWithPagination retrieves paginated questions from a specific question bank with optional tag filtering
func (s *QuizService) GetQuestionsWithPagination(questionBankID uint, tag string, page int, pageSize int) ([]models.Question, int64, error) {
	var questions []models.Question
	var total int64

	// 计算总记录数
	query := s.db.Model(&models.Question{}).Where("question_bank_id = ?", questionBankID)
	if tag != "" {
		query = query.Joins("JOIN question_tags qt ON qt.question_id = questions.id").
			Joins("JOIN tags t ON t.id = qt.tag_id").Where("t.name = ?", tag)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取指定页的数据
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&questions).Error; err != nil {
		return nil, 0, err
	}

	return questions, total, nil
}

// GetRandomQuestions retrieves a set of random questions from the specified question bank
func (s *QuizService) GetRandomQuestions(questionBankID uint, limit int) ([]models.Question, error) {
	var questions []models.Question

	// 使用随机函数获取指定数量的随机问题
	if err := s.db.Where("question_bank_id = ?", questionBankID).
		Order("RANDOM()").Limit(limit).Find(&questions).Error; err != nil {
		return nil, err
	}

	// 根据问题类型有选择地预加载对应的答案关联数据
	for i := range questions {
		switch questions[i].QuestionType {
		case models.QuestionTypeSingleChoice, models.QuestionTypeMultipleChoice:
			// 预加载选择题的选项
			if err := s.db.Preload("AnswerOptions").Find(&questions[i]).Error; err != nil {
				return nil, err
			}
		case models.QuestionTypeTrueFalse:
			// 预加载判断题的答案
			if err := s.db.Preload("TrueFalseAnswer").Find(&questions[i]).Error; err != nil {
				return nil, err
			}
		case models.QuestionTypeWrittenAnswer:
			// 预加载问答题的答案
			if err := s.db.Preload("WrittenAnswer").Find(&questions[i]).Error; err != nil {
				return nil, err
			}
		case models.QuestionTypeFillInTheBlank:
			// 预加载填空题的答案
			if err := s.db.Preload("FillInTheBlanks").Find(&questions[i]).Error; err != nil {
				return nil, err
			}
		}
	}

	return questions, nil
}

// GetQuestionAttempt retrieves a specific question attempt by user and question ID
func (s *QuizService) GetQuestionAttempt(userID uint, questionID uint) (*models.QuestionAttempt, error) {
	var attempt models.QuestionAttempt
	if err := s.db.Where("user_id = ? AND question_id = ?", userID, questionID).First(&attempt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if no attempt record found
		}
		return nil, err
	}
	return &attempt, nil
}

func (s *QuizService) RecordQuestionAttempt(userID uint, questionID uint, answer interface{}) (*models.QuestionAttempt, error) {
	var question models.Question
	if err := s.db.First(&question, "id = ?", questionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("question not found")
		}
		return nil, err
	}

	isCorrect := false

	// Preload relevant associations based on question type
	switch question.QuestionType {
	case models.QuestionTypeSingleChoice, models.QuestionTypeMultipleChoice:
		if err := s.db.Preload("AnswerOptions").First(&question).Error; err != nil {
			return nil, err
		}
	case models.QuestionTypeTrueFalse:
		if err := s.db.Preload("TrueFalseAnswer").First(&question).Error; err != nil {
			return nil, err
		}
	case models.QuestionTypeWrittenAnswer:
		if err := s.db.Preload("WrittenAnswer").First(&question).Error; err != nil {
			return nil, err
		}
	case models.QuestionTypeFillInTheBlank:
		if err := s.db.Preload("FillInTheBlanks").First(&question).Error; err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown question type")
	}

	// Verify the answer based on question type
	var lastAnswerJSON []byte
	switch question.QuestionType {
	case models.QuestionTypeSingleChoice, models.QuestionTypeMultipleChoice:
		// For multiple choice questions, compare the answer options
		providedAnswersFloat, ok := answer.([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid answer format for multiple choice question")
		}

		// Convert []interface{} (float64) to []uint
		providedAnswers := make([]uint, len(providedAnswersFloat))
		for i, val := range providedAnswersFloat {
			floatVal, ok := val.(float64)
			if !ok {
				return nil, fmt.Errorf("invalid answer format for multiple choice question")
			}
			providedAnswers[i] = uint(floatVal)
		}

		correctAnswers := []uint{}
		for _, option := range question.AnswerOptions {
			if option.IsCorrect {
				correctAnswers = append(correctAnswers, option.ID)
			}
		}

		isCorrect = compareAnswers(providedAnswers, correctAnswers)

		// Record the provided answer
		answerJSON, err := json.Marshal(providedAnswers)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal answer: %v", err)
		}
		lastAnswerJSON = answerJSON

	case models.QuestionTypeTrueFalse:
		// For true/false questions, compare the boolean value
		isCorrect = false
		if answer != nil {
			providedAnswer, ok := answer.(bool)
			if !ok {
				return nil, fmt.Errorf("invalid answer format for true/false question")
			}
			isCorrect = (providedAnswer == question.TrueFalseAnswer.IsTrue)
			// Record the provided answer
			answerJSON, err := json.Marshal(providedAnswer)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal answer: %v", err)
			}
			lastAnswerJSON = answerJSON
		}

	case models.QuestionTypeWrittenAnswer:
		// For written questions, simply record the answer for manual grading
		isCorrect = true
		if answer != nil {
			providedAnswer, ok := answer.(string)
			if !ok {
				return nil, fmt.Errorf("invalid answer format for written answer question")
			}
			// Record the provided answer
			answerJSON, err := json.Marshal(providedAnswer)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal answer: %v", err)
			}
			lastAnswerJSON = answerJSON
		}

	case models.QuestionTypeFillInTheBlank:
		// For fill-in-the-blank questions, compare each blank answer
		isCorrect = false
		if answer != nil {
			providedAnswersInterface, ok := answer.([]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answer format for fill-in-the-blank question")
			}

			// Convert []interface{} to []string
			providedAnswers := make([]string, len(providedAnswersInterface))
			for i, val := range providedAnswersInterface {
				strVal, ok := val.(string)
				if !ok {
					return nil, fmt.Errorf("invalid answer format for fill-in-the-blank question")
				}
				providedAnswers[i] = strVal
			}

			if len(providedAnswers) != len(question.FillInTheBlanks) {
				isCorrect = false
			} else {
				isCorrect = true
				for i, blank := range question.FillInTheBlanks {
					if providedAnswers[i] != blank.BlankText {
						isCorrect = false
						break
					}
				}
			}

			// Record the provided answers
			answerJSON, err := json.Marshal(providedAnswers)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal answers: %v", err)
			}
			lastAnswerJSON = answerJSON
		}

	default:
		return nil, fmt.Errorf("unknown question type")
	}

	// Record the attempt
	var attempt models.QuestionAttempt
	err := s.db.Where("user_id = ? AND question_id = ?", userID, questionID).First(&attempt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create a new attempt record if not found
			attempt = models.QuestionAttempt{
				UserID:             userID,
				QuestionID:         questionID,
				Attempts:           1,
				ConsecutiveCorrect: 0,
				LastAnswerAt:       time.Now(),
				LastAnswer:         lastAnswerJSON,
			}
			if isCorrect {
				attempt.ConsecutiveCorrect = 1
			}
			return &attempt, s.db.Create(&attempt).Error
		}
		return nil, err
	}

	attempt.UpdateAnswer(lastAnswerJSON, isCorrect)
	if err := s.db.Save(&attempt).Error; err != nil {
		return nil, err
	}
	return &attempt, nil
}

func compareAnswers(providedAnswers, correctAnswers []uint) bool {
	if len(providedAnswers) != len(correctAnswers) {
		return false
	}
	answerMap := make(map[uint]bool)
	for _, answer := range correctAnswers {
		answerMap[answer] = true
	}
	for _, answer := range providedAnswers {
		if !answerMap[answer] {
			return false
		}
	}
	return true
}

// GetQuestionBankErrorFrequency retrieves the error frequency for questions in a question bank since a given time
func (s *QuizService) GetQuestionAttempts(userID uint, questionBankID uint, consecutiveCorrectThreshold uint) ([]dto.QuestionAttemptResponse, error) {
	var attempts []models.QuestionAttempt
	if err := s.db.Joins("JOIN questions ON questions.id = question_attempts.question_id").
		Where("questions.question_bank_id = ? AND question_attempts.consecutive_correct < ?",
			questionBankID, consecutiveCorrectThreshold).
		Find(&attempts).Error; err != nil {
		return nil, err
	}

	var result []dto.QuestionAttemptResponse
	for _, attempt := range attempts {
		result = append(result, dto.QuestionAttemptResponse{
			QuestionID:         attempt.QuestionID,
			Attempts:           attempt.Attempts,
			Wrong:              attempt.Wrong,
			ConsecutiveCorrect: attempt.ConsecutiveCorrect,
			LastAnswerAt:       attempt.LastAnswerAt,
		})
	}

	return result, nil
}
