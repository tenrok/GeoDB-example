package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/kong"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/mattn/go-colorable"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"geodb-example/internal/rotatehook"
)

type AppContext struct {
	*kong.Context
	config *viper.Viper
	logger *logrus.Logger
}

var cli struct {
	Run     RunCmd           `cmd:"" help:"Start HTTP server, listen and serve."`
	Service ServiceCmd       `cmd:"" help:"Control the system service."`
	Version kong.VersionFlag `name:"version" short:"v" help:"Print version information and quit."`
}

func main() {
	// Создаём логгер
	logger := &logrus.Logger{
		Out: colorable.NewColorableStdout(),
		Formatter: &nested.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			HideKeys:        true,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.InfoLevel,
	}

	// Определяем директории
	execPath, err := os.Executable()
	if err != nil {
		logger.Fatalf("Error: %v", err)
	}
	_, execFile := filepath.Split(execPath)

	workingDir, err := os.Getwd()
	if err != nil {
		logger.Fatalf("Error: %v", err)
	}

	// Читаем конфигурационный файл
	config := viper.New()

	config.SetDefault("service.name", "example")                                                         // Имя службы
	config.SetDefault("service.display_name", "Example Service")                                         // Отображаемое имя службы
	config.SetDefault("service.description", "Example Service")                                          // Описание службы
	config.SetDefault("server.host", "")                                                                 // Хост сервера
	config.SetDefault("server.port", 8080)                                                               // Порт сервера
	config.SetDefault("server.forwarded_for", false)                                                     // Проксируется ли запросы на сервер (используется обратный прокси, добавляющий заголовок X-Forwarded-For)?
	config.SetDefault("server.trusted_proxies", []string{"127.0.0.1"})                                   // Доверенные прокси, которые могут устанавливать заголовок X-Forwarded-For
	config.SetDefault("log.enabled", false)                                                              // Вести log-файл?
	config.SetDefault("log.level", "info")                                                               //
	config.SetDefault("log.path", filepath.Join(workingDir, "logs", "example.log"))                      // Путь до log-файла
	config.SetDefault("log.max_size", 5)                                                                 //
	config.SetDefault("log.max_age", 30)                                                                 //
	config.SetDefault("log.max_backups", 10)                                                             //
	config.SetDefault("log.local_time", true)                                                            //
	config.SetDefault("log.compress", true)                                                              //
	config.SetDefault("database.dsn", fmt.Sprintf("file:%s", filepath.Join(workingDir, "GeoDB.sqlite"))) // Путь до базы данных SQLite

	config.SetConfigName("example")
	config.SetConfigType("yaml")

	config.AddConfigPath(filepath.Join(workingDir, "configs"))
	switch runtime.GOOS {
	case "linux":
		config.AddConfigPath("/etc/example")
	case "windows":
		config.AddConfigPath(filepath.Join(os.Getenv("PROGRAMDATA"), "Example Service"))
	}

	if err := config.ReadInConfig(); err != nil {
		logger.Fatalf("Error: %v", err)
	}

	// Настраиваем логирование
	level, err := logrus.ParseLevel(config.GetString("log.level"))
	if err != nil {
		logger.Fatalln(err.Error())
	}
	logger.SetLevel(level)

	hook := rotatehook.NewRotateHook(&rotatehook.Config{
		Filename:   config.GetString("log.path"),
		MaxSize:    config.GetInt("log.max_size"),
		MaxAge:     config.GetInt("log.max_age"),
		MaxBackups: config.GetInt("log.max_backups"),
		LocalTime:  config.GetBool("log.local_time"),
		Compress:   config.GetBool("log.compress"),
		Formatter: &nested.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			HideKeys:        true,
			NoColors:        true,
		},
		Level:   level,
		Enabled: config.GetBool("log.enabled"),
	})

	logger.AddHook(hook)

	// Переменные параметров командной строки
	ctx := kong.Parse(&cli,
		kong.Name(execFile),
		kong.Description("Example Service"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.Vars{
			"version": "1.0.0",
		},
	)

	app := &AppContext{Context: ctx}
	app.config = config
	app.logger = logger

	err = ctx.Run(app)
	ctx.FatalIfErrorf(err)
}
