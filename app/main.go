package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/internal/inmem"
	"github.com/codecrafters-io/redis-starter-go/app/internal/qulifi"
)

func main() {
	addr := "0.0.0.0:6379"

	srv := qulifi.NewServer(
		qulifi.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))),
		qulifi.WithStore(inmem.New()),
	)

	log.Printf("qulifi server starting @ %s\n", addr)
	err := srv.ListenAndServe(addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
