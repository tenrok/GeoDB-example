package models

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"geodbsvc/internal/database"
)

type Server interface {
	GetConfig() *viper.Viper
	GetDB() database.DB
	GetLogger() *logrus.Logger
}
