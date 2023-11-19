package libgobuster

import (
	"log"
	"os"

	"github.com/fatih/color"
)

type Logger struct {
	log      *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
	warnLog  *log.Logger
	infoLog  *log.Logger
	debug    bool
}

func NewLogger(debug bool) *Logger {
	return &Logger{
		log:      log.New(os.Stdout, "", 0),
		errorLog: log.New(os.Stderr, color.New(color.FgRed).Sprint("[ERROR] "), 0),
		debugLog: log.New(os.Stderr, color.New(color.FgBlue).Sprint("[DEBUG] "), 0),
		warnLog:  log.New(os.Stderr, color.New(color.FgYellow).Sprint("[WARN] "), 0),
		infoLog:  log.New(os.Stderr, color.New(color.FgCyan).Sprint("[INFO] "), 0),
		debug:    debug,
	}
}

func (l Logger) Debug(v ...any) {
	if !l.debug {
		return
	}
	l.debugLog.Print(v...)
}

func (l Logger) Debugf(format string, v ...any) {
	if !l.debug {
		return
	}
	l.debugLog.Printf(format, v...)
}

func (l Logger) Warn(v ...any) {
	l.warnLog.Print(v...)
}

func (l Logger) Warnf(format string, v ...any) {
	l.warnLog.Printf(format, v...)
}

func (l Logger) Info(v ...any) {
	l.infoLog.Print(v...)
}

func (l Logger) Infof(format string, v ...any) {
	l.infoLog.Printf(format, v...)
}

func (l Logger) Print(v ...any) {
	l.log.Print(v...)
}

func (l Logger) Printf(format string, v ...any) {
	l.log.Printf(format, v...)
}

func (l Logger) Println(v ...any) {
	l.log.Println(v...)
}

func (l Logger) Error(v ...any) {
	l.errorLog.Print(v...)
}

func (l Logger) Errorf(format string, v ...any) {
	l.errorLog.Printf(format, v...)
}

func (l Logger) Fatal(v ...any) {
	l.errorLog.Fatal(v...)
}

func (l Logger) Fatalf(format string, v ...any) {
	l.errorLog.Fatalf(format, v...)
}

func (l Logger) Fatalln(v ...any) {
	l.errorLog.Fatalln(v...)
}
