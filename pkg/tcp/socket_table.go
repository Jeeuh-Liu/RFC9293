package tcp

import "sync"

type SocketTable struct {
	mu sync.Mutex
	//tuple := remoteIP::remotePort::localPort
	counter           uint16
	id2Conns          map[uint16]*VTCPConn
	tuple2NormalConns map[string]*VTCPConn
	id2Listeners      map[uint16]*VTCPListener
	port2Listeners    map[uint16]*VTCPListener
}

func NewSocketTable() *SocketTable {
	return &SocketTable{
		counter:           uint16(0),
		id2Conns:          make(map[uint16]*VTCPConn),
		tuple2NormalConns: make(map[string]*VTCPConn),
		id2Listeners:      make(map[uint16]*VTCPListener),
		port2Listeners:    make(map[uint16]*VTCPListener),
	}
}

func (table *SocketTable) OfferListener(port uint16) *VTCPListener {
	table.mu.Lock()
	defer table.mu.Unlock()
	listener := NewListener(port)
	listener.ID = table.counter
	table.port2Listeners[port] = listener
	table.id2Listeners[listener.ID] = listener
	go listener.acceptLoop()
	table.counter++
	return listener
}

func (table *SocketTable) OfferConn(tuple string, conn *VTCPConn) {
	table.mu.Lock()
	defer table.mu.Unlock()
	conn.ID = table.counter
	table.id2Conns[conn.ID] = conn
	table.tuple2NormalConns[tuple] = conn
	table.counter++
}

func (table *SocketTable) FindListener(port uint16) *VTCPListener {
	table.mu.Lock()
	defer table.mu.Unlock()
	return table.port2Listeners[port]
}

func (table *SocketTable) FindConn(tuple string) *VTCPConn {
	table.mu.Lock()
	defer table.mu.Unlock()
	return table.tuple2NormalConns[tuple]
}

func (table *SocketTable) FindConnByID(id uint16) *VTCPConn {
	table.mu.Lock()
	defer table.mu.Unlock()
	return table.id2Conns[id]
}
