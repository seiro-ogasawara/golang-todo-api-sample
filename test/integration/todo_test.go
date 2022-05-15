package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/model"
	"github.com/seiro-ogasawara/golang-todo-api-sample/domain/repository"
	"github.com/seiro-ogasawara/golang-todo-api-sample/infra/persistence/onmemory"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/handler"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/middleware"
	"github.com/seiro-ogasawara/golang-todo-api-sample/usecase"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility/config"
	"github.com/seiro-ogasawara/golang-todo-api-sample/utility/db"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestTodoCreateWithOnmemoryRepository(t *testing.T) {
	router, db, userRepo := createRouterWithOnmemoryRepository(t)
	testTodoCreate(t, router, db, userRepo)
}

func TestTodoCreateWithDatabaseRepository(t *testing.T) {
	router, db, userRepo := createRouterWithDatabaseRepository(t)
	testTodoCreate(t, router, db, userRepo)
}

func testTodoCreate(t *testing.T, router *gin.Engine, db *gorm.DB, userRepo repository.UserRepository) {
	t.Helper()

	now := time.Now()
	ctx := getContext(t, db)

	_ = userRepo.Create(ctx, "userid", "password")

	cases := []struct {
		name           string
		auth           string
		body           handler.CreateTodoRequest
		expectStatus   int
		expectResponse handler.TodoResponse
	}{
		{
			name: "success, full",
			auth: "userid:password",
			body: handler.CreateTodoRequest{
				Title:       "title string",
				Description: "description string",
				Status:      int(model.StatusNotReady),
				Priority:    int(model.PriorityHigh),
			},
			expectStatus: http.StatusCreated,
			expectResponse: handler.TodoResponse{
				Title:       "title string",
				Description: "description string",
				Status:      int(model.StatusNotReady),
				Priority:    int(model.PriorityHigh),
			},
		},
		{
			name: "success, minimum",
			auth: "userid:password",
			body: handler.CreateTodoRequest{
				Title: "title string2",
			},
			expectStatus: http.StatusCreated,
			expectResponse: handler.TodoResponse{
				Title:       "title string2",
				Description: "",
				Status:      int(model.StatusNotReady),
				Priority:    int(model.PriorityMiddle),
			},
		},
		{
			name: "fail, empty title",
			auth: "userid:password",
			body: handler.CreateTodoRequest{
				Title: "",
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, title too long",
			auth: "userid:password",
			body: handler.CreateTodoRequest{
				Title: fmt.Sprintf("%051s", "title string"),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, description too long",
			auth: "userid:password",
			body: handler.CreateTodoRequest{
				Title:       "title string",
				Description: fmt.Sprintf("%0501s", "description string"),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, unknown status",
			auth: "userid:password",
			body: handler.CreateTodoRequest{
				Title:       "title string",
				Description: "description string",
				Status:      int(model.StatusDone) + 1,
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, unknown priority",
			auth: "userid:password",
			body: handler.CreateTodoRequest{
				Title:       "title string",
				Description: "description string",
				Priority:    int(model.PriorityLow) + 1,
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, unauthorized",
			auth: "userid:invalidpassword",
			body: handler.CreateTodoRequest{
				Title: "title string2",
			},
			expectStatus: http.StatusUnauthorized,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			b, _ := json.Marshal(c.body)
			body := ioutil.NopCloser(bytes.NewBuffer(b))
			req, _ := http.NewRequest("POST", "/todos", body)
			req.Header.Set("Authorization", c.auth)
			router.ServeHTTP(w, req)

			assert.Equal(t, c.expectStatus, w.Code, w.Body.String())
			if c.expectStatus != http.StatusCreated {
				return
			}

			var actual handler.TodoResponse
			if err := json.Unmarshal(w.Body.Bytes(), &actual); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, c.expectResponse.Title, actual.Title)
			assert.Equal(t, c.expectResponse.Description, actual.Description)
			assert.Equal(t, c.expectResponse.Status, actual.Status)
			assert.Equal(t, c.expectResponse.Priority, actual.Priority)
			createdAt, err := time.Parse(time.RFC3339, actual.CreatedAt)
			if err != nil {
				t.Fatal(err)
			}
			assert.True(t, now.UnixNano() < createdAt.UnixNano(), createdAt.Format(time.RFC3339Nano))
			updatedAt, err := time.Parse(time.RFC3339, actual.UpdatedAt)
			if err != nil {
				t.Fatal(err)
			}
			assert.True(t, now.UnixNano() < updatedAt.UnixNano(), updatedAt.Format(time.RFC3339Nano))
			assert.Equal(t, createdAt.UnixNano(), updatedAt.UnixNano())
		})
	}
}

func TestTodoGetWithOnmemoryRepository(t *testing.T) {
	router, db, userRepo := createRouterWithOnmemoryRepository(t)
	testTodoGet(t, router, db, userRepo)
}

func TestTodoGetWithDatabaseRepository(t *testing.T) {
	router, db, userRepo := createRouterWithDatabaseRepository(t)
	testTodoGet(t, router, db, userRepo)
}

func testTodoGet(t *testing.T, router *gin.Engine, db *gorm.DB, userRepo repository.UserRepository) {
	t.Helper()

	_ = userRepo.Create(getContext(t, db), "userid", "password")
	_ = userRepo.Create(getContext(t, db), "userid2", "password2")

	// prepare todo
	var existingTodo handler.TodoResponse
	{
		reqBody := handler.CreateTodoRequest{
			Title:       "title string",
			Description: "description string",
			Status:      int(model.StatusNotReady),
			Priority:    int(model.PriorityHigh),
		}
		existingTodo = createTodo(t, router, "userid:password", reqBody)
	}
	id, err := strconv.Atoi(existingTodo.ID)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name           string
		id             int
		auth           string
		expectStatus   int
		expectResponse handler.TodoResponse
	}{
		{
			name:           "success",
			id:             id,
			auth:           "userid:password",
			expectStatus:   http.StatusOK,
			expectResponse: existingTodo,
		},
		{
			name:         "not found",
			auth:         "userid:password",
			id:           id + 1,
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "not found(others todo)",
			auth:         "userid2:password2",
			id:           id,
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "unauthorized",
			auth:         "userid:invalidpassword",
			id:           id,
			expectStatus: http.StatusUnauthorized,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/todos/%d", c.id), nil)
			req.Header.Set("Authorization", c.auth)
			router.ServeHTTP(w, req)

			assert.Equal(t, c.expectStatus, w.Code, w.Body.String())
			if c.expectStatus != http.StatusCreated {
				return
			}

			var actual handler.TodoResponse
			if err := json.Unmarshal(w.Body.Bytes(), &actual); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, c.expectResponse.Title, actual.Title)
			assert.Equal(t, c.expectResponse.Description, actual.Description)
			assert.Equal(t, c.expectResponse.Status, actual.Status)
			assert.Equal(t, c.expectResponse.Priority, actual.Priority)
			assert.Equal(t, c.expectResponse.CreatedAt, actual.CreatedAt)
			assert.Equal(t, c.expectResponse.UpdatedAt, actual.UpdatedAt)
		})
	}
}

func TestTodoListWithOnmemoryRepository(t *testing.T) {
	router, db, userRepo := createRouterWithOnmemoryRepository(t)
	testTodoList(t, router, db, userRepo)
}

func TestTodoListWithDatabaseRepository(t *testing.T) {
	router, db, userRepo := createRouterWithDatabaseRepository(t)
	testTodoList(t, router, db, userRepo)
}

func testTodoList(t *testing.T, router *gin.Engine, db *gorm.DB, userRepo repository.UserRepository) {
	t.Helper()

	_ = userRepo.Create(getContext(t, db), "userid", "password")
	_ = userRepo.Create(getContext(t, db), "userid2", "password2")

	// prepare todo
	userTodos := make([]handler.TodoResponse, 0)
	tParams := []handler.CreateTodoRequest{
		{Title: "t11", Description: "d11", Status: int(model.StatusNotReady), Priority: int(model.PriorityHigh)},
		{Title: "t12", Description: "d12", Status: int(model.StatusReady), Priority: int(model.PriorityMiddle)},
		{Title: "t13", Description: "d13", Status: int(model.StatusDoing), Priority: int(model.PriorityLow)},
		{Title: "t14", Description: "d14", Status: int(model.StatusDone), Priority: int(model.PriorityHigh)},
		{Title: "t15", Description: "d15", Status: int(model.StatusNotReady), Priority: int(model.PriorityMiddle)},
	}
	for _, tp := range tParams {
		td := createTodo(t, router, "userid:password", tp)
		userTodos = append(userTodos, td)
	}

	tParams2 := []handler.CreateTodoRequest{
		{Title: "t21", Description: "d21", Status: int(model.StatusNotReady), Priority: int(model.PriorityHigh)},
	}
	for _, tp := range tParams2 {
		_ = createTodo(t, router, "userid2:password2", tp)
	}

	cases := []struct {
		name         string
		auth         string
		param        handler.ListTodoRequest
		expectStatus int
		expects      handler.ListTodoResponse
	}{
		{
			name:         "ok, no param", // default: order by id asc, not include done todo.
			auth:         "userid:password",
			param:        handler.ListTodoRequest{},
			expectStatus: http.StatusOK,
			expects: handler.ListTodoResponse{
				Entries: []handler.TodoResponse{userTodos[0], userTodos[1], userTodos[2], userTodos[4]},
			},
		},
		{
			name: "ok, param: sortby id, order by desc",
			auth: "userid:password",
			param: handler.ListTodoRequest{
				SortBy:  "id",
				OrderBy: "desc",
			},
			expectStatus: http.StatusOK,
			expects: handler.ListTodoResponse{
				Entries: []handler.TodoResponse{userTodos[4], userTodos[2], userTodos[1], userTodos[0]},
			},
		},
		{
			name: "ok, param: sortby priority, order by asc",
			auth: "userid:password",
			param: handler.ListTodoRequest{
				SortBy:  "Priority",
				OrderBy: "aSC",
			},
			expectStatus: http.StatusOK,
			expects: handler.ListTodoResponse{
				Entries: []handler.TodoResponse{userTodos[0], userTodos[1], userTodos[4], userTodos[2]},
			},
		},
		{
			name: "ok, param: sortby priority, order by desc, including done todos",
			auth: "userid:password",
			param: handler.ListTodoRequest{
				SortBy:      "PRIORITY",
				OrderBy:     "Desc",
				IncludeDone: true,
			},
			expectStatus: http.StatusOK,
			expects: handler.ListTodoResponse{
				Entries: []handler.TodoResponse{userTodos[2], userTodos[1], userTodos[4], userTodos[0], userTodos[3]},
			},
		},
		{
			name: "ng, invalid sortby param",
			auth: "userid:password",
			param: handler.ListTodoRequest{
				SortBy: "invalid",
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "ng, invalid orderby param",
			auth: "userid:password",
			param: handler.ListTodoRequest{
				OrderBy: "invalid",
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "ng, unauthorized",
			auth:         "userid:invalidpassword",
			param:        handler.ListTodoRequest{},
			expectStatus: http.StatusUnauthorized,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			url := fmt.Sprintf("/todos?")
			if c.param.SortBy != "" {
				url = fmt.Sprintf("%ssortby=%s&", url, c.param.SortBy)
			}
			if c.param.OrderBy != "" {
				url = fmt.Sprintf("%sorderby=%s&", url, c.param.OrderBy)
			}
			if c.param.IncludeDone {
				url = fmt.Sprintf("%sincludeDone=true", url)
			}
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Authorization", c.auth)
			router.ServeHTTP(w, req)

			assert.Equal(t, c.expectStatus, w.Code, w.Body.String())
			if c.expectStatus != http.StatusOK {
				return
			}

			var actuals handler.ListTodoResponse
			if err := json.Unmarshal(w.Body.Bytes(), &actuals); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, len(c.expects.Entries), len(actuals.Entries))
			for i := 0; i < len(c.expects.Entries); i++ {
				expect := c.expects.Entries[i]
				actual := actuals.Entries[i]
				assert.Equal(t, expect, actual, i)
			}
		})
	}
}

func TestTodoUpdateWithOnmemoryRepository(t *testing.T) {
	router, db, userRepo := createRouterWithOnmemoryRepository(t)
	testTodoUpdate(t, router, db, userRepo)
}

func TestTodoUpdateWithDatabaseRepository(t *testing.T) {
	router, db, userRepo := createRouterWithDatabaseRepository(t)
	testTodoUpdate(t, router, db, userRepo)
}

func testTodoUpdate(t *testing.T, router *gin.Engine, db *gorm.DB, userRepo repository.UserRepository) {
	t.Helper()

	_ = userRepo.Create(getContext(t, db), "userid", "password")
	_ = userRepo.Create(getContext(t, db), "userid2", "password2")

	// prepare todo
	var existingTodo handler.TodoResponse
	{
		reqBody := handler.CreateTodoRequest{
			Title:       "title string",
			Description: "description string",
			Status:      int(model.StatusNotReady),
			Priority:    int(model.PriorityHigh),
		}
		existingTodo = createTodo(t, router, "userid:password", reqBody)
	}
	id, err := strconv.Atoi(existingTodo.ID)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name         string
		id           int
		auth         string
		body         handler.UpdateTodoRequest
		expectStatus int
		expect       handler.TodoResponse
	}{
		{
			name:         "unauthorized",
			id:           id,
			auth:         "userid:invalidpassword",
			body:         handler.UpdateTodoRequest{},
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "not found",
			id:           id + 1,
			auth:         "userid:password",
			body:         handler.UpdateTodoRequest{},
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "not found(others todo)",
			id:           id,
			auth:         "userid2:password2",
			body:         handler.UpdateTodoRequest{},
			expectStatus: http.StatusNotFound,
		},
		{
			name: "fail, title too long",
			id:   id,
			auth: "userid:password",
			body: handler.UpdateTodoRequest{
				Title: ptr(fmt.Sprintf("%051s", "title string")),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, description too long",
			id:   id,
			auth: "userid:password",
			body: handler.UpdateTodoRequest{
				Description: ptr(fmt.Sprintf("%0501s", "description string")),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, invalid status",
			id:   id,
			auth: "userid:password",
			body: handler.UpdateTodoRequest{
				Status: ptr(int(model.StatusDone + 1)),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, invalid priority",
			id:   id,
			auth: "userid:password",
			body: handler.UpdateTodoRequest{
				Priority: ptr(int(model.PriorityLow + 1)),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "success, update title",
			id:   id,
			auth: "userid:password",
			body: handler.UpdateTodoRequest{
				Title: ptr("updated title"),
			},
			expectStatus: http.StatusOK,
			expect: handler.TodoResponse{
				Title:       "updated title",
				Description: existingTodo.Description,
				Status:      existingTodo.Status,
				Priority:    existingTodo.Priority,
				CreatedAt:   existingTodo.CreatedAt,
			},
		},
		{
			name: "success, update full",
			id:   id,
			auth: "userid:password",
			body: handler.UpdateTodoRequest{
				Title:       ptr("updated title2"),
				Description: ptr("updated description"),
				Status:      ptr(int(model.StatusDone)),
				Priority:    ptr(int(model.PriorityMiddle)),
			},
			expectStatus: http.StatusOK,
			expect: handler.TodoResponse{
				Title:       "updated title2",
				Description: "updated description",
				Status:      int(model.StatusDone),
				Priority:    int(model.PriorityMiddle),
				CreatedAt:   existingTodo.CreatedAt,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			b, _ := json.Marshal(c.body)
			body := ioutil.NopCloser(bytes.NewBuffer(b))
			req, _ := http.NewRequest("PATCH", fmt.Sprintf("/todos/%d", c.id), body)
			req.Header.Set("Authorization", c.auth)
			router.ServeHTTP(w, req)

			assert.Equal(t, c.expectStatus, w.Code, w.Body.String())
			if c.expectStatus != http.StatusOK {
				return
			}

			var actual handler.TodoResponse
			if err := json.Unmarshal(w.Body.Bytes(), &actual); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, c.expect.Title, actual.Title)
			assert.Equal(t, c.expect.Description, actual.Description)
			assert.Equal(t, c.expect.Status, actual.Status)
			assert.Equal(t, c.expect.Priority, actual.Priority)
			assert.Equal(t, c.expect.CreatedAt, actual.CreatedAt)
			createdAt, err := time.Parse(time.RFC3339Nano, actual.CreatedAt)
			if err != nil {
				t.Fatal(err)
			}
			updatedAt, err := time.Parse(time.RFC3339Nano, actual.UpdatedAt)
			if err != nil {
				t.Fatal(err)
			}
			assert.True(t, updatedAt.After(createdAt))
		})
	}
}

func TestTodoDeleteWithOnmemoryRepository(t *testing.T) {
	router, db, userRepo := createRouterWithOnmemoryRepository(t)
	testTodoDelete(t, router, db, userRepo)
}

func TestTodoDeleteWithDatabaseRepository(t *testing.T) {
	router, db, userRepo := createRouterWithDatabaseRepository(t)
	testTodoDelete(t, router, db, userRepo)
}

func testTodoDelete(t *testing.T, router *gin.Engine, db *gorm.DB, userRepo repository.UserRepository) {
	t.Helper()

	_ = userRepo.Create(getContext(t, db), "userid", "password")
	_ = userRepo.Create(getContext(t, db), "userid2", "password2")

	// prepare todo
	var existingTodo handler.TodoResponse
	{
		reqBody := handler.CreateTodoRequest{
			Title:       "title string",
			Description: "description string",
			Status:      int(model.StatusNotReady),
			Priority:    int(model.PriorityHigh),
		}
		existingTodo = createTodo(t, router, "userid:password", reqBody)
	}
	id, err := strconv.Atoi(existingTodo.ID)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name         string
		id           int
		auth         string
		expectStatus int
	}{
		{
			name:         "unauthorized",
			id:           id,
			auth:         "userid:invalidpassword",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "not found",
			id:           id + 1,
			auth:         "userid:password",
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "not found(others todo)",
			id:           id,
			auth:         "userid2:password2",
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "success",
			id:           id,
			auth:         "userid:password",
			expectStatus: http.StatusOK,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/todos/%d", c.id), nil)
			req.Header.Set("Authorization", c.auth)
			router.ServeHTTP(w, req)

			assert.Equal(t, c.expectStatus, w.Code, w.Body.String())
			if c.expectStatus != http.StatusCreated {
				return
			}

			// verify that it has been deleted
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest("GET", fmt.Sprintf("/todos/%d", c.id), nil)
			req.Header.Set("Authorization", c.auth)
			router.ServeHTTP(w2, req2)
			assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		})
	}
}

func createRouterWithDatabaseRepository(t *testing.T) (*gin.Engine, *gorm.DB, repository.UserRepository) {
	db := db.GetTestDBConn(t)
	todoRepo := onmemory.NewOnmemoryTodoRepository()
	userRepo := onmemory.NewOnmemoryUserRepository()
	usecase := usecase.NewTodoUsecase(todoRepo)
	handler := handler.NewTodoHandler(usecase)
	authMiddleware := middleware.NewAuthMiddleware(userRepo)
	dbMiddleware := middleware.NewDBMiddleware(db)
	return api.Route(authMiddleware, dbMiddleware, handler), db, userRepo
}

func createRouterWithOnmemoryRepository(t *testing.T) (*gin.Engine, *gorm.DB, repository.UserRepository) {
	todoRepo := onmemory.NewOnmemoryTodoRepository()
	userRepo := onmemory.NewOnmemoryUserRepository()
	usecase := usecase.NewTodoUsecase(todoRepo)
	handler := handler.NewTodoHandler(usecase)
	authMiddleware := middleware.NewAuthMiddleware(userRepo)
	return api.Route(authMiddleware, nil, handler), nil, userRepo
}

func getContext(t *testing.T, db *gorm.DB) context.Context {
	t.Helper()

	return context.WithValue(context.TODO(), config.DBKey, db)
}

func createTodo(
	t *testing.T, router *gin.Engine, auth string, reqBody handler.CreateTodoRequest,
) handler.TodoResponse {
	t.Helper()

	w := httptest.NewRecorder()
	b, _ := json.Marshal(reqBody)
	body := ioutil.NopCloser(bytes.NewBuffer(b))
	req, _ := http.NewRequest("POST", "/todos", body)
	req.Header.Set("Authorization", auth)
	router.ServeHTTP(w, req)

	var td handler.TodoResponse
	assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	if err := json.Unmarshal(w.Body.Bytes(), &td); err != nil {
		t.Fatal(err)
	}
	return td
}

// -----
// utilities

type pointable interface {
	int | string // NOTE: float, bool, and a few other things, but I'll ignore them for now.
}

func ptr[T pointable](v T) *T {
	return &v
}
