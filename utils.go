package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	mp3 "github.com/hajimehoshi/go-mp3"
	"github.com/skratchdot/open-golang/open"
)

// External reference to global single instance for cleanup
// This is set in main.go and used for cleanup in SafeExit
var globalSingleInstance *SingleInstance

// NotificationManager handles system notifications with throttling
type NotificationManager struct {
	enabled           bool
	showSuccess       bool
	showErrors        bool
	lastNotifications map[string]time.Time // Track last notification time per error type
	errorCounts       map[string]int       // Track consecutive error counts per type
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(config *Config) *NotificationManager {
	return &NotificationManager{
		enabled:           config.Notifications.Enabled,
		showSuccess:       config.Notifications.ShowSuccess,
		showErrors:        config.Notifications.ShowErrors,
		lastNotifications: make(map[string]time.Time),
		errorCounts:       make(map[string]int),
	}
}

// NotifySuccess sends a success notification (only when transitioning from error state)
func (nm *NotificationManager) NotifySuccess(message string) {
	if !nm.enabled || !nm.showSuccess {
		return
	}

	// Only notify success if we had previous errors (recovering from error state)
	if nm.hasRecentErrors() {
		err := beeep.Notify("NFC Karten-Lesung erfolgreich", message, "")
		if err != nil {
			log.Printf("Failed to send success notification: %v", err)
		}

		// Clear error counts on successful operation
		nm.clearErrorCounts()
	}
}

// NotifyError sends an error notification with smart throttling
func (nm *NotificationManager) NotifyError(message string) {
	if !nm.enabled || !nm.showErrors {
		return
	}

	errorType := nm.categorizeError(message)

	if nm.shouldNotifyError(errorType, message) {
		title := "NFC Reader-Fehler"
		if count := nm.errorCounts[errorType]; count > 1 {
			title = fmt.Sprintf("NFC Reader-Fehler (x%d)", count)
		}

		err := beeep.Alert(title, message, "")
		if err != nil {
			log.Printf("Failed to send error notification: %v", err)
		}

		nm.lastNotifications[errorType] = time.Now()
	}

	nm.errorCounts[errorType]++
}

// NotifyErrorThrottled sends throttled error notifications for system failures
func (nm *NotificationManager) NotifyErrorThrottled(errorType, message string) {
	if !nm.enabled || !nm.showErrors {
		return
	}

	if nm.shouldNotifyError(errorType, message) {
		title := "NFC System-Fehler"
		if count := nm.errorCounts[errorType]; count > 1 {
			title = fmt.Sprintf("NFC System-Fehler (x%d)", count)
		}

		err := beeep.Alert(title, message, "")
		if err != nil {
			log.Printf("Failed to send error notification: %v", err)
		}

		nm.lastNotifications[errorType] = time.Now()
	}

	nm.errorCounts[errorType]++
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
	
	// Clean up single instance lock if it exists
	if globalSingleInstance != nil {
		globalSingleInstance.Release()
	}
	
	os.Exit(code)
}

// RestartManager handles application self-restart functionality
type RestartManager struct {
	config              *Config
	notificationManager *NotificationManager
	contextFailureCount int
}

// NewRestartManager creates a new restart manager
func NewRestartManager(config *Config, notificationManager *NotificationManager) *RestartManager {
	return &RestartManager{
		config:              config,
		notificationManager: notificationManager,
		contextFailureCount: 0,
	}
}

// TrackContextFailure increments the context failure counter and triggers restart if threshold reached
func (rm *RestartManager) TrackContextFailure(err error) bool {
	return rm.trackSystemFailure("PC/SC Context", err)
}

// TrackSystemFailure increments the system failure counter and triggers restart if threshold reached
// Use this for any PC/SC system-level errors (readers, connections, system operations)
func (rm *RestartManager) TrackSystemFailure(operation string, err error) bool {
	return rm.trackSystemFailure(operation, err)
}

// trackSystemFailure is the internal implementation for tracking any PC/SC system failure
func (rm *RestartManager) trackSystemFailure(operation string, err error) bool {
	rm.contextFailureCount++

	fmt.Printf("PC/SC %s failure %d/%d: %v\n", operation, rm.contextFailureCount, rm.config.Advanced.MaxContextFailures, err)

	if rm.config.Advanced.SelfRestart && rm.contextFailureCount >= rm.config.Advanced.MaxContextFailures {
		rm.performSelfRestart(operation)
		return true // This will never actually return due to restart, but for clarity
	}

	return false
}

// ResetFailureCount resets the context failure counter (called on successful context establishment)
func (rm *RestartManager) ResetFailureCount() {
	if rm.contextFailureCount > 0 {
		fmt.Printf("PC/SC Context established successfully, resetting failure count\n")
		rm.contextFailureCount = 0
	}
}

// performSelfRestart performs the actual application restart
func (rm *RestartManager) performSelfRestart(operation string) {
	message := fmt.Sprintf("Maximale PC/SC %s Fehler erreicht (%d). Anwendung wird neu gestartet...", operation, rm.config.Advanced.MaxContextFailures)
	fmt.Println(message)

	if rm.notificationManager != nil {
		rm.notificationManager.NotifyInfo("NFC Lesegerät", message)
	}

	// Give time for notifications to be displayed
	time.Sleep(2 * time.Second)

	if rm.config.Advanced.RestartDelay > 0 {
		fmt.Printf("Waiting %d seconds before restart...\n", rm.config.Advanced.RestartDelay)
		time.Sleep(time.Duration(rm.config.Advanced.RestartDelay) * time.Second)
	}

	// Get the current executable path and arguments
	executable, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path for restart: %v\n", err)
		SafeExit(1, "Cannot restart application", rm.notificationManager)
		return
	}

	// Get original arguments (excluding the program name) and add restart flag
	args := os.Args[1:]
	args = append(args, "--auto-restart")

	fmt.Printf("Restarting application: %s %v\n", executable, args)

	// Start new process
	cmd := exec.Command(executable, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Start()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to restart application: %v", err)
		fmt.Println(errorMsg)
		if rm.notificationManager != nil {
			rm.notificationManager.NotifyError(errorMsg)
		}
		SafeExit(1, "Restart failed", rm.notificationManager)
		return
	}

	// Notify about successful restart initiation
	if rm.notificationManager != nil {
		rm.notificationManager.NotifyInfo("NFC Lesegerät", "Anwendungsneustart erfolgreich eingeleitet")
	}

	fmt.Println("New process started successfully. Exiting current instance.")
	os.Exit(0)
}

// categorizeError categorizes error messages into types for throttling
func (nm *NotificationManager) categorizeError(message string) string {
	switch {
	case strings.Contains(message, "PC/SC Context"):
		return "pc-sc-context"
	case strings.Contains(message, "Reader"):
		return "reader-error"
	case strings.Contains(message, "Card"):
		return "card-error"
	case strings.Contains(message, "Keyboard"):
		return "keyboard-error"
	case strings.Contains(message, "Browser"):
		return "browser-error"
	case strings.Contains(message, "Service"):
		return "service-error"
	default:
		return "general-error"
	}
}

// shouldNotifyError determines if an error notification should be sent based on throttling rules
func (nm *NotificationManager) shouldNotifyError(errorType, message string) bool {
	now := time.Now()

	// Always notify first occurrence of any error type
	lastNotification, exists := nm.lastNotifications[errorType]
	if !exists {
		return true
	}

	// Get error count for this type
	count := nm.errorCounts[errorType]

	// Throttling rules based on error count and type
	switch errorType {
	case "pc-sc-context", "reader-error":
		// Critical system errors: notify on 1st, 3rd, 5th, then every 10th
		if count == 0 || count == 2 || count == 4 {
			return true
		}
		if count >= 10 && count%10 == 0 {
			return true
		}
		// Also notify if it's been more than 5 minutes since last notification
		if now.Sub(lastNotification) > 5*time.Minute {
			return true
		}
	case "card-error":
		// Card errors: notify on 1st, then every 5th, or after 2 minutes
		if count == 0 || count%5 == 0 {
			return true
		}
		if now.Sub(lastNotification) > 2*time.Minute {
			return true
		}
	case "service-error":
		// Service errors: notify on 1st, 2nd, then every 5th
		if count <= 1 || count%5 == 0 {
			return true
		}
		if now.Sub(lastNotification) > 3*time.Minute {
			return true
		}
	default:
		// Other errors: notify on 1st, then every 3rd, or after 1 minute
		if count == 0 || count%3 == 0 {
			return true
		}
		if now.Sub(lastNotification) > 1*time.Minute {
			return true
		}
	}

	return false
}

// hasRecentErrors checks if there were any recent errors (for success notification logic)
func (nm *NotificationManager) hasRecentErrors() bool {
	for _, count := range nm.errorCounts {
		if count > 0 {
			return true
		}
	}
	return false
}

// clearErrorCounts resets all error counters (called on successful operation)
func (nm *NotificationManager) clearErrorCounts() {
	nm.errorCounts = make(map[string]int)
}

// AudioManager handles audio feedback for successful scans and errors
type AudioManager struct {
	enabled      bool
	successSound string
	errorSound   string
	volume       int
}

// NewAudioManager creates a new audio manager
func NewAudioManager(config *Config) *AudioManager {
	return &AudioManager{
		enabled:      config.Audio.Enabled,
		successSound: config.Audio.SuccessSound,
		errorSound:   config.Audio.ErrorSound,
		volume:       config.Audio.Volume,
	}
}

// PlaySuccessSound plays the configured success sound
func (am *AudioManager) PlaySuccessSound() {
	if !am.enabled {
		return
	}

	go am.playSound(am.successSound)
}

// PlayErrorSound plays the configured error sound
func (am *AudioManager) PlayErrorSound() {
	if !am.enabled {
		return
	}

	go am.playSound(am.errorSound)
}

// playSound plays the specified sound
func (am *AudioManager) playSound(soundType string) {
	switch soundType {
	case "beep":
		am.playSystemBeep()
	case "error":
		am.playSystemError()
	case "none", "":
		// No sound
		return
	default:
		// Assume it's a file path
		am.playAudioFile(soundType)
	}
}

// playSystemBeep plays a system beep sound
func (am *AudioManager) playSystemBeep() {
	switch runtime.GOOS {
	case "windows":
		// Windows system beep
		exec.Command("rundll32", "user32.dll,MessageBeep", "64").Run()
	case "darwin":
		// macOS system beep
		exec.Command("afplay", "/System/Library/Sounds/Ping.aiff").Run()
	case "linux":
		// Linux system beep - try multiple methods
		if exec.Command("pactl", "list", "short", "modules").Run() == nil {
			// PulseAudio available
			exec.Command("pactl", "upload-sample", "/usr/share/sounds/freedesktop/stereo/complete.oga", "beep").Run()
			exec.Command("pactl", "play-sample", "beep").Run()
		} else if exec.Command("which", "beep").Run() == nil {
			// beep command available
			exec.Command("beep", "-f", "800", "-l", "200").Run()
		} else {
			// Fallback to terminal bell
			fmt.Print("\a")
		}
	default:
		// Fallback to terminal bell
		fmt.Print("\a")
	}
}

// playSystemError plays a system error sound
func (am *AudioManager) playSystemError() {
	switch runtime.GOOS {
	case "windows":
		// Windows error sound
		exec.Command("rundll32", "user32.dll,MessageBeep", "16").Run()
	case "darwin":
		// macOS error sound
		exec.Command("afplay", "/System/Library/Sounds/Sosumi.aiff").Run()
	case "linux":
		// Linux error sound - lower pitch beeps
		if exec.Command("which", "beep").Run() == nil {
			exec.Command("beep", "-f", "300", "-l", "500").Run()
		} else {
			// Multiple terminal bells for error
			fmt.Print("\a\a")
		}
	default:
		// Fallback to double terminal bell
		fmt.Print("\a\a")
	}
}

// playAudioFile plays an audio file from the specified path
func (am *AudioManager) playAudioFile(filePath string) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Audio file not found: %s", filePath)
		return
	}

	// Check if it's an MP3 file - use go-mp3 for validation
	if strings.HasSuffix(strings.ToLower(filePath), ".mp3") {
		am.playMP3File(filePath)
		return
	}

	// For other file types, use system players directly
	am.playWithSystemPlayer(filePath)
}

// playMP3File plays an MP3 file using go-mp3 library
func (am *AudioManager) playMP3File(filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open MP3 file %s: %v", filePath, err)
		return
	}
	defer f.Close()

	// Verify it's a valid MP3 file
	_, err = mp3.NewDecoder(f)
	if err != nil {
		log.Printf("Failed to create MP3 decoder for %s: %v", filePath, err)
		// Fall back to system player
		am.playWithSystemPlayer(filePath)
		return
	}

	// Use system player for actual playback since oto has dependency issues
	am.playWithSystemPlayer(filePath)
}

// playWithSystemPlayer plays audio files using system players
func (am *AudioManager) playWithSystemPlayer(filePath string) {
	switch runtime.GOOS {
	case "windows":
		// On Windows, use PowerShell with MediaPlayer
		cmd := exec.Command("powershell", "-c", fmt.Sprintf(`
			Add-Type -AssemblyName presentationCore
			$mediaPlayer = New-Object system.windows.media.mediaplayer
			$mediaPlayer.open('%s')
			$mediaPlayer.Play()
			Start-Sleep -Seconds 2
		`, filePath))
		cmd.Run()
	case "darwin":
		// macOS - use afplay
		exec.Command("afplay", filePath).Run()
	case "linux":
		// Linux - try multiple audio players
		players := []string{"mpg123", "ffplay", "paplay", "aplay"}
		for _, player := range players {
			if exec.Command("which", player).Run() == nil {
				if player == "ffplay" {
					exec.Command(player, "-nodisp", "-autoexit", filePath).Run()
				} else {
					exec.Command(player, filePath).Run()
				}
				return
			}
		}
		log.Printf("No audio player found for: %s", filePath)
	default:
		log.Printf("Audio file playback not supported on this platform: %s", filePath)
	}
}
