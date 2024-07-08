package main

import (
	"fmt"

	"go_barrier/barrierconfig"
	"go_barrier/barriercontrol"
	"go_barrier/utils"
	"go_barrier/webhandler"

	g "go_barrier/globals"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	g.BarrierIds = make([]int, 0)
	g.BarrierId2IP = make(map[int]string)
	g.Filename = "barrier.yaml"
	utils.InitLogger()

	log.Debug().Msg("***** Gobarrier Start Initialisation 3.01 *****")
	err := barrierconfig.LoadConfig(g.Filename)
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading barrier.yaml")
	}
	// Initialize barrier connections
	// barriercontrol.InitializeConnections(g.Config.Barriers)
	utils.BuildElkaControllerMap()

	r := gin.Default()
	// Apply middleware globally
	// r.Use(middleware.CheckMiddleware())

	//WEB
	r.GET("/", webhandler.ServeHomePage)
	r.LoadHTMLGlob("templates/*")

	//debug
	r.GET("/debug", utils.DebugHandler)

	r.GET("/Barrier", barrierconfig.GetAllBarriers())
	r.GET("/Barrier/:id", barrierconfig.GetBarrier())
	r.POST("/add", barrierconfig.AddBarrier())
	r.PUT("/modify/:id", barrierconfig.ModifyBarrier())
	r.DELETE("/delete/:id", barrierconfig.DeleteBarrier())

	barrierRoutes := r.Group("/", utils.CheckMiddleware())
	{
		// Add new routes for barrier control
		barrierRoutes.POST("/open/:id", barriercontrol.OpenBarrier())
		barrierRoutes.POST("/close/:id", barriercontrol.CloseBarrier())
		barrierRoutes.POST("/unlock/:id", barriercontrol.UnlockBarrier())
		barrierRoutes.POST("/lock/:id", barriercontrol.LockBarrier())
		barrierRoutes.GET("/status/:id", barriercontrol.GetBarrierStatus())
		barrierRoutes.POST("/config/:id", barriercontrol.SetBarrierConfig())
		barrierRoutes.GET("/query/:id", barriercontrol.Querydata())
	}
	port := barrierconfig.GetAppPort()

	log.Info().Int("port", port).Msg("Starting server")
	r.Run(fmt.Sprintf(":%d", port))
}
