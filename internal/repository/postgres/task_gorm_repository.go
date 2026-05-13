package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"todo-api/internal/model"
	"todo-api/internal/repository"
)

type TaskGormRepository struct {
	db *gorm.DB
}

func NewTaskGormRepository(db *gorm.DB) *TaskGormRepository {
	return &TaskGormRepository{db: db}
}

func (r *TaskGormRepository) Create(ctx context.Context, task *model.Task) error {
	record := toRecord(task)

	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return err
	}

	*task = toDomain(record)
	return nil
}

func (r *TaskGormRepository) GetById(ctx context.Context, id int) (*model.Task, error) {
	var record TaskRecord

	err := r.db.WithContext(ctx).
		First(&record, id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrTaskNotFound
		}
		return nil, err
	}

	task := toDomain(record)
	return &task, nil
}

func (r *TaskGormRepository) List(ctx context.Context) ([]model.Task, error) {
	var records []TaskRecord

	if err := r.db.WithContext(ctx).
		Order("id asc").
		Find(&records).
		Error; err != nil {
		return nil, err
	}

	tasks := make([]model.Task, 0, len(records))
	for _, record := range records {
		tasks = append(tasks, toDomain(record))
	}

	return tasks, nil
}

func (r *TaskGormRepository) Update(ctx context.Context, task *model.Task) error {
	record := toRecord(task)

	result := r.db.WithContext(ctx).
		Model(&TaskRecord{}).
		Where("id = ?", task.ID).
		Updates(map[string]any{
			"title":       record.Title,
			"description": record.Description,
			"done":        record.Done,
			"updated_at":  record.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrTaskNotFound
	}

	return nil
}

func (r *TaskGormRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).
		Delete(&TaskRecord{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrTaskNotFound
	}

	return nil
}

func toRecord(task *model.Task) TaskRecord {
	return TaskRecord{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Done:        task.Done,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func toDomain(record TaskRecord) model.Task {
	return model.Task{
		BaseModel: model.BaseModel{
			ID:        record.ID,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
		},
		Title:       record.Title,
		Description: record.Description,
		Done:        record.Done,
	}
}
