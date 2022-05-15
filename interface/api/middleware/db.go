package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/model"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility/config"
	"gorm.io/gorm"
)

type DBMiddleware struct {
	db *gorm.DB
}

func NewDBMiddleware(db *gorm.DB) *DBMiddleware {
	return &DBMiddleware{db}
}

func (m *DBMiddleware) NewTransaction() gin.HandlerFunc {
	if m == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		tx := m.db.Begin()
		if err := tx.Error; err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				model.ErrorResponse{ErrCode: http.StatusInternalServerError, Detail: "can't start transaction"},
			)
		}
		c.Set(config.DBKey, tx)

		defer func() {
			err := c.Errors.Last()
			if r := recover(); r != nil || err != nil {
				if rerr := tx.Rollback().Error; rerr != nil {
					log.Printf("failed to rollback: %v\n", rerr)
				}
				if r != nil {
					// re-throw error to make recoverty handler work
					panic(r)
				}
			} else {
				if cerr := tx.Commit().Error; cerr != nil {
					log.Printf("failed to commit: %v\n", cerr)
					_ = tx.Rollback()
				}
			}
		}()

		c.Next()
	}
}

func (m *DBMiddleware) NewDB() gin.HandlerFunc {
	if m == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		c.Set(config.DBKey, m.db)
		c.Next()
	}
}
