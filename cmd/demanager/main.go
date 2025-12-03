package main

import (
	"fmt"
	"os/signal"
	"os"
	"context"
	"syscall"

	"github.com/sopcoerik/fictional-robot/internal/parser"
	"github.com/sopcoerik/fictional-robot/internal/sorter"
	"github.com/sopcoerik/fictional-robot/internal/starter"
)

func main() {
	fmt.Println("hello, fictional robot")

	config := parser.ParseConfig("devenv.yaml")

	orderedServices := sorter.SortServices(config) 

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serviceChan := make(chan error)

	for _, s := range(orderedServices) {
		service := config.Services[s]
		
		fmt.Printf("Starting %s\n", s)

		go starter.StartService(&service, ctx, serviceChan)

		err := <-serviceChan

		if err != nil {
			fmt.Println("an error occurred while starting process\n", err.Error())
		}

		fmt.Printf("Started %s\n", s)

		// if no error we try to ping the service
		// should be a new function that tries
		// to get a health check for <time>
	}

	<-ctx.Done()
}
