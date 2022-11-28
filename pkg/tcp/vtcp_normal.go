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
	// Write Condition Variable
	wcv sync.Cond
	// Send Buffer
	scv sync.Cond
	sb  *SendBuffer // send buffer
	// Retransmission
	rtmQueue      chan *proto.Segment  // retransmission queue
	seq2timestamp map[uint32]time.Time // seq # of 1 segment to expiration time
	//Recv
	NonEmptyCond *sync.Cond
	RcvBuf       *RecvBuffer
	BlockChan    chan *proto.NodeCLI
	// ZeroProbe
	zeroProbe bool
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
		zeroProbe:     false,
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
	// [HandShake1] Send Syn
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn, conn.seqNum), []byte{})
	conn.NodeSegSendChan <- seg
	conn.rtmQueue <- seg
	myDebug.Debugln("[Handshake 1] sent to %v, SEQ: %v", conn.RemoteAddr.String(), conn.seqNum)
	// [Handshake2] Rev Syn+ACK
	for {
		conn.mu.Unlock()
		segRev := <-conn.SegRcvChan
		conn.mu.Lock()
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			myDebug.Debugln("[Handshake 3] %v:%v sent to %v:%v, SEQ: %v, ACK %v",
				conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
				conn.RemotePort, segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum)
			// [Handshake3] Send Ack
			conn.seqNum = segRev.TCPhdr.AckNum
			conn.ackNum = segRev.TCPhdr.SeqNum + 1
			seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
			conn.NodeSegSendChan <- seg
			conn.state = proto.ESTABLISH
			// [Client] Create send buffer
			conn.sb = NewSendBuffer(conn.seqNum, uint32(segRev.TCPhdr.WindowSize))
			conn.scv = *sync.NewCond(&conn.mu)
			conn.wcv = *sync.NewCond(&conn.mu)
			go conn.VSBufferSend()
			// go conn.VSBufferRcv()
			// [Double: Client] Create rcv buffer
			conn.RcvBuf = NewRecvBuffer(conn.ackNum, DEFAULTWINDOWSIZE)
			// [Client] Rev Segments
			go conn.estabRevAndSend()
			return
		}
	}
}

// ********************************************************************************************
// Server
func (conn *VTCPConn) SynRev() {
	// [Handshake2] Send Syn|ACK
	conn.mu.Lock()
	conn.seqNum -= 1000000000
	conn.seqNum++
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck, conn.seqNum), []byte{})
	conn.NodeSegSendChan <- seg
	conn.rtmQueue <- seg
	myDebug.Debugln("[Handshake 2] sent to %v, SEQ: %v WIN: %v", conn.RemoteAddr.String(), conn.seqNum, conn.windowSize)
	conn.mu.Unlock()

	// [Handshake3] Rev ACK
	for {
		segRev := <-conn.SegRcvChan
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			myDebug.Debugln("[Handshake 3] %v:%v receive packet from %v:%v, SEQ: %v, ACK %v",
				conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
				conn.RemotePort, segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum)
			conn.seqNum = segRev.TCPhdr.AckNum
			conn.ackNum = segRev.TCPhdr.SeqNum
			conn.state = proto.ESTABLISH
			// [Server] Create rcv buffer
			conn.RcvBuf = NewRecvBuffer(conn.ackNum, DEFAULTWINDOWSIZE)
			// go conn.estabRev()
			// [Server] Rev Segments
			go conn.estabRevAndSend()
			// [Double: Server] Create send buffer
			conn.sb = NewSendBuffer(conn.seqNum, uint32(segRev.TCPhdr.WindowSize))
			conn.scv = *sync.NewCond(&conn.mu)
			conn.wcv = *sync.NewCond(&conn.mu)
			go conn.VSBufferSend()
			return
		}
	}
}

// ********************************************************************************************
// Handle SegRcv in both send buffer and rcv buffer
func (conn *VTCPConn) estabRevAndSend() {
	for {
		segRev := <-conn.SegRcvChan
		// it is possible ACK is lost and we get another SynAck
		myDebug.Debugln("156: %v, %v, %v", segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum, string(segRev.Payload))
		if len(segRev.Payload) == 0 {
			// Rcv segments In Send Buffer
			conn.HandleRcvSegInSendBuffer(segRev)
		} else {
			// Rcv segments In Rcv Buffer
			conn.HandleRcvSegInRcvBuffer(segRev)
		}
	}
}

// ********************************************************************************************
// Send TCP Packet through Established Normal Conn

func (conn *VTCPConn) VSBufferWrite(content []byte) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// fmt.Println("Hello")
	total := uint32(len(content))
	for total > 0 {
		// fmt.Printf("Hello1, isFull: %v\n", conn.sb.IsFull())
		if !conn.sb.IsFull() {
			bnum := conn.sb.WriteIntoBuffer(content)
			myDebug.Debugln("[Client] %v:%v writes %v bytes into send buffer, CurrSendBuffer:%v", conn.LocalAddr.String(), conn.LocalPort, bnum, string(conn.sb.buffer))
			total -= bnum
			content = content[bnum:]
			conn.scv.Signal()
			// fmt.Println("Hello2")
		} else {
			conn.wcv.Wait()
		}
	}
}

func (conn *VTCPConn) VSBufferSend() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	mtu := proto.DEFAULTPACKETMTU - proto.DEFAULTIPHDRLEN - proto.DEFAULTTCPHDRLEN
	for conn.state == proto.ESTABLISH {
		if conn.sb.CanSend() && !conn.zeroProbe {
			if conn.sb.win == 0 {
				//buffer
				// Zero probe
				payload, seqNum := conn.sb.GetZeroProbe()
				conn.send(payload, seqNum)
				conn.zeroProbe = true
			} else {
				// Get one segment, send it out and add it to retransmission queue
				payload, seqNum := conn.sb.GetSegmentToSendAndUpdateNxt(mtu)
				conn.send(payload, seqNum)
			}
		} else {
			conn.scv.Wait()
		}
	}
}

func (conn *VTCPConn) VSBufferRcv() {
	for {
		segRev := <-conn.SegRcvChan
		// it is possible ACK is lost and we get another SynAck
		if segRev.TCPhdr.Flags == (header.TCPFlagSyn | header.TCPFlagAck) {
			seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
			conn.NodeSegSendChan <- seg
			fmt.Printf("[HandShake3] Handshake Msg -> Send back another ACK %v\n", seg.TCPhdr.AckNum)
			continue
		}
		conn.HandleRcvSegInSendBuffer(segRev)
	}
}

func (conn *VTCPConn) HandleRcvSegInSendBuffer(segRev *proto.Segment) {
	// It is possible we still get another Syn|Ack Segment
	if segRev.TCPhdr.Flags == (header.TCPFlagSyn | header.TCPFlagAck) {
		seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
		conn.NodeSegSendChan <- seg
		fmt.Printf("[HandShake3] Handshake Msg -> Send back another ACK %v\n", seg.TCPhdr.AckNum)
		return
	}

	conn.mu.Lock()
	myDebug.Debugln("[Client] %v:%v receive from %v:%v, SEQ: %v, ACK %v, WIN: %v",
		conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
		conn.RemotePort, segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum, segRev.TCPhdr.WindowSize)
	acked := conn.sb.UpdateUNA(segRev)
	myDebug.Debugln("[Client] After ACK, Send Buffer Content: %v", string(conn.sb.buffer))
	conn.wcv.Signal()
	conn.sb.UpdateWin(segRev.TCPhdr.WindowSize)
	if segRev.TCPhdr.WindowSize > 0 {
		conn.zeroProbe = false
		conn.scv.Signal()
	}
	conn.seqNum += acked
	// myDebug.Debugln("[SendBuffer_RevACK] %v send buffer remaing bytes %v", conn.LocalAddr.String(), conn.sb.GetRemainBytes())
	conn.mu.Unlock()
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
		fmt.Printf("[Client] has successfully retransmitted 1 HS segment flag: %v\n", segR.TCPhdr.Flags)
		return
	}
	// retransmit if not acked
	fmt.Printf("[Client] retransmit 1 HS segment flag: %v because current seqNum is %v and should be at least %v\n", segR.TCPhdr.Flags, conn.seqNum, segR.TCPhdr.SeqNum+1)
	conn.NodeSegSendChan <- segR
	conn.rtmQueue <- segR
}

func (conn *VTCPConn) retransmit(segR *proto.Segment) {
	time.Sleep(300 * time.Millisecond)
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// for ACK segments, seq number should increment by len(payload)
	if conn.seqNum >= segR.TCPhdr.SeqNum+uint32(len(segR.Payload)) {
		return
	}
	// retransmit if not acked
	myDebug.Debugln("[Client] retransmit Payload %v: to ack it, SEQ needs to be at least %v. Curr SEQ: %v,  ", string(segR.Payload), segR.TCPhdr.SeqNum+uint32(len(segR.Payload)), conn.seqNum)
	conn.NodeSegSendChan <- segR
	conn.rtmQueue <- segR
}

// ********************************************************************************************
// Recv
func (conn *VTCPConn) estabRev() {
	for {
		segRev := <-conn.SegRcvChan
		conn.HandleRcvSegInRcvBuffer(segRev)
	}
}

func (conn *VTCPConn) HandleRcvSegInRcvBuffer(segRev *proto.Segment) {
	conn.mu.Lock()
	if conn.windowSize == 0 {
		// myDebug.Debugln("[Server] %v:%v receive zero probe from %v:%v, SEQ: %v, ACK %v",
		// 	conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
		// 	conn.RemotePort, segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum)
		seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
		conn.NodeSegSendChan <- seg
		conn.mu.Unlock()
		return
	}
	status := conn.RcvBuf.GetSegStatus(segRev)
	// fmt.Println("status:", status)
	if status == OUTSIDEWINDOW {
		// bug_fix: unlock when call return
		conn.mu.Unlock()
		return
	}
	if status == UNDEFINED {
		conn.mu.Unlock()
		return
	}
	myDebug.Debugln("[Server] %v:%v receive from %v:%v, SEQ: %v, ACK %v, Payload %v",
		conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
		conn.RemotePort, segRev.TCPhdr.SeqNum, segRev.TCPhdr.AckNum, string(segRev.Payload))
	// headAcked := conn.RcvBuf.IsHeadAcked()
	// bug_fix: already acked can also write some bytes
	// if status == EARLYARRIVAL || status == NEXTUNACKSEG || status == ALREADYACKED {
	// 	// bug_fix: change early arr
	// 	ackNum, windowSize := conn.RcvBuf.WriteSeg2Buf(segRev)
	// 	if status == NEXTUNACKSEG {
	// 		conn.ackNum = ackNum
	// 		conn.windowSize = windowSize
	// 	}
	// }
	ackNum, windowSize := conn.RcvBuf.WriteSeg2Buf(segRev)
	headAcked := conn.RcvBuf.IsHeadAcked()
	if headAcked {
		conn.ackNum = ackNum
		conn.windowSize = windowSize
		conn.NonEmptyCond.Broadcast()
	}

	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
	myDebug.Debugln("[Server] Current recv buffer content: %v", conn.RcvBuf.DisplayBuf())
	myDebug.Debugln("[Server] %v:%v sent to %v:%v, SEQ: %v, ACK: %v, Win: %v",
		conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
		conn.RemotePort, conn.seqNum, conn.ackNum, conn.windowSize)
	conn.NodeSegSendChan <- seg
	conn.mu.Unlock()
}

func (conn *VTCPConn) Retriv(numBytes uint32, isBlock bool) {
	res := []byte{}
	totalRead := uint32(0)
	// conn.BlockChan <- &proto.NodeCLI{CLIType: proto.CLI_BLOCKCLI}
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
		myDebug.Debugln("[Server] To READ %v bytes, return %v bytes, content %v, buffer %v, currWindowSize %v",
			numBytes, totalRead, string(res), conn.RcvBuf.DisplayBuf(), conn.windowSize)

		if string(res) == proto.TestString {
			println("************************************************")
			fmt.Println("Woww!!!!!!!!")
			println("************************************************")
		}
		conn.mu.Unlock()
		if !isBlock || totalRead == numBytes {
			break
		}
	}
	// conn.BlockChan <- &proto.NodeCLI{CLIType: proto.CLI_UNBLOCKCLI}
	fmt.Printf("now head point to %v\n", conn.RcvBuf.head)
}

// ********************************************************************************************
// helper function
func (conn *VTCPConn) send(content []byte, seqNum uint32) {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, seqNum), content)
	conn.NodeSegSendChan <- seg
	// add the segment to retransmission queue
	conn.rtmQueue <- seg

	myDebug.Debugln("[Client] %v:%v sent to %v:%v, SEQ: %v, ACK: %v, Payload: %v",
		conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(), conn.RemotePort, seg.TCPhdr.SeqNum, conn.ackNum, string(seg.Payload))
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
