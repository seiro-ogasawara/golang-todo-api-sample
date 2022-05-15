package db

import (
	"context"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type environmentVariables struct {
	User     string `envconfig:"DB_USER" default:"postgres"`
	Password string `envconfig:"DB_PASSWORD" default:"postgres"`
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     int    `envconfig:"DB_PORT" default:"5432"`
	Database string `envconfig:"DB_NAME" default:"todo_app"`
}

func GetDBFromEnvironmentVariables() (*gorm.DB, error) {
	var ev environmentVariables
	if err := envconfig.Process("", &ev); err != nil {
		return nil, err
	}
	return newDB(ev)
}

func newDB(ev environmentVariables) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=false",
		ev.Host, ev.Port, ev.User, ev.Password, ev.Database,
	)

	return gorm.Open(
		postgres.New(
			postgres.Config{
				DriverName: "postgres",
				DSN:        dsn,
			},
		),
		&gorm.Config{},
	)
}

func GetDBFromContext(ctx context.Context) *gorm.DB {
	v := ctx.Value(config.DBKey)
	if v == nil {
		panic("can't get db object from context")
	}
	ret, ok := v.(*gorm.DB)
	if !ok {
		panic("can't get db object from context")
	}
	return ret
}
