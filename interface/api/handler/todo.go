package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/model"
	servermodel "github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/model"
	"github.com/seiro-ogasawara/golang-todo-api-sample/usecase"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility/config"
)

// TodoHandler is API interface of Todo service.
type TodoHandler interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

// todoHandler is a structure that implements TodoHandler.
type todoHandler struct {
	u usecase.TodoUsecase
}

func NewTodoHandler(u usecase.TodoUsecase) TodoHandler {
	return &todoHandler{u: u}
}

// CreateTodoRequest is the structure representation of the request body of `POST /todos`.
type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Status      int    `json:"status,omitempty"`   // 1: Not Ready, 2: Ready, 3: Doing, 4: Done
	Priority    int    `json:"priority,omitempty"` // 1: High, 2: Middle, 3: Low
}

// TodoResponse is the structure representation of the response of Todo information.
type TodoResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      int    `json:"status"`   // 1: Not Ready, 2: Ready, 3: Doing, 4: Done
	Priority    int    `json:"priority"` // 1: High, 2: Middle, 3: Low
	CreatedAt   string `json:"createAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func buildTodoResponse(todo *model.Todo) TodoResponse {
	return TodoResponse{
		ID:          strconv.Itoa(todo.ID),
		Title:       todo.Title,
		Description: todo.Description,
		Status:      int(todo.Status),
		Priority:    int(todo.Priority),
		CreatedAt:   todo.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   todo.UpdatedAt.Format(time.RFC3339Nano),
	}
}

// Create processes the request of `POST /todos`.
func (h *todoHandler) Create(c *gin.Context) {
	userID := c.GetString(config.UserIDKey)

	json := CreateTodoRequest{
		Status:   int(model.StatusNotReady),
		Priority: int(model.PriorityMiddle),
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			servermodel.ErrorResponse{ErrCode: http.StatusBadRequest, Detail: err.Error()},
		)
		return
	}

	newTodo, err := h.u.Create(c, userID, json.Title, json.Description, json.Status, json.Priority)
	if err != nil {
		sendErrorResponse(c, err)
		return
	}
	c.JSON(http.StatusCreated, buildTodoResponse(newTodo))
}

// Get processes the request of `GET /todos/:id`.
func (h *todoHandler) Get(c *gin.Context) {
	userID := c.GetString(config.UserIDKey)
	todoID := c.Param("id")

	todo, err := h.u.Get(c, userID, todoID)
	if err != nil {
		sendErrorResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, buildTodoResponse(todo))
}

// ListTodoRequest is the structure representation of the request body of `GET /todos`.
type ListTodoRequest struct {
	SortBy      string `form:"sortby"`  // "id" or "priority"
	OrderBy     string `form:"orderby"` // "asc" or "desc"
	IncludeDone bool   `form:"includeDone"`
}

// ListTodoResponse is the structure representation of the response body of `GET /todos`.
type ListTodoResponse struct {
	Entries []TodoResponse
}

// List processes the request of `GET /todos`.
func (h todoHandler) List(c *gin.Context) {
	userID := c.GetString(config.UserIDKey)

	query := ListTodoRequest{
		SortBy:      string(model.SortByID),
		OrderBy:     string(model.OrderByASC),
		IncludeDone: false,
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			servermodel.ErrorResponse{ErrCode: http.StatusBadRequest, Detail: err.Error()},
		)
		return
	}

	todos, err := h.u.List(c, userID, query.SortBy, query.OrderBy, query.IncludeDone)
	if err != nil {
		sendErrorResponse(c, err)
		return
	}
	res := make([]TodoResponse, 0, len(todos))
	for _, todo := range todos {
		t := buildTodoResponse(todo)
		res = append(res, t)
	}
	c.JSON(http.StatusOK, ListTodoResponse{res})
}

// UpdateTodoRequest is the structure representation of the request body of `PATCH /todos/:id`.
type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *int    `json:"status,omitempty"`
	Priority    *int    `json:"priority,omitempty"`
}

// Update processes the request of `PATCH /todos/:id`.
func (h todoHandler) Update(c *gin.Context) {
	userID := c.GetString(config.UserIDKey)
	todoID := c.Param("id")

	json := UpdateTodoRequest{}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			servermodel.ErrorResponse{ErrCode: http.StatusBadRequest, Detail: err.Error()},
		)
		return
	}
	todo, err := h.u.Update(c, userID, todoID, json.Title, json.Description, json.Status, json.Priority)
	if err != nil {
		sendErrorResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, buildTodoResponse(todo))
}

// Delete processes the request of `DELETE /todos/:id`.
func (h todoHandler) Delete(c *gin.Context) {
	userID := c.GetString(config.UserIDKey)
	todoID := c.Param("id")

	if err := h.u.Delete(c, userID, todoID); err != nil {
		sendErrorResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, servermodel.MessageResponse{Message: fmt.Sprintf("todo %s is deleted", todoID)})
}

func sendErrorResponse(c *gin.Context, err error) {
	var httpErr *utility.HTTPError
	if errors.As(err, &httpErr) {
		c.AbortWithStatusJSON(
			httpErr.ErrCode(),
			servermodel.ErrorResponse{ErrCode: httpErr.ErrCode(), Detail: httpErr.Error()},
		)
		return
	}
	c.AbortWithStatusJSON(
		http.StatusInternalServerError,
		servermodel.ErrorResponse{ErrCode: http.StatusInternalServerError, Detail: err.Error()},
	)
}
