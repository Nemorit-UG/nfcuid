# Service Installation Guide for NFC UID Reader
1. Service Artefakt herunterladen und in C:\\Program Files\\NFCUID kopieren

2. Verknüpfung in Autostart erstellen
win + r
```
shell:startup
```

3. Config.yml anlegen
```yaml
# NFC UID Reader Configuration
# Copy this file to config.yaml and modify as needed

# NFC Reader Settings
nfc:
  # Device number (0 for manual selection, or specific device number)
  device: 1
  
  # Output formatting options
  caps_lock: false     # UID output with uppercase letters
  reverse: false       # Reverse the UID byte order
  decimal: true        # Output UID in decimal format instead of hex
  decimal_padding: 10  # Pad decimal numbers with leading zeros to this length (0 = no padding)
  
  # Character options: none, space, tab, hyphen, enter, semicolon, colon, comma
  end_char: "enter"     # Character to append at end of UID
  in_char: "none"      # Character to insert between UID bytes

# Web Browser Integration
web:
  # Whether to open a browser window when the application starts
  open_website: true
  
  # URL to open in the browser
  website_url: "https://app.kitafino.de/sys_k2/index.php?action=login"
  
  # Try to open browser in fullscreen mode
  fullscreen: true

# System Notifications
notifications:
  # Enable system notifications
  enabled: true
  
  # Show notifications for successful card reads
  show_success: false
  
  # Show notifications for errors and issues
  show_errors: true

# Advanced Settings
advanced:
  # Number of times to retry failed card reads
  retry_attempts: 3
  
  # Seconds to wait before attempting to reconnect after disconnection
  reconnect_delay: 2
  
  # Automatically attempt to reconnect to readers when disconnected
  auto_reconnect: true
  self_restart: true              # Enable automatic application restart
  max_context_failures: 5        # Max consecutive PC/SC context failures before restart
  restart_delay: 10               # Seconds to wait before restarting

audio:
  # Enable audio feedback for successful scans and errors
  enabled: true
  
  # Success sound options: "beep", "none", or path to custom sound file
  success_sound: "C:\\Program Files\\NFCUID\\sound.mp3"
  
  # Error sound options: "error", "none", or path to custom sound file
  error_sound: "error"
  
  # Volume level (0-100, currently not implemented but reserved for future use)
  volume: 70
```

4. sound.mp3 in C:\\Program Files\\NFCUID kopieren
5. Ggf. Systemsound für anschließen eines neuen Gerätes und trennen eines extenen Gerätes abschalten (Kantine).

FERTIG!
