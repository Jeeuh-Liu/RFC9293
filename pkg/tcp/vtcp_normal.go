package tcp

import (
	"fmt"
	"math/rand"
	"net"
	"tcpip/pkg/myDebug"
	"tcpip/pkg/proto"

	"github.com/google/netstack/tcpip/header"
)

const (
	MINACKNUM         = 1
	DEFAULTDATAOFFSET = 20
	DEFAULTWINDOWSIZE = 65535
)

type VTCPConn struct {
	state           string
	seqNum          uint32
	ackNum          uint32
	LocalAddr       net.IP
	LocalPort       uint16
	RemoteAddr      net.IP
	RemotePort      uint16
	windowSize      uint16
	ID              uint16
	SegRcvChan      chan *proto.Segment
	NodeSegSendChan chan *proto.Segment
}

func NewNormalSocket(seqNum uint32, dstPort, srcPort uint16, dstIP, srcIP net.IP) *VTCPConn {
	conn := &VTCPConn{
		state:      proto.SYN_RECV,
		seqNum:     rand.Uint32(),
		ackNum:     seqNum + 1, // seqNum can be 0 if created by cli
		LocalPort:  srcPort,
		LocalAddr:  srcIP,
		RemoteAddr: dstIP,
		RemotePort: dstPort,
		windowSize: DEFAULTWINDOWSIZE,
		SegRcvChan: make(chan *proto.Segment),
	}
	return conn
}

// ********************************************************************************************
// Server
// Q: handle RST-?
// TODO: resubmission
func (conn *VTCPConn) SynRev() {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck), []byte{})
	conn.NodeSegSendChan <- seg
	for {
		segRev := <-conn.SegRcvChan
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			conn.state = proto.ESTABLISH
			return
		}
	}
}

// ********************************************************************************************
// Client
func (conn *VTCPConn) SynSend() {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn), []byte{})
	conn.NodeSegSendChan <- seg
	myDebug.Debugln("%v sent connection request to %v, SEQ: %v", conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.seqNum)
	for {
		segRev := <-conn.SegRcvChan
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			conn.seqNum++
			conn.ackNum = segRev.TCPhdr.SeqNum + 1
			conn.send([]byte{})
			conn.state = proto.ESTABLISH
			return
		}
	}
}

func (conn *VTCPConn) send(content []byte) {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck), content)
	conn.NodeSegSendChan <- seg
	myDebug.Debugln("%v sent segment to %v, SEQ: %v, ACK: %v, Payload: %v",
		conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.seqNum, conn.ackNum, string(seg.Payload))
}

func (conn *VTCPConn) SimpleSend(content []byte) {
	mtu := proto.DEFAULTPACKETMTU - proto.DEFAULTIPHDRLEN - proto.DEFAULTTCPHDRLEN
	for len(content) > 0 {
		var payload []byte
		if len(content) <= mtu {
			payload = content
			content = []byte{}
		} else {
			payload = content[:mtu]
			content = content[mtu:]
		}
		conn.send(payload)
		conn.seqNum += uint32(len(payload))
		segRev := <-conn.SegRcvChan
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			conn.seqNum++
			conn.ackNum = segRev.TCPhdr.SeqNum + 1
		}
	}
}

// ********************************************************************************************
// helper function
func (conn *VTCPConn) GetTuple() string {
	return fmt.Sprintf("%v:%v:%v:%v", conn.RemoteAddr.String(), conn.RemotePort,
		conn.LocalAddr.String(), conn.LocalPort)
}

func (conn *VTCPConn) buildTCPHdr(flags int) *header.TCPFields {
	return &header.TCPFields{
		SrcPort:       conn.LocalPort,
		DstPort:       conn.RemotePort,
		SeqNum:        conn.seqNum,
		AckNum:        conn.ackNum,
		DataOffset:    DEFAULTDATAOFFSET,
		Flags:         uint8(flags),
		WindowSize:    conn.windowSize,
		Checksum:      0,
		UrgentPointer: 0,
	}
}
