package storage

import (
	"os"
	"path"
)

type StorageDriver interface {
	Init() error
	Create(name string) error
	Remove(name string) error
	Exists(name string) bool
}

type FileDriver struct{}

func (f *FileDriver) Init() error {
	dir := []string{ImageRoot, ContainRoot}
	for _, dir := range dir {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileDriver) Create(name string) error {
	imagePath := path.Join(ImageRoot, name)
	return os.MkdirAll(imagePath, 0755)
}

func (f *FileDriver) Remove(name string) error {
	imagePath := path.Join(ImageRoot, name)
	return os.RemoveAll(imagePath)
}

func (f *FileDriver) Exists(name string) bool {
	imagePath := path.Join(ImageRoot, name)
	_, err := os.Stat(imagePath)
	return !os.IsNotExist(err)
}
