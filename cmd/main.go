package main

import (
	"fmt"
	"log"

	ds "github.com/alnah/task-tracker/internal/data_store"
	tf "github.com/alnah/task-tracker/internal/task_factory"
	tr "github.com/alnah/task-tracker/internal/task_repository"
)

func main() {

	dataStore, err := ds.NewJSONFileDataStore[tf.Tasks]("tasks")
	taskFactory := tf.DefaultTaskFactory{
		Timer:       &tf.DefaultTimeProvider{},
		IDGenerator: &tf.DefaultIDGenerator{},
	}
	if err != nil {
		log.Fatal(err)
	}

	taskRepository, err := tr.NewFileTaskRepository(taskFactory, *dataStore)
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
	var status tf.Status = tf.Done
	if _, err := taskRepository.UpdateTask(tr.UpdateTaskParams{
		ID:          3,
		Description: &description,
		Status:      &status,
	}); err != nil {
		log.Fatal(err)
	}

	tasks, err := taskRepository.ReadManyTasks(tf.Done)
	if err != nil {
		log.Fatalf("Failed to read tasks: %v", err) // Handle error
	}
	fmt.Printf("%+v\n", tasks[3].Description)
}
