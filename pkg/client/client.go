package client

import (
	"net"

	"github.com/notnil/chess"
)

// struct to store all client data
type Client struct {
	BaseServerName string
	NumServers     int
	Game           chess.Game
	Position       chess.Position
	Conn           net.Conn
	PosId          int
	JobId          int
}
