package network

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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
					cli := NewCLI(LI, 0, []byte{})
					node.NodeCLIChan <- cli
				}
			} else if len(line) >= 2 && line[:2] == "lr" {
				if len(line) == 2 {
					cli := NewCLI(LR, 0, []byte{})
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
				cli := NewCLI(uint8(SetUpT), uint8(id), []byte{})
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
				cli := NewCLI(uint8(SetDownT), uint8(id), []byte{})
				node.NodeCLIChan <- cli
			} else if line == "q" {
				cli := NewCLI(Quit, 0, []byte{})
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
		case LI:
			node.PrintInterfaces()
			fmt.Printf("> ")
		case SetUpT:
			node.SetUp(cli.ID)
			fmt.Printf("> ")
		case SetDownT:
			node.SetDown(cli.ID)
			fmt.Printf("> ")
		case Quit:
			node.Quit()
			fmt.Printf("> ")
		case LR:
			node.PrintRoutes()
			fmt.Printf("> ")
		case RIPBroadcast:
			node.BroadcastRIP()
		case RIPHandle:
			node.HandleRIP(cli.Bytes)
		}
	}
}
