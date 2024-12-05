package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"

	"github.com/rpnahm/distsys-chess-engine/pkg/client"
)

/* This main file will be used to create a simple gui to play chess against. We will use another
   go file to make a test script that will test the performance by running the cluster against a local
   version of stockfish... notnils/chess has a lot of documentation on it. I'm also not sure how to redraw
   the board in the same position so that we aren't creating something really long to scroll down in the terminal
*/

func main() {
	fmt.Println("Hello from Client Main")
	// handle command line input
	if len(os.Args) != 2 {
		log.Fatal("Usage: ./server <BaseServerName>")
	}

	eng := client.Init(os.Args[1], 1)
	//eng.Game.UseNotation = *chess.NewGame()
	err := eng.ConnectAll()
	if err != nil {
		log.Fatal("Unable to connect to all servers: ", err)
	}

	err = eng.NewGame(*eng.Game.Position(), []uci.CmdSetOption{})
	if err != nil {
		log.Fatal(err)
	}

	for eng.Game.Outcome() == chess.NoOutcome {
		reader := bufio.NewReader(os.Stdin)
		// select a random move

		eng.Run() //add error handling

		//clear board before printing game state to avoid stacking boards
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

		fmt.Println(eng.Game.Position().Board().Draw())
		for {
			fmt.Printf("Enter a valid move:")
			move, _ := reader.ReadString('\n')
			move = move[:len(move)-1]
			err = eng.Game.MoveStr(move)
			if err != nil {
				fmt.Println("Invalid move\nValid Moves:")
				moves := eng.Game.ValidMoves()
				for i := 0; i < len(moves); i++ {
					fmt.Printf("%s\n", moves[i])
				}
				continue
			} else {
				break
			}
		}
	}

	err = eng.NewPos(*eng.Game.Position())
	if err != nil {
		log.Println("Unable to update position", err)
	}
	log.Println("Success")
	eng.Shutdown()
}
