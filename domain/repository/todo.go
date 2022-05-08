package repository

import (
	"context"

	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/model"
)

type TodoRepository interface {
	Create(ctx context.Context, todo model.Todo) (int, error)
	Get(ctx context.Context, userID string, id int) (*model.Todo, error)
	List(ctx context.Context, userID string, sortBy model.Sorter, orderBy model.Order, includeDone bool) ([]*model.Todo, error)
	Update(ctx context.Context, todo *model.Todo) error
	Delete(ctx context.Context, id int) error
}
