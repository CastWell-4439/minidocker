package isolation

import (
	"log/slog"
	"os"
	"path"
	"strconv"
)

// 这里也是为了简化的硬编码嗯。。。无所谓了
const (
	cgroupRoot = "/sys/fs/cgroup"
)

type CgroupManager struct {
	Name string
}

func NewCgroupManager(id string) *CgroupManager {
	return &CgroupManager{
		Name: "easydocker-" + id,
	}
}

func (c *CgroupManager) Set(pid int) error {
	//需要配置cpu，内存和进程数
	subsystems := []string{"cpu", "memory", "pids"}

	for _, sub := range subsystems {
		cgroupPath := path.Join(cgroupRoot, sub, c.Name)
		if err := os.MkdirAll(cgroupPath, 0777); err != nil {
			return err
		}
		pidPath := path.Join(cgroupPath, "cgroup.procs")
		if err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0644); err != nil {
			return err
		}

		if sub == "cpu" {
			//cpu.shares用于设置CPU份额，512表示在CPU竞争时，该cgroup获得的CPU时间是512
			if err := os.WriteFile(path.Join(cgroupPath, "cpu.shares"), []byte("512"), 0644); err != nil {
				slog.Error("fail to set cpu shares", "error", err)
			}
		} else if sub == "memory" {
			//内存上限100m
			if err := os.WriteFile(path.Join(cgroupPath, "memory.limit_in_bytes"), []byte("100M"), 0644); err != nil {
				slog.Error("fail to set memory limit_in_bytes", "error", err)
			}
		} else if sub == "pids" {
			//最多100个进程
			if err := os.WriteFile(path.Join(cgroupPath, "pids.max"), []byte("100"), 0644); err != nil {
				slog.Error("fail to set pids", "error", err)
			}
		} else {
			return nil
		}
	}
	return nil
}

func (c *CgroupManager) Remove() error {
	subsystems := []string{"cpu", "memory", "pids"}

	for _, sub := range subsystems {
		cgroupPath := path.Join(cgroupRoot, sub, c.Name)
		if err := os.RemoveAll(cgroupPath); err != nil {
			return err
		}
	}
	return nil
}
