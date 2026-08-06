package middleware

import (
	"net/http"
	"testing"

	"github.com/gophish/gophish/models"
)

func TestResolveTenantGrant(t *testing.T) {
	companyID := int64(20)
	grants := []models.TenantUser{
		{TenantID: 10, UserID: 1, CompanyID: &companyID, Role: "member"},
		{TenantID: 11, UserID: 1, Role: "tenant_admin"},
	}

	grant, status, err := resolveTenantGrant(grants, "11")
	if err != nil || status != http.StatusOK || grant.TenantID != 11 || grant.Role != "tenant_admin" {
		t.Fatalf("explicit grant = %#v, status = %d, err = %v", grant, status, err)
	}
	_, status, err = resolveTenantGrant(grants, "12")
	if err == nil || status != http.StatusForbidden {
		t.Fatalf("ungranted tenant status = %d, err = %v", status, err)
	}
	_, status, err = resolveTenantGrant(grants, "invalid")
	if err == nil || status != http.StatusBadRequest {
		t.Fatalf("invalid tenant status = %d, err = %v", status, err)
	}
	_, status, err = resolveTenantGrant(grants, "")
	if err == nil || status != http.StatusConflict {
		t.Fatalf("ambiguous tenant status = %d, err = %v", status, err)
	}
}

func TestResolveTenantGrantUsesOnlyGrant(t *testing.T) {
	grant, status, err := resolveTenantGrant([]models.TenantUser{{TenantID: 10, UserID: 1, Role: "member"}}, "")
	if err != nil || status != http.StatusOK || grant.TenantID != 10 {
		t.Fatalf("single grant = %#v, status = %d, err = %v", grant, status, err)
	}
	_, status, err = resolveTenantGrant(nil, "10")
	if err == nil || status != http.StatusForbidden {
		t.Fatalf("no grant status = %d, err = %v", status, err)
	}
}
