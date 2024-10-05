package main

import (
	"log"
	"os"

	f "github.com/alnah/task-tracker/internal/factory"
	r "github.com/alnah/task-tracker/internal/repository"
	s "github.com/alnah/task-tracker/internal/storage"
)

func main() {
	if err := os.MkdirAll("../data", os.ModePerm); err != nil {
		log.Fatalf("Failed to create directory: %w", err)
	}

	dataStore := s.NewJSONFileDataStore[f.Tasks]("../data/tasks.json")
	taskFactory := f.TaskFactory{
		Timer:       f.RealTimer{},
		IDGenerator: f.IDGenerator{},
	}

	taskRepository, err := r.NewFileTaskRepository(taskFactory, dataStore)
	if err != nil {
		log.Fatalf("Failed to create task repository: %v", err) // Handle error
	}

	taskRepository.CreateTask("Hello, Task!")
	// tasks, err := taskRepository.ReadAllTasks()
	// if err != nil {
	// 	log.Fatalf("Failed to read tasks: %v", err) // Handle error
	// }
}
