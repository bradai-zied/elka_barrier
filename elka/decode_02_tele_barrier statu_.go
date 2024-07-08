package elka

func DecodeTeleBarrierStatus(message []byte) map[string]bool {
	if len(message) != 9 { // 1 byte for 0x07 + 8 bytes for Status0-7
		panic("Invalid message length for tele_barrier status")
	}

	status := make(map[string]bool)
	statusBytes := message[1:] // Skip the first byte (0x07)

	// Status0
	status["Wireless"] = statusBytes[0]&0x01 != 0
	status["BT"] = statusBytes[0]&0x02 != 0
	status["BTA1"] = statusBytes[0]&0x04 != 0
	status["BTA2"] = statusBytes[0]&0x08 != 0
	status["BTA3"] = statusBytes[0]&0x10 != 0
	status["BTZ1A"] = statusBytes[0]&0x20 != 0
	status["BTZ1B"] = statusBytes[0]&0x40 != 0
	status["BTZ2"] = statusBytes[0]&0x80 != 0

	// Status1
	status["BTS"] = statusBytes[1]&0x01 != 0
	status["TreeOff"] = statusBytes[1]&0x02 != 0
	status["LS"] = statusBytes[1]&0x04 != 0
	status["SLZ"] = statusBytes[1]&0x08 != 0
	status["FEventMd"] = statusBytes[1]&0x10 != 0
	status["SERVICE"] = statusBytes[1]&0x20 != 0
	status["NetworkMonitoring"] = statusBytes[1]&0x40 != 0

	// Status2
	status["AdjustmentLoopA"] = statusBytes[2]&0x01 != 0
	status["AdjustmentLoopB"] = statusBytes[2]&0x02 != 0
	status["AdjustmentLoopC"] = statusBytes[2]&0x04 != 0
	status["LoopA"] = statusBytes[2]&0x08 != 0
	status["LoopB"] = statusBytes[2]&0x10 != 0
	status["LoopC"] = statusBytes[2]&0x20 != 0
	status["LoopEXT"] = statusBytes[2]&0x40 != 0

	// Status3
	status["Multi1"] = statusBytes[3]&0x01 != 0
	status["Multi2"] = statusBytes[3]&0x02 != 0
	status["Multi3"] = statusBytes[3]&0x04 != 0
	status["Multi4"] = statusBytes[3]&0x08 != 0
	status["Multi5"] = statusBytes[3]&0x10 != 0
	status["Multi6"] = statusBytes[3]&0x20 != 0
	status["Multi7"] = statusBytes[3]&0x40 != 0
	status["Multi8"] = statusBytes[3]&0x80 != 0

	// Status4
	status["Multi9"] = statusBytes[4]&0x01 != 0
	status["Multi10"] = statusBytes[4]&0x02 != 0
	status["Multi11"] = statusBytes[4]&0x04 != 0
	status["Multi12"] = statusBytes[4]&0x08 != 0
	status["Multi13"] = statusBytes[4]&0x10 != 0
	status["Multi14"] = statusBytes[4]&0x20 != 0

	// Status5
	status["BUSBT"] = statusBytes[5]&0x01 != 0
	status["BUSBA"] = statusBytes[5]&0x02 != 0
	status["BUSBZ"] = statusBytes[5]&0x04 != 0
	status["BUSBS"] = statusBytes[5]&0x08 != 0

	// Status6
	status["BUSMulti1"] = statusBytes[6]&0x01 != 0
	status["BUSMulti2"] = statusBytes[6]&0x02 != 0
	status["BUSMulti3"] = statusBytes[6]&0x04 != 0
	status["BUSMulti4"] = statusBytes[6]&0x08 != 0
	status["BUSMulti5"] = statusBytes[6]&0x10 != 0
	status["BUSMulti6"] = statusBytes[6]&0x20 != 0
	status["BUSMulti7"] = statusBytes[6]&0x40 != 0
	status["BUSMulti8"] = statusBytes[6]&0x80 != 0

	// Status7
	status["BUSMulti9"] = statusBytes[7]&0x01 != 0
	status["BUSMulti10"] = statusBytes[7]&0x02 != 0
	status["BUSMulti11"] = statusBytes[7]&0x04 != 0
	status["BUSMulti12"] = statusBytes[7]&0x08 != 0
	status["BUSMulti13"] = statusBytes[7]&0x10 != 0
	status["BUSMulti14"] = statusBytes[7]&0x20 != 0

	return status
}
