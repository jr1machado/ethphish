package context

import (
	"errors"
	"net/http"
)

// TenantScope describes the authorization boundary selected for one request.
// A company is optional because a tenant administrator can be authorized for
// all companies in the tenant.
type TenantScope struct {
	TenantID  int64
	CompanyID *int64
	UserID    int64
	Role      string
}

type tenantScopeKey struct{}

var ErrTenantScopeMissing = errors.New("tenant scope is missing")

// WithTenantScope attaches a verified tenant boundary to a request. It is
// intentionally separate from the legacy string context keys to prevent an
// untrusted handler from accidentally replacing the scope.
func WithTenantScope(r *http.Request, scope TenantScope) *http.Request {
	return Set(r, tenantScopeKey{}, scope)
}

// TenantScopeFromRequest retrieves the verified tenant boundary, if present.
func TenantScopeFromRequest(r *http.Request) (TenantScope, bool) {
	scope, ok := Get(r, tenantScopeKey{}).(TenantScope)
	if !ok || scope.TenantID <= 0 || scope.UserID <= 0 {
		return TenantScope{}, false
	}
	return scope, true
}

// RequireTenantScope returns the scope or a stable error suitable for an
// authorization middleware. Callers must not execute tenant-owned queries
// without a scope.
func RequireTenantScope(r *http.Request) (TenantScope, error) {
	scope, ok := TenantScopeFromRequest(r)
	if !ok {
		return TenantScope{}, ErrTenantScopeMissing
	}
	return scope, nil
}
