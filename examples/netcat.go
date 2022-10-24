package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cyberluisda/saverserver-go/server"
)

func main() {
	var tlsKey string
	flag.StringVar(&tlsKey, "tlskey", "", "Key file to enable tls (over connections like tcp)")
	var tlsCert string
	flag.StringVar(&tlsCert, "tlscert", "", "Certificate file to enable tls (over connections like tcp)")
	var user bool
	flag.BoolVar(&user, "user", false, "If true, client authentication with certificate is required")
	var display bool
	flag.BoolVar(
		&display,
		"display",
		false,
		"If true, address and user certificate info (if available) is displayed before each incoming data",
	)
	var address string
	flag.StringVar(
		&address,
		"address",
		"",
		"listening address, protocol included, Examples: udp://localhost:12000, tcp://localhost:13000. "+
			"To choose a random port: tcp://localhost:0 or udp://localhost:0")
	flag.Parse()

	if address == "" {
		fmt.Printf(
			`Usage: %s [ --tlskey keyfile.key --tlscert certfile.crt [ --user ] ] [ --display ] --address address.
	address is the listening address, protocol included, Examples: udp://localhost:12000, tcp://localhost:13000.
	To choose random ports: tcp://localhost:0 or udp://localhost:0

	keyfile.key and certfile.crt are the key and certificate file to enable tls connections. Only valid for tcp
	and similar connections.

`,
			os.Args[0],
		)
		os.Exit(1)
	}

	var bs server.BasicServer

	listenerProtocols := []string{"tcp", "tcp4", "tcp6", "unix", "unixpacket"}
	isListener := false
	for _, proto := range listenerProtocols {
		if strings.HasPrefix(address, proto) {
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

	if tlsKey != "" && tlsCert != "" {
		bsKey, err := ioutil.ReadFile(tlsKey)
		if err != nil {
			log.Fatalf("Error while load key file %s: %v", tlsKey, err)
		}

		bsCert, err := ioutil.ReadFile(tlsCert)
		if err != nil {
			log.Fatalf("Error while load certificate file %s: %v", tlsCert, err)
		}

		tll := server.TLSListener{
			ConnectionMgr: server.ConnectionMgr{
				Address: address,
			},
			PayloadStorage: server.PayloadStorage{
				CallBack: callFunc,
			},
			KeyPem:  bsKey,
			CertPem: bsCert,
		}

		if user {
			tll.ClientAuth = tls.RequireAnyClientCert
		}

		bs = &tll

	} else if isListener {
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
