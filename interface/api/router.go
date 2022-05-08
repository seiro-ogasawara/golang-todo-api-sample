package api

import (
	"github.com/gin-gonic/gin"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/handler"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/middleware"
)

func Route(auth middleware.AuthMiddleware, handler handler.TodoHandler) *gin.Engine {
	r := gin.Default()

	todoAPIGroup := r.Group("/todos")
	todoAPIGroup.Use(auth.NewAuthentication())

	todoAPIGroup.POST("", handler.Create)
	todoAPIGroup.GET("", handler.List)
	todoAPIGroup.GET("/:id", handler.Get)
	todoAPIGroup.PATCH("/:id", handler.Update)
	todoAPIGroup.DELETE("/:id", handler.Delete)

	return r
}
