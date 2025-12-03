package main

import (
	"fmt"
	"os/signal"
	"os"
	"context"
	"syscall"
	"net"
	"time"

	"github.com/sopcoerik/fictional-robot/internal/parser"
	"github.com/sopcoerik/fictional-robot/internal/sorter"
	"github.com/sopcoerik/fictional-robot/internal/starter"
)

func CheckHealth(ctx context.Context, url string, timeout time.Duration) (error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("health check timeout")
		case <-ticker.C:
			conn, err := net.Dial("tcp", url)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}

func main() {
	defer func() {
		fmt.Println("Maybe this works")
	}()

	fmt.Println("hello, fictional robot")

	config := parser.ParseConfig("devenv.yaml")

	orderedServices := sorter.SortServices(config) 

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serviceChan := make(chan error)

	for _, s := range(orderedServices) {
		service := config.Services[s]
		serviceUrl := fmt.Sprintf("localhost:%d", service.Port)
		
		fmt.Printf("Starting %s\n", s)

		go starter.StartService(&service, ctx, serviceChan)

		err := <-serviceChan
		if err != nil {
			fmt.Println("an error occurred while starting process\n", err.Error())
			stop()
		}

		err = CheckHealth(ctx, serviceUrl, 30 * time.Second)

		if err != nil {
			fmt.Printf("error making request to %s", s)
			stop()
		}

		fmt.Printf("Started %s\n", s)
	}

	<-ctx.Done()
}
