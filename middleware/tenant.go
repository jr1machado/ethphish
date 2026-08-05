package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
)

// TenantIDHeader identifies the tenant requested by an authenticated API
// caller. It is an authorization input, never an authorization decision: the
// middleware verifies it against tenant_users before it reaches a handler.
const TenantIDHeader = "X-EthPhish-Tenant-ID"

func resolveTenantGrant(grants []models.TenantUser, rawTenantID string) (models.TenantUser, int, error) {
	if len(grants) == 0 {
		return models.TenantUser{}, http.StatusForbidden, fmt.Errorf("no active tenant grant")
	}

	rawTenantID = strings.TrimSpace(rawTenantID)
	if rawTenantID == "" {
		if len(grants) != 1 {
			return models.TenantUser{}, http.StatusConflict, fmt.Errorf("tenant selection is required")
		}
		return grants[0], http.StatusOK, nil
	}

	tenantID, err := strconv.ParseInt(rawTenantID, 10, 64)
	if err != nil || tenantID <= 0 {
		return models.TenantUser{}, http.StatusBadRequest, fmt.Errorf("invalid tenant selection")
	}
	for _, grant := range grants {
		if grant.TenantID == tenantID {
			return grant, http.StatusOK, nil
		}
	}
	return models.TenantUser{}, http.StatusForbidden, fmt.Errorf("tenant access denied")
}

// ResolveTenantScope loads the authenticated user's authorized tenant scope.
// It must run after RequireAPIKey or a session-authentication middleware.
// Handlers protected by this middleware can rely on the typed request scope
// instead of trusting URL values or client supplied headers.
func ResolveTenantScope(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := ctx.Get(r, "user").(models.User)
		if !ok || user.Id <= 0 {
			JSONError(w, http.StatusUnauthorized, "Authenticated user is required")
			return
		}
		grants, err := models.GetTenantUsers(user.Id)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "Unable to resolve tenant access")
			return
		}
		grant, status, err := resolveTenantGrant(grants, r.Header.Get(TenantIDHeader))
		if err != nil {
			JSONError(w, status, err.Error())
			return
		}
		scope := ctx.TenantScope{
			TenantID:  grant.TenantID,
			CompanyID: grant.CompanyID,
			UserID:    user.Id,
			Role:      grant.Role,
		}
		next.ServeHTTP(w, ctx.WithTenantScope(r, scope))
	}
}
