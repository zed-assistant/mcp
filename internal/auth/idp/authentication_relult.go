package idp

type AuthenticationResult struct {
	Email            string
	Sub              string
	IDP              string
	PendingRequestID string
}
