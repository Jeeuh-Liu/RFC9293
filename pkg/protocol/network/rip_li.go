package network

import (
	"fmt"
	"time"
)

// Broadcast RIP
func (node *Node) RIPDaemon() {
	for {
		fmt.Println("Try to broadcast RIP")
		rip := node.NewRIP()
		bytes := rip.Marshal()
		for _, li := range node.ID2Interface {
			li.SendRIP(bytes)
		}
		time.Sleep(5 * time.Second)
	}
}
