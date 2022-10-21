package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	shttp "github.com/upbound/upbound-go-api3/internal/client/http"
	serrors "github.com/upbound/upbound-go-api3/internal/errors"
	"github.com/upbound/upbound-go-api3/internal/types"
)

const (
	errInvalidSessionRequestBody  = "invalid session request body"
	errCreateSessionRequest       = "could not create session request"
	errDoSessionRequest           = "session request failed"
	errNotFound                   = "could not find session"
	errSessionResponse            = "session request was not successful"
	errInvalidSessionResponseBody = "invalid session response body"
)

// SessionCookieName is the name of the cookie that contains an Upbound session
// token.
const SessionCookieName = "SID"

type authctxkey int

var (
	// UserKey is used to identify a user in a context.
	UserKey authctxkey
	// RobotKey is used to identify a robot in a context.
	RobotKey authctxkey
)

// UserIDFromContext extracts the user ID from the supplied context.
func UserIDFromContext(ctx context.Context) (uint, bool) {
	p, ok := ctx.Value(UserKey).(uint)
	return p, ok
}

// RobotIDFromContext extracts the robot ID from the supplied context.
func RobotIDFromContext(ctx context.Context) (types.UUID, bool) {
	p, ok := ctx.Value(RobotKey).(types.UUID)
	return p, ok
}

// SessionRequest is the request body for session information.
type SessionRequest struct {
	JWTToken string `json:"jwtToken"`
}

// SessionResponse is the response body for session information.
type SessionResponse struct {
	UserID uint `json:"userID"`
}

// An Entity is a type of entity that can authenticate using an API token.
type Entity string

// Types of entities.
const (
	User  Entity = "user"
	Robot Entity = "robot"
)

// EntityResponse is the response body for the entity information.
type EntityResponse struct {
	ID         types.UUID `json:"id"`
	Name       string     `json:"name"`
	OwnerType  string     `json:"ownerType"`
	OwnerID    string     `json:"ownerID"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

// Client is an auth client.
type Client interface {
	GetUserID(ctx context.Context, token string) (uint, error)
	GetEntityID(ctx context.Context, token string) (Entity, string, error)
}

// ExternalClient manages authentication and authorization using an external
// identity upbound-go-api3.
type ExternalClient struct {
	log         logging.Logger
	authHost    url.URL
	privateHost url.URL
	client      shttp.Client
}

// ClientOpt is an option that modifies an external client.
type ClientOpt func(m *ExternalClient)

// WithLogger sets a logger for the external client.
func WithLogger(log logging.Logger) ClientOpt {
	return func(c *ExternalClient) {
		c.log = log
	}
}

// WithClient sets the HTTP client for the external client.
func WithClient(client shttp.Client) ClientOpt {
	return func(c *ExternalClient) {
		c.client = client
	}
}

// New constructs a new external auth client.
func New(authHost, privateHost url.URL, opts ...ClientOpt) *ExternalClient {
	m := &ExternalClient{
		log:         logging.NewNopLogger(),
		authHost:    authHost,
		privateHost: privateHost,
		client: &http.Client{
			Timeout: 10 * time.Second,
			// TODO(hasheddan): consider passing base transport with more
			// granular timeouts.
			Transport: otelhttp.NewTransport(nil),
		},
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

const (
	sessionTokenPath = "/v1/session/token/user"
	apiTokenPath     = "/v1/tokens/validate"
)

// GetUserID gets the UserID from a session token.
func (c *ExternalClient) GetUserID(ctx context.Context, token string) (uint, error) {
	b, err := json.Marshal(&SessionRequest{
		JWTToken: token,
	})
	if err != nil {
		c.log.Debug(errInvalidSessionRequestBody, "error", err)
		return 0, errors.Wrap(err, errInvalidSessionRequestBody)
	}
	c.authHost.Path = sessionTokenPath
	req, err := http.NewRequestWithContext(ctx, "POST", c.authHost.String(), bytes.NewReader(b))
	if err != nil {
		c.log.Debug(errCreateSessionRequest, "error", err)
		return 0, errors.Wrap(err, errCreateSessionRequest)
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		c.log.Debug(errDoSessionRequest, "error", err)
		return 0, errors.Wrap(err, errDoSessionRequest)
	}
	defer res.Body.Close() //nolint:errcheck
	if res.StatusCode == http.StatusNotFound {
		c.log.Debug(errNotFound)
		return 0, serrors.NewNotFound(errors.New(errNotFound))
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		c.log.Debug(errSessionResponse, "status", res.StatusCode)
		return 0, errors.New(errSessionResponse)
	}
	var session SessionResponse
	if err := json.NewDecoder(res.Body).Decode(&session); err != nil {
		c.log.Debug(errInvalidSessionResponseBody, "error", err)
		return 0, errors.Wrap(err, errInvalidSessionResponseBody)
	}
	return session.UserID, nil
}

// GetEntityID gets the entity for the API token.
func (c *ExternalClient) GetEntityID(ctx context.Context, token string) (Entity, string, error) {
	b, err := json.Marshal(&SessionRequest{
		JWTToken: token,
	})
	if err != nil {
		c.log.Debug(errInvalidSessionRequestBody, "error", err)
		return "", "", errors.Wrap(err, errInvalidSessionRequestBody)
	}
	c.privateHost.Path = apiTokenPath
	req, err := http.NewRequestWithContext(ctx, "POST", c.privateHost.String(), bytes.NewReader(b))
	if err != nil {
		c.log.Debug(errCreateSessionRequest, "error", err)
		return "", "", errors.Wrap(err, errCreateSessionRequest)
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		c.log.Debug(errDoSessionRequest, "error", err)
		return "", "", errors.Wrap(err, errDoSessionRequest)
	}
	defer res.Body.Close() //nolint:errcheck
	if res.StatusCode == http.StatusNotFound {
		c.log.Debug(errNotFound)
		return "", "", serrors.NewNotFound(errors.New(errNotFound))
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		c.log.Debug(errSessionResponse, "status", res.StatusCode)
		return "", "", errors.New(errSessionResponse)
	}
	var tokenRes EntityResponse
	if err := json.NewDecoder(res.Body).Decode(&tokenRes); err != nil {
		c.log.Debug(errInvalidSessionResponseBody, "error", err)
		return "", "", errors.Wrap(err, errInvalidSessionResponseBody)
	}
	return Entity(tokenRes.OwnerType), tokenRes.OwnerID, nil
}
