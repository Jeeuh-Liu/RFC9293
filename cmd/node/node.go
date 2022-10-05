package main

import (
	"fmt"
	"log"
	"os"
	"tcpip/pkg/ip"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Usage: node + inx file")
	}
	node := &ip.Node{}
	node.Make(os.Args)

	for {
		fmt.Printf("> ")
		cli := <-node.NodeCLIChan
		switch cli.CLIType {
		case ip.LI:
			node.PrintInterfaces()
		case ip.SetUpT:
			node.SetUp(cli.ID)
		case ip.SetDownT:
			node.SetDown(cli.ID)
		case ip.Quit:
			node.Quit()
		}
	}
}
