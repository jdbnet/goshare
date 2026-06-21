package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var std *log.Logger

func init() {
	std = log.New(os.Stdout, "", 0)
}

func Info(msg string, args ...interface{}) {
	output("INFO", fmt.Sprintf(msg, args...))
}

func Error(msg string, args ...interface{}) {
	output("ERROR", fmt.Sprintf(msg, args...))
}

func Warn(msg string, args ...interface{}) {
	output("WARN", fmt.Sprintf(msg, args...))
}

func output(level, msg string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	std.Printf("[%s] [%s] %s\n", timestamp, level, msg)
}
