package browser

type Browser interface {
	Start() error
	Stop() error
}
