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

	"github.com/micmonay/keybd_event"
	"github.com/ebfe/scard"
)

type Service interface {
	Start()
	Flags() Flags
}

func NewService(flags Flags, config *Config, notificationManager *NotificationManager, restartManager *RestartManager) Service {
	return &service{
		flags:               flags,
		config:              config,
		notificationManager: notificationManager,
		restartManager:      restartManager,
		retryManager:        NewRetryManager(config.Advanced.RetryAttempts, config.Advanced.ReconnectDelay),
	}
}

type Flags struct {
	CapsLock bool
	Reverse  bool
	Decimal  bool
	EndChar  CharFlag
	InChar   CharFlag
	Device   int
}

type service struct {
	flags               Flags
	config              *Config
	notificationManager *NotificationManager
	restartManager      *RestartManager
	retryManager        *RetryManager
}

func UIDToUint32(uid []byte) (uint32, error) {
    if len(uid) != 4 {
        return 0, fmt.Errorf("UID must be 4 bytes, got %d bytes", len(uid))
    }
    return binary.LittleEndian.Uint32(uid), nil
}

func (s *service) Start() {
	for {
		if err := s.runServiceLoop(); err != nil {
			s.notificationManager.NotifyErrorThrottled("service-error", "Verbindung zum NFC-Lesegerät verloren. Bitte Gerät überprüfen.")
			fmt.Printf("Service encountered an error: %v\n", err)
			
			if s.config.Advanced.AutoReconnect {
				fmt.Printf("Attempting to restart service in %d seconds...\n", s.config.Advanced.ReconnectDelay)
				time.Sleep(time.Duration(s.config.Advanced.ReconnectDelay) * time.Second)
				continue
			} else {
				SafeExit(1, "Service stopped due to error", s.notificationManager)
			}
		}
	}
}

func (s *service) runServiceLoop() error {
	// Establish PC/SC context with retry logic
	var ctx *scard.Context
	err := s.retryManager.Retry(func() error {
		var err error
		ctx, err = scard.EstablishContext()
		if err != nil {
			// Track context establishment failure
			if s.restartManager.TrackContextFailure(err) {
				// Restart was triggered, this will never return
				return nil
			}
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to establish PC/SC context: %v", err)
	}
	
	// Context established successfully, reset failure counter
	s.restartManager.ResetFailureCount()
	defer ctx.Release()

	// List available readers
	readers, err := ctx.ListReaders()
	if err != nil {
		// Track reader enumeration failure
		if s.restartManager.TrackSystemFailure("Reader Enumeration", err) {
			// Restart was triggered, this will never return
			return nil
		}
		return fmt.Errorf("failed to list readers: %v", err)
	}

	if len(readers) < 1 {
		return errors.New("Kein NFC-Lesegerät gefunden. Bitte Gerät anschließen und Anwendung neu starten.")
	}

	fmt.Printf("Found %d device(s):\n", len(readers))
	for i, reader := range readers {
		fmt.Printf("[%d] %s\n", i+1, reader)
	}

	// Select device
	if err := s.selectDevice(readers); err != nil {
		return err
	}

	fmt.Printf("Selected device: [%d] %s\n", s.flags.Device, readers[s.flags.Device-1])
	selectedReaders := []string{readers[s.flags.Device-1]}

	// Initialize keyboard
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return fmt.Errorf("failed to initialize keyboard: %v", err)
	}

	// Linux requires a delay for keyboard initialization
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	// Main card reading loop
	return s.cardReadingLoop(ctx, selectedReaders, kb)
}

func (s *service) Flags() Flags {
	return s.flags
}


func (s *service) formatOutput(rx []byte) string {
	var output string
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
			s.flags.Decimal = false
		} else {
			output = fmt.Sprintf("%d", number)
		}
	}
	
	if !s.flags.Decimal {
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
	for {
		fmt.Println("Waiting for a Card...")
		
		// Wait for card present with error handling
		index, err := s.waitForCardWithRetry(ctx, selectedReaders)
		if err != nil {
			s.notificationManager.NotifyErrorThrottled("card-error", "Karte konnte nicht erkannt werden. Bitte NFC-Lesegerät überprüfen.")
			if s.config.Advanced.AutoReconnect {
				continue
			}
			return err
		}

		// Process the card
		if err := s.processCard(ctx, selectedReaders, index, kb); err != nil {
			s.notificationManager.NotifyErrorThrottled("card-error", "Karte konnte nicht gelesen werden. Bitte erneut versuchen.")
			fmt.Printf("Card processing failed: %v\n", err)
			// Continue to next card instead of exiting
			continue
		}
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
	fmt.Println("Connecting to card...")
	
	// Connect to card with retry
	var card *scard.Card
	err := s.retryManager.Retry(func() error {
		var err error
		card, err = ctx.Connect(selectedReaders[index], scard.ShareShared, scard.ProtocolAny)
		if err != nil {
			// Track reader connection failure
			if s.restartManager.TrackSystemFailure("Reader Connection", err) {
				// Restart was triggered, this will never return
				return nil
			}
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to connect to card: %v", err)
	}
	defer card.Disconnect(scard.ResetCard)

	// Read UID with retry
	uidBytes, err := s.readCardUID(card)
	if err != nil {
		return err
	}

	fmt.Printf("UID is: % x\n", uidBytes)

	// Format and send keyboard output
	output := s.formatOutput(uidBytes)
	fmt.Print("Writing as keyboard input...")
	
	if err := KeyboardWrite(output, kb); err != nil {
		s.notificationManager.NotifyErrorThrottled("keyboard-error", "Karten-ID konnte nicht eingegeben werden. Cursor im richtigen Feld?")
		return fmt.Errorf("failed to write keyboard output: %v", err)
	}

	fmt.Println("Success!")
	s.notificationManager.NotifySuccess(fmt.Sprintf("Card UID: %s", output))

	// Wait for card removal
	fmt.Print("Waiting for card release...")
	err = s.waitUntilCardRelease(ctx, selectedReaders, index)
	if err != nil {
		s.notificationManager.NotifyError("Fehler beim Warten auf Karten-Entfernung. Karte wurde trotzdem gelesen.")
	} else {
		fmt.Println("Card released")
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
