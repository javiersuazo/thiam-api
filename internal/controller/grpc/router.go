package grpc

import (
	"github.com/evrone/go-clean-template/pkg/logger"
	pbgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewRouter -.
func NewRouter(app *pbgrpc.Server, l logger.Interface) {
	_ = l // placeholder for future use

	reflection.Register(app)
}
