package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("NFC UID Reader - Enhanced Version")
	fmt.Printf("Version: %s\n", Version)
	fmt.Println("==================================")

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize notification manager
	notificationManager := NewNotificationManager(config)

	// Initialize update checker and check for updates if enabled
	if config.Updates.Enabled && config.Updates.CheckOnStartup {
		updateChecker := NewUpdateChecker(config, notificationManager)
		go func() {
			// Run update check in background to avoid blocking startup
			if err := updateChecker.PerformUpdateCheck(); err != nil {
				fmt.Printf("Update check failed: %v\n", err)
			}
		}()
	}

	// Initialize audio manager
	audioManager := NewAudioManager(config)

	// Initialize restart manager
	restartManager := NewRestartManager(config, notificationManager)

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
	service := NewService(appFlags, config, notificationManager, restartManager, audioManager)

	fmt.Println("Starting NFC card reader service...")
	notificationManager.NotifyInfo("NFC Lesegerät", "Service gestartet - bereit zum Kartenlesen")

	service.Start()
}
