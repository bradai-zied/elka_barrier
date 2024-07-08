package barrierconfig

import (
	"go_barrier/def"
	g "go_barrier/globals"
	"go_barrier/utils"
	"io/ioutil"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func GetAppPort() int {
	return g.Config.AppPort
}

func LoadConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Err(err).Msg("Error Read barrier.yaml File")
		return err
	}

	g.Config = def.YAMLConfig{}
	err = yaml.Unmarshal(data, &g.Config)
	if err != nil {
		log.Err(err).Msg("Error Unmarshal Yaml File")
		return err
	}

	for _, yamlBarrier := range g.Config.Barriers {
		log.Debug().Msgf(" ################### COnfig FIle: %v ################################", yamlBarrier)
		for _, id := range yamlBarrier.ID {
			if !utils.ContainsInt(g.BarrierIds, id) {
				g.BarrierIds = append(g.BarrierIds, id)
				g.BarrierId2IP[id] = yamlBarrier.IP
			}
			// barrier := def.Barrier{
			// 	IP:          yamlBarrier.IP,
			// 	Name:        yamlBarrier.Name,
			// 	ID:          id, // Assuming the first ID is the main one
			// 	BarrierType: yamlBarrier.BarrierType,
			// 	Port:        yamlBarrier.Port,
			// }
			// response.Barriers = append(response.Barriers, barrier)
		}

	}
	log.Debug().Msgf("All Barrier ID: %v", g.BarrierIds)
	return nil
}

func SaveConfig() error {
	filename := g.Filename
	data, err := yaml.Marshal(g.Config)
	if err != nil {
		log.Err(err).Msg("Error save barrier.yaml File")
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}
