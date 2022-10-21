package acl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	shttp "github.com/upbound/upbound-go-api3/internal/client/http"
	serrors "github.com/upbound/upbound-go-api3/internal/errors"
	"github.com/upbound/upbound-go-api3/internal/types"
)

var _ Client = &ExternalClient{}
var _ Client = &MockClient{}

func TestGetACL_Users(t *testing.T) {
	u, _ := url.Parse("https://api-private:8080")
	errBoom := errors.New("boom")
	a := ACL{
		Accounts: []AccountAccess{
			{
				ID:         1,
				Name:       "hasheddan",
				Permission: Owner,
			},
		},
		Teams: []TeamAccess{
			{
				ID: types.UUID{
					UUID: uuid.MustParse("fc5105af-e023-47eb-9e45-7d07872f0fbc"),
				},
				Permission: Member,
			},
		},
	}
	b, _ := json.Marshal(a)
	type arguments struct {
		userID uint
	}
	type want struct {
		acl ACL
		err error
	}
	cases := map[string]struct {
		reason string
		c      shttp.Client
		args   arguments
		want   want
	}{
		"Success": {
			reason: "If status code is OK and response body is valid, ACL should be returned with no error.",
			c: &shttp.MockClient{
				DoFn: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader(b)),
						StatusCode: http.StatusOK,
					}, nil
				},
			},
			args: arguments{
				userID: 1,
			},
			want: want{
				acl: a,
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
				userID: 1,
			},
			want: want{
				acl: ACL{},
				err: errors.Wrap(errBoom, errDoACLRequest),
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
				userID: 1,
			},
			want: want{
				acl: ACL{},
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
				userID: 1,
			},
			want: want{
				acl: ACL{},
				err: errors.New(errACLResponse),
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
				userID: 1,
			},
			want: want{
				acl: ACL{},
				err: errors.Wrap(io.EOF, errInvalidACLResponseBody),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := New(*u, WithClient(tc.c))
			acl, err := c.GetACL(context.Background(), fmt.Sprintf("%d", tc.args.userID))
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nGetACL(...): -want err, +got err:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.acl, acl); diff != "" {
				t.Errorf("\n%s\nGetCL(...): -want ACL, +got ACL:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestGetACL_Robots(t *testing.T) {
	uuid1 := types.NewUUID()
	u, _ := url.Parse("https://api-private:8080")
	errBoom := errors.New("boom")
	a := ACL{
		Accounts: []AccountAccess{
			{
				ID:         1,
				Name:       "hasheddan",
				Permission: Owner,
			},
		},
		Teams: []TeamAccess{
			{
				ID: types.UUID{
					UUID: uuid.MustParse("fc5105af-e023-47eb-9e45-7d07872f0fbc"),
				},
				Permission: Member,
			},
		},
	}
	b, _ := json.Marshal(a)
	type arguments struct {
		robotID types.UUID
	}
	type want struct {
		acl ACL
		err error
	}
	cases := map[string]struct {
		reason string
		c      shttp.Client
		args   arguments
		want   want
	}{
		"Success": {
			reason: "If status code is OK and response body is valid, ACL should be returned with no error.",
			c: &shttp.MockClient{
				DoFn: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(bytes.NewReader(b)),
						StatusCode: http.StatusOK,
					}, nil
				},
			},
			args: arguments{
				robotID: uuid1,
			},
			want: want{
				acl: a,
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
				robotID: uuid1,
			},
			want: want{
				acl: ACL{},
				err: errors.Wrap(errBoom, errDoACLRequest),
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
				robotID: uuid1,
			},
			want: want{
				acl: ACL{},
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
				robotID: uuid1,
			},
			want: want{
				acl: ACL{},
				err: errors.New(errACLResponse),
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
				robotID: uuid1,
			},
			want: want{
				acl: ACL{},
				err: errors.Wrap(io.EOF, errInvalidACLResponseBody),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := New(*u, WithClient(tc.c))
			acl, err := c.GetACL(context.Background(), tc.args.robotID.String())
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nGetACL(...): -want err, +got err:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.acl, acl); diff != "" {
				t.Errorf("\n%s\nGetCL(...): -want ACL, +got ACL:\n%s", tc.reason, diff)
			}
		})
	}
}
