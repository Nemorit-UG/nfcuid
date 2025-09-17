package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogManager handles file-based logging with rotation
type LogManager struct {
	logFile     *os.File
	logger      *log.Logger
	logFilePath string
	enabled     bool
}

// NewLogManager creates a new log manager with file output
func NewLogManager() *LogManager {
	lm := &LogManager{
		enabled: true,
	}

	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create logs directory: %v\n", err)
		lm.enabled = false
		return lm
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	lm.logFilePath = filepath.Join(logsDir, fmt.Sprintf("nfcuid_%s.log", timestamp))

	// Open log file
	if err := lm.openLogFile(); err != nil {
		fmt.Printf("Warning: Failed to open log file: %v\n", err)
		lm.enabled = false
		return lm
	}

	// Create a multi-writer to write to both console and file
	multiWriter := io.MultiWriter(os.Stdout, lm.logFile)
	lm.logger = log.New(multiWriter, "", log.LstdFlags)

	lm.LogInfo("Log file created", "path", lm.logFilePath)

	return lm
}

func (lm *LogManager) openLogFile() error {
	var err error
	lm.logFile, err = os.OpenFile(lm.logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	return err
}

// LogInfo logs an informational message
func (lm *LogManager) LogInfo(message string, keyValuePairs ...string) {
	if !lm.enabled {
		fmt.Printf("[INFO] %s\n", message)
		return
	}

	logMsg := fmt.Sprintf("[INFO] %s", message)
	for i := 0; i < len(keyValuePairs); i += 2 {
		if i+1 < len(keyValuePairs) {
			logMsg += fmt.Sprintf(" %s=%s", keyValuePairs[i], keyValuePairs[i+1])
		}
	}

	lm.logger.Println(logMsg)
}

// LogError logs an error message
func (lm *LogManager) LogError(message string, err error, keyValuePairs ...string) {
	if !lm.enabled {
		fmt.Printf("[ERROR] %s: %v\n", message, err)
		return
	}

	logMsg := fmt.Sprintf("[ERROR] %s", message)
	if err != nil {
		logMsg += fmt.Sprintf(": %v", err)
	}
	for i := 0; i < len(keyValuePairs); i += 2 {
		if i+1 < len(keyValuePairs) {
			logMsg += fmt.Sprintf(" %s=%s", keyValuePairs[i], keyValuePairs[i+1])
		}
	}

	lm.logger.Println(logMsg)
}

// LogWarning logs a warning message
func (lm *LogManager) LogWarning(message string, keyValuePairs ...string) {
	if !lm.enabled {
		fmt.Printf("[WARNING] %s\n", message)
		return
	}

	logMsg := fmt.Sprintf("[WARNING] %s", message)
	for i := 0; i < len(keyValuePairs); i += 2 {
		if i+1 < len(keyValuePairs) {
			logMsg += fmt.Sprintf(" %s=%s", keyValuePairs[i], keyValuePairs[i+1])
		}
	}

	lm.logger.Println(logMsg)
}

// LogCardRead logs a successful card read
func (lm *LogManager) LogCardRead(uid string, device string) {
	lm.LogInfo("Card read successful", "uid", uid, "device", device)
}

// LogDeviceStatus logs device connection/disconnection
func (lm *LogManager) LogDeviceStatus(device string, status string) {
	lm.LogInfo("Device status change", "device", device, "status", status)
}

// LogScanningStatus logs scanning state changes
func (lm *LogManager) LogScanningStatus(status string, device string) {
	lm.LogInfo("Scanning status", "status", status, "device", device)
}

// GetLogFilePath returns the current log file path
func (lm *LogManager) GetLogFilePath() string {
	return lm.logFilePath
}

// Close closes the log file
func (lm *LogManager) Close() {
	if lm.logFile != nil {
		lm.LogInfo("Closing log file")
		lm.logFile.Close()
	}
}

// ListLogFiles returns a list of all log files in the logs directory
func (lm *LogManager) ListLogFiles() ([]string, error) {
	logsDir := "logs"
	files, err := filepath.Glob(filepath.Join(logsDir, "nfcuid_*.log"))
	if err != nil {
		return nil, err
	}
	return files, nil
}

// GetLogFileContent returns the content of a specific log file
func (lm *LogManager) GetLogFileContent(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
