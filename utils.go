package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/skratchdot/open-golang/open"
)

// NotificationManager handles system notifications
type NotificationManager struct {
	enabled     bool
	showSuccess bool
	showErrors  bool
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(config *Config) *NotificationManager {
	return &NotificationManager{
		enabled:     config.Notifications.Enabled,
		showSuccess: config.Notifications.ShowSuccess,
		showErrors:  config.Notifications.ShowErrors,
	}
}

// NotifySuccess sends a success notification
func (nm *NotificationManager) NotifySuccess(message string) {
	if !nm.enabled || !nm.showSuccess {
		return
	}
	
	err := beeep.Notify("NFC Card Read Success", message, "")
	if err != nil {
		log.Printf("Failed to send success notification: %v", err)
	}
}

// NotifyError sends an error notification
func (nm *NotificationManager) NotifyError(message string) {
	if !nm.enabled || !nm.showErrors {
		return
	}
	
	err := beeep.Alert("NFC Reader Error", message, "")
	if err != nil {
		log.Printf("Failed to send error notification: %v", err)
	}
}

// NotifyInfo sends an informational notification
func (nm *NotificationManager) NotifyInfo(title, message string) {
	if !nm.enabled {
		return
	}
	
	err := beeep.Notify(title, message, "")
	if err != nil {
		log.Printf("Failed to send info notification: %v", err)
	}
}

// BrowserManager handles browser operations
type BrowserManager struct {
	fullscreen bool
}

// NewBrowserManager creates a new browser manager
func NewBrowserManager(fullscreen bool) *BrowserManager {
	return &BrowserManager{
		fullscreen: fullscreen,
	}
}

// OpenURL opens the specified URL in the default browser
func (bm *BrowserManager) OpenURL(url string) error {
	fmt.Printf("Opening browser at: %s\n", url)
	
	if bm.fullscreen {
		return bm.openFullscreen(url)
	}
	
	return bm.openMaximized(url)
}

// openMaximized opens URL in maximized browser window
func (bm *BrowserManager) openMaximized(url string) error {
	switch runtime.GOOS {
	case "windows":
		// On Windows, try to open maximized
		cmd := exec.Command("cmd", "/c", "start", "/max", url)
		return cmd.Start()
	case "darwin":
		// On macOS, open normally first, then try to maximize
		if err := open.Run(url); err != nil {
			return err
		}
		// Give browser time to load
		time.Sleep(500 * time.Millisecond)
		// Try to maximize using AppleScript
		script := `tell application "System Events" to tell process "Safari" to set frontmost to true`
		exec.Command("osascript", "-e", script).Run()
		return nil
	case "linux":
		// On Linux, open normally - window managers handle maximization differently
		return open.Run(url)
	default:
		return open.Run(url)
	}
}

// openFullscreen opens URL in fullscreen browser window
func (bm *BrowserManager) openFullscreen(url string) error {
	switch runtime.GOOS {
	case "windows":
		// Windows: Use Chrome with kiosk mode if available, fallback to maximized
		chromeCmd := exec.Command("chrome", "--kiosk", url)
		if err := chromeCmd.Start(); err != nil {
			// Fallback to Edge kiosk mode
			edgeCmd := exec.Command("msedge", "--kiosk", url)
			if err := edgeCmd.Start(); err != nil {
				// Fallback to maximized window
				return bm.openMaximized(url)
			}
		}
		return nil
	case "darwin":
		// macOS: Try Chrome kiosk mode, fallback to Safari fullscreen
		chromeCmd := exec.Command("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "--kiosk", url)
		if err := chromeCmd.Start(); err != nil {
			// Fallback to Safari with fullscreen AppleScript
			if err := open.Run(url); err != nil {
				return err
			}
			time.Sleep(1 * time.Second)
			script := `
				tell application "Safari"
					activate
					tell application "System Events"
						key code 36 using {command down, control down}
					end tell
				end tell`
			exec.Command("osascript", "-e", script).Run()
		}
		return nil
	case "linux":
		// Linux: Try various browsers with fullscreen flags
		browsers := [][]string{
			{"google-chrome", "--kiosk", url},
			{"chromium-browser", "--kiosk", url},
			{"firefox", "--kiosk", url},
		}
		
		for _, browserCmd := range browsers {
			cmd := exec.Command(browserCmd[0], browserCmd[1:]...)
			if err := cmd.Start(); err == nil {
				return nil
			}
		}
		
		// Fallback to default browser
		if err := open.Run(url); err != nil {
			return err
		}
		
		// Try to make it fullscreen using xdotool if available
		time.Sleep(1 * time.Second)
		exec.Command("xdotool", "key", "F11").Run()
		return nil
	default:
		return open.Run(url)
	}
}

// RetryManager handles retry logic with exponential backoff
type RetryManager struct {
	maxAttempts int
	baseDelay   time.Duration
}

// NewRetryManager creates a new retry manager
func NewRetryManager(maxAttempts int, baseDelaySeconds int) *RetryManager {
	return &RetryManager{
		maxAttempts: maxAttempts,
		baseDelay:   time.Duration(baseDelaySeconds) * time.Second,
	}
}

// Retry executes the given function with retry logic
func (rm *RetryManager) Retry(operation func() error) error {
	var lastErr error
	
	for attempt := 1; attempt <= rm.maxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		if attempt < rm.maxAttempts {
			delay := time.Duration(attempt) * rm.baseDelay
			fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", attempt, err, delay)
			time.Sleep(delay)
		}
	}
	
	return fmt.Errorf("operation failed after %d attempts, last error: %v", rm.maxAttempts, lastErr)
}

// SafeExit performs a graceful shutdown
func SafeExit(code int, message string, notificationManager *NotificationManager) {
	if message != "" {
		fmt.Println(message)
		if notificationManager != nil {
			notificationManager.NotifyError(message)
		}
	}
	os.Exit(code)
}