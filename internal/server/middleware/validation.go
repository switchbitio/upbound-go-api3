package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/pkg/errors"
)

// Copy of middleware from go-chi customized for supporting custom return values
// for OCI specification errors

// Options to customize request validation, openapi3filter specified options
// will be passed through.
type Options struct {
	Options openapi3filter.Options
}

// OCIRequestValidatorWithOptions Creates middleware to validate request by
// swagger spec. This middleware is good for net/http either since go-chi is
// 100% compatible with net/http.
func OCIRequestValidatorWithOptions(swagger *openapi3.T, options *Options) func(next http.Handler) http.Handler {
	router, err := gorillamux.NewRouter(swagger)
	if err != nil {
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// validate request
			if statusCode, err := validateRequest(r, router, options); err != nil {
				http.Error(w, err.Error(), statusCode)
				return
			}

			// serve
			next.ServeHTTP(w, r)
		})
	}
}

func validateRequest(r *http.Request, router routers.Router, options *Options) (int, error) {
	// Find route
	route, pathParams, err := router.FindRoute(r)
	if err != nil {
		return http.StatusBadRequest, err // We failed to find a matching route for the request.
	}

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}

	if options != nil {
		requestValidationInput.Options = &options.Options
	}

	if err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput); err != nil {
		switch e := errors.Cause(err).(type) { //nolint
		case *openapi3filter.RequestError:
			return getOCIResponseFromError(e)
		case *openapi3filter.SecurityRequirementsError:
			return http.StatusUnauthorized, err
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected
			// error.
			return http.StatusInternalServerError, fmt.Errorf("error validating route: %w", err)
		}
	}

	return http.StatusOK, nil
}

func getOCIResponseFromError(err *openapi3filter.RequestError) (int, error) {
	switch e := err.Err.(type) { //nolint
	case *openapi3.SchemaError:
		return http.StatusNotFound, fmt.Errorf(e.Error())
	case *openapi3filter.ParseError:
		return http.StatusBadRequest, nil
	default:
		errorLines := strings.Split(err.Error(), "\n")
		return http.StatusBadRequest, fmt.Errorf(errorLines[0])
	}
}
