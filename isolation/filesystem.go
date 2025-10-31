package isolation

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"syscall"
)

// rootfs系统
func setFileSystem(rootfs string) error {
	//切换目录
	if err := syscall.Chroot(rootfs); err != nil {
		slog.Error("fail to change root", "error", err)
		return err
	}
	//切换工作目录到新根目录
	if err := syscall.Chdir("/"); err != nil {
		slog.Error("chdir error", "error", err)
		return err
	}

	slog.Info("success to set filesystem", "rootfs", rootfs)
	return nil
}

// 挂载文件系统
// 这里挂载proc，tmpfs，devpts三个
func setMount(rootfs string) error {
	//proc系统提供进程信息
	procPath := path.Join(rootfs, "proc")
	if err := os.MkdirAll(procPath, 0755); err != nil {
		slog.Error("fail to create proc dir", "error", err)
	}

	if err := syscall.Mount("proc", procPath, "proc", 0, ""); err != nil {
		return fmt.Errorf("fail to mount proc,%v", err)
	}

	//tmpfs系统提供临时文件存储
	tmpfsPath := path.Join(rootfs, "tmpfs")
	if err := os.Mkdir(tmpfsPath, 0755); err != nil {
		slog.Error("fail to create tmp dir", "error", err)
	}
	if err := syscall.Mount("tmpfs", tmpfsPath, "tmpfs", 0, ""); err != nil {
		return fmt.Errorf("fail to mount tmpfs,%v", err)
	}

	//devpts提供伪终端设备
	devptsPath := path.Join(rootfs, "dev", "pts")
	if err := os.MkdirAll(devptsPath, 0755); err != nil {
		return fmt.Errorf("fail to mount devpts,%v", err)
	}
	if err := syscall.Mount("devpts", devptsPath, "devpts", 0, ""); err != nil {
		return fmt.Errorf("fail to mount devpts,%v", err)
	}
	slog.Info("all success")
	return nil
}
