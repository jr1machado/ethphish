package context

import (
	"net/http/httptest"
	"testing"
)

func TestTenantScopeRoundTrip(t *testing.T) {
	companyID := int64(14)
	req := httptest.NewRequest("GET", "/", nil)
	req = WithTenantScope(req, TenantScope{
		TenantID:  7,
		CompanyID: &companyID,
		UserID:    9,
		Role:      "tenant_admin",
	})

	scope, err := RequireTenantScope(req)
	if err != nil {
		t.Fatalf("requiring tenant scope: %v", err)
	}
	if scope.TenantID != 7 || scope.UserID != 9 || scope.CompanyID == nil || *scope.CompanyID != 14 {
		t.Fatalf("scope = %#v, want tenant 7, user 9, company 14", scope)
	}
}

func TestTenantScopeRejectsMissingOrInvalidScope(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if _, err := RequireTenantScope(req); err != ErrTenantScopeMissing {
		t.Fatalf("missing scope error = %v, want %v", err, ErrTenantScopeMissing)
	}

	req = WithTenantScope(req, TenantScope{TenantID: 1})
	if _, err := RequireTenantScope(req); err != ErrTenantScopeMissing {
		t.Fatalf("invalid scope error = %v, want %v", err, ErrTenantScopeMissing)
	}
}
