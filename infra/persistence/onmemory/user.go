package onmemory

import (
	"context"
	"sync"

	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/model"
	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/repository"
)

type onmemoryUserRepository struct {
	sync sync.Mutex
	data []model.User
}

func NewOnmemoryUserRepository() repository.UserRepository {
	users := make([]model.User, 0)
	return &onmemoryUserRepository{data: users}
}

func (r *onmemoryUserRepository) Authenticate(ctx context.Context, id, password string) (bool, error) {
	for _, u := range r.data {
		if id == u.UserID {
			if password == u.Password {
				return true, nil
			}
			return false, nil
		}
	}
	return false, nil
}

func (r *onmemoryUserRepository) Create(ctx context.Context, id, password string) error {
	r.sync.Lock()
	defer r.sync.Unlock()

	user := model.User{UserID: id, Password: password}
	r.data = append(r.data, user)

	return nil
}
