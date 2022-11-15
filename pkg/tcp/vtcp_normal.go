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
	DEFAULTWINDOWSIZE = 5
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
	//Recv
	NonEmptyCond *sync.Cond
	RcvBuf       *RecvBuffer
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
	conn.NonEmptyCond = sync.NewCond(&conn.mu)
	go conn.retransmissionLoop()
	return conn
}

// ********************************************************************************************
// Client
func (conn *VTCPConn) SynSend() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// Send Syn
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn, conn.seqNum), []byte{})
	conn.NodeSegSendChan <- seg
	conn.rtmQueue <- seg
	myDebug.Debugln("[Threeway Handshake - Send Syn] %v sent connection request to %v, SEQ: %v", conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.seqNum)
	// Rev Syn+ACK
	for {
		conn.mu.Unlock()
		segRev := <-conn.SegRcvChan
		conn.mu.Lock()

		myDebug.Debugln("[Threeway Handshake - Rev SYN+ACK] %v:%v receive packet from %v:%v, SEQ: %v, ACK %v",
			conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
			conn.RemotePort, segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum)
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			conn.seqNum = segRev.TCPhdr.AckNum
			conn.ackNum = segRev.TCPhdr.SeqNum + 1
			// Send Ack
			seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
			conn.NodeSegSendChan <- seg
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
func (conn *VTCPConn) SynRev() {
	conn.mu.Lock()
	conn.seqNum -= 1000000000
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck, conn.seqNum), []byte{})
	conn.NodeSegSendChan <- seg
	myDebug.Debugln("[Threeway Handshake - Send Syn + Ack] %v sent connection request to %v, SEQ: %v", conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.seqNum)
	conn.seqNum++
	conn.mu.Unlock()

	for {
		segRev := <-conn.SegRcvChan
		if conn.seqNum == segRev.TCPhdr.AckNum {
			myDebug.Debugln("[Threeway Handshake - Rev ACK] %v:%v receive packet from %v:%v, SEQ: %v, ACK %v",
				conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
				conn.RemotePort, segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum)
			conn.state = proto.ESTABLISH
			conn.RcvBuf = NewRecvBuffer(segRev.TCPhdr.SeqNum, DEFAULTWINDOWSIZE)
			go conn.estabRev()
			return
		}
	}
}

// ********************************************************************************************
// Send TCP Packet through Established Normal Conn

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
			if conn.sb.win == 0 {
				//buffer
				//
				// Zero probe
				payload, seqNum := conn.sb.UpdateNxt(1)
				conn.send(payload, seqNum)
			} else {
				// Get one segment, send it out and add it to retransmission queue
				payload, seqNum := conn.sb.UpdateNxt(mtu)
				conn.send(payload, seqNum)
			}
		} else {
			conn.sCv.Wait()
		}
	}
}

func (conn *VTCPConn) VSBufferRcv() {
	for {
		ack := <-conn.SegRcvChan
		// it is possible ACK is lost and we get another SynAck
		if ack.TCPhdr.Flags == (header.TCPFlagSyn | header.TCPFlagAck) {
			// myDebug.Debugln("[SendBuffer_RevACK] %v:%v receive packet from %v:%v, SEQ: %v, ACK %v",
			// 	conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
			// 	conn.RemotePort, ack.TCPhdr.SeqNum, ack.TCPhdr.AckNum)
			fmt.Println("[SB Rcv] Handshake Msg in VSBuffer")
			if ack.TCPhdr.Flags == (header.TCPFlagSyn | header.TCPFlagAck) {
				fmt.Println("[SB Rcv] Handshake Msg -> Send back another ACK")
				seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
				conn.NodeSegSendChan <- seg
			}

			continue
		}

		conn.mu.Lock()
		myDebug.Debugln("[SendBuffer_RevACK] %v:%v receive packet from %v:%v, SEQ: %v, ACK %v, WIN: %v",
			conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
			conn.RemotePort, ack.TCPhdr.SeqNum, ack.TCPhdr.AckNum, ack.TCPhdr.WindowSize)
		conn.sb.UpdateUNA(ack)
		conn.sb.UpdateWin(ack.TCPhdr.WindowSize)
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
		if segR.TCPhdr.Flags == header.TCPFlagAck && len(segR.Payload) > 0 {
			go conn.retransmit(segR)
		} else if segR.TCPhdr.Flags == header.TCPFlagSyn || (segR.TCPhdr.Flags == (header.TCPFlagSyn | header.TCPFlagAck)) {
			go conn.retransmitHS(segR)
		}

	}
}

func (conn *VTCPConn) retransmitHS(segR *proto.Segment) {
	time.Sleep(300 * time.Millisecond)
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// for handshake segments, seq number should increment by 1
	if conn.seqNum >= segR.TCPhdr.SeqNum+1 {
		myDebug.Debugln("[SendBuffer_RevACK] Segment with SEQ: %v, have been acked",
			segR.TCPhdr.SeqNum)
		return
	}
	// retransmit if not acked
	fmt.Printf("retransmit 1 HS segment flag: %v because current seqNum is %v and should be at least %v\n", segR.TCPhdr.Flags, conn.seqNum, segR.TCPhdr.SeqNum+1)
	conn.NodeSegSendChan <- segR
	conn.rtmQueue <- segR
}

func (conn *VTCPConn) retransmit(segR *proto.Segment) {
	time.Sleep(300 * time.Millisecond)
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// for ACK segments, seq number should increment by len(payload)
	if conn.seqNum >= segR.TCPhdr.SeqNum+uint32(len(segR.Payload)) {
		myDebug.Debugln("[SendBuffer_RevACK] Segment with SEQ: %v, have been acked",
			segR.TCPhdr.SeqNum)
		return
	}
	// retransmit if not acked
	fmt.Printf("retransmit 1 ACK segment because current seqNum is %v and should be at least %v; Payload: %v\n", conn.seqNum, segR.TCPhdr.SeqNum+uint32(len(segR.Payload)), segR.Payload)
	conn.NodeSegSendChan <- segR
	conn.rtmQueue <- segR
}

// ********************************************************************************************
// Recv
func (conn *VTCPConn) estabRev() {
	myDebug.Debugln("initial window size: %v, window size in buffer", conn.windowSize, conn.RcvBuf.GetWindowSize())
	for {
		segRev := <-conn.SegRcvChan
		conn.mu.Lock()
		myDebug.Debugln("Received segment %v:%v receive packet from %v:%v, SEQ: %v, ACK %v, Payload %v, CurrBuffer: %v",
			conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
			conn.RemotePort, segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum, segRev.Payload, conn.RcvBuf.DisplayBuf())
		if conn.windowSize == 0 {
			seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
			conn.NodeSegSendChan <- seg
			conn.mu.Unlock()
			fmt.Println("enter 247")
			continue
		}
		myDebug.Debugln("%v:%v receive packet from %v:%v, SEQ: %v, ACK %v, Payload %v",
			conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
			conn.RemotePort, segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum, segRev.Payload)
		status := conn.RcvBuf.GetSegStatus(segRev)
		if status == OUTSIDEWINDOW {
			fmt.Println("enter 254")
			continue
		}
		if status == UNDEFINED {
			fmt.Println("enter 258")
			continue
		}
		headAcked := conn.RcvBuf.IsHeadAcked()
		if status == EARLYARRIVAL || status == NEXTUNACKSEG {
			conn.ackNum, conn.windowSize = conn.RcvBuf.WriteSeg2Buf(segRev)
		}
		seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
		myDebug.Debugln("%v sent segment to %v, SEQ: %v, ACK: %v, Win: %v, Buffer: %v",
			conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.seqNum,
			conn.ackNum, conn.windowSize, conn.RcvBuf.DisplayBuf())
		conn.NodeSegSendChan <- seg
		conn.mu.Unlock()
		if !headAcked && status == NEXTUNACKSEG {
			conn.NonEmptyCond.Broadcast()
		}
	}
}

func (conn *VTCPConn) Retriv(numBytes uint32, isBlock bool) {
	res := []byte{}
	totalRead := uint32(0)
	for {
		conn.mu.Lock()
		if !conn.RcvBuf.IsHeadAcked() {
			conn.NonEmptyCond.Wait()
		}
		output, numRead := conn.RcvBuf.ReadBuf(numBytes)
		conn.windowSize += numRead
		conn.RcvBuf.SetWindowSize(uint32(conn.windowSize))

		res = append(res, output...)
		totalRead += uint32(numRead)
		myDebug.Debugln("to read %v bytes, return %v bytes, content %v, buffer %v, currWindowSize %v",
			numBytes, totalRead, res, conn.RcvBuf.DisplayBuf(), conn.windowSize)

		conn.mu.Unlock()
		if !isBlock || totalRead == numBytes {
			break
		}
	}
	fmt.Printf("now head point to %v\n", conn.RcvBuf.head)
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

func (conn *VTCPConn) Lock() {
	conn.mu.Lock()
}

func (conn *VTCPConn) Unlock() {
	conn.mu.Unlock()
}
