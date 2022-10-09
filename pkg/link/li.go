package link

import (
	"fmt"
	"log"
	"net"
)

type LinkInterface struct {
	ID        uint8
	MACLocal  *net.UDPAddr
	MACRemote *net.UDPAddr
	IPLocal   string
	IPRemote  string
	Status    string
	// we use RemoteConn to send data to remote machine
	LinkConn *net.UDPConn
}

func (li *LinkInterface) Make(udpIp, udpPortRemote, ipLocal, ipRemote string, id uint8, MACLocal string) {
	li.ID = id
	li.IPLocal = ipLocal
	li.IPRemote = ipRemote
	if li.IPLocal == "" {
		return
	}
	// LocalAddr
	// Setup RemoteConn
	remoteAddr, err := net.ResolveUDPAddr("udp", ToIPColonAddr(udpIp, udpPortRemote))
	li.MACRemote = remoteAddr
	if err != nil {
		log.Fatalln(err)
	}
	LocalAddr, err := net.ResolveUDPAddr("udp", MACLocal)
	li.MACLocal = LocalAddr
	if err != nil {
		log.Fatalln(err)
	}
	// fmt.Println(li.Addr, li.IpLocal, li.IpRemote)
	li.LinkConn, err = net.ListenUDP("udp", li.MACLocal)
	if err != nil {
		log.Fatalln(err)
	}
	li.Status = "up"
}

func ToIPColonAddr(udpIp, udpPort string) string {
	return fmt.Sprintf("%v:%v", udpIp, udpPort)
}
