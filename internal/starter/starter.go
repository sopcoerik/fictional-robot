package starter

import (
	"os/exec"
	"os"
	"log"
	"syscall"
	"context"

	"github.com/sopcoerik/fictional-robot/internal/parser"
)


func StartService(service *parser.Service, ctx context.Context) {
	


	for {
		select {
			case <-ctx.Done():
				log.Println("Service shutting down...")

				return

			default:
				cmd := exec.Command("sh", "-c", service.Command)	
	
				cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr



				err := cmd.Start()
				if err != nil {
					log.Fatal(err)
				}

				log.Printf("Waiting for command to finish...")
				err = cmd.Wait()
				log.Printf("Command finished with error: %v", err)

		}
	}

}
