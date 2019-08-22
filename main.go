package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()

	r.LoadHTMLGlob("templates/*.tpl")
	r.Static("/static", "./static")
	r.Static("/articles/img", "./articles/img")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tpl", nil)
	})
	r.Run(":8080")
}
