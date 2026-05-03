package repository

import (
	"context"
	"errors"
	"sync"
	"time"
	"todo-api/internal/model"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository struct {
	mu     sync.RWMutex
	tasks  map[int]*model.Task
	nextID int
}

func NewTaskRespository() *TaskRepository {
	return &TaskRepository{
		tasks:  make(map[int]*model.Task),
		nextID: 1,
	}
}

func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	task.ID = r.nextID
	task.CreatedAt = now
	task.UpdatedAt = now

	r.tasks[task.ID] = task
	r.nextID++

	return nil
}

func (r *TaskRepository) GetById(ctx context.Context, id int) (*model.Task, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]

	if !ok {
		return nil, ErrTaskNotFound
	}

	copied := *task
	return &copied, nil
}

func (r *TaskRepository) List(ctx context.Context) ([]model.Task, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]model.Task, 0, len(r.tasks))

	for _, task := range r.tasks {
		result = append(result, *task)
	}

	return result, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.tasks[task.ID]
	if !ok {
		return ErrTaskNotFound
	}

	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = task
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return ErrTaskNotFound
	}

	delete(r.tasks, id)
	return nil
}
