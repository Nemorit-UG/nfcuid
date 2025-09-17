package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"context"
	"sync"
)

// HotkeyManager handles global hotkey registration and monitoring
type HotkeyManager struct {
	service    Service
	hotkey     string
	stopCtx    context.Context
	stopCancel context.CancelFunc
	mu         sync.Mutex
	running    bool
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
	go hm.monitorCommands()
	hm.running = true
	
	fmt.Printf("Repeat functionality ready - Type 'repeat' or 'r' + Enter to repeat last input (hotkey: %s)\n", hm.hotkey)
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

// monitorCommands monitors for text commands to trigger repeat
func (hm *HotkeyManager) monitorCommands() {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		select {
		case <-hm.stopCtx.Done():
			return
		default:
			// Non-blocking check for input
			if scanner.Scan() {
				input := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if input == "repeat" || input == "r" {
					hm.handleRepeatCommand()
				}
			}
			if scanner.Err() != nil {
				return // Exit on scanner error
			}
		}
	}
}

// handleRepeatCommand processes repeat command activation
func (hm *HotkeyManager) handleRepeatCommand() {
	fmt.Printf("Repeat command triggered - repeating last input\n")
	
	if err := hm.service.RepeatLastInput(); err != nil {
		fmt.Printf("Failed to repeat last input: %v\n", err)
	}
}

// Note: This simplified implementation uses text commands instead of global hotkeys
// for cross-platform compatibility and ease of testing. In a production environment,
// you would implement platform-specific global hotkey detection:
// - Windows: RegisterHotKey API or SetWindowsHookEx
// - Linux: X11 XGrabKey or similar
// - macOS: Carbon RegisterEventHotKey or Cocoa