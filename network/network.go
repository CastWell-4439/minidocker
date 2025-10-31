package network

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type Network struct {
	Name    string
	GateWay string
	Subnet  string
}

func SetNetwork(containerID string, pid int) error {
	//用veth pair来连接namespace
	//主机接口
	hostInterface := "veth_" + containerID[:8]
	//容器内部接口
	containerInterface := "eth0"

	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:  hostInterface,
			Flags: net.FlagUp, //创建后立即启用
		},
		PeerName: containerInterface, //接口是veth pair的另一端
	}

	var err error
	//错误的话立刻删除veth
	defer func() {
		if err != nil {
			slog.Warn("fail to clean veth")
			netlink.LinkDel(veth)
		}
	}()

	//注册并创建veth pair
	if err := netlink.LinkAdd(veth); err != nil {
		slog.Error("fail to add veth")
		return err
	}

	//通过name查找veth pair容器端接口
	peer, err := netlink.LinkByName(containerInterface)
	if err != nil {
		slog.Error("fail to get peer")
		return err
	}

	//接口移动到container的ns
	if err := netlink.LinkSetNsPid(peer, pid); err != nil {
		slog.Error("fail to move peer")
		return err
	}

	bridge, err := netlink.LinkByName("easydocker")
	if err != nil {
		return fmt.Errorf("fail to get bridge")
	}
	//主机veth接口加入bridge
	if err = netlink.LinkSetMaster(veth, bridge); err != nil {
		return fmt.Errorf("fail to set bridge")
	}

	//启动主机veth
	if err = netlink.LinkSetUp(veth); err != nil {
		return fmt.Errorf("fail to set veth up")
	}

	nsHandle, err := netns.GetFromPid(pid)
	if err != nil {
		return fmt.Errorf("fail to get netns")
	}
	defer nsHandle.Close()

	//方便恢复，先存一下
	originNs, err := netns.Get()
	if err != nil {
		return fmt.Errorf("fail to get netns")
	}
	defer originNs.Close()
	defer netns.Set(originNs) //切回原来的

	//ns换过去
	if err = netns.Set(nsHandle); err != nil {
		return fmt.Errorf("fail to set netns")
	}

	peerNs, err := netlink.LinkByName(containerInterface)
	if err != nil {
		return fmt.Errorf("fail to get peer")
	}

	//docker的默认网段是172.17.0.0/16
	//这边选择最后一段在2-255中间是为了避免和172.17.0.1冲突嗯
	ip := fmt.Sprintf("172.17.0.%d/16", pid%254+2)
	addr, err := netlink.ParseAddr(ip)
	if err != nil {
		return fmt.Errorf("fail to parse ip")
	}
	//配置ip
	if err = netlink.AddrAdd(peerNs, addr); err != nil {
		return fmt.Errorf("fail to add ip")
	}

	if err = netlink.LinkSetUp(peerNs); err != nil {
		return fmt.Errorf("fail to set up")
	}

	defaultGateway := net.ParseIP("172.17.0.1")
	route := &netlink.Route{
		Dst: nil, //nil的话是默认的0.0.0.0/0
		Gw:  defaultGateway,
	}
	if err = netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("fail to add route")
	}

	slog.Info("created successful", "bridge", containerID)
	return nil
}
