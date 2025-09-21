package main

import (
	"github.com/micmonay/keybd_event"
)

type keySet struct {
	code  int
	shift bool
}


//KeyboardWrite emulate keyboard input from string with CAPS Lock protection
func KeyboardWrite(textInput string, kb keybd_event.KeyBonding) error {
	// Create CAPS Lock manager
	capsManager := NewCapsLockManager(kb)
	
	// Disable CAPS Lock if it's on
	if err := capsManager.DisableCapsLock(); err != nil {
		return err
	}
	
	// Defer restoration of CAPS Lock state
	defer func() {
		capsManager.RestoreCapsLock() // Ignore error in defer
	}()

	//Should we skip next character in string
	//Used if we found some escape sequence
	skip := false
	for i, c := range textInput {
		if !skip {
			if c != '\\' {
				kb.SetKeys(names[string(c)].code)
				kb.HasSHIFT(names[string(c)].shift)
			} else {
				//Found backslash escape character
				//Check next character
				switch textInput[i+1] {
				case 'n':
					//Found newline character sequence
					kb.SetKeys(names["ENTER"].code)
					skip = true
				case '\\':
					//Found backslash character sequence
					kb.SetKeys(names["\\"].code)
					kb.HasSHIFT(names["\\"].shift)
					skip = true
				case 'b':
					//Found backspace character sequence
					kb.SetKeys(names["BACKSPACE"].code)
					skip = true
				case 't':
					//Found tab character sequence
					kb.SetKeys(names["TAB"].code)
					skip = true
				case '"':
					//Found double quote character sequence
					kb.SetKeys(names["\""].code)
					kb.HasSHIFT(names["\""].shift)
					skip = true
				default:
					//Nothing special, jsut backslash output
					kb.SetKeys(names["\\"].code)
					kb.HasSHIFT(names["\\"].shift)
				}

			}
			var err = kb.Launching()
			if err != nil {
				return err
			}
		} else {
			skip = false
		}

	}
	return nil

}
