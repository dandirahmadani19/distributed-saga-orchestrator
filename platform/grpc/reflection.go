package grpc

import "google.golang.org/grpc/reflection"

func EnableReflection(s *Server) {
	reflection.Register(s.Instance())
}
