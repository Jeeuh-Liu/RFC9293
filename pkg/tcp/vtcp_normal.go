package tcp

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"tcpip/pkg/myDebug"
	"tcpip/pkg/proto"
	"time"

	"github.com/google/netstack/tcpip/header"
)

const (
	MINACKNUM         = 1
	DEFAULTDATAOFFSET = 20
	DEFAULTWINDOWSIZE = 65535
)

type VTCPConn struct {
	mu              sync.Mutex
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
	// Send Buffer
	sCv sync.Cond
	sb  *SendBuffer // send buffer
	// Retransmission
	rtmQueue      chan *proto.Segment  // retransmission queue
	seq2timestamp map[uint32]time.Time // seq # of 1 segment to expiration time
}

func NewNormalSocket(seqNumber uint32, dstPort, srcPort uint16, dstIP, srcIP net.IP) *VTCPConn {
	conn := &VTCPConn{
		mu:         sync.Mutex{},
		state:      proto.SYN_RECV,
		seqNum:     rand.Uint32(),
		ackNum:     seqNumber + 1, // first ackNum can be 1 by giving seqNumber 0 (client --> NConn)
		LocalPort:  srcPort,
		LocalAddr:  srcIP,
		RemoteAddr: dstIP,
		RemotePort: dstPort,
		windowSize: DEFAULTWINDOWSIZE,
		SegRcvChan: make(chan *proto.Segment),
		// Retransmission
		rtmQueue:      make(chan *proto.Segment),
		seq2timestamp: make(map[uint32]time.Time),
	}
	go conn.retransmissionLoop()
	return conn
}

// ********************************************************************************************
// Client
func (conn *VTCPConn) SynSend() {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn, conn.seqNum), []byte{})
	conn.NodeSegSendChan <- seg
	myDebug.Debugln("%v sent connection request to %v, SEQ: %v", conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.seqNum)
	// Rev Syn+ACK
	for {
		segRev := <-conn.SegRcvChan
		myDebug.Debugln("[Threeway Handshake - Rev SYN+ACK] %v:%v receive packet from %v:%v, SEQ: %v, ACK %v",
			conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
			conn.RemotePort, segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum)
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			conn.seqNum = segRev.TCPhdr.AckNum
			conn.ackNum = segRev.TCPhdr.SeqNum + 1
			// Send ACK
			conn.send([]byte{}, conn.seqNum)
			conn.state = proto.ESTABLISH
			// create send buffer
			conn.sb = NewSendBuffer(conn.seqNum, uint32(segRev.TCPhdr.WindowSize))
			conn.sCv = *sync.NewCond(&conn.mu)
			go conn.VSBufferSend()
			go conn.VSBufferRcv()
			return
		}
	}
}

// ********************************************************************************************
// Server
// Q: handle RST-?
// TODO: resubmission
func (conn *VTCPConn) SynRev() {
	// Send Syn + Ack
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck, conn.seqNum), []byte{})
	conn.NodeSegSendChan <- seg
	// Rev Ack
	for {
		segRev := <-conn.SegRcvChan
		myDebug.Debugln("[Threeway Handshake - Rev ACK] %v:%v receive packet from %v:%v, SEQ: %v, ACK %v",
			conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
			conn.RemotePort, segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum)
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			conn.state = proto.ESTABLISH
			// create send buffer
			conn.sb = NewSendBuffer(conn.seqNum, uint32(segRev.TCPhdr.WindowSize))
			conn.sCv = *sync.NewCond(&conn.mu)
			go conn.VSBufferSend()
			go conn.VSBufferRcv()
			return
		}
	}
}

// ********************************************************************************************
// Send TCP Packet through Established Normal Conn

// func (conn *VTCPConn) SimpleSend(content []byte) {
// 	mtu := proto.DEFAULTPACKETMTU - proto.DEFAULTIPHDRLEN - proto.DEFAULTTCPHDRLEN
// 	for len(content) > 0 {
// 		var payload []byte
// 		if len(content) <= mtu {
// 			payload = content
// 			content = []byte{}
// 		} else {
// 			payload = content[:mtu]
// 			content = content[mtu:]
// 		}
// 		conn.send(payload, conn.seqNum)
// 		conn.seqNum += uint32(len(payload))
// 		segRev := <-conn.SegRcvChan
// 		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
// 			conn.seqNum++
// 			conn.ackNum = segRev.TCPhdr.SeqNum + 1
// 		}
// 	}
// }

func (conn *VTCPConn) VSBufferWrite(content []byte) {
	bnum := conn.sb.WriteIntoBuffer(content)
	myDebug.Debugln("[VSBufferWrite] %v writes %v bytes into send buffer, send buffer remaining bytes %v\n", conn.LocalAddr.String(), bnum, conn.sb.GetRemainBytes())
	// fmt.Printf("Send buffer becomes %v\n", conn.sb.buffer)
	conn.sCv.Signal()
}

func (conn *VTCPConn) VSBufferSend() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	mtu := proto.DEFAULTPACKETMTU - proto.DEFAULTIPHDRLEN - proto.DEFAULTTCPHDRLEN
	for conn.state == proto.ESTABLISH {
		if conn.sb.CanSend() {
			// Get one segment, send it out and add it to retransmission queue
			payload, seqNum := conn.sb.UpdateNxt(mtu)
			conn.send(payload, seqNum)
		} else {
			conn.sCv.Wait()
		}
	}
}

func (conn *VTCPConn) VSBufferRcv() {
	for {
		ack := <-conn.SegRcvChan
		conn.mu.Lock()
		myDebug.Debugln("[SendBuffer_RevACK] %v:%v receive packet from %v:%v, SEQ: %v, ACK %v",
			conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
			conn.RemotePort, ack.TCPhdr.SeqNum, ack.TCPhdr.AckNum)
		conn.sb.UpdateUNA(ack)
		conn.seqNum = ack.TCPhdr.AckNum
		myDebug.Debugln("[SendBuffer_RevACK] %v send buffer remaing bytes %v", conn.LocalAddr.String(), conn.sb.GetRemainBytes())
		conn.mu.Unlock()
	}
}

// ********************************************************************************************
// Retransmission Queue
func (conn *VTCPConn) retransmissionLoop() {
	for {
		segR := <-conn.rtmQueue
		go conn.retransmit(segR)
	}
}

func (conn *VTCPConn) retransmit(segR *proto.Segment) {
	time.Sleep(100 * time.Millisecond)
	conn.mu.Lock()
	defer conn.mu.Unlock()
	if segR.TCPhdr.AckNum <= conn.seqNum {
		myDebug.Debugln("[SendBuffer_RevACK] Segment with SEQ: %v, have been acked",
			segR.TCPhdr.SeqNum)
		return
	}
	conn.rtmQueue <- segR
}

// ********************************************************************************************
// helper function
func (conn *VTCPConn) send(content []byte, seqNum uint32) {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, seqNum), content)
	conn.NodeSegSendChan <- seg
	// add the segment to retransmission queue
	conn.rtmQueue <- seg

	myDebug.Debugln("[VSBufferSend] %v sent segment to %v, SEQ: %v, ACK: %v, Payload: %v\n",
		conn.LocalAddr.String(), conn.RemoteAddr.String(), seg.TCPhdr.SeqNum, conn.ackNum, string(seg.Payload))
}

func (conn *VTCPConn) GetTuple() string {
	return fmt.Sprintf("%v:%v:%v:%v", conn.RemoteAddr.String(), conn.RemotePort,
		conn.LocalAddr.String(), conn.LocalPort)
}

func (conn *VTCPConn) buildTCPHdr(flags int, seqNum uint32) *header.TCPFields {
	return &header.TCPFields{
		SrcPort:       conn.LocalPort,
		DstPort:       conn.RemotePort,
		SeqNum:        seqNum,
		AckNum:        conn.ackNum,
		DataOffset:    DEFAULTDATAOFFSET,
		Flags:         uint8(flags),
		WindowSize:    conn.windowSize,
		Checksum:      0,
		UrgentPointer: 0,
	}
}
