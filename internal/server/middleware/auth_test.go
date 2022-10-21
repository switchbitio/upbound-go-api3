package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/upbound/upbound-go-api3/internal/client/auth"
)

func sessionCheck(t *testing.T, authenticated bool, userID uint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, b := auth.UserIDFromContext(r.Context())
		if diff := cmp.Diff(userID, u); diff != "" {
			t.Errorf("\nUserIDFromContext(...): -want user, +got user:\n%s", diff)
		}
		if diff := cmp.Diff(authenticated, b); diff != "" {
			t.Errorf("\nUserIDFromContext(...): -want authenticated, +got authenticated:\n%s", diff)
		}
	}
}

func TestRequired(t *testing.T) {
	errBoom := errors.New("boom")
	type arguments struct {
		next   http.Handler
		cookie *http.Cookie
	}
	type want struct {
		status int
	}
	cases := map[string]struct {
		reason string
		m      auth.Client
		args   arguments
		want   want
	}{
		"Authenticated": {
			reason: "If session cookie is present and valid next handler should be called.",
			m: &auth.MockClient{
				GetUserIDFn: func(_ context.Context, _ string) (uint, error) {
					return 1, nil
				},
			},
			args: arguments{
				next: sessionCheck(t, true, 1),
				cookie: &http.Cookie{
					Name:  auth.SessionCookieName,
					Value: "inconsequential",
				},
			},
			want: want{
				status: http.StatusOK,
			},
		},
		"MissingSession": {
			reason: "If session cookie is not present an unauthorized status code should be returned.",
			args: arguments{
				cookie: &http.Cookie{},
			},
			want: want{
				status: http.StatusUnauthorized,
			},
		},
		"InvalidSession": {
			reason: "If session cookie is present but is not valid an unauthorized status code should be returned.",
			m: &auth.MockClient{
				GetUserIDFn: func(_ context.Context, _ string) (uint, error) {
					return 0, errBoom
				},
			},
			args: arguments{
				cookie: &http.Cookie{
					Name:  auth.SessionCookieName,
					Value: "inconsequential",
				},
			},
			want: want{
				status: http.StatusUnauthorized,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			a := NewAuthN(tc.m)
			rr := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(context.Background(), "GET", "doesnt/matter", nil)
			req.AddCookie(tc.args.cookie)
			a.Required(tc.args.next).ServeHTTP(rr, req)
			res := rr.Result()
			defer res.Body.Close()
			if diff := cmp.Diff(tc.want.status, res.StatusCode); diff != "" {
				t.Errorf("\n%s\nRequired(...): -want status, +got status:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestOptional(t *testing.T) {
	errBoom := errors.New("boom")
	type arguments struct {
		next   http.Handler
		cookie *http.Cookie
	}
	type want struct {
		status int
	}
	cases := map[string]struct {
		reason string
		m      auth.Client
		args   arguments
		want   want
	}{
		"Authenticated": {
			reason: "If session cookie is present and valid next handler should be called with user ID in context.",
			m: &auth.MockClient{
				GetUserIDFn: func(_ context.Context, _ string) (uint, error) {
					return 1, nil
				},
			},
			args: arguments{
				next: sessionCheck(t, true, 1),
				cookie: &http.Cookie{
					Name:  auth.SessionCookieName,
					Value: "inconsequential",
				},
			},
			want: want{
				status: http.StatusOK,
			},
		},
		"MissingSession": {
			reason: "If session cookie is not present the next handler should be called without user ID in context.",
			args: arguments{
				next:   sessionCheck(t, false, 0),
				cookie: &http.Cookie{},
			},
			want: want{
				status: http.StatusOK,
			},
		},
		"InvalidSession": {
			reason: "If session cookie is present but is not valid the next handler should be called without user ID in context.",
			m: &auth.MockClient{
				GetUserIDFn: func(_ context.Context, _ string) (uint, error) {
					return 0, errBoom
				},
			},
			args: arguments{
				next: sessionCheck(t, false, 0),
				cookie: &http.Cookie{
					Name:  auth.SessionCookieName,
					Value: "inconsequential",
				},
			},
			want: want{
				status: http.StatusOK,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			a := NewAuthN(tc.m)
			rr := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(context.Background(), "GET", "doesnt/matter", nil)
			req.AddCookie(tc.args.cookie)
			a.Optional(tc.args.next).ServeHTTP(rr, req)
			res := rr.Result()
			defer res.Body.Close()
			if diff := cmp.Diff(tc.want.status, res.StatusCode); diff != "" {
				t.Errorf("\n%s\nOptional(...): -want status, +got status:\n%s", tc.reason, diff)
			}
		})
	}
}
