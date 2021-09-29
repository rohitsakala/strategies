package authenticator

type Authenticator interface {
	GetTOTP() (string, error)
}
