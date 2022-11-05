package tcp

import (
	"fmt"
	"tcpip/pkg/myDebug"
	"tcpip/pkg/proto"

	"github.com/google/netstack/tcpip/header"
)

type VTCPListener struct {
	ID          uint16
	state       uint8
	localPort   uint16
	spawnChan   chan *VTCPConn
	AcceptQueue chan *proto.Segment
}

func NewListener(port uint16) *VTCPListener {
	listener := &VTCPListener{
		localPort:   port,
		state:       proto.LISTENER,
		spawnChan:   make(chan *VTCPConn),
		AcceptQueue: make(chan *proto.Segment),
	}
	go listener.acceptLoop()
	return listener
}

func (listener *VTCPListener) acceptLoop() error {
	for {
		packet := <-listener.AcceptQueue
		if packet.TCPhdr.Flags == header.TCPFlagSyn {
			myDebug.Debugln("socket listening on %v receives a request from %v:%v",
				listener.localPort, packet.IPhdr.Src.String(), packet.TCPhdr.SrcPort)
			conn := NewNormalSocket(packet)
			listener.spawnChan <- conn
		}
	}
}

func (listener *VTCPListener) VAccept() (*VTCPConn, error) {
	conn := <-listener.spawnChan
	if conn == nil {
		return nil, fmt.Errorf("fail to produce a new socket")
	}
	return conn, nil
}
