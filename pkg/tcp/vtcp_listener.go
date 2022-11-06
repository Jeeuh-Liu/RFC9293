package tcp

import (
	"fmt"
	"tcpip/pkg/myDebug"
	"tcpip/pkg/proto"

	"github.com/google/netstack/tcpip/header"
)

type VTCPListener struct {
	ID          uint16
	state       string
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
	// go listener.acceptLoop()
	return listener
}

func (listener *VTCPListener) acceptLoop() error {
	for {
		segment := <-listener.AcceptQueue
		if segment.TCPhdr.Flags == header.TCPFlagSyn {
			// Check if socket for the application of the same port has been created
			myDebug.Debugln("socket listening on %v receives a request from %v:%v",
				listener.localPort, segment.IPhdr.Src.String(), segment.TCPhdr.SrcPort)
			conn := NewNormalSocket(segment)
			fmt.Println(conn.seqNum)
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
