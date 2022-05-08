package main

import (
	"github.com/gin-gonic/gin"
	"github.com/seiro-ogasawara/golang-todo-api-sample/infra/persistence/onmemory"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/handler"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/middleware"
	"github.com/seiro-ogasawara/golang-todo-api-sample/usecase"
)

func Route() *gin.Engine {
	todoRepo := onmemory.NewOnmemoryTodoRepository()
	userRepo := onmemory.NewOnmemoryUserRepository()
	usecase := usecase.NewTodoUsecase(todoRepo)
	handler := handler.NewTodoHandler(usecase)
	authMiddleware := middleware.NewAuthMiddleware(userRepo)

	return api.Route(authMiddleware, handler)
}

func main() {
	r := Route()
	if err := r.Run(); err != nil {
		panic(err)
	}
}
