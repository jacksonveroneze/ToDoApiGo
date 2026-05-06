package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"todo-api/internal/dto"
	"todo-api/internal/model"
	"todo-api/internal/service"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) RegisterRouters(mux *http.ServeMux) {
	mux.Handle("/tasks", otelhttp.NewHandler(http.HandlerFunc(h.handleTasks), "tasks"))
	mux.Handle("/tasks/", otelhttp.NewHandler(http.HandlerFunc(h.handleTaskByID), "tasksById"))
}

func (h *TaskHandler) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createTask(w, r)
	case http.MethodGet:
		h.ListTasks(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, dto.APIResponse[string]{
			Error: "method not allowed",
		})
	}
}

func (h *TaskHandler) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.APIResponse[string]{
			Error: "invalid task id",
		})
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetById(w, r, id)
	case http.MethodDelete:
		h.Delete(w, r, id)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, dto.APIResponse[string]{
			Error: "method not allowed",
		})
	}
}

func (h *TaskHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.APIResponse[string]{
			Error: "invalid request body",
		})
		return
	}

	task, err := h.service.Create(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.APIResponse[string]{
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, dto.APIResponse[*model.Task]{
		Data: task,
	})
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.List(r.Context())

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.APIResponse[string]{
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, dto.APIResponse[[]model.Task]{
		Data: tasks,
	})
}

func (h *TaskHandler) GetById(w http.ResponseWriter, r *http.Request, id int) {
	task, err := h.service.GetById(r.Context(), id)

	if err != nil {
		if h.service.IsNotFound(err) {
			writeJSON(w, http.StatusNotFound, dto.APIResponse[string]{
				Error: "task not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, dto.APIResponse[string]{
			Error: err.Error(),
		})
	}

	writeJSON(w, http.StatusOK, dto.APIResponse[*model.Task]{
		Data: task,
	})
}

func (h *TaskHandler) Delete(w http.ResponseWriter, t *http.Request, id int) {
	writeJSON(w, http.StatusMethodNotAllowed, dto.APIResponse[string]{
		Error: "method not allowed",
	})
}

func writeJSON[T any](w http.ResponseWriter, statusCode int, payload dto.APIResponse[T]) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(payload)
}

func extractID(path string) (int, error) {
	trimmed := strings.TrimPrefix(path, "/tasks/")

	if trimmed == "" {
		return 0, errors.New("missing id")
	}

	id, err := strconv.Atoi(trimmed)

	if err != nil {
		return 0, err
	}

	return id, nil
}
