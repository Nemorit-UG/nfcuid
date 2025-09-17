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

### CAPS Lock Protection
- **Automatic CAPS Lock management**: Detects and temporarily disables CAPS Lock during input
- **State restoration**: Automatically restores original CAPS Lock state after input
- **Prevents character corruption**: Ensures consistent input regardless of CAPS Lock state
- **Cross-platform support**: Works on Windows, Linux, and macOS

### Repeat Last Input
- **Configurable hotkey**: Repeat the last successful card input with a keyboard shortcut
- **Global hotkey detection**: Uses special keys like F12, Pos1, etc. for activation
- **Always stays in foreground**: Hotkey monitoring runs in background
- **No card required**: Replay the last output without scanning a new card
- **Configurable**: Set your preferred hotkey in the configuration file

### Advanced Configuration Options
- Configurable retry attempts and reconnection delays
- Success/error notification preferences
- Auto-reconnection toggle
- Fullscreen browser mode selection

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

# Hotkey Settings
hotkeys:
  repeat_last_input: "F12"    # Keyboard shortcut to repeat last input
                              # Options: F1-F12, Pos1, End, Insert, Delete, etc.
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

### Using Repeat Last Input
The application supports repeating the last successful card input:

1. **Scan a card** - The UID is automatically stored for repeat functionality
2. **Press the configured hotkey** - Default is F12, can be changed to Pos1 or other special keys
3. **The last input is replayed** - The stored UID is sent as keyboard input to the active field

```yaml
# Configure the repeat hotkey
hotkeys:
  repeat_last_input: "F12"  # Change to your preferred key (F12, Pos1, etc.)
```

This feature is perfect for:
- **Data entry workflows**: Quickly re-enter the same card multiple times
- **Testing applications**: Repeat the same input without re-scanning
- **Kiosk applications**: Allow users to retry failed entries

*Note: Global hotkey detection requires platform-specific implementation. The framework is ready for Windows (RegisterHotKey), Linux (X11 XGrabKey), and macOS (Carbon/Cocoa) implementations.*

### Kiosk Mode Example
```yaml
# config.yaml for kiosk application
nfc:
  device: 1
  end_char: "enter"
web:
  open_website: true
  website_url: "https://your-kiosk-app.com/checkin"
  fullscreen: true
notifications:
  show_success: false  # Quiet mode
```

### Development/Testing
```yaml
# config.yaml for development
nfc:
  device: 0           # Manual device selection
  caps_lock: true
  in_char: "hyphen"
web:
  open_website: true
  website_url: "http://localhost:3000"
  fullscreen: false
advanced:
  retry_attempts: 1   # Fail fast for debugging
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
