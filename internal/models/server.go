package models

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"example/internal/database"
)

type Server interface {
	Config() *viper.Viper
	DB() database.DB
	Logger() *logrus.Logger
}
