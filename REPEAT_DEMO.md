# Repeat Last Input - Demo

This document demonstrates the new "repeat last input" functionality.

## How it works

1. When you scan an NFC card, the application stores the formatted output
2. You can replay this output using the repeat functionality
3. No card scan is required for repeat - just trigger the repeat command

## Available triggers

### Text commands (always available)
- Type `repeat` + Enter in the terminal
- Type `r` + Enter for quick access

### Hotkey (configurable)
- Default: F12 key
- Configure in `config.yaml` under `hotkeys.repeat_last_input`
- Note: Global hotkey detection requires platform-specific implementation

## Configuration

Add to your `config.yaml`:

```yaml
hotkeys:
  repeat_last_input: "F12"  # Set your preferred hotkey
```

## Use cases

- **Data entry**: Quickly re-enter the same card ID multiple times
- **Testing**: Replay inputs without re-scanning cards
- **Error recovery**: Retry failed entries without scanning again
- **Kiosk applications**: Let users retry operations easily

## Example workflow

1. Scan card → Output: "1234567890"
2. Type `repeat` → Output: "1234567890" (no scan needed)
3. Type `r` → Output: "1234567890" (shorthand)
4. Press F12 → Output: "1234567890" (hotkey, when implemented)

## Implementation notes

The current implementation provides:
- ✅ Text-based commands (`repeat`, `r`)
- ✅ Configurable hotkey setting
- ✅ Cross-platform compatibility
- ⏳ Global hotkey detection (foundation ready for platform-specific implementation)

For production environments requiring true global hotkeys, platform-specific APIs would be integrated:
- Windows: RegisterHotKey or SetWindowsHookEx
- Linux: X11 XGrabKey
- macOS: Carbon RegisterEventHotKey