package admincommand

type AdminCommand[T any] interface {
	ToCommand() string
	ParseResponse(response string) (T, error)
}
