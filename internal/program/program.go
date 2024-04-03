package program

import (
	"fmt"

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"geodbsvc/internal/database"
	"geodbsvc/internal/server"
)

type Program struct {
	exit   chan struct{}
	cfg    *viper.Viper
	logger *logrus.Logger
	srv    *server.Server
}

// New
func New(cfg *viper.Viper, db database.DB, logger *logrus.Logger) *Program {
	p := new(Program)
	p.cfg = cfg
	p.logger = logger
	p.srv = server.New(cfg, db, logger)
	return p
}

// Start вызывается при запуске службы
func (p *Program) Start(s service.Service) error {
	p.exit = make(chan struct{})

	// Основная работа программы
	go func() {
		addr := fmt.Sprintf("%s:%d", p.cfg.GetString("server.host"), p.cfg.GetInt("server.port"))
		p.srv.Run(addr)
		p.logger.Printf("Server is running at %s", addr)
		<-p.exit
		p.srv.Shutdown()
	}()

	return nil
}

// Stop вызывается при остановке службы
func (p *Program) Stop(s service.Service) error {
	close(p.exit)
	return nil
}
