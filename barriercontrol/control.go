package barriercontrol

import (
	"fmt"
	"go_barrier/elka"
	g "go_barrier/globals"
	"go_barrier/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func OpenBarrier() gin.HandlerFunc {
	return func(c *gin.Context) {
		lock := c.DefaultQuery("lock", "false")
		id, err := strconv.Atoi(c.Param("id"))
		_ = c.DefaultQuery("extradata", "")
		if err != nil {
			log.Error().Err(err).Msg("Invalid barrier ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barrier ID"})
			return
		}
		log.Debug().Int("Lane ", id).Msgf("Send open to barrier IP: %s", elka.ElkaController[id].Barrierip)

		if elka.ElkaController[id].IsLockedDown {
			log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Barrier locked Down")
			c.JSON(http.StatusConflict, gin.H{"message": "Barrier is Locked Down"})
			return
		}
		if elka.ElkaController[id].BarrierPositionStr == "Open" {
			elka.ElkaController[id].IsClosed = false
			log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Barrier opened already")
			c.JSON(http.StatusOK, gin.H{"message": "Barrier Already Open"})
			return
		}

		if len(elka.ElkaController[id].MessageToApi) > 0 {
			select {

			case OldMsg := <-elka.ElkaController[id].MessageToApi:
				log.Debug().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Int("ChannelLength", len(elka.ElkaController[id].MessageToApi)).Msgf("OldMsg: %s" + OldMsg)

			default:
				// Channel is empty now
			}

		}

		if lock == "true" && !elka.ElkaController[id].IsLockedDown {
			elka.ElkaController[id].LockOpen()
			elka.ElkaController[id].IsLockedUp = true
		} else {
			elka.ElkaController[id].Open()
		}

		select {
		case <-time.After(time.Duration(g.Config.TimeOutHttpResp) * time.Millisecond):
			log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Failed to Close barrier")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Barrier Not responding"})
			return
		case openrepsone := <-elka.ElkaController[id].MessageToApi:
			log.Debug().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Int("ChannelLength", len(elka.ElkaController[id].MessageToApi)).Str("openrepsone", openrepsone).Msg("+++++++++++++++++++++WHTF")
			switch openrepsone {
			case "Open":
				// log.Debug().Msg("Ignore OlD message")
				elka.ElkaController[id].IsClosed = false
				c.JSON(http.StatusOK, gin.H{"message": "Barrier already open successfully"})
				return
			case "Close":
				log.Debug().Msg("Ignore OlD message")
			case "ack":
				log.Debug().Msg("Ignore ack message")
			case "nak":
				log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Str("BArrier message", openrepsone).Msg("Barrier Refused to close")
				c.JSON(http.StatusOK, gin.H{"message": "Barrier Refused to open"})

				// default:
				// 	log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Str("BArrier message", closerepsone).Msg("Barrier Not responding")
			}
		}
		openresponse := "default"
		select {
		case <-time.After(time.Duration(g.Config.TimeOutHttpResp) * time.Millisecond):
			log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Failed to Open barrier")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Barrier Not responding"})
			return
		case openresponse = <-elka.ElkaController[id].MessageToApi:
			log.Debug().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Str("BArrier message", openresponse).Msg("________Barrier Open Response :")
			if openresponse == "Open" || openresponse == "Opening" {
				elka.ElkaController[id].IsClosed = false
				c.JSON(http.StatusOK, gin.H{"message": "Barrier open successfully"})
				return
			}
		}
		// log.Info().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Barrier Closed successfully")
		log.Error().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Str("openresponse", openresponse).Msg("Uknown message while waiting Open message")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Barrier Not responding"})
	}
}

func CloseBarrier() gin.HandlerFunc {
	return func(c *gin.Context) {
		lock := c.DefaultQuery("lock", "false")
		_ = c.DefaultQuery("extradata", "")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Error().Err(err).Msg("Invalid barrier ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barrier ID"})
			return
		}
		log.Debug().Msgf("send close to id: %d", id)

		if elka.ElkaController[id].IsLockedUp {
			log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Barrier locked Up")
			elka.ElkaController[id].IsClosed = false
			c.JSON(http.StatusConflict, gin.H{"message": "Barrier is Locked UP"})
			return
		}

		if elka.ElkaController[id].IsLockedDown {
			log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Barrier locked down")
			// elka.ElkaController[id].IsClosed = true
			c.JSON(http.StatusOK, gin.H{"message": "Barrier is Locked down"})
			return
		}
		//empty channel
		if len(elka.ElkaController[id].MessageToApi) > 0 {
			select {
			case OldMsg := <-elka.ElkaController[id].MessageToApi:
				log.Debug().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Int("ChannelLength", len(elka.ElkaController[id].MessageToApi)).Msgf("OldMsg: %s" + OldMsg)
			default:
				// Channel is empty now
			}
		}

		// }
		if lock == "true" {
			elka.ElkaController[id].LockClosed()
			elka.ElkaController[id].IsLockedDown = true
			elka.ElkaController[id].IsClosed = true
		} else {
			elka.ElkaController[id].Close()
			elka.ElkaController[id].IsClosed = true
		}
		time.Sleep(300 * time.Millisecond)
		emptybool := true
		for emptybool {
			select {
			case <-time.After(time.Duration(g.Config.TimeOutHttpResp) * time.Second):
				log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Failed to unlock barrier")
				// c.JSON(http.StatusBadRequest, gin.H{"message": "Barrier Unlocked Refused"})
			case Oldmsg := <-elka.ElkaController[id].MessageToApi:
				log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msgf("Oldmsg %s", Oldmsg)
				// c.JSON(http.StatusBadRequest, gin.H{"message": "Barrier Unlocked Refused"})
			default:
				emptybool = false
			}
		}
		log.Debug().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Str("BArrier message", elka.ElkaController[id].BarrierPositionStr).Msg("________Barrier Close Response :")
		if elka.ElkaController[id].BarrierPositionStr == "Closed" || elka.ElkaController[id].BarrierPositionStr == "Closing" {
			c.JSON(http.StatusOK, gin.H{"message": "Barrier Closed successfully"})
			return
		}

		// log.Info().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Barrier Closed successfully")
		log.Error().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Str("closerepsone", elka.ElkaController[id].BarrierPositionStr).Msg("Uknown message while waiting closing message")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Barrier Not responding"})
	}
}

func UnlockBarrier() gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = c.DefaultQuery("extradata", "")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Error().Err(err).Msg("Invalid barrier ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barrier ID"})
			return
		}

		log.Debug().Int("Lane ", id).Msgf("Send Unlock to barrier IP: %s", elka.ElkaController[id].Barrierip)
		//empty channel
		if len(elka.ElkaController[id].MessageToApi) > 0 {
			select {

			case OldMsg := <-elka.ElkaController[id].MessageToApi:
				log.Debug().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Int("ChannelLength", len(elka.ElkaController[id].MessageToApi)).Msgf("OldMsg: %s" + OldMsg)

			default:
				// Channel is empty now
			}

		}

		//send telegram to check barrier inputs
		elka.ElkaController[id].SendQueryTelegram(0x02)
		time.Sleep(1000 * time.Millisecond)
		emptybool := true
		for emptybool {
			select {
			case <-time.After(time.Duration(g.Config.TimeOutHttpResp) * time.Second):
				log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Failed to unlock barrier")
				// c.JSON(http.StatusBadRequest, gin.H{"message": "Barrier Unlocked Refused"})
			case Oldmsg := <-elka.ElkaController[id].MessageToApi:
				log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msgf("Oldmsg %s", Oldmsg)
				// c.JSON(http.StatusBadRequest, gin.H{"message": "Barrier Unlocked Refused"})
			default:
				emptybool = false
			}
		}
		elka.ElkaController[id].Unlock()
		elka.ElkaController[id].IsLockedUp = false
		elka.ElkaController[id].IsLockedDown = false
		time.Sleep(200 * time.Millisecond)
		log.Debug().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Status")
		if elka.ElkaController[id].LoopA {
			log.Debug().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("LoopA is active open Barrier after unlock")
			elka.ElkaController[id].Open()
		}

		// barrierstatus, err := elka.ElkaController[id].GetBarrierStatus()
		// if err != nil {
		// 	log.Error().Err(err).Int("id", id).Msg("Failed to get barrier status")
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get barrier status"})
		// 	return
		// }
		// log.Debug().Msgf("barrierstatus: %s", string(barrierstatus))

		log.Info().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Barrier Unlocked successfully")
		c.JSON(http.StatusOK, gin.H{"message": "Barrier Unlocked successfully"})
	}
}

func LockBarrier() gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = c.DefaultQuery("extradata", "")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Error().Err(err).Msg("Invalid barrier ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barrier ID"})
			return
		}

		log.Debug().Int("Lane ", id).Msgf("Send LockOpen to barrier IP: %s", elka.ElkaController[id].Barrierip)

		elka.ElkaController[id].LockOpen()
		// elka.ElkaController[id].Open() // to romve when implemnt check loop
		elka.ElkaController[id].IsLockedUp = false
		// elka.ElkaController[id].IsLockedDown = false

		// barrierstatus, err := elka.ElkaController[id].GetBarrierStatus()
		// if err != nil {
		// 	log.Error().Err(err).Int("id", id).Msg("Failed to get barrier status")
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get barrier status"})
		// 	return
		// }
		// log.Debug().Msgf("barrierstatus: %s", string(barrierstatus))

		log.Info().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msg("Barrier Locked Up successfully")
		c.JSON(http.StatusOK, gin.H{"message": "Barrier Locked Up successfully"})
	}
}
func GetBarrierStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = c.DefaultQuery("extradata", "")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Error().Err(err).Msg("Invalid barrier ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barrier ID"})
			return
		}

		log.Debug().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msgf("send GetStatus : %d", id)
		status := elka.ElkaController[id].BarrierPositionStr
		if elka.ElkaController[id].IsLockedDown {
			status = "LockedDown"
		}
		if elka.ElkaController[id].IsLockedUp {
			status = "LockedUp"
		}
		c.JSON(http.StatusOK, status)
	}
}

var validFunctions = []string{"status", "service", "maintenance", "gate", "error", "vehicle", "position"}

func SetBarrierConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Error().Err(err).Msg("Invalid barrier ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barrier ID"})
			return
		}
		// possible value :service,maintenance,gate,error,vehicle,position never use motor,debug
		var requestBody struct {
			Function []string `json:"function"`
		}

		// Bind the JSON body to the requestBody struct
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			log.Error().Err(err).Msg("Invalid request body")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		// Check if the strings in the `function` array are part of the predefined list
		var validFunctionList []string
		if len(requestBody.Function) == 0 {
			validFunctionList = []string{}

		} else {

			for _, fn := range requestBody.Function {
				log.Debug().Msgf("function to check: %s", fn)
				if utils.Contains(validFunctions, fn) {
					validFunctionList = append(validFunctionList, fn)
				} else {
					log.Warn().Str("function", fn).Msg("Invalid function")
				}
			}

			if len(validFunctionList) == 0 {
				log.Error().Msg("No valid functions provided")
				c.JSON(http.StatusBadRequest, gin.H{"error": "No valid functions provided all",
					"bodyStruct": "{\"function\": [\"service\", \"maintenance\", \"gate\", \"error\", \"position\", \"vehicle\"]}",
					"details":    validFunctions})
				return
			}
		}
		// Do something with the validFunctionList
		log.Info().Int("id", id).Strs("valid_functions", validFunctionList).Msg("Valid functions processed")

		c.JSON(http.StatusOK, gin.H{"valid_functions": validFunctionList})

		flags := elka.ParseNotificationFlags(validFunctionList)
		log.Debug().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Msgf("Try to set config : %d", id)
		err = elka.ElkaController[id].SetChangeNotifications(flags)
		if err != nil {
			log.Error().Err(err).Msgf("%v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Error setting configuration: %v", err)})
		} else {
			log.Error().Err(err).Msg("Configuration set successfully")
		}
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// var validquery = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0A", "0B", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}

func Querydata() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		query := c.Query("query")
		if err != nil {
			log.Error().Err(err).Msg("Invalid barrier ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barrier ID"})
			return
		}

		// possible value :service,maintenance,gate,error,vehicle,position never use motor,debug
		var requestBody struct {
			Function []string `json:"function"`
		}

		// Bind the JSON body to the requestBody struct
		// if err := c.ShouldBindJSON(&requestBody); err != nil {
		// 	log.Error().Err(err).Msg("Invalid request body")
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		// 	return
		// }
		queryType, err := strconv.ParseUint(query, 16, 8)
		if err != nil {
			log.Warn().Str("query", query).Msg("Invalid query")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query"})
		}
		// Check if the integer is between 0x00 and 0x1C
		if queryType > 0x1C {
			log.Error().Uint64("value", queryType).Msg("Value out of range")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Value out of range"})
			return
		}
		// Do something with the validFunctionList
		log.Info().Int("id", id).Strs("valid_functions", requestBody.Function).Msg("Valid functions processed")

		result, err := elka.ElkaController[id].SendQueryTelegram(byte(queryType))
		if err != nil {
			log.Error().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Str("query", query).Msgf("Query failed: %v\n", err)
		} else {
			log.Debug().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Str("query", query).Msgf("Query response: %v\n", result)
		}

		select {
		case <-time.After(time.Duration(g.Config.TimeOutHttpResp) * time.Millisecond):
			log.Warn().Str("BarrierIP", elka.ElkaController[id].Barrierip).Int("id", id).Str("query", query).Msgf("Api call timeout reach waiting response after %d millisec", g.Config.TimeOutHttpResp)
			c.JSON(http.StatusOK, gin.H{"message": result})
			return
		case result := <-elka.ElkaController[id].MessageToApi:
			c.JSON(http.StatusOK, gin.H{"message": result})
		}

		c.JSON(http.StatusOK, gin.H{"message": result})

	}
}

// func (bc *BarrierConnection) sendOpenCommand() error {
// 	// Implement the actual open command based on ELKA barrier protocol
// 	return nil
// }

// func (bc *BarrierConnection) sendCloseCommand() error {
// 	// Implement the actual close command based on ELKA barrier protocol
// 	return nil
// }

// func (bc *BarrierConnection) getStatus() (string, error) {
// 	// Implement the actual status check based on ELKA barrier protocol
// 	return "", nil
// }
