package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"todo-api/internal/dto"
	"todo-api/internal/model"
	"todo-api/internal/repository"
	"todo-api/internal/worker"
)

type TaskService struct {
	repo   *repository.TaskRepository
	events chan<- worker.AuditEvent
}

func NewTaskService(repo *repository.TaskRepository, events chan<- worker.AuditEvent) *TaskService {
	return &TaskService{
		repo:   repo,
		events: events,
	}
}

func (s *TaskService) Create(ctx context.Context, req dto.CreateTaskRequest) (*model.Task, error) {
	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)

	if title == "" || description == "" {
		return nil, fmt.Errorf("title or description are required")
	}

	task := &model.Task{
		Title:       title,
		Description: description,
		Done:        false,
	}

	err := s.repo.Create(ctx, task)

	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	s.publishEvent("created", task.ID)

	return task, nil
}

func (s *TaskService) List(ctx context.Context) ([]model.Task, error) {
	tasks, err := s.repo.List(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, nil

}

func (s *TaskService) GetById(ctx context.Context, id int) (*model.Task, error) {
	task, err := s.repo.GetById(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get task by id: %w", err)
	}

	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	s.publishEvent("deleted", id)

	return nil
}

func (s *TaskService) IsNotFound(err error) bool {
	return errors.Is(err, repository.ErrTaskNotFound)
}

func (s *TaskService) publishEvent(action string, taskID int) {
	select {
	case s.events <- worker.AuditEvent{
		Action: action,
		TaskID: taskID,
		When:   time.Now(),
	}:
	default:
	}
}
