package network

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"tcpip/pkg/proto"
	"time"
)

// ScanCLI
/*
	When to output '> ' ?
	(1) After sending a cli to chan and  handling the cli, we can output a '>'
	(2) If the command is invalid and we cannot send it to chan, we output a '>' after error msg
*/

func (node *Node) ScanClI() {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) >= 2 && line[:2] == "li" {
				if len(line) == 2 {
					cli := proto.NewCLI(proto.LI, 0, []byte{}, "")
					node.NodeCLIChan <- cli
				}
			} else if len(line) >= 2 && line[:2] == "lr" {
				if len(line) == 2 {
					cli := proto.NewCLI(proto.LR, 0, []byte{}, "")
					node.NodeCLIChan <- cli
				}

			} else if len(line) >= 2 && len(strings.Split(line, " ")) == 2 && line[:2] == "up" {
				cmds := strings.Split(line, " ")
				id, err := strconv.Atoi(cmds[1])
				if err != nil {
					fmt.Printf("strconv.Atoi: parsing %v: invalid syntax\n> ", cmds[1])
					continue
				}
				if id >= len(node.ID2Interface) {
					fmt.Printf("interface %v does not exist\n> ", id)
					continue
				}
				// open link
				cli := proto.NewCLI(uint8(proto.SetUpT), uint8(id), []byte{}, "")
				node.NodeCLIChan <- cli
			} else if len(line) >= 4 && len(strings.Split(line, " ")) == 2 && line[:4] == "down" {
				cmds := strings.Split(line, " ")
				id, err := strconv.Atoi(cmds[1])
				if err != nil {
					fmt.Printf("strconv.Atoi: parsing %v: invalid syntax\n> ", cmds[1])
					continue
				}
				if id >= len(node.ID2Interface) {
					fmt.Printf("interface %v does not exist\n> ", id)
					continue
				}
				// close link
				cli := proto.NewCLI(uint8(proto.SetDownT), uint8(id), []byte{}, "")
				node.NodeCLIChan <- cli
			} else if line == "q" {
				cli := proto.NewCLI(proto.Quit, 0, []byte{}, "")
				node.NodeCLIChan <- cli
			} else {
				fmt.Printf("Invalid command\n> ")
			}
		}
	}
}

// ******************************************************************
// Output the data of CLI
func (node *Node) HandleCLI() {
	fmt.Printf("> ")
	for {
		cli := <-node.NodeCLIChan
		switch cli.CLIType {
		case proto.LI:
			node.HandlePrintInterfaces()
			fmt.Printf("> ")
		case proto.SetUpT:
			node.HandleSetUp(cli.ID)
			fmt.Printf("> ")
		case proto.SetDownT:
			node.HandleSetDown(cli.ID)
			fmt.Printf("> ")
		case proto.Quit:
			node.HandleQuit()
			fmt.Printf("> ")
		case proto.LR:
			node.HandlePrintRoutes()
			fmt.Printf("> ")
		case proto.TypeBroadcastRIPResp:
			node.HandleBroadcastRIPResp()
		case proto.TypeBroadcastRIPReq:
			node.HandleBroadcastRIPReq()
		case proto.TypeHandlePacket:
			node.HandlePacket(cli.Bytes)
		case proto.TypeHandleRIPResp:
			node.HandleRIPResp(cli.Bytes)
		case proto.TypeRouteEx:
			node.HandleRouteEx(cli.DestIP)
		}
	}
}

// Broadcast RIP through LinkInterface
func (node *Node) RIPRespDaemon() {
	for {
		cli := proto.NewCLI(proto.TypeBroadcastRIPResp, 0, []byte{}, "")
		node.NodeCLIChan <- cli
		time.Sleep(5 * time.Second)
	}
}

func (node *Node) RIPReqDaemon() {
	cli := proto.NewCLI(proto.TypeBroadcastRIPReq, 0, []byte{}, "")
	node.NodeCLIChan <- cli
}

// Route Ex timeout
func (node *Node) SendExTimeCLI(destIP string) {
	// sleep 13 second and check whether the time expires
	time.Sleep(13 * time.Second)
	cli := proto.NewCLI(proto.TypeRouteEx, 0, []byte{}, destIP)
	node.NodeCLIChan <- cli
}
