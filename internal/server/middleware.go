package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cicconee/weather-app/internal/admin"
	"github.com/cicconee/weather-app/internal/app"
)

const adminTokenCookieKey = "admin_token"

// AdminValidater is a middleware that is wrapped around admin paths.
// Any HTTP request that requires a valid admin should be wrapped in the
// Validate func.
type AdminValidater struct {
	admins *admin.Service
	logger *log.Logger
}

// Validate will verify that the caller is a admin. If the user making the request
// has a valid admin token cookie, next will execute. The request context passed to next
// will contain a key "admin_id" that will contain the id of the validated admin.
func (v *AdminValidater) Validate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lw := NewLogWriter(v.logger, w, r)

		cookie, err := r.Cookie(adminTokenCookieKey)
		if err != nil {
			appErr := &app.ServerResponseError{
				Err:        fmt.Errorf("getting %s cookie: %v\n", adminTokenCookieKey, err),
				Msg:        "Please login",
				StatusCode: http.StatusUnauthorized,
			}
			v.logAbort(r, appErr, "AdminValidater.Validate")
			lw.WriteError(appErr)
			return
		}

		account, err := v.admins.Validate(r.Context(), cookie.Value)
		if err != nil {
			err = fmt.Errorf("validating token: %w", err)
			v.logAbort(r, err, "AdminValidater.Validate")
			lw.WriteError(err)
			return
		}

		if !account.IsApproved() {
			appErr := &app.ServerResponseError{
				Err:        fmt.Errorf("admin not approved (id=%d)", account.ID),
				Msg:        "Your admin rights are under review",
				StatusCode: http.StatusUnauthorized,
			}
			v.logAbort(r, appErr, "AdminValidater.Validate")
			lw.WriteError(appErr)
			return
		}

		next(w, r.WithContext(context.WithValue(r.Context(), "admin_id", account.ID)))
	}
}

type logParams struct {
	AccountID int
}

func (v *AdminValidater) logAbort(r *http.Request, err error, entry string) {
	v.logger.Printf("%s %s %s: aborting admin request: %v\n", r.Method, r.URL.Path, entry, err)
}
