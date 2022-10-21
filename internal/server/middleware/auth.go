package middleware

import (
	"context"
	"net/http"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/upbound/upbound-go-api3/internal/client/auth"
)

const (
	errMissingSession = "failed to extract required session cookie"
	errGetUserID      = "failed to get user ID for session"
)

// AuthN is authentication middleware.
type AuthN struct {
	log logging.Logger
	mgr auth.Client
}

// AuthNOpt modifies authentication middleware.
type AuthNOpt func(a *AuthN)

// AuthNWithLogger sets the logger for authentication middleware.
func AuthNWithLogger(l logging.Logger) AuthNOpt {
	return func(a *AuthN) {
		a.log = l
	}
}

// NewAuthN constructs new authentication middleware.
func NewAuthN(mgr auth.Client, opts ...AuthNOpt) *AuthN {
	a := &AuthN{
		log: logging.NewNopLogger(),
		mgr: mgr,
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

// Required verifies a user is authenticated or aborts the request.
func (a *AuthN) Required(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie.
		c, err := r.Cookie(auth.SessionCookieName)
		if err != nil {
			a.log.Debug(errMissingSession, "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		id, err := a.mgr.GetUserID(r.Context(), c.Value)
		if err != nil {
			a.log.Debug(errGetUserID, "error", err)
			// TODO(hasheddan): consider returning another status code if error
			// is our fault.
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), auth.UserKey, id)))
	})
}

// Optional verifies a user if a session cookie is present.
func (a *AuthN) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie.
		c, err := r.Cookie(auth.SessionCookieName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		id, err := a.mgr.GetUserID(r.Context(), c.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), auth.UserKey, id)))
	})
}
