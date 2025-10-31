package network

import (
	"fmt"
	"log/slog"

	"github.com/vishvananda/netlink"
)

func SetBridge() error {
	name := "easydocker"

	_, err := netlink.LinkByName(name)
	if err == nil {
		slog.Info("already exist", "bridge", name)
		return nil
	}

	bridge := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: name,
		},
	}
	if err := netlink.LinkAdd(bridge); err != nil {
		return fmt.Errorf("fail to create bridge,%v", err)
	}

	//获取创建的
	bri, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("fail to get bridge,%v", err)
	}

	//配置ip
	addr, err := netlink.ParseAddr("172.17.0.1/24")
	if err != nil {
		return fmt.Errorf("fail to parse addr,%v", err)
	}

	if err := netlink.AddrAdd(bri, addr); err != nil {
		return fmt.Errorf("fail to add addr,%v", err)
	}

	//启动！
	if err := netlink.LinkSetUp(bri); err != nil {
		return fmt.Errorf("fail to start bridge,%v", err)
	}
	slog.Info("created successful", "bridge", name)

	return nil
}
