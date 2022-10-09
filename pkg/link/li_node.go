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
	bnum, err := li.LinkConn.WriteToUDP(bytes, li.MACRemote)
	if err != nil {
		log.Fatalln("sendRIP", err)
	}
	fmt.Printf("Send %v bytes\n", bnum)
}

// Handle commands to open and close of link
func (li *LinkInterface) OpenRemoteLink() {
	if li.Status == "up" {
		fmt.Printf("interface %v is already up\n", li.ID)
		return
	}
	// fmt.Println(li.Addr, li.IpLocal, li.IpRemote)
	linkConn, err := net.ListenUDP("udp", li.MACLocal)
	li.LinkConn = linkConn
	if err != nil {
		log.Fatalln("Open LinkConn", err)
	}
	li.Status = "up"
	fmt.Printf("interface %v is now enabled, Dial to udp %v\n", li.ID, li.MACRemote)
}

func (li *LinkInterface) CloseRemoteLink() {
	if li.Status == "down" {
		fmt.Printf("interface %v is already down\n", li.ID)
		return
	}
	li.LinkConn.Close()
	li.Status = "dn"
	fmt.Printf("interface %v is now disabled\n", li.ID)
}
