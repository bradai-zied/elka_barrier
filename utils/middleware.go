package utils

import (
	g "go_barrier/globals"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CheckMiddleware is a sample middleware for checking conditions
func CheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Example: Check if a specific query parameter is present
		requiredParam := c.Param("id")
		if requiredParam == "" {
			log.Error().Msg("Missing required query parameter id")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required query parameter"})
			c.Abort()
			return
		}
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Error().Err(err).Msg("Cannot Convert BArrier id " + c.Param("id"))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid barrier ID" + c.Param("id")})
			c.Abort()
			return
		}

		// Find the barrier connection
		_, exists := g.BarrierId2IP[id]
		if !exists {
			log.Error().Int("id", id).Msg("Barrier not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Barrier " + c.Param("id") + " not found"})
			c.Abort()
			return
		}

		// Allow the request to proceed
		c.Next()
	}
}
