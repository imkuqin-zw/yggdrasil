package application

type InternalServer interface {
	Serve() error
	Stop() error
}
