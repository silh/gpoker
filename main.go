package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	app := gin.Default()
	app.LoadHTMLGlob("templates/*")

	app.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.gohtml", nil)
	})
	app.Run()
}
