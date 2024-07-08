package utils

import (
	"fmt"
	"go_barrier/elka"
	g "go_barrier/globals"
	"time"

	"github.com/rs/zerolog/log"
)

func SetDefaultConfig(c *elka.Controller) {

	args := g.Config.DefaultNotification
	flags := elka.ParseNotificationFlags(args)
	err := c.SetChangeNotifications(flags)
	fmt.Sprintf("Response from SetChangeNotifications :%v", err)
	time.Sleep(2000 * time.Millisecond)
	c.SendQueryTelegram(byte(0x02))
}
func BuildElkaControllerMap() {
	n := len(g.Config.Barriers)
	elka.ElkaController = make(map[int]*elka.Controller, n)
	for _, yamlBarrier := range g.Config.Barriers {
		elkaC := elka.NewController(yamlBarrier.IP)
		for _, id := range yamlBarrier.ID {
			elka.ElkaController[id] = elkaC
		}
		log.Info().Msgf("Connect to Barrier:%s", elkaC.GetBarrierIP())
		go func() {
			err := elkaC.Connect()
			if err != nil {
				log.Err(err).Msgf("Failed to connect Barrier IP:%s ", elkaC.GetBarrierIP())
				return
			}
			SetDefaultConfig(elkaC)

			// defer elkaC.Disconnect()
		}()
		log.Debug().Msgf("Sleeping for 100 milliseconds... for next barrier")
		time.Sleep(100 * time.Millisecond)
	}
}
