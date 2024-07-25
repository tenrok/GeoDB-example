package main

import (
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/kardianos/service"
	_ "github.com/mattn/go-sqlite3"

	sqlitedb "example/internal/database/sqlite"
	"example/internal/program"
	"example/migrations"
)

type RunCmd struct {
}

func (c *RunCmd) Run(app *AppContext) error {
	app.logger.Infof("Конфигурационный файл: %q", app.config.ConfigFileUsed())
	app.logger.Infof("Уровень логирования: %s", app.logger.Level.String())

	// Открываем БД
	db, err := sqlitedb.Open(app.config.GetString("database.dsn"))
	if err != nil {
		app.logger.Fatalf("Возникла ошибка при открытии БД: %v", err)
	}
	defer db.Close()

	// Выполняем миграцию БД
	ds, err := iofs.New(migrations.FS, "sqlite")
	if err != nil {
		app.logger.Fatalln(err)
	}
	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		app.logger.Fatalln(err)
	}
	migrator, err := migrate.NewWithInstance("iofs", ds, "sqlite3", driver)
	if err != nil {
		app.logger.Fatalln(err)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		app.logger.Fatalln(err)
	}

	var ver string
	if err := db.QueryRow("select version from schema_migrations limit 1").Scan(&ver); err != nil {
		app.logger.Fatalln(err)
	}
	db.SetVersion(ver)

	// Создаём программу
	prg := program.New(app.config, db, app.logger)

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
		app.logger.Fatalln(err)
	}

	errs := make(chan error, 5)

	// Открываем системный логгер
	svcLogger, err := svc.Logger(errs)
	if err != nil {
		app.logger.Fatalln(err)
	}

	// Вывод ошибок
	go func() {
		for {
			if err := <-errs; err != nil {
				app.logger.Errorln(err)
			}
		}
	}()

	if err := svc.Run(); err != nil {
		svcLogger.Error(err)
	}

	return nil
}
