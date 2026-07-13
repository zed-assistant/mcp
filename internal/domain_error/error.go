package domainerror

type ErrorCode string

const (
	InvalidInput         ErrorCode = "INVALID_INPUT"
	NotFound             ErrorCode = "NOT_FOUND"
	InvalidInternalState ErrorCode = "INVALID_INTERNAL_STATE"
	NotAllowed           ErrorCode = "NOT_ALLOWED"
	PreconditionFailed   ErrorCode = "PRECONDITION_FAILED"
	Conflict             ErrorCode = "CONFLICT"
)

type DomainError struct {
	InternalMessage string
	PublicMessage   string
	InternalCode    ErrorCode
	Err             error
}

func (e DomainError) Error() string {
	return e.InternalMessage
}

func (e DomainError) Unwrap() error {
	return e.Err
}
