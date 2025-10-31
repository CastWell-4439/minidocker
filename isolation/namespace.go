package isolation

import (
	"docker/network"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/vishvananda/netlink"
)

var DefaultNetwork = &network.Network{
	Name:    "bridge",
	GateWay: "10.0.0.1/24",
	Subnet:  "10.0.0.0/24",
}

type ContainerConfig struct {
	ID      string
	Name    string
	Image   string
	Command string
	Pid     int64
}

func setHostName(ID string) error {
	hostname := ID

	//设置hostname
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		slog.Error("fail to set hostname", "error", err)
		return err
	}
	slog.Info("set container hostname", "hostname", hostname)
	return nil
}

func executeCommand(cmd []string) error {
	if len(cmd) == 0 {
		return fmt.Errorf("empty command")
	}

	commandPath, err := exec.LookPath(cmd[0])
	if err != nil {
		slog.Error("fail to find command", "command", cmd[0], "error", err)
		return err
	}

	argv := make([]string, 0, len(cmd))
	argv = append(argv, cmd...)

	//设置环境变量
	//这个你要用自己手动设置吧唉
	env := []string{
		"PATH=",
		"TERM=",
		"HOME=",
	}

	slog.Info("executing command", "path", commandPath, "args", argv)

	if err := syscall.Exec(commandPath, argv, env); err != nil {
		slog.Error("fail to exec command", "error", err)
		return err
	}
	return nil
}

func SetNameSpace(id string, pid int) error {
	slog.Info("set namespace", "id", id, "pid", pid)

	if err := network.SetNetwork(id, pid); err != nil {
		slog.Error("fail to set namespace", "id", id, "error", err)
		return fmt.Errorf("fail to set namespace,%v", err)
	}
	slog.Info("success to set namespace", "id", id)
	return nil
}

func cleanVethInterface(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		if strings.Contains(err.Error(), "no such network interface") {
			slog.Info("already removed", "interface", name)
			return nil
		}
		return fmt.Errorf("fail to find interface,%v", err)
	}

	//先关了
	if link.Attrs().Flags&net.FlagUp != 0 {
		if err := netlink.LinkSetDown(link); err != nil {
			slog.Error("fail to stop interface", "interface", name, "err", err)
		}
	}

	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("fail to remove interface,%v", err)
	}
	slog.Info("success to remove interface", "interface", name)
	return nil
}

func cleanRoutes(ID string) error {
	//列出所有route规则
	routes, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("fail to list routes,%v", err)
	}
	hostInterface := path.Join("veth_", ID)
	count := 0

	for _, route := range routes {
		if route.LinkIndex > 0 {
			link, err := netlink.LinkByIndex(route.LinkIndex)
			if err != nil && strings.Contains(err.Error(), "no such interface") {
				continue
			} else if err != nil {
				slog.Error("fail to get link by index,%v", err)
				continue
			}
			if link.Attrs().Name == hostInterface {
				if err := netlink.RouteDel(&route); err != nil {
					slog.Error("fail to delete route", "route", route, "err", err)
				} else {
					count++
				}
			}
		}
	}
	if count > 0 {
		slog.Info("success to delete routes", "count", count)
	}
	return nil
}

func cleanNATRule(id string) error {
	rules := []string{
		//这里面写你需要的规则
		fmt.Sprintf(""),
	}
	for _, rule := range rules {
		cmd := exec.Command("iptables", "-t", "nat", "-D", rule)
		if out, err := cmd.CombinedOutput(); err != nil {
			if !strings.Contains(string(out), "no target") {
				slog.Error("fail to remove rule", "rule", rule, "err", err)
			}
		} else {
			slog.Info("success to remove rule", "rule", rule)
		}
	}
	return nil
}

func cleanFilter(ID string) error {
	hostInterface := path.Join("veth_", ID)
	rules := []string{
		fmt.Sprintf("%s%s", hostInterface, ID),
		//还是写你要清的rule
	}
	for _, rule := range rules {
		cmd := exec.Command("iptables", "-D", rule)
		if out, err := cmd.CombinedOutput(); err != nil {
			if !strings.Contains(string(out), "no target") {
				slog.Error("fail to remove filter rule", "rule", rule, "err", err)
			}
		}
	}
	return nil
}

func cleanResource(id string) error {
	hostInterface := path.Join("veth_", id)

	if err := cleanVethInterface(hostInterface); err != nil {
		return fmt.Errorf("fail to clean veth interface,%v", err)
	}
	if err := cleanNATRule(id); err != nil {
		return fmt.Errorf("fail to clean nat rule,%v", err)
	}
	if err := cleanFilter(id); err != nil {
		return fmt.Errorf("fail to clean fliter,%v", err)
	}
	if err := cleanRoutes(id); err != nil {
		return fmt.Errorf("fail to clean routes,%v", err)
	}
	slog.Info("success to clean resource", "id", id)
	return nil
}

func CleanNameSpace(id string) error {
	if err := cleanResource(id); err != nil {
		slog.Error("fail to clean resource", "id", id)
		return fmt.Errorf("fail to clean resource,%v", err)
	}
	return nil
}

func StartContainer(ID, rootfs string, cmd []string, interactive bool) (int, error) {
	slog.Info("start container", "containID", ID, "command", cmd)

	c := exec.Command("/proc/self/exe", "init")

	//如果发现这段代码报错了请切换到linux
	//Cloneflags这个字面量只有在linux的syscall包里有
	//嗯而且本来这个就是写给linux用的
	c.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	} //CLONE_NEWUTS: 隔离主机名和域名，CLONE_NEWPID: 隔离进程ID空间，CLONE_NEWNS: 隔离挂载点，CLONE_NEWNET: 隔离网络栈，CLONE_NEWIPC: 隔离IPC

	if interactive {
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}

	c.Dir = rootfs

	if err := c.Start(); err != nil {
		return 0, err
	}

	return c.Process.Pid, nil
}

func InitProcess(ID string, rootfs string, cmd []string) error {
	slog.Info("container init", "containerID", ID)

	if err := setHostName(ID); err != nil {
		return err
	}
	if err := setFileSystem(rootfs); err != nil {
		return err
	}
	if err := setMount(rootfs); err != nil {
		return err
	}

	return executeCommand(cmd)

}

func ExecInContainer(ID string, rootfs string, cmd []string) error {
	slog.Info("execute command", "containerID", ID, "command", cmd)

	if err := setHostName(ID); err != nil {
		return err
	}
	if err := setFileSystem(rootfs); err != nil {
		return err
	}
	return executeCommand(cmd)
}

func InitContainer(rootfs string) error {
	if err := setMount(rootfs); err != nil {
		return fmt.Errorf("fail to mount rootfs,%v", err)
	}
	return nil
}
