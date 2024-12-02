package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/notnil/chess/uci"
	"github.com/rpnahm/distsys-chess-engine/pkg/common"
)

// global variables for catalog addresses and such
var catalogAddr = "catalog.cse.nd.edu"
var catalogPort = 9097

type Worker struct {
	name     string
	listener net.Listener
	address  string
	port     string
	eng      *uci.Engine
	conn     net.Conn
	// posId    int
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
	return
}

// Run the worker, handles the main for loop
func (w *Worker) Run() {
	defer w.listener.Close()
	var err error

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
			// If the read is empty for some reason, loop
			if n == 0 {
				continue
			}

			log.Println("Unable to read connection: ", err)
			break
		}

		print(buf[:n])
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

	nsAddressString := fmt.Sprintf("%s:%d", catalogAddr, catalogPort)
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
