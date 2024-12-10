package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"

	"github.com/rpnahm/distsys-chess-engine/pkg/client"
)

// *** UPDATES NEEDED ***

// Later on: game select and such

func main() {
	fmt.Println("Hello from Client Main")
	// handle command line input
	if len(os.Args) != 4 {
		log.Fatal("Usage: ./server <BaseServerName> <num servers> <turntime>")
	}
	turntime, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal("invalid turn time")
	}
	numservers, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("invalid turn time")
	}
	eng := client.Init(os.Args[1], numservers, time.Duration(turntime)*time.Millisecond, 50*time.Millisecond)
	//eng.Game.UseNotation = *chess.NewGame()
	err = eng.ConnectAll()
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

		//add error handling

		//clear board before printing game state to avoid stacking boards
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

		fmt.Println(eng.Game.Position().Board().Draw())
		for {
			fmt.Printf("Enter a valid move:")
			move, _ := reader.ReadString('\n')
			move = move[:len(move)-1]
			if move == "quit" {
				log.Fatal("Game Ended, No Outcome")
				eng.Shutdown()

			}
			err = eng.Game.MoveStr(move)
			if err != nil {
				cmd = exec.Command("clear")
				cmd.Stdout = os.Stdout
				cmd.Run()
				fmt.Println("Invalid move\nValid Moves:")
				moves := eng.Game.ValidMoves()
				for i := 0; i < len(moves); i++ {
					fmt.Printf("%s\n", moves[i])
				}
				fmt.Println(eng.Game.Position().Board().Draw())
				continue
			} else {
				break
			}
		}

		cmd = exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

		fmt.Println(eng.Game.Position().Board().Draw())

		eng.Run()
	}

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	fmt.Println(eng.Game.Position().Board().Draw())

	fmt.Println(eng.Game.Outcome())
	if eng.Game.Outcome() == chess.WhiteWon {
		fmt.Println("Checkmate. You Won!")
	} else if eng.Game.Outcome() == chess.BlackWon {
		fmt.Println("Checkmate. You Lost.")
	} else if eng.Game.Outcome() == chess.Draw {
		fmt.Println("Stalemate. You Tied!")
	}

	fmt.Println("Game Ended")

	log.Println("Success")
	eng.Shutdown()
}
