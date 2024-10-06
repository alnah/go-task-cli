package store

import (
	"os"
)

type Store[T any] interface {
	Init() (*os.File, error)
	LoadData(string) (T, error)
	SaveData(T, string) error
	CloseFile(*os.File) error
}

type JSONInitData string

const (
	EmptyArray  JSONInitData = "[]"
	EmptyObject JSONInitData = "{}"
)

type JSONFileStore[T any] struct {
	DestDir  string
	Filename string
	InitData JSONInitData
}
