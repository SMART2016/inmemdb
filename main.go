package main

import (
	"flag"
	"github.com/inmemdb/inmem/config"
	"github.com/inmemdb/inmem/server"
	"log"
)

func main() {
	setupFlags()
	log.Println("Starting InMemDb server !")
	server.RunInMemDBServer()
}

func setupFlags() {
	//Says the host
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for inmem server")

	// INFO: default Port for REDIS server is 6379
	flag.IntVar(&config.Port, "port", 7379, "port for inmem server")
	flag.Parse()
}
