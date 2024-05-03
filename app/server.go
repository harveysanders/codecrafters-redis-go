package main

import (
	"fmt"
	"log"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/internal/qulifi"
)

func main() {
	addr := "0.0.0.0:6379"
	srv := qulifi.Server{}

	log.Printf("qulifi server starting @ %s\n", addr)
	err := srv.ListenAndServe(addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
