package models

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

// Assignment statuses.
const (
	AssignmentAssigned   string = "assigned"
	AssignmentInProgress string = "in_progress"
	AssignmentCompleted  string = "completed"
	AssignmentFailed     string = "failed"
)

// assignmentTokenLength matches the other opaque tokens in this package
// (see models/token.go).
const assignmentTokenLength = 32

// ErrAssignmentTokenInvalid is thrown when a training access token
// doesn't match any assignment.
var ErrAssignmentTokenInvalid = errors.New("Training link is invalid")

// TrainingAssignment is one target's access to one training. Its token is
// stored in cleartext, not hashed — unlike the approval/portal-login
// magic links (which grant a one-shot security decision and are hashed
// like a password), this token is a revisitable resource identifier: the
// same assignment is visited repeatedly across lessons and quiz attempts,
// exactly like a campaign's Result.RId, which is the precedent this
// follows (see generateResultId in models/result.go).
type TrainingAssignment struct {
	Id          int64      `json:"id"`
	TenantID    int64      `json:"-" gorm:"column:tenant_id;default:1"`
	TrainingId  int64      `json:"training_id"`
	CampaignId  *int64     `json:"campaign_id,omitempty"`
	ResultId    *int64     `json:"result_id,omitempty"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Token       string     `json:"-"`
	Status      string     `json:"status"`
	Attempts    int        `json:"attempts"`
	BestScore   int        `json:"best_score"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func (a *TrainingAssignment) TableName() string { return "training_assignments" }

// TrainingLessonView records that an assignment has seen a lesson.
type TrainingLessonView struct {
	Id           int64     `json:"id"`
	AssignmentId int64     `json:"assignment_id"`
	LessonId     int64     `json:"lesson_id"`
	ViewedAt     time.Time `json:"viewed_at"`
}

func (v *TrainingLessonView) TableName() string { return "training_lesson_views" }

// QuizAttempt is one graded submission of a training's quiz.
type QuizAttempt struct {
	Id            int64     `json:"id"`
	AssignmentId  int64     `json:"assignment_id"`
	AttemptNumber int       `json:"attempt_number"`
	Score         int       `json:"score"`
	Passed        bool      `json:"passed"`
	Answers       string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
}

func (a *QuizAttempt) TableName() string { return "quiz_attempts" }

// CreateTrainingAssignment mints a new assignment and its access token
// for one target. campaignID/resultID are nil for a direct group
// assignment, set for one spawned by a campaign's post-click trigger.
func CreateTrainingAssignment(tenantID, trainingID int64, campaignID, resultID *int64, email, name string) (TrainingAssignment, error) {
	token, err := generateOpaqueToken(assignmentTokenLength)
	if err != nil {
		return TrainingAssignment{}, err
	}
	a := TrainingAssignment{
		TenantID:   tenantID,
		TrainingId: trainingID,
		CampaignId: campaignID,
		ResultId:   resultID,
		Email:      email,
		Name:       name,
		Token:      token,
		Status:     AssignmentAssigned,
		CreatedAt:  time.Now().UTC(),
	}
	err = db.Save(&a).Error
	return a, err
}

// GetOrCreateTrainingAssignmentForResult returns the existing assignment
// for a (campaign, result) pair if one was already spawned by an earlier
// matching event, or creates one — so a "both" trigger (click then
// submit) only ever produces one assignment per result, not two, and a
// second matching event still has a token to redirect through.
func GetOrCreateTrainingAssignmentForResult(tenantID, trainingID, campaignID, resultID int64, email, name string) (TrainingAssignment, error) {
	existing := TrainingAssignment{}
	err := db.Where("campaign_id = ? AND result_id = ?", campaignID, resultID).First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return existing, err
	}
	cID, rID := campaignID, resultID
	return CreateTrainingAssignment(tenantID, trainingID, &cID, &rID, email, name)
}

// GetTrainingAssignmentByToken resolves an access token to its
// assignment.
func GetTrainingAssignmentByToken(token string) (TrainingAssignment, error) {
	a := TrainingAssignment{}
	err := db.Where("token = ?", token).First(&a).Error
	if err == gorm.ErrRecordNotFound {
		return a, ErrAssignmentTokenInvalid
	}
	return a, err
}

// MarkLessonViewed records a lesson as seen by an assignment (idempotent
// — viewing the same lesson twice doesn't duplicate the row) and, on an
// assignment's first viewed lesson, flips its status to in_progress.
func MarkLessonViewed(assignmentID, lessonID int64) error {
	var count int
	db.Model(&TrainingLessonView{}).Where("assignment_id = ? AND lesson_id = ?", assignmentID, lessonID).Count(&count)
	if count == 0 {
		if err := db.Save(&TrainingLessonView{AssignmentId: assignmentID, LessonId: lessonID, ViewedAt: time.Now().UTC()}).Error; err != nil {
			return err
		}
	}
	return db.Model(&TrainingAssignment{}).Where("id = ? AND status = ?", assignmentID, AssignmentAssigned).
		Update("status", AssignmentInProgress).Error
}

// MarkAssignmentCompleted flips an assignment straight to completed —
// used for a quiz-less training, where finishing the lessons is the
// entire completion criterion.
func MarkAssignmentCompleted(assignmentID int64) error {
	now := time.Now().UTC()
	return db.Model(&TrainingAssignment{}).Where("id = ?", assignmentID).
		Updates(map[string]interface{}{"status": AssignmentCompleted, "completed_at": now}).Error
}

// ViewedLessonIDs returns the lesson IDs an assignment has already seen.
func ViewedLessonIDs(assignmentID int64) (map[int64]bool, error) {
	views := []TrainingLessonView{}
	if err := db.Where("assignment_id = ?", assignmentID).Find(&views).Error; err != nil {
		return nil, err
	}
	seen := make(map[int64]bool, len(views))
	for _, v := range views {
		seen[v.LessonId] = true
	}
	return seen, nil
}

// GradeQuizAttempt scores a submitted answer set against a quiz,
// enforces the max-attempts limit, records the attempt, and updates the
// assignment's status/best score. answers maps question ID (as string,
// matching form field naming) to the submitted answer.
func GradeQuizAttempt(assignment TrainingAssignment, quiz TrainingQuiz, answers map[string]string) (QuizAttempt, error) {
	if quiz.MaxAttempts > 0 && assignment.Attempts >= quiz.MaxAttempts {
		return QuizAttempt{}, errors.New("Maximum quiz attempts reached")
	}

	correct := 0
	for _, q := range quiz.Questions {
		submitted, ok := answers[strconv.FormatInt(q.Id, 10)]
		if ok && normalizeAnswer(submitted) == normalizeAnswer(q.CorrectAnswer) {
			correct++
		}
	}
	total := len(quiz.Questions)
	score := 0
	if total > 0 {
		score = correct * 100 / total
	}
	passed := score >= quiz.PassPercent

	encodedAnswers, _ := json.Marshal(answers)
	attempt := QuizAttempt{
		AssignmentId:  assignment.Id,
		AttemptNumber: assignment.Attempts + 1,
		Score:         score,
		Passed:        passed,
		Answers:       string(encodedAnswers),
		CreatedAt:     time.Now().UTC(),
	}
	if err := db.Save(&attempt).Error; err != nil {
		return attempt, err
	}

	updates := map[string]interface{}{"attempts": assignment.Attempts + 1}
	if score > assignment.BestScore {
		updates["best_score"] = score
	}
	if passed {
		updates["status"] = AssignmentCompleted
		updates["completed_at"] = time.Now().UTC()
	} else if quiz.MaxAttempts > 0 && assignment.Attempts+1 >= quiz.MaxAttempts {
		updates["status"] = AssignmentFailed
	}
	if err := db.Model(&TrainingAssignment{}).Where("id = ?", assignment.Id).Updates(updates).Error; err != nil {
		return attempt, err
	}
	return attempt, nil
}

func normalizeAnswer(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
}
