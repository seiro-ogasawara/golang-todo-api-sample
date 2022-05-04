package integration

import (
	"bytes"
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
	"github.com/seiro-ogasawara/golang-todo-api-sample/infra/persistence/onmemory"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api"
	"github.com/seiro-ogasawara/golang-todo-api-sample/interface/api/handler"
	"github.com/seiro-ogasawara/golang-todo-api-sample/usecase"
	"github.com/stretchr/testify/assert"
)

func TestTodoCreate(t *testing.T) {
	now := time.Now()
	router := createRouterWithOnmemoryRepository(t)

	cases := []struct {
		name           string
		body           handler.CreateTodoRequest
		expectStatus   int
		expectResponse handler.TodoResponse
	}{
		{
			name: "success, full",
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
			body: handler.CreateTodoRequest{
				Title: "",
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, title too long",
			body: handler.CreateTodoRequest{
				Title: fmt.Sprintf("%051s", "title string"),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, description too long",
			body: handler.CreateTodoRequest{
				Title:       "title string",
				Description: fmt.Sprintf("%0501s", "description string"),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, unknown status",
			body: handler.CreateTodoRequest{
				Title:       "title string",
				Description: "description string",
				Status:      int(model.StatusDone) + 1,
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, unknown priority",
			body: handler.CreateTodoRequest{
				Title:       "title string",
				Description: "description string",
				Priority:    int(model.PriorityLow) + 1,
			},
			expectStatus: http.StatusBadRequest,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			b, _ := json.Marshal(c.body)
			body := ioutil.NopCloser(bytes.NewBuffer(b))
			req, _ := http.NewRequest("POST", "/todos", body)
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

func TestTodoGet(t *testing.T) {
	router := createRouterWithOnmemoryRepository(t)

	// prepare todo
	var existingTodo handler.TodoResponse
	{
		reqBody := handler.CreateTodoRequest{
			Title:       "title string",
			Description: "description string",
			Status:      int(model.StatusNotReady),
			Priority:    int(model.PriorityHigh),
		}
		w := httptest.NewRecorder()
		b, _ := json.Marshal(reqBody)
		body := ioutil.NopCloser(bytes.NewBuffer(b))
		req, _ := http.NewRequest("POST", "/todos", body)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		if err := json.Unmarshal(w.Body.Bytes(), &existingTodo); err != nil {
			t.Fatal(err)
		}
	}
	id, err := strconv.Atoi(existingTodo.ID)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name           string
		id             int
		expectStatus   int
		expectResponse handler.TodoResponse
	}{
		{
			name:           "success",
			id:             id,
			expectStatus:   http.StatusOK,
			expectResponse: existingTodo,
		},
		{
			name:         "not found",
			id:           id + 1,
			expectStatus: http.StatusNotFound,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/todos/%d", c.id), nil)
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

func TestTodoList(t *testing.T) {
	router := createRouterWithOnmemoryRepository(t)

	// prepare todo
	todos := make([]handler.TodoResponse, 0)
	tParams := []handler.CreateTodoRequest{
		{Title: "t1", Description: "d1", Status: int(model.StatusNotReady), Priority: int(model.PriorityHigh)},
		{Title: "t2", Description: "d2", Status: int(model.StatusReady), Priority: int(model.PriorityMiddle)},
		{Title: "t3", Description: "d3", Status: int(model.StatusDoing), Priority: int(model.PriorityLow)},
		{Title: "t4", Description: "d4", Status: int(model.StatusDone), Priority: int(model.PriorityHigh)},
		{Title: "t5", Description: "d5", Status: int(model.StatusNotReady), Priority: int(model.PriorityMiddle)},
	}
	for _, tp := range tParams {
		reqBody := tp
		w := httptest.NewRecorder()
		b, _ := json.Marshal(reqBody)
		body := ioutil.NopCloser(bytes.NewBuffer(b))
		req, _ := http.NewRequest("POST", "/todos", body)
		router.ServeHTTP(w, req)

		var td handler.TodoResponse
		assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		if err := json.Unmarshal(w.Body.Bytes(), &td); err != nil {
			t.Fatal(err)
		}
		todos = append(todos, td)
	}

	cases := []struct {
		name         string
		param        handler.ListTodoRequest
		expectStatus int
		expects      handler.ListTodoResponse
	}{
		{
			name:         "ok, no param", // default: order by id asc, not include done todo.
			param:        handler.ListTodoRequest{},
			expectStatus: http.StatusOK,
			expects: handler.ListTodoResponse{
				Entries: []handler.TodoResponse{todos[0], todos[1], todos[2], todos[4]},
			},
		},
		{
			name: "ok, param: sortby id, order by desc",
			param: handler.ListTodoRequest{
				SortBy:  "id",
				OrderBy: "desc",
			},
			expectStatus: http.StatusOK,
			expects: handler.ListTodoResponse{
				Entries: []handler.TodoResponse{todos[4], todos[2], todos[1], todos[0]},
			},
		},
		{
			name: "ok, param: sortby priority, order by asc",
			param: handler.ListTodoRequest{
				SortBy:  "Priority",
				OrderBy: "aSC",
			},
			expectStatus: http.StatusOK,
			expects: handler.ListTodoResponse{
				Entries: []handler.TodoResponse{todos[0], todos[1], todos[4], todos[2]},
			},
		},
		{
			name: "ok, param: sortby priority, order by desc, including done todos",
			param: handler.ListTodoRequest{
				SortBy:      "PRIORITY",
				OrderBy:     "Desc",
				IncludeDone: true,
			},
			expectStatus: http.StatusOK,
			expects: handler.ListTodoResponse{
				Entries: []handler.TodoResponse{todos[2], todos[1], todos[4], todos[0], todos[3]},
			},
		},
		{
			name: "ng, invalid sortby param",
			param: handler.ListTodoRequest{
				SortBy: "invalid",
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "ng, invalid orderby param",
			param: handler.ListTodoRequest{
				OrderBy: "invalid",
			},
			expectStatus: http.StatusBadRequest,
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

func TestTodoUpdate(t *testing.T) {
	router := createRouterWithOnmemoryRepository(t)

	// prepare todo
	var existingTodo handler.TodoResponse
	{
		reqBody := handler.CreateTodoRequest{
			Title:       "title string",
			Description: "description string",
			Status:      int(model.StatusNotReady),
			Priority:    int(model.PriorityHigh),
		}
		w := httptest.NewRecorder()
		b, _ := json.Marshal(reqBody)
		body := ioutil.NopCloser(bytes.NewBuffer(b))
		req, _ := http.NewRequest("POST", "/todos", body)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		if err := json.Unmarshal(w.Body.Bytes(), &existingTodo); err != nil {
			t.Fatal(err)
		}
	}
	id, err := strconv.Atoi(existingTodo.ID)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name         string
		id           int
		body         handler.UpdateTodoRequest
		expectStatus int
		expect       handler.TodoResponse
	}{
		{
			name:         "not found",
			id:           id + 1,
			body:         handler.UpdateTodoRequest{},
			expectStatus: http.StatusNotFound,
		},
		{
			name: "fail, title too long",
			id:   id,
			body: handler.UpdateTodoRequest{
				Title: ptr(fmt.Sprintf("%051s", "title string")),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, description too long",
			id:   id,
			body: handler.UpdateTodoRequest{
				Description: ptr(fmt.Sprintf("%0501s", "description string")),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, invalid status",
			id:   id,
			body: handler.UpdateTodoRequest{
				Status: ptr(int(model.StatusDone + 1)),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "fail, invalid priority",
			id:   id,
			body: handler.UpdateTodoRequest{
				Priority: ptr(int(model.PriorityLow + 1)),
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "success, update title",
			id:   id,
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

func TestTodoDelete(t *testing.T) {
	router := createRouterWithOnmemoryRepository(t)

	// prepare todo
	var existingTodo handler.TodoResponse
	{
		reqBody := handler.CreateTodoRequest{
			Title:       "title string",
			Description: "description string",
			Status:      int(model.StatusNotReady),
			Priority:    int(model.PriorityHigh),
		}
		w := httptest.NewRecorder()
		b, _ := json.Marshal(reqBody)
		body := ioutil.NopCloser(bytes.NewBuffer(b))
		req, _ := http.NewRequest("POST", "/todos", body)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		if err := json.Unmarshal(w.Body.Bytes(), &existingTodo); err != nil {
			t.Fatal(err)
		}
	}
	id, err := strconv.Atoi(existingTodo.ID)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name         string
		id           int
		expectStatus int
	}{
		{
			name:         "not found",
			id:           id + 1,
			expectStatus: http.StatusNotFound,
		},
		{
			name:         "success",
			id:           id,
			expectStatus: http.StatusOK,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/todos/%d", c.id), nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, c.expectStatus, w.Code, w.Body.String())
			if c.expectStatus != http.StatusCreated {
				return
			}

			// verify that it has been deleted
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest("GET", fmt.Sprintf("/todos/%d", c.id), nil)
			router.ServeHTTP(w2, req2)
			assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		})
	}
}

func createRouterWithOnmemoryRepository(t *testing.T) *gin.Engine {
	repo := onmemory.NewOnmemoryTodoRepository()
	usecase := usecase.NewTodoUsecase(repo)
	handler := handler.NewTodoHandler(usecase)
	return api.Route(handler)
}

// -----
// utilities

type pointable interface {
	int | string // NOTE: float, bool, and a few other things, but I'll ignore them for now.
}

func ptr[T pointable](v T) *T {
	return &v
}
