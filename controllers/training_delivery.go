package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gophish/gophish/auth"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// registerTrainingRoutes mounts the public, token-identified training
// delivery surface on the phishing server. Unlike /approvals/* and
// /portal/*, there's no login step — the token in the URL identifies the
// assignment directly, the same trust model as a campaign's rid. CSRF is
// still applied to the quiz POST, scoped to this subtree.
func (ps *PhishingServer) registerTrainingRoutes(router *mux.Router) {
	sub := mux.NewRouter()
	sub.HandleFunc("/training/{token}", ps.TrainingStart).Methods(http.MethodGet)
	sub.HandleFunc("/training/{token}/lessons/{n:[0-9]+}", ps.TrainingLesson).Methods(http.MethodGet)
	sub.HandleFunc("/training/{token}/quiz", ps.TrainingQuizView).Methods(http.MethodGet)
	sub.HandleFunc("/training/{token}/quiz", ps.TrainingQuizSubmit).Methods(http.MethodPost)

	csrfKey := []byte(ps.config.CSRFKey)
	if len(csrfKey) == 0 {
		csrfKey = []byte(auth.GenerateSecureKey(auth.APIKeyLength))
	}
	protected := csrf.Protect(csrfKey, csrf.Path("/training"), csrf.Secure(ps.config.UseTLS))(sub)
	router.PathPrefix("/training").Handler(protected)
}

// maybeRedirectToTraining implements the "teachable moment" delivery
// path: if the campaign names a training and the request method matches
// its configured trigger (GET -> click, POST -> submit, or "both"
// matches either), it gets-or-creates a TrainingAssignment for this
// result and redirects there instead of letting PhishHandler continue to
// the normal landing-page response. Returns false (no redirect) for any
// campaign that doesn't opt into this.
func (ps *PhishingServer) maybeRedirectToTraining(w http.ResponseWriter, r *http.Request, c models.Campaign, rs models.Result) bool {
	if c.TrainingID == nil {
		return false
	}
	trigger := c.TrainingTrigger
	matches := trigger == models.TrainingTriggerBoth ||
		(r.Method == http.MethodGet && trigger == models.TrainingTriggerClick) ||
		(r.Method == http.MethodPost && trigger == models.TrainingTriggerSubmit)
	if !matches {
		return false
	}
	a, err := models.GetOrCreateTrainingAssignmentForResult(c.TenantID, *c.TrainingID, c.Id, rs.Id, rs.Email, rs.FirstName)
	if err != nil {
		log.Error(err)
		return false
	}
	http.Redirect(w, r, fmt.Sprintf("/training/%s", a.Token), http.StatusFound)
	return true
}

// resolveTrainingAssignment loads the assignment and training for a
// token, or renders a generic "invalid link" page and returns ok=false.
func execTrainingTemplate(w http.ResponseWriter, name string, data interface{}) {
	if err := approvalPortalTemplate(name).Execute(w, data); err != nil {
		log.Error(err)
	}
}

func resolveTrainingAssignment(w http.ResponseWriter, r *http.Request) (models.TrainingAssignment, models.Training, bool) {
	token := mux.Vars(r)["token"]
	a, err := models.GetTrainingAssignmentByToken(token)
	if err != nil {
		execTrainingTemplate(w, "training_invalid", nil)
		return a, models.Training{}, false
	}
	t, err := models.GetTrainingByID(a.TrainingId)
	if err != nil {
		execTrainingTemplate(w, "training_invalid", nil)
		return a, t, false
	}
	return a, t, true
}

// TrainingStart redirects to the first unseen lesson, or the quiz if
// every lesson has been seen (or there are none).
func (ps *PhishingServer) TrainingStart(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	a, t, ok := resolveTrainingAssignment(w, r)
	if !ok {
		return
	}
	seen, err := models.ViewedLessonIDs(a.Id)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	for i, lesson := range t.Lessons {
		if !seen[lesson.Id] {
			http.Redirect(w, r, fmt.Sprintf("/training/%s/lessons/%d", token, i+1), http.StatusFound)
			return
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/training/%s/quiz", token), http.StatusFound)
}

// TrainingLesson renders one lesson (1-indexed by its position in the
// training's lesson list) and marks it viewed.
func (ps *PhishingServer) TrainingLesson(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	a, t, ok := resolveTrainingAssignment(w, r)
	if !ok {
		return
	}
	n, _ := strconv.Atoi(mux.Vars(r)["n"])
	if n < 1 || n > len(t.Lessons) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	lesson := t.Lessons[n-1]
	if err := models.MarkLessonViewed(a.Id, lesson.Id); err != nil {
		log.Error(err)
	}
	nextURL := fmt.Sprintf("/training/%s/quiz", token)
	if n < len(t.Lessons) {
		nextURL = fmt.Sprintf("/training/%s/lessons/%d", token, n+1)
	}
	execTrainingTemplate(w, "training_lesson", struct {
		Training   models.Training
		Lesson     models.TrainingLesson
		LessonHTML template.HTML
		Index      int
		Total      int
		NextURL    string
	}{t, lesson, template.HTML(lesson.HTML), n, len(t.Lessons), nextURL})
}

// TrainingQuizView renders the quiz form, or a "no quiz configured"
// completion notice if the training has no quiz.
func (ps *PhishingServer) TrainingQuizView(w http.ResponseWriter, r *http.Request) {
	a, t, ok := resolveTrainingAssignment(w, r)
	if !ok {
		return
	}
	if t.Quiz == nil {
		// No quiz to take — a lessons-only training is "complete" once
		// every lesson has been viewed.
		if err := models.MarkAssignmentCompleted(a.Id); err != nil {
			log.Error(err)
		}
		execTrainingTemplate(w, "training_complete", struct {
			Training models.Training
			Passed   bool
		}{t, true})
		return
	}
	execTrainingTemplate(w, "training_quiz", struct {
		Training   models.Training
		Assignment models.TrainingAssignment
		Token      string
		Error      string
	}{t, a, csrf.Token(r), ""})
}

// TrainingQuizSubmit grades a quiz submission and shows the result.
func (ps *PhishingServer) TrainingQuizSubmit(w http.ResponseWriter, r *http.Request) {
	a, t, ok := resolveTrainingAssignment(w, r)
	if !ok {
		return
	}
	if t.Quiz == nil {
		http.Error(w, "This training has no quiz", http.StatusBadRequest)
		return
	}
	r.ParseForm()
	answers := map[string]string{}
	for _, q := range t.Quiz.Questions {
		key := strconv.FormatInt(q.Id, 10)
		answers[key] = r.FormValue("q_" + key)
	}
	attempt, err := models.GradeQuizAttempt(a, *t.Quiz, answers)
	if err != nil {
		execTrainingTemplate(w, "training_quiz", struct {
			Training   models.Training
			Assignment models.TrainingAssignment
			Token      string
			Error      string
		}{t, a, csrf.Token(r), err.Error()})
		return
	}
	execTrainingTemplate(w, "training_complete", struct {
		Training models.Training
		Passed   bool
		Score    int
	}{t, attempt.Passed, attempt.Score})
}
