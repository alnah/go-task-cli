package service

import (
	r "github.com/alnah/task-tracker/internal/repository"
	s "github.com/alnah/task-tracker/internal/storage"
)

type Service interface {
	AddTask(string) (r.Task, error)
	UpdateTask(uint, *string, *r.Status) (r.Task, error)
	DeleteTask(uint) (bool, error)
	ListTasks(*r.Status) (r.Tasks, error)
}

type TaskService struct {
	Storage    s.DataStore[r.Tasks]
	Repository r.TaskRepository
}

type UpdateTaskParams struct {
	ID          uint
	Description *string
	Status      *r.Status
}

func (s *TaskService) AddTask(description string) (r.Task, error) {
	if _, loadErr := s.loadAndImportTasks(); loadErr != nil {
		return r.Task{}, loadErr
	}

	task, addErr := s.Repository.AddTask(description)
	if addErr != nil {
		return r.Task{}, addErr
	}

	_, saveErr := s.Storage.SaveTasks(s.Writer, s.Repository.FindAll())
	if saveErr != nil {
		return r.Task{}, saveErr
	}

	return task, nil
}

func (s *TaskService) UpdateTask(params UpdateTaskParams) (r.Task, error) {
	if _, loadErr := s.loadAndImportTasks(); loadErr != nil {
		return r.Task{}, loadErr
	}

	updateParams := r.UpdateTaskParams{
		ID:          params.ID,
		Description: params.Description,
		Status:      params.Status,
	}

	task, updateErr := s.Repository.UpdateTask(updateParams)
	if updateErr != nil {
		return r.Task{}, updateErr
	}

	_, saveErr := s.Storage.SaveTasks(s.Writer, s.Repository.FindAll())
	if saveErr != nil {
		return r.Task{}, saveErr
	}

	return task, nil
}

func (s *TaskService) DeleteTask(id uint) (bool, error) {
	if _, loadErr := s.loadAndImportTasks(); loadErr != nil {
		return false, loadErr
	}

	if _, deleteErr := s.Repository.DeleteTask(id); deleteErr != nil {
		return false, deleteErr
	}

	_, saveErr := s.Storage.SaveTasks(s.Writer, s.Repository.FindAll())
	if saveErr != nil {
		return false, saveErr
	}

	return true, nil
}

func (s *TaskService) ListTasks(status *r.Status) (r.Tasks, error) {
	if _, loadErr := s.loadAndImportTasks(); loadErr != nil {
		return r.Tasks{}, loadErr
	}

	if status != nil {
		return s.Repository.FindMany(*status), nil
	}

	return s.Repository.FindAll(), nil
}

func (s *TaskService) loadAndImportTasks() (r.Tasks, error) {
	loadedTasks, loadErr := s.Storage.LoadTasks(s.Reader)
	if loadErr != nil {
		return r.Tasks{}, loadErr
	}

	tasks := s.Repository.ImportTasksData(loadedTasks)
	return tasks, nil
}
