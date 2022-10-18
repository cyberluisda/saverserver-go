/*
Package serer define a tcp/udp server that accepts connections and saves received payload.

The payload items can be listed while server is running of after it is closed
*/
package server

import (
	"regexp"
	"sync"
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
