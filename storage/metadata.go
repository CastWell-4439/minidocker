package storage

import (
	"encoding/json"
	"os"
	"path"
	"time"
)

var (
	//根目录
	Root        = "/var/lib/easydocker"
	ImageRoot   = path.Join(Root, "images")
	ContainRoot = path.Join(Root, "contains")
	BlobRoot    = path.Join(Root, "blobs")
)

type ImageMetadata struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Size    int64  `json:"size"`
}

type ContainerMetadata struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Image      string    `json:"image"`
	Status     string    `json:"status"`
	CreateTime time.Time `json:"createTime"`
}

func SaveImageMetadata(data *ImageMetadata) error {
	imagePath := path.Join(ImageRoot, data.Name)
	metadataPath := path.Join(imagePath, "metadata.json")
	if err := os.MkdirAll(imagePath, 0755); err != nil {
		return err
	}

	file, err := os.Create(metadataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	//第二个参数是前缀，第三个参数是缩进
	datas, err := json.MarshalIndent(data, "", "	")
	if err != nil {
		return err
	}
	_, err = file.Write(datas)
	if err != nil {
		return err
	}
	return nil
}

func LoadImageMetadata(name string) (*ImageMetadata, error) {
	metadataPath := ImageRoot + name + "metadata.json"
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata ImageMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}
