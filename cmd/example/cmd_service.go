package main

import (
	"runtime"
	"strings"

	"github.com/kardianos/service"

	"example/internal/utils"
)

type ServiceCmd struct {
	Action string `arg:"" required:"" help:"start, stop, restart, install, uninstall"`
}

type DummyProgram struct{}

func (DummyProgram) Start(service.Service) error { return nil }
func (DummyProgram) Stop(service.Service) error  { return nil }

func (c *ServiceCmd) Run(app *AppContext) error {
	// Создаём программу "постышку" для управления службой
	prg := &DummyProgram{}

	// Задаём настройки для службы
	options := make(service.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"

	svcConfig := &service.Config{
		Name:        app.config.GetString("service.name"),
		DisplayName: app.config.GetString("service.display_name"),
		Description: app.config.GetString("service.description"),
		Option:      options,
		Arguments:   []string{"run"},
	}
	if runtime.GOOS == "linux" {
		svcConfig.Dependencies = []string{
			"Requires=network.target",
			"After=network-online.target syslog.target",
		}
	}

	// Создаём службу
	svc, err := service.New(prg, svcConfig)
	if err != nil {
		app.Fatalf("%s", err)
	}

	if !utils.Contains(service.ControlAction[:], c.Action, true) {
		app.Fatalf("valid actions is %s", strings.Join(service.ControlAction[:], ", "))
	} else if err := service.Control(svc, c.Action); err != nil {
		app.Fatalf("%s", err)
	}

	return nil
}
