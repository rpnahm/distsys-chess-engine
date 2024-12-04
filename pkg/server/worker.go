package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
	"github.com/rpnahm/distsys-chess-engine/pkg/common"
)

type Worker struct {
	name     string
	listener net.Listener
	address  string
	port     string
	eng      *uci.Engine
	game     *chess.Game
	conn     net.Conn
	posId    int
}

// struct for json messages to catalog server
type message struct {
	Type    string `json:"type"`
	Owner   string `json:"owner"`
	Port    string `json:"port"`
	Project string `json:"project"`
}

// Create a worker instance and start listening and such
func Startup() *Worker {

	// startup server
	w := &Worker{}

	e, err := uci.New("bin/stockfish")
	if err != nil {
		log.Fatal("Unable to start server", err)
	}
	w.eng = e

	// start listening on any address and any port
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		log.Fatal("Error opening listener", err)
	}

	w.address = ln.Addr().String()
	_, port, err := net.SplitHostPort(w.address)
	if err != nil {
		log.Fatal("Error splitting port from address")
	}
	w.port = port
	w.listener = ln

	return w
}

// Set the name of the server
func (w *Worker) SetName(name string) {
	w.name = name
}

// Run the worker, handles the main for loop
func (w *Worker) Run() {
	defer w.listener.Close()
	defer w.eng.Close()

	// Start UCI on engine
	err := w.eng.Run(uci.CmdUCI, uci.CmdIsReady)
	if err != nil {
		log.Fatal("Unable to start UCI on engine:", err)
	}

	for {
		w.conn, err = w.listener.Accept()
		if err != nil {
			log.Println("Error accepting connection", err)
		}
		w.handle()
	}
}

// Handles the operations of a single client connection
func (w *Worker) handle() {
	defer w.conn.Close()

	buf := make([]byte, common.BufSize)

	// loop forever for each client
	for {
		n, err := w.conn.Read(buf)
		if err != nil {
			log.Println("Unable to read connection: ", err)
			break
		}

		// *** FROM HERE ON WE HAVE TO REPORT ERRORS TO THE CLIENT ***
		// decode the json
		var request map[string]interface{}
		err = json.Unmarshal(buf[:n], &request)
		if err != nil {
			w.reportError(fmt.Sprint("Error unmarshalling json data: ", buf[:n], err))
			continue
		}

		// Switch to run the correct operation
		opType := request["type"]
		switch opType {
		case "new_game":
			// Handle newgame request
			w.newGame(buf[:n])
			continue
		case "parse_moves":
			// Handle parsemoves request
			continue
		case "new_pos":
			// Handle newpos request
			w.newPos(buf[:n])
			continue
		case "stop":
			// Handle stop request
			return
		case "exit":
			// Shut down server
			log.Fatal("Shutting Down Server")
		default:
			w.reportError(fmt.Sprint("Unknown message type: ", opType))
		}
	}
}

// Handle a newgame request
func (w *Worker) newGame(data []byte) {
	// unmarshall the data
	var info common.NewGame
	err := json.Unmarshal(data, &info)
	if err != nil {
		w.reportError(fmt.Sprint("Unable to decode new_game JSON, ", data, err))
		log.Println("Unable to decode new_game JSON, ", err)
	}

	// Reset the posid (only matters within each game)
	w.posId = info.PosId

	// set the options or the instance
	// First interpret each option as a CmdSetOption
	var options []uci.Cmd
	for _, option_string := range info.Options {
		// split the string by whitespace
		temp := strings.Fields(option_string)
		// test for whitespace
		if len(temp) > 2 {
			w.reportError(fmt.Sprint("Unable to decode option: ", option_string))
			log.Println("Unable to decode option: ", option_string)
		}
		if len(temp) == 2 {
			options = append(options, uci.CmdSetOption{Name: temp[0], Value: temp[1]})
		} else {
			options = append(options, uci.CmdSetOption{Name: temp[0], Value: ""})
		}
	}

	//now run the options on the engine
	err = w.eng.Run(options...)
	if err != nil {
		w.reportError(fmt.Sprint("Unable to run options", err))
		log.Println("Unable to run options", err)
	}

	// Set a new game board
	fen, err := chess.FEN(info.Position)
	if err != nil {
		w.reportError(fmt.Sprint("Unable to decode FEN string: ", info.Position, err))
		log.Println("Unable to decode Fen string: ", info.Position, err)
	}
	// interpret the starting position
	w.game = chess.NewGame(fen)
	pos := uci.CmdPosition{Position: w.game.Position()}
	// reset the engine game and set the position
	err = w.eng.Run(uci.CmdUCINewGame, pos, uci.CmdIsReady)
	if err != nil {
		w.reportError(fmt.Sprint("Unable to run ucinewgame, position command on engine", err))
		log.Println("Unable to run ucinewgame, position command on engine", err)
	}

	// Now the engine is correctly configured so we can return readyok
	oJson := common.ReadyOk{
		Type:  "ready_ok",
		PosId: w.posId,
	}
	// skipping error checking because it's within our control
	oData, _ := json.Marshal(&oJson)
	// send to the client
	_, err = w.conn.Write(oData)
	if err != nil {
		log.Println("Unable to send data to conn @ ", w.conn.RemoteAddr(), err)
	}
}

// Set a new position of the game
func (w *Worker) newPos(data []byte) {
	// must unmarshall data first
	_ = data
}

// Function to stop the worker from considering the current case
func (w *Worker) Stop() {
	err := w.eng.Run(uci.CmdStop)
	if err != nil {
		w.reportError(fmt.Sprint("Error stopping engine: ", err))
		log.Println("Error stopping engine: ", err)
	}
}

// Send an error message back to the client
func (w *Worker) reportError(errString string) {
	output := common.Error{
		Type:   "error",
		Reason: errString,
	}

	data, err := json.Marshal(output)
	if err != nil {
		log.Println("Unable to marshall error data ", output, err)
	}

	_, err = w.conn.Write(data)
	if err != nil {
		log.Println("Unable to send errror data ", output, err)
	} else {
		log.Println("Successfully sent error ", output)
	}
}

// Send the server info to the catalog once per minute
func (w *Worker) CatalogMessage(owner, project string) {
	m := message{
		Type:    w.name,
		Owner:   owner,
		Project: project,
		Port:    w.port,
	}

	// encode the json data
	jsonData, err := json.Marshal(m)
	if err != nil {
		log.Fatal("Error Marshalling Name-Server json", err)
	}

	nsAddressString := fmt.Sprintf("%s:%d", common.CatalogAddr, common.CatalogPort)
	nsAddress, err := net.ResolveUDPAddr("udp", nsAddressString)
	if err != nil {
		log.Fatal("Error resolving Name Server Address", err)
	}

	// connect to nameserver and update every 60 seconds
	conn, err := net.Dial("udp", nsAddress.String())
	if err != nil {
		log.Fatal("Error connecting to Nameserver for posting:", err)
	}
	defer conn.Close()

	for {
		_, err = conn.Write(jsonData)
		if err != nil {
			log.Fatal("Error sending message to Nameserver", err)
		}
		time.Sleep(1 * time.Minute)
	}

}
