package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rpnahm/distsys-chess-engine/pkg/server"
)

func main() {
	fmt.Println("Hello from the server Executable")

	// handle command line input
	if len(os.Args) != 2 {
		log.Fatal("Usage: ./server <serverName>")
	}

	// start the engine and server
	worker := server.Startup()
	worker.SetName(os.Args[1])

	// Run a separate thread that communicates with the nameserver
	go worker.CatalogMessage("rnahm")

	worker.Run()

}
