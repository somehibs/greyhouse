package log

import (
	"log"
)

func Print(v ...interface{}) {
	log.Print(v...)
}

func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func Warn(v ...interface{}) {
	Print(v...)
}

func Warnf(format string, v ...interface{}) {
	Printf(format, v...)
}
