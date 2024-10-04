package main

import (
	"fmt"

	"github.com/alnah/task-tracker/internal/repository"
	"github.com/alnah/task-tracker/internal/service"
	"github.com/alnah/task-tracker/internal/storage"
)

func main() {
	service := service.TaskService{
		Storage: &storage.StorageJSON{},
		Repository: &repository.MemoTaskRepo{
			Timer:   repository.RealTimer{},
			Counter: repository.IDCounter{},
			Tasks:   repository.Tasks{},
		},
	}
	fmt.Print(service)
}
