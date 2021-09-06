package strategy

type Strategy interface {
	Start() error
	CheckMargin()
}
