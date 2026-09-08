package service

import (
	"context"
	"fmt"

	"github.com/newaccg/todo-api/internal/model"
)

func (s *Service) DeleteTaskByIDWithUserID(ctx context.Context, taskID, userID int64) error {
	if err := s.repo.DeleteByIDWithUserID(ctx, taskID, userID); err != nil {
		return fmt.Errorf("could not delete task by task ID %d and user ID %d: %w", taskID, userID, err)
	}

	return nil
}

func (s *Service) UpdateTaskByIDWithUserID(ctx context.Context, taskID, userID int64, title, description string) (*model.Task, error) {
	task := model.Task{
		Title:       title,
		Description: description,
	}

	newTask, err := s.repo.UpdateByIDWithUserID(ctx, taskID, userID, &task)
	if err != nil {
		return nil, fmt.Errorf("could not update task by task ID %d and user ID %d: %w", taskID, userID, err)
	}

	return newTask, nil
}

func (s *Service) GetAllTasksWithUserID(ctx context.Context, userID int64, filter, order string, page, limit int) ([]model.Task, error) {
	tasks, err := s.repo.GetAllWithUserID(ctx, userID, filter, order, page, limit)
	if err != nil {
		return nil, fmt.Errorf("could not get all tasks: %w", err)
	}

	return tasks, nil
}

func (s *Service) CreateTaskWithUserID(ctx context.Context, title, description string, id int64) (*model.Task, error) {
	task := model.Task{
		Title:       title,
		Description: description,
	}

	newTask, err := s.repo.CreateWithUserID(ctx, &task, id)
	if err != nil {
		return nil, fmt.Errorf("could not create task: %w", err)
	}

	return newTask, nil
}
