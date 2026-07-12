package main

import (
	authapi "github.com/zed-assistant/mcp/internal/api/auth_api"
	wellknownapi "github.com/zed-assistant/mcp/internal/api/well_known_api"
)

type serverDeps struct {
	authApi      *authapi.AuthApi
	wellKnownApi *wellknownapi.WellKnownApi
}
