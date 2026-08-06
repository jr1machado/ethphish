package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gophish/gophish/approvals"
	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

// trainingPortalBaseURL mirrors approvalPortalBaseURL — the public base
// URL used to build absolute links in training-assignment e-mails. Set by
// InitApprovalServices, since both live on the same public phishing
// server and share one config value.
var trainingPortalBaseURL string

// Trainings handles listing and creating trainings.
func (as *Server) Trainings(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	switch r.Method {
	case "GET":
		ts, err := models.GetTrainingsForTenant(scope.TenantID)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, ts, http.StatusOK)
	case "POST":
		t := models.Training{}
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		userID := ctx.Get(r, "user_id").(int64)
		if err := models.PostTrainingForTenant(&t, scope.TenantID, userID); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, t, http.StatusCreated)
	default:
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
	}
}

// TrainingID handles retrieving, updating, and deleting a single
// training.
func (as *Server) TrainingID(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	switch r.Method {
	case "GET":
		t, err := models.GetTrainingForTenant(id, scope.TenantID)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
			return
		}
		JSONResponse(w, t, http.StatusOK)
	case "PUT":
		t := models.Training{}
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		t.Id = id
		if err := models.PutTrainingForTenant(&t, scope.TenantID); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, t, http.StatusOK)
	case "DELETE":
		if err := models.DeleteTrainingForTenant(id, scope.TenantID); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Training deleted"}, http.StatusOK)
	default:
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
	}
}

// TrainingLessons handles adding a lesson to a training.
func (as *Server) TrainingLessons(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	trainingID, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	if _, err := models.GetTrainingForTenant(trainingID, scope.TenantID); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}
	l := models.TrainingLesson{}
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
		return
	}
	l.TrainingId = trainingID
	if err := models.AddTrainingLesson(&l); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}
	JSONResponse(w, l, http.StatusCreated)
}

// TrainingLessonID handles deleting a single lesson.
func (as *Server) TrainingLessonID(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	trainingID, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	lessonID, _ := strconv.ParseInt(mux.Vars(r)["lessonId"], 0, 64)
	if _, err := models.GetTrainingForTenant(trainingID, scope.TenantID); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	if r.Method != "DELETE" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}
	if err := models.DeleteTrainingLesson(lessonID, trainingID); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}
	JSONResponse(w, models.Response{Success: true, Message: "Lesson deleted"}, http.StatusOK)
}

// TrainingQuiz handles creating/updating a training's quiz settings.
func (as *Server) TrainingQuiz(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	trainingID, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	if _, err := models.GetTrainingForTenant(trainingID, scope.TenantID); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}
	q := models.TrainingQuiz{}
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
		return
	}
	q.TrainingId = trainingID
	if q.PassPercent <= 0 {
		q.PassPercent = 80
	}
	if err := models.PostTrainingQuiz(&q); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}
	JSONResponse(w, q, http.StatusOK)
}

// TrainingQuizQuestions handles adding a question to a training's quiz.
func (as *Server) TrainingQuizQuestions(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	trainingID, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	if _, err := models.GetTrainingForTenant(trainingID, scope.TenantID); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	quiz, err := models.GetTrainingQuiz(trainingID)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Create the quiz before adding questions"}, http.StatusBadRequest)
		return
	}
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}
	q := models.QuizQuestion{}
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
		return
	}
	q.QuizId = quiz.Id
	if q.Type != models.QuestionTrueFalse {
		q.Type = models.QuestionMultipleChoice
	}
	if err := models.AddQuizQuestion(&q); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}
	JSONResponse(w, q, http.StatusCreated)
}

// trainingAssignRequest is the body of POST /api/trainings/{id}/assign.
type trainingAssignRequest struct {
	GroupID int64 `json:"group_id"`
}

// TrainingAssign creates one TrainingAssignment per target in the given
// group and e-mails each their unique access link.
func (as *Server) TrainingAssign(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	trainingID, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	training, err := models.GetTrainingForTenant(trainingID, scope.TenantID)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	req := trainingAssignRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
		return
	}
	userID := ctx.Get(r, "user_id").(int64)
	group, err := models.GetGroupForTenant(req.GroupID, scope.TenantID, userID)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
		return
	}
	assigned := 0
	for _, target := range group.Targets {
		if target.Email == "" {
			continue
		}
		a, err := models.CreateTrainingAssignment(scope.TenantID, training.Id, nil, nil, target.Email, target.FirstName)
		if err != nil {
			log.Error(err)
			continue
		}
		link := fmt.Sprintf("%s/training/%s", trainingPortalBaseURL, a.Token)
		if err := approvals.SendTrainingAssignmentEmail(scope.TenantID, userID, target.Email, target.FirstName, training.Name, link); err != nil {
			log.Error(err)
			continue
		}
		assigned++
	}
	JSONResponse(w, models.Response{Success: true, Message: fmt.Sprintf("Assigned to %d target(s)", assigned)}, http.StatusOK)
}
