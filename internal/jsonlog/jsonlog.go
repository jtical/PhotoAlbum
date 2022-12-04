//Filename: internal/jsonlog/jsonlog.go

package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// we can have diffrent severity levels of logging entries
type Level int8

// specify our levels. Levels start at zero
const (
	LevelInfo  Level = iota //value will be 0 here
	LevelError              //value will be 1
	LevelFatal              //value will be 2
	LevelOff                //value will be 3
)

// the severity levels as a human readable friendly format
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// define a custome logger
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// The New() function creates a new instance of logger
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// helper methods
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1)
}

func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	//ensure that the severity level is atleast the minimum
	if level < l.minLevel {
		return 0, nil
	}
	//create a struct to hold log entry data
	data := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}
	//should we include the stack trace?
	if level >= LevelError {
		data.Trace = string(debug.Stack())
	}
	//encode the log entry to JSON
	var entry []byte
	entry, err := json.Marshal(data)
	if err != nil {
		entry = []byte(LevelError.String() + ": unable to marshal log message: " + err.Error())
	}
	//prepare to write the log entry
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out.Write(append(entry, '\n'))
}

// Implement the io.Writer interface
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
