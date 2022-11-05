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
	ackNum     uint32
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
		seqNum:     0xd599,
		ackNum:     pkt.TCPhdr.SeqNum + 1,
		LocalPort:  pkt.TCPhdr.DstPort,
		LocalAddr:  pkt.IPhdr.Dst,
		RemoteAddr: pkt.IPhdr.Src,
		RemotePort: pkt.TCPhdr.SrcPort,
		windowSize: DEFAULTWINDOWSIZE,
		Buffer:     make(chan *proto.Segment),
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
		if conn.ackNum == rev.TCPhdr.SeqNum+1 {
			conn.seqNum++
			conn.state = proto.ESTABLISH
			ack := &proto.Segment{
				TCPhdr:  conn.buildTCPHdr(),
				Payload: []byte{},
			}
			conn.Upstream <- &proto.SegmentMsg{
				SocketID: conn.ID,
				Seg:      ack,
			}
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
		AckNum:        conn.ackNum,
		DataOffset:    DEFAULTDATAOFFSET,
		Flags:         header.TCPFlagAck,
		WindowSize:    conn.windowSize,
		Checksum:      0,
		UrgentPointer: 0,
	}
}