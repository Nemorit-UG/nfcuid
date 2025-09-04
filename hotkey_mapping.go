package main

import (
	"fmt"
	"strings"
)

// KeyMapping provides cross-platform key code mapping
type KeyMapping struct {
	keyMap      map[string]uint16
	modifierMap map[string]uint16
}

// NewKeyMapping creates a new key mapping system
func NewKeyMapping() *KeyMapping {
	km := &KeyMapping{
		keyMap:      make(map[string]uint16),
		modifierMap: make(map[string]uint16),
	}
	
	km.initializeKeyMaps()
	return km
}

// initializeKeyMaps sets up the key code mappings
func (km *KeyMapping) initializeKeyMaps() {
	// Letters (using string-based key names for robotgo)
	km.keyMap["a"] = uint16('a')
	km.keyMap["b"] = uint16('b')
	km.keyMap["c"] = uint16('c')
	km.keyMap["d"] = uint16('d')
	km.keyMap["e"] = uint16('e')
	km.keyMap["f"] = uint16('f')
	km.keyMap["g"] = uint16('g')
	km.keyMap["h"] = uint16('h')
	km.keyMap["i"] = uint16('i')
	km.keyMap["j"] = uint16('j')
	km.keyMap["k"] = uint16('k')
	km.keyMap["l"] = uint16('l')
	km.keyMap["m"] = uint16('m')
	km.keyMap["n"] = uint16('n')
	km.keyMap["o"] = uint16('o')
	km.keyMap["p"] = uint16('p')
	km.keyMap["q"] = uint16('q')
	km.keyMap["r"] = uint16('r')
	km.keyMap["s"] = uint16('s')
	km.keyMap["t"] = uint16('t')
	km.keyMap["u"] = uint16('u')
	km.keyMap["v"] = uint16('v')
	km.keyMap["w"] = uint16('w')
	km.keyMap["x"] = uint16('x')
	km.keyMap["y"] = uint16('y')
	km.keyMap["z"] = uint16('z')
	
	// Numbers
	km.keyMap["0"] = uint16('0')
	km.keyMap["1"] = uint16('1')
	km.keyMap["2"] = uint16('2')
	km.keyMap["3"] = uint16('3')
	km.keyMap["4"] = uint16('4')
	km.keyMap["5"] = uint16('5')
	km.keyMap["6"] = uint16('6')
	km.keyMap["7"] = uint16('7')
	km.keyMap["8"] = uint16('8')
	km.keyMap["9"] = uint16('9')
	
	// Function keys
	km.keyMap["f1"] = 0x70   // F1
	km.keyMap["f2"] = 0x71   // F2
	km.keyMap["f3"] = 0x72   // F3
	km.keyMap["f4"] = 0x73   // F4
	km.keyMap["f5"] = 0x74   // F5
	km.keyMap["f6"] = 0x75   // F6
	km.keyMap["f7"] = 0x76   // F7
	km.keyMap["f8"] = 0x77   // F8
	km.keyMap["f9"] = 0x78   // F9
	km.keyMap["f10"] = 0x79  // F10
	km.keyMap["f11"] = 0x7A  // F11
	km.keyMap["f12"] = 0x7B  // F12
	
	// Special keys
	km.keyMap["home"] = 0x24     // Home
	km.keyMap["end"] = 0x23      // End
	km.keyMap["insert"] = 0x2D   // Insert
	km.keyMap["delete"] = 0x2E   // Delete
	km.keyMap["backspace"] = 0x08 // Backspace
	km.keyMap["tab"] = 0x09      // Tab
	km.keyMap["enter"] = 0x0D    // Enter
	km.keyMap["space"] = 0x20    // Space
	km.keyMap["escape"] = 0x1B   // Escape
	
	// Arrow keys
	km.keyMap["up"] = 0x26       // Up
	km.keyMap["down"] = 0x28     // Down
	km.keyMap["left"] = 0x25     // Left
	km.keyMap["right"] = 0x27    // Right
	
	// Page navigation
	km.keyMap["pageup"] = 0x21   // Page Up
	km.keyMap["pagedown"] = 0x22 // Page Down
	
	// Numpad keys
	km.keyMap["numpad0"] = 0x60  // Numpad 0
	km.keyMap["numpad1"] = 0x61  // Numpad 1
	km.keyMap["numpad2"] = 0x62  // Numpad 2
	km.keyMap["numpad3"] = 0x63  // Numpad 3
	km.keyMap["numpad4"] = 0x64  // Numpad 4
	km.keyMap["numpad5"] = 0x65  // Numpad 5
	km.keyMap["numpad6"] = 0x66  // Numpad 6
	km.keyMap["numpad7"] = 0x67  // Numpad 7
	km.keyMap["numpad8"] = 0x68  // Numpad 8
	km.keyMap["numpad9"] = 0x69  // Numpad 9
	km.keyMap["numpadmultiply"] = 0x6A // Numpad *
	km.keyMap["numpadadd"] = 0x6B      // Numpad +
	km.keyMap["numpadsubtract"] = 0x6D // Numpad -
	km.keyMap["numpaddivide"] = 0x6F   // Numpad /
	km.keyMap["numpaddecimal"] = 0x6E  // Numpad .
	km.keyMap["numpadenter"] = 0x0D    // Numpad Enter
	
	// Modifier keys (can be used as primary keys)
	km.keyMap["ctrl"] = 0x11     // Ctrl
	km.keyMap["alt"] = 0x12      // Alt
	km.keyMap["shift"] = 0x10    // Shift
	km.keyMap["cmd"] = 0x5B      // Windows/Cmd key
	km.keyMap["win"] = 0x5B      // Windows key
	
	// Modifier map for combinations (same values for simplicity)
	km.modifierMap["ctrl"] = 0x11
	km.modifierMap["alt"] = 0x12
	km.modifierMap["shift"] = 0x10
	km.modifierMap["cmd"] = 0x5B
	km.modifierMap["win"] = 0x5B
}

// GetKeyCode returns the key code for a given key name
func (km *KeyMapping) GetKeyCode(keyName string) (uint16, bool) {
	code, exists := km.keyMap[strings.ToLower(keyName)]
	return code, exists
}

// GetModifierCode returns the modifier code for a given modifier name
func (km *KeyMapping) GetModifierCode(modifierName string) (uint16, bool) {
	code, exists := km.modifierMap[strings.ToLower(modifierName)]
	return code, exists
}

// BuildModifierMask creates a modifier mask from a list of modifier names
func (km *KeyMapping) BuildModifierMask(modifiers []string) uint16 {
	var mask uint16
	for _, modifier := range modifiers {
		if code, exists := km.GetModifierCode(modifier); exists {
			mask |= code
		}
	}
	return mask
}

// HotkeyDefinition represents a complete hotkey definition with codes
type HotkeyDefinition struct {
	Name              string   // User-friendly name
	KeyCode           uint16   // Primary key code
	ModifierMask      uint16   // Combined modifier mask
	OriginalKey       string   // Original key string from config
	OriginalModifiers []string // Original modifier strings (normalized, for robotgo)
}

// BuildHotkeyDefinition creates a HotkeyDefinition from a Hotkey config
func (km *KeyMapping) BuildHotkeyDefinition(hotkey Hotkey) (*HotkeyDefinition, error) {
	keyCode, exists := km.GetKeyCode(hotkey.Key)
	if !exists {
		return nil, fmt.Errorf("unsupported key: %s", hotkey.Key)
	}
	
	modifierMask := km.BuildModifierMask(hotkey.Modifiers)
	
	name := hotkey.Name
	if name == "" {
		if len(hotkey.Modifiers) > 0 {
			name = strings.Join(hotkey.Modifiers, "+") + "+" + hotkey.Key
		} else {
			name = hotkey.Key
		}
	}
	
	// Normalize modifiers to lowercase for robotgo
	normalizedMods := make([]string, len(hotkey.Modifiers))
	for i, m := range hotkey.Modifiers {
		normalizedMods[i] = strings.ToLower(m)
	}
	
	return &HotkeyDefinition{
		Name:              name,
		KeyCode:           keyCode,
		ModifierMask:      modifierMask,
		OriginalKey:       hotkey.Key,
		OriginalModifiers: normalizedMods,
	}, nil
}