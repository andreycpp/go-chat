package main

import (
	"flag"
	"log"
	"net"

	"chat/server"
	"chat/userdb"
)

func main() {
	addr := flag.String("addr", ":9999", "listen address")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		panic(err)
	}
	log.Printf("Listening for connections on %v", *addr)
	defer ln.Close()

	s := server.NewServer(userdb.NewInMemoryUserDb())

	// start an infinite server loop
	go s.Run()

	// infinitely accept connections and send to server's join chan
	for {
		c, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		log.Printf("New client connected %v", c.RemoteAddr())
		s.Join <- server.NewClient(c, s)
	}
}
