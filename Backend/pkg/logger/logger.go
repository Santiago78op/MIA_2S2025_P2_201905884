package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// Logger define los métodos comunes del logger central.
// Lo usamos en toda la app para mantener un formato uniforme.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
	Fatal(msg string, args ...any)
}

// logStd es la implementación basada en log.Stdlib.
type logStd struct {
	debug bool
	core  *log.Logger
}

// New crea un nuevo logger con o sin modo debug.
func New(debug bool) Logger {
	return &logStd{
		debug: debug,
		core:  log.New(os.Stdout, "", 0), // sin prefix; lo formateamos nosotros
	}
}

// helper para timestamp + nivel + caller
func (l *logStd) format(level, msg string, args ...any) string {
	now := time.Now().Format("2006-01-02 15:04:05")
	colored := colorize(level)
	final := fmt.Sprintf(msg, args...)
	file, line := caller(3)
	return fmt.Sprintf("%s %s [%s:%d] %s", now, colored, file, line, final)
}

func (l *logStd) Info(msg string, args ...any) {
	l.core.Println(l.format("INFO", msg, args...))
}

func (l *logStd) Warn(msg string, args ...any) {
	l.core.Println(l.format("WARN", msg, args...))
}

func (l *logStd) Error(msg string, args ...any) {
	l.core.Println(l.format("ERROR", msg, args...))
}

func (l *logStd) Debug(msg string, args ...any) {
	if l.debug {
		l.core.Println(l.format("DEBUG", msg, args...))
	}
}

func (l *logStd) Fatal(msg string, args ...any) {
	l.core.Fatalln(l.format("FATAL", msg, args...))
}

// caller obtiene el nombre corto del archivo y línea.
func caller(skip int) (string, int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "???", 0
	}
	fn := runtime.FuncForPC(pc)
	name := filepathShort(file)
	if fn != nil {
		name = fmt.Sprintf("%s", name)
	}
	return name, line
}

// filepathShort devuelve solo el nombre del archivo.
func filepathShort(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return path
	}
	return parts[len(parts)-1]
}

// colorize aplica color ANSI al nivel.
func colorize(level string) string {
	switch level {
	case "INFO":
		return "\033[1;34mINFO\033[0m" // azul
	case "WARN":
		return "\033[1;33mWARN\033[0m" // amarillo
	case "ERROR":
		return "\033[1;31mERROR\033[0m" // rojo
	case "DEBUG":
		return "\033[1;36mDEBUG\033[0m" // cyan
	case "FATAL":
		return "\033[1;35mFATAL\033[0m" // magenta
	default:
		return level
	}
}
