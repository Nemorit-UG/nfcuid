# nfcUID - Enhanced Version
Cross-platform terminal application for reading NFC tag UID with web browser integration and robust error handling.

## Overview
Application reads NFC tag UID using PC/SC API and provides keyboard output to any text field. Features include:
- **YAML Configuration**: Configure all settings via `config.yaml` file
- **Web Browser Integration**: Automatically open URLs in maximized/fullscreen browser windows
- **System Notifications**: User-friendly error handling with desktop notifications
- **Robust Error Recovery**: Automatic reconnection and retry mechanisms
- **Cross-platform Support**: Windows, MacOS, and Linux

PC/SC is a standard interface for smartcards, available on most operating systems. UID is written to the active text input field by generating keystrokes.

## New Features (Enhanced Version)

### Configuration File Support
- Create a `config.yaml` file for persistent settings
- All command-line flags now available as YAML configuration
- Command-line flags override config file settings when provided
- Copy `config.yaml.example` to get started

### Flexible Hotkey System (ENHANCED)
- **Multiple Hotkeys**: Configure any number of custom hotkey combinations
- **Flexible Key Support**: Support for letters, numbers, function keys, arrows, and special keys
- **Modifier Combinations**: Support Ctrl, Alt, Shift, Win/Cmd key combinations
- **Standalone Modifiers**: Use Ctrl, Alt, or Shift keys as standalone triggers
- **Configurable timeout**: Set how long scanned content remains available for repeat
- **Smart notifications**: Optional notifications when hotkey is used
- **Thread-safe storage**: Secure handling of last scanned content
- **Cross-platform support**: Works on Windows, macOS, and Linux
- **Backward Compatible**: Old `key: "home"` configuration still supported

### Virtual Scanner for Testing (NEW)
- **Hardware-free testing**: Test the application without physical NFC readers
- **Configurable test cards**: Define custom UIDs for simulation
- **Auto-cycle mode**: Automatically cycle through test cards
- **Manual mode**: Trigger card presentation manually
- **Full integration**: Works with all existing features including repeat key

### Web Browser Integration
- Automatically open websites when the application starts
- Support for maximized and fullscreen browser windows
- Cross-platform browser opening (Chrome, Firefox, Safari, Edge)
- Perfect for kiosk applications or web-based card management systems

### Enhanced Error Handling
- **No more crashes**: Graceful error handling with system notifications
- **Auto-reconnection**: Automatically reconnect when NFC readers disconnect
- **Retry mechanisms**: Configurable retry attempts for failed card reads
- **Desktop notifications**: Success and error notifications via system tray

### Advanced Configuration Options
- Configurable retry attempts and reconnection delays
- Success/error notification preferences
- Auto-reconnection toggle
- Fullscreen browser mode selection
- Repeat key settings and timeout configuration
- Virtual scanner test card definitions

## Supported readers
Application works with any PC/SC compatible reader. Tested with:
  - ACR122U
  - ACR1281U-C1
  - ACR1252U-M1

## Supported NFC tags
Application works with any NFC tag with UID. Tested with:
  - Mifare Classic
  - Mifare Ultralight
  - NTAG203, NTAG213, NTAG216

## Installation & Build

### Download
Binaries for Windows, MacOS and Linux platforms available from release page.

### Build from source
```bash
go get github.com/taglme/nfcuid
cd nfcuid
go mod tidy
go build
```

## Configuration

### YAML Configuration File
Create `config.yaml` (copy from `config.yaml.example`):

```yaml
# NFC Reader Settings
nfc:
  device: 0              # 0 for manual selection
  caps_lock: false       # Uppercase hex output
  reverse: false         # Reverse UID byte order
  decimal: false         # Decimal format instead of hex
  decimal_padding: 0     # Pad decimal numbers with leading zeros to this length (0 = no padding)
  end_char: "enter"      # Character after UID
  in_char: "hyphen"      # Character between bytes

# Flexible Hotkey System
repeat_key:
  enabled: true          # Enable hotkey functionality
  hotkeys:               # List of configurable hotkeys
    - key: "home"        # Primary key (letters, numbers, f1-f12, etc.)
      modifiers: []      # Modifier keys (ctrl, alt, shift, cmd/win)
      name: "Home Key"   # Optional display name
    - key: "r"           # Ctrl+R combination
      modifiers: ["ctrl"]
      name: "Ctrl+R"
    - key: "ctrl"        # Standalone Ctrl key (as requested)
      modifiers: []
      name: "Ctrl Key"
  content_timeout: 300   # Seconds to keep last content (0 = no timeout)
  notification: true     # Show notification when hotkey used
  require_previous_scan: true  # Only work if there was a previous scan

# Virtual Scanner for Testing (without hardware)
virtual_scanner:
  enabled: false        # Enable virtual scanner mode
  test_cards:           # List of test card UIDs (hex format)
    - "04a1b2c3"
    - "deadbeef"
    - "12345678"
  cycle_delay: 2000     # Milliseconds between auto card cycles
  auto_cycle: false     # Automatically cycle through cards

# Web Browser Integration
web:
  open_website: true                    # Open browser on startup
  website_url: "https://example.com"    # URL to open
  fullscreen: true                      # Fullscreen mode

# System Notifications
notifications:
  enabled: true          # Enable notifications
  show_success: true     # Notify on successful reads
  show_errors: true      # Notify on errors

# Advanced Settings
advanced:
  retry_attempts: 3           # Retry failed operations
  reconnect_delay: 2          # Seconds between reconnection attempts
  auto_reconnect: true        # Auto-reconnect on disconnection
  self_restart: true          # Enable self-restart on critical failures
  max_context_failures: 5     # Max PC/SC context failures before restart
  restart_delay: 10           # Seconds to wait before restarting
```

### Command-line Options
All YAML options available as flags (override config file):

```bash
# NFC Options
-device int            Device number (0 for manual selection)
-caps-lock bool        UID with uppercase letters
-reverse bool          Reverse UID byte order
-decimal bool          Output in decimal format
-end-char string       End character: none,space,tab,hyphen,enter,semicolon,colon,comma
-in-char string        Between-bytes character (same options as end-char)

# Web Options
-open-website bool     Open browser on startup
-website-url string    URL to open
-fullscreen bool       Use fullscreen browser mode

# Run with -h for complete help
nfcuid -h
```

## Usage Examples

### Basic Usage
```bash
# Use config.yaml settings
./nfcuid

# Override specific settings
./nfcuid -device=1 -end-char=enter
```

### Kiosk Mode with Repeat Key
```yaml
# config.yaml for kiosk application
nfc:
  device: 1
  end_char: "enter"
web:
  open_website: true
  website_url: "https://your-kiosk-app.com/checkin"
  fullscreen: true
repeat_key:
  enabled: true
  notification: false  # Silent repeat for kiosk mode
  content_timeout: 0   # Never expire
notifications:
  show_success: false  # Quiet mode
```

### Development/Testing with Virtual Scanner
```yaml
# config.yaml for development
virtual_scanner:
  enabled: true        # Use virtual scanner instead of hardware
  auto_cycle: true     # Automatically present cards
  cycle_delay: 1000    # 1 second between cards
  test_cards:
    - "04a1b2c3"       # Custom test UIDs
    - "deadbeef"
    - "12345678"
repeat_key:
  enabled: true
  content_timeout: 60  # Short timeout for testing
nfc:
  caps_lock: true
  in_char: "hyphen"
web:
  open_website: true
  website_url: "http://localhost:3000"
  fullscreen: false
advanced:
  retry_attempts: 1    # Fail fast for debugging
```

### Production with Repeat Key
```yaml
# config.yaml for production use
nfc:
  device: 1
  decimal: true
  decimal_padding: 10
  end_char: "enter"
repeat_key:
  enabled: true
  content_timeout: 300  # 5 minutes
  notification: true
  require_previous_scan: true
notifications:
  show_success: false
  show_errors: true
advanced:
  auto_reconnect: true
  self_restart: true
```

### Output Examples
```
# Hex format with hyphens
04-AE-65-CA-82-49-80

# Decimal format
310838458

# Hex format, no separators
04AE65CA824980
```

## Flexible Hotkey System

### How to Use
1. Scan an NFC card normally
2. Press any of your configured hotkey combinations to repeat the last scanned content
3. The content will be typed again as keyboard input

### Supported Keys
- **Letters**: a-z (case-insensitive)
- **Numbers**: 0-9
- **Function Keys**: F1-F12
- **Special Keys**: Home, End, Insert, Delete, Backspace, Tab, Enter, Space, Escape
- **Arrow Keys**: Up, Down, Left, Right
- **Page Navigation**: PageUp, PageDown
- **Modifier Keys**: Ctrl, Alt, Shift, Win/Cmd (can be used as standalone keys or modifiers)
- **Numpad**: Numpad0-9, NumpadMultiply, NumpadAdd, etc.

### Features
- **Multiple Hotkeys**: Configure unlimited hotkey combinations
- **Flexible Combinations**: Mix any key with any modifier combination
- **Standalone Modifiers**: Use Ctrl, Alt, Shift as trigger keys (not just modifiers)
- **Configurable timeout**: Content expires after set time (default: 5 minutes)
- **Smart notifications**: Shows when hotkey is used and content age
- **Requirement checking**: Can require previous successful scan
- **Thread-safe**: Works safely with concurrent NFC operations
- **Cross-platform**: Uses robotgo library for reliable key state monitoring

### Configuration Options
```yaml
repeat_key:
  enabled: true                    # Enable/disable feature
  
  # New flexible hotkey configuration
  hotkeys:
    - key: "home"                  # Simple key
      modifiers: []
      name: "Home Key"
    - key: "r"                     # Ctrl+R
      modifiers: ["ctrl"]
      name: "Ctrl+R"
    - key: "space"                 # Ctrl+Alt+Space
      modifiers: ["ctrl", "alt"]
      name: "Ctrl+Alt+Space"
    - key: "ctrl"                  # Standalone Ctrl key
      modifiers: []
      name: "Ctrl Key"
    - key: "f5"                    # F5 function key
      modifiers: []
      name: "F5 Refresh"
  
  # Backward compatibility (deprecated)
  # key: "home"                    # Old format still supported
  
  content_timeout: 300             # Seconds before content expires (0 = never)
  notification: true               # Show notification on repeat
  require_previous_scan: true      # Only work after successful scan
```

### Example Configurations
```yaml
# Basic single hotkey (backward compatible)
repeat_key:
  enabled: true
  key: "home"

# Multiple flexible hotkeys
repeat_key:
  enabled: true
  hotkeys:
    - key: "home"
      modifiers: []
    - key: "ctrl"           # Standalone Ctrl key
      modifiers: []
    - key: "r"
      modifiers: ["ctrl"]   # Ctrl+R
    - key: "f12"
      modifiers: []

# Advanced combinations
repeat_key:
  enabled: true
  hotkeys:
    - key: "insert"
      modifiers: ["shift"]        # Shift+Insert
    - key: "space"
      modifiers: ["ctrl", "alt"]  # Ctrl+Alt+Space
    - key: "alt"                  # Standalone Alt key
      modifiers: []
```

## Virtual Scanner for Testing

### How to Use
1. Enable virtual scanner in config: `virtual_scanner.enabled: true`
2. Define test cards: List hex UIDs in `test_cards`
3. Choose mode:
   - **Auto-cycle**: Cards appear automatically at intervals
   - **Manual**: Trigger card presentation manually

### Features
- **No hardware needed**: Test without physical NFC readers
- **Custom UIDs**: Define any hex UID for testing
- **Full integration**: Works with all features including repeat key
- **Realistic simulation**: Mimics real PC/SC card behavior

### Configuration Options
```yaml
virtual_scanner:
  enabled: false              # Enable virtual scanner mode
  test_cards:                # List of test UIDs (hex format, even length)
    - "04a1b2c3"
    - "deadbeef"
    - "12345678"
  cycle_delay: 2000          # Milliseconds between auto cycles
  auto_cycle: false          # Auto-cycle through cards vs manual
```

## Error Handling & Troubleshooting

### System Notifications
- **Success**: "Card UID: [uid-value]"
- **Errors**: Specific error descriptions
- **Connection**: Reader disconnect/reconnect status
- **Browser**: Browser opening confirmation

### Common Issues
1. **No readers found**: Check USB connections, install drivers
2. **Permission denied**: Run with appropriate permissions (especially Linux)
3. **Browser won't open**: Check URL format, browser availability
4. **Cards not reading**: Try different retry settings, check card compatibility

### Logging & Debug
- Console output shows detailed operation status
- Notifications provide user-friendly error messages
- Auto-recovery attempts logged with delays
- Configuration validation on startup

## Advanced Features

### Auto-Recovery
- Automatic reconnection when readers disconnect
- Configurable retry attempts for failed operations
- Exponential backoff for reconnection delays
- Graceful fallback when errors occur

### Self-Restart Mechanism
- **Automatic restart** on critical PC/SC context failures (default: after 5 consecutive failures)
- **Configurable threshold** via `max_context_failures` setting
- **Graceful restart** with notification and configurable delay
- **Process preservation** with same command-line arguments
- **Perfect for kiosk/service environments** where human intervention isn't available
- **Failure tracking** resets on successful context establishment

The self-restart feature monitors PC/SC context establishment failures. When `scard.EstablishContext()` fails consecutively for the configured number of times (default: 5), the application will:
1. Display notification about restart
2. Wait for configured delay (default: 10 seconds)
3. Launch new process with same arguments
4. Exit current process gracefully

This ensures maximum uptime in unattended environments.

### Cross-Platform Browser Support
- **Windows**: Chrome/Edge kiosk mode, fallback to default
- **macOS**: Chrome kiosk mode, Safari with AppleScript fullscreen
- **Linux**: Chrome/Firefox kiosk mode, xdotool F11 fallback

### Configuration Priority
1. Command-line flags (highest priority)
2. YAML configuration file
3. Built-in defaults (lowest priority)

## Enhanced Version
The enhanced version has been developed by [Nemorit UG (haftungsbeschränkt)](https://nemorit.de).
Further information and help is available via mail [info@nemorit.de](mailto:info@nemorit.de).

## Based on NFC Software by Tagl.me
The software is based on the original work by [Tagl.me](https://tagl.me)
