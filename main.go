package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("NFC UID Reader - Enhanced Version")
	fmt.Println("==================================")

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize notification manager
	notificationManager := NewNotificationManager(config)
	
	// Initialize audio manager
	audioManager := NewAudioManager(config)
	
	// Initialize restart manager
	restartManager := NewRestartManager(config, notificationManager)
	
	// Initialize last content manager for repeat key functionality
	lastContentManager := NewLastContentManager(config.RepeatKey.ContentTimeout)
	
	// Initialize hotkey monitor for repeat key functionality
	hotkeyMonitor := NewHotkeyMonitor(config, lastContentManager, notificationManager, audioManager)
	
	// Initialize browser manager
	var browserManager *BrowserManager
	if config.Web.OpenWebsite {
		browserManager = NewBrowserManager(config.Web.Fullscreen)
		
		// Open browser window on startup
		fmt.Printf("Opening browser: %s\n", config.Web.WebsiteURL)
		if err := browserManager.OpenURL(config.Web.WebsiteURL); err != nil {
			notificationManager.NotifyErrorThrottled("browser-error", fmt.Sprintf("Failed to open browser: %v", err))
			fmt.Printf("Warning: Failed to open browser: %v\n", err)
		}
	}

	// Convert config to legacy Flags struct for compatibility
	appFlags := config.ToFlags()

	// Initialize and start the NFC service
	service := NewService(appFlags, config, notificationManager, restartManager, audioManager, lastContentManager, hotkeyMonitor)
	
	fmt.Println("Starting NFC card reader service...")

	notificationManager.NotifyInfo("NFC Lesegerät", "Service gestartet - bereit zum Kartenlesen")
	
	service.Start()
}
