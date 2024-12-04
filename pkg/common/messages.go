package common

import "time"

/*
	This file will contain different json formats
	to help with json marshall and json unmarshalling.
	*** All messages must start with a type value ***

	Notices:
		All position messages must be in FEN notation (notnils/chess )
*/

// Error message: An error message to let the client know that the previous operation failed for some reason
type Error struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

// NewGame message: for signalling a newgame to the server
type NewGame struct {
	Type     string   `json:"type"`
	Options  []string `json:"options"` // A list of strings of the format "name value" separated by a space
	Position string   `json:"position"`
	PosId    int      `json:"pos_id"`
}

// ReadyOk message: server confirming the position
type ReadyOk struct {
	Type  string `json:"type"`
	PosId int    `json:"pos_id"`
}

// ParseMoves message: Sends all necessary data to the server to look at moves
type ParseMoves struct {
	Type     string    `json:"type"`
	Position string    `json:"position"`
	PosId    int       `json:"pos_id"`
	Moves    []string  `json:"moves"`
	DueTime  time.Time `json:"due_time"`
	JobId    int       `json:"job_id"`
}

// Working message: Acknowledges that the server is working
type Working struct {
	Type  string `json:"type"`
	PosId int    `json:"pos_id"`
	JobId int    `json:"job_id"`
}

// Results message: Returns the results of the search (ideally by the response time)
type Results struct {
	Type     string `json:"type"`
	JobId    int    `json:"job_id"`
	BestMove string `json:"best_move"`
	Score    int    `json:"score"`
	Mate     int    `json:"mate"`
	Nodes    int    `json:"nodes"`
}

// NewPosition message: updates the position of the board
type NewPos struct {
	Type     string `json:"type"`
	Position string `json:"position"`
	PosId    int    `json:"pos_id"`
}

// Stop message: Signals to the server to close the connection
// Can either contain type: stop or stop: exit depending on the need
type Stop struct {
	Type string `json:"type"`
}
