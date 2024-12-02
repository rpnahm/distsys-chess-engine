package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/notnil/chess"
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
}

// Stores connection information about each server
type server struct {
	name  string
	conn  net.Conn
	posId int
	jobId int
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

// connects a single client
func (c *Client) Connect(serverNum int) {
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
	}

	var results []map[string]interface{}
	err = json.Unmarshal(body, &results)
	if err != nil {
		log.Println("Unable to decode json", err)
	}

	// Now we have the input in result we should
	// iterate over it now to find our server
	newTime := 0
	var newServerInfo map[string]interface{}
	for _, value := range results {
		if value["type"] == c.conns[serverNum].name && newTime < int(value["lastheardfrom"].(float64)) {
			newServerInfo = value
		}
	}

	fmt.Printf("%s:%s\n", newServerInfo["address"], newServerInfo["port"])
	// set the conn values to the correct state, and return
	c.conns[serverNum].conn, err = net.Dial("tcp", fmt.Sprintf("%s:%s", newServerInfo["address"], newServerInfo["port"]))
	if err != nil {
		log.Println("Unable to connect to server ", c.conns[serverNum].name, err)
		c.conns[serverNum].conn = nil
	}
	fmt.Println("Done")
}
