package main

import (
	"fmt"
	"path/filepath"
)

// UIStatus represents the current status of the application for future UI integration
type UIStatus struct {
	IsScanning       bool     `json:"is_scanning"`
	DeviceName       string   `json:"device_name"`
	DeviceIndex      int      `json:"device_index"`
	AvailableDevices []string `json:"available_devices"`
	LastCardUID      string   `json:"last_card_uid"`
	LastError        string   `json:"last_error"`
	LogFilePath      string   `json:"log_file_path"`
	Status           string   `json:"status"`
}

// UIManager handles UI-related functionality and prepares for future web UI integration
type UIManager struct {
	status     *UIStatus
	logManager *LogManager
}

// NewUIManager creates a new UI manager
func NewUIManager(logManager *LogManager) *UIManager {
	return &UIManager{
		status: &UIStatus{
			IsScanning:       false,
			DeviceName:       "",
			DeviceIndex:      0,
			AvailableDevices: []string{},
			LastCardUID:      "",
			LastError:        "",
			LogFilePath:      logManager.GetLogFilePath(),
			Status:           "Initializing",
		},
		logManager: logManager,
	}
}

// UpdateStatus updates the current application status
func (ui *UIManager) UpdateStatus(status string) {
	ui.status.Status = status
	ui.logManager.LogInfo("UI Status updated", "status", status)
}

// SetDeviceInfo sets the device information
func (ui *UIManager) SetDeviceInfo(deviceName string, deviceIndex int, availableDevices []string) {
	ui.status.DeviceName = deviceName
	ui.status.DeviceIndex = deviceIndex
	ui.status.AvailableDevices = availableDevices
	ui.logManager.LogInfo("Device info updated", "device", deviceName, "index", fmt.Sprintf("%d", deviceIndex))
}

// SetScanningState sets whether the application is currently scanning
func (ui *UIManager) SetScanningState(isScanning bool) {
	ui.status.IsScanning = isScanning
	if isScanning {
		ui.status.Status = "Scanning for cards"
	} else {
		ui.status.Status = "Ready"
	}
	ui.logManager.LogScanningStatus(ui.status.Status, ui.status.DeviceName)
}

// SetLastCardUID sets the last successful card UID
func (ui *UIManager) SetLastCardUID(uid string) {
	ui.status.LastCardUID = uid
	ui.logManager.LogInfo("Last card UID updated", "uid", uid)
}

// SetLastError sets the last error message
func (ui *UIManager) SetLastError(error string) {
	ui.status.LastError = error
}

// GetStatus returns the current application status (for future API endpoint)
func (ui *UIManager) GetStatus() *UIStatus {
	// Update log file path in case it changed
	ui.status.LogFilePath = ui.logManager.GetLogFilePath()
	return ui.status
}

// GetLogFiles returns a list of available log files (for future UI log viewer)
func (ui *UIManager) GetLogFiles() ([]string, error) {
	files, err := ui.logManager.ListLogFiles()
	if err != nil {
		return nil, err
	}

	// Convert to relative paths for UI display
	var relativePaths []string
	for _, file := range files {
		relativePaths = append(relativePaths, filepath.Base(file))
	}

	return relativePaths, nil
}

// GetLogFileContent returns the content of a specific log file (for future UI log viewer)
func (ui *UIManager) GetLogFileContent(filename string) (string, error) {
	// For security, only allow files in the logs directory
	fullPath := filepath.Join("logs", filepath.Base(filename))
	return ui.logManager.GetLogFileContent(fullPath)
}

// DisplayCurrentStatus shows the current status in the console
func (ui *UIManager) DisplayCurrentStatus() {
	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("│  📊 Current Status: %-35s │\n", ui.status.Status)
	fmt.Printf("│  📡 Device: %-43s │\n", ui.status.DeviceName)
	if ui.status.IsScanning {
		fmt.Println("│  🔍 State: SCANNING                                     │")
	} else {
		fmt.Println("│  🔍 State: READY                                        │")
	}
	if ui.status.LastCardUID != "" {
		fmt.Printf("│  🏷️  Last Card: %-37s │\n", ui.status.LastCardUID)
	}
	fmt.Printf("│  📝 Log File: %-39s │\n", filepath.Base(ui.status.LogFilePath))
	fmt.Println("└─────────────────────────────────────────────────────────┘")
	fmt.Println()
}

// DisplayDeviceSelectionHelp shows help for future UI-based device selection
func (ui *UIManager) DisplayDeviceSelectionHelp() {
	fmt.Println()
	fmt.Println("💡 Future Enhancement: UI-based Device Selection")
	fmt.Println("   In future versions, you will be able to:")
	fmt.Println("   • Select NFC devices from a web interface")
	fmt.Println("   • Monitor scanning status in real-time")
	fmt.Println("   • View and download log files")
	fmt.Println("   • Configure settings through the UI")
	fmt.Println()
}

// DisplayLogAccessInfo shows information about log file access
func (ui *UIManager) DisplayLogAccessInfo() {
	fmt.Println()
	fmt.Println("📋 Log File Information:")
	fmt.Printf("   Current log file: %s\n", ui.status.LogFilePath)
	fmt.Println("   📁 All logs are stored in the 'logs' directory")
	fmt.Println("   💡 Future versions will allow viewing logs through a web interface")

	// List available log files
	files, err := ui.GetLogFiles()
	if err == nil && len(files) > 0 {
		fmt.Println("   Available log files:")
		for _, file := range files {
			fmt.Printf("     • %s\n", file)
		}
	}
	fmt.Println()
}
