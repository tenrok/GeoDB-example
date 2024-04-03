package api

import "geodbsvc/internal/models"

type Controller struct {
	srv models.Server
}

// NewController создаёт новый контроллер
func NewController(srv models.Server) *Controller {
	ctrl := new(Controller)
	ctrl.srv = srv
	return ctrl
}
