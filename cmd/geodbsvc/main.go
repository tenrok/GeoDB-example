package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fsnotify/fsnotify"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/kardianos/service"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"

	sqlitedb "geodbsvc/internal/database/sqlite"
	"geodbsvc/internal/loggerx"
	"geodbsvc/internal/program"
	"geodbsvc/internal/utils"
	"geodbsvc/migrations"
)

func main() {
	// Парсируем командную строку
	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()

	// Определяем директории
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	execDir, _ := filepath.Split(execPath)

	// Читаем конфигурационный файл
	cfg := viper.New()

	cfg.SetDefault("service.name", "geodbsvc")                                                     // Имя службы
	cfg.SetDefault("service.display_name", "GeoDB Service")                                        // Отображаемое имя службы
	cfg.SetDefault("service.description", "GeoDB Service")                                         // Описание службы
	cfg.SetDefault("server.host", "")                                                              // Хост сервера
	cfg.SetDefault("server.port", 8080)                                                            // Порт сервера
	cfg.SetDefault("server.forwarded_for", false)                                                  // Проксируется ли запросы на сервер (используется обратный прокси, добавляющий заголовок X-Forwarded-For)?
	cfg.SetDefault("server.trusted_proxies", []string{"127.0.0.1"})                                // Доверенные прокси, которые могут устанавливать заголовок X-Forwarded-For
	cfg.SetDefault("log.enabled", false)                                                           // Вести log-файл?
	cfg.SetDefault("log.file", filepath.Join(execDir, "logs", "geodbsvc.log"))                     // Путь до log-файла
	cfg.SetDefault("database.dsn", fmt.Sprintf("file:%s", filepath.Join(execDir, "GeoDB.sqlite"))) // Путь до базы данных SQLite

	cfg.SetConfigName("geodbsvc")
	cfg.SetConfigType("yaml")

	switch runtime.GOOS {
	case "linux":
		cfg.AddConfigPath("/etc/geodbsvc")
	case "windows":
		cfg.AddConfigPath(filepath.Join(os.Getenv("PROGRAMDATA"), "GeoDB-Service"))
	}
	cfg.AddConfigPath(filepath.Join(execDir, "configs"))

	if err := cfg.ReadInConfig(); err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Настраиваем логирование
	logx := loggerx.New(
		&lumberjack.Logger{
			Filename:   cfg.GetString("log.file"),
			MaxSize:    5,
			MaxAge:     30,
			MaxBackups: 10,
			LocalTime:  true,
			Compress:   true,
		},
		cfg.GetBool("log.enabled"),
	)

	multiWriter := io.MultiWriter(os.Stdout, logx)

	logger := &logrus.Logger{
		Out: multiWriter,
		Formatter: &logrus.TextFormatter{
			ForceColors:     true,
			DisableColors:   false,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.InfoLevel,
	}

	// Открываем БД
	db, err := sqlitedb.Open(cfg.GetString("database.dsn"))
	if err != nil {
		logger.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Выполняем миграцию БД
	ds, err := iofs.New(migrations.FS, "sqlite")
	if err != nil {
		logger.Fatalf("Error: %v", err)
	}
	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		logger.Fatalf("Error: %v", err)
	}
	migrator, err := migrate.NewWithInstance("iofs", ds, "sqlite3", driver)
	if err != nil {
		logger.Fatalf("Error: %v", err)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatalf("Error: %v", err)
	}

	// Создаём программу
	prg := program.New(cfg, db, logger)

	// Задаём настройки для службы
	options := make(service.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"

	svcConfig := &service.Config{
		Name:        cfg.GetString("service.name"),
		DisplayName: cfg.GetString("service.display_name"),
		Description: cfg.GetString("service.description"),
		Option:      options,
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
		logger.Fatalf("Error: %v", err)
	}

	errs := make(chan error, 5)

	// Открываем системный логгер
	svcLogger, err := svc.Logger(errs)
	if err != nil {
		logger.Fatalf("Error: %v", err)
	}

	// Вывод ошибок
	go func() {
		for {
			if err := <-errs; err != nil {
				logger.Errorf("Error: %v", err)
			}
		}
	}()

	// Управление службой
	if len(*svcFlag) != 0 {
		if !utils.Contains(service.ControlAction[:], *svcFlag, true) {
			fmt.Fprintf(os.Stdout, "Valid actions: %q\n", service.ControlAction)
		} else if err := service.Control(svc, *svcFlag); err != nil {
			fmt.Fprintln(os.Stdout, err)
		}
		return
	}

	logger.Infoln(`Used config file "` + cfg.ConfigFileUsed() + `"`)

	// Следим за изменениями конфигурационного файла
	cfg.OnConfigChange(func(e fsnotify.Event) {
		logger.Infoln("Config file changed:", e.Name)
		logx.SetEnabled(cfg.GetBool("log.enabled"))
	})
	cfg.WatchConfig()

	// Запускаем службу
	if err := svc.Run(); err != nil {
		svcLogger.Error(err)
	}
}
