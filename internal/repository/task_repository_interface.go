package repository

import (
	"context"
	"errors"
	"todo-api/internal/model"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetById(ctx context.Context, id int) (*model.Task, error)
	List(ctx context.Context) ([]model.Task, error)
	Delete(ctx context.Context, id int) error
}
