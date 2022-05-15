package api

import (
	"github.com/gin-gonic/gin"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/handler"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/middleware"
)

func Route(
	auth middleware.AuthMiddleware,
	//dbMiddleware middleware.DBMiddleware,
	dbMiddleware *middleware.DBMiddleware,
	handler handler.TodoHandler,
) *gin.Engine {

	r := gin.Default()

	todoAPIGroup := r.Group("/todos")
	todoAPIGroup.Use(auth.NewAuthentication())

	todoAPIGroup.POST(
		"",
		dbMiddleware.NewTransaction(),
		handler.Create,
	)
	todoAPIGroup.GET(
		"",
		dbMiddleware.NewDB(),
		handler.List,
	)
	todoAPIGroup.GET(
		"/:id",
		dbMiddleware.NewDB(),
		handler.Get,
	)
	todoAPIGroup.PATCH(
		"/:id",
		dbMiddleware.NewTransaction(),
		handler.Update,
	)
	todoAPIGroup.DELETE(
		"/:id",
		dbMiddleware.NewTransaction(),
		handler.Delete,
	)

	return r
}
