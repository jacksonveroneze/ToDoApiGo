package handler

import (
	"encoding/json"
	"net/http"
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
}

func (h *TaskHandler) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createTask(w, r)
	case http.MethodGet:
		h.ListTasks(w, r)
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

func writeJSON[T any](w http.ResponseWriter, statusCode int, payload dto.APIResponse[T]) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(payload)
}
