package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gophish/gophish/config"
	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/reports"
	"github.com/gorilla/mux"
)

// contractFileManager stores uploaded contract documents, sibling to the
// reports storage directory (see InitContractServices).
var contractFileManager *reports.FileManager

// maxContractUploadBytes caps the total request body read for a contract
// document upload (not just the in-memory multipart buffer).
const maxContractUploadBytes = 32 << 20

// InitContractServices wires up contract document storage. Reuses the
// reports.FileManager (year/month sharding + path-traversal guard) rather
// than a new implementation — see reports/file_manager.go.
func InitContractServices(reportConfig *config.Reports) {
	base := filepath.Join(".", "reports", "contracts")
	if reportConfig != nil && reportConfig.StoragePath != "" {
		base = filepath.Join(filepath.Dir(reportConfig.StoragePath), "contracts")
	}
	contractFileManager = reports.NewFileManagerWithPath(base)
}

// Contracts handles listing and creating contracts.
func (as *Server) Contracts(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	switch r.Method {
	case "GET":
		cs, err := models.GetContractsForTenant(scope.TenantID)
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, cs, http.StatusOK)
	case "POST":
		c := models.Contract{}
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		userID := ctx.Get(r, "user_id").(int64)
		if err := models.PostContractForTenant(&c, scope.TenantID, userID); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, c, http.StatusCreated)
	default:
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
	}
}

// ContractID handles retrieving, updating, and deleting a single contract.
func (as *Server) ContractID(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	switch r.Method {
	case "GET":
		c, err := models.GetContractForTenant(id, scope.TenantID)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
			return
		}
		JSONResponse(w, c, http.StatusOK)
	case "PUT":
		c := models.Contract{}
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		c.Id = id
		if err := models.PutContractForTenant(&c, scope.TenantID); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, c, http.StatusOK)
	case "DELETE":
		if err := models.DeleteContractForTenant(id, scope.TenantID); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Contract deleted"}, http.StatusOK)
	default:
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
	}
}

// ContractVersions handles uploading a new document version for a contract.
// Expects a multipart form with a "file" field plus optional
// "scope_description" and "notes" fields.
func (as *Server) ContractVersions(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	contractID, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)

	if _, err := models.GetContractForTenant(contractID, scope.TenantID); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	switch r.Method {
	case "GET":
		vs, err := models.GetContractVersions(contractID)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, vs, http.StatusOK)
	case "POST":
		r.Body = http.MaxBytesReader(w, r.Body, maxContractUploadBytes)
		if err := r.ParseMultipartForm(maxContractUploadBytes); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error parsing upload (file may be too large)"}, http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "A contract document file is required"}, http.StatusBadRequest)
			return
		}
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error reading upload"}, http.StatusInternalServerError)
			return
		}
		// filepath.Base strips any directory components an attacker-supplied
		// filename might smuggle in (e.g. "../../evil.sh"), since
		// FileManager.SaveReport joins this straight onto its storage path
		// with no containment check of its own.
		safeName := filepath.Base(header.Filename)
		storedName := fmt.Sprintf("contract_%d_%d_%s", contractID, time.Now().UTC().UnixNano(), safeName)
		filePath, err := contractFileManager.SaveReport(contractID, storedName, data)
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: "Error storing document"}, http.StatusInternalServerError)
			return
		}
		userID := ctx.Get(r, "user_id").(int64)
		v := models.ContractVersion{
			FilePath:         filePath,
			FileName:         header.Filename,
			ScopeDescription: r.FormValue("scope_description"),
			Notes:            r.FormValue("notes"),
		}
		saved, err := models.AddContractVersion(contractID, &v, userID)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, saved, http.StatusCreated)
	default:
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
	}
}

// ContractVersionDownload streams a stored contract document back to an
// authenticated admin.
func (as *Server) ContractVersionDownload(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	id, _ := strconv.ParseInt(mux.Vars(r)["versionId"], 0, 64)
	v, err := models.GetContractVersionForTenant(id, scope.TenantID)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	data, err := contractFileManager.GetReportFile(v.FilePath)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", v.FileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

// ContractApprovers handles listing and adding named approvers for a
// contract.
func (as *Server) ContractApprovers(w http.ResponseWriter, r *http.Request) {
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	contractID, _ := strconv.ParseInt(mux.Vars(r)["id"], 0, 64)
	if _, err := models.GetContractForTenant(contractID, scope.TenantID); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	switch r.Method {
	case "GET":
		as_, err := models.GetContractApprovers(contractID)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, as_, http.StatusOK)
	case "POST":
		a := models.ContractApprover{}
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		a.ContractId = contractID
		if a.Name == "" || a.Email == "" {
			JSONResponse(w, models.Response{Success: false, Message: "Approver name and email are required"}, http.StatusBadRequest)
			return
		}
		if err := models.AddContractApprover(&a); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, a, http.StatusCreated)
	default:
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
	}
}

// ContractApproverID handles removing a single approver.
func (as *Server) ContractApproverID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}
	scope, err := ctx.RequireTenantScope(r)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Tenant scope is required"}, http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	contractID, _ := strconv.ParseInt(vars["id"], 0, 64)
	approverID, _ := strconv.ParseInt(vars["approverId"], 0, 64)
	if _, err := models.GetContractForTenant(contractID, scope.TenantID); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}
	if err := models.DeleteContractApprover(approverID, contractID); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	JSONResponse(w, models.Response{Success: true, Message: "Approver removed"}, http.StatusOK)
}
