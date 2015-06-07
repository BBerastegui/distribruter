package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jroimartin/rpcmq"
)

func main() {
	c := rpcmq.NewClient("amqp://localhost:5672",
		"rcp-queue", "rpc-client", "rpc-exchange", "direct")
	if err := c.Init(); err != nil {
		log.Fatalf("Init: %v", err)
	}
	defer c.Shutdown()

	// Keep getting results
	go func() {
		for r := range c.Results() {
			if r.Err != "" {
				log.Printf("Received error: %v (%v)", r.Err, r.UUID)
				continue
			}
			log.Printf("Received: %v (%v)\n", string(r.Data), r.UUID)
		}
	}()

	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)
	signal.Notify(ctrlC, syscall.SIGTERM)
	<-ctrlC
}
