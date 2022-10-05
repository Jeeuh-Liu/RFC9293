package link

import (
	"fmt"
	"net"
)

type Interface struct {
	ID       uint8
	Addr     string
	IpLocal  string
	IpRemote string
	Status   string
	LinkConn *net.UDPConn
}

func (li *Interface) Make(id uint8, udpIp, udpPort, ipLocal, ipRemote string) {
	li.ID = id
	li.Addr = ToIpColonAddr(udpIp, udpPort)
	li.IpLocal = ipLocal
	li.IpRemote = ipRemote
	li.Status = "up"
	if li.IpLocal == "" {
		return
	}
	go li.ServeLink()
}

func ToIpColonAddr(udpIp, udpPort string) string {
	return fmt.Sprintf("%v:%v", udpIp, udpPort)
}
