package main

import (
	"github.com/micmonay/keybd_event"
)

// CapsLockManager handles CAPS Lock state management during keyboard input (Linux stub)
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

// IsCapsLockOn checks if CAPS Lock is currently enabled (Linux implementation would need X11)
func (c *CapsLockManager) IsCapsLockOn() bool {
	// TODO: Implement using X11 or other Linux methods
	// For now, assume CAPS Lock is off
	return false
}

// DisableCapsLock disables CAPS Lock and saves the original state
func (c *CapsLockManager) DisableCapsLock() error {
	c.originalState = c.IsCapsLockOn()

	if c.originalState {
		// CAPS Lock is on, turn it off
		c.kb.SetKeys(58) // VK_CAPSLOCK for Linux
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
		c.kb.SetKeys(58) // VK_CAPSLOCK for Linux
		if err := c.kb.Launching(); err != nil {
			return err
		}
	}

	return nil
}
