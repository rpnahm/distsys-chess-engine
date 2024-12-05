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

// *** UPDATES NEEDED ***
// Client makes first move
// re-print board after client and server move
// num servers  + movetime => init arguments + command line
// Endgame handling (print out final board state and result)

// Later on: game select and such

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
