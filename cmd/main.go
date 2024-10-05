package main

import (
	"fmt"
	"log"

	s "github.com/alnah/task-tracker/internal/storage"
	f "github.com/alnah/task-tracker/internal/task_factory"
	r "github.com/alnah/task-tracker/internal/task_repository"
)

func main() {

	dataStore, err := s.NewJSONFileDataStore[f.Tasks]("tasks")
	taskFactory := f.TaskFactory{
		Timer:       f.RealTimer{},
		IDGenerator: f.IDGenerator{},
	}
	if err != nil {
		log.Fatal(err)
	}

	taskRepository, err := r.NewFileTaskRepository(taskFactory, dataStore)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := taskRepository.CreateTask("Hello, Task!"); err != nil {
		log.Fatal(err)
	}

	if _, err := taskRepository.CreateTask("Hello, Task!"); err != nil {
		log.Fatal(err)
	}

	if _, err := taskRepository.CreateTask("Hello, Task!"); err != nil {
		log.Fatal(err)
	}

	if _, err := taskRepository.CreateTask("Hello, Task!"); err != nil {
		log.Fatal(err)
	}

	if _, err := taskRepository.DeleteTask(1); err != nil {
		log.Fatal(err)
	}

	if _, err := taskRepository.DeleteTask(2); err != nil {
		log.Fatal(err)
	}

	var description string = "Hello, World!"
	var status f.Status = f.Done
	if _, err := taskRepository.UpdateTask(r.UpdateTaskParams{
		ID:          3,
		Description: &description,
		Status:      &status,
	}); err != nil {
		log.Fatal(err)
	}

	tasks, err := taskRepository.ReadManyTasks(f.Done)
	if err != nil {
		log.Fatalf("Failed to read tasks: %v", err) // Handle error
	}
	fmt.Printf("%+v\n", tasks[3].Description)
}
