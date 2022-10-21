package auth

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	shttp "github.com/upbound/upbound-go-api3/internal/client/http"
	serrors "github.com/upbound/upbound-go-api3/internal/errors"
)

var _ Client = &ExternalClient{}
var _ Client = &MockClient{}

func TestGetUserID(t *testing.T) {
	auth, _ := url.Parse("https://api-private-auth:8080")
	private, _ := url.Parse("https://api-private:8080")
	errBoom := errors.New("boom")
	type arguments struct {
		token string
	}
	type want struct {
		id  uint
		err error
	}
	cases := map[string]struct {
		reason string
		c      shttp.Client
		args   arguments
		want   want
	}{
		"Success": {
			reason: "If status code is OK and response body is valid, user ID should be returned with no error.",
			c: &shttp.MockClient{
				DoFn: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(strings.NewReader(`{"userID": 1}`)),
						StatusCode: http.StatusOK,
					}, nil
				},
			},
			args: arguments{
				token: "test",
			},
			want: want{
				id:  1,
				err: nil,
			},
		},
		"ErrorDo": {
			reason: "If performing the request causes an error then an error should be returned.",
			c: &shttp.MockClient{
				DoFn: func(req *http.Request) (*http.Response, error) {
					return &http.Response{}, errBoom
				},
			},
			args: arguments{
				token: "test",
			},
			want: want{
				id:  0,
				err: errors.Wrap(errBoom, errDoSessionRequest),
			},
		},
		"ErrorNotFound": {
			reason: "If response has not found response code then a not found error should be returned.",
			c: &shttp.MockClient{
				DoFn: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(strings.NewReader("not found")),
						StatusCode: http.StatusNotFound,
					}, nil
				},
			},
			args: arguments{
				token: "test",
			},
			want: want{
				id:  0,
				err: serrors.NewNotFound(errors.New(errNotFound)),
			},
		},
		"ErrorResponseCode": {
			reason: "If response has unsuccessful response code then an error should be returned.",
			c: &shttp.MockClient{
				DoFn: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(strings.NewReader("error")),
						StatusCode: http.StatusInternalServerError,
					}, nil
				},
			},
			args: arguments{
				token: "test",
			},
			want: want{
				id:  0,
				err: errors.New(errSessionResponse),
			},
		},
		"ErrorBadResponse": {
			reason: "If response code is success, but body is invalid an error should be returned.",
			c: &shttp.MockClient{
				DoFn: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(strings.NewReader("")),
						StatusCode: http.StatusOK,
					}, nil
				},
			},
			args: arguments{
				token: "test",
			},
			want: want{
				id:  0,
				err: errors.Wrap(io.EOF, errInvalidSessionResponseBody),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			m := New(*auth, *private, WithClient(tc.c))
			uid, err := m.GetUserID(context.Background(), tc.args.token)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nGetUserID(...): -want err, +got err:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.id, uid); diff != "" {
				t.Errorf("\n%s\nGetUserID(...): -want id, +got id:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestUserIDFromContext(t *testing.T) {
	type arguments struct {
		ctx context.Context
	}
	type want struct {
		id     uint
		exists bool
	}
	cases := map[string]struct {
		reason string
		args   arguments
		want   want
	}{
		"Exists": {
			reason: "If values exists for context key it should be returned.",
			args: arguments{
				ctx: context.WithValue(context.Background(), UserKey, uint(1)),
			},
			want: want{
				exists: true,
				id:     1,
			},
		},
		"WrongType": {
			reason: "If value for context key is wrong type it does not exist.",
			args: arguments{
				ctx: context.WithValue(context.Background(), UserKey, 1),
			},
			want: want{
				exists: false,
				id:     0,
			},
		},
		"NoKey": {
			reason: "If there is no key then it does not exist.",
			args: arguments{
				ctx: context.Background(),
			},
			want: want{
				exists: false,
				id:     0,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			uid, exists := UserIDFromContext(tc.args.ctx)
			if diff := cmp.Diff(tc.want.id, uid); diff != "" {
				t.Errorf("\n%s\nUserIDFromContext(...): -want id, +got id:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.exists, exists); diff != "" {
				t.Errorf("\n%s\nUserIDFromContext(...): -want exists, +got exists:\n%s", tc.reason, diff)
			}
		})
	}
}
