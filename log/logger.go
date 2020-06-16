package log

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

func logger() *logrus.Logger {
	log := logrus.New()
	log.SetReportCaller(true)
	debug := flag.Bool("debug", false, "for debugging")
	flag.Parse()
	logrus.SetOutput(os.Stdout)
	if *debug {
		log.Formatter = &logrus.TextFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				repopath := fmt.Sprintf(os.ExpandEnv("$PWD"))
				filename := strings.Replace(f.File, repopath, "", -1)
				funcname := strings.Split(f.Function, "/")
				return fmt.Sprintf("%s()", funcname[len(funcname)-1]), fmt.Sprintf("%s:%d", filename, f.Line)
			},
		}
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
	return log
}

var Logger = logger()
