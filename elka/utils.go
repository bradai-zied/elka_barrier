package elka

import (
	"fmt"
	g "go_barrier/globals"
	"strings"
	"time"
)

var (
	// address     = "192.168.25.206" // Replace with your controller's IP and port
	maxRetries = g.Config.MaxRetries
	// retryDelay             = time.Duration(g.Config.TimeBetweenRetrySec) * time.Second
	// readTimeout            = time.Duration(g.Config.TimeOutHttpResp) * time.Second
	ConnectTimeOutAfterMax = time.Duration(g.Config.TimeRetryAfterMaxSec) * time.Second
)

const (
	FlagBarrierStatus      uint16 = 1 << 0
	FlagServiceCounter     uint16 = 1 << 1
	FlagMaintenanceCounter uint16 = 1 << 2
	FlagGateState          uint16 = 1 << 3
	FlagErrorMemory        uint16 = 1 << 4
	FlagVehicleCounter     uint16 = 1 << 5
	FlagBarrierPosition    uint16 = 1 << 6
	FlagMotorStatus        uint16 = 1 << 7
	FlagDebugLoops         uint16 = 1 << 8
)

// Define ANSI color codes as constants
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	// Extended colors
	Black        = "\033[30m"
	LightRed     = "\033[91m"
	LightGreen   = "\033[92m"
	LightYellow  = "\033[93m"
	LightBlue    = "\033[94m"
	LightPurple  = "\033[95m"
	LightCyan    = "\033[96m"
	LightGray    = "\033[37m"
	DarkGray     = "\033[90m"
	BrightRed    = "\033[38;5;196m"
	BrightGreen  = "\033[38;5;46m"
	BrightYellow = "\033[38;5;226m"
	BrightBlue   = "\033[38;5;21m"
	BrightPurple = "\033[38;5;201m"
	BrightCyan   = "\033[38;5;51m"
	BrightWhite  = "\033[97m"
	Orange       = "\033[38;5;208m"
	Pink         = "\033[38;5;218m"
	LightBrown   = "\033[38;5;94m"
)

func BytesToHex(bytes []byte) string {
	hex := make([]string, len(bytes))
	for i, b := range bytes {
		hex[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(hex, " ")
}

func CalculateChecksum(data []byte) uint16 {
	// Implement CRC-CCITT (0xFFFF) checksum calculation here
	// This is a placeholder implementation
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
