package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/static"
)

func main() {
	router := gin.Default()
	router.Use(static.Serve("/static", static.LocalFile("static", true)))
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Upload Page",
		})
	})
	router.Run("0.0.0.0:20605")
}