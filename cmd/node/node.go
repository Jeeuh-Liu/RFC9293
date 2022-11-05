package main

import (
	"log"
	"os"
	"tcpip/pkg/network"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Usage: node + inx file")
	}
	node := &network.Node{}
	node.Make(os.Args)
	// node.HandleCLI()
	node.ReceiveOpFromChan()
}
