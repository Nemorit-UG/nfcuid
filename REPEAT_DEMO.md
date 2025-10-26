# Repeat Last Input - Demo

This document demonstrates the new "repeat last input" functionality.

## How it works

1. When you scan an NFC card, the application stores the formatted output
2. You can replay this output using the configured hotkey
3. No card scan is required for repeat - just press the hotkey

## Available triggers

### Global hotkey (configurable)
- Default: F12 key
- Alternative options: Pos1, F1-F12, or other special keys
- Configure in `config.yaml` under `hotkeys.repeat_last_input`
- Requires platform-specific implementation for actual global detection

## Configuration

Add to your `config.yaml`:

```yaml
hotkeys:
  repeat_last_input: "F12"  # Set your preferred hotkey (F12, Pos1, etc.)
```

## Use cases

- **Data entry**: Quickly re-enter the same card ID multiple times
- **Testing**: Replay inputs without re-scanning cards
- **Error recovery**: Retry failed entries without scanning again
- **Kiosk applications**: Let users retry operations easily

## Example workflow

1. Scan card → Output: "1234567890"
2. Press F12 → Output: "1234567890" (no scan needed)
3. Press F12 → Output: "1234567890" (repeat again)

## Implementation notes

The current implementation provides:
- ✅ Configurable hotkey setting
- ✅ Cross-platform compatibility framework
- ✅ Background monitoring structure
- ⏳ Global hotkey detection (foundation ready for platform-specific implementation)

For production environments requiring global hotkeys, platform-specific APIs need to be integrated:
- Windows: RegisterHotKey or SetWindowsHookEx
- Linux: X11 XGrabKey
- macOS: Carbon RegisterEventHotKey