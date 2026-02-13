package grpc

import (
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	s *grpc.Server
	l net.Listener
}

func New(port string, opts ...grpc.ServerOption) (*Server, error) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(opts...)

	return &Server{
		s: s,
		l: lis,
	}, nil
}

func (g *Server) Instance() *grpc.Server {
	return g.s
}

func (g *Server) Start() error {
	return g.s.Serve(g.l)
}

func (g *Server) Stop() {
	g.s.GracefulStop()
}
