package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	TurnTime       int
}

// Stores connection information about each server
type server struct {
	name  string
	conn  net.Conn
	posId int
	jobId int
}

type newError struct {
	Code    int
	Message string
}

func (e *newError) Error() string {
	return fmt.Sprintf("Code %d, Error: %s\n", e.Code, e.Message)
}

// intialize the Client struct for operations
func Init(baseServer string, num int) *Client {
	c := &Client{baseServerName: baseServer, numServers: num, posId: 0, jobId: 0}
	c.Game = *chess.NewGame()
	for i := 0; i < c.numServers; i++ {
		name := fmt.Sprintf("%s-%02d", c.baseServerName, i)
		s := server{name: name, conn: nil, posId: 0, jobId: 0}
		c.conns = append(c.conns, s)
	}
	return c
}

// Closes all connections
func (c *Client) Shutdown() {
	for _, server := range c.conns {
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
		if value["type"] == c.conns[serverNum].name && newTime < value["lastheardfrom"].(float64) {
			newServerInfo = value
			newTime = value["lastheardfrom"].(float64)
		}
	}

	// set the conn values to the correct state, and return
	c.conns[serverNum].conn, err = net.Dial("tcp", fmt.Sprintf("%s:%s", newServerInfo["address"], newServerInfo["port"]))
	if err != nil {
		c.conns[serverNum].conn = nil
		return err
	}
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
		opts = append(opts, option.String())
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
		server.posId = c.posId
		log.Println("Sending to server", server.name)
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
				} else if response["type"] == "ready_ok" && int(response["pos_id"].(float64)) == server.posId { // ready_ok continues
					break
				} else { // random message gets a resend
					continue
				}
			}
		}
	}
	return nil
}
