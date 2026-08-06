package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

// Quiz question types.
const (
	QuestionMultipleChoice string = "multiple_choice"
	QuestionTrueFalse      string = "true_false"
)

// ErrTrainingNameNotSpecified is thrown when a training is submitted
// without a name.
var ErrTrainingNameNotSpecified = errors.New("Training name not specified")

// ErrTrainingNotFound is thrown when a training can't be found.
var ErrTrainingNotFound = errors.New("Training not found")

// Training is an admin-authored awareness training: an ordered set of
// lessons with an optional quiz at the end.
type Training struct {
	Id          int64            `json:"id"`
	TenantID    int64            `json:"-" gorm:"column:tenant_id;default:1"`
	Name        string           `json:"name" sql:"not null"`
	Description string           `json:"description"`
	CreatedBy   int64            `json:"-" gorm:"column:created_by"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Lessons     []TrainingLesson `json:"lessons,omitempty" gorm:"foreignkey:TrainingId"`
	Quiz        *TrainingQuiz    `json:"quiz,omitempty" gorm:"-"`
}

func (t *Training) TableName() string { return "trainings" }

// TrainingLesson is one page of content within a training, shown in
// Position order.
type TrainingLesson struct {
	Id         int64  `json:"id"`
	TrainingId int64  `json:"training_id"`
	Position   int    `json:"position"`
	Title      string `json:"title" sql:"not null"`
	HTML       string `json:"html"`
}

func (l *TrainingLesson) TableName() string { return "training_lessons" }

// TrainingQuiz is the (optional, at most one) quiz at the end of a
// training's lessons.
type TrainingQuiz struct {
	Id          int64          `json:"id"`
	TrainingId  int64          `json:"training_id"`
	PassPercent int            `json:"pass_percent"`
	MaxAttempts int            `json:"max_attempts"`
	Questions   []QuizQuestion `json:"questions,omitempty" gorm:"foreignkey:QuizId"`
}

func (q *TrainingQuiz) TableName() string { return "training_quizzes" }

// QuizQuestion is one question in a quiz. Type is independently chosen
// per question — a quiz can mix multiple-choice and true/false
// questions. Options is a JSON-encoded array of choices, used only for
// multiple_choice; true_false questions compare CorrectAnswer directly
// against "true"/"false".
type QuizQuestion struct {
	Id       int64  `json:"id"`
	QuizId   int64  `json:"quiz_id"`
	Position int    `json:"position"`
	Type     string `json:"type"`
	Text     string `json:"text" sql:"not null"`
	// Options/CorrectAnswer are plain JSON fields, not "-": the admin
	// authoring API (controllers/api/training.go) needs to read and write
	// them same as every other field. The public quiz-taking template
	// (templates/training_quiz.html) never has a JSON boundary to leak
	// through — it accesses these as Go struct fields inside
	// html/template, and only ever renders OptionsList(), never
	// CorrectAnswer.
	Options       string `json:"options" gorm:"column:options"`
	CorrectAnswer string `json:"correct_answer" gorm:"column:correct_answer"`
}

func (q *QuizQuestion) TableName() string { return "quiz_questions" }

// OptionsList decodes Options into a slice, empty for true_false
// questions or malformed data. Value receiver, not pointer: called from
// html/template on a range variable (models.QuizQuestion, not
// *QuizQuestion), which isn't addressable — a pointer-receiver method
// there fails at template execution time, silently, because none of this
// package's Execute() call sites check its returned error (matching this
// file's existing pattern elsewhere).
func (q QuizQuestion) OptionsList() []string {
	if q.Options == "" {
		return nil
	}
	var opts []string
	if err := json.Unmarshal([]byte(q.Options), &opts); err != nil {
		return nil
	}
	return opts
}

// Validate checks a training has the minimum required fields.
func (t *Training) Validate() error {
	if t.Name == "" {
		return ErrTrainingNameNotSpecified
	}
	return nil
}

// PostTrainingForTenant creates a new training scoped to the tenant.
func PostTrainingForTenant(t *Training, tenantID, uid int64) error {
	if err := t.Validate(); err != nil {
		return err
	}
	t.TenantID = tenantID
	t.CreatedBy = uid
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now
	return db.Save(t).Error
}

// PutTrainingForTenant updates a training's name/description, verifying
// tenant ownership first.
func PutTrainingForTenant(t *Training, tenantID int64) error {
	if err := t.Validate(); err != nil {
		return err
	}
	existing, err := GetTrainingForTenant(t.Id, tenantID)
	if err != nil {
		return err
	}
	existing.Name = t.Name
	existing.Description = t.Description
	existing.UpdatedAt = time.Now().UTC()
	return db.Save(&existing).Error
}

// GetTrainingsForTenant returns every training belonging to the tenant,
// with lessons and quiz (with its questions) preloaded.
func GetTrainingsForTenant(tenantID int64) ([]Training, error) {
	ts := []Training{}
	if err := db.Preload("Lessons", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("position asc")
	}).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&ts).Error; err != nil {
		return nil, err
	}
	for i := range ts {
		quiz, err := GetTrainingQuiz(ts[i].Id)
		if err == nil {
			ts[i].Quiz = &quiz
		}
	}
	return ts, nil
}

// GetTrainingForTenant returns a single tenant-scoped training, preloaded
// with its lessons and quiz.
func GetTrainingForTenant(id, tenantID int64) (Training, error) {
	t := Training{}
	err := db.Preload("Lessons", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("position asc")
	}).Where("id = ? AND tenant_id = ?", id, tenantID).First(&t).Error
	if err == gorm.ErrRecordNotFound {
		return t, ErrTrainingNotFound
	}
	if err != nil {
		return t, err
	}
	quiz, err := GetTrainingQuiz(t.Id)
	if err == nil {
		t.Quiz = &quiz
	}
	return t, nil
}

// GetTrainingByID returns a training by id with no tenant check — for
// call sites (like the public delivery routes) that already resolved
// tenant membership via the assignment's own tenant_id.
func GetTrainingByID(id int64) (Training, error) {
	t := Training{}
	err := db.Preload("Lessons", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("position asc")
	}).Where("id = ?", id).First(&t).Error
	if err == gorm.ErrRecordNotFound {
		return t, ErrTrainingNotFound
	}
	if err != nil {
		return t, err
	}
	quiz, err := GetTrainingQuiz(t.Id)
	if err == nil {
		t.Quiz = &quiz
	}
	return t, nil
}

// DeleteTrainingForTenant removes a training and its lessons/quiz.
func DeleteTrainingForTenant(id, tenantID int64) error {
	t, err := GetTrainingForTenant(id, tenantID)
	if err != nil {
		return err
	}
	if quiz, err := GetTrainingQuiz(t.Id); err == nil {
		db.Where("quiz_id = ?", quiz.Id).Delete(QuizQuestion{})
		db.Delete(&quiz)
	}
	db.Where("training_id = ?", t.Id).Delete(TrainingLesson{})
	return db.Delete(&t).Error
}

// AddTrainingLesson appends a lesson to a training. Position is computed
// server-side (count of existing lessons) rather than trusting the
// client's count of DOM rows at click time, which raced with the
// previous lesson's save and produced duplicate positions.
func AddTrainingLesson(l *TrainingLesson) error {
	var count int
	db.Model(&TrainingLesson{}).Where("training_id = ?", l.TrainingId).Count(&count)
	l.Position = count
	return db.Save(l).Error
}

// DeleteTrainingLesson removes a single lesson.
func DeleteTrainingLesson(id, trainingID int64) error {
	return db.Where("id = ? AND training_id = ?", id, trainingID).Delete(TrainingLesson{}).Error
}

// GetTrainingQuiz returns the quiz for a training, if one exists, with
// its questions ordered by position.
func GetTrainingQuiz(trainingID int64) (TrainingQuiz, error) {
	q := TrainingQuiz{}
	err := db.Where("training_id = ?", trainingID).First(&q).Error
	if err != nil {
		return q, err
	}
	err = db.Where("quiz_id = ?", q.Id).Order("position asc").Find(&q.Questions).Error
	return q, err
}

// PostTrainingQuiz creates or replaces the quiz settings (pass
// percentage, max attempts) for a training. A training has at most one
// quiz.
func PostTrainingQuiz(q *TrainingQuiz) error {
	existing, err := GetTrainingQuiz(q.TrainingId)
	if err == nil {
		q.Id = existing.Id
		return db.Model(&TrainingQuiz{}).Where("id = ?", q.Id).
			Updates(map[string]interface{}{"pass_percent": q.PassPercent, "max_attempts": q.MaxAttempts}).Error
	}
	return db.Save(q).Error
}

// AddQuizQuestion appends a question to a quiz. Position is computed
// server-side, same reasoning as AddTrainingLesson.
func AddQuizQuestion(q *QuizQuestion) error {
	var count int
	db.Model(&QuizQuestion{}).Where("quiz_id = ?", q.QuizId).Count(&count)
	q.Position = count
	return db.Save(q).Error
}

// DeleteQuizQuestion removes a single question.
func DeleteQuizQuestion(id, quizID int64) error {
	return db.Where("id = ? AND quiz_id = ?", id, quizID).Delete(QuizQuestion{}).Error
}
