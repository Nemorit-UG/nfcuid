package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HotkeyManager handles global hotkey registration and monitoring
type HotkeyManager struct {
	service    Service
	hotkey     string
	stopCtx    context.Context
	stopCancel context.CancelFunc
	mu         sync.Mutex
	running    bool
	// lastDown tracks the previous state of the hotkey to detect rising edges
	lastDown   bool
}

// NewHotkeyManager creates a new hotkey manager
func NewHotkeyManager(service Service, hotkeyConfig string) *HotkeyManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &HotkeyManager{
		service:    service,
		hotkey:     hotkeyConfig,
		stopCtx:    ctx,
		stopCancel: cancel,
	}
}

// Start begins monitoring for the configured hotkey
func (hm *HotkeyManager) Start() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.running {
		return fmt.Errorf("hotkey manager already running")
	}

	// Start hotkey monitoring in a goroutine
	go hm.monitorHotkey()
	hm.running = true

	fmt.Printf("Hotkey monitoring started - Press %s to repeat last input\n", hm.hotkey)
	return nil
}

// Stop terminates hotkey monitoring
func (hm *HotkeyManager) Stop() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if !hm.running {
		return
	}

	hm.stopCancel()
	hm.running = false
}

// monitorHotkey monitors for the configured global hotkey
func (hm *HotkeyManager) monitorHotkey() {
	// Polling-based detection with rising-edge trigger.
	// Platform-specific key state retrieval is implemented in per-OS files.
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-hm.stopCtx.Done():
			return
		case <-ticker.C:
			pressed := isHotkeyPressed(hm)
			// Trigger only on key down transition (rising edge)
			if pressed && !hm.lastDown {
				hm.handleHotkeyPress()
			}
			hm.lastDown = pressed
		}
	}
}

// isHotkeyPressed is implemented in platform-specific files (e.g., hotkey_windows.go, hotkey_default.go)

// handleHotkeyPress processes hotkey activation
func (hm *HotkeyManager) handleHotkeyPress() {
	fmt.Printf("Hotkey %s detected - repeating last input\n", hm.hotkey)

	if err := hm.service.RepeatLastInput(); err != nil {
		fmt.Printf("Failed to repeat last input: %v\n", err)
	}
}

// Note: This implementation provides the framework for global hotkey detection.
// For actual hotkey functionality, platform-specific implementations are required:
// - Windows: Use RegisterHotKey API or SetWindowsHookEx for global hotkey detection
// - Linux: Use X11 XGrabKey or similar low-level input monitoring
// - macOS: Use Carbon RegisterEventHotKey or Cocoa event monitoring
//
// The current implementation provides the structure but always returns false
// for hotkey detection, serving as a foundation for future platform-specific work.
