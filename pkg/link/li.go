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
	RemoteConn *net.UDPConn
}

func (li *LinkInterface) Make(udpIp, udpPortRemote, ipLocal, ipRemote string, id uint8) {
	li.ID = id
	li.MACRemote = ToIPColonAddr(udpIp, udpPortRemote)
	li.IPLocal = ipLocal
	li.IPRemote = ipRemote
	if li.IPLocal == "" {
		return
	}
	// LocalAddr
	// ipPort := strings.Split(MACLocal, ":")
	// port, _ := strconv.Atoi(ipPort[1])
	// localAddr := net.UDPAddr{
	// 	IP:   net.ParseIP(ipPort[0]),
	// 	Port: port,
	// }
	// Setup RemoteConn
	remoteAddr, err := net.ResolveUDPAddr("udp", li.MACRemote)
	if err != nil {
		log.Fatalln(err)
	}
	// fmt.Println(li.Addr, li.IpLocal, li.IpRemote)
	li.RemoteConn, err = net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		log.Fatalln(err)
	}
	li.Status = "up"
}

func ToIPColonAddr(udpIp, udpPort string) string {
	return fmt.Sprintf("%v:%v", udpIp, udpPort)
}
