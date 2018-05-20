package adaptor

type Adaptor interface {
	Init() error
	Run()
}
