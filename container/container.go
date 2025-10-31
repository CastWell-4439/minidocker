package container

import (
	"docker/image"
	"docker/isolation"
	"docker/storage"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	//这里不直接用storage的是因为一个得是var一个得是const
	ContainerRoot = "/var/lib/easydocker/container"
	ContainerLog  = "container.log"
	ContainerInfo = "container.json"
)

type Container struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Image      string    `json:"image"`
	Command    string    `json:"command"`
	CreateTime time.Time `json:"create_time"`
	Status     string    `json:"status"`
	Pid        int       `json:"pid"`
}

func makeID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func NewContainer(name, image, command string, pid int) *Container {
	return &Container{
		ID:         makeID(),
		Name:       name,
		Image:      image,
		Command:    command,
		CreateTime: time.Now(),
		Status:     "running",
		Pid:        pid,
	}
}

func saveContainerInfo(container *Container) error {
	infoPath := path.Join(ContainerRoot, container.ID, "info")
	file, err := os.Create(infoPath)
	if err != nil {
		return fmt.Errorf("fail to create container,%v", err)
	}
	defer file.Close()

	data, err := json.MarshalIndent(container, "", "	")
	if err != nil {
		return fmt.Errorf("fail to marshal container,%v", err)
	}
	_, err = file.Write(data)
	return err
}

func ListContainers() error {
	entries, err := os.ReadDir(ContainerRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("did not exist,%v", err)
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		infoPath := path.Join(ContainerRoot, entry.Name(), "info")
		data, err := os.ReadFile(infoPath)
		if err != nil {
			continue
		}

		var container Container
		if err := json.Unmarshal(data, &container); err != nil {
			continue
		}
		fmt.Printf("%s  %s  %s  %s\n", container.ID, container.Name, container.Image, container.Status)
	}
	return nil
}

func Stop(ID string) error {
	infoPath := path.Join(ContainerRoot, ID, "info")
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return fmt.Errorf("fail to read file,%v", err)
	}

	var info Container
	if err := json.Unmarshal(data, &info); err != nil {
		return fmt.Errorf("fail to unmarshal container,%v", err)
	}

	process, err := os.FindProcess(info.Pid)
	if err != nil {
		return err
	}
	if err := process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("fail to signal,%v", err)
	}

	timeout := time.After(10 * time.Second)
	ticker := time.Tick(100 * time.Millisecond)

	for {
		select {
		case <-timeout:
			if err := process.Kill(); err != nil {
				return fmt.Errorf("fail to kill container,%v", err)
			}
			return nil
		case <-ticker:
			if err := process.Signal(syscall.Signal(0)); err != nil {
				info.Status = "stopped"
				return saveContainerInfo(&info)
			}
		}

	}
}

func Exec(ID string, cmd []string) error {
	infoPath := path.Join(ContainerRoot, ID, "info")
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return fmt.Errorf("fail to read file,%v", err)
	}
	var info Container
	if err := json.Unmarshal(data, &info); err != nil {
		return fmt.Errorf("fail to unmarshal container,%v", err)
	}
	rootfs := path.Join(storage.ImageRoot, info.Image, "rootfs")
	return isolation.ExecInContainer(info.ID, rootfs, cmd)
}

func Run(name, images string, command []string, interactive bool) (int, error) {
	container := NewContainer(name, images, strings.Join(command, " "), os.Getpid())
	containerDir := path.Join(ContainerRoot, container.ID)
	var err error
	if err := os.MkdirAll(containerDir, 0755); err != nil {
		return container.Pid, fmt.Errorf("fail to create dir,%v", err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(containerDir)
		}
	}()

	rootfs, err := image.Check(images, container.ID)
	if err != nil {
		return container.Pid, fmt.Errorf("rootfs error,%v", err)
	}
	if err := saveContainerInfo(container); err != nil {
		return container.Pid, err
	}
	return isolation.StartContainer(container.ID, rootfs, command, interactive)
}
