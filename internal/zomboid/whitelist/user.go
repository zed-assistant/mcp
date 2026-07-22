package whitelist

import "time"

type User struct {
	ID             int
	World          string
	Username       string
	LastConnection *time.Time
	RoleID         int
	RoleName       string
	SteamID        string
}
