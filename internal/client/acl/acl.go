package acl

import (
	"context"
	"encoding/json"
	"fmt"
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
	errCreateACLRequest       = "could not create ACL request"
	errDoACLRequest           = "ACL request failed"
	errNotFound               = "could not find ACL"
	errACLResponse            = "ACL request was not successful"
	errInvalidACLResponseBody = "invalid ACL response body"
)

// Permission is a valid permission level for a user in an organizational
// unit.
type Permission string

// Owner and Member are the only accepted ACL permission levels.
const (
	Owner  Permission = "owner"
	Member Permission = "member"
)

// AccountAccess defines a user's access in an account.
type AccountAccess struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	Permission Permission `json:"perm"`
}

// TeamAccess defines a user's access in a team.
type TeamAccess struct {
	ID         types.UUID `json:"id"`
	Permission Permission `json:"perm"`
}

// ACL is the ACL that is returned by the external ACL upbound-go-api3.
type ACL struct {
	// Accounts includes all user accounts and the corresponding permission.
	Accounts []AccountAccess `json:"accounts"`
	// Teams includes all user teams and the corresponding permission.
	Teams []TeamAccess `json:"teams"`
}

// Client is an auth client.
type Client interface {
	GetACL(ctx context.Context, entityID string) (ACL, error)
}

// ExternalClient manages authentication and authorization using an external
// identity upbound-go-api3.
type ExternalClient struct {
	log    logging.Logger
	host   url.URL
	client shttp.Client
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
func New(host url.URL, opts ...ClientOpt) *ExternalClient {
	m := &ExternalClient{
		log:  logging.NewNopLogger(),
		host: host,
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

const aclPathFmt = "/v1/entities/%s/acl"

// GetACL gets an ACL for the supplied entity ID.
func (c *ExternalClient) GetACL(ctx context.Context, entityID string) (ACL, error) {
	c.host.Path = fmt.Sprintf(aclPathFmt, entityID)
	req, err := http.NewRequestWithContext(ctx, "GET", c.host.String(), nil)
	if err != nil {
		c.log.Debug(errCreateACLRequest, "error", err)
		return ACL{}, errors.Wrap(err, errCreateACLRequest)
	}
	res, err := c.client.Do(req)
	if err != nil {
		c.log.Debug(errDoACLRequest, "error", err)
		return ACL{}, errors.Wrap(err, errDoACLRequest)
	}
	defer res.Body.Close() //nolint:errcheck
	if res.StatusCode == http.StatusNotFound {
		c.log.Debug(errNotFound)
		return ACL{}, serrors.NewNotFound(errors.New(errNotFound))
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		c.log.Debug(errACLResponse, "status", res.StatusCode)
		return ACL{}, errors.New(errACLResponse)
	}
	var acl ACL
	if err := json.NewDecoder(res.Body).Decode(&acl); err != nil {
		c.log.Debug(errInvalidACLResponseBody, "error", err)
		return ACL{}, errors.Wrap(err, errInvalidACLResponseBody)
	}
	return acl, nil
}
