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
	state      uint8
	seqNum     uint32
	expectACK  uint32
	LocalAddr  net.IP
	LocalPort  uint16
	RemoteAddr net.IP
	RemotePort uint16
	windowSize uint16
	ID         uint16
	Buffer     chan *proto.Segment
	Upstream   chan *proto.SegmentMsg
}

func NewNormalSocket(pkt *proto.Segment) *VTCPConn {
	return &VTCPConn{
		state: proto.SYN_RECV,
		// seqNum:     rand.Uint32(),
		seqNum:     0,
		expectACK:  MINACKNUM,
		LocalPort:  pkt.TCPhdr.DstPort,
		LocalAddr:  pkt.IPhdr.Dst,
		RemoteAddr: pkt.IPhdr.Src,
		RemotePort: pkt.TCPhdr.SrcPort,
		windowSize: DEFAULTWINDOWSIZE,
	}
}

// Q: handle RST-?
// TODO: resubmission
func (conn *VTCPConn) SynRecv() {
	myDebug.Debugln("connection with %v:%v enter syn_recv state",
		conn.RemoteAddr.String(), conn.RemotePort)
	ack := &proto.Segment{
		TCPhdr:  conn.buildTCPHdr(),
		Payload: []byte{},
	}
	ack.TCPhdr.Flags |= header.TCPFlagSyn
	conn.Upstream <- &proto.SegmentMsg{
		SocketID: conn.ID,
		Seg:      ack,
	}
	for {
		rev := <-conn.Buffer
		if rev.TCPhdr.SeqNum == conn.seqNum+1 {
			conn.state = proto.ESTABLISH
			conn.seqNum += 2
			return
		}
	}
}

func (conn *VTCPConn) GetTuple() string {
	return fmt.Sprintf("%v:%v:%v:%v", conn.RemoteAddr.String(), conn.RemotePort,
		conn.LocalAddr.String(), conn.LocalPort)
}

func (conn *VTCPConn) buildTCPHdr() *header.TCPFields {
	return &header.TCPFields{
		SrcPort:       conn.LocalPort,
		DstPort:       conn.RemotePort,
		SeqNum:        conn.seqNum,
		AckNum:        conn.expectACK,
		DataOffset:    DEFAULTDATAOFFSET,
		Flags:         header.TCPFlagAck,
		WindowSize:    conn.windowSize,
		Checksum:      0,
		UrgentPointer: 0,
	}
}
