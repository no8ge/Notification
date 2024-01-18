package router

import (
	"github.com/gin-gonic/gin"
	handler "github.com/no8geo/notify/pkg/handler"
	"github.com/olahol/melody"
)

func V1(r *gin.Engine, m *melody.Melody) {

	notifcation := r.Group("/v1")
	{
		notifcation.GET("/index", handler.Index)
	}

	webSocket := r.Group("v1")
	{
		webSocket.GET("/ws/pull", handler.Pull(m))
		webSocket.GET("/ws/push", handler.Push(m))
		webSocket.GET("/ws/watch", handler.Watch(m))
	}
}
