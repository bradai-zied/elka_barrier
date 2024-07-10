package elka

import (
	"encoding/binary"
	"fmt"
	g "go_barrier/globals"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var ElkaController map[int]*Controller
var IPElkaController map[string]*Controller

// var address = "192.168.25.206" // Replace with your controller's IP address
// var Response chan []byte

type Controller struct {
	conn        net.Conn
	mu          sync.Mutex
	IsConnected bool
	// stopChan                chan struct{} `json:"-"`
	messages                chan []byte `json:"-"`
	changeNotificationFlags uint16

	BarrierStatus      [8]byte     `json:"-"`
	serviceCounter     uint32      `json:"-"`
	maintenanceCounter uint32      `json:"-"`
	gateState          byte        `json:"-"`
	errorMemory        [10][8]byte // 10 levels of error memory
	vehicleCounter     int32       `json:"-"`
	BarrierPosition    int8        `json:"-"`
	motorStatus        []byte      `json:"-"`
	debugLoops         []byte      `json:"-"`
	Barrierip          string
	BarrierPositionStr string
	IsClosed           bool
	IsLockedUp         bool
	IsLockedDown       bool
	LoopA              bool
	LoopB              bool
	MessageToApi       chan string `json:"-"`
	Id                 int
}

func NewController(barrierip string) *Controller {
	return &Controller{
		Barrierip:          barrierip,
		IsLockedUp:         false,
		IsLockedDown:       false,
		LoopA:              false,
		LoopB:              false,
		BarrierPositionStr: "Uknown",
		Id:                 0,
		// stopChan:           make(chan struct{}),
		messages:     make(chan []byte, 200), // Buffer for 100 messages
		MessageToApi: make(chan string, 1),
	}
}

func (c *Controller) Connect() error {
	log.Debug().Str("ip", c.Barrierip).Msg("Try to connect to Barrier  ")
	var err error
	for i := 0; i < g.Config.MaxRetries; i++ {
		log.Debug().Str("ip", c.Barrierip).Msgf("Try to connect to Barrier attempt %d ", i)
		c.conn, err = net.Dial("tcp", c.Barrierip+":52719")
		if err == nil {
			c.IsConnected = true
			log.Debug().Str("ip", c.Barrierip).Msg("******************* Connected ")
			go c.readLoop()
			time.Sleep(1 * time.Second)
			// Assert the connection as a TCP connection to access TCP-specific methods
			tcpConn, ok := c.conn.(*net.TCPConn)
			if !ok {
				fmt.Println("Failed to assert net.Conn as *net.TCPConn")
				// return
			}
			// Enable TCP keep-alive
			if err := tcpConn.SetKeepAlive(true); err != nil {
				fmt.Println("Error setting keep-alive:", err)
				// return
			}
			// Set the keep-alive period to 3 minutes
			if err := tcpConn.SetKeepAlivePeriod(10 * time.Second); err != nil {
				fmt.Println("Error setting keep-alive period:", err)
				// return
			}
			return nil
		}
		//
		log.Err(err).Str("Barrier ip", c.Barrierip).Msgf("Connection attempt %d failed: . Retrying... in %d \n", i+1, g.Config.TimeBetweenRetrySec)
		time.Sleep(time.Duration(g.Config.TimeBetweenRetrySec) * time.Second)
	}
	log.Err(err).Str("Barrier ip", c.Barrierip).Msgf("failed to connect after %d attempts next Retry in %d min", g.Config.MaxRetries, g.Config.TimeRetryAfterMaxSec)
	time.Sleep(time.Duration(g.Config.TimeRetryAfterMaxSec) * time.Minute)
	go c.Connect()

	return fmt.Errorf("failed to connect after %d attempts", g.Config.TimeBetweenRetrySec)

}

func (c *Controller) Disconnect() {
	c.IsConnected = false
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()

	}
	// close(c.stopChan)
}

// func (c *Controller) Reconnect() error {
// 	c.mu.Lock()
// 	defer c.mu.Unlock()

// 	if c.conn != nil {
// 		c.conn.Close()
// 	}
// 	c.IsConnected = false
// 	close(c.stopChan)
// 	c.stopChan = make(chan struct{})
// 	c.messages = make(chan []byte, 200)

// 	var err error
// 	for i := 0; i < maxRetries; i++ {
// 		// c.conn, err = net.Dial("tcp", c.Barrierip+":52719")
// 		err = c.Connect()
// 		if err == nil {
// 			c.IsConnected = true
// 			go c.readLoop()
// 			return nil
// 		}

// 		log.Warn().Str("Barrier ip", c.Barrierip).Msgf("Connection attempt %d failed: %v. Retrying... \n", i+1, err)
// 		time.Sleep(time.Duration(g.Config.TimeBetweenRetrySec))
// 	}
// 	log.Warn().Str("Barrier ip", c.Barrierip).Msgf("failed to reconnect after %d attempts", maxRetries)
// 	return fmt.Errorf("failed to reconnect after %d attempts", maxRetries)
// }

func (c *Controller) readLoop() {
	// c.mu.Lock()
	time.Sleep(1000 * time.Millisecond)

	// defer c.Disconnect()
	buffer := make([]byte, 1024)
	closeloop := true
	log.Debug().Str("BarrierIP", c.Barrierip).Bool("closeloop", closeloop).Msg("********** readLoop started")
	// select {
	// case <-c.stopChan:
	// default:
	// 	//chanle epmpty

	// }

	for closeloop {
		// select {
		// case <-c.stopChan:
		// 	log.Error().Str("BarrierIP", c.Barrierip).Bool("closeloop", closeloop).Msg("********** readLoop stopped")
		// 	return
		// default:
		n, err := c.conn.Read(buffer)
		if err != nil {
			// c.Disconnect()
			log.Err(err).Str("BarrierIP", c.Barrierip).Bool("IsConnected", c.IsConnected).Msg(" Error Read read from TCP connection kill in 1 sec")
			c.Disconnect()
			go c.Connect()
			closeloop = false
			// if c.IsConnected {
			// 	log.Error().Str("BarrierIP", c.Barrierip).Msg("Retry read loop in 1 sec ")
			// 	time.Sleep(500 * time.Millisecond)
			// 	go c.readLoop()
			// 	closeloop = false
			// } else {
			// 	// c.Disconnect()
			// 	c.IsConnected = false
			// 	go c.Connect()
			// 	closeloop = false
			// }
			// os.Exit(0)
			// go c.Reconnect() // Trigger reconnection
			// return
			// defer c.mu.Unlock()
		} else {
			c.processMessage(buffer[:n])
		}
		// }
	}

}

// func (c *Controller) triggerReconnect() {
// 	for {
// 		log.Warn().Str("BarrierIP", c.Barrierip).Msg("trigger, Connection lost. Attempting to reconnect...")
// 		err := c.Reconnect()
// 		if err == nil {
// 			c.IsConnected = true
// 			log.Info().Str("BarrierIP", c.Barrierip).Msgf("Reconnected successfully")
// 			return
// 		}
// 		log.Warn().Str("BarrierIP", c.Barrierip).Msgf("Failed to reconnect: %v. Retrying in %v...\n", err, g.Config.TimeBetweenRetrySec)
// 		time.Sleep(time.Duration(g.Config.TimeBetweenRetrySec) * time.Second)
// 	}
// }

func (c *Controller) processMessage(data []byte) {
	// Process the received message
	// This is where you'd implement the logic to handle different message types
	// fmt.Printf("Received response (hex): %s\n", bytesToHex(data))
	// fmt.Printf("X Received message: %v\n", data)
	select {
	case c.messages <- data:
		log.Debug().Str("BarrierIP", c.Barrierip).Msgf("Received Message: %s ", BytesToHex(data))
		HandleMessage(data, c)
	default:
		log.Warn().Msg("MessageApi buffer full, Empty before sending")
		for {
			select {
			case old := <-c.messages:
				log.Warn().Str("BarrierIP", c.Barrierip).Msgf("Message buffer full, discarding message empty channel,Old:%v", old)
				// Remove an item from the channel
			default:
				// Channel is empty now
				c.messages <- data
				return
			}
		}
	}
}

func (c *Controller) SendQueryTelegram(queryType byte) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.IsConnected {
		return "", fmt.Errorf("not connected")
	}
	log.Warn().Msgf("*****************Request queryType : %v", queryType)
	// Construct the query telegram
	data := []byte{0x55, 0x02, 0x02, queryType, 0x00, 0x00}
	checksum := CalculateChecksum(data[:4])
	binary.BigEndian.PutUint16(data[4:], checksum)

	// // Send the query
	log.Debug().Str("BarrierIP", c.Barrierip).Msgf("Send Query: %s", BytesToHex(data))
	_, err := c.conn.Write(data)
	if err != nil {
		log.Err(err).Str("BarrierIP", c.Barrierip).Msg("Failed to send query to TCP connection")
		return "", fmt.Errorf("failed to send query: %v", err)
	}

	// // Read the response
	// response := make([]byte, 1024)
	// select {
	// case response = <-c.messages:
	// 	// fmt.Print("Received Response")
	// case <-time.After(readTimeout):
	// 	response = []byte{}
	// 	return "", fmt.Errorf("timeout waiting for query response %v", queryType)
	// }
	// result := fmt.Sprintf("Query Type: 0x%02X\n", queryType)
	// result += fmt.Sprintf("Full Response (hex): %s\n", BytesToHex(response))
	// // Interpret the response based on the query type
	// switch queryType {
	// case 0x00:
	// 	result += fmt.Sprintf("Device ID: 0x%04X\n", binary.BigEndian.Uint16(response[3:5]))
	// case 0x01:
	// 	result += fmt.Sprintf("Program Version: 0x%04X\n", binary.BigEndian.Uint16(response[3:5]))
	// case 0x02:
	// 	result += InterpretBarrierStatus(response[3:11])
	// case 0x03:
	// 	result += InterpretBarrierStatusMask(response[3:11])
	// case 0x04:
	// 	result += InterpretChangeMonitoring(binary.BigEndian.Uint16(response[3:5]))
	// case 0x05:
	// 	result += fmt.Sprintf("Service Counter: %d\n", binary.BigEndian.Uint32(response[3:7]))
	// case 0x06:
	// 	result += fmt.Sprintf("Maintenance Counter: %d\n", binary.BigEndian.Uint32(response[3:7]))
	// case 0x07:
	// 	result += InterpretBarrierState(response[3])
	// case 0x08:
	// 	result += fmt.Sprintf("Open Hold Time: %d ms\n", binary.BigEndian.Uint16(response[3:5])*10)
	// case 0x09:
	// 	result += fmt.Sprintf("Pre-warning Time before Opening: %d ms\n", binary.BigEndian.Uint16(response[3:5])*10)
	// case 0x0A:
	// 	result += fmt.Sprintf("Pre-warning Time before Closing: %d ms\n", binary.BigEndian.Uint16(response[3:5])*10)
	// case 0x0B:
	// 	result += InterpretRadioCode(binary.BigEndian.Uint32(response[3:7]))
	// case 0x0C:
	// 	result += InterpretCountFunction(response[3], response[4], response[5])
	// case 0x0D:
	// 	result += InterpretInductionLoops(response[3:11])
	// case 0x0E:
	// 	result += InterpretDirectionLogics(response[3:11])
	// case 0x0F:
	// 	result += fmt.Sprintf("Serial Number: %d\n", binary.BigEndian.Uint32(response[3:7]))
	// case 0x10:
	// 	result += fmt.Sprintf("MAC Address: %s\n", BytesToHex(response[3:9]))
	// case 0x11:
	// 	result += fmt.Sprintf("Operating Hours Counter: %d hours\n", binary.BigEndian.Uint32(response[3:7]))
	// case 0x12:
	// 	result += InterpretErrorMemory(response[3:])
	// case 0x13:
	// 	result += InterpretConfigFlags(response[3:11])
	// case 0x14:
	// 	result += InterpretMultiRelayModes(response[3:17])
	// case 0x15:
	// 	result += fmt.Sprintf("Maintenance Interval: %d\n", binary.BigEndian.Uint32(response[3:7]))
	// case 0x16:
	// 	result += InterpretInductionLoopPeriods(response[3:27])
	// case 0x17:
	// 	result += fmt.Sprintf("Vehicle Counter: %d\n", int32(binary.BigEndian.Uint32(response[3:7])))
	// case 0x18:
	// 	result += fmt.Sprintf("Barrier Position: %d%%\n", int8(response[3]))
	// case 0x19:
	// 	result += fmt.Sprintf("Password: %d\n", binary.BigEndian.Uint16(response[3:5]))
	// case 0x1A:
	// 	result += InterpretLoopCalibrationCounters(response[3:9])
	// case 0x1B:
	// 	result += InterpretBarrierParameters(response[3:])
	// case 0x1C:
	// 	result += InterpretBarrierType(response[3], response[4])
	// default:
	// 	result += "Interpretation not implemented for this query type\n"
	// }

	return "ok", nil
}

func (c *Controller) SendCommand(command byte, function byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.IsConnected {
		return fmt.Errorf("not connected")
	}

	data := []byte{0x55, 0x03, 0x01, command, function, 0x00, 0x00} // SD, LE, D0, D1, D2, CSH, CSL
	checksum := CalculateChecksum(data[:5])
	binary.BigEndian.PutUint16(data[5:], checksum)

	_, err := c.conn.Write(data)
	return err
}

func (c *Controller) GetStatus() (string, error) {
	BarrierPosition := "Unknown"
	// err := c.SendCommand(0x02, 0x00) // Status request command
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.IsConnected {
		return BarrierPosition, fmt.Errorf("not connected")
	}

	data := []byte{0x55, 0x02, 0x02, 0x07, 0x35, 0x5B}

	_, err := c.conn.Write(data)
	if err != nil {
		return BarrierPosition, fmt.Errorf("failed to send custom status request: %v", err)
	}

	select {
	case response := <-c.messages:
		// fmt.Printf("len:%v, %x ", len(response), response[2])
		if response[2] == 0x0c {
			// fmt.Printf("%x", response[3])
			Handle_0C_Message(response, c)
			// log.Debug().Msgf("Wrong message interprter Status: %x : %s", response, BarrierPosition)
		}
		log.Info().Str("BarrierIP", c.Barrierip).Msgf("Received response: %s : %s", BytesToHex(response), BarrierPosition)
		// return BarrierPositionnil
	case <-time.After(time.Duration(g.Config.TimeOutHttpResp) * time.Second):
		return BarrierPosition, fmt.Errorf("timeout waiting for status response")
	}
	return BarrierPosition, nil
}

func (c *Controller) SendCustomStatusRequest() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.IsConnected {
		return fmt.Errorf("not connected")
	}

	data := []byte{0x55, 0x02, 0x02, 0x02, 0x65, 0xFE}

	_, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send custom status request: %v", err)
	}

	return nil
}

func DecodeStatus(status []byte) (string, error) {

	if len(status) < 8 {
		return "", fmt.Errorf("status response too short")
	}

	debug := strings.Builder{}
	debug.WriteString("Controller Debug Status:\n")

	// Full message in hex
	debug.WriteString(fmt.Sprintf("Full message (hex): %s\n\n", BytesToHex(status)))

	// Interpret Status0
	debug.WriteString(fmt.Sprintf("Status0: %08b\n", status[3]))
	debug.WriteString(fmt.Sprintf("  Funk: %v\n", status[3]&(1<<0) != 0))
	debug.WriteString(fmt.Sprintf("  BT: %v\n", status[3]&(1<<1) != 0))
	debug.WriteString(fmt.Sprintf("  BTA1: %v\n", status[3]&(1<<2) != 0))
	debug.WriteString(fmt.Sprintf("  BTA2: %v\n", status[3]&(1<<3) != 0))
	debug.WriteString(fmt.Sprintf("  BTA3: %v\n", status[3]&(1<<4) != 0))
	debug.WriteString(fmt.Sprintf("  BTZ1A: %v\n", status[3]&(1<<5) != 0))
	debug.WriteString(fmt.Sprintf("  BTZ1B: %v\n", status[3]&(1<<6) != 0))
	debug.WriteString(fmt.Sprintf("  BTZ2: %v\n", status[3]&(1<<7) != 0))

	// Interpret Status1
	debug.WriteString(fmt.Sprintf("Status1: %08b\n", status[4]))
	debug.WriteString(fmt.Sprintf("  BTS: %v\n", status[4]&(1<<0) != 0))
	debug.WriteString(fmt.Sprintf("  Baum Ab: %v\n", status[4]&(1<<1) != 0))
	debug.WriteString(fmt.Sprintf("  LS: %v\n", status[4]&(1<<2) != 0))
	debug.WriteString(fmt.Sprintf("  SLZ: %v\n", status[4]&(1<<3) != 0))
	debug.WriteString(fmt.Sprintf("  F_ereignis_md: %v\n", status[4]&(1<<4) != 0))
	debug.WriteString(fmt.Sprintf("  SERVICE: %v\n", status[4]&(1<<5) != 0))
	debug.WriteString(fmt.Sprintf("  Netzüberwachung: %v\n", status[4]&(1<<6) != 0))

	// Interpret Status2
	debug.WriteString(fmt.Sprintf("Status2: %08b\n", status[5]))
	debug.WriteString(fmt.Sprintf("  Abgleich Schleife A: %v\n", status[5]&(1<<0) != 0))
	debug.WriteString(fmt.Sprintf("  Abgleich Schleife B: %v\n", status[5]&(1<<1) != 0))
	debug.WriteString(fmt.Sprintf("  Abgleich Schleife C: %v\n", status[5]&(1<<2) != 0))
	debug.WriteString(fmt.Sprintf("  Schleife A: %v\n", status[5]&(1<<3) != 0))
	debug.WriteString(fmt.Sprintf("  Schleife B: %v\n", status[5]&(1<<4) != 0))
	debug.WriteString(fmt.Sprintf("  Schleife C: %v\n", status[5]&(1<<5) != 0))
	debug.WriteString(fmt.Sprintf("  Schleife EXT: %v\n", status[5]&(1<<6) != 0))

	// Interpret Status3-7 (Multi relays)
	for i := 6; i < 12; i++ {
		debug.WriteString(fmt.Sprintf("Status%d (Multi %d-%d): %08b\n", i, (i-3)*8+1, (i-2)*8, status[i]))
	}

	return debug.String(), nil
}

// func main() {
// 	fmt.Println("************************* ELKA barrier Tester ******************************")
// 	Response = make(chan []byte, 100)
// 	// Define a string flag with a default value and a description
// 	strFlag := flag.String("ip", "192.168.25.200", "Input string")
// 	if *strFlag == "" {
// 		fmt.Println("No input provided. Usage: go run main.go -ip=\"BarrierIP\"")
// 		return
// 	}
// 	flag.Parse()
// 	address = *strFlag
// 	fmt.Printf("Working with : %v\n", address)
// 	controller := NewController()
// 	err := controller.Connect()
// 	if err != nil {
// 		fmt.Printf("Failed to connect: %v\n", err)
// 		return
// 	}
// 	defer controller.Disconnect()
// 	// Display the memory address of the variable
// 	// fmt.Printf("The memory address of controller is: %p\n", &controller)
// 	// fmt.Printf("The memory address of controller.messages is: %p\n", &controller.messages)
// 	// Start a goroutine to handle incoming messages

// 	go func() {
// 		for {
// 			// fmt.Printf("The memory address of controller.messages is: %p\n", &controller.messages)
// 			select {
// 			case msg := <-controller.messages:
// 				fmt.Printf("Received response len:%d message: %v (hex): %s\n", len(msg), msg, bytesToHex(msg))
// 				if msg[1] == 0x0A && msg[2] == 0x07 {
// 					decoded, _ := DecodeStatus(msg)
// 					fmt.Printf("decoded message: %v\n", decoded)
// 				} else if msg[0] == 0x55 && msg[1] == 0x02 && msg[2] == 0x0C {
// 					if msg[3] == 0x04 {
// 						fmt.Printf("Barrier status Open: %v\n", msg)
// 					} else if msg[03] == 0x05 {
// 						fmt.Printf("Barrier status Closed: %v\n", msg)
// 					} else {
// 						fmt.Printf("Barrier status Unknown: %v\n", msg)
// 					}

// 				}
// 				// fmt.Println("----------->send ro another channel")
// 				Response <- msg
// 				// Process the message as needed
// 			case <-controller.stopChan:
// 				fmt.Printf("Listening thread is dead")
// 				// return
// 			}
// 		}

// 	}()

// 	// Start a goroutine to check connection status and reconnect if necessary
// 	go func() {
// 		for {
// 			if !controller.IsConnected {
// 				fmt.Println("Connection lost. Attempting to reconnect...")
// 				err = controller.Reconnect()
// 				if err != nil {
// 				if err != nil {
// 					fmt.Printf("Failed to reconnect: %v\n", err)
// 					time.Sleep(retryDelay)
// 					continue
// 				}
// 				fmt.Println("Reconnected successfully")
// 				fmt.Printf("Reconnected memory address of controller is: %p\n", &controller)
// 			}
// 			time.Sleep(3 * time.Second)

// 		}
// 	}()

// 	// Main loop for keyboard input
// 	scanner := bufio.NewScanner(os.Stdin)
// 	// fmt.Println("Enter commands (open, close, lockopen, lockclose, unlockopen, unlockclose, unlock, status, details, , query, exit):")

// 	for {
// 		fmt.Println(Cyan + "Enter commands (open, close, lockopen, lockclose, unlockopen, unlockclose, unlock, status, details, , query, exit):" + Reset)
// 		fmt.Println(LightPurple + "For query command: query X ( from 1 to  1c / 155.2.2 Query telegram ) " + Reset)
// 		fmt.Println(Red + "For set command: set notifications X Y .... possible option (status,service,maintenance,gate,error,vehicle,position,motor,debug ) " + Reset)

// 		fmt.Print("> ")

// 		scanner.Scan()
// 		input := strings.ToLower(strings.TrimSpace(scanner.Text()))

// 		var err error
// 		switch {
// 		case input == "details":
// 			err = controller.SendCustomStatusRequest()
// 			if err != nil {
// 				fmt.Printf("Failed to send custom status request: %v\n", err)
// 			}
// 		case input == "open":
// 			err = controller.Open()
// 		case input == "close":
// 			err = controller.Close()
// 		case input == "lockopen":
// 			err = controller.LockOpen()
// 		case input == "lockclose":
// 			err = controller.LockClosed()
// 		case input == "unlockopen":
// 			err = controller.SendCommand(0x01, 0x02) // BA command, Deactivate function
// 		case input == "unlockclose":
// 			err = controller.SendCommand(0x02, 0x02) // BZ command, Deactivate function
// 		case input == "unlock":
// 			err = controller.Unlock()
// 		case strings.HasPrefix(input, "query "):
// 			queryTypeStr := strings.TrimPrefix(input, "query ")
// 			queryType, err := strconv.ParseUint(queryTypeStr, 16, 8)
// 			if err != nil {
// 				fmt.Printf("Invalid query type: %v\n", err)
// 				continue
// 			}
// 			result, err := controller.SendQueryTelegram(byte(queryType))
// 			if err != nil {
// 				fmt.Printf("Query failed: %v\n", err)
// 			} else {
// 				fmt.Println(result)
// 			}
// 		case input == "status":
// 			status := controller.GetStatus()
// 			fmt.Printf("Status: %v\n", status)
// 			continue

// 		case input == "exit":
// 			fmt.Println("Exiting...")
// 			return
// 		case strings.HasPrefix(input, "set notifications"):
// 			args := strings.Fields(input)[2:] // Get all words after "set notifications"
// 			flags := ParseNotificationFlags(args)
// 			err := controller.SetChangeNotifications(flags)
// 			if err != nil {
// 				fmt.Printf("Failed to set notifications: %v\n", err)
// 			} else {
// 				fmt.Println("Notifications set successfully")
// 			}

// 		default:
// 			fmt.Println("Unknown command. Ignored.")
// 			continue
// 		}

// 		if err != nil {
// 			fmt.Printf("Command failed: %v\n", err)
// 		} else {
// 			fmt.Println("Command sent successfully")
// 		}
// 	}
// }
