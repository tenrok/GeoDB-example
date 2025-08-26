package api

import "geodb-example/internal/models"

type Controller struct {
	srv models.Server
}

// NewController создаёт новый контроллер
func NewController(srv models.Server) *Controller {
	return &Controller{srv}
}
