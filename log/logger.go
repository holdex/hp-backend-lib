package liblog

import (
	"io/ioutil"
	"os"

	"google.golang.org/grpc/grpclog"
)

var logger grpclog.LoggerV2

func init() {
	errorW := ioutil.Discard
	warningW := ioutil.Discard
	infoW := ioutil.Discard

	logLevel := os.Getenv("LOG_SEVERITY_LEVEL")
	switch logLevel {
	case "", "ERROR": // If env is unset, set level to ERROR.
		errorW = os.Stderr
	case "WARNING":
		warningW = os.Stderr
	case "INFO":
		infoW = os.Stderr
	}

	logger = grpclog.NewLoggerV2(infoW, warningW, errorW)
}

func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warning(args ...interface{}) {
	logger.Warning(args...)
}

func Warningf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

func Warningln(args ...interface{}) {
	logger.Warningln(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
	os.Exit(1)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
	os.Exit(1)
}

func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
	os.Exit(1)
}

type Redactor interface {
	Redact() string
}
