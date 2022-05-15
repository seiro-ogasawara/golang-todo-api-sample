package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/model"
	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/repository"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility/db"
	"gorm.io/gorm"
)

type databaseUserRepository struct {
}

func NewDatabaseUserRepository() repository.UserRepository {
	return &databaseUserRepository{}
}

func (r *databaseUserRepository) Authenticate(ctx context.Context, id, password string) (bool, error) {
	var u model.User
	if err := db.GetDBFromContext(ctx).Where("user_id = ? AND password = ?", id, password).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, utility.InternalServerError("can't find user from db", err)
	}
	return true, nil
}

func (r *databaseUserRepository) Create(ctx context.Context, id, password string) error {
	u := model.User{
		UserID:   id,
		Password: password,
	}
	if err := db.GetDBFromContext(ctx).Create(&u).Error; err != nil {
		pgErr, ok := err.(*pq.Error)
		if ok {
			if pgErr.Code.Name() == "unique_violation" {
				return utility.Conflict(fmt.Sprintf("user %s already exists", id), pgErr)
			}
		}
		return utility.InternalServerError("failed to create user", err)
	}
	return nil
}
