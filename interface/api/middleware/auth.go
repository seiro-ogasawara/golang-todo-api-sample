package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/repository"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/model"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility"
)

type AuthMiddleware interface {
	NewAuthentication() gin.HandlerFunc
}

type authMiddleware struct {
	repo repository.UserRepository
}

func NewAuthMiddleware(repo repository.UserRepository) AuthMiddleware {
	return &authMiddleware{repo: repo}
}

func (m *authMiddleware) NewAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		val := c.Request.Header.Get("authorization")
		pair := strings.SplitN(val, ":", 2)
		if len(pair) < 2 {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				model.ErrorResponse{ErrCode: http.StatusUnauthorized, Detail: "invalid authentication"},
			)
			return
		}

		authenticated, err := m.repo.Authenticate(c, pair[0], pair[1])
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				model.ErrorResponse{ErrCode: http.StatusInternalServerError, Detail: err.Error()},
			)
			return
		}
		if !authenticated {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				model.ErrorResponse{ErrCode: http.StatusUnauthorized, Detail: "uset not found or invalid password"},
			)
			return
		}

		c.Set(utility.UserIDKey, pair[0])
	}
}
