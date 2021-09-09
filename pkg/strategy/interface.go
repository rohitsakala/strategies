package strategy

type Strategy interface {
	Start() error
	Stop() error
}
