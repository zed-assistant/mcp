package wellknownapi

import "net/http"

func (a *WellKnownApi) getMCPProtectedResourceMetadata(w http.ResponseWriter, _ *http.Request) {
	metadata := map[string]interface{}{
		"resource":                 a.appConfig.Server.ExternalUrl + "/mcp",
		"authorization_servers":    []string{a.appConfig.Server.ExternalUrl},
		"scopes_supported":         []string{"mcp:tools"},
		"bearer_methods_supported": []string{"header"},
	}

	writeWellKnownJSON(w, metadata)
}
