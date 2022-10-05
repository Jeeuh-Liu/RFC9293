package ip

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

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
					cli := NewCLI(LI, 0)
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
				cli := NewCLI(uint8(SetUpT), uint8(id))
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
				cli := NewCLI(uint8(SetDownT), uint8(id))
				node.NodeCLIChan <- cli
			} else if line == "q" {
				cli := NewCLI(Quit, 0)
				node.NodeCLIChan <- cli
			} else {
				fmt.Printf("Invalid command\n> ")
			}
		}
	}
}
