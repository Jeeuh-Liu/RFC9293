package link

import (
	"fmt"
	"log"
	"net"
)

// Send bytes through link
func (li *LinkInterface) SendPacket(packetBytes []byte) {
	if li.Status != "up" {
		return
	}
	// fmt.Printf("Link try to send a RIP to %v\n", li.MACRemote)
	remoteAddr, err := net.ResolveUDPAddr("udp", li.MACRemote)
	if err != nil {
		log.Fatalln(err)
	}
	bnum, err := li.LinkConn.WriteToUDP(packetBytes, remoteAddr)
	if err != nil {
		log.Fatalln("sendRIP", err)
	}
	fmt.Printf("Send %v bytes\n", bnum)
}
