package ip

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

func (node *Node) ScanClI() {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "li" {
				cli := NewCLI(LI, 0)
				node.NodeCLIChan <- cli
			} else if line == "q" {
				cli := NewCLI(Quit, 0)
				node.NodeCLIChan <- cli
			} else if len(strings.Split(line, " ")) >= 2 {
				cmds := strings.Split(line, " ")
				if cmds[0] == "up" {
					// open link
					id, err := strconv.Atoi(cmds[1])
					if err != nil {
						log.Fatalln(err)
					}
					cli := NewCLI(uint8(SetUpT), uint8(id))
					node.NodeCLIChan <- cli
				} else if cmds[0] == "down" {
					// close link
					id, err := strconv.Atoi(cmds[1])
					if err != nil {
						log.Fatalln(err)
					}
					cli := NewCLI(uint8(SetDownT), uint8(id))
					node.NodeCLIChan <- cli
				}
			}
		}
	}
}
