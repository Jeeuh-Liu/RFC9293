package tcp

import (
	"fmt"
	"tcpip/pkg/proto"
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
		conn := NewNormalSocket(packet)
		listener.spawnChan <- conn
	}
}

func (listener *VTCPListener) VAccept() (*VTCPConn, error) {
	conn := <-listener.spawnChan
	if conn == nil {
		return nil, fmt.Errorf("fail to produce a new socket")
	}
	return conn, nil
}
