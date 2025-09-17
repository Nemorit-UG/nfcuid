package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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
	Hotkeys struct {
		RepeatLastInput string `yaml:"repeat_last_input"`
	} `yaml:"hotkeys"`
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
	
	// Hotkey defaults
	config.Hotkeys.RepeatLastInput = "F12" // Default hotkey for repeat function
	
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
	
	// Validate hotkey configuration
	if config.Hotkeys.RepeatLastInput == "" {
		return fmt.Errorf("repeat_last_input hotkey cannot be empty")
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