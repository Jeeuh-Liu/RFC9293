package link

import (
	"fmt"
	"net"
)

type LinkInterface struct {
	ID        uint8
	MACRemote string
	IPLocal   string
	IPRemote  string
	Status    string
	// we use RemoteConn to send data to remote machine
	RemoteConn *net.UDPConn
}

func (li *LinkInterface) Make(id uint8, udpIp, udpPortRemote, ipLocal, ipRemote string) {
	li.ID = id
	li.MACRemote = ToIPColonAddr(udpIp, udpPortRemote)
	li.IPLocal = ipLocal
	li.IPRemote = ipRemote
	if li.IPLocal == "" {
		return
	}
	li.OpenRemoteLink()
	li.Status = "up"
}

func ToIPColonAddr(udpIp, udpPort string) string {
	return fmt.Sprintf("%v:%v", udpIp, udpPort)
}
