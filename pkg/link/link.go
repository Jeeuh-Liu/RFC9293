package link

import (
	"fmt"
	"log"
	"net"
)

func (li *Interface) ServeLink() {
	addr, err := net.ResolveUDPAddr("udp", li.Addr)
	if err != nil {
		log.Fatalln(err)
	}
	// fmt.Println(li.Addr, li.IpLocal, li.IpRemote)
	li.LinkConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	bytes := make([]byte, 1400)
	li.LinkConn.Read(bytes)
}

func (li *Interface) OpenLink() {
	if li.Status == "up" {
		fmt.Printf("interface %v is already up\n", li.ID)
		return
	}
	go li.ServeLink()
	fmt.Printf("interface %v is now enabled\n", li.ID)
}

func (li *Interface) CloseLink() {
	if li.Status == "down" {
		fmt.Printf("interface %v is already down\n", li.ID)
		return
	}
	li.Status = "dn"
	li.LinkConn.Close()
	fmt.Printf("interface %v is now disabled\n", li.ID)
}
