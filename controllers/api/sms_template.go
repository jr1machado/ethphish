package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// SMSTemplates handles requests for the /api/sms_templates/ endpoint
func (as *Server) SMSTemplates(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	userID := ctx.Get(r, "user_id").(int64)
	switch {
	case r.Method == "GET":
		ts, err := models.GetSMSTemplatesForTenant(scope.TenantID, userID)
		if err != nil {
			log.Error(err)
		}
		JSONResponse(w, ts, http.StatusOK)
	// DELETE: Bulk delete SMS templates
	case r.Method == "DELETE":
		var req struct {
			IDs []int64 `json:"ids"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		if len(req.IDs) == 0 {
			JSONResponse(w, models.Response{Success: false, Message: "No SMS template IDs provided"}, http.StatusBadRequest)
			return
		}
		err = models.DeleteSMSTemplatesForTenant(req.IDs, scope.TenantID, userID)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting SMS templates: " + err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "SMS templates deleted successfully!"}, http.StatusOK)
	//POST: Create a new SMS template and return it as JSON
	case r.Method == "POST":
		t := models.SMSTemplate{}
		// Put the request into a template
		err := json.NewDecoder(r.Body).Decode(&t)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}
		// Check to make sure the name is unique
		_, err = models.GetSMSTemplateByNameForTenant(t.Name, scope.TenantID, userID)
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "SMS template name already in use"}, http.StatusConflict)
			log.Error(err)
			return
		}
		t.UserId = userID
		t.TenantID = scope.TenantID
		err = t.Validate()
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		err = models.PostSMSTemplateForTenant(&t, scope.TenantID)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, t, http.StatusCreated)
	}
}

// SMSTemplate contains functions to handle the GET'ing, DELETE'ing, and PUT'ing
// of a SMS template object
func (as *Server) SMSTemplate(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	userID := ctx.Get(r, "user_id").(int64)
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	t, err := models.GetSMSTemplateForTenant(id, scope.TenantID, userID)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "SMS template not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, t, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteSMSTemplateForTenant(id, scope.TenantID, userID)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting SMS template"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "SMS Template Deleted Successfully"}, http.StatusOK)
	case r.Method == "PUT":
		t = models.SMSTemplate{}
		err = json.NewDecoder(r.Body).Decode(&t)
		if err != nil {
			log.Error(err)
		}
		if t.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "/:id and /:template_id mismatch"}, http.StatusBadRequest)
			return
		}
		t.UserId = userID
		t.TenantID = scope.TenantID
		err = t.Validate()
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		err = models.PutSMSTemplateForTenant(&t, scope.TenantID)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error updating SMS template"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, t, http.StatusOK)
	}
}
