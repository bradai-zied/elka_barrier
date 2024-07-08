package utils

import (
	"go_barrier/elka"
	g "go_barrier/globals"

	"github.com/gin-gonic/gin"
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
