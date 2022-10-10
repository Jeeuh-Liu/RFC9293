package link

import (
	"fmt"
	"log"
	"net"
)

type LinkInterface struct {
	ID        uint8
	MACLocal  string
	MACRemote string
	IPLocal   string
	IPRemote  string
	Status    string
	// we use RemoteConn to send data to remote machine
	LinkConn *net.UDPConn
}

func (li *LinkInterface) Make(udpIp, udpPortRemote, ipLocal, ipRemote string, id uint8, udpPortLocal string) {
	li.ID = id
	li.MACLocal = udpIp + ":" + udpPortLocal
	li.MACRemote = udpIp + ":" + udpPortRemote
	li.IPLocal = ipLocal
	li.IPRemote = ipRemote
	if li.IPLocal == "" {
		return
	}
	// LocalAddr
	// Setup RemoteConn
	remoteAddr, err := net.ResolveUDPAddr("udp", li.MACRemote)
	if err != nil {
		log.Fatalln(err)
	}
	localAddr, err := net.ResolveUDPAddr("udp", li.MACLocal)
	if err != nil {
		log.Fatalln(err)
	}
	li.MACRemote = remoteAddr.String()
	li.MACLocal = localAddr.String()
	// fmt.Println(li.MACLocal, li.MACRemote)
	linkConn, err := net.ListenUDP("udp", localAddr)
	li.LinkConn = linkConn
	if err != nil {
		log.Fatalln("Open LinkConn", err)
	}
	li.Status = "up"
	go li.ServeLink()
}

func ToIPColonAddr(udpIp, udpPort string) string {
	return fmt.Sprintf("%v:%v", udpIp, udpPort)
}
