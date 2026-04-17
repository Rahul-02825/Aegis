package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Logger struct {
	file *os.File
	mu   sync.Mutex 
}

var (
	Instance *Logger
	once     sync.Once
)

// InitLogger initializes the logger (singleton)
func InitLogger(filename string) *Logger {
	logpath := "./logs/app.log"
	// Ensure parent directory exists
	err := os.MkdirAll("./logs", os.ModePerm)
	if err != nil {
		panic(err)
	}
	once.Do(func() {
		f, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		Instance = &Logger{file: f}
	})
	return Instance
}

// log writes a log entry with timestamp and level
func (l *Logger) log(level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("%s [%s] %s\n", timestamp, level, msg)
	l.file.WriteString(entry)
	fmt.Print(entry) // also print to console
}

// Public methods
func (l *Logger) Info(msg string)  { l.log("INFO", msg) }
func (l *Logger) Warn(msg string)  { l.log("WARN", msg) }
func (l *Logger) Error(msg string) { l.log("ERROR", msg) }

// Close the file when shutting down
func (l *Logger) Close() {
	l.file.Close()
}
