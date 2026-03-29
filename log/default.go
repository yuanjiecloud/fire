package log

import (
	"fmt"
	"os"
	"runtime/debug"
)

var (
	Verbose    = false
	PrintStack = false
)

func Debug(data ...interface{}) {
	if Verbose {
		fmt.Println(data...)
	}
}

func Fatal(data ...interface{}) {
	fmt.Println(data...)
	if PrintStack {
		debug.PrintStack()
	}
	os.Exit(-1)
}

func Error(data ...interface{}) {
	os.Stderr.WriteString(fmt.Sprint(data...) + "\n")
}

func Info(data ...interface{}) {
	fmt.Println(data...)
}

func CheckAndFatal(err error) {
	if err != nil {
		Fatal(err)
	}
}
