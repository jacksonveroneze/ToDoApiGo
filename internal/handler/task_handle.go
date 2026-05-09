package handler

import (
	"net/http"
	"strconv"
	"todo-api/internal/dto"
	"todo-api/internal/service"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandlerGin(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) RegisterRoutes(router gin.IRouter) {
	tasks := router.Group("/tasks")
	{
		tasks.POST("", h.createTask)
		tasks.GET("", h.listTasks)
		tasks.GET("/:id", h.getTaskById)
		tasks.DELETE("/:id", h.deleteTask)
	}
}

func (h *TaskHandler) createTask(c *gin.Context) {
	var req dto.CreateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	taskInput := service.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
	}

	task, err := h.service.Create(c.Request.Context(), taskInput)

	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}

	writeData(c, http.StatusCreated, task)
}

func (h *TaskHandler) listTasks(c *gin.Context) {
	tasks, err := h.service.List(c.Request.Context())

	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}

	writeData(c, http.StatusOK, tasks)
}

func (h *TaskHandler) getTaskById(c *gin.Context) {
	id, err := parseId(c.Param("id"))

	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.service.GetById(c.Request.Context(), id)

	if err != nil {
		if h.service.IsNotFound(err) {
			writeError(c, http.StatusNotFound, "task not found")
			return
		}

		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}

	writeData(c, http.StatusOK, task)
}

func (h *TaskHandler) deleteTask(c *gin.Context) {
	id, err := parseId(c.Param("id"))

	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid task id")
		return
	}

	errDelete := h.service.Delete(c.Request.Context(), id)

	if errDelete != nil {
		if h.service.IsNotFound(errDelete) {
			writeError(c, http.StatusNotFound, "task not found")
			return
		}

		writeError(c, http.StatusInternalServerError, errDelete.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func parseId(raw string) (int, error) {
	return strconv.Atoi(raw)
}

func writeData[T any](c *gin.Context, status int, data T) {
	c.JSON(status, dto.APIResponse[T]{Data: data})
}

func writeError(c *gin.Context, status int, msg string) {
	c.JSON(status, dto.APIResponse[any]{Error: msg})
}
