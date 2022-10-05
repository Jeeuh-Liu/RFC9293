package ip

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"tcpip/pkg/link"
)

type Node struct {
	ID2Interface map[uint8]*link.Interface
	NodeCLIChan  chan *CLI
}

func (node *Node) Make(args []string) {
	node.ID2Interface = make(map[uint8]*link.Interface)
	inx := args[1]
	f, err := os.Open(inx)
	if err != nil {
		log.Fatalln(err)
	}
	r := bufio.NewReader(f)

	id := uint8(0)
	for {
		bytes, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln(err)
		}
		eles := strings.Split(string(bytes), " ")
		li := &link.Interface{}
		if len(eles) == 2 {
			continue
		}
		li.Make(id, eles[0], eles[1], eles[2], eles[3])
		fmt.Printf("%v: %v\n", id, eles[2])
		node.ID2Interface[id] = li
		id++
	}
	// fmt.Println(node)
	node.NodeCLIChan = make(chan *CLI)
	go node.ScanClI()
}
