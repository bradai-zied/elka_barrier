package elka

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

func Handle_0C_Message(response []byte, c *Controller) {
	BarrierPosition := InterpretBarrierState(response[3])
	c.BarrierPositionStr = BarrierPosition
	if response[3] == 0x05 || response[3] == 0x01 {
		c.IsClosed = true
		log.Debug().Str("ip", c.Barrierip).Msg("Set IsClosed to true")
	}
	if response[3] == 0x04 || response[3] == 0x00 {
		c.IsClosed = false
		log.Debug().Str("ip", c.Barrierip).Msg("Set IsClosed to false")
	}
}
func HandleMessage(msg []byte, c *Controller) {
	responsetype := msg[2]

	switch responsetype {
	case 0x01:

		SendResponseToApi("ack", c)
	case 0x02:
		SendResponseToApi("nak", c)
	case 0x03:
		SendResponseToApi("busy", c)
	case 0x04:
		SendResponseToApi("syn", c)
	case 0x06:
		HandleQueryMessage(0x01, msg, c) //Query the program version		tele_programversion
	case 0x07:
		HandleQueryMessage(0x02, msg, c) //Query the barrier status 		tele_barrier status
	case 0x08:
		HandleQueryMessage(0x03, msg, c) //Query the barrier status mask tele_barrier status mask
	case 0x09:
		HandleQueryMessage(0x04, msg, c) //Change monitoring query 		tele_change monitoring
	case 0x0A:
		HandleQueryMessage(0x05, msg, c) //Query the service counter		tele_servicecounter
	case 0x0B:
		HandleQueryMessage(0x06, msg, c) //Query the maintenance counter		tele_maintenancecounter
	case 0x0C:
		HandleQueryMessage(0x07, msg, c) //Query the gate status		tele_gatestate
	case 0x0D:
		HandleQueryMessage(0x08, msg, c) //Query the hold open time		tele_t_open
	case 0x0E:
		HandleQueryMessage(0x09, msg, c) //Query the advance warning before opening		tele_t_vw_auf
	case 0x0F:
		HandleQueryMessage(0x0A, msg, c) //Query the advance warning before closing		tele_t_vw_zu
	case 0x10:
		HandleQueryMessage(0x0B, msg, c) //Query the radio code BT		tele_dec_bt
	case 0x11:
		HandleQueryMessage(0x0C, msg, c) //Query the counting function 		tele_counting function
	case 0x12:
		HandleQueryMessage(0x0D, msg, c) //Query the induction loops		tele_induction loops
	case 0x13:
		HandleQueryMessage(0x0E, msg, c) //Query the directional logics 		tele_directional logics
	case 0x14:
		HandleQueryMessage(0x0F, msg, c) //Query the serial number		tele_serial_number
	case 0x15:
		HandleQueryMessage(0x10, msg, c) //Query the MAC address		tele_macaddress
	case 0x16:
		HandleQueryMessage(0x11, msg, c) //Query the operating hours counter		tele_operating hours counter
	case 0x17:
		HandleQueryMessage(0x12, msg, c) //Query the error memory		tele_error_memory(index)
	case 0x18:
		HandleQueryMessage(0x13, msg, c) //Query the config flags		tele_configflags
	case 0x19:
		HandleQueryMessage(0x14, msg, c) //Query the multi-relay operating modes		tele_multi
	case 0x1A:
		HandleQueryMessage(0x15, msg, c) //Query the maintenance interval		tele_maintenanceinterval
	case 0x1B:
		HandleQueryMessage(0x16, msg, c) //Query the induction loop periods		tele_induction loop periods
	case 0x1C:
		HandleQueryMessage(0x17, msg, c) //Vehicle counter query		tele_vehicle counter
	case 0x1D:
		HandleQueryMessage(0x18, msg, c) //Query the barrier position		tele_barrierposition
	case 0x1E:
		HandleQueryMessage(0x19, msg, c) //Query the password		tele_password
	case 0x1F:
		HandleQueryMessage(0x1A, msg, c) //Query the loop adjustment counters		tele_schleifenabgleich_counter
	case 0x20:
		HandleQueryMessage(0x1B, msg, c) //Query the barrier parameters		tele_parameter
	case 0x21:
		HandleQueryMessage(0x1C, msg, c) //Query the barrier type		tele_barrier type
	default:
		log.Warn().Str("ip", c.Barrierip).Msg("Unhandled Decription Message" + BytesToHex(msg))
		// 	HandleQueryMessage(0x10, msg, c)
		// case 0x0A:
		// 	HandleQueryMessage(0x05, msg, c)
		// case 0x0B:
		// 	HandleQueryMessage(0x06, msg, c)
		// case 0x0C:
		// 	HandleQueryMessage(0x07, msg, c)
		// case 0x0D:
		// 	HandleQueryMessage(0x08, msg, c)
	}
	// }
	// if msgtype == 0x0C {
	// 	Handle_0C_Message(msg, c)
	// } else if msgtype == 0x08 {
	// 	HandleQueryMessage(0x03, msg, c)
	// 	log.Debug().Str("ip", c.Barrierip).Msg(" message: " + BytesToHex(msg))
	// } else if msgtype == 0x09 {
	// 	HandleQueryMessage(0x04, msg, c)
	// 	log.Debug().Str("ip", c.Barrierip).Msg(" message: " + BytesToHex(msg))
	// } else {
	// 	// Handle_09_Message(msg, c)
	// 	log.Warn().Str("ip", c.Barrierip).Msg("Unhandled message: " + BytesToHex(msg))
	// }

}

func SendResponseToApi(msg string, c *Controller) {
	log.Debug().Str("ip", c.Barrierip).Msgf("Receive Message :%s", msg)
	for {
		select {
		case c.MessageToApi <- msg:
			return
			// Remove an item from the channel
		default:
			for {
				log.Warn().Str("BarrierIP", c.Barrierip).Msgf("MessageApi buffer not empty %d, Empty before sending", len(c.MessageToApi))
				select {
				case xx := <-c.MessageToApi:
					log.Warn().Str("BarrierIP", c.Barrierip).Str("Message on Buffer :", xx)
					// Remove an item from the channel
				default:
					// log.Warn().Str("BarrierIP", c.Barrierip).Str("Message on Buffer :", <-c.MessageToApi)
					// Channel is empty now
					c.MessageToApi <- msg
					return
				}
			}

		}
	}

}

func HandleQueryMessage(queryType byte, response []byte, c *Controller) {
	// result := fmt.Sprintf("Query Type: 0x%02X\n", queryType)
	result := fmt.Sprintf("Full Response (hex): %s\n", BytesToHex(response))

	switch queryType {
	case 0x00:
		result += fmt.Sprintf("Device ID: 0x%04X\n", binary.BigEndian.Uint16(response[3:5]))
	case 0x01:
		result += fmt.Sprintf("Program Version: 0x%04X\n", binary.BigEndian.Uint16(response[3:5]))
	case 0x02:
		// log.Debug().Msg("************************************")
		// Function that bring all high low flags
		// result += InterpretBarrierStatus(response[3:11])
		status := DecodeTeleBarrierStatus(response[2:11])

		SendResponseToApi(MapToString(status), c)

		resultx := fmt.Sprintf("Barrier Status : %v\n", queryType)

		for key, value := range status {
			if value {
				resultx += fmt.Sprintf("  %s : %s%v-High%s\n", key, Yellow, value, Reset)
			} else {
				resultx += fmt.Sprintf("  %s : %s%v-Low%s\n", key, Green, value, Reset)
			}
		}
		c.LoopA = status["LoopA"]
		c.LoopB = status["LoopB"]
		c.IsLockedUp = status["BUSBA"]
		c.IsLockedDown = status["BUSBZ"]
		c.IsClosed = status["Multi6"]
		log.Debug().Str("BarrierIP", c.Barrierip).Str("_Msg", resultx).Send()
		return
		// fmt.Printf(resultx)
	case 0x03:
		result += InterpretBarrierStatusMask(response[3:11])
		SendResponseToApi(result, c)
	case 0x04:
		// result += InterpretChangeMonitoring(binary.BigEndian.Uint16(response[3:5]))
		_ = DecodeTeleChangeMonitoring(response)
		// log.Debug().Msgf("DecodeTeleChangeMonitoring: %v", xresult)
	case 0x05:
		result += fmt.Sprintf("Service Counter: %d\n", binary.BigEndian.Uint32(response[3:7]))
	case 0x06:
		result += fmt.Sprintf("Maintenance Counter: %d\n", binary.BigEndian.Uint32(response[3:7]))
	case 0x07: //this message is for barier position when it move, we ill have message each open close
		barrierstat := InterpretBarrierState(response[3])
		result += barrierstat
		c.BarrierPositionStr = barrierstat
		SendResponseToApi(barrierstat, c)
		// return
	case 0x08:
		result += fmt.Sprintf("Open Hold Time: %d ms\n", binary.BigEndian.Uint16(response[3:5])*10)
	case 0x09:
		result += fmt.Sprintf("Pre-warning Time before Opening: %d ms\n", binary.BigEndian.Uint16(response[3:5])*10)
	case 0x0A:
		result += fmt.Sprintf("Pre-warning Time before Closing: %d ms\n", binary.BigEndian.Uint16(response[3:5])*10)
	case 0x0B:
		result += InterpretRadioCode(binary.BigEndian.Uint32(response[3:7]))
	case 0x0C:
		result += InterpretCountFunction(response[3], response[4], response[5])
	case 0x0D:
		result += InterpretInductionLoops(response[3:11])
	case 0x0E:
		result += InterpretDirectionLogics(response[3:11])
	case 0x0F:
		result += fmt.Sprintf("Serial Number: %d\n", binary.BigEndian.Uint32(response[3:7]))
	case 0x10:
		result += fmt.Sprintf("MAC Address: %s\n", BytesToHex(response[3:9]))
	case 0x11:
		result += fmt.Sprintf("Operating Hours Counter: %d hours\n", binary.BigEndian.Uint32(response[3:7]))
	case 0x12:
		result += InterpretErrorMemory(response[3:])
	case 0x13:
		result += InterpretConfigFlags(response[3:11])
	case 0x14:
		result += InterpretMultiRelayModes(response[3:17])
	case 0x15:
		result += fmt.Sprintf("Maintenance Interval: %d\n", binary.BigEndian.Uint32(response[3:7]))
	case 0x16:
		result += InterpretInductionLoopPeriods(response[3:27])
	case 0x17:
		result += fmt.Sprintf("Vehicle Counter: %d\n", int32(binary.BigEndian.Uint32(response[3:7])))
	case 0x18:
		result += fmt.Sprintf("Barrier Position: %d%%\n", int8(response[3]))
	case 0x19:
		result += fmt.Sprintf("Password: %d\n", binary.BigEndian.Uint16(response[3:5]))
	case 0x1A:
		result += InterpretLoopCalibrationCounters(response[3:9])
	case 0x1B:
		result += InterpretBarrierParameters(response[3:])
	case 0x1C:
		result += InterpretBarrierType(response[3], response[4])
	default:
		result += "Interpretation not implemented for this query type\n"
	}
	// SendResponseToApi(result, c)
	log.Debug().Msgf("queryType: %x  Message %s", queryType, result)
}

func MapToString(m map[string]bool) string {
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for k, v := range m {
		if !first {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("'%s': %v", k, v))
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}
