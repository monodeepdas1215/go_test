package config

import "github.com/sirupsen/logrus"

const (

	DbRecreate = false
	DbDebugStatus = true

	DbDriver  = "postgres"
	DbName    = "go_test"
	DbHost    = "localhost"
	DbPort    = "5432"
	DbUser    = "md5"
	DbPass    = "****"

	AppLogLevel = logrus.InfoLevel
)
