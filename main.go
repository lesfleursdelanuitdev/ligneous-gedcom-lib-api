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

	srv := server.New(port)
	fmt.Printf("ligneous-gedcom-lib-api listening on :%d\n", port)
	log.Fatal(srv.ListenAndServe())
}
