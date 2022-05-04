package main

import (
	"github.com/gin-gonic/gin"
	"github.com/seiro-ogasawara/golang-todo-api-sample/infra/persistence/onmemory"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/handler"
	"github.com/seiro-ogasawara/golang-todo-api-sample/usecase"
)

func Route() *gin.Engine {
	repo := onmemory.NewOnmemoryTodoRepository()
	usecase := usecase.NewTodoUsecase(repo)
	handler := handler.NewTodoHandler(usecase)

	return api.Route(handler)
}

func main() {
	r := Route()
	if err := r.Run(); err != nil {
		panic(err)
	}
}
