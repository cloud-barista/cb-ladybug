package common

import (
	"github.com/cloud-barista/cb-store/config"
	"github.com/sirupsen/logrus"
)

var CBLog *logrus.Logger

// var CBStore icbs.Store

func init() {
	// cblog is a global variable.
	CBLog = config.Cblogger
	// CBStore = cbstore.GetStore()
}
