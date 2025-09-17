package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("NFC UID Reader - Enhanced Version")
	fmt.Println("==================================")

	// Check for existing instances
	singleInstance := NewSingleInstance("nfcuid")
	globalSingleInstance = singleInstance  // Store globally for cleanup
	
	if !singleInstance.TryLock() {
		// Check if another instance is actually running
		isRunning, pid, err := singleInstance.GetRunningInstanceInfo()
		if err != nil {
			fmt.Printf("Error checking for existing instances: %v\n", err)
			os.Exit(1)
		}
		
		if isRunning {
			fmt.Printf("Another instance of NFC UID Reader is already running (PID: %d)\n", pid)
			fmt.Println("Please close the existing instance before starting a new one.")
			fmt.Println("This prevents conflicts with keyboard input from multiple instances.")
			os.Exit(1)
		} else {
			// Stale lock file was cleaned up, try again
			if !singleInstance.TryLock() {
				fmt.Println("Failed to acquire application lock after cleanup. Please try again.")
				os.Exit(1)
			}
		}
	}

	// Setup cleanup on exit
	setupGracefulShutdown(singleInstance)

	fmt.Println("✓ Single instance lock acquired successfully")

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		singleInstance.Release()
		os.Exit(1)
	}

	// Initialize notification manager
	notificationManager := NewNotificationManager(config)
	
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

// setupGracefulShutdown sets up signal handlers for graceful shutdown
func setupGracefulShutdown(singleInstance *SingleInstance) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		fmt.Println("\nReceived shutdown signal, cleaning up...")
		singleInstance.Release()
		os.Exit(0)
	}()
}
