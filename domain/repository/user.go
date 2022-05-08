package repository

import "context"

type UserRepository interface {
	Authenticate(ctx context.Context, id, password string) (bool, error)
	Create(ctx context.Context, id, password string) error
}
