package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type LogLevel string

const (
	InfoLevel  LogLevel = "INFO"
	ErrorLevel LogLevel = "ERROR"
	DebugLevel LogLevel = "DEBUG"
	WarnLevel  LogLevel = "WARNING"
)

var (
	Enable  = false
	Debug   = false
	logPath = "logs"

	red    = "\x1b[31m"
	orange = "\033[38;5;208m"
	reset  = "\x1b[37m"

	infoFile  *os.File
	errorFile *os.File
	debugFile *os.File

	logMutex sync.Mutex
)
var (
	infoPath  string
	errorPath string
	debugPath string
)

const maxLogFileSize = 5 * 1024 * 1024 // 5 MB

func InitLogger() error {
	if !Enable {
		return nil
	}

	timestamp := time.Now().Format("2006-01-02_15-04")
	os.MkdirAll(logPath, 0755)

	var err error
	infoPath = filepath.Join(logPath, fmt.Sprintf("info-%s.log", timestamp))
	infoFile, err = os.OpenFile(infoPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	errorPath = filepath.Join(logPath, fmt.Sprintf("error-%s.log", timestamp))
	errorFile, err = os.OpenFile(errorPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	debugPath = filepath.Join(logPath, fmt.Sprintf("debug-%s.log", timestamp))
	debugFile, err = os.OpenFile(debugPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	return nil
}

func CloseLogger() {
	if infoFile != nil {
		infoFile.Close()
	}
	if errorFile != nil {
		errorFile.Close()
	}
	if debugFile != nil {
		debugFile.Close()
	}
}

func Log(message string, level LogLevel) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formatted := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	logMutex.Lock()
	defer logMutex.Unlock()

	switch level {
	case InfoLevel, WarnLevel:
		fmt.Printf(orange+"["+string(level)+"]"+reset+" %s\n", message)
		if infoFile != nil {
			rotateIfNeeded(&infoFile, &infoPath)
			infoFile.WriteString(formatted)
		}
	case ErrorLevel:
		fmt.Printf(red+"["+string(level)+"]"+reset+" %s\n", message)
		if errorFile != nil {
			rotateIfNeeded(&errorFile, &errorPath)
			errorFile.WriteString(formatted)
		}
	case DebugLevel:
		if Debug {
			fmt.Printf(orange+"["+string(level)+"]"+reset+" %s\n", message)
		}
		if debugFile != nil {
			rotateIfNeeded(&debugFile, &debugPath)
			debugFile.WriteString(formatted)
		}
	}
}

func rotateIfNeeded(file **os.File, path *string) {
	info, err := (*file).Stat()
	if err != nil {
		return
	}
	if info.Size() >= maxLogFileSize {
		(*file).Close()
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		newName := fmt.Sprintf("%s-%s.log", strings.TrimSuffix(*path, ".log"), timestamp)

		if err := os.Rename(*path, newName); err != nil {
			fmt.Println("Failed to rotate log:", err)
			return
		}

		*file, err = os.OpenFile(*path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Println("Failed to reopen log file:", err)
			return
		}
	}
}

func help() string {
	return `
[Usage]
  go run main.go -h                 Provides help menu
  go run main.go -d, --debug       Runs server with debug logs
  go run main.go -l, --logs        Save logs during runtime
  go run main.go -s, --seed        Creates dummy users and posts for testing

  * Multiple flags can be used in one command (flag order can be random) *
`
}
