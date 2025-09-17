# Service Installation Guide for NFC UID Reader
Installations-Anleitung für Mensen der Kupferstadt Stolberg

1. Service Artefakt herunterladen und in C:\\NFCUID kopieren

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
  success_sound: "C:\\NFCUID\\sound.mp3"
  
  # Error sound options: "error", "none", or path to custom sound file
  error_sound: "error"
  
  # Volume level (0-100, currently not implemented but reserved for future use)
  volume: 70

# Update Checker Settings (automatic updates)
updates:
  # Enable automatic update checking
  enabled: true
  
  # Check for updates on application startup
  check_on_startup: true
  
  # Automatically download available updates
  auto_download: true
  
  # Automatically install downloaded updates (requires restart)
  # WARNING: Setting this to true will automatically replace the executable
  auto_install: false
  
  # Check interval in hours (for future periodic checks)
  check_interval_hours: 24
```

4. sound.mp3 in C:\\NFCUID kopieren
5. Ggf. Systemsound für anschließen eines neuen Gerätes und trennen eines extenen Gerätes abschalten (Kantine).

## Automatische Updates

Die Anwendung kann automatisch nach Updates suchen und diese installieren:

- **Update-Prüfung**: Automatische Prüfung beim Start der Anwendung
- **Download**: Automatischer Download verfügbarer Updates
- **Installation**: Optionale automatische Installation (standardmäßig deaktiviert für Sicherheit)
- **Benachrichtigungen**: Systembenachrichtigungen über Update-Status

### Update-Konfiguration in config.yaml:
```yaml
updates:
  enabled: true              # Update-Checker aktivieren
  check_on_startup: true     # Beim Start auf Updates prüfen
  auto_download: true        # Updates automatisch herunterladen
  auto_install: true         # Updates automatisch installieren und neustarten
```

**Wichtige Hinweise für C:\\Program Files Installation:**
- Automatische Installation erfordert möglicherweise Administrator-Rechte
- Bei Berechtigungsproblemen: Anwendung als Administrator ausführen oder in Benutzerverzeichnis installieren
- Alternative: Manuelle Installation mit `nfcuid.exe -update`

### Manuelle Update-Kontrolle:
```cmd
# Version anzeigen
nfcuid.exe -version

# Updates deaktivieren
nfcuid.exe -updates=false

# Update-Prüfung beim Start deaktivieren
nfcuid.exe -check-updates=false

# Sofortiges Update prüfen und installieren
nfcuid.exe -update
```

FERTIG!

## Hilfe & Support
Für Unterstützung und Fragen wenden Sie sich bitte an [info@nemorit.de](mailto:info@nemorit.de).
