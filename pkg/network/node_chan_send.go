package network

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"tcpip/pkg/proto"
	"time"
)

// Scan NodeCLI
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
			ws := strings.Split(line, " ")
			// fmt.Println(ws, len(ws), ws[0])
			if (len(ws) == 1 || len(ws) == 2) && ws[0] == "li" {
				if len(ws) == 1 {
					cli := proto.NewNodeCLI(proto.LI, 0, []byte{}, "", 0, "")
					node.NodeCLIChan <- cli
				} else {
					// print li to a file
				}
			} else if (len(ws) == 1 || len(ws) == 2) && ws[0] == "lr" {
				if len(ws) == 1 {
					cli := proto.NewNodeCLI(proto.LR, 0, []byte{}, "", 0, "")
					node.NodeCLIChan <- cli
				} else {
					// print lr to a file
				}
			} else if len(ws) == 2 && ws[0] == "up" {
				id, err := strconv.Atoi(ws[1])
				if err != nil {
					fmt.Printf("strconv.Atoi: parsing %v: invalid syntax\n> ", ws[1])
					continue
				}
				if id >= len(node.ID2Interface) {
					fmt.Printf("interface %v does not exist\n> ", id)
					continue
				}
				// open link
				cli := proto.NewNodeCLI(uint8(proto.SetUpT), uint8(id), []byte{}, "", 0, "")
				node.NodeCLIChan <- cli
			} else if len(ws) == 2 && ws[0] == "down" {
				id, err := strconv.Atoi(ws[1])
				if err != nil {
					fmt.Printf("strconv.Atoi: parsing %v: invalid syntax\n> ", ws[1])
					continue
				}
				if id >= len(node.ID2Interface) {
					fmt.Printf("interface %v does not exist\n> ", id)
					continue
				}
				// close link
				cli := proto.NewNodeCLI(uint8(proto.SetDownT), uint8(id), []byte{}, "", 0, "")
				node.NodeCLIChan <- cli
			} else if len(ws) == 1 && ws[0] == "q" {
				cli := proto.NewNodeCLI(proto.Quit, 0, []byte{}, "", 0, "")
				node.NodeCLIChan <- cli
			} else if len(ws) >= 4 && ws[0] == "send" {
				destIP := net.ParseIP(ws[1]).String()
				protoID, err := strconv.Atoi(ws[2])
				if err != nil {
					fmt.Printf("strconv.Atoi: parsing %v: invalid syntax\n> ", ws[1])
					continue
				}
				msg := line[len(ws[0])+len(ws[1])+len(ws[2])+3:]
				// fmt.Println(msg)
				cli := proto.NewNodeCLI(proto.TypeSendPacket, 0, []byte{}, destIP, protoID, msg)
				node.NodeCLIChan <- cli
			} else {
				fmt.Printf("Invalid command\n> ")
			}
		}
	}
}

// Send NodeBroadcast
func (node *Node) RIPRespDaemon() {
	for {
		cli := proto.NewNodeBC(proto.TypeBroadcastRIPResp, 0, []byte{}, "", 0, "")
		node.NodeBCChan <- cli
		time.Sleep(5 * time.Second)
	}
}

func (node *Node) RIPReqDaemon() {
	cli := proto.NewNodeBC(proto.TypeBroadcastRIPReq, 0, []byte{}, "", 0, "")
	node.NodeBCChan <- cli
}

// Send NodeEx
func (node *Node) SendExTimeCLI(destIP string) {
	// sleep 12 second and check whether the time expires
	time.Sleep(12 * time.Second)
	cli := proto.NewNodeEx(proto.TypeRouteEx, 0, []byte{}, destIP, 0, "")
	node.NodeExChan <- cli
}

// Send NodePktOp
func (node *Node) BroadcastRIPRespTU(entity proto.Entry) {
	for _, li := range node.ID2Interface {
		if !li.IsUp() {
			continue
		}
		entries := []proto.Entry{}
		entries = append(entries, entity)
		rip := proto.NewPktRIP(li.IPLocal, li.IPRemote, 2, entries)
		bytes := rip.Marshal()
		li.SendPacket(bytes)
	}
}
