package image

import (
	"docker/storage"
	"os"
	"path"
)

func CreateRootfs(name string) error {
	imagePath := path.Join(storage.ImageRoot, name, "rootfs")
	if err := os.MkdirAll(imagePath, 0777); err != nil {
		return err
	}

	dirs := []string{
		"bin",   //存二进制可执行文件
		"dev",   //设备文件
		"etc",   //系统配置
		"home",  //一般路过用户的主目录
		"lib",   //32位系统库
		"lib64", // 64位系统库
		"proc",  //虚拟文件系统，提供进程和系统信息
		"root",  //root的根目录
		"sys",   //也是虚拟文件系统，但是这个用于访问内核设备和subsystem信息
		"tmp",   //临时文件
		"usr",   //用户相关的文件
		"var",   //经常变化的比如日志缓存啥的
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(path.Join(imagePath, dir), 0755); err != nil {
			return err
		}
	}

	return nil
}
