//go:build windows
// +build windows

package main

import (
	"strconv"
	"strings"
	"syscall"
)

var (
	user32Hotkey        = syscall.NewLazyDLL("user32.dll")
	procGetAsyncKeyState = user32Hotkey.NewProc("GetAsyncKeyState")
)

// normalizeHotkey maps synonyms and normalizes case
func normalizeHotkey(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	switch s {
	case "POS1":
		return "HOME"
	default:
		return s
	}
}

// vkFromHotkey returns a Windows virtual-key code for the configured hotkey
func vkFromHotkey(h string) (uintptr, bool) {
	switch h {
	case "HOME":
		return uintptr(0x24), true
	case "END":
		return uintptr(0x23), true
	case "INSERT":
		return uintptr(0x2D), true
	case "DELETE":
		return uintptr(0x2E), true
	}
	if strings.HasPrefix(h, "F") {
		nStr := strings.TrimPrefix(h, "F")
		if n, err := strconv.Atoi(nStr); err == nil {
			if n >= 1 && n <= 24 {
				return uintptr(0x70 + (n - 1)), true // VK_F1 = 0x70
			}
		}
	}
	return 0, false
}

// isHotkeyPressed checks the real-time state of the configured key using GetAsyncKeyState
func isHotkeyPressed(hm *HotkeyManager) bool {
	hk := normalizeHotkey(hm.hotkey)
	vk, ok := vkFromHotkey(hk)
	if !ok {
		return false
	}
	// High-order bit set means key is currently down
	r, _, _ := procGetAsyncKeyState.Call(vk)
	return int16(r) < 0
}