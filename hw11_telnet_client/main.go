package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Println("use: go-telnet [parameters] host port")
		return
	}

	address := net.JoinHostPort(flag.Arg(0), flag.Arg(1))
	client := NewTelnetClient(address, *timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		log.Fatalf("Cannot connect to %s: %v", address, err)
	}
	defer func(client TelnetClient) {
		err := client.Close()
		if err != nil {
			log.Fatalf("Cannot close connection: %v", err)
		}
	}(client)

	log.Printf("...Connected to %s \n", address)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		if err := client.Send(); err != nil {
			log.Println("...Send error:", err)
		}
		log.Printf("...EOF")
		cancel()
	}()
	go func() {
		if err := client.Receive(); err != nil {
			log.Println("...Receive error:", err)
		}
		log.Printf("...Connection was closed by peer")
		cancel()
	}()

	<-ctx.Done()
}
