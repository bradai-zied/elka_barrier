package elka

import (
	"errors"
	"fmt"
	g "go_barrier/globals"
	"time"

	"github.com/rs/zerolog/log"
)

func (c *Controller) SetChangeNotifications(flags uint16) error {
	// c.mu.Lock()
	// defer c.mu.Unlock()

	if !c.IsConnected {
		return errors.New("Barrier is  not connected")
	}

	// Construct the telegram
	data := []byte{0x04, 0x04, byte(flags), byte(flags >> 8)}
	telegram := append([]byte{0x55, byte(len(data))}, data...)
	checksum := CalculateChecksum(telegram)
	telegram = append(telegram, byte(checksum>>8), byte(checksum))
	// telegram = append(telegram, byte(checksum), byte(checksum>>8))

	// fmt.Print(LightBlue)
	// fmt.Printf("Send set msg: %v (hex): %s", telegram, BytesToHex(telegram))
	// fmt.Println(Reset)
	// Send the telegram
	// if len(c.MessageToApi) > 0 {
	// 	log.Warn().Str("BarrierIP", c.Barrierip).Msg("Old message not processed clear it before send")
	select {
	case <-c.MessageToApi:
		log.Warn().Str("BarrierIP", c.Barrierip).Int("ChannelLength", len(c.MessageToApi)).Msg("Old message not processed clear it before send")
	default:
		// Channel is empty now
	}
	// }
	log.Debug().Str("BarrierIP", c.Barrierip).Msgf("Send Message: %s", BytesToHex(telegram))
	_, err := c.conn.Write(telegram)
	if err != nil {
		return fmt.Errorf("failed to send change notification settings: %v", err)
	}

	// Wait for response
	// response := make([]byte, 512)
	select {
	case response := <-c.MessageToApi:
		// fmt.Printf("%sResponse set notification:%s%s\n", Orange, BytesToHex(response), Reset)
		if response == "ack" || response == "syn" {
			log.Debug().Str("BarrierIP", c.Barrierip).Str("ConfigurationResponse", response).Msg("Setting notification accepted")
			return nil
		} else {
			log.Error().Str("BarrierIP", c.Barrierip).Str("ConfigurationResponse", response).Msg("Setting notification refused")
			return fmt.Errorf("configuration refused:  %v", response)
		}
	case <-time.After(time.Duration(g.Config.TimeOutHttpResp) * time.Second):
		// response = []byte{}
		log.Error().Str("BarrierIP", c.Barrierip).Msg("Setting notification refused")
		return fmt.Errorf("timeout waiting for query response %v", flags)
	}
	// n, err := c.conn.Read(response)
	// if err != nil {
	// 	return fmt.Errorf("failed to read response: %v", err)
	// }
	// fmt.Printf("%v", response)
	// Process response
	// if n < 1 {
	// 	return errors.New("invalid response")
	// }
	// return nil
}

func (c *Controller) sendInitialData() {
	if c.changeNotificationFlags&FlagBarrierStatus != 0 {
		status, err := c.GetBarrierStatus()
		if err == nil {
			c.sendTelegram(0x07, status) // tele_barrier status
		}
	}
	if c.changeNotificationFlags&FlagServiceCounter != 0 {
		counter, err := c.GetServiceCounter()
		if err == nil {
			c.sendTelegram(0x0A, counter) // tele_servicecounter
		}
	}
	if c.changeNotificationFlags&FlagMaintenanceCounter != 0 {
		counter, err := c.GetMaintenanceCounter()
		if err == nil {
			c.sendTelegram(0x0B, counter) // tele_maintenancecounter
		}
	}
	if c.changeNotificationFlags&FlagGateState != 0 {
		state, err := c.GetGateState()
		if err == nil {
			c.sendTelegram(0x0C, []byte{state}) // tele_gatestate
		}
	}
	if c.changeNotificationFlags&FlagErrorMemory != 0 {
		errorMem, err := c.GetErrorMemory(0) // Index 0 for the most recent error
		if err == nil {
			c.sendTelegram(0x17, errorMem) // tele_error_memory
		}
	}
	if c.changeNotificationFlags&FlagVehicleCounter != 0 {
		counter, err := c.GetVehicleCounter()
		if err == nil {
			c.sendTelegram(0x1C, counter) // tele_vehicle counter
		}
	}
	if c.changeNotificationFlags&FlagBarrierPosition != 0 {
		position, err := c.GetBarrierPosition()
		if err == nil {
			c.sendTelegram(0x1D, []byte{position}) // tele_barrierposition
		}
	}
	if c.changeNotificationFlags&FlagMotorStatus != 0 {
		status, err := c.GetMotorStatus()
		if err == nil {
			c.sendTelegram(0x38, status) // tele_motorstatus
		}
	}
	if c.changeNotificationFlags&FlagDebugLoops != 0 {
		debug, err := c.GetDebugLoops()
		if err == nil {
			c.sendTelegram(0x39, debug) // tele_debug_loops
		}
	}
}

func (c *Controller) sendTelegram(telegramType byte, data []byte) {
	telegram := []byte{0x55, byte(len(data) + 1), telegramType}
	telegram = append(telegram, data...)
	checksum := CalculateChecksum(telegram)
	telegram = append(telegram, byte(checksum), byte(checksum>>8))

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsConnected {
		_, err := c.conn.Write(telegram)
		if err != nil {
			// Handle error (log it, etc.)
			fmt.Printf("Failed to send telegram: %v\n", err)
		}
	}
}

// This function would be called whenever a parameter changes
func (c *Controller) HandleParameterChange(flag uint16) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.IsConnected || c.changeNotificationFlags&flag == 0 {
		return // Not connected or this change is not being monitored
	}

	var telegramType byte
	var data []byte
	var err error

	switch flag {
	case FlagBarrierStatus:
		telegramType = 0x07 // tele_barrier status
		data, err = c.GetBarrierStatus()
	case FlagServiceCounter:
		telegramType = 0x0A // tele_servicecounter
		data, err = c.GetServiceCounter()
	case FlagMaintenanceCounter:
		telegramType = 0x0B // tele_maintenancecounter
		data, err = c.GetMaintenanceCounter()
	case FlagGateState:
		telegramType = 0x0C // tele_gatestate
		var state byte
		state, err = c.GetGateState()
		data = []byte{state}
	case FlagErrorMemory:
		telegramType = 0x17             // tele_error_memory
		data, err = c.GetErrorMemory(0) // Always send the most recent error (index 0)
	case FlagVehicleCounter:
		telegramType = 0x1C // tele_vehicle counter
		data, err = c.GetVehicleCounter()
	case FlagBarrierPosition:
		telegramType = 0x1D // tele_barrierposition
		var position byte
		position, err = c.GetBarrierPosition()
		data = []byte{position}
	case FlagMotorStatus:
		telegramType = 0x38 // tele_motorstatus
		data, err = c.GetMotorStatus()
	case FlagDebugLoops:
		telegramType = 0x39 // tele_debug_loops
		data, err = c.GetDebugLoops()
	default:
		err = fmt.Errorf("unknown parameter change flag: %d", flag)
	}

	if err != nil {
		// Log the error or handle it as appropriate for your application
		fmt.Printf("Error getting data for parameter change notification: %v\n", err)
		return
	}

	// Construct and send the telegram
	telegram := []byte{0x55, byte(len(data) + 1), telegramType}
	telegram = append(telegram, data...)
	checksum := CalculateChecksum(telegram)
	telegram = append(telegram, byte(checksum), byte(checksum>>8))

	_, err = c.conn.Write(telegram)
	if err != nil {
		// Log the error or handle it as appropriate for your application
		fmt.Printf("Failed to send parameter change notification: %v\n", err)
	}
}

func ParseNotificationFlags(args []string) uint16 {
	var flags uint16
	flagMap := map[string]uint16{
		"status":      FlagBarrierStatus,
		"service":     FlagServiceCounter,
		"maintenance": FlagMaintenanceCounter,
		"gate":        FlagGateState,
		"error":       FlagErrorMemory,
		"vehicle":     FlagVehicleCounter,
		"position":    FlagBarrierPosition,
		"motor":       FlagMotorStatus,
		"debug":       FlagDebugLoops,
	}

	for _, arg := range args {
		if flag, ok := flagMap[arg]; ok {
			flags |= flag
		} else {
			fmt.Printf("Warning: Unknown notification type '%s'\n", arg)
		}
	}

	return flags
}
