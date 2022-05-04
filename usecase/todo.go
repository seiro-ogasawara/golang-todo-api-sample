package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/model"
	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/repository"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility"
)

type TodoUsecase interface {
	Create(ctx context.Context, title, description string, status, priority int) (*model.Todo, error)
	Get(ctx context.Context, id string) (*model.Todo, error)
	List(ctx context.Context, sortByStr, orderByStr string, includeDone bool) ([]*model.Todo, error)
	Update(ctx context.Context, idStr string, title, description *string, status, priority *int) (*model.Todo, error)
	Delete(ctx context.Context, idStr string) error
}

type todoUsecase struct {
	repo repository.TodoRepository
}

func NewTodoUsecase(repo repository.TodoRepository) TodoUsecase {
	return &todoUsecase{repo: repo}
}

func (u *todoUsecase) Create(
	ctx context.Context, title, description string, statusInt, priorityInt int,
) (*model.Todo, error) {
	if err := validateTitle(title); err != nil {
		return nil, utility.BadRequest("", err)
	}
	if err := validateDescription(description); err != nil {
		return nil, utility.BadRequest("", err)
	}

	status, err := parseStatus(statusInt)
	if err != nil {
		return nil, utility.BadRequest("", err)
	}
	priority, err := parsePriority(priorityInt)
	if err != nil {
		return nil, utility.BadRequest("", err)
	}

	newTodo := model.Todo{
		Title:       title,
		Description: description,
		Status:      status,
		Priority:    priority,
	}
	newID, err := u.repo.Create(ctx, newTodo)
	if err != nil {
		return nil, err
	}

	return u.repo.Get(ctx, newID)
}

func (u *todoUsecase) Get(ctx context.Context, idStr string) (*model.Todo, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, utility.BadRequest(fmt.Sprintf("id must be integer, but %s", idStr), err)
	}

	return u.repo.Get(ctx, id)
}

func (u *todoUsecase) List(ctx context.Context, sortByStr, orderByStr string, includeDone bool) ([]*model.Todo, error) {
	sortBy, err := model.ToSorter(sortByStr)
	if err != nil {
		return nil, utility.BadRequest("", err)
	}
	orderBy, err := model.ToOrder(orderByStr)
	if err != nil {
		return nil, utility.BadRequest("", err)
	}
	return u.repo.List(ctx, sortBy, orderBy, includeDone)
}

func (u *todoUsecase) Update(
	ctx context.Context, idStr string, title, description *string, statusIntP, priorityIntP *int,
) (*model.Todo, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, utility.BadRequest(fmt.Sprintf("id must be integer, but %s", idStr), err)
	}
	todo, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if title == nil && description == nil && statusIntP == nil && priorityIntP == nil {
		err := errors.New("no fields to be updated")
		return nil, utility.BadRequest("", err)
	}

	if title != nil {
		if err := validateTitle(*title); err != nil {
			return nil, utility.BadRequest("", err)
		}
		todo.Title = *title
	}

	if description != nil {
		if err := validateDescription(*description); err != nil {
			return nil, utility.BadRequest("", err)
		}
		todo.Description = *description
	}

	if statusIntP != nil {
		status, err := parseStatus(*statusIntP)
		if err != nil {
			return nil, utility.BadRequest("", err)
		}
		todo.Status = status
	}
	if priorityIntP != nil {
		priority, err := parsePriority(*priorityIntP)
		if err != nil {
			return nil, utility.BadRequest("", err)
		}
		todo.Priority = priority
	}

	if err := u.repo.Update(ctx, todo); err != nil {
		return nil, err
	}

	return u.repo.Get(ctx, id)
}

func (u *todoUsecase) Delete(ctx context.Context, idStr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utility.BadRequest(fmt.Sprintf("id must be integer, but %s", idStr), err)
	}

	return u.repo.Delete(ctx, id)
}
