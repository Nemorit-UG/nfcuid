package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/micmonay/keybd_event"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

// LastContentManager manages the storage and retrieval of the last scanned content
type LastContentManager struct {
	lastContent   string
	lastTimestamp time.Time
	timeout       time.Duration
	mutex         sync.RWMutex
}

// NewLastContentManager creates a new LastContentManager with the given timeout
func NewLastContentManager(timeoutSeconds int) *LastContentManager {
	return &LastContentManager{
		timeout: time.Duration(timeoutSeconds) * time.Second,
	}
}

// Store saves the content with current timestamp
func (lcm *LastContentManager) Store(content string) {
	lcm.mutex.Lock()
	defer lcm.mutex.Unlock()
	
	lcm.lastContent = content
	lcm.lastTimestamp = time.Now()
}

// Retrieve gets the last stored content if it hasn't expired
func (lcm *LastContentManager) Retrieve() (string, bool) {
	lcm.mutex.RLock()
	defer lcm.mutex.RUnlock()
	
	// Check if we have content
	if lcm.lastContent == "" {
		return "", false
	}
	
	// Check if content has expired (0 timeout means no expiration)
	if lcm.timeout > 0 && time.Since(lcm.lastTimestamp) > lcm.timeout {
		return "", false
	}
	
	return lcm.lastContent, true
}

// HasContent returns true if there's valid content available
func (lcm *LastContentManager) HasContent() bool {
	_, hasContent := lcm.Retrieve()
	return hasContent
}

// Clear removes the stored content
func (lcm *LastContentManager) Clear() {
	lcm.mutex.Lock()
	defer lcm.mutex.Unlock()
	
	lcm.lastContent = ""
	lcm.lastTimestamp = time.Time{}
}

// GetAge returns how long ago the content was stored
func (lcm *LastContentManager) GetAge() time.Duration {
	lcm.mutex.RLock()
	defer lcm.mutex.RUnlock()
	
	if lcm.lastContent == "" {
		return 0
	}
	
	return time.Since(lcm.lastTimestamp)
}

// HotkeyMonitor manages global hotkey detection for repeat functionality
type HotkeyMonitor struct {
	config              *Config
	lastContentManager  *LastContentManager
	notificationManager *NotificationManager
	audioManager        *AudioManager
	keyboardBinding     keybd_event.KeyBonding
	keyMapping          *KeyMapping
	hotkeyDefinitions   []*HotkeyDefinition
	registeredHotkeys   []*HotkeyDefinition  // Successfully registered hotkeys
	failedHotkeys       []FailedHotkey       // Failed hotkey registrations
	stopChannel         chan bool
	running             bool
	hookActive          bool                 // Track if gohook is active
	mutex               sync.RWMutex
}

// FailedHotkey represents a hotkey that failed to register
type FailedHotkey struct {
	Definition *HotkeyDefinition
	Error      error
	Timestamp  time.Time
}

// NewHotkeyMonitor creates a new HotkeyMonitor
func NewHotkeyMonitor(config *Config, lcm *LastContentManager, nm *NotificationManager, am *AudioManager) *HotkeyMonitor {
	return &HotkeyMonitor{
		config:              config,
		lastContentManager:  lcm,
		notificationManager: nm,
		audioManager:        am,
		keyMapping:          NewKeyMapping(),
		stopChannel:         make(chan bool, 1),
	}
}

// Start begins monitoring for hotkey presses
func (hm *HotkeyMonitor) Start() error {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	
	if hm.running {
		return fmt.Errorf("hotkey monitor is already running")
	}
	
	if !hm.config.RepeatKey.Enabled {
		fmt.Println("Repeat key functionality is disabled")
		return nil
	}
	
	// Initialize keyboard binding for output
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return fmt.Errorf("failed to initialize keyboard binding for repeat key: %v", err)
	}
	hm.keyboardBinding = kb
	
	// Build hotkey definitions from configuration
	if err := hm.buildHotkeyDefinitions(); err != nil {
		return fmt.Errorf("failed to build hotkey definitions: %v", err)
	}
	
	if len(hm.hotkeyDefinitions) == 0 {
		fmt.Println("No valid hotkeys configured")
		return nil
	}
	
	// Attempt to register hotkeys with error handling
	hm.registeredHotkeys = make([]*HotkeyDefinition, 0, len(hm.hotkeyDefinitions))
	hm.failedHotkeys = make([]FailedHotkey, 0)
	
	for _, def := range hm.hotkeyDefinitions {
		if err := hm.validateHotkeyDefinition(def); err != nil {
			hm.failedHotkeys = append(hm.failedHotkeys, FailedHotkey{
				Definition: def,
				Error:      fmt.Errorf("validation failed: %v", err),
				Timestamp:  time.Now(),
			})
			fmt.Printf("  ⚠ Skipping invalid hotkey %s: %v\n", def.Name, err)
		} else {
			hm.registeredHotkeys = append(hm.registeredHotkeys, def)
		}
	}
	
	if len(hm.registeredHotkeys) == 0 {
		return fmt.Errorf("no valid hotkeys could be registered")
	}
	
	fmt.Printf("Starting repeat key monitor for %d hotkey(s)...\n", len(hm.registeredHotkeys))
	for _, def := range hm.registeredHotkeys {
		fmt.Printf("  ✓ %s\n", def.Name)
	}
	
	if len(hm.failedHotkeys) > 0 {
		fmt.Printf("Failed to register %d hotkey(s):\n", len(hm.failedHotkeys))
		for _, failed := range hm.failedHotkeys {
			fmt.Printf("  ✗ %s: %v\n", failed.Definition.Name, failed.Error)
		}
	}
	
	hm.running = true
	
	// Start monitoring in a separate goroutine with error recovery
	go hm.monitorLoopWithRecovery()
	
	return nil
}

// Stop stops the hotkey monitoring
func (hm *HotkeyMonitor) Stop() {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	
	if !hm.running {
		return
	}
	
	fmt.Println("Stopping repeat key monitor...")
	hm.running = false
	
	// Ensure event processing is ended to unblock monitor loop
	if hm.hookActive {
		robotgo.EventEnd()
		hm.hookActive = false
	}
	
	select {
	case hm.stopChannel <- true:
	default:
	}
}

// IsRunning returns whether the monitor is currently running
func (hm *HotkeyMonitor) IsRunning() bool {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	return hm.running
}

// monitorLoopWithRecovery runs the main monitoring loop with error recovery
func (hm *HotkeyMonitor) monitorLoopWithRecovery() {
	maxRetries := 3
	retryCount := 0
	
	for {
		// Check if we should stop
		select {
		case <-hm.stopChannel:
			fmt.Println("Repeat key monitor stopped")
			return
		default:
		}
		
		if !hm.running {
			return
		}
		
		err := hm.monitorLoop()
		
		// If monitoring stopped normally, exit
		if !hm.running {
			return
		}
		
		// If we hit an error, try to recover
		retryCount++
		if retryCount > maxRetries {
			fmt.Printf("Hotkey monitor failed after %d retries, giving up\n", maxRetries)
			if hm.config.Notifications.ShowErrors {
				hm.notificationManager.NotifyError("Hotkey monitor failed permanently")
			}
			return
		}
		
		fmt.Printf("Hotkey monitor error (retry %d/%d): %v\n", retryCount, maxRetries, err)
		if hm.config.Notifications.ShowErrors {
			hm.notificationManager.NotifyError(fmt.Sprintf("Hotkey monitor restarting (attempt %d)", retryCount))
		}
		
		// Wait before retry
		time.Sleep(time.Duration(retryCount) * time.Second)
	}
}

// monitorLoop sets up global hotkey hooks using robotgo.EventHook and processes events
func (hm *HotkeyMonitor) monitorLoop() error {
	hm.mutex.Lock()
	hm.hookActive = true
	hm.mutex.Unlock()

	fmt.Println("Repeat key monitor started with global hotkey hooks")

	// Register hooks for each configured hotkey
	for _, def := range hm.registeredHotkeys {
		keyName := hm.keyCodeToRobotgoName(def.KeyCode, def.OriginalKey)
		// Build the key sequence: modifiers + main key
		keys := append([]string{}, def.OriginalModifiers...)
		keys = append(keys, keyName)

		robotgo.EventHook(hook.KeyDown, keys, func(e hook.Event) {
			// Double-check running state
			if !hm.IsRunning() {
				return
			}
			fmt.Printf("Hotkey detected: %s\n", def.Name)
			if err := hm.TriggerRepeat(); err != nil {
				fmt.Printf("Failed to trigger repeat: %v\n", err)
				if hm.config.Notifications.ShowErrors {
					hm.notificationManager.NotifyError(fmt.Sprintf("Repeat failed: %v", err))
				}
			}
		})
	}

	// Start event processing
	evChan := robotgo.EventStart()
	defer robotgo.EventEnd()

	for {
		select {
		case <-hm.stopChannel:
			fmt.Println("Repeat key monitor stopped")
			hm.mutex.Lock()
			hm.hookActive = false
			hm.mutex.Unlock()
			return nil
		case _, ok := <-evChan:
			if !ok {
				return fmt.Errorf("event channel closed unexpectedly")
			}
			// We rely on the registered callbacks; nothing to do here
		}
	}
}

// TriggerRepeat manually triggers a repeat action (for testing and fallback)
func (hm *HotkeyMonitor) TriggerRepeat() error {
	if !hm.config.RepeatKey.Enabled {
		return fmt.Errorf("repeat key functionality is disabled")
	}
	
	// Check if we require a previous scan and don't have one
	if hm.config.RepeatKey.RequirePreviousScan && !hm.lastContentManager.HasContent() {
		if hm.config.RepeatKey.Notification {
			hm.notificationManager.NotifyInfo("Repeat Key", "Kein vorheriger Scan verfügbar zum Wiederholen")
		}
		return fmt.Errorf("no previous scan available to repeat")
	}
	
	// Get the last content
	content, hasContent := hm.lastContentManager.Retrieve()
	if !hasContent {
		if hm.config.RepeatKey.Notification {
			hm.notificationManager.NotifyInfo("Repeat Key", "Kein Inhalt zum Wiederholen verfügbar")
		}
		return fmt.Errorf("no content available to repeat")
	}
	
	fmt.Printf("Repeating last scanned content: %s\n", content)
	
	// Output the content using keyboard
	if err := KeyboardWrite(content, hm.keyboardBinding); err != nil {
		hm.notificationManager.NotifyError("Wiederholte Eingabe fehlgeschlagen. Cursor im richtigen Feld?")
		hm.audioManager.PlayErrorSound()
		return fmt.Errorf("failed to write repeated keyboard output: %v", err)
	}
	
	// Provide feedback
	if hm.config.RepeatKey.Notification {
		age := hm.lastContentManager.GetAge()
		message := fmt.Sprintf("Letzte Karte wiederholt (vor %v gescannt)", formatDuration(age))
		hm.notificationManager.NotifyInfo("Repeat Key", message)
	}
	
	hm.audioManager.PlaySuccessSound()
	fmt.Println("Repeat action completed successfully")
	
	return nil
}

// buildHotkeyDefinitions builds hotkey definitions from configuration
func (hm *HotkeyMonitor) buildHotkeyDefinitions() error {
	hm.hotkeyDefinitions = make([]*HotkeyDefinition, 0)
	
	// Use new hotkeys configuration if available
	if len(hm.config.RepeatKey.Hotkeys) > 0 {
		for i, hotkey := range hm.config.RepeatKey.Hotkeys {
			def, err := hm.keyMapping.BuildHotkeyDefinition(hotkey)
			if err != nil {
				return fmt.Errorf("hotkey %d: %v", i+1, err)
			}
			hm.hotkeyDefinitions = append(hm.hotkeyDefinitions, def)
		}
	} else if hm.config.RepeatKey.Key != "" {
		// Backward compatibility: convert old Key format
		hotkey := Hotkey{
			Key:       hm.config.RepeatKey.Key,
			Modifiers: []string{},
			Name:      fmt.Sprintf("%s Key", strings.Title(hm.config.RepeatKey.Key)),
		}
		def, err := hm.keyMapping.BuildHotkeyDefinition(hotkey)
		if err != nil {
			return fmt.Errorf("legacy key configuration: %v", err)
		}
		hm.hotkeyDefinitions = append(hm.hotkeyDefinitions, def)
	}
	
	return nil
}

// validateHotkeyDefinition validates a hotkey definition
func (hm *HotkeyMonitor) validateHotkeyDefinition(def *HotkeyDefinition) error {
	if def == nil {
		return fmt.Errorf("hotkey definition is nil")
	}
	
	if def.KeyCode == 0 {
		return fmt.Errorf("invalid key code")
	}
	
	if def.Name == "" {
		return fmt.Errorf("hotkey name cannot be empty")
	}
	
	if def.OriginalKey == "" {
		return fmt.Errorf("original key cannot be empty")
	}
	
	return nil
}


// keyCodeToRobotgoName converts internal key codes to robotgo key names

// keyCodeToRobotgoName converts internal key codes to robotgo key names
func (hm *HotkeyMonitor) keyCodeToRobotgoName(keyCode uint16, originalKey string) string {
	// Map common keys to robotgo names
	robotgoKeyMap := map[string]string{
		"ctrl":      "ctrl",
		"alt":       "alt",
		"shift":     "shift",
		"cmd":       "cmd",
		"win":       "cmd",
		"home":      "home",
		"end":       "end",
		"insert":    "insert",
		"delete":    "delete",
		"backspace": "backspace",
		"tab":       "tab",
		"enter":     "enter",
		"space":     "space",
		"escape":    "esc",
		"up":        "up",
		"down":      "down",
		"left":      "left",
		"right":     "right",
		"pageup":    "pageup",
		"pagedown":  "pagedown",
		"f1":        "f1",
		"f2":        "f2",
		"f3":        "f3",
		"f4":        "f4",
		"f5":        "f5",
		"f6":        "f6",
		"f7":        "f7",
		"f8":        "f8",
		"f9":        "f9",
		"f10":       "f10",
		"f11":       "f11",
		"f12":       "f12",
	}
	
	if robotgoKey, exists := robotgoKeyMap[strings.ToLower(originalKey)]; exists {
		return robotgoKey
	}
	
	// For letters and numbers, return as-is
	return strings.ToLower(originalKey)
}

// getModifiersFromMask extracts modifier key names from a modifier mask
func (hm *HotkeyMonitor) getModifiersFromMask(mask uint16) []string {
	var modifiers []string
	
	if mask&0x11 != 0 { // Ctrl
		modifiers = append(modifiers, "ctrl")
	}
	if mask&0x12 != 0 { // Alt
		modifiers = append(modifiers, "alt")
	}
	if mask&0x10 != 0 { // Shift
		modifiers = append(modifiers, "shift")
	}
	if mask&0x5B != 0 { // Cmd/Win
		modifiers = append(modifiers, "cmd")
	}
	
	return modifiers
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// GetRegisteredHotkeys returns the list of successfully registered hotkeys
func (hm *HotkeyMonitor) GetRegisteredHotkeys() []*HotkeyDefinition {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	
	result := make([]*HotkeyDefinition, len(hm.registeredHotkeys))
	copy(result, hm.registeredHotkeys)
	return result
}

// GetFailedHotkeys returns the list of hotkeys that failed to register
func (hm *HotkeyMonitor) GetFailedHotkeys() []FailedHotkey {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	
	result := make([]FailedHotkey, len(hm.failedHotkeys))
	copy(result, hm.failedHotkeys)
	return result
}

// GetStatus returns the current status of the hotkey monitor
func (hm *HotkeyMonitor) GetStatus() map[string]interface{} {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	
	return map[string]interface{}{
		"running":            hm.running,
		"hookActive":         hm.hookActive,
		"registeredCount":    len(hm.registeredHotkeys),
		"failedCount":        len(hm.failedHotkeys),
		"totalConfigured":    len(hm.hotkeyDefinitions),
	}
}