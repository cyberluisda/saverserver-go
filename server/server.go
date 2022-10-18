/*
Package serer define a tcp/udp server that accepts connections and saves received payload.

The payload items can be listed while server is running of after it is closed
*/
package server

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
