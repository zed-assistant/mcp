package wellknownapi

import "net/http"

func (a *WellKnownApi) getMCPProtectedResourceMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := map[string]any{
		"resource":                 a.appConfig.Server.ExternalUrl + "/mcp",
		"authorization_servers":    []string{a.appConfig.Server.ExternalUrl},
		"scopes_supported":         []string{"mcp:tools"},
		"bearer_methods_supported": []string{"header"},
	}

	a.writeWellKnownJSON(r.Context(), w, metadata)
}
