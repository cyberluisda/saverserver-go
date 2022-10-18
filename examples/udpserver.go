package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyberluisda/saverserver-go/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf(
			`Usage: %s address.
	address is the listening address, protocol included, Examples: udp://localhost:13000, udp4://localhost:0 (random port)
`,
			os.Args[0],
		)
		os.Exit(1)
	}

	lp := server.ListenerPacket{
		Address: os.Args[1],
	}

	err := lp.Start()
	if err != nil {
		log.Fatalf("Error while starts the server: %v", err)
	}

	log.Printf("Listening on address: %s, port %d\n", lp.GetAddress(), lp.Port())

	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	<-cancelChan
	err = lp.Stop()
	if err != nil {
		log.Fatalf("Error while stops the server: %v", err)
	}

	for k, v := range lp.GetPayloads() {
		fmt.Println(k, "->", string(v))
	}
}
