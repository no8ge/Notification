package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

func Index(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "template/index.html")
}

func Pull(m *melody.Melody) gin.HandlerFunc {
	return func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	}
}

func Push(m *melody.Melody) gin.HandlerFunc {
	return func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	}
}

func Watch(m *melody.Melody) gin.HandlerFunc {
	return func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	}
}
