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
		Cost:    route.Cost,
		Address: str2ipv4Num(route.Dest),
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

func UnmarshalEntry(bytes []byte) Entry {
	cost := uint32(binary.BigEndian.Uint32(bytes[:4]))
	address := uint32(binary.BigEndian.Uint32(bytes[4:8]))
	mask := uint32(binary.BigEndian.Uint32(bytes[8:]))
	entry := Entry{
		Cost:    cost,
		Address: address,
		Mask:    mask,
	}
	return entry
}

func str2ipv4Num(addr string) uint32 {
	numStrs := strings.Split(addr, ".")
	res := uint32(0)
	for _, numStr := range numStrs {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			log.Fatalln(err)
		}
		res = res<<8 + uint32(num)
		// fmt.Println(res)
	}
	return res
}

func ipv4Num2str(addr uint32) string {
	mask := 1<<8 - 1
	res := strconv.Itoa(int(addr) & mask)
	addr >>= 8
	for i := 0; i < 3; i++ {
		res = strconv.Itoa(int(addr)&mask) + "." + res
		addr >>= 8
	}
	return res
}
