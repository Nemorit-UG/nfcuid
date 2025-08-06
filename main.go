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
	
	// Initialize browser manager
	var browserManager *BrowserManager
	if config.Web.OpenWebsite {
		browserManager = NewBrowserManager(config.Web.Fullscreen)
		
		// Open browser window on startup
		fmt.Printf("Opening browser: %s\n", config.Web.WebsiteURL)
		if err := browserManager.OpenURL(config.Web.WebsiteURL); err != nil {
			notificationManager.NotifyError(fmt.Sprintf("Failed to open browser: %v", err))
			fmt.Printf("Warning: Failed to open browser: %v\n", err)
		} else {
			notificationManager.NotifyInfo("NFC Reader", "Browser opened successfully")
		}
	}

	// Convert config to legacy Flags struct for compatibility
	appFlags := config.ToFlags()

	// Initialize and start the NFC service
	service := NewService(appFlags, config, notificationManager)
	
	fmt.Println("Starting NFC card reader service...")
	notificationManager.NotifyInfo("NFC Reader", "Service started - ready to read cards")
	
	service.Start()
}
