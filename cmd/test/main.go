package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
	"github.com/rpnahm/distsys-chess-engine/pkg/client"
	"github.com/rpnahm/distsys-chess-engine/pkg/common"
)

func main() {
	// handle command line input
	if len(os.Args) != 6 {
		log.Fatal("Usage: ./server <BaseServerName> <numServers> <turnTime(ms)> <numGames> <threads>")
	}
	nServers, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("Must use integer for numServers", err)
	}
	t, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal("Must use integer for turn time", err)
	}
	turnTime := time.Duration(t) * time.Millisecond
	nGames, err := strconv.Atoi(os.Args[4])
	if err != nil {
		log.Fatal("Must use integer for nGames", err)
	}
	nThreads, err := strconv.Atoi(os.Args[5])
	if err != nil {
		log.Fatal("Must use integer for nThreads", err)
	}

	// Start up engines
	log.Println("Starting up engines")
	client := client.Init(os.Args[1], nServers, turnTime, 50*time.Millisecond)

	localEng, err := uci.New("bin/stockfish")
	if err != nil {
		log.Fatal("Unable to start local stockfish")
	}

	defer localEng.Close()

	// Connecting to all servers
	log.Println("Connecting to all servers")
	err = client.ConnectAll()
	if err != nil {
		log.Fatal("Unable to connect to all servers", err)
	}
	defer client.Shutdown()

	// tracking information
	systemWins, systemDraws, systemLosses := 0, 0, 0
	fd, err := os.OpenFile(fmt.Sprintf("%s-%d-%d-%d.log", os.Args[1], nServers, t, nThreads), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("Unable to open log file", err)
	}
	defer fd.Close()
	fd.WriteString("ServerNodesProcessed, ClientNodesProcessed\n")

	recordFd, err := os.OpenFile(fmt.Sprintf("%s-%d-%d-%d-record.log", os.Args[1], nServers, t, nThreads), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("Unable to open record log file", err)
	}
	defer recordFd.Close()

	var results common.Results

	log.Println("All servers connected")
	// games loop
	for game := 0; game < nGames; game++ {
		//setup each game
		// options
		var options []uci.CmdSetOption
		options = append(options, uci.CmdSetOption{Name: "Threads", Value: fmt.Sprint(nThreads)})
		options = append(options, uci.CmdSetOption{Name: "Hash", Value: fmt.Sprint((10240 * nThreads))})

		err = localEng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame)
		if err != nil {
			log.Fatal("Unable configure local stockfish", err)
		}
		var cmds []uci.Cmd
		for _, c := range options {
			cmds = append(cmds, c)
		}
		err = localEng.Run(cmds...)
		if err != nil {
			log.Fatal("Unable to run setoptions on local engine", err)
		}

		client.Game = *chess.NewGame(chess.UseNotation(chess.UCINotation{}))

		log.Printf("Remotely creating newgame %d:\n", game)
		err = client.NewGame(*client.Game.Position(), options)
		if err != nil {
			log.Fatal("Unable to start newgames on servers", err)
		}

		fmt.Println("Entering Game Loop")
		// game loop
		for client.Game.Outcome() == chess.NoOutcome {
			// Local Move
			if game%2 == 0 {

				cmdPos := uci.CmdPosition{Position: client.Game.Position()}
				cmdGo := uci.CmdGo{MoveTime: turnTime}
				localEng.Run(cmdPos, cmdGo)

				move := localEng.SearchResults().BestMove
				client.Game.Move(move)

				if client.Game.Outcome() != chess.NoOutcome {
					break
				}

				// server move
				results, err = client.Run()
				if err != nil {
					log.Println(err)
				}
			} else {
				// server move
				results, err = client.Run()
				if err != nil {
					log.Println(err)
				}

				if client.Game.Outcome() != chess.NoOutcome {
					break
				}

				cmdPos := uci.CmdPosition{Position: client.Game.Position()}
				cmdGo := uci.CmdGo{MoveTime: turnTime}
				localEng.Run(cmdPos, cmdGo)

				move := localEng.SearchResults().BestMove
				client.Game.Move(move)
			}
			fd.WriteString(fmt.Sprintf("%d, %d\n", results.Nodes, localEng.SearchResults().Info.Nodes))

		}

		log.Println(client.Game.Outcome(), client.Game.Method(), len(client.Game.MoveHistory()), "moves")
		if client.Game.Outcome() == chess.Draw {
			systemDraws++
		} else if game%2 == 0 {
			if client.Game.Outcome() == chess.WhiteWon {
				systemLosses++
			} else {
				systemWins++
			}
		} else {
			if client.Game.Outcome() == chess.BlackWon {
				systemWins++
			} else {
				systemLosses++
			}
		}
		recordFd.WriteString(fmt.Sprintf("%d-%d-%d\n", systemWins, systemDraws, systemLosses))
	}
	fd.WriteString(fmt.Sprintf("Distributed Chess Record against Local Stockfish:\n%d-%d-%d\n", systemWins, systemDraws, systemLosses))
}
