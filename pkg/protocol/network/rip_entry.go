package network

import (
	"bytes"
	"encoding/binary"
	"log"
	"strconv"
	"strings"
)

type Entry struct {
	// 12 byte
	Cost    uint32
	Address uint32
	Mask    uint32
}

func NewEntry(route Route) Entry {
	entry := Entry{
		Cost:    route.Cost + 1,
		Address: str2ipv4(route.Dest),
		Mask:    1<<32 - 1,
	}
	return entry
}

func (entry Entry) Marshal() []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, entry.Cost)
	if err != nil {
		log.Fatalln(err)
	}
	err = binary.Write(buf, binary.BigEndian, entry.Address)
	if err != nil {
		log.Fatalln(err)
	}
	err = binary.Write(buf, binary.BigEndian, entry.Mask)
	if err != nil {
		log.Fatalln(err)
	}
	bytes := buf.Bytes()
	return bytes
}

func str2ipv4(addr string) uint32 {
	numStrs := strings.Split(addr, ".")
	res := uint32(0)
	for _, numStr := range numStrs {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			log.Fatalln(err)
		}
		res = res<<8 + uint32(num)
	}
	return res
}
