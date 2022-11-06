package tcp

import (
	"fmt"
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

func NewNormalSocket(pkt *proto.Segment) *VTCPConn {
	conn := &VTCPConn{
		state: proto.SYN_RECV,
		// seqNum: rand.Uint32(),
		seqNum:     0xd599,
		ackNum:     pkt.TCPhdr.SeqNum + 1,
		LocalPort:  pkt.TCPhdr.DstPort,
		LocalAddr:  pkt.IPhdr.Dst,
		RemoteAddr: pkt.IPhdr.Src,
		RemotePort: pkt.TCPhdr.SrcPort,
		windowSize: DEFAULTWINDOWSIZE,
		SegRcvChan: make(chan *proto.Segment),
	}
	return conn
}

// Q: handle RST-?
// TODO: resubmission
func (conn *VTCPConn) VTCPConnSynHandler() {
	myDebug.Debugln("connection with %v:%v enter syn_recv state",
		conn.RemoteAddr.String(), conn.RemotePort)
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck), []byte{})
	fmt.Printf("Conn's Seq Num is %v\n", conn.seqNum)
	conn.NodeSegSendChan <- seg
	go conn.VTCPConnSegHandler()
}

func (conn *VTCPConn) VTCPConnSegHandler() {
	for {
		segRev := <-conn.SegRcvChan
		fmt.Println(conn.seqNum, segRev.TCPhdr.AckNum)
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			conn.seqNum++
			conn.state = proto.ESTABLISH
			seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck), []byte{})
			conn.NodeSegSendChan <- seg
		}
	}
}

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
