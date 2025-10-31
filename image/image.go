package image

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"docker/storage"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
)

type Image struct {
	Name string
	Tag  string
	Size int64
}

type descriptor struct {
	Type   string `json:"type"`
	Size   int64  `json:"size"`
	Digest string `json:"digest"`
}
type manifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	Type          string       `json:"type"`
	Config        descriptor   `json:"config"`
	Layers        []descriptor `json:"layers"`
}

func parse(imagename string) (registry, repo, tag string, err error) {
	part := strings.SplitN(imagename, ":", 2)
	name := part[0]
	tag = "latest"
	if len(part) > 1 {
		tag = part[1]
	}

	if strings.Contains(name, ":") {
		registryRepo := strings.SplitN(name, "/", 2)
		if len(registryRepo) != 2 {
			return "", "", "", fmt.Errorf("fail to find image: %s", name)
		}
		registry = registryRepo[0]
		repo = registryRepo[1]
	} else {
		registry = "docker.io"
		if !strings.Contains(name, "/") {
			repo = "library/" + name
		} else {
			repo = name
		}
	}
	return registry, repo, tag, nil
}

func getImageManifest(registry, repo, tag string) (*manifest, error) {
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, repo, tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("User-Agent", "easydocker/1.0")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		url = fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, repo, tag)
		res, err = http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("fail to get image manifest: %s", err)
		}
	}
	defer res.Body.Close()

	//其实点进去发现http.StatusOK完全就是200吧。。。真的有必要封装这个吗
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fail to get  manifest")
	}

	var m manifest
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("fail to parse manifest,%v", err)
	}
	return &m, nil
}

func saveManifest(m *manifest, imagePath string) error {
	manifestPath := path.Join(imagePath, "manifest.json")
	f, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(m)
}

func getBlobPath(digest, blobRoot string) string {
	hash := strings.TrimPrefix(digest, "sha256:")
	return path.Join(blobRoot, "sha256", hash)
}

func verify(blobPath, digest string, size int64) error {
	f, err := os.Stat(blobPath)
	if err != nil {
		return err
	}
	if f.Size() != size {
		return fmt.Errorf("error size")
	}
	ff, err := os.Open(blobPath)
	if err != nil {
		return err
	}
	defer ff.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, ff); err != nil {
		return err
	}
	Digest := fmt.Sprintf("sha256:%x", hash.Sum(nil))
	if Digest != digest {
		return fmt.Errorf("error hash")
	}
	return nil
}

func downloadImage(registry, repo, blobRoot string, des descriptor) error {
	blobPath := getBlobPath(des.Digest, blobRoot)
	if err := verify(blobPath, des.Digest, des.Size); err != nil {
		slog.Warn("blob already exist")
		return err
	}
	url := fmt.Sprintf("https://%s/v2/%s/blobs/%s", registry, repo, des.Digest)
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("fail to download blob,%v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to download blob")
	}

	if err := os.MkdirAll(path.Dir(blobPath), 0755); err != nil {
		return err
	}
	f, err := os.Create(blobPath)
	if err != nil {
		return fmt.Errorf("fail to create blob,%v", err)
	}
	defer f.Close()

	hash := sha256.New()
	tee := io.TeeReader(res.Body, hash)
	if _, err := io.Copy(f, tee); err != nil {
		return fmt.Errorf("fail to write blob,%v", err)
	}
	Digest := fmt.Sprintf("sha256:%x", hash.Sum(nil))
	if Digest != des.Digest {
		os.Remove(blobPath)
		return fmt.Errorf("hash do not match")
	}
	if ff, _ := f.Stat(); ff.Size() != des.Size {
		os.Remove(blobPath)
		return fmt.Errorf("size do not match")
	}
	return nil
}

func ExtractImage(imagepath, des string) error {
	file, err := os.Open(imagepath)
	if err != nil {
		return err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()

	tars := tar.NewReader(gz)

	for {
		header, err := tars.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := path.Join(des, header.Name)

		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(targetPath, 0777); err != nil {
				return err
			}
		} else if header.Typeflag == tar.TypeReg {
			out, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tars); err != nil {
				out.Close()
				return err
			}
			out.Close()
			os.Chmod(targetPath, os.FileMode(header.Mode))
		}
	}
	return nil
}

func Pull(name string) error {
	slog.Info("pulling image:", name)

	registry, repo, tag, err := parse(name)
	if err != nil {
		return fmt.Errorf("fail to parse image: registry=%s,repo=%s,tag=%s,err=%v", registry, repo, tag, err)
	}
	slog.Debug("pulling image:registry=%s,repo=%s,tag=%s,err=%v", registry, repo, tag, err)

	manifest, err := getImageManifest(registry, repo, tag)
	if err != nil {
		return fmt.Errorf("fail to get image manifest: registry=%s,repo=%s,tag=%s", name, repo, tag)
	}

	imagePath := path.Join(storage.ImageRoot, name)
	if err := os.MkdirAll(imagePath, 0755); err != nil {
		return fmt.Errorf("fail to create image dir,%v", err)
	}
	rootfsPath := path.Join(imagePath, "rootfs")
	if err := os.MkdirAll(rootfsPath, 0755); err != nil {
		return fmt.Errorf("fail to create rootfs dir,%v", err)
	}

	if err := saveManifest(manifest, imagePath); err != nil {
		return fmt.Errorf("fail to save manifest,%v", err)
	}

	if err := downloadImage(registry, repo, storage.BlobRoot, manifest.Config); err != nil {
		return fmt.Errorf("fail to download blob,%v", err)
	}

	for i, layer := range manifest.Layers {
		slog.Info("layer:", i+1, layer.Digest)
		if err := downloadImage(registry, repo, storage.BlobRoot, layer); err != nil {
			return fmt.Errorf("fail to download in layer:%d,%v", i+1, err)
		}
		layerPath := getBlobPath(layer.Digest, storage.Root)
		if err := ExtractImage(layerPath, rootfsPath); err != nil {
			return fmt.Errorf("fail to extract in layer:%d,digest:%s,%v", i+1, layer.Digest, err)
		}
	}
	return nil
}
