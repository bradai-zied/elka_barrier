package elka

import (
	"encoding/binary"
	"fmt"

	"github.com/rs/zerolog/log"
	// "github.com/rs/zerolog/log"
)

func InterpretBarrierStatus(status []byte) string {
	fmt.Printf("len:%v, %x \n", len(status), status[2])
	if len(status) < 8 {
		fmt.Printf("%x", status[2])
		return "Wrong message interprter Status"
	}

	result := "Barrier Status:\n"
	result += fmt.Sprintf("  Loop A: %v\n", status[2]&(1<<3) != 0)
	result += fmt.Sprintf("  Loop B: %v\n", status[2]&(1<<4) != 0)
	result += fmt.Sprintf("  Loop C: %v\n", status[2]&(1<<5) != 0)
	result += fmt.Sprintf("  BTZ1A: %v\n", status[0]&(1<<5) != 0)
	result += fmt.Sprintf("  BTA1: %v\n", status[0]&(1<<2) != 0)

	return result
}

// this function detec the barrier when it start moving
func InterpretBarrierState(state byte) string {
	states := []string{
		"Opening",
		"Closing",
		"Pre-warning before opening",
		"Pre-warning before closing",
		"Open",
		"Closed",
		"Intermediate position",
	}

	if int(state) < len(states) {
		return states[state]
	}
	return fmt.Sprintf("%d", state)
}

// Helper functions for interpretation (you'll need to implement these):

func InterpretBarrierStatusMask(mask []byte) string {
	if len(mask) < 8 {
		return "Error: Insufficient data for barrier status mask"
	}

	result := "Barrier Status Mask:\n"

	// Define status bits
	statusBits := []struct {
		byte     int
		bit      uint8
		name     string
		register string
	}{
		{0, 0, "Funk", "Status0"},
		{0, 1, "BT", "Status0"},
		{0, 2, "BTA1", "Status0"},
		{0, 3, "BTA2", "Status0"},
		{0, 4, "BTA3", "Status0"},
		{0, 5, "BTZ1A", "Status0"},
		{0, 6, "BTZ1B", "Status0"},
		{0, 7, "BTZ2", "Status0"},

		{1, 0, "BTS", "Status1"},
		{1, 1, "Baum Ab", "Status1"},
		{1, 2, "LS", "Status1"},
		{1, 3, "SLZ", "Status1"},
		{1, 4, "F_ereignis_md", "Status1"},
		{1, 5, "SERVICE", "Status1"},
		{1, 6, "Netzüberwachung", "Status1"},

		{2, 0, "Abgleich Schleife A", "Status2"},
		{2, 1, "Abgleich Schleife B", "Status2"},
		{2, 2, "Abgleich Schleife C", "Status2"},
		{2, 3, "Schleife A", "Status2"},
		{2, 4, "Schleife B", "Status2"},
		{2, 5, "Schleife C", "Status2"},
		{2, 6, "Schleife EXT", "Status2"},

		// Status3 to Status7 are for Multi relays 1-40
	}

	for _, status := range statusBits {
		if mask[status.byte]&(1<<status.bit) != 0 {
			result += fmt.Sprintf("  %s (%s, Bit %d): %sMonitored%s\n", status.name, status.register, status.bit, Yellow, Reset)
		} else {
			result += fmt.Sprintf("  %s (%s, Bit %d): %sNot monitored%s\n", status.name, status.register, status.bit, Green, Reset)
		}
	}

	// Handle Multi relays (Status3 to Status7)
	for i := 3; i < 8; i++ {
		for bit := 0; bit < 8; bit++ {
			relayNum := (i-3)*8 + bit + 1
			if mask[i]&(1<<bit) != 0 {
				result += fmt.Sprintf("  Multi %d (Status%d, Bit %d): %sMonitored%s\n", relayNum, i, bit, Yellow, Reset)
			} else {
				result += fmt.Sprintf("  Multi %d (Status%d, Bit %d): %sNot monitored%s\n", relayNum, i, bit, Green, Reset)
			}
		}
	}
	return result
}

func DecodeTeleChangeMonitoring(message []byte) map[string]bool {
	log.Warn().Msgf("Decode:%s", BytesToHex(message))
	// result := "5.1.9 tele_change monitoring:\n"
	result := ""
	// Define the monitored messages
	monitoredMessages := []string{
		"tele_barrier status",
		"tele_servicecounter",
		"tele_maintenancecounter",
		"tele_gatestate",
		"tele_error_memory",
		"tele_vehicle counter",
		"tele_barrierposition",
		"tele_motorstatus",
		"tele_debug_loops",
	}

	// Extract the flags from the message
	flags := binary.LittleEndian.Uint16(message[3:5])

	// Create a map to store the status of each message
	status := make(map[string]bool)

	// Check each bit and set the status
	for i, msg := range monitoredMessages {
		status[msg] = (flags & (1 << uint(i))) != 0

		if status[msg] {
			result += fmt.Sprintf("  - %s: %sMonitored%s\n", msg, Yellow, Reset)
		} else {
			result += fmt.Sprintf("  - %s : %sNot monitored%s\n", msg, Green, Reset)
		}
	}
	fmt.Print(result)
	return status
}

func InterpretChangeMonitoring(flags uint16) string {
	log.Warn().Msgf("Monit Flags: %x", flags)
	result := "Change Monitoring Flags:\n"

	monitoringFlags := []struct {
		bit  uint16
		name string
	}{
		{0, "Barrier Status"},
		{1, "Service Counter"},
		{2, "Maintenance Counter"},
		{3, "Barrier State"},
		{4, "Error Memory"},
		{5, "Vehicle Counter"},
		{6, "Barrier Position"},
		{7, "Motor Status"},
		{8, "Loop Debug"},
	}

	for _, flag := range monitoringFlags {
		if flags&(1<<flag.bit) != 0 {
			result += fmt.Sprintf("  - %s: %sMonitored%s\n", flag.name, Yellow, Reset)
		} else {
			result += fmt.Sprintf("  -%s : %sNot monitored%s\n", flag.name, Green, Reset)
		}
	}

	// Special notes for certain flags
	if flags&(1<<7) != 0 {
		result += "    Note: Motor Status monitoring should only be used for troubleshooting.\n"
		result += "    It generates a high data volume and should not be enabled in production.\n"
	}
	if flags&(1<<8) != 0 {
		result += "    Note: Loop Debug monitoring should only be used for troubleshooting.\n"
		result += "    It generates a high data volume and should not be enabled in production.\n"
	}

	return result
}

func InterpretRadioCode(code uint32) string {
	// Implement interpretation of radio code
	return fmt.Sprintf("Radio Code: %08X\n", code)
}

func InterpretCountFunction(lower, upper, current byte) string {
	return fmt.Sprintf("Count Function:\n  Lower Limit: %d\n  Upper Limit: %d\n  Current Value: %d\n",
		int8(lower), upper, int8(current))
}

func InterpretInductionLoops(data []byte) string {
	// Implement interpretation of induction loop settings
	return "Induction Loops: [Implement interpretation]\n"
}

func InterpretDirectionLogics(data []byte) string {
	// Implement interpretation of direction logics
	return "Direction Logics: [Implement interpretation]\n"
}

func InterpretErrorMemory(data []byte) string {
	// Implement interpretation of error memory
	return "Error Memory: [Implement interpretation]\n"
}

func InterpretConfigFlags(data []byte) string {
	// Implement interpretation of configuration flags
	return "Configuration Flags: [Implement interpretation]\n"
}

func InterpretMultiRelayModes(data []byte) string {
	// Implement interpretation of multi-relay modes
	return "Multi-relay Modes: [Implement interpretation]\n"
}

func InterpretInductionLoopPeriods(data []byte) string {
	// Implement interpretation of induction loop periods
	return "Induction Loop Periods: [Implement interpretation]\n"
}

func InterpretLoopCalibrationCounters(data []byte) string {
	return fmt.Sprintf("Loop Calibration Counters:\n  Loop A: %d\n  Loop B: %d\n  Loop C: %d\n",
		binary.BigEndian.Uint16(data[0:2]), binary.BigEndian.Uint16(data[2:4]), binary.BigEndian.Uint16(data[4:6]))
}

func InterpretBarrierParameters(data []byte) string {
	if len(data) < 86 { // 43 words * 2 bytes per word
		return "Error: Insufficient data for barrier parameters"
	}

	result := "Barrier Parameters:\n"

	// Helper function to read 16-bit values
	readUint16 := func(offset int) uint16 {
		return binary.BigEndian.Uint16(data[offset : offset+2])
	}

	// Helper function to read 8-bit values
	readUint8 := func(offset int) uint8 {
		return data[offset]
	}

	// Interpret opening angles (wa1 to wa8)
	for i := 0; i < 8; i++ {
		result += fmt.Sprintf("  wa%d: %d\n", i+1, readUint16(i*2))
	}

	// Interpret soft stop angle
	result += fmt.Sprintf("  ws: %d\n", readUint16(16))

	// Interpret closing angles (wz1 to wz7)
	for i := 0; i < 7; i++ {
		result += fmt.Sprintf("  wz%d: %d\n", i+1, readUint16(18+i*2))
	}

	// Interpret opening currents (ia0 to ia8)
	for i := 0; i < 9; i++ {
		result += fmt.Sprintf("  ia%d: %d\n", i, readUint8(32+i))
	}

	// Interpret closing currents (iz0 to iz8)
	for i := 0; i < 9; i++ {
		result += fmt.Sprintf("  iz%d: %d\n", i, readUint8(41+i))
	}

	// Interpret opening times (da0 to da8)
	for i := 0; i < 9; i++ {
		result += fmt.Sprintf("  da%d: %d\n", i, readUint8(50+i))
	}

	// Interpret closing times (dz0 to dz8)
	for i := 0; i < 9; i++ {
		result += fmt.Sprintf("  dz%d: %d\n", i, readUint8(59+i))
	}

	// Interpret sync parameters
	result += fmt.Sprintf("  iasync: %d\n", readUint8(68))
	result += fmt.Sprintf("  izsync: %d\n", readUint8(69))
	result += fmt.Sprintf("  dasync: %d\n", readUint8(70))
	result += fmt.Sprintf("  dzsync: %d\n", readUint8(71))

	// Interpret controller parameters
	result += fmt.Sprintf("  Ki_star_normal: %d\n", readUint8(72))
	result += fmt.Sprintf("  Kp_tilde_normal: %d\n", readUint8(73))
	result += fmt.Sprintf("  Ki_star_start: %d\n", readUint8(74))
	result += fmt.Sprintf("  Kp_tilde_start: %d\n", readUint8(75))
	result += fmt.Sprintf("  Ki_star_softstop: %d\n", readUint8(76))
	result += fmt.Sprintf("  Kp_tilde_softstop: %d\n", readUint8(77))
	result += fmt.Sprintf("  Ki_star_quickstop: %d\n", readUint8(78))
	result += fmt.Sprintf("  Kp_tilde_quickstop: %d\n", readUint8(79))

	// Interpret timing parameters
	result += fmt.Sprintf("  t_mdu: %d\n", readUint8(80))
	result += fmt.Sprintf("  t_softstop: %d\n", readUint8(81))
	result += fmt.Sprintf("  t_quickstop: %d\n", readUint8(82))
	result += fmt.Sprintf("  t_startup: %d\n", readUint8(83))
	result += fmt.Sprintf("  t_startup_cp: %d\n", readUint8(84))

	// Interpret flags
	flags := readUint8(85)
	result += "  Flags:\n"
	result += fmt.Sprintf("    Direction inverted: %v\n", flags&0x01 != 0)
	result += fmt.Sprintf("    Holding force level: %d\n", (flags>>1)&0x03)
	result += fmt.Sprintf("    Power failure detection speed level: %d\n", (flags>>3)&0x03)

	return result
}

func InterpretBarrierType(series, speed byte) string {
	return fmt.Sprintf("Barrier Type:\n  Series: %d\n  Speed: %d\n", series, speed)
}
