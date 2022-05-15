package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/seiro-ogasawara/golang-todo-api-sample/infra/persistence/database"

	// "github.com/seiro-ogasawara/golang-todo-api-sample/infra/persistence/onmemory"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/handler"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/middleware"
	"github.com/seiro-ogasawara/golang-todo-api-sample/usecase"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility/db"
)

func Route() *gin.Engine {

	db, err := db.GetDBFromEnvironmentVariables()
	if err != nil {
		log.Fatalf("failed to access database: %v\n", err)
	}

	//todoRepo := onmemory.NewOnmemoryTodoRepository()
	//userRepo := onmemory.NewOnmemoryUserRepository()
	todoRepo := database.NewDatabaseTodoRepository()
	userRepo := database.NewDatabaseUserRepository()
	usecase := usecase.NewTodoUsecase(todoRepo)
	handler := handler.NewTodoHandler(usecase)
	authMiddleware := middleware.NewAuthMiddleware(userRepo)
	dbMiddleware := middleware.NewDBMiddleware(db)

	return api.Route(authMiddleware, dbMiddleware, handler)
}

func main() {
	r := Route()
	if err := r.Run(); err != nil {
		panic(err)
	}
}
