package v1

import (
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/rabbitmq/rmq_rpc/server"
)

// NewRouter -.
func NewRouter(l logger.Interface) map[string]server.CallHandler {
	_ = l // placeholder for future use

	routes := make(map[string]server.CallHandler)

	return routes
}
