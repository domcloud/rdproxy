package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/tidwall/redcon"
)

func main() {
	var s Config
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	handler := NewHandler(s)

	log.Printf("started server at %s", s.Listen)

	err = redcon.ListenAndServe(s.Listen,
		handler.ServeRESP,
		handler.AcceptConn,
		handler.ClosedConn,
	)
	if err != nil {
		log.Fatal(err)
	}
}
