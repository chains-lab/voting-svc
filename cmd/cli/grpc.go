package cli

import (
	"net"

	"google.golang.org/grpc"
)

// MyRegistrar — это наш “мост” к реальному *grpc.Server
type MyRegistrar struct {
	srv *grpc.Server
}

// NewRegistrar создаёт экземпляр со спрингованным grpc.Server
func NewRegistrar() *MyRegistrar {
	return &MyRegistrar{
		srv: grpc.NewServer(),
	}
}

// RegisterService нужным образом проксирует вызов к grpc.Server
func (r *MyRegistrar) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	r.srv.RegisterService(desc, impl)
}

// Serve — удобный метод, чтобы стартануть HTTP/2 слушатель
func (r *MyRegistrar) Serve(lis net.Listener) error {
	return r.srv.Serve(lis)
}

// Start your CLI command here
