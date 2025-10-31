package container

import (
	"encoding/json"
	"os"
	"path"
)

func GetContainer(ID string) (*Container, error) {
	infoPath := path.Join(ContainerRoot, ID, "info")

	data, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, err
	}

	var info Container
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func UpdateContainer(ID, status string) error {
	info, err := GetContainer(ID)
	if err != nil {
		return err
	}
	info.Status = status
	return saveContainerInfo(info)
}
