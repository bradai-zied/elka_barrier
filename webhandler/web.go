package webhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ServeHomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Hello World",
	})
}
