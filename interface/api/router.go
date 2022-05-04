package api

import (
	"github.com/gin-gonic/gin"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/handler"
)

func Route(handler handler.TodoHandler) *gin.Engine {
	r := gin.Default()
	r.POST("/todos", handler.Create)
	r.GET("/todos", handler.List)
	r.GET("/todos/:id", handler.Get)
	r.PATCH("/todos/:id", handler.Update)
	r.DELETE("/todos/:id", handler.Delete)

	return r
}
