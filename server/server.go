/*
Package serer define a tcp/udp server that accepts connections and saves received payload.

The payload items can be listed while server is running of after it is closed
*/
package server

import (
	"crypto/tls"
	"crypto/x509"
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

	// GetPayload returns the list of payloads received by a client identified with its source address.
	GetPayload(remoteAddr string) []byte

	// GetPayloads returns the full list of payloads in a map which key is the source address of the client.
	GetPayloads() map[string][]byte

	// Reset cleans the list of payloads received until now.
	Reset()

	// GetAddress returns the address where the server is listening.
	GetAddress() string

	// Port returns the listening port number or 0 if it is unknown or -1 if server is not running after call Start.
	Port() int

	// Accepting connections.
	Accepting() bool

	// Connections return the number of active connections.
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
	// DefaultMaxConnections are the default max connections used if value is not defined
	DefaultMaxConnections = 128
	// DefaultSopTimeout is the default stop timeout
	DefaultSopTimeout = time.Second
	// Default listen address for listeners
	DefaultListenAddressListener = "tcp://localhost:0"
)

// Start starts the server (listener) and enable the input data processing.
func (lst *Listener) Start() error {
	if lst.Address == "" {
		lst.Address = DefaultListenAddressListener
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
	if lst.Address == DefaultListenAddressListener {
		lst.Address = fmt.Sprintf("tcp://localhost:%d", lst.Port())
	}
	lst.Init()
	if lst.MaxConnections <= 0 {
		lst.MaxConnections = DefaultMaxConnections
	}
	if lst.StopTimeout == 0 {
		lst.StopTimeout = DefaultSopTimeout
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

const tickerWhileStopping = time.Millisecond * 100

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
			time.Sleep(tickerWhileStopping)
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

// GetAddress returns the address where the server is listening.
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

// Connections return the number of active connections.
func (lst *Listener) Connections() int {
	lst.activeConnsMtx.Lock()
	defer lst.activeConnsMtx.Unlock()
	return lst.activeConns
}

const readBufferSize = 1024

func (lst *Listener) handleIncomingConnection(conn net.Conn) {
	defer func() {
		lst.activeConnsMtx.Lock()
		lst.activeConns--
		lst.activeConnsMtx.Unlock()
	}()

	lst.activeConnsMtx.Lock()
	lst.activeConns++
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

// DefaultListenAddressListenerPacket is the default listen address for ListenerPacket
const DefaultListenAddressListenerPacket = "udp://localhost:0"

// Start starts the server (listener) and enable the input data processing.
func (lp *ListenerPacket) Start() error {
	if lp.Address == "" {
		lp.Address = DefaultListenAddressListenerPacket
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
	if lp.Address == DefaultListenAddressListenerPacket {
		lp.Address = fmt.Sprintf("udp://localhost:%d", lp.Port())
	}

	// Initialization
	lp.Init()

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

// GetAddress returns the address where the server is listening.
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

// Connections returns 0 because in udp we have not any active connection
func (lp *ListenerPacket) Connections() int {
	return 0
}

const packetsBufferSize = 1024

func (lp *ListenerPacket) handleIncomingPackets() {
	buffer := make([]byte, packetsBufferSize)
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

type TLSListener struct {
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

	// KeyPem is the key used to enable TLS protocol
	KeyPem []byte
	// CertPem is the cert used to enable TLS protocol
	CertPem []byte
	// ClientAuth is the value with same name described in https://pkg.go.dev/crypto/tls@go1.16.15#Config
	ClientAuth tls.ClientAuthType
	// ClientCAs is the list of CAs used to verify client certs. See https://pkg.go.dev/crypto/tls@go1.16.15#Config
	ClientCAs [][]byte
	// MinVersion is the value with same name described in https://pkg.go.dev/crypto/tls@go1.16.15#Config
	MinVersion uint16
	// MinVersion is the value with same name described in https://pkg.go.dev/crypto/tls@go1.16.15#Config
	MaxVersion uint16
	// KeyLogWriter is the value with same name described in https://pkg.go.dev/crypto/tls@go1.16.15#Config
	KeyLogWriter io.Writer

	activeConns    int
	activeConnsMtx sync.Mutex
	listener       net.Listener
	isStarted      bool
}

// Start starts the server (listener) and enable the input data processing.
func (tll *TLSListener) Start() error {
	if tll.Address == "" {
		tll.Address = DefaultListenAddressListener
	}

	netType, addr, err := splitAddress(tll.Address)
	if err != nil {
		return fmt.Errorf("while extracts protocol, address and port: %w", err)
	}

	cert, err := tls.X509KeyPair(tll.CertPem, tll.KeyPem)
	if err != nil {
		return fmt.Errorf("while loads Key and Certificate: %w", err)
	}

	if tll.MinVersion == 0 {
		tll.MinVersion = tls.VersionTLS12
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tll.ClientAuth,
		MinVersion:   tll.MinVersion,
		MaxVersion:   tll.MaxVersion,
		KeyLogWriter: tll.KeyLogWriter,
	}

	if len(tll.ClientCAs) > 0 {
		config.ClientCAs = x509.NewCertPool()
		for i, bs := range tll.ClientCAs {
			ok := config.ClientCAs.AppendCertsFromPEM(bs)
			if !ok {
				return fmt.Errorf("ClientCA certificate number %d to authenticate user can not be loaded", i+1)
			}
		}
	}

	tll.listener, err = tls.Listen(netType, addr, config)
	if err != nil {
		return fmt.Errorf("while starts listener: %w, '%s' '%s' '%+v'", err, netType, addr, config)
	}

	// Default values
	if tll.Address == DefaultListenAddressListener {
		tll.Address = fmt.Sprintf("tcp://localhost:%d", tll.Port())
	}
	tll.Init()
	if tll.MaxConnections <= 0 {
		tll.MaxConnections = DefaultMaxConnections
	}
	if tll.StopTimeout == 0 {
		tll.StopTimeout = DefaultSopTimeout
	}

	// Start the server to accept connection
	ensureStarted := make(chan bool)
	go func() {
		firstConn := true
		for {
			if tll.Connections() == tll.MaxConnections {
				log.Printf("Max active connections limit (%d) reached with %d\n", tll.MaxConnections, tll.activeConns)
			} else {
				if firstConn {
					ensureStarted <- true
					firstConn = false
				}

				conn, err := tll.listener.Accept()

				if err != nil {
					log.Printf("Error while accept connection %v\n", err)
					break
				} else {
					tlsConn := conn.(*tls.Conn)
					go tll.handleIncomingTLSConnection(tlsConn)
				}
			}
		}
	}()

	<-ensureStarted
	tll.isStarted = true
	close(ensureStarted)

	return nil
}

// Stop stops the listener, no more connections will be allowed and data processing is stopped.
func (tll *TLSListener) Stop() error {
	defer func() {
		tll.isStarted = false
		tll.activeConns = 0
	}()

	connPending := make(chan bool)
	defer close(connPending)
	go func() {
		for {
			tll.activeConnsMtx.Lock()
			if tll.activeConns <= 0 {
				tll.activeConnsMtx.Unlock()
				connPending <- true
				break
			}
			tll.activeConnsMtx.Unlock()
			time.Sleep(tickerWhileStopping)
		}
	}()

	select {
	case <-connPending:
	case <-time.After(tll.StopTimeout):
		defer tll.listener.Close()
		return fmt.Errorf("stop timeout %v reached while wait for stopping", tll.StopTimeout)
	}

	err := tll.listener.Close()
	if err != nil {
		return fmt.Errorf("while close the listener: %w", err)
	}

	return nil
}

// GetAddress returns the address where the server is listening.
func (tll *TLSListener) GetAddress() string {
	return tll.Address
}

// Port returns the listening port number or 0 if it is unknown or -1 if server is not running after call Start.
func (tll *TLSListener) Port() int {
	if tll.listener == nil {
		return -1
	}

	tcpAddr, ok := tll.listener.Addr().(*net.TCPAddr)
	if ok {
		return tcpAddr.Port
	}

	return 0
}

// Accepting connections.
func (tll *TLSListener) Accepting() bool {
	tll.activeConnsMtx.Lock()
	defer tll.activeConnsMtx.Unlock()
	return tll.activeConns < tll.MaxConnections && tll.isStarted
}

// Connections return the number of active connections.
func (tll *TLSListener) Connections() int {
	tll.activeConnsMtx.Lock()
	defer tll.activeConnsMtx.Unlock()
	return tll.activeConns
}

func (tll *TLSListener) handleIncomingTLSConnection(conn *tls.Conn) {
	defer func() {
		tll.activeConnsMtx.Lock()
		tll.activeConns--
		tll.activeConnsMtx.Unlock()
	}()

	tll.activeConnsMtx.Lock()
	tll.activeConns++
	tll.activeConnsMtx.Unlock()

	// store incoming data
	clientID := conn.RemoteAddr().String()

	err := conn.Handshake()
	if err != nil {
		log.Println("Error while make handshake:", err)
		return
	}

	cs := conn.ConnectionState()
	nCerts := len(cs.PeerCertificates)

	if nCerts > 0 {
		clientID = "@" + clientID
		for i, c := range cs.PeerCertificates {
			clientID = c.Subject.String() + clientID
			if i < nCerts-1 {
				clientID = "-" + clientID
			}
		}
	}

	for {
		buffer := make([]byte, readBufferSize)
		var n int
		n, err = conn.Read(buffer)
		if err != nil && err != io.EOF {
			log.Println("while close connection:", err)
			break
		}
		if n != 0 {
			tll.AddPayload(clientID, buffer, n)
		}
		if err == io.EOF {
			break
		}
	}

	// close conn
	err = conn.Close()
	if err != nil {
		log.Println("while close connection:", err)
	}
}

type StoppableConnectionsMgr struct {
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

func (scm *StoppableConnectionsMgr) Stop() error {
	defer func() {
		scm.isStarted = false
		scm.activeConns = 0
	}()

	connPending := make(chan bool)
	defer close(connPending)
	go func() {
		for {
			scm.activeConnsMtx.Lock()
			if scm.activeConns <= 0 {
				scm.activeConnsMtx.Unlock()
				connPending <- true
				break
			}
			scm.activeConnsMtx.Unlock()
			time.Sleep(tickerWhileStopping)
		}
	}()

	select {
	case <-connPending:
	case <-time.After(scm.StopTimeout):
		defer scm.listener.Close()
		return fmt.Errorf("stop timeout %v reached while wait for stopping", scm.StopTimeout)
	}

	err := scm.listener.Close()
	if err != nil {
		return fmt.Errorf("while close the listener: %w", err)
	}

	return nil
}

// GetAddress returns the address where the server is listening.
func (scm *StoppableConnectionsMgr) GetAddress() string {
	return scm.Address
}

// Port returns the listening port number or 0 if it is unknown or -1 if server is not running after call Start.
func (scm *StoppableConnectionsMgr) Port() int {
	if scm.listener == nil {
		return -1
	}

	tcpAddr, ok := scm.listener.Addr().(*net.TCPAddr)
	if ok {
		return tcpAddr.Port
	}

	return 0
}

// Accepting connections.
func (scm *StoppableConnectionsMgr) Accepting() bool {
	scm.activeConnsMtx.Lock()
	defer scm.activeConnsMtx.Unlock()
	return scm.activeConns < scm.MaxConnections && scm.isStarted
}

// Connections return the number of active connections.
func (scm *StoppableConnectionsMgr) Connections() int {
	scm.activeConnsMtx.Lock()
	defer scm.activeConnsMtx.Unlock()
	return scm.activeConns
}

type PayloadStorage struct {
	payloads    map[string][]byte
	payloadsMtx sync.RWMutex

	// CallBack is a function called in each time that new payload is arrived. The func
	//	receive the address and the payload received and it should return true if payload
	// 	must be saved or false if payload must be discarded.
	CallBack func(addr string, payload []byte) bool
}

func (ps *PayloadStorage) Init() {
	if ps.payloads == nil {
		ps.payloads = make(map[string][]byte)
	}

	if ps.CallBack == nil {
		ps.CallBack = func(addr string, payload []byte) bool { return true }
	}
}

// NPayloadItems returns the number of payloads items received by clients.
func (ps *PayloadStorage) NPayloadItems() int {
	ps.payloadsMtx.RLock()
	defer ps.payloadsMtx.RUnlock()
	return len(ps.payloads)
}

// Reset cleans the list of payloads received until now.
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

	// First apply callback and check if we have to save payload
	save := ps.CallBack(addr, buffer[0:n])

	if save {
		v, ok := ps.payloads[addr]
		if ok {
			v = append(v, buffer[0:n]...)
		} else {
			v = make([]byte, n)
			copy(v, buffer)
		}
		ps.payloads[addr] = v
	}
}

// GetPayloadAddresses returns the list of source address of the clients sent data.
func (ps *PayloadStorage) GetPayloadAddresses() []string {
	ps.payloadsMtx.RLock()
	defer ps.payloadsMtx.RUnlock()
	r := make([]string, len(ps.payloads))
	i := 0
	for k := range ps.payloads {
		r[i] = k
		i++
	}

	return r
}

// GetPayload returns the list of payloads received by a client identified with its source address.
func (ps *PayloadStorage) GetPayload(remoteAddr string) []byte {
	ps.payloadsMtx.RLock()
	defer ps.payloadsMtx.RUnlock()
	return ps.payloads[remoteAddr]
}

// GetPayloads returns the full list of payloads in a map which key is the source address of the client.
func (ps *PayloadStorage) GetPayloads() map[string][]byte {
	ps.payloadsMtx.RLock()
	defer ps.payloadsMtx.RUnlock()

	return ps.payloads
}

func splitAddress(a string) (protocol, address string, err error) {
	ptrStr := `^(\w+)://([^:]*:\d+)$`
	ptr := regexp.MustCompile(ptrStr)
	r := ptr.FindStringSubmatch(a)
	if len(r) != 3 { // full string, first submatch and second submatch
		return "", "", fmt.Errorf("%s is not matching with pattern expected for a valid url (%s)", a, ptrStr)
	}

	return r[1], r[2], nil
}
