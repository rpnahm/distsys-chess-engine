package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/rpnahm/distsys-chess-engine/pkg/server"
)

func main() {
	fmt.Println("Hello from the server Executable")

	// handle command line input
	if len(os.Args) != 2 {
		log.Fatal("Usage: ./server <serverName>")
	}

	// set the server name
	server.Name = os.Args[1]

	ln, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		log.Fatal("Error opening listner", err)
	}
	defer ln.Close()

	// Run a separate thread that communicates with the nameserver
	go server.CatalogMessage(server.Name, "rnahm", "chess-engine", ln.Addr().(*net.TCPAddr))

}
