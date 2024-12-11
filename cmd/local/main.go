package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
)

// This runs stockfish against itself
// should be run with 12 cores

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: ./bin/local <numGames>")
	}
	nGames, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Usage: ./bin/local <numGames>")
	}
	weak, err1 := uci.New("bin/stockfish")
	strong, err2 := uci.New("bin/stockfish")
	if err1 != nil || err2 != nil {
		log.Fatal("Unable tos start engines", err1, err2)
	}
	defer weak.Close()
	defer strong.Close()

	storagePerThread := 1024

	var options1 []uci.Cmd
	options1 = append(options1, uci.CmdUCI)
	options1 = append(options1, uci.CmdIsReady)
	options1 = append(options1, uci.CmdUCINewGame)
	options1 = append(options1, uci.CmdSetOption{Name: "Threads", Value: fmt.Sprint(1)})
	options1 = append(options1, uci.CmdSetOption{Name: "Hash", Value: fmt.Sprint(storagePerThread * 1)})

	options2 := options1
	options2[3] = uci.CmdSetOption{Name: "Threads", Value: fmt.Sprint(12)}
	options2[4] = uci.CmdSetOption{Name: "Hash", Value: fmt.Sprint(storagePerThread * 12)}

	// games loop
	// tracking information
	systemWins, systemDraws, systemLosses := 0, 0, 0
	cmdGo := uci.CmdGo{MoveTime: time.Second / 10}

	fd, _ := os.OpenFile(fmt.Sprintf("localonly-%dgames.log", nGames), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer fd.Close()

	for i := 0; i < nGames; i++ {
		// startup engines
		weak.Run(options1...)
		strong.Run(options2...)
		game := chess.NewGame()

		// Game loop
		for game.Outcome() == chess.NoOutcome {
      // slow move first
			cmdPos := uci.CmdPosition{Position: game.Position()}
			if i%2 == 0 {
				weak.Run(cmdPos, cmdGo)
				game.Move(weak.SearchResults().BestMove)

				if game.Outcome() != chess.NoOutcome {
					break
				}

				cmdPos = uci.CmdPosition{Position: game.Position()}
				strong.Run(cmdPos, cmdGo)
				game.Move(strong.SearchResults().BestMove)
			  } else {
				strong.Run(cmdPos, cmdGo)
				game.Move(strong.SearchResults().BestMove)

				if game.Outcome() != chess.NoOutcome {
					break
				}

				cmdPos = uci.CmdPosition{Position: game.Position()}
				weak.Run(cmdPos, cmdGo)
				game.Move(weak.SearchResults().BestMove)
			}
		}
		if game.Outcome() == chess.Draw {
			systemDraws++
		} else if i%2 == 0 {
			if game.Outcome() == chess.WhiteWon {
				systemLosses++
			} else {
				systemWins++
			}
		} else {
			if game.Outcome() == chess.WhiteWon {
				systemWins++
			} else {
				systemLosses++
			}
		}

		fd.WriteString(fmt.Sprintf("%d-%d-%d\n", systemWins, systemDraws, systemLosses))

	}
}
