package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cyberluisda/saverserver-go/server"
)

var user bool

func main() {
	if len(os.Args) < 2 {
		fmt.Printf(
			`Usage: %s address.
	address is the listening address, protocol included, Examples: udp://localhost:12000, tcp://localhost:13000.
	With random ports: tcp://localhost:0 or udp://localhost:0
`,
			os.Args[0],
		)
		os.Exit(1)
	}

	var bs server.BasicServer

	listenerProtocols := []string{"tcp", "tcp4", "tcp6", "unix", "unixpacket"}
	isListener := false
	for _, proto := range listenerProtocols {
		if strings.HasPrefix(os.Args[1], proto) {
			isListener = true
			break
		}
	}

	callFunc := func(a string, bs []byte) bool {
		if display {
			fmt.Printf("%s - ", a)
		}
		fmt.Print(string(bs))
		return false
	}

	if isListener {
		lst := server.Listener{
			ConnectionMgr: server.ConnectionMgr{
				Address: address,
			},
		}
		lst.CallBack = callFunc

		bs = &lst
	} else {
		lp := server.ListenerPacket{
			Address: address,
		}
		lp.CallBack = callFunc

		bs = &lp
	}

	err := bs.Start()
	if err != nil {
		log.Fatalf("Error while starts the server: %v", err)
	}

	log.Printf("Listening on address: %s, port %d\n", bs.GetAddress(), bs.Port())

	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	<-cancelChan
	err = bs.Stop()
	if err != nil {
		log.Fatalf("Error while stops the server: %v", err)
	}
}
