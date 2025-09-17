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

	// Display prominent warning message
	displayWarningMessage()

	// Check for existing instances
	singleInstance := NewSingleInstance("nfcuid")
	globalSingleInstance = singleInstance // Store globally for cleanup

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
		SafeExit(1, fmt.Sprintf("Failed to load configuration: %v", err), nil)
	}

	// Initialize logging system
	logManager := NewLogManager()
	defer logManager.Close()
	logManager.LogInfo("Application starting", "version", "Enhanced Version")

	// Initialize UI manager for status tracking and future UI integration
	uiManager := NewUIManager(logManager)
	uiManager.UpdateStatus("Initializing")

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
	service := NewService(appFlags, config, notificationManager, restartManager, audioManager, logManager, uiManager)

	// Display helpful information about logs and future UI features
	uiManager.DisplayLogAccessInfo()
	uiManager.DisplayDeviceSelectionHelp()

	logManager.LogInfo("Starting NFC card reader service")
	fmt.Println("Starting NFC card reader service...")
	uiManager.UpdateStatus("Starting service")
	notificationManager.NotifyInfo("NFC Lesegerät", "Service gestartet - bereit zum Kartenlesen")

	service.Start()
}

// displayWarningMessage shows a prominent warning about not closing the application
func displayWarningMessage() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                  ⚠️  WARNUNG  ⚠️                            ║")
	fmt.Println("║                                                                              ║")
	fmt.Println("║  DIESE ANWENDUNG NICHT SCHLIESSEN!                                           ║")
	fmt.Println("║  DO NOT CLOSE THIS APPLICATION!                                              ║")
	fmt.Println("║                                                                              ║")
	fmt.Println("║  Das NFC-Lesegerät läuft kontinuierlich im Hintergrund.                      ║")
	fmt.Println("║  Das Schließen dieser Anwendung beendet den Karten-Lesedienst.               ║")
	fmt.Println("║                                                                              ║")
	fmt.Println("║  The NFC reader runs continuously in the background.                         ║")
	fmt.Println("║  Closing this application will stop the card reading service.                ║")
	fmt.Println("║                                                                              ║")
	fmt.Println("║  Logs werden in './logs/' gespeichert | Logs are saved in './logs/'          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
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
