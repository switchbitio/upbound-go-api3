package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	oapifilter "github.com/getkin/kin-openapi/openapi3filter"
)

type ctxkey int

const (
	// OriginalPathKey is the key for the original request path before modification.
	// It is passed to the handler in the request context.
	OriginalPathKey ctxkey = 1

	// OriginalDomainKey is the key for the original request domain before modification.
	OriginalDomainKey ctxkey = 2
)

// OriginalPathFromContext extracts original path from the supplied context.
func OriginalPathFromContext(ctx context.Context) (string, bool) {
	p, ok := ctx.Value(OriginalPathKey).(string)
	return p, ok
}

// SetDomainInContext sets the originally requested domain in the context for use
// as needed in rewriting responses later
func SetDomainInContext(req *http.Request, domain string) *http.Request {
	return req.Clone(context.WithValue(req.Context(), OriginalDomainKey, domain))
}

// GetDomainInContext gets the originally requested domain in the context for use
// as needed in rewriting responses
func GetDomainInContext(req *http.Request) (string, bool) {
	p, ok := req.Context().Value(OriginalDomainKey).(string)
	return p, ok
}

// StripSlashes is a middleware that will match request paths with a trailing
// slash, strip it from the path and continue routing through the mux, if a
// route matches, then it will serve the handler.
func StripSlashes(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var path = r.URL.Path
		r.URL.Path = strings.TrimSuffix(path, "/")
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), OriginalPathKey, path)))
	}
	return http.HandlerFunc(fn)
}

// JSONBodyDecoder decoder function for custom content types, just copied from
// chi req_resp_decoder.go
func JSONBodyDecoder(body io.Reader, header http.Header, schema *openapi3.SchemaRef, encFn oapifilter.EncodingFn) (any, error) {
	var value any
	if err := json.NewDecoder(body).Decode(&value); err != nil {
		return nil, &oapifilter.ParseError{Kind: oapifilter.KindInvalidFormat, Cause: err}
	}
	return value, nil
}
