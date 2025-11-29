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

	for _, s := range(orderedServices) {
		service := config.Services[s]

		go starter.StartService(&service, ctx)
	}

	<-ctx.Done()
}
