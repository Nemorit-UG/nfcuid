package main

import (
	"syscall"

	"github.com/micmonay/keybd_event"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	getKeyState      = user32.NewProc("GetKeyState")
)

// CapsLockManager handles CAPS Lock state management during keyboard input
type CapsLockManager struct {
	originalState bool
	kb            keybd_event.KeyBonding
}

// NewCapsLockManager creates a new CAPS Lock manager
func NewCapsLockManager(kb keybd_event.KeyBonding) *CapsLockManager {
	return &CapsLockManager{
		kb: kb,
	}
}

// IsCapsLockOn checks if CAPS Lock is currently enabled
func (c *CapsLockManager) IsCapsLockOn() bool {
	ret, _, _ := getKeyState.Call(uintptr(keybd_event.VK_CAPSLOCK))
	// GetKeyState returns a value where the low-order bit indicates if the key is toggled
	return (ret & 0x0001) != 0
}

// DisableCapsLock disables CAPS Lock and saves the original state
func (c *CapsLockManager) DisableCapsLock() error {
	c.originalState = c.IsCapsLockOn()
	
	if c.originalState {
		// CAPS Lock is on, turn it off
		c.kb.SetKeys(keybd_event.VK_CAPSLOCK)
		if err := c.kb.Launching(); err != nil {
			return err
		}
	}
	
	return nil
}

// RestoreCapsLock restores the original CAPS Lock state
func (c *CapsLockManager) RestoreCapsLock() error {
	currentState := c.IsCapsLockOn()
	
	// Only toggle if the current state differs from the original state
	if currentState != c.originalState {
		c.kb.SetKeys(keybd_event.VK_CAPSLOCK)
		if err := c.kb.Launching(); err != nil {
			return err
		}
	}
	
	return nil
}