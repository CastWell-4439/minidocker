package image

import (
	"docker/storage"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
)

func copy(src, dst string) error {
	entry, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("fail to read dir, %v", err)
	}

	for _, e := range entry {
		srcPath := path.Join(src, e.Name())
		dstPath := path.Join(dst, e.Name())

		if e.IsDir() {
			//递归地创建/复制
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copy(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			srcFile, err := os.Open(srcPath)
			if err != nil {
				return fmt.Errorf("fail to open file, %v", err)
			}
			defer srcFile.Close()

			dstFile, err := os.Create(dstPath)
			if err != nil {
				return fmt.Errorf("fail to create file, %v", err)
			}
			defer dstFile.Close()

			if _, err := io.Copy(dstFile, srcFile); err != nil {
				return fmt.Errorf("fail to copy file, %v", err)
			}
			if err := dstFile.Chmod(0755); err != nil {
				slog.Warn("fail to chmod file, %v", err)
			}
		}
	}
	return nil
}

func Check(name, containerID string) (string, error) {
	imagePath := path.Join(storage.ImageRoot, name)

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		if err := Pull(name); err != nil {
			return "", err
		}
	}

	containerRootfs := path.Join(storage.ContainRoot, containerID, "rootfs")
	if err := os.MkdirAll(containerRootfs, 0755); err != nil {
		return "", err
	}

	imageRootfs := path.Join(imagePath, "rootfs")
	if err := copy(imageRootfs, containerRootfs); err != nil {
		return "", err
	}
	return imagePath, nil
}
