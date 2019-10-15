package gcloudsql

import (
	"io/ioutil"
	"log"
)

type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

var infoLogger Logger
var debugLogger Logger

func init() {
	infoLogger = log.New(ioutil.Discard, "", 0)
	debugLogger = log.New(ioutil.Discard, "", 0)
}

func SetInfoLogger(logger Logger) {
	infoLogger = logger
}

func SetDebugLogger(logger Logger) {
	debugLogger = logger
}
