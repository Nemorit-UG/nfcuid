package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// SingleInstance provides functionality to prevent multiple instances of the application
type SingleInstance struct {
	lockFile *os.File
	lockPath string
}

// NewSingleInstance creates a new SingleInstance manager
func NewSingleInstance(appName string) *SingleInstance {
	// Get appropriate temp directory based on OS
	tempDir := os.TempDir()
	lockPath := filepath.Join(tempDir, fmt.Sprintf("%s.lock", appName))

	return &SingleInstance{
		lockPath: lockPath,
	}
}

// TryLock attempts to acquire the lock to prevent multiple instances
// Returns true if lock was acquired successfully, false if another instance is running
func (si *SingleInstance) TryLock() bool {
	// Try to open/create the lock file
	file, err := os.OpenFile(si.lockPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		// If file already exists, check if the process is still running
		if os.IsExist(err) {
			return si.checkExistingInstance()
		}
		// Other error occurred
		fmt.Printf("Warning: Failed to create lock file: %v\n", err)
		return false
	}

	// Write current process ID to the lock file
	pid := os.Getpid()
	if _, err := file.WriteString(strconv.Itoa(pid)); err != nil {
		file.Close()
		os.Remove(si.lockPath)
		fmt.Printf("Warning: Failed to write PID to lock file: %v\n", err)
		return false
	}

	si.lockFile = file
	return true
}

// checkExistingInstance checks if the process in the lock file is still running
func (si *SingleInstance) checkExistingInstance() bool {
	// Read the PID from the existing lock file
	data, err := os.ReadFile(si.lockPath)
	if err != nil {
		// If we can't read the file, assume it's stale and try to remove it
		os.Remove(si.lockPath)
		return si.TryLock()
	}

	pidStr := string(data)
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		// Invalid PID in file, assume it's stale
		os.Remove(si.lockPath)
		return si.TryLock()
	}

	// Check if the process is still running
	if !si.isProcessRunning(pid) {
		// Process is not running, remove stale lock file
		os.Remove(si.lockPath)
		return si.TryLock()
	}

	// Process is still running, another instance exists
	return false
}

// isProcessRunning checks if a process with the given PID is running
func (si *SingleInstance) isProcessRunning(pid int) bool {
	// Try to find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Windows, FindProcess always succeeds, so we need to actually test it
	// Send signal 0 to test if process exists (works on Unix-like systems)
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false
	}

	return true
}

// Release releases the lock when the application is shutting down
func (si *SingleInstance) Release() {
	if si.lockFile != nil {
		si.lockFile.Close()
		si.lockFile = nil
	}

	// Remove the lock file
	if si.lockPath != "" {
		os.Remove(si.lockPath)
	}
}

// GetRunningInstanceInfo returns information about any running instance
func (si *SingleInstance) GetRunningInstanceInfo() (bool, int, error) {
	data, err := os.ReadFile(si.lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, 0, nil
		}
		return false, 0, err
	}

	pidStr := string(data)
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return false, 0, fmt.Errorf("invalid PID in lock file: %s", pidStr)
	}

	isRunning := si.isProcessRunning(pid)
	return isRunning, pid, nil
}
