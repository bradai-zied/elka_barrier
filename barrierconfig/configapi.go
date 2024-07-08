package barrierconfig

import (
	"go_barrier/def"
	g "go_barrier/globals"
	"go_barrier/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func GetAllBarriers() gin.HandlerFunc {
	return func(c *gin.Context) {
		response := def.BarrierResponse{
			Barriers: make([]def.Barrier, 0, len(g.Config.Barriers)),
		}

		for _, yamlBarrier := range g.Config.Barriers {
			for _, id := range yamlBarrier.ID {
				if !utils.ContainsInt(g.BarrierIds, id) {
					g.BarrierIds = append(g.BarrierIds, id)
				}
				barrier := def.Barrier{
					IP:          yamlBarrier.IP,
					Name:        yamlBarrier.Name,
					ID:          id, // Assuming the first ID is the main one
					BarrierType: yamlBarrier.BarrierType,
					Port:        yamlBarrier.Port,
				}
				response.Barriers = append(response.Barriers, barrier)
			}

		}
		c.JSON(http.StatusOK, response)
	}
}

func GetBarrier() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barrier ID"})
			return
		}
		// log.Debug().Msgf("All Barrier ID: %v", g.BarrierIds)
		if !utils.ContainsInt(g.BarrierIds, id) {
			log.Warn().Msgf("Barrier %s not found", idStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Barrier not found"})
			return
		}

		for _, yamlBarrier := range g.Config.Barriers {
			for _, barrierid := range yamlBarrier.ID {
				if barrierid == id {
					barrier := def.Barrier{
						IP:          yamlBarrier.IP,
						Name:        yamlBarrier.Name,
						ID:          id,
						BarrierType: yamlBarrier.BarrierType,
						Port:        yamlBarrier.Port,
					}
					c.JSON(http.StatusOK, barrier)
					return
				}
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Barrier not found"})
	}
}

func AddBarrier() gin.HandlerFunc {
	return func(c *gin.Context) {

		var newBarrier def.Barrier
		if err := c.ShouldBindJSON(&newBarrier); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if utils.ContainsInt(g.BarrierIds, newBarrier.ID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Barrier ID already exists"})
			return
		}

		// Check if a barrier with the same IP exists
		var existingBarrierIndex = -1
		for i, yamlBarrier := range g.Config.Barriers {
			if yamlBarrier.IP == newBarrier.IP {
				existingBarrierIndex = i
				break
			}
		}

		if existingBarrierIndex != -1 {
			// Barrier with the same IP exists, update it
			existingBarrier := &g.Config.Barriers[existingBarrierIndex]

			// Add the new ID if it doesn't exist
			idExists := false
			for _, id := range existingBarrier.ID {
				if id == newBarrier.ID {
					idExists = true
					break
				}
			}
			if !idExists {
				existingBarrier.ID = append(existingBarrier.ID, newBarrier.ID)
			}

			// Update other fields
			existingBarrier.Name = newBarrier.Name
			existingBarrier.BarrierType = newBarrier.BarrierType
			existingBarrier.Port = newBarrier.Port

			c.JSON(http.StatusOK, gin.H{"message": "Barrier updated successfully"})
		} else {
			// Barrier doesn't exist, add a new one
			yamlBarrier := def.YAMLBarrier{
				IP:          newBarrier.IP,
				Name:        newBarrier.Name,
				ID:          []int{newBarrier.ID},
				BarrierType: newBarrier.BarrierType,
				Port:        newBarrier.Port,
			}

			g.Config.Barriers = append(g.Config.Barriers, yamlBarrier)

			c.JSON(http.StatusOK, gin.H{"message": "Barrier added successfully"})
		}

		err := SaveConfig()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save barrier configuration"})
			return
		}
		LoadConfig(g.Filename)
	}
}

func ModifyBarrier() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Error().Err(err).Msg("Invalid barrier ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barrier ID"})
			return
		}

		var modifiedBarrier def.Barrier
		if err := c.ShouldBindJSON(&modifiedBarrier); err != nil {
			log.Error().Err(err).Msg("Invalid barrier data")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !utils.ContainsInt(g.BarrierIds, modifiedBarrier.ID) {
			log.Error().Err(err).Msg("Barrier ID dont exists")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Barrier ID dont exists"})
			return
		}

		found := false
		for i, yamlBarrier := range g.Config.Barriers {
			for j, barrierID := range yamlBarrier.ID {
				if barrierID == id {
					// Update the existing barrier
					g.Config.Barriers[i].Name = modifiedBarrier.Name
					g.Config.Barriers[i].BarrierType = modifiedBarrier.BarrierType
					g.Config.Barriers[i].Port = modifiedBarrier.Port

					// Update IP if it's different
					if yamlBarrier.IP != modifiedBarrier.IP {
						g.Config.Barriers[i].IP = modifiedBarrier.IP
					}

					// Update ID if it's different
					if modifiedBarrier.ID != id {
						g.Config.Barriers[i].ID[j] = modifiedBarrier.ID
					}

					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			log.Warn().Int("id", id).Msg("Barrier not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Barrier not found"})
			return
		}

		err = SaveConfig()
		if err != nil {
			log.Error().Err(err).Msg("Failed to save barrier configuration")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save barrier configuration"})
			return
		}

		log.Info().Int("id", id).Msg("Barrier modified successfully")
		c.JSON(http.StatusOK, gin.H{"message": "Barrier modified successfully"})
		LoadConfig(g.Filename)
	}
}

func DeleteBarrier() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Error().Err(err).Msg("Invalid barrier ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barrier ID"})
			return
		}
		if !utils.ContainsInt(g.BarrierIds, id) {
			log.Error().Err(err).Msg("Barrier ID dont exists")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Barrier ID dont exists"})
			return
		}

		found := false
		for i, yamlBarrier := range g.Config.Barriers {
			for j, barrierID := range yamlBarrier.ID {
				if barrierID == id {
					// Remove the ID from the list
					g.Config.Barriers[i].ID = append(g.Config.Barriers[i].ID[:j], g.Config.Barriers[i].ID[j+1:]...)

					// If it was the last ID, remove the entire barrier
					if len(g.Config.Barriers[i].ID) == 0 {
						g.Config.Barriers = append(g.Config.Barriers[:i], g.Config.Barriers[i+1:]...)
					}

					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			log.Warn().Int("id", id).Msg("Barrier not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Barrier not found"})
			return
		}

		err = SaveConfig()
		if err != nil {
			log.Error().Err(err).Msg("Failed to save barrier configuration")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save barrier configuration"})
			return
		}

		log.Info().Int("id", id).Msg("Barrier deleted successfully")
		c.JSON(http.StatusOK, gin.H{"message": "Barrier deleted successfully"})
		LoadConfig(g.Filename)
	}
}
