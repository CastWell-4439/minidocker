package storage

import (
	"io"
	"os"
	"path"
)

func Copy(src, dst string) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer to.Close()
	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}
	return nil
}

func CreateDir(root string, dir []string) error {
	for _, d := range dir {
		if err := os.MkdirAll(path.Join(root, d), 0777); err != nil {
			return err
		}
	}
	return nil
}
