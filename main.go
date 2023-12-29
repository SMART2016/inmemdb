package main

import (
	"flag"
	"log"

	"github.com/inmemdb/inmem/config"
	"github.com/inmemdb/inmem/server"
)

func main() {
	setupFlags()
	log.Println("Starting InMemDb server !")
	//server.RunInMemDBSyncServer()

	asyncServer := server.NewAsyncServer()
	asyncServer.RunInMemDBASyncServer()
}

func setupFlags() {
	//Says the host
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for inmem server")

	// INFO: default Port for REDIS server is 6379
	flag.IntVar(&config.Port, "port", 7379, "port for inmem server")
	flag.Parse()
}
