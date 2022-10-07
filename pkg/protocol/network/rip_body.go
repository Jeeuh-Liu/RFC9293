package network

import (
	"bytes"
	"encoding/binary"
	"log"
)

type RIPBody struct {
	// command + num_entries = 4 bytes
	Command     uint16
	Num_Entries uint16
	// one entry = 12 bytes
	Entries []Entry
}

func (node *Node) NewRIPBody() *RIPBody {
	num_entries := len(node.Routes)
	entries := []Entry{}
	for _, route := range node.Routes {
		entry := NewEntry(route)
		entries = append(entries, entry)
	}
	body := &RIPBody{
		Command:     1,
		Num_Entries: uint16(num_entries),
		Entries:     entries,
	}
	return body
}

func (body *RIPBody) Marshal() []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, body.Command)
	if err != nil {
		log.Fatalln(err)
	}
	err = binary.Write(buf, binary.BigEndian, body.Num_Entries)
	if err != nil {
		log.Fatalln(err)
	}
	bytes := buf.Bytes()
	// fmt.Printf("Length of body [Command + Num_Entries] is %v bytes\n", len(bytes))
	for _, entry := range body.Entries {
		bytes = append(bytes, entry.Marshal()...)
		// fmt.Printf("Length of body [Command + Num_Entries] + entry is %v bytes\n", len(bytes))
	}
	// fmt.Printf("Length of body is %v bytes\n", len(bytes))
	return bytes
}
