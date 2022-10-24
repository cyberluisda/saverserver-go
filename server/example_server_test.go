package server

import (
	"crypto/tls"
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

//nolint:lll
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
		_, err = conn.Write(msg)
		if err != nil {
			panic(
				fmt.Sprintf("While send data to %s: %v", tcpAddr, err),
			)
		}
		if i == 500 {
			fmt.Println("Accepting while client is sending data", lst.Accepting())
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
		_, err = conn.Write(msg)
		if err != nil {
			panic(
				fmt.Sprintf("While send data to %s: %v", tcpAddr, err),
			)
		}

		if i == 500 {
			fmt.Println("Connections while client is sending data", lst.Connections())
		}
	}

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

func ExampleListener_CallBack() {
	lst := Listener{}
	lst.CallBack = func(a string, b []byte) bool {
		fmt.Println(string(b))
		return false
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

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(
			fmt.Sprintf("while connect to address %s address: %v", tcpAddr, err),
		)
	}

	msg := "This is a test"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		panic(
			fmt.Sprintf("while send data to socket: %v", err),
		)
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

	//Output:
	// This is a test
	// #Items 0
}

func ExampleListenerPacket_random_port() {
	lp := ListenerPacket{}
	err := lp.Start()
	if err != nil {
		panic(err)
	}
	p := lp.Port()
	if p > 1024 && p < 65536 {
		fmt.Println("OK")
	} else {
		fmt.Printf("Port returned and unexpected value: %d\n", p)
	}

	err = lp.Stop()
	if err != nil {
		panic(err)
	}

	//Output:
	// OK
}
func ExampleListenerPacket_protocol_error() {
	lp := ListenerPacket{
		Address: "tcp://:1536",
	}
	err := lp.Start()
	if err == nil {
		fmt.Println("No Error")
	} else {
		fmt.Println("Error", err)
	}

	//Output:
	// Error while starts listener: listen tcp :1536: address :1536: unexpected address type, 'tcp' ':1536'
}
func ExampleListenerPacket_Accepting() {
	lst := ListenerPacket{}

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

func ExampleListenerPacket_Connections() {
	fmt.Println("Number of connections must be 0 in all cases. UDP style communications does not support connections")
	lp := ListenerPacket{}

	fmt.Println("Connections before start", lp.Connections())

	err := lp.Start()
	if err != nil {
		panic(err)
	}
	if lp.Port() <= 0 {
		panic("Port unknown")
	}
	fmt.Println("Connections after start", lp.Connections())

	addr, err := net.ResolveUDPAddr("udp", strings.TrimPrefix(lp.GetAddress(), "udp://"))
	if err != nil {
		panic(
			fmt.Sprintf("while resolve address: %v", err),
		)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(
			fmt.Sprintf("while connect to address %s address: %v", addr, err),
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
		if i == 9 {
			fmt.Println("Connections while client is sending data", lp.Connections())
		}
	}

	err = conn.Close()
	if err != nil {
		panic(
			fmt.Sprintf("while close client: %v", err),
		)
	}

	err = lp.Stop()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connections after stop", lp.Connections())

	//Output:
	// Number of connections must be 0 in all cases. UDP style communications does not support connections
	// Connections before start 0
	// Connections after start 0
	// Connections while client is sending data 0
	// Connections after stop 0
}

//nolint:lll
func ExampleListenerPacket() {
	lp := ListenerPacket{}
	err := lp.Start()
	if err != nil {
		panic(err)
	}
	if lp.Port() <= 0 {
		panic("Port unknown")
	}

	addr, err := net.ResolveUDPAddr("udp", strings.TrimPrefix(lp.GetAddress(), "udp://"))
	if err != nil {
		panic(
			fmt.Sprintf("while resolve address: %v", err),
		)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(
			fmt.Sprintf("while connect to address %s address: %v", addr, err),
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

	time.Sleep(time.Millisecond * 100) // Wait to ensure data was received

	err = lp.Stop()
	if err != nil {
		panic(err)
	}

	fmt.Println("#Items", lp.NPayloadItems())
	rAddress := lp.GetPayloadAddresses()[0]
	fmt.Println(string(lp.GetPayload(rAddress)))

	//Output:
	// #Items 1
	//  this is test number 0  this is test number 1  this is test number 2  this is test number 3  this is test number 4  this is test number 5  this is test number 6  this is test number 7  this is test number 8  this is test number 9
}

func ExampleListenerPacket_CallBack() {
	lp := ListenerPacket{}
	lp.CallBack = func(a string, b []byte) bool {
		fmt.Println(string(b))
		return false
	}
	err := lp.Start()
	if err != nil {
		panic(err)
	}
	if lp.Port() <= 0 {
		panic("Port unknown")
	}

	addr, err := net.ResolveUDPAddr("udp", strings.TrimPrefix(lp.GetAddress(), "udp://"))
	if err != nil {
		panic(
			fmt.Sprintf("while resolve address: %v", err),
		)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(
			fmt.Sprintf("while connect to address %s address: %v", addr, err),
		)
	}

	msg := "This is the test"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		panic(
			fmt.Sprintf("while send data to socket: %v", err),
		)
	}

	err = conn.Close()
	if err != nil {
		panic(
			fmt.Sprintf("while close client: %v", err),
		)
	}

	time.Sleep(time.Millisecond * 100) // Wait to ensure data was received

	err = lp.Stop()
	if err != nil {
		panic(err)
	}

	fmt.Println("#Items", lp.NPayloadItems())

	//Output:
	// This is the test
	// #Items 0
}

func ExampleTLSListener() {
	lst := TLSListener{
		CertPem: []byte(`-----BEGIN CERTIFICATE-----
MIIFbTCCA1WgAwIBAgIUBuOu++jpX8WN+gv03btFbAbqitgwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAgFw0yMjEwMjQxMDI4MTlaGA8yMTIy
MDkzMDEwMjgxOVowRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUx
ITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDCCAiIwDQYJKoZIhvcN
AQEBBQADggIPADCCAgoCggIBANb1LSQ+pYzkSzOHwObC7TtL48/PdWFaoil6rtjp
T1wVEoaTUJg/V/kuV5mr3iFYOsp2C8OqcNiRg3Ez+bn/VaAqV+ON/FGI61I2zukd
57yX3DJXsoxF5lNx80BfqaDGhPs7YC3EzqRgg7T1wJpWt++I6TO2kLw0UZMb7oBd
kLlEcPX+DrBYB6J/+yZFN5J/7TuXVJVy8z8CMY3L7rA2C6rw6g6SM/aZdvycmi9z
J9m/CZjgmFT9EZeIhphrUb+nV+p7juGirAPm37+l5Y1TpoKyzlaMfXtu/l9TX+jz
uOw7oQ8ZdxUDCA8441g3E0M2UYG5dm67UAt0LPtFowWQfrWj/pD3vCmkOyIWOm9o
HkLm2Ue/eiOBiauz6XTBmef9FVxbtauPC553mDRtQwfRFIuPWXDC2WcUmnmHDC7l
ouSE7DU16l0TMDEl2NZAkKNW5ZVzCgjOjCbgiitwHhIFliH+9wMTrClSgusrANDw
a8xv4nrIA00v90eu8vblk9Rx/UUGQ9D15xK3rAtw6OohW93vESH5I1fGfUCdsa+o
1b0T0cVyUxp30iYCgO8qLsZq4JLD6ikf+9qwXoQL7IdOBSqH6kjmqdmzVbNvo/Se
NizVTRG6hVCWeg2ZOHVDICWhC+G3YAoUlT36SQFJibGYI0W8v+cv0CUvGeGkzaC2
y/cFAgMBAAGjUzBRMB0GA1UdDgQWBBTKTzI9KELxrCZ0RC+3AbB9Y3m6ATAfBgNV
HSMEGDAWgBTKTzI9KELxrCZ0RC+3AbB9Y3m6ATAPBgNVHRMBAf8EBTADAQH/MA0G
CSqGSIb3DQEBCwUAA4ICAQCrP8XstH4JHKism3V6oNqH38ZWfhnLOplVZ79WctSF
D0qCeP1cRK127VC2YZxDlRgdHlvFVYZLp4DvT8nYHLlYtqnwI0fR4N7Siq/swOvt
YbHNEtGaVeA/jSwc8VyLo1XtV9EpBqIJVp9cp6c0YUQlksH2tmP55r2laCiqVVQv
9SgSzDyKeRbqV9KE3A6YJvSYvabHQPU1YL8mem2T34JKCv2+raJGR3+tfx8DGDIT
VpPsWsCseM1CxXFlrJLy042f/hhldexvG/qSdh9bTUui5bovc876YHeGgzVD9e+F
HEU+pdJgh/KdJ+J35VpMyVuT6IcTprVpn9vKcetNKOMeVWOrJovJKzVmuFXEcetl
Wg1I9ggxS8p+3GFoimvg6WLdsQkOWl+9N0PuesGLwYIlkKoZvsI5Mgii/2vimknz
Uk5y/xpaDktCWmKNCDHLvfUVnXBi2Kg+h3+IoQwDE3vEFArAe/MAHuTrVgzz5PcA
D+J+s0VMw6fLpRh84H529EKxSzVYkPtm4AvY4e/Tclm8N0ybf5byARkc3g3Ow8vq
m9JRMJOmgAX99g+MP8GcIZNzd568kE+6PRLQsgWcn6fvrdz+MhZ4pWKIVMaTjZSB
dAT40ewxeO9p9gapTidLrVv5XHsAtq0tDMyH9N7B3Gqlboo8O5eM2vif8cCHy1jA
bw==
-----END CERTIFICATE-----
`),
		KeyPem: []byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`),
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

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(
			fmt.Sprintf("while connect to address %s address: %v", tcpAddr, err),
		)
	}
	cfg := &tls.Config{
		InsecureSkipVerify: true,
	}
	connTLS := tls.Client(conn, cfg)
	err = connTLS.Handshake()
	if err != nil {
		panic(
			fmt.Sprintf("while make TLS handshake to address %s address: %v", tcpAddr, err),
		)
	}

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf(" this is test number %d ", i)
		_, err = connTLS.Write([]byte(msg))
		if err != nil {
			panic(
				fmt.Sprintf("while send data to socket: %v", err),
			)
		}
	}

	err = connTLS.Close()
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

func ExampleTLSListener_client_certs() {
	lst := TLSListener{
		ClientAuth: tls.RequestClientCert,
		CertPem: []byte(`-----BEGIN CERTIFICATE-----
MIIFbTCCA1WgAwIBAgIUBuOu++jpX8WN+gv03btFbAbqitgwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAgFw0yMjEwMjQxMDI4MTlaGA8yMTIy
MDkzMDEwMjgxOVowRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUx
ITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDCCAiIwDQYJKoZIhvcN
AQEBBQADggIPADCCAgoCggIBANb1LSQ+pYzkSzOHwObC7TtL48/PdWFaoil6rtjp
T1wVEoaTUJg/V/kuV5mr3iFYOsp2C8OqcNiRg3Ez+bn/VaAqV+ON/FGI61I2zukd
57yX3DJXsoxF5lNx80BfqaDGhPs7YC3EzqRgg7T1wJpWt++I6TO2kLw0UZMb7oBd
kLlEcPX+DrBYB6J/+yZFN5J/7TuXVJVy8z8CMY3L7rA2C6rw6g6SM/aZdvycmi9z
J9m/CZjgmFT9EZeIhphrUb+nV+p7juGirAPm37+l5Y1TpoKyzlaMfXtu/l9TX+jz
uOw7oQ8ZdxUDCA8441g3E0M2UYG5dm67UAt0LPtFowWQfrWj/pD3vCmkOyIWOm9o
HkLm2Ue/eiOBiauz6XTBmef9FVxbtauPC553mDRtQwfRFIuPWXDC2WcUmnmHDC7l
ouSE7DU16l0TMDEl2NZAkKNW5ZVzCgjOjCbgiitwHhIFliH+9wMTrClSgusrANDw
a8xv4nrIA00v90eu8vblk9Rx/UUGQ9D15xK3rAtw6OohW93vESH5I1fGfUCdsa+o
1b0T0cVyUxp30iYCgO8qLsZq4JLD6ikf+9qwXoQL7IdOBSqH6kjmqdmzVbNvo/Se
NizVTRG6hVCWeg2ZOHVDICWhC+G3YAoUlT36SQFJibGYI0W8v+cv0CUvGeGkzaC2
y/cFAgMBAAGjUzBRMB0GA1UdDgQWBBTKTzI9KELxrCZ0RC+3AbB9Y3m6ATAfBgNV
HSMEGDAWgBTKTzI9KELxrCZ0RC+3AbB9Y3m6ATAPBgNVHRMBAf8EBTADAQH/MA0G
CSqGSIb3DQEBCwUAA4ICAQCrP8XstH4JHKism3V6oNqH38ZWfhnLOplVZ79WctSF
D0qCeP1cRK127VC2YZxDlRgdHlvFVYZLp4DvT8nYHLlYtqnwI0fR4N7Siq/swOvt
YbHNEtGaVeA/jSwc8VyLo1XtV9EpBqIJVp9cp6c0YUQlksH2tmP55r2laCiqVVQv
9SgSzDyKeRbqV9KE3A6YJvSYvabHQPU1YL8mem2T34JKCv2+raJGR3+tfx8DGDIT
VpPsWsCseM1CxXFlrJLy042f/hhldexvG/qSdh9bTUui5bovc876YHeGgzVD9e+F
HEU+pdJgh/KdJ+J35VpMyVuT6IcTprVpn9vKcetNKOMeVWOrJovJKzVmuFXEcetl
Wg1I9ggxS8p+3GFoimvg6WLdsQkOWl+9N0PuesGLwYIlkKoZvsI5Mgii/2vimknz
Uk5y/xpaDktCWmKNCDHLvfUVnXBi2Kg+h3+IoQwDE3vEFArAe/MAHuTrVgzz5PcA
D+J+s0VMw6fLpRh84H529EKxSzVYkPtm4AvY4e/Tclm8N0ybf5byARkc3g3Ow8vq
m9JRMJOmgAX99g+MP8GcIZNzd568kE+6PRLQsgWcn6fvrdz+MhZ4pWKIVMaTjZSB
dAT40ewxeO9p9gapTidLrVv5XHsAtq0tDMyH9N7B3Gqlboo8O5eM2vif8cCHy1jA
bw==
-----END CERTIFICATE-----
`),
		KeyPem: []byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`),
	}
	err := lst.Start()
	if err != nil {
		panic(err)
	}
	if lst.Port() <= 0 {
		panic("Port unknown")
	}

	clientCrt, err := tls.X509KeyPair(
		[]byte(`-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----
`),
		[]byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`),
	)
	if err != nil {
		panic(
			fmt.Sprintf("While load client certificate: %v", err),
		)
	}
	cfg := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{clientCrt},
	}

	addr := strings.TrimPrefix(lst.GetAddress(), "tcp://")
	conn, err := tls.Dial("tcp", addr, cfg)
	if err != nil {
		panic(
			fmt.Sprintf("while connect to address %s: %v", addr, err),
		)
	}

	state := conn.ConnectionState()
	fmt.Println("Handshake: ", state.HandshakeComplete)

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
	fmt.Println("Client cert:", strings.Split(rAddress, "@")[0])
	fmt.Println(string(lst.GetPayload(rAddress)))

	//Output:
	// Handshake:  true
	// #Items 1
	// Client cert: CN=client.com,OU=IT,O=Random Company,L=Earth,ST=NRW,C=DE,1.2.840.113549.1.9.1=#0c0c7573657240666f6f2e636f6d
	//  this is test number 0  this is test number 1  this is test number 2  this is test number 3  this is test number 4  this is test number 5  this is test number 6  this is test number 7  this is test number 8  this is test number 9
}
