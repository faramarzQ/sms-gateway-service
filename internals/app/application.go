package app

type Application interface {
	Build() error
	Run() error
}
