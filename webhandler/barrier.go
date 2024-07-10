package webhandler

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ServeBarriersPage(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	c.HTML(http.StatusOK, "barriers.html", gin.H{
		"title": "Barriers Management",
		"user":  user,
	})
}
