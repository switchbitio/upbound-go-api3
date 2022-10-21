package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/upbound/upbound-go-api3/internal/client/acl"
	"github.com/upbound/upbound-go-api3/internal/client/auth"
	"github.com/upbound/upbound-go-api3/internal/types"
)

func main() {
	http.ListenAndServe(":9090", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:errcheck
		if r.URL.Path == "/cookie" {
			http.SetCookie(w, &http.Cookie{
				Name:  auth.SessionCookieName,
				Value: "2",
			})
			return
		}
		if r.URL.Path == "/v1/session/token/user" {
			b, _ := json.Marshal(&auth.SessionResponse{ //nolint:errcheck,errchkjson
				UserID: 2,
			})
			w.Write(b) //nolint:errcheck
			return
		}
		if r.URL.Path == "/v1/entities/2/acl" {
			b, _ := json.Marshal(&acl.ACL{ //nolint:errcheck,errchkjson
				Accounts: []acl.AccountAccess{
					{
						ID:         2,
						Name:       "",
						Permission: acl.Owner,
					},
				},
				Teams: []acl.TeamAccess{
					{
						ID:         types.NewUUID(),
						Permission: acl.Owner,
					},
				},
			})
			w.Write(b) //nolint:errcheck
			return
		}
		if strings.HasPrefix(r.URL.Path, "/v1/accounts/") {
			type accountResponse struct {
				ID uint `json:"id"`
			}
			b, _ := json.Marshal(&accountResponse{ //nolint:errcheck,errchkjson
				ID: 2,
			})
			w.Write(b) //nolint:errcheck
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}
