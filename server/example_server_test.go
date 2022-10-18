package server

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func ExampleListener_random_port() {
	lst := Listener{}
	err := lst.Start()
	if err != nil {
		panic(err)
	}
	p := lst.Port()
	if p > 1024 && p < 65536 {
		fmt.Println("OK")
	} else {
		fmt.Printf("Port returned and unexpected value: %d\n", p)
	}

	err = lst.Stop()
	if err != nil {
		panic(err)
	}

	//Output:
	// OK
}

func ExampleListener_protocol_error() {
	lst := Listener{
		Address: "udp://:1536",
	}
	err := lst.Start()
	if err == nil {
		fmt.Println("No Error")
	} else {
		fmt.Println("Error", err)
	}

	//Output:
	// Error while starts listener: listen udp :1536: address :1536: unexpected address type, 'udp' ':1536'
}

func ExampleListener() {
	lst := Listener{}
	err := lst.Start()
	if err != nil {
		panic(err)
	}
	if lst.Port() <= 0 {
		panic("Port unknown")
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", strings.TrimPrefix(lst.GetAddress(), "tcp://"))
	if err != nil {
		panic(
			fmt.Sprintf("while resolve address: %v", err),
		)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(
			fmt.Sprintf("while connect to address %s address: %v", tcpAddr, err),
		)
	}

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf(" this is test number %d ", i)
		_, err = conn.Write([]byte(msg))
		if err != nil {
			panic(
				fmt.Sprintf("while send data to socket: %v", err),
			)
		}
	}

	err = conn.Close()
	if err != nil {
		panic(
			fmt.Sprintf("while close client: %v", err),
		)
	}

	err = lst.Stop()
	if err != nil {
		panic(err)
	}

	fmt.Println("#Items", lst.NPayloadItems())
	rAddress := lst.GetPayloadAddresses()[0]
	fmt.Println(string(lst.GetPayload(rAddress)))

	//Output:
	// #Items 1
	//  this is test number 0  this is test number 1  this is test number 2  this is test number 3  this is test number 4  this is test number 5  this is test number 6  this is test number 7  this is test number 8  this is test number 9
}

func ExampleListener_Accepting() {
	lst := Listener{}

	fmt.Println("Accepting before start", lst.Accepting())

	err := lst.Start()
	if err != nil {
		panic(err)
	}

	fmt.Println("Accepting after start", lst.Accepting())

	err = lst.Stop()
	if err != nil {
		panic(err)
	}

	fmt.Println("Accepting after stop", lst.Accepting())

	//Output:
	// Accepting before start false
	// Accepting after start true
	// Accepting after stop false
}

func ExampleListener_Accepting_max_cons() {
	lst := Listener{
		MaxConnections: 1,
	}
	err := lst.Start()
	if err != nil {
		panic(err)
	}
	if lst.Port() <= 0 {
		panic("Port unknown")
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", strings.TrimPrefix(lst.GetAddress(), "tcp://"))
	if err != nil {
		panic(
			fmt.Sprintf("while resolve address: %v", err),
		)
	}

	fmt.Println("Accepting before client connection", lst.Accepting())

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(
			fmt.Sprintf("while connect to address %s address: %v", tcpAddr, err),
		)
	}

	fmt.Println("Accepting after client connection", lst.Accepting())
	msg := []byte("")
	for i := 0; i < 1000; i++ {
		_, err := conn.Write(msg)
		if err != nil {
			panic(
				fmt.Sprintf("While send data to %s: %v", tcpAddr, err),
			)
		}
	}

	fmt.Println("Accepting while client is sending data", lst.Accepting())

	err = conn.Close()
	if err != nil {
		panic(
			fmt.Sprintf("while close client: %v", err),
		)
	}

	err = lst.Stop()
	if err != nil {
		panic(err)
	}

	//Output:
	// Accepting before client connection true
	// Accepting after client connection true
	// Accepting while client is sending data false
}

func ExampleListener_Connections() {
	lst := Listener{
		MaxConnections: 1,
	}
	err := lst.Start()
	if err != nil {
		panic(err)
	}
	if lst.Port() <= 0 {
		panic("Port unknown")
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", strings.TrimPrefix(lst.GetAddress(), "tcp://"))
	if err != nil {
		panic(
			fmt.Sprintf("while resolve address: %v", err),
		)
	}

	fmt.Println("Connections before client connection", lst.Connections())

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(
			fmt.Sprintf("while connect to address %s address: %v", tcpAddr, err),
		)
	}

	fmt.Println("Connections after client connection", lst.Connections())

	msg := []byte("")
	for i := 0; i < 1000; i++ {
		_, err := conn.Write(msg)
		if err != nil {
			panic(
				fmt.Sprintf("While send data to %s: %v", tcpAddr, err),
			)
		}
	}

	fmt.Println("Connections while client is sending data", lst.Connections())

	err = conn.Close()
	if err != nil {
		panic(
			fmt.Sprintf("while close client: %v", err),
		)
	}

	time.Sleep(time.Millisecond * 100) // Wait to ensure connection was released
	fmt.Println("Connections after client is closed and connection was released", lst.Connections())

	err = lst.Stop()
	if err != nil {
		panic(err)
	}

	//Output:
	// Connections before client connection 0
	// Connections after client connection 0
	// Connections while client is sending data 1
	// Connections after client is closed and connection was released 0
}
