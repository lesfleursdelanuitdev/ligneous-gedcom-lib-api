package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom-lib-api/internal/server"
)

func main() {
	portFlag := flag.Int("port", 8091, "server port")
	flag.Parse()

	port := *portFlag
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":" + strconv.Itoa(port)
	}

	srv := server.New(listenAddr)
	fmt.Printf("ligneous-gedcom-lib-api listening on %s\n", listenAddr)
	log.Fatal(srv.ListenAndServe())
}
