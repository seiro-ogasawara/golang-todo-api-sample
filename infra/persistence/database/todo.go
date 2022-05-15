package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/model"
	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/repository"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility/db"
	"gorm.io/gorm"
)

type databaseTodoRepository struct {
}

func NewDatabaseTodoRepository() repository.TodoRepository {
	return &databaseTodoRepository{}
}

func (r *databaseTodoRepository) Create(ctx context.Context, todo model.Todo) (int, error) {
	now := time.Now()
	todo.CreatedAt = now
	todo.UpdatedAt = now
	if err := db.GetDBFromContext(ctx).Create(&todo).Error; err != nil {
		return 0, utility.InternalServerError("can't create todo", err)
	}
	return todo.ID, nil
}

func (r *databaseTodoRepository) Get(ctx context.Context, userID string, id int) (*model.Todo, error) {
	var ret model.Todo
	if err := db.GetDBFromContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&ret).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utility.NotFound(fmt.Sprintf("todo with id %d is not found", id), err)
		}
		return nil, utility.InternalServerError(fmt.Sprintf("todo with id %d is not found", id), err)
	}
	return &ret, nil
}

func (r *databaseTodoRepository) List(
	ctx context.Context, userID string, sortBy model.Sorter, orderBy model.Order, includeDone bool,
) ([]*model.Todo, error) {
	query := db.GetDBFromContext(ctx).
		Where("user_id = ?", userID).
		Order(fmt.Sprintf("%s %s", string(sortBy), string(orderBy)))
	if !includeDone {
		query.Where("status <> ?", int(model.StatusDone))
	}

	var ret []*model.Todo
	if err := query.Find(&ret).Error; err != nil {
		return nil, utility.InternalServerError(fmt.Sprintf("can't find todo for user %s from db", userID), err)
	}
	return ret, nil
}

func (r *databaseTodoRepository) Update(ctx context.Context, todo *model.Todo) error {
	todo.UpdatedAt = time.Now()
	result := db.GetDBFromContext(ctx).Save(todo)
	if err := result.Error; err != nil {
		return utility.InternalServerError(fmt.Sprintf("todo with id %d is not found", todo.ID), err)
	}
	if result.RowsAffected == 0 {
		return utility.NotFound("", fmt.Errorf(fmt.Sprintf("todo with id %d is not found", todo.ID)))
	}
	return nil
}

func (r *databaseTodoRepository) Delete(ctx context.Context, id int) error {
	result := db.GetDBFromContext(ctx).
		Where("id = ?", id).
		Delete(&model.Todo{})
	if err := result.Error; err != nil {
		return utility.InternalServerError(fmt.Sprintf("can't delete todo with id %d from db", id), err)
	}
	if result.RowsAffected == 0 {
		return utility.NotFound("", fmt.Errorf(fmt.Sprintf("todo with id %d is not found", id)))
	}
	return nil
}
