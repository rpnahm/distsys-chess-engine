package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rpnahm/distsys-chess-engine/pkg/client"
)

func main() {
	fmt.Println("Hello from Client Main")
	// handle command line input
	if len(os.Args) != 2 {
		log.Fatal("Usage: ./server <BaseServerName>")
	}

	eng := client.Init(os.Args[1], 1)

	eng.Connect(0)
}
