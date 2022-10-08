package link

import (
	"fmt"
	"log"
	"net"
)

// Send bytes through link
func (li *LinkInterface) SendRIP(bytes []byte) {
	if li.Status != "up" {
		return
	}
	// fmt.Printf("Link try to send a RIP to %v\n", li.MACRemote)
	li.RemoteConn.Write(bytes)
}

// Handle commands to open and close of link
func (li *LinkInterface) OpenRemoteLink() {
	if li.Status == "up" {
		fmt.Printf("interface %v is already up\n", li.ID)
		return
	}
	addr, err := net.ResolveUDPAddr("udp", li.MACRemote)
	if err != nil {
		log.Fatalln(err)
	}
	// fmt.Println(li.Addr, li.IpLocal, li.IpRemote)
	li.RemoteConn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalln(err)
	}
	li.Status = "up"
	fmt.Printf("interface %v is now enabled, Dial to udp %v\n", li.ID, li.MACRemote)
}

func (li *LinkInterface) CloseRemoteLink() {
	if li.Status == "down" {
		fmt.Printf("interface %v is already down\n", li.ID)
		return
	}
	li.RemoteConn.Close()
	li.Status = "dn"
	fmt.Printf("interface %v is now disabled\n", li.ID)
}
