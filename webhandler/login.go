package webhandler

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		// Abort the request with the appropriate error code
		c.Abort()
		c.Redirect(http.StatusFound, "/login")
		return
	}
	// Continue down the chain to handler etc
	c.Next()
}

func ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login",
	})
}

func PerformLogin(c *gin.Context) {
	session := sessions.Default(c)
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Validate form input
	if username == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Parameters can't be empty"})
		return
	}

	// Check for username and password match, usually from a database
	if (username == "admin" && password == "password") || (username == "rami" && password == "rami") || (username == "zied" && password == "zied") {
		session.Set("user", username) // In real world, set user ID instead of username
		if err := session.Save(); err != nil {
			c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "Failed to save session"})
			return
		}
		c.Redirect(http.StatusFound, "/")
	} else {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Authentication failed"})
	}
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}
	session.Delete("user")
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.Redirect(http.StatusFound, "/login")
}
