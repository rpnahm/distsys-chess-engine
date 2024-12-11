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
newgame:
	//start new game
	err = eng.NewGame(*eng.Game.Position(), []uci.CmdSetOption{})
	if err != nil {
		log.Fatal(err)
	}
	//loop as long as there is no outcome in the game
	for eng.Game.Outcome() == chess.NoOutcome {
		reader := bufio.NewReader(os.Stdin)
		//clear terminal before printing board
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

		fmt.Println(eng.Game.Position().Board().Draw())
		//loop waiting for valid move input
		for {
			fmt.Printf("Enter a valid move:")
			move, _ := reader.ReadString('\n')
			move = move[:len(move)-1]
			//check if user wants to quit
			if move == "quit" {
				log.Fatal("Game Ended, No Outcome")
				eng.Shutdown()

			}
			//make move
			err = eng.Game.MoveStr(move)
			if err != nil {
				cmd = exec.Command("clear")
				cmd.Stdout = os.Stdout
				cmd.Run()
				fmt.Println("Invalid move\nValid Moves:")
				//print valid moves
				moves := eng.Game.ValidMoves()
				for i := 0; i < len(moves); i++ {
					fmt.Printf("%s\n", moves[i])
				}
				fmt.Println(eng.Game.Position().Board().Draw())
				continue
			} else {
				//exit loop if user gives valid move
				break
			}
		}

		cmd = exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

		fmt.Println(eng.Game.Position().Board().Draw())
		//make computer move
		eng.Run()
	}

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	fmt.Println(eng.Game.Position().Board().Draw())

	//print outcome of game
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

	//check if user wants to play again
	log.Println("Play Again?(y/n)")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = response[:len(response)-1]
	if response == "y" {
		//start a new game and go to beginning of game
		eng.Game = *chess.NewGame(chess.UseNotation(chess.UCINotation{}))
		goto newgame
	}

	eng.Shutdown()
}
