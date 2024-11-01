package tcp

import (
	"fmt"
	"tcpip/pkg/myDebug"
	"tcpip/pkg/proto"
)

const (
	NEXTUNACKSEG  = 0
	OUTSIDEWINDOW = 1
	ALREADYACKED  = 2
	EARLYARRIVAL  = 3
	UNDEFINED     = 4
)

type RecvBuffer struct {
	// |_|newhead-----una -------early(last)------bound}-
	una    uint32 //oldest number that is not acknowledge
	head   uint32
	window uint32          //
	buffer map[uint32]byte //segment-seq to bit
}

func NewRecvBuffer(seq uint32, sz uint32) *RecvBuffer {
	return &RecvBuffer{
		una:    seq,
		head:   seq,
		window: sz,
		buffer: make(map[uint32]byte),
	}
}

func (buf *RecvBuffer) WriteSeg2Buf(seg *proto.Segment) (uint32, uint16) {
	// pos := seg.TCPhdr.SeqNum
	// // bug fix: first byte acked doesn't mean all bytes in payload have been acked
	// // we must add it to map as many as possible (length of payload can get out of range)
	// // _, acked := buf.buffer[calcIndex(pos)]
	// // pos cannot write into a position too far away
	// // if acked {
	// // 	return buf.una, uint16(buf.window)
	// // }
	// //ack,b
	// acked := uint32(0)
	// for _, b := range seg.Payload {
	// 	// ignore the acked
	// 	if pos < buf.head {
	// 		acked += 1
	// 		pos++
	// 	} else if pos >= buf.head+DEFAULTWINDOWSIZE {
	// 		// break if too far away
	// 		break
	// 	} else {
	// 		buf.buffer[calcIndex(pos)] = b
	// 		pos++
	// 	}
	// }
	// start := seg.TCPhdr.SeqNum + uint32(acked)
	// fmt.Println("start is:", start)
	// //-------|-----|--------xxxxxx
	// if buf.una == start {
	// 	_, found := buf.buffer[calcIndex(pos)]
	// 	// at most ack windowsize bytes and cannot move back to buf.head
	// 	for found && pos < buf.head+DEFAULTWINDOWSIZE {
	// 		pos++
	// 		_, found = buf.buffer[calcIndex(pos)]
	// 	}
	// 	buf.una = pos
	// }
	// newWindow := DEFAULTWINDOWSIZE - (buf.una - buf.head)
	// // myDebug.Debugln("old win: %v, new win: %v, pos: %v, head: %v", buf.window, newWindow, pos, buf.head)
	// if newWindow < buf.window {
	// 	buf.window = newWindow
	// }
	// return buf.una, uint16(buf.window)

	pos := seg.TCPhdr.SeqNum
	_, acked := buf.buffer[pos]
	if acked {
		return buf.una, uint16(buf.window)
	}
	//ack,b
	numPrev := uint32(0)
	for _, b := range seg.Payload {
		// ignore the acked
		if pos < buf.head {
			numPrev += 1
			pos++
		} else if pos >= buf.head+DEFAULTWINDOWSIZE {
			// break if too far away
			break
		} else {
			buf.buffer[pos] = b
			pos++
		}
	}
	//buf.una - buf.head
	newWindow := DEFAULTWINDOWSIZE - (pos - buf.head)
	// myDebug.Debugln("old win: %v, new win: %v, pos: %v, head: %v", buf.window, newWindow, pos, buf.head)
	if newWindow < buf.window {
		buf.window = newWindow
	}
	//-------|-----|--------xxxxxx
	start := seg.TCPhdr.SeqNum + numPrev
	fmt.Println(buf.una, start)
	if buf.una == start {
		_, found := buf.buffer[pos]
		for found {
			pos++
			_, found = buf.buffer[pos]
		}
		buf.una = pos
	}
	return buf.una, uint16(buf.window)
}

func (buf *RecvBuffer) ReadBuf(numBytes uint32) ([]byte, uint16) {
	output := []byte{}
	cnt := uint32(0)
	for cnt < numBytes && buf.head < buf.una {
		index := buf.head
		b := buf.buffer[index]
		delete(buf.buffer, index)

		output = append(output, b)
		buf.head++
		cnt++
	}
	return output, uint16(len(output))
}

func (buf *RecvBuffer) IsHeadAcked() bool {
	return buf.una != buf.head
}

func (buf *RecvBuffer) DisplayBuf() string {
	res := []byte{}
	pos := buf.head
	for cnt := 0; cnt < DEFAULTWINDOWSIZE; cnt++ {
		val, acked := buf.buffer[pos]
		if acked {
			res = append(res, val)
		} else {
			res = append(res, byte('*'))
		}
		pos++
	}
	return string(res)
}

func (buf *RecvBuffer) GetSegStatus(seg *proto.Segment) uint8 {
	seq := seg.TCPhdr.SeqNum
	// if seq+uint32(len(seg.Payload)) > buf.head+uint32(DEFAULTWINDOWSIZE) {
	// 	return SENDERDUTY
	// }
	// bug_fix: seq > buf.head+uint32(DEFAULTWINDOWSIZE)
	// bug_fix if seq < buf.head => unacked
	if seq >= buf.head+uint32(DEFAULTWINDOWSIZE) {
		myDebug.Debugln("Out of range : %v, %v", seq, buf.head)
		return OUTSIDEWINDOW
	}
	if seq == buf.una {
		return NEXTUNACKSEG
	}
	if seq < buf.una {
		return ALREADYACKED
	}
	if seq > buf.una {
		return EARLYARRIVAL
	}
	return UNDEFINED
}

func (buf *RecvBuffer) SetWindowSize(size uint32) {
	buf.window = size
}

func calcIndex(pos uint32) uint32 {
	return pos % DEFAULTWINDOWSIZE
}

func (buf *RecvBuffer) GetWindowSize() uint16 {
	return uint16(buf.window)
}
