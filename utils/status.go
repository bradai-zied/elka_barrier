package utils

import (
	"go_barrier/elka"
	g "go_barrier/globals"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func DebugHandler(c *gin.Context) {

	c.JSON(200, gin.H{
		"elka.ElkaController": &elka.ElkaController,
		"BarrierIds":          g.BarrierIds,
		"BarrierId2IP":        g.BarrierId2IP,
		"g.Config.Barriers":   g.Config.Barriers,
		// "Cam2Lane":    g.Cam2Lane,
		// "FreeFlowApi": g.FreeFlowApi,
	})
}

func Restart(c *gin.Context) {

	go func() {
		log.Warn().Msg("Request to kill applicaiton")
		// Wait a bit to ensure the response is sent before shutdown
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()
	// var response []elka.Controller
	c.JSON(200, gin.H{"message": "Server is shutting down..."})
}
func AllStatusHandler(c *gin.Context) {
	// var response []elka.Controller
	c.JSON(200, gin.H{
		"elka.elka.IPElkaController": &elka.IPElkaController,
		// "BarrierIds":          g.BarrierIds,
		// "BarrierId2IP":        g.BarrierId2IP,
		// "g.Config.Barriers":   g.Config.Barriers,
		// "Cam2Lane":    g.Cam2Lane,
		// "FreeFlowApi": g.FreeFlowApi,
	})
}
