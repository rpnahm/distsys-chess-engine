package main

import (
	"fmt"
	"log"
	"os"

	"github.com/notnil/chess/uci"
	"github.com/rpnahm/distsys-chess-engine/pkg/client"
)

func main() {
	fmt.Println("Hello from Client Main")
	// handle command line input
	if len(os.Args) != 2 {
		log.Fatal("Usage: ./server <BaseServerName>")
	}

	eng := client.Init(os.Args[1], 2)

	err := eng.ConnectAll()
	if err != nil {
		log.Fatal("Unable to connect to all servers: ", err)
	}

	err = eng.NewGame(*eng.Game.Position(), []uci.CmdSetOption{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Success")
	eng.Shutdown()
}
