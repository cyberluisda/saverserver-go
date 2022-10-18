/*
Package serer define a tcp/udp server that accepts connections and saves received payload.

The payload items can be listed while server is running of after it is closed
*/
package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"sync"
	"time"
)

type BasicServer interface {
	// Start starts the server (listener) and enable the input data processing.
	Start() error

	// Stop stops the listener, no more connections will be allowed and data processing is stopped.
	Stop() error

	// NPayloadItems returns the number of payloads items received by clients.
	NPayloadItems() int

	// GetPayloadAddresses returns the list of source address of the clients sent data.
	GetPayloadAddresses() []string

	//GetPayload returns the list of payloads received by a client identified with its source address.
	GetPayload(remoteAddr string) []byte

	//GetPayloads returns the full list of payloads in a map which key is the source address of the client.
	GetPayloads() map[string][]byte

	//Reset cleans the list of payloads received until now.
	Reset()

	//GetAddress returns the address where the server is listening.
	GetAddress() string

	// Port returns the listening port number or 0 if it is unknown or -1 if server is not running after call Start.
	Port() int

	// Accepting connections.
	Accepting() bool

	//Connections return the number of active connections.
	Connections() int
}

type Listener struct {
	PayloadStorage

	// Address is the address in ip:port format where server is listening.
	// If is not defined tcp://localhost:free_port will be used where free_port is a random port > 1024 that is not in
	// using when server is started
	// In the case of udp protocol de ip must be empty. For example udp://:13000
	Address string

	// Max number of connections to accept,
	MaxConnections int

	// StopTimeout is the timeout to wait for read data during Stop operation
	StopTimeout time.Duration

	activeConns    int
	activeConnsMtx sync.Mutex
	listener       net.Listener
	isStarted      bool
}

const (
	// Default max connections used if value is not defined
	DEFAULT_MAX_CONNECTIONS = 128
	// Default stop timeout
	DEFAULT_STOP_TIMEOUT = time.Second
)

// Start starts the server (listener) and enable the input data processing.
func (lst *Listener) Start() error {
	if lst.Address == "" {
		lst.Address = "tcp://localhost:0"
	}

	netType, addr, err := splitAddress(lst.Address)
	if err != nil {
		return fmt.Errorf("while extracts protocol, address and port: %w", err)
	}

	lst.listener, err = net.Listen(netType, addr)
	if err != nil {
		return fmt.Errorf("while starts listener: %w, '%s' '%s'", err, netType, addr)
	}

	// Default values
	if lst.Address == "tcp://localhost:0" {
		lst.Address = fmt.Sprintf("tcp://localhost:%d", lst.Port())
	}
	lst.InitPayload()
	if lst.MaxConnections <= 0 {
		lst.MaxConnections = DEFAULT_MAX_CONNECTIONS
	}
	if lst.StopTimeout == 0 {
		lst.StopTimeout = DEFAULT_STOP_TIMEOUT
	}

	// Start the server to accept connection
	ensureStarted := make(chan bool)
	go func() {
		firstConn := true
		for {

			if lst.Connections() == lst.MaxConnections {
				log.Printf("Max active connections limit (%d) reached with %d\n", lst.MaxConnections, lst.activeConns)
			} else {
				if firstConn {
					ensureStarted <- true
					firstConn = false
				}

				conn, err := lst.listener.Accept()
				if err != nil {
					log.Printf("Error while accept connection %v\n", err)
					break
				} else {
					go lst.handleIncomingConnection(conn)
				}
			}
		}
	}()

	<-ensureStarted
	lst.isStarted = true
	close(ensureStarted)

	return nil
}

// Stop stops the listener, no more connections will be allowed and data processing is stopped.
func (lst *Listener) Stop() error {
	defer func() {
		lst.isStarted = false
		lst.activeConns = 0
	}()

	connPending := make(chan bool)
	defer close(connPending)
	go func() {
		for {
			lst.activeConnsMtx.Lock()
			if lst.activeConns <= 0 {
				lst.activeConnsMtx.Unlock()
				connPending <- true
				break
			}
			lst.activeConnsMtx.Unlock()
			time.Sleep(time.Millisecond * 100)
		}
	}()

	select {
	case <-connPending:
	case <-time.After(lst.StopTimeout):
		defer lst.listener.Close()
		return fmt.Errorf("stop timeout %v reached while wait for stopping", lst.StopTimeout)
	}

	err := lst.listener.Close()
	if err != nil {
		return fmt.Errorf("while close the listener: %w", err)
	}

	return nil
}

//GetAddress returns the address where the server is listening.
func (lst *Listener) GetAddress() string {
	return lst.Address
}

// Port returns the listening port number or 0 if it is unknown or -1 if server is not running after call Start.
func (lst *Listener) Port() int {
	if lst.listener == nil {
		return -1
	}

	tcpAddr, ok := lst.listener.Addr().(*net.TCPAddr)
	if ok {
		return tcpAddr.Port
	}

	return 0
}

// Accepting connections.
func (lst *Listener) Accepting() bool {
	lst.activeConnsMtx.Lock()
	defer lst.activeConnsMtx.Unlock()
	return lst.activeConns < lst.MaxConnections && lst.isStarted
}

//Connections return the number of active connections.
func (lst *Listener) Connections() int {
	lst.activeConnsMtx.Lock()
	defer lst.activeConnsMtx.Unlock()
	return lst.activeConns
}

const readBufferSize = 1024

func (lst *Listener) handleIncomingConnection(conn net.Conn) {
	defer func() {
		lst.activeConnsMtx.Lock()
		lst.activeConns = lst.activeConns - 1
		lst.activeConnsMtx.Unlock()
	}()

	lst.activeConnsMtx.Lock()
	lst.activeConns = lst.activeConns + 1
	lst.activeConnsMtx.Unlock()

	// store incoming data
	remoteAddress := conn.RemoteAddr().String()
	for {
		buffer := make([]byte, readBufferSize)
		n, err := conn.Read(buffer)
		if err != nil && err != io.EOF {
			log.Println("while close connection: %w", err)
			break
		}
		if n != 0 {
			lst.AddPayload(remoteAddress, buffer, n)
		}
		if err == io.EOF {
			break
		}
	}

	// close conn
	err := conn.Close()
	if err != nil {
		log.Println("while close connection: %w", err)
	}
}

type ListenerPacket struct {
	PayloadStorage
	Address string

	conn    net.PacketConn
	started bool
}

// Start starts the server (listener) and enable the input data processing.
func (lp *ListenerPacket) Start() error {
	if lp.Address == "" {
		lp.Address = "udp://localhost:0"
	}

	netType, addr, err := splitAddress(lp.Address)
	if err != nil {
		return fmt.Errorf("while extracts protocol, address and port: %w", err)
	}

	lp.conn, err = net.ListenPacket(netType, addr)
	if err != nil {
		return fmt.Errorf("while starts listener: %w, '%s' '%s'", err, netType, addr)
	}

	// Default values
	if lp.Address == "udp://localhost:0" {
		lp.Address = fmt.Sprintf("udp://localhost:%d", lp.Port())
	}

	// Initialization
	lp.InitPayload()

	// Start the server to accept connections
	go lp.handleIncomingPackets()

	lp.started = true
	return nil
}

// Stop stops the listener, no more connections will be allowed and data processing is stopped.
func (lp *ListenerPacket) Stop() error {
	defer func() {
		lp.started = false
	}()

	if lp.started {
		err := lp.conn.Close()
		if err != nil {
			return fmt.Errorf("while close packet connection: %w", err)
		}
	}

	return nil
}

//GetAddress returns the address where the server is listening.
func (lp *ListenerPacket) GetAddress() string {
	return lp.Address
}

// Port returns the listening port number or 0 if it is unknown or -1 if server is not running after call Start.
func (lp *ListenerPacket) Port() int {
	if lp.conn == nil {
		return -1
	}

	updAddr, ok := lp.conn.LocalAddr().(*net.UDPAddr)
	if ok {
		return updAddr.Port
	}

	return 0
}

// Accepting connections.
func (lp *ListenerPacket) Accepting() bool {
	return lp.started
}

//Connections returns 0 because in udp we have not any active connection
func (lp *ListenerPacket) Connections() int {
	return 0
}

func (lp *ListenerPacket) handleIncomingPackets() {
	buffer := make([]byte, 1024)
	//n, remoteAddr, err := 0, new(net.Addr), error(nil)
	err := error(nil)
	for err == nil {
		n, remoteAddr, err := lp.conn.ReadFrom(buffer)
		if err != nil {
			log.Println("while read data from packet connection: %w", err)
		} else if n > 0 {
			addr := remoteAddr.String()
			lp.AddPayload(addr, buffer, n)
		}
	}
}

type PayloadStorage struct {
	payloads    map[string][]byte
	payloadsMtx sync.RWMutex
}

func (ps *PayloadStorage) InitPayload() {
	if ps.payloads == nil {
		ps.payloads = make(map[string][]byte)
	}
}

// NPayloadItems returns the number of payloads items received by clients.
func (ps *PayloadStorage) NPayloadItems() int {
	ps.payloadsMtx.RLock()
	defer ps.payloadsMtx.RUnlock()
	return len(ps.payloads)
}

//Reset cleans the list of payloads received until now.
func (ps *PayloadStorage) Reset() {
	ps.payloadsMtx.Lock()
	defer ps.payloadsMtx.Unlock()
	for k := range ps.payloads {
		delete(ps.payloads, k)
	}
}

func (ps *PayloadStorage) AddPayload(addr string, buffer []byte, n int) {
	ps.payloadsMtx.Lock()
	defer ps.payloadsMtx.Unlock()

	v, ok := ps.payloads[addr]
	if ok {
		v = append(v, buffer[0:n]...)
	} else {
		v = make([]byte, n)
		copy(v, buffer)
	}
	ps.payloads[addr] = v
}

// GetPayloadAddresses returns the list of source address of the clients sent data.
func (ps *PayloadStorage) GetPayloadAddresses() []string {
	ps.payloadsMtx.RLock()
	defer ps.payloadsMtx.RUnlock()
	r := make([]string, len(ps.payloads))
	i := 0
	for k := range ps.payloads {
		r[i] = k
		i = i + 1
	}

	return r
}

//GetPayload returns the list of payloads received by a client identified with its source address.
func (ps *PayloadStorage) GetPayload(remoteAddr string) []byte {
	ps.payloadsMtx.RLock()
	defer ps.payloadsMtx.RUnlock()
	return ps.payloads[remoteAddr]
}

//GetPayloads returns the full list of payloads in a map which key is the source address of the client.
func (ps *PayloadStorage) GetPayloads() map[string][]byte {
	ps.payloadsMtx.RLock()
	defer ps.payloadsMtx.RUnlock()

	return ps.payloads
}

func splitAddress(a string) (string, string, error) {
	ptrStr := `^(\w+)://([^:]*:\d+)$`
	ptr := regexp.MustCompile(ptrStr)
	r := ptr.FindStringSubmatch(a)
	if len(r) != 3 { // full string, first submatch and second submatch
		return "", "", fmt.Errorf("%s is not matching with pattern expected for a valid url (%s)", a, ptrStr)
	}

	return string(r[1]), string(r[2]), nil
}
