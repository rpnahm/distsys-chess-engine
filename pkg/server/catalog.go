package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// global variables for catalog addresses and such
var catalogAddr = "catalog.cse.nd.edu"
var catalogPort = 9097

// Accessable variables
var Name = ""

type message struct {
	Type  string `json:"type"`
	Owner string `json:"owner"`
	Port  int    `json:"port"`
	// Address string
	Project string `json:"project"`
}

func CatalogMessage(typeServer, owner, project string, address *net.TCPAddr) {
	m := message{
		Type:  typeServer,
		Owner: owner,
		Port:  address.Port,
		// Address: address.String(),
		Project: project,
	}

	// encode the json data
	jsonData, err := json.Marshal(m)
	_ = jsonData
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
