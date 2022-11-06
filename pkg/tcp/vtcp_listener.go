package tcp

import (
	"fmt"
	"tcpip/pkg/myDebug"
	"tcpip/pkg/proto"
)

type VTCPListener struct {
	ID        uint16
	state     string
	localPort uint16
	// listern.acceptLoop(), conn => spawnCh => VAccpet()
	ConnQueue chan *VTCPConn
	// listen.
	SegRcvChan chan *proto.Segment
}

func NewListener(port uint16) *VTCPListener {
	listener := &VTCPListener{
		localPort:  port,
		state:      proto.LISTENER,
		ConnQueue:  make(chan *VTCPConn),
		SegRcvChan: make(chan *proto.Segment),
	}
	go listener.VListenerAcceptLoop()
	return listener
}

func (listener *VTCPListener) VListenerAcceptLoop() error {
	for {
		segment := <-listener.SegRcvChan
		myDebug.Debugln("socket listening on %v receives a request from %v:%v",
			listener.localPort, segment.IPhdr.Src.String(), segment.TCPhdr.SrcPort)
		conn := NewNormalSocket(segment)
		// fmt.Println(conn.seqNum)
		listener.ConnQueue <- conn
	}
}

func (listener *VTCPListener) VAccept() (*VTCPConn, error) {
	conn := <-listener.ConnQueue
	if conn == nil {
		return nil, fmt.Errorf("fail to produce a new socket")
	}
	return conn, nil
}
