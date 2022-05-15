package db

import (
	"fmt"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-txdb"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var once sync.Once

func GetTestDBConn(t *testing.T) *gorm.DB {
	t.Helper()

	var ev environmentVariables
	if err := envconfig.Process("", &ev); err != nil {
		t.Fatal(err)
	}

	once.Do(func() {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			ev.Host, ev.User, ev.Password, ev.Database, ev.Port,
		)
		txdb.Register("txdb", "postgres", dsn)
	})

	ret, err := gorm.Open(
		postgres.New(postgres.Config{
			DriverName: "txdb",
		}),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}
	return ret
}
