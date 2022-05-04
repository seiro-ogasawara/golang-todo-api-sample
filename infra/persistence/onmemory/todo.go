package onmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	linq "github.com/ahmetb/go-linq/v3"

	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/model"
	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/repository"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility"
)

type onmemoryTodoRepository struct {
	sync sync.Mutex
	id   int
	data []model.Todo
}

func NewOnmemoryTodoRepository() repository.TodoRepository {
	todos := make([]model.Todo, 0)
	return &onmemoryTodoRepository{data: todos}
}

func (r *onmemoryTodoRepository) Create(ctx context.Context, todo model.Todo) (int, error) {
	r.sync.Lock()
	defer r.sync.Unlock()

	now := time.Now()
	r.id += 1
	todo.ID = r.id
	todo.CreatedAt = now
	todo.UpdatedAt = now
	r.data = append(r.data, todo)
	return todo.ID, nil
}

func (r *onmemoryTodoRepository) Get(ctx context.Context, id int) (*model.Todo, error) {
	for i := 0; i < len(r.data); i++ {
		todo := r.data[i]
		if todo.ID == id {
			ret := todo
			return &ret, nil
		}
	}
	return nil, utility.NotFound("", fmt.Errorf("todo with id %d is not found", id))
}

func (r *onmemoryTodoRepository) List(
	ctx context.Context, sortBy model.Sorter, orderBy model.Order, includeDone bool,
) ([]*model.Todo, error) {
	sortedTodos := []model.Todo{}
	query := linq.From(r.data)
	if !includeDone {
		// exclude finished todo
		query = query.WhereT(
			func(t model.Todo) bool {
				return t.Status != model.StatusDone
			},
		)
	}
	query.SortT(
		func(t1, t2 model.Todo) bool {
			if sortBy == model.SortByID && orderBy == model.OrderByASC {
				return t1.ID < t2.ID
			} else if sortBy == model.SortByID && orderBy == model.OrderByDESC {
				return t1.ID > t2.ID
			} else if sortBy == model.SortByPriority && orderBy == model.OrderByASC {
				return t1.Priority < t2.Priority
			} else {
				return t1.Priority > t2.Priority
			}
		},
	).ToSlice(&sortedTodos)

	ret := make([]*model.Todo, 0, len(sortedTodos))
	for i := 0; i < len(sortedTodos); i++ {
		ret = append(ret, &sortedTodos[i])
	}
	return ret, nil
}

func (r *onmemoryTodoRepository) Update(ctx context.Context, todo *model.Todo) error {
	r.sync.Lock()
	defer r.sync.Unlock()

	processed := false
	for i := 0; i < len(r.data); i++ {
		if r.data[i].ID == todo.ID {
			r.data[i] = *todo
			r.data[i].UpdatedAt = time.Now()
			processed = true
			break
		}
	}
	if !processed {
		return utility.NotFound("", fmt.Errorf("todo with id %d is not found", todo.ID))
	}
	return nil
}

func (r *onmemoryTodoRepository) Delete(ctx context.Context, id int) error {
	targetNum := 0
	found := false
	for i := 0; i < len(r.data); i++ {
		if r.data[i].ID == id {
			targetNum = i
			found = true
			break
		}
	}
	if !found {
		return utility.NotFound("", fmt.Errorf("todo with id %d is not found", id))
	}

	r.data = r.data[:targetNum+copy(r.data[targetNum:], r.data[targetNum+1:])]
	return nil
}
