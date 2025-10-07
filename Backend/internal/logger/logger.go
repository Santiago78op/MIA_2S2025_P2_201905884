package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level representa el nivel de severidad del log
type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Entry representa una entrada de log
type Entry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     Level                  `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// Logger es un logger thread-safe con almacenamiento en memoria
type Logger struct {
	mu       sync.RWMutex
	entries  []Entry
	maxSize  int
	file     *os.File
	toStdout bool
}

var (
	globalLogger *Logger
	once         sync.Once
)

// Init inicializa el logger global
func Init(logFile string, maxEntries int, stdout bool) error {
	var err error
	once.Do(func() {
		globalLogger = &Logger{
			entries:  make([]Entry, 0, maxEntries),
			maxSize:  maxEntries,
			toStdout: stdout,
		}

		if logFile != "" {
			// Crear el directorio si no existe
			dir := filepath.Dir(logFile)
			if err = os.MkdirAll(dir, 0755); err != nil {
				return
			}
			globalLogger.file, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		}
	})
	return err
}

// GetLogger retorna el logger global
func GetLogger() *Logger {
	if globalLogger == nil {
		// Init con valores por defecto si no se ha inicializado
		_ = Init("", 1000, true)
	}
	return globalLogger
}

// log registra una entrada de log
func (l *Logger) log(level Level, message string, context map[string]interface{}) {
	entry := Entry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Context:   context,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Agregar a memoria (circular buffer)
	l.entries = append(l.entries, entry)
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[1:]
	}

	// Escribir a archivo si existe
	if l.file != nil {
		data, _ := json.Marshal(entry)
		l.file.Write(append(data, '\n'))
	}

	// Escribir a stdout si está habilitado
	if l.toStdout {
		fmt.Printf("[%s] %s: %s", entry.Level, entry.Timestamp.Format(time.RFC3339), entry.Message)
		if len(entry.Context) > 0 {
			contextJSON, _ := json.Marshal(entry.Context)
			fmt.Printf(" | %s", string(contextJSON))
		}
		fmt.Println()
	}
}

// Debug registra un mensaje de debug
func (l *Logger) Debug(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(LevelDebug, message, ctx)
}

// Info registra un mensaje informativo
func (l *Logger) Info(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(LevelInfo, message, ctx)
}

// Warn registra una advertencia
func (l *Logger) Warn(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(LevelWarn, message, ctx)
}

// Error registra un error
func (l *Logger) Error(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(LevelError, message, ctx)
}

// GetEntries retorna todas las entradas de log
func (l *Logger) GetEntries() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Retornar copia para evitar race conditions
	entries := make([]Entry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

// GetEntriesByLevel retorna entradas filtradas por nivel
func (l *Logger) GetEntriesByLevel(level Level) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var filtered []Entry
	for _, entry := range l.entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// GetRecentEntries retorna las últimas N entradas
func (l *Logger) GetRecentEntries(n int) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if n > len(l.entries) {
		n = len(l.entries)
	}

	start := len(l.entries) - n
	entries := make([]Entry, n)
	copy(entries, l.entries[start:])
	return entries
}

// Clear limpia todos los logs en memoria
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = make([]Entry, 0, l.maxSize)
}

// Close cierra el archivo de log si existe
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Funciones de conveniencia para usar el logger global
func Debug(message string, context ...map[string]interface{}) {
	GetLogger().Debug(message, context...)
}

func Info(message string, context ...map[string]interface{}) {
	GetLogger().Info(message, context...)
}

func Warn(message string, context ...map[string]interface{}) {
	GetLogger().Warn(message, context...)
}

func Error(message string, context ...map[string]interface{}) {
	GetLogger().Error(message, context...)
}
