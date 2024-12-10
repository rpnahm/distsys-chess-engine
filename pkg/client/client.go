package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
	"github.com/rpnahm/distsys-chess-engine/pkg/common"
)

// struct to store all client data
type Client struct {
	baseServerName string
	numServers     int
	Game           chess.Game
	conns          []server
	posId          int
	jobId          int
	TurnTime       time.Duration
	latencyBuff    time.Duration
}

// Stores connection information about each server
type server struct {
	name  string
	conn  net.Conn
	jobId int
	ready bool
}

type newError struct {
	Code    int
	Message string
}

func (e *newError) Error() string {
	return fmt.Sprintf("Code %d, Error: %s\n", e.Code, e.Message)
}

// intialize the Client struct for operations
func Init(baseServer string, numServers int, turnTime time.Duration, latency time.Duration) *Client {
	c := &Client{baseServerName: baseServer, numServers: numServers, posId: 0, jobId: 0}

	c.TurnTime = turnTime
	c.latencyBuff = latency
	c.Game = *chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	for i := 0; i < c.numServers; i++ {
		name := fmt.Sprintf("%s-%02d", c.baseServerName, i)
		s := server{name: name, conn: nil, jobId: 0, ready: false}
		c.conns = append(c.conns, s)
	}
	return c
}

// Closes all connections
func (c *Client) Shutdown() {
	// stop message
	o := common.Stop{Type: "stop"}
	data, _ := json.Marshal(o)

	for _, server := range c.conns {
		server.conn.Write(data)
		server.conn.Close()
	}
}

// connects a single server
func (c *Client) Connect(serverNum int) error {
	// get response from server
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/query.json", common.CatalogAddr, common.CatalogPort))
	if err != nil {
		log.Fatal("Unable to contact catalog server", serverNum, err)
	}
	defer resp.Body.Close()

	// parse the input
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Unable to parse body of catalog", err)
		return err
	}

	var results []map[string]interface{}
	err = json.Unmarshal(body, &results)
	if err != nil {
		log.Println("Unable to decode json", err)
		return err
	}

	// Now we have the input in result we should
	// iterate over it now to find our server
	newTime := 0.0
	var newServerInfo map[string]interface{}
	for _, value := range results {
		if value["type"] == "chess-worker" {
			if value["project"] == c.conns[serverNum].name && newTime < value["lastheardfrom"].(float64) {
				newServerInfo = value
				newTime = value["lastheardfrom"].(float64)
			}
		}
	}

	// set the conn values to the correct state, and return
	c.conns[serverNum].conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", newServerInfo["address"], int(newServerInfo["port"].(float64))))
	if err != nil {
		c.conns[serverNum].conn = nil
		log.Println("Unable to connect to server: ", c.conns[serverNum].name)
		return err
	}
	c.conns[serverNum].ready = true
	return nil
}

// Connect to all servers
func (c *Client) ConnectAll() error {
	for i := 0; i < c.numServers; i++ {
		err := http.ErrAbortHandler
		tried := false
		for err != nil {
			err = c.Connect(i)
			if err != nil {
				if tried {
					return err
				}
				time.Sleep(common.Wait)
				tried = true
			}
		}
	}
	return nil
}

// sets up the newgame on all machines
func (c *Client) NewGame(position chess.Position, options []uci.CmdSetOption) error {
	o := common.NewGame{
		Type:     "new_game",
		Position: position.String(),
		PosId:    c.posId,
	}

	// add options to the message
	var opts []string
	for _, option := range options {
		opts = append(opts, fmt.Sprintf("%s %s", option.Name, option.Value))
	}

	o.Options = opts

	// Send to all clients and get their responses one at a time
	oData, err := json.Marshal(o)
	if err != nil {
		return err
	}
	err = c.sendAll(oData)
	if err != nil {
		return err
	}
	return nil
}

// Sends the same message to all clients expects ReadyOk
func (c *Client) sendAll(data []byte) error {
	buf := make([]byte, common.BufSize)
	var response map[string]interface{}

	for i, server := range c.conns {
		// loop forever until sent
		for {
			for {
				_, err := server.conn.Write(data)
				if err != nil {
					log.Println("Unable to send data to server:", i, err, "retrying")
					time.Sleep(common.Wait)
					server.conn.Close()
					c.Connect(i)
				} else {
					break
				}
			}

			// Now get readyok
			n, err := server.conn.Read(buf)
			if err != nil {
				log.Println("Unable to recieve readyok from server", err)
				server.conn.Close()
				c.Connect(i)
			} else {
				// decode json
				err = json.Unmarshal(buf[:n], &response)
				// bad json gets a return
				if err != nil {
					return err
				}
				// server error gets return
				if response["type"] == "error" {
					return &newError{Code: 1, Message: response["reason"].(string)}
				} else if response["type"] == "ready_ok" && int(response["pos_id"].(float64)) == c.posId { // ready_ok continues
					break
				} else { // random message gets a resend
					continue
				}
			}
		}
	}
	return nil
}

// updates the postition of all clients
func (c *Client) NewPos(position chess.Position) error {
	c.posId++
	o := common.NewPos{
		Type:     "new_pos",
		Position: position.String(),
		PosId:    c.posId,
	}

	// marshall the json
	data, err := json.Marshal(o)
	if err != nil {
		return err
	}

	// send the messages and expect a readyok
	err = c.sendAll(data)
	if err != nil {
		return err
	}

	return nil
}

// Main function that handles server operations
// Parses the current position, and returns the best move
func (c *Client) Run() (common.Results, error) {
	// build generic message
	// calculate duetime because it is the same for all servers
	dueTime := time.Now().Add(c.TurnTime - c.latencyBuff)
	base := common.ParseMoves{
		Type:     "parse_moves",
		Position: c.Game.FEN(),
		PosId:    c.posId,
		DueTime:  dueTime,
	}

	var readyServers []int
	// find out how many servers are ready
	for i, server := range c.conns {
		if server.ready {
			readyServers = append(readyServers, i)
		}
	}

	// create an array of messages for all servers
	var messages []common.ParseMoves
	for i := 0; i < len(readyServers); i++ {
		base.JobId = c.jobId + i
		messages = append(messages, base)
	}

	// Get the list of possible moves
	moves := c.Game.ValidMoves()
	assignments := len(moves)
	// Assign the moves to the servers
	for i, move := range moves {
		messages[i%len(readyServers)].Moves = append(messages[i%len(readyServers)].Moves, move.String())
	}

	// Iterate over servers and build + send their messages while splitting up moves and incrementing jobid's
	// figure out how many messages there actually are
	if assignments > len(readyServers) {
		assignments = len(readyServers)
	}
	for i, num := range readyServers {
		// break if the index is greater or equal to assignments
		if i >= assignments {
			break
		}
		// marshall and send the data
		data, err := json.Marshal(messages[i])
		if err != nil {
			log.Fatal("Unable to encode parse_moves json", err)
		}

		_, err = c.conns[num].conn.Write(data)
		if err != nil {
			err = c.Connect(i)
			if err != nil {
				return common.Results{}, err
			}
		}
	}

	var results []common.Results
	var result common.Results
	// listen for results message, combine them into some sort of datastructure and pick based off of score/mate
	// sleep until the deadline
	time.Sleep(time.Until(dueTime.Add(c.latencyBuff)))
	// Now perform all read results with very short blocking
	buf := make([]byte, common.BufSize)

	for i, num := range readyServers {
		if i > assignments {
			break
		}

		c.conns[num].conn.SetReadDeadline(time.Now().Add(100 * time.Microsecond))
		defer c.conns[num].conn.SetReadDeadline(time.Time{})
		// ignore errors, just skip
		n, err := c.conns[num].conn.Read(buf)
		if err != nil {
			log.Println(err)
			c.conns[num].conn.Close()
			c.conns[num].ready = false
			go c.Connect(num)
		}
		if n != 0 {
			json.Unmarshal(buf[:n], &result)
			results = append(results, result)
		}
	}

	// ouput results struct to handle testing
	// Apply the chosen move to the game
	// if results are empty
	if len(results) == 0 {
		// Choose a random move
		move := moves[rand.Intn(len(moves))]
		log.Println("No input from servers, choosing random move")
		c.Game.Move(move)
		return common.Results{}, nil
	}
	output := results[0]
	// loop through results
	if len(results) > 1 {
		for _, result := range results {
			// update nodes visited
			output.Nodes += result.Nodes

			// select best_move
			if output.Score < result.Score {
				output.BestMove = result.BestMove
				output.Score = result.Score
				output.Mate = result.Mate
			}
		}
	}

	// apply the move
	c.Game.MoveStr(output.BestMove)
	// Update position at all of the servers
	return output, nil
}
