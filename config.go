package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	NFC struct {
		Device         int    `yaml:"device"`
		CapsLock       bool   `yaml:"caps_lock"`
		Reverse        bool   `yaml:"reverse"`
		Decimal        bool   `yaml:"decimal"`
		DecimalPadding int    `yaml:"decimal_padding"`
		EndChar        string `yaml:"end_char"`
		InChar         string `yaml:"in_char"`
	} `yaml:"nfc"`
	Web struct {
		OpenWebsite bool   `yaml:"open_website"`
		WebsiteURL  string `yaml:"website_url"`
		Fullscreen  bool   `yaml:"fullscreen"`
	} `yaml:"web"`
	Notifications struct {
		Enabled     bool `yaml:"enabled"`
		ShowSuccess bool `yaml:"show_success"`
		ShowErrors  bool `yaml:"show_errors"`
	} `yaml:"notifications"`
	Audio struct {
		Enabled        bool   `yaml:"enabled"`
		SuccessSound   string `yaml:"success_sound"`
		ErrorSound     string `yaml:"error_sound"`
		Volume         int    `yaml:"volume"`
	} `yaml:"audio"`
	Advanced struct {
		RetryAttempts      int  `yaml:"retry_attempts"`
		ReconnectDelay     int  `yaml:"reconnect_delay"`
		AutoReconnect      bool `yaml:"auto_reconnect"`
		SelfRestart        bool `yaml:"self_restart"`
		MaxContextFailures int  `yaml:"max_context_failures"`
		RestartDelay       int  `yaml:"restart_delay"`
	} `yaml:"advanced"`
	RepeatKey struct {
		Enabled              bool       `yaml:"enabled"`
		Key                 string      `yaml:"key"`                  // Deprecated: use Hotkeys instead
		Hotkeys             []Hotkey    `yaml:"hotkeys"`              // New flexible hotkey support
		ContentTimeout      int         `yaml:"content_timeout"`
		Notification        bool        `yaml:"notification"`
		RequirePreviousScan bool        `yaml:"require_previous_scan"`
	} `yaml:"repeat_key"`
}

// Hotkey represents a configurable hotkey combination
type Hotkey struct {
	Key       string   `yaml:"key"`       // Primary key (e.g., "r", "f1", "ctrl", "home")
	Modifiers []string `yaml:"modifiers"` // Modifier keys (e.g., ["ctrl", "alt"])
	Name      string   `yaml:"name"`      // Optional name for this hotkey
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	config := &Config{}
	
	// NFC defaults
	config.NFC.Device = 0
	config.NFC.CapsLock = false
	config.NFC.Reverse = false
	config.NFC.Decimal = false
	config.NFC.DecimalPadding = 0
	config.NFC.EndChar = "none"
	config.NFC.InChar = "none"
	
	// Web defaults
	config.Web.OpenWebsite = false
	config.Web.WebsiteURL = "https://example.com"
	config.Web.Fullscreen = true
	
	// Notification defaults
	config.Notifications.Enabled = true
	config.Notifications.ShowSuccess = true
	config.Notifications.ShowErrors = true
	
	// Advanced defaults
	config.Advanced.RetryAttempts = 3
	config.Advanced.ReconnectDelay = 2
	config.Advanced.AutoReconnect = true
	config.Advanced.SelfRestart = true
	config.Advanced.MaxContextFailures = 5
	config.Advanced.RestartDelay = 10
	
	// Audio defaults
	config.Audio.Enabled = true
	config.Audio.SuccessSound = "beep"     // Built-in beep sound
	config.Audio.ErrorSound = "error"     // Built-in error sound
	config.Audio.Volume = 70               // 70% volume
	
	// Repeat key defaults
	config.RepeatKey.Enabled = true
	config.RepeatKey.Key = "home"  // Backward compatibility
	config.RepeatKey.Hotkeys = []Hotkey{
		{
			Key:       "home",
			Modifiers: []string{},
			Name:      "Home Key",
		},
	}
	config.RepeatKey.ContentTimeout = 300  // 5 minutes
	config.RepeatKey.Notification = true
	config.RepeatKey.RequirePreviousScan = true
	
	return config
}

// LoadConfig loads configuration from YAML file with fallback to command-line flags
func LoadConfig() (*Config, error) {
	config := DefaultConfig()
	
	// Try to load from config.yaml
	configPath := "config.yaml"
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Loading configuration from %s\n", configPath)
		if err := loadConfigFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %v", err)
		}
	} else {
		fmt.Println("No config.yaml found, using defaults and command-line flags")
	}
	
	// Override with command-line flags if provided
	overrideWithFlags(config)
	
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}
	
	return config, nil
}

// loadConfigFromFile loads configuration from a YAML file
func loadConfigFromFile(config *Config, filename string) error {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return err
	}
	
	return yaml.Unmarshal(data, config)
}

// overrideWithFlags applies command-line flags over configuration file settings
func overrideWithFlags(config *Config) {
	var endChar, inChar string
	var autoRestart bool
	
	// Define flags
	flag.StringVar(&endChar, "end-char", config.NFC.EndChar, "Character at the end of UID. Options: "+CharFlagOptions())
	flag.StringVar(&inChar, "in-char", config.NFC.InChar, "Character between bytes of UID. Options: "+CharFlagOptions())
	flag.BoolVar(&config.NFC.CapsLock, "caps-lock", config.NFC.CapsLock, "UID with Caps Lock")
	flag.BoolVar(&config.NFC.Reverse, "reverse", config.NFC.Reverse, "UID reverse order")
	flag.BoolVar(&config.NFC.Decimal, "decimal", config.NFC.Decimal, "UID in decimal format")
	flag.IntVar(&config.NFC.DecimalPadding, "decimal-padding", config.NFC.DecimalPadding, "Pad decimal numbers with leading zeros to this length (0 = no padding)")
	flag.IntVar(&config.NFC.Device, "device", config.NFC.Device, "Device number to use")
	flag.BoolVar(&config.Web.OpenWebsite, "open-website", config.Web.OpenWebsite, "Open website URL in browser on startup")
	flag.StringVar(&config.Web.WebsiteURL, "website-url", config.Web.WebsiteURL, "URL to open in browser")
	flag.BoolVar(&config.Web.Fullscreen, "fullscreen", config.Web.Fullscreen, "Open browser in fullscreen mode")
	flag.BoolVar(&autoRestart, "auto-restart", false, "Internal flag indicating automatic restart")
	
	// Parse flags
	flag.Parse()
	
	// If this is an auto-restart, disable browser opening
	if autoRestart {
		config.Web.OpenWebsite = false
	}
	
	// Apply character flags
	if endChar != config.NFC.EndChar {
		config.NFC.EndChar = endChar
	}
	if inChar != config.NFC.InChar {
		config.NFC.InChar = inChar
	}
}

// validateConfig validates the configuration values
func validateConfig(config *Config) error {
	// Validate character flags
	if _, ok := StringToCharFlag(config.NFC.EndChar); !ok {
		return fmt.Errorf("invalid end character: %s", config.NFC.EndChar)
	}
	
	if _, ok := StringToCharFlag(config.NFC.InChar); !ok {
		return fmt.Errorf("invalid in character: %s", config.NFC.InChar)
	}
	
	// Validate device number
	if config.NFC.Device < 0 {
		return fmt.Errorf("device number must be positive, got: %d", config.NFC.Device)
	}
	
	// Validate decimal padding
	if config.NFC.DecimalPadding < 0 {
		return fmt.Errorf("decimal padding must be non-negative, got: %d", config.NFC.DecimalPadding)
	}
	
	// Validate retry attempts
	if config.Advanced.RetryAttempts < 1 {
		return fmt.Errorf("retry attempts must be at least 1, got: %d", config.Advanced.RetryAttempts)
	}
	
	// Validate reconnect delay
	if config.Advanced.ReconnectDelay < 0 {
		return fmt.Errorf("reconnect delay must be non-negative, got: %d", config.Advanced.ReconnectDelay)
	}
	
	// Validate self-restart settings
	if config.Advanced.MaxContextFailures < 1 {
		return fmt.Errorf("max context failures must be at least 1, got: %d", config.Advanced.MaxContextFailures)
	}
	
	if config.Advanced.RestartDelay < 0 {
		return fmt.Errorf("restart delay must be non-negative, got: %d", config.Advanced.RestartDelay)
	}
	
	// Validate repeat key settings
	if config.RepeatKey.ContentTimeout < 0 {
		return fmt.Errorf("repeat key content timeout must be non-negative, got: %d", config.RepeatKey.ContentTimeout)
	}
	
	// Validate hotkey configurations
	if err := validateHotkeys(config.RepeatKey.Hotkeys); err != nil {
		return fmt.Errorf("invalid hotkey configuration: %v", err)
	}
	
	// Handle backward compatibility: if Key is set but Hotkeys is empty, migrate
	if config.RepeatKey.Key != "" && len(config.RepeatKey.Hotkeys) == 0 {
		config.RepeatKey.Hotkeys = []Hotkey{
			{
				Key:       config.RepeatKey.Key,
				Modifiers: []string{},
				Name:      fmt.Sprintf("%s Key", strings.Title(config.RepeatKey.Key)),
			},
		}
	}
	
	return nil
}

// ToFlags converts Config to the legacy Flags struct for compatibility
func (c *Config) ToFlags() Flags {
	flags := Flags{
		CapsLock:       c.NFC.CapsLock,
		Reverse:        c.NFC.Reverse,
		Decimal:        c.NFC.Decimal,
		DecimalPadding: c.NFC.DecimalPadding,
		Device:         c.NFC.Device,
	}
	
	// Convert character flags
	endChar, _ := StringToCharFlag(c.NFC.EndChar)
	inChar, _ := StringToCharFlag(c.NFC.InChar)
	
	flags.EndChar = endChar
	flags.InChar = inChar
	
	return flags
}

// validateHotkeys validates the hotkey configurations
func validateHotkeys(hotkeys []Hotkey) error {
	validKeys := map[string]bool{
		// Letters
		"a": true, "b": true, "c": true, "d": true, "e": true, "f": true, "g": true, "h": true,
		"i": true, "j": true, "k": true, "l": true, "m": true, "n": true, "o": true, "p": true,
		"q": true, "r": true, "s": true, "t": true, "u": true, "v": true, "w": true, "x": true,
		"y": true, "z": true,
		// Numbers
		"0": true, "1": true, "2": true, "3": true, "4": true, "5": true, "6": true, "7": true,
		"8": true, "9": true,
		// Function keys
		"f1": true, "f2": true, "f3": true, "f4": true, "f5": true, "f6": true, "f7": true,
		"f8": true, "f9": true, "f10": true, "f11": true, "f12": true,
		// Special keys
		"home": true, "end": true, "insert": true, "delete": true, "backspace": true,
		"tab": true, "enter": true, "space": true, "escape": true, "pause": true,
		// Arrow keys
		"up": true, "down": true, "left": true, "right": true,
		// Page navigation
		"pageup": true, "pagedown": true,
		// Modifier keys (can be used as primary keys too)
		"ctrl": true, "alt": true, "shift": true, "cmd": true, "win": true,
		// Numpad
		"numpad0": true, "numpad1": true, "numpad2": true, "numpad3": true, "numpad4": true,
		"numpad5": true, "numpad6": true, "numpad7": true, "numpad8": true, "numpad9": true,
		"numpadmultiply": true, "numpadadd": true, "numpadsubtract": true, "numpaddivide": true,
		"numpaddecimal": true, "numpadenter": true,
	}
	
	validModifiers := map[string]bool{
		"ctrl": true, "alt": true, "shift": true, "cmd": true, "win": true,
	}
	
	for i, hotkey := range hotkeys {
		if hotkey.Key == "" {
			return fmt.Errorf("hotkey %d: key cannot be empty", i+1)
		}
		
		// Normalize key to lowercase
		key := strings.ToLower(hotkey.Key)
		if !validKeys[key] {
			return fmt.Errorf("hotkey %d: unsupported key '%s'", i+1, hotkey.Key)
		}
		
		// Validate modifiers
		for j, modifier := range hotkey.Modifiers {
			mod := strings.ToLower(modifier)
			if !validModifiers[mod] {
				return fmt.Errorf("hotkey %d, modifier %d: unsupported modifier '%s'", i+1, j+1, modifier)
			}
		}
	}
	
	return nil
}