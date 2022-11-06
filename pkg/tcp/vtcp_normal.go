package tcp

import (
	"fmt"
	"math/rand"
	"net"
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
func (conn *VTCPConn) VTCPConnSynRev() {
	// myDebug.Debugln("connection with %v:%v enter syn_recv state",
	// conn.RemoteAddr.String(), conn.RemotePort)
	conn.VTCPConnSynAckSend()
	go conn.VTCPConnAckRev()
}

func (conn *VTCPConn) VTCPConnSynAckSend() {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck), []byte{})
	conn.NodeSegSendChan <- seg
	fmt.Printf("[SYNACK] Push 1 msg into SegSendChan\n")
}

func (conn *VTCPConn) VTCPConnAckRev() {
	for {
		segRev := <-conn.SegRcvChan
		fmt.Println(conn.seqNum, segRev.TCPhdr.AckNum, segRev.TCPhdr.Flags)
		// fmt.Println(conn.seqNum, segRev.TCPhdr.AckNum)
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			conn.seqNum++
			conn.state = proto.ESTABLISH
			// seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck), []byte{})
			// conn.NodeSegSendChan <- seg
			return
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

// ********************************************************************************************
// Client
func (conn *VTCPConn) VTCPConnSynSend() {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn), []byte{})
	conn.NodeSegSendChan <- seg
	fmt.Printf("[SYN] Push 1 msg into SegSendChan\n")
	go conn.VTCPConnSYNAckRev()
}

func (conn *VTCPConn) VTCPConnSYNAckRev() {
	for {
		segRev := <-conn.SegRcvChan
		fmt.Println(conn.seqNum, segRev.TCPhdr.AckNum, segRev.TCPhdr.Flags)
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			conn.seqNum++
			conn.state = proto.ESTABLISH
			conn.ackNum = segRev.TCPhdr.SeqNum + 1
			seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck), []byte{})
			conn.NodeSegSendChan <- seg
			fmt.Printf("[ACK] Push 1 msg into SegSendChan\n")
			return
		}
	}
}
