package wellknownapi

import "net/http"

func (a *WellKnownApi) getAuthServerdMetadata(w http.ResponseWriter, _ *http.Request) {
	metadata := map[string]interface{}{
		"issuer":                 a.appConfig.Server.ExternalUrl,
		"authorization_endpoint": a.appConfig.Server.ExternalUrl + "/auth/authorize",
		"token_endpoint":         a.appConfig.Server.ExternalUrl + "/auth/token",

		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"none"},
		"scopes_supported":                      []string{"mcp:tools", "offline_access"},

		"client_id_metadata_document_supported": true,
	}

	writeWellKnownJSON(w, metadata)
}
