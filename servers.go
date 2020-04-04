package sbproclient

import (
	"log"
	"time"
)

type Server struct {
	Address  string
	Location *time.Location
}

var defaultTicksServer = &Server{Address: "62.138.11.5:6777"}
var defauktTicksHistoryServer = &Server{Address: "62.138.11.5:6627"}
var defaultSymbolsServer = &Server{Address: "62.138.11.5:7730"}

func init() {
	log.SetFlags(0)

	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		log.Fatalln(err)
	}

	defaultTicksServer.Location = loc
	defauktTicksHistoryServer.Location = loc
}
