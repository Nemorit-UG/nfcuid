package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ebfe/scard"
	"github.com/micmonay/keybd_event"
)

type Service interface {
	Start()
	Flags() Flags
}

func NewService(flags Flags, config *Config, notificationManager *NotificationManager, restartManager *RestartManager, audioManager *AudioManager, logManager *LogManager, uiManager *UIManager) Service {
	return &service{
		flags:               flags,
		config:              config,
		notificationManager: notificationManager,
		restartManager:      restartManager,
		audioManager:        audioManager,
		logManager:          logManager,
		uiManager:           uiManager,
		retryManager:        NewRetryManager(config.Advanced.RetryAttempts, config.Advanced.ReconnectDelay),
	}
}

type Flags struct {
	CapsLock       bool
	Reverse        bool
	Decimal        bool
	DecimalPadding int
	EndChar        CharFlag
	InChar         CharFlag
	Device         int
}

type service struct {
	flags               Flags
	config              *Config
	notificationManager *NotificationManager
	restartManager      *RestartManager
	audioManager        *AudioManager
	logManager          *LogManager
	uiManager           *UIManager
	retryManager        *RetryManager
}

func UIDToUint32(uid []byte) (uint32, error) {
	if len(uid) != 4 {
		return 0, fmt.Errorf("UID must be 4 bytes, got %d bytes", len(uid))
	}
	return binary.LittleEndian.Uint32(uid), nil
}

func (s *service) Start() {
	s.logManager.LogInfo("Service starting main loop")
	for {
		if err := s.runServiceLoop(); err != nil {
			s.logManager.LogError("Service loop error", err)
			s.notificationManager.NotifyErrorThrottled("service-error", "Verbindung zum NFC-Lesegerät verloren. Bitte Gerät überprüfen.")
			fmt.Printf("Service encountered an error: %v\n", err)

			if s.config.Advanced.AutoReconnect {
				s.logManager.LogWarning("Attempting to restart service", "delay_seconds", fmt.Sprintf("%d", s.config.Advanced.ReconnectDelay))
				fmt.Printf("Attempting to restart service in %d seconds...\n", s.config.Advanced.ReconnectDelay)
				time.Sleep(time.Duration(s.config.Advanced.ReconnectDelay) * time.Second)
				continue
			} else {
				s.logManager.LogError("Service stopped due to error with auto-reconnect disabled", err)
				SafeExit(1, "Service stopped due to error", s.notificationManager)
			}
		}
	}
}

func (s *service) runServiceLoop() error {
	s.logManager.LogInfo("Starting service loop")

	// Establish PC/SC context with retry logic
	var ctx *scard.Context
	err := s.retryManager.Retry(func() error {
		var err error
		ctx, err = scard.EstablishContext()
		if err != nil {
			s.logManager.LogError("Failed to establish PC/SC context", err)
			// Track context establishment failure
			if s.restartManager.TrackContextFailure(err) {
				// Restart was triggered, this will never return
				return nil
			}
		} else {
			s.logManager.LogInfo("PC/SC context established successfully")
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to establish PC/SC context: %v", err)
	}

	// Context established successfully, reset failure counter
	s.restartManager.ResetFailureCount()
	defer func() {
		ctx.Release()
		s.logManager.LogInfo("PC/SC context released")
	}()

	// List available readers
	readers, err := ctx.ListReaders()
	if err != nil {
		s.logManager.LogError("Failed to list readers", err)
		// Track reader enumeration failure
		if s.restartManager.TrackSystemFailure("Reader Enumeration", err) {
			// Restart was triggered, this will never return
			return nil
		}
		return fmt.Errorf("failed to list readers: %v", err)
	}

	if len(readers) < 1 {
		s.logManager.LogWarning("No NFC readers found")
		return errors.New("Kein NFC-Lesegerät gefunden. Bitte Gerät anschließen und Anwendung neu starten.")
	}

	s.logManager.LogInfo("Found NFC readers", "count", fmt.Sprintf("%d", len(readers)))
	fmt.Printf("Found %d device(s):\n", len(readers))
	for i, reader := range readers {
		fmt.Printf("[%d] %s\n", i+1, reader)
		s.logManager.LogInfo("Available reader", "index", fmt.Sprintf("%d", i+1), "name", reader)
	}

	// Select device
	if err := s.selectDevice(readers); err != nil {
		s.logManager.LogError("Device selection failed", err)
		return err
	}

	s.logManager.LogInfo("Device selected", "device_index", fmt.Sprintf("%d", s.flags.Device), "device_name", readers[s.flags.Device-1])
	fmt.Printf("Selected device: [%d] %s\n", s.flags.Device, readers[s.flags.Device-1])
	selectedReaders := []string{readers[s.flags.Device-1]}

	// Update UI manager with device info
	s.uiManager.SetDeviceInfo(readers[s.flags.Device-1], s.flags.Device, readers)
	s.uiManager.UpdateStatus("Device selected")

	// Initialize keyboard
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		s.logManager.LogError("Failed to initialize keyboard", err)
		return fmt.Errorf("failed to initialize keyboard: %v", err)
	}

	s.logManager.LogInfo("Keyboard initialized successfully")

	// Linux requires a delay for keyboard initialization
	if runtime.GOOS == "linux" {
		s.logManager.LogInfo("Applying Linux keyboard delay")
		time.Sleep(2 * time.Second)
	}

	// Main card reading loop
	s.logManager.LogScanningStatus("Starting card reading loop", readers[s.flags.Device-1])
	s.uiManager.UpdateStatus("Ready for scanning")
	return s.cardReadingLoop(ctx, selectedReaders, kb)
}

func (s *service) Flags() Flags {
	return s.flags
}

func (s *service) formatOutput(rx []byte) string {
	var output string
	var errorHexFallback bool = false
	//Reverse UID in flag set
	if s.flags.Reverse {
		for i, j := 0, len(rx)-1; i < j; i, j = i+1, j-1 {
			rx[i], rx[j] = rx[j], rx[i]
		}
	}

	if s.flags.Decimal {
		number, err := UIDToUint32(rx)
		if err != nil {
			s.notificationManager.NotifyError("Fehler beim Umwandeln der Karten-ID. Verwende Standard-Format.")
			// Fallback to hex format
			errorHexFallback = true
		} else {
			if s.flags.DecimalPadding > 0 {
				output = fmt.Sprintf("%0*d", s.flags.DecimalPadding, number)
			} else {
				output = fmt.Sprintf("%d", number)
			}
		}
	}

	if !s.flags.Decimal || errorHexFallback {
		for i, rxByte := range rx {
			var byteStr string
			if s.flags.CapsLock {
				byteStr = fmt.Sprintf("%02X", rxByte)
			} else {
				byteStr = fmt.Sprintf("%02x", rxByte)
			}

			output = output + byteStr
			if i < len(rx)-1 {
				output = output + s.flags.InChar.Output()
			}
		}
	}

	output = output + s.flags.EndChar.Output()
	return output
}

func (s *service) waitUntilCardPresent(ctx *scard.Context, readers []string) (int, error) {
	rs := make([]scard.ReaderState, len(readers))
	for i := range rs {
		rs[i].Reader = readers[i]
		rs[i].CurrentState = scard.StateUnaware
	}

	for {
		for i := range rs {
			if rs[i].EventState&scard.StatePresent != 0 {
				return i, nil
			}
			rs[i].CurrentState = rs[i].EventState
		}
		err := ctx.GetStatusChange(rs, -1)
		if err != nil {
			// Track reader status monitoring failure
			if s.restartManager.TrackSystemFailure("Reader Status Monitoring", err) {
				// Restart was triggered, this will never return
				return -1, nil
			}
			return -1, err
		}
	}
}

func (s *service) waitUntilCardRelease(ctx *scard.Context, readers []string, index int) error {
	rs := make([]scard.ReaderState, 1)

	rs[0].Reader = readers[index]
	rs[0].CurrentState = scard.StatePresent

	for {

		if rs[0].EventState&scard.StateEmpty != 0 {
			return nil
		}
		rs[0].CurrentState = rs[0].EventState

		err := ctx.GetStatusChange(rs, -1)
		if err != nil {
			// Track reader status monitoring failure
			if s.restartManager.TrackSystemFailure("Reader Status Monitoring", err) {
				// Restart was triggered, this will never return
				return nil
			}
			return err
		}
	}
}

func (s *service) selectDevice(readers []string) error {
	if s.flags.Device == 0 {
		// Interactive device selection
		for {
			fmt.Print("Enter device number to start: ")
			inputReader := bufio.NewReader(os.Stdin)
			deviceStr, _ := inputReader.ReadString('\n')

			if runtime.GOOS == "windows" {
				deviceStr = strings.Replace(deviceStr, "\r\n", "", -1)
			} else {
				deviceStr = strings.Replace(deviceStr, "\n", "", -1)
			}

			deviceInt, err := strconv.Atoi(deviceStr)
			if err != nil {
				fmt.Println("Please input integer value")
				continue
			}
			if deviceInt < 1 || deviceInt > len(readers) {
				fmt.Printf("Value should be between 1 and %d\n", len(readers))
				continue
			}
			s.flags.Device = deviceInt
			break
		}
	} else if s.flags.Device < 1 || s.flags.Device > len(readers) {
		return fmt.Errorf("device number should be between 1 and %d, got: %d", len(readers), s.flags.Device)
	}

	return nil
}

func (s *service) cardReadingLoop(ctx *scard.Context, selectedReaders []string, kb keybd_event.KeyBonding) error {
	deviceName := selectedReaders[0]
	s.logManager.LogScanningStatus("Ready for card scanning", deviceName)
	s.uiManager.SetScanningState(true)

	for {
		// Display status with visual separator
		s.uiManager.DisplayCurrentStatus()
		fmt.Println("┌─────────────────────────────────────────────────────────┐")
		fmt.Println("│  📡 NFC Reader Status: AKTIV / ACTIVE                   │")
		fmt.Println("│  🔍 Scanning Mode: BEREIT / READY                       │")
		fmt.Println("│  ⏳ Waiting for card... / Warte auf Karte...            │")
		fmt.Println("└─────────────────────────────────────────────────────────┘")
		s.logManager.LogScanningStatus("Waiting for card", deviceName)
		s.uiManager.UpdateStatus("Waiting for card")

		// Wait for card present with error handling
		index, err := s.waitForCardWithRetry(ctx, selectedReaders)
		if err != nil {
			s.logManager.LogError("Card detection failed", err, "device", deviceName)
			s.notificationManager.NotifyErrorThrottled("card-error", "Karte konnte nicht erkannt werden. Bitte NFC-Lesegerät überprüfen.")
			if s.config.Advanced.AutoReconnect {
				s.logManager.LogWarning("Retrying card detection due to auto-reconnect enabled")
				continue
			}
			return err
		}

		// Update status for card detected
		fmt.Println("✅ Card detected! / Karte erkannt!")
		s.logManager.LogScanningStatus("Card detected", deviceName)
		s.uiManager.UpdateStatus("Processing card")

		// Process the card
		if err := s.processCard(ctx, selectedReaders, index, kb); err != nil {
			s.logManager.LogError("Card processing failed", err, "device", deviceName)
			s.uiManager.SetLastError(err.Error())
			s.notificationManager.NotifyErrorThrottled("card-error", "Karte konnte nicht gelesen werden. Bitte erneut versuchen.")
			fmt.Printf("❌ Card processing failed: %v\n", err)
			// Continue to next card instead of exiting
			continue
		}

		fmt.Println("✅ Card processed successfully! / Karte erfolgreich verarbeitet!")
		s.uiManager.UpdateStatus("Waiting for card")
	}
}

func (s *service) waitForCardWithRetry(ctx *scard.Context, readers []string) (int, error) {
	var index int
	err := s.retryManager.Retry(func() error {
		var err error
		index, err = s.waitUntilCardPresent(ctx, readers)
		return err
	})
	return index, err
}

func (s *service) processCard(ctx *scard.Context, selectedReaders []string, index int, kb keybd_event.KeyBonding) error {
	deviceName := selectedReaders[index]
	fmt.Println("🔌 Connecting to card... / Verbindung zur Karte...")
	s.logManager.LogInfo("Attempting to connect to card", "device", deviceName)

	// Connect to card with retry
	var card *scard.Card
	err := s.retryManager.Retry(func() error {
		var err error
		card, err = ctx.Connect(selectedReaders[index], scard.ShareShared, scard.ProtocolAny)
		if err != nil {
			s.logManager.LogError("Failed to connect to card", err, "device", deviceName)
			// Track reader connection failure
			if s.restartManager.TrackSystemFailure("Reader Connection", err) {
				// Restart was triggered, this will never return
				return nil
			}
		} else {
			s.logManager.LogInfo("Connected to card successfully", "device", deviceName)
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to connect to card: %v", err)
	}
	defer func() {
		card.Disconnect(scard.ResetCard)
		s.logManager.LogInfo("Disconnected from card", "device", deviceName)
	}()

	// Read UID with retry
	s.logManager.LogInfo("Reading card UID", "device", deviceName)
	uidBytes, err := s.readCardUID(card)
	if err != nil {
		s.logManager.LogError("Failed to read card UID", err, "device", deviceName)
		return err
	}

	uidHex := fmt.Sprintf("% x", uidBytes)
	s.logManager.LogCardRead(uidHex, deviceName)
	fmt.Printf("📊 UID is: %s\n", uidHex)

	// Format and send keyboard output
	output := s.formatOutput(uidBytes)
	fmt.Print("⌨️  Writing as keyboard input... / Eingabe als Tastatur...")
	s.logManager.LogInfo("Sending keyboard output", "output", output, "device", deviceName)

	if err := KeyboardWrite(output, kb); err != nil {
		s.logManager.LogError("Keyboard output failed", err, "output", output, "device", deviceName)
		s.notificationManager.NotifyErrorThrottled("keyboard-error", "Karten-ID konnte nicht eingegeben werden. Cursor im richtigen Feld?")
		s.audioManager.PlayErrorSound()
		return fmt.Errorf("failed to write keyboard output: %v", err)
	}

	fmt.Println("✅ Success! / Erfolgreich!")
	s.logManager.LogInfo("Card processing completed successfully", "uid", uidHex, "output", output, "device", deviceName)
	s.uiManager.SetLastCardUID(output)
	s.notificationManager.NotifySuccess(fmt.Sprintf("Card UID: %s", output))
	s.audioManager.PlaySuccessSound()

	// Wait for card removal
	fmt.Print("⏳ Waiting for card release... / Warte auf Karten-Entfernung...")
	s.logManager.LogInfo("Waiting for card removal", "device", deviceName)
	err = s.waitUntilCardRelease(ctx, selectedReaders, index)
	if err != nil {
		s.logManager.LogError("Error waiting for card removal", err, "device", deviceName)
		s.notificationManager.NotifyError("Fehler beim Warten auf Karten-Entfernung. Karte wurde trotzdem gelesen.")
	} else {
		fmt.Println("✅ Card released / Karte entfernt")
		s.logManager.LogInfo("Card removed successfully", "device", deviceName)
	}

	return nil
}

func (s *service) readCardUID(card *scard.Card) ([]byte, error) {
	var uidBytes []byte

	err := s.retryManager.Retry(func() error {
		// GET DATA command
		cmd := []byte{0xFF, 0xCA, 0x00, 0x00, 0x00}

		rsp, err := card.Transmit(cmd)
		if err != nil {
			return fmt.Errorf("card transmission failed: %v", err)
		}

		if len(rsp) < 2 {
			return errors.New("insufficient response bytes from card")
		}

		// Check response code - two last bytes of response
		rspCodeBytes := rsp[len(rsp)-2:]
		successResponseCode := []byte{0x90, 0x00}
		if !bytes.Equal(rspCodeBytes, successResponseCode) {
			return fmt.Errorf("card operation failed, response code: % x", rspCodeBytes)
		}

		uidBytes = rsp[0 : len(rsp)-2]
		return nil
	})

	return uidBytes, err
}
