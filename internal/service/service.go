package service

import (
	"io"

	rp "github.com/alnah/task-tracker/internal/repository"
	s "github.com/alnah/task-tracker/internal/storage"
)

type Service interface {
	AddTask(io.Reader, io.Writer, string) (rp.Task, error)
	UpdateTask(io.Reader, io.Writer, UpdateTaskParams) (rp.Task, error)
	DeleteTask(io.Reader, uint) (bool, error)
	ListTasks(io.Reader, *rp.Status) (rp.Tasks, error)
}

type TaskService struct {
	Storage    s.Storage
	Repository rp.Repository
}

type UpdateTaskParams struct {
	ID          uint
	Description *string
	Status      *rp.Status
}

func (s *TaskService) AddTask(
	r io.Reader,
	w io.Writer,
	description string,
) (rp.Task, error) {
	if _, loadErr := s.loadAndImportTasks(r); loadErr != nil {
		return rp.Task{}, loadErr
	}

	task, addErr := s.Repository.AddTask(description)
	if addErr != nil {
		return rp.Task{}, addErr
	}

	_, saveErr := s.Storage.SaveTasks(w, s.Repository.FindAll())
	if saveErr != nil {
		return rp.Task{}, saveErr
	}

	return task, nil
}

func (s *TaskService) UpdateTask(
	r io.Reader,
	w io.Writer,
	params UpdateTaskParams,
) (rp.Task, error) {
	if _, loadErr := s.loadAndImportTasks(r); loadErr != nil {
		return rp.Task{}, loadErr
	}

	updateParams := rp.UpdateTaskParams{
		ID:          params.ID,
		Description: params.Description,
		Status:      params.Status,
	}

	task, updateErr := s.Repository.UpdateTask(updateParams)
	if updateErr != nil {
		return rp.Task{}, updateErr
	}

	_, saveErr := s.Storage.SaveTasks(w, s.Repository.FindAll())
	if saveErr != nil {
		return rp.Task{}, saveErr
	}

	return task, nil
}

func (s *TaskService) DeleteTask(
	r io.Reader,
	w io.Writer,
	id uint,
) (bool, error) {
	if _, loadErr := s.loadAndImportTasks(r); loadErr != nil {
		return false, loadErr
	}

	if _, deleteErr := s.Repository.DeleteTask(id); deleteErr != nil {
		return false, deleteErr
	}

	_, saveErr := s.Storage.SaveTasks(w, s.Repository.FindAll())
	if saveErr != nil {
		return false, saveErr
	}

	return true, nil
}

func (s *TaskService) ListTasks(
	r io.Reader,
	status *rp.Status,
) (rp.Tasks, error) {
	if _, loadErr := s.loadAndImportTasks(r); loadErr != nil {
		return rp.Tasks{}, loadErr
	}

	if status != nil {
		return s.Repository.FindMany(*status), nil
	}

	return s.Repository.FindAll(), nil
}

func (s *TaskService) loadAndImportTasks(r io.Reader) (rp.Tasks, error) {
	loadedTasks, loadErr := s.Storage.LoadTasks(r)
	if loadErr != nil {
		return rp.Tasks{}, loadErr
	}

	tasks := s.Repository.ImportTasksData(loadedTasks)
	return tasks, nil
}
